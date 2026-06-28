package evaluator

import (
	"fmt"
	"regexp"
	"strings"
)

type Result struct {
	Passed bool    `json:"passed"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type Evaluator interface {
	Name() string
	Evaluate(output string, expected string) Result
}

func ByName(name string) (Evaluator, error) {
	switch name {
	case "", "exact_match":
		return ExactMatch{}, nil
	case "contains":
		return Contains{}, nil
	case "regex":
		return Regex{}, nil
	default:
		return nil, fmt.Errorf("unknown evaluator %q", name)
	}
}

type ExactMatch struct{}

func (ExactMatch) Name() string { return "exact_match" }

func (ExactMatch) Evaluate(output string, expected string) Result {
	passed := normalize(output) == normalize(expected)
	return boolResult(passed, "normalized output must equal expected answer")
}

type Contains struct{}

func (Contains) Name() string { return "contains" }

func (Contains) Evaluate(output string, expected string) Result {
	passed := strings.Contains(strings.ToLower(output), strings.ToLower(expected))
	return boolResult(passed, "output must contain expected answer")
}

type Regex struct{}

func (Regex) Name() string { return "regex" }

func (Regex) Evaluate(output string, expected string) Result {
	re, err := regexp.Compile(expected)
	if err != nil {
		return Result{Passed: false, Score: 0, Reason: "invalid expected regex: " + err.Error()}
	}
	passed := re.MatchString(output)
	return boolResult(passed, "output must match expected regex")
}

func normalize(v string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(v)), " ")
}

func boolResult(passed bool, reason string) Result {
	if passed {
		return Result{Passed: true, Score: 1, Reason: reason}
	}
	return Result{Passed: false, Score: 0, Reason: reason}
}
