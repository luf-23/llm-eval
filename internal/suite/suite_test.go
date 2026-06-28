package suite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suite.yaml")
	err := os.WriteFile(path, []byte(`
name: sample
cases:
  - input: "1 + 1"
    expected: "2"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Cases[0].ID != "case-1" {
		t.Fatalf("expected default case id, got %q", s.Cases[0].ID)
	}
	if s.Cases[0].Evaluator != "exact_match" {
		t.Fatalf("expected default evaluator, got %q", s.Cases[0].Evaluator)
	}
}
