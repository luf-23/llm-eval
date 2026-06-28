package suite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Suite struct {
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Prompt      string     `json:"prompt" yaml:"prompt"`
	Cases       []TestCase `json:"cases" yaml:"cases"`
}

type TestCase struct {
	ID        string `json:"id" yaml:"id"`
	Input     string `json:"input" yaml:"input"`
	Expected  string `json:"expected" yaml:"expected"`
	Evaluator string `json:"evaluator" yaml:"evaluator"`
}

func Load(path string) (Suite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Suite{}, err
	}

	var s Suite
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &s)
	case ".json":
		err = json.Unmarshal(data, &s)
	default:
		return Suite{}, fmt.Errorf("unsupported suite file type %q", filepath.Ext(path))
	}
	if err != nil {
		return Suite{}, err
	}
	if s.Name == "" {
		return Suite{}, fmt.Errorf("suite name is required")
	}
	if len(s.Cases) == 0 {
		return Suite{}, fmt.Errorf("suite must contain at least one case")
	}
	for i := range s.Cases {
		if s.Cases[i].ID == "" {
			s.Cases[i].ID = fmt.Sprintf("case-%d", i+1)
		}
		if s.Cases[i].Evaluator == "" {
			s.Cases[i].Evaluator = "exact_match"
		}
	}
	return s, nil
}
