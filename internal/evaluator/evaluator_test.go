package evaluator

import "testing"

func TestExactMatchNormalizesWhitespace(t *testing.T) {
	result := ExactMatch{}.Evaluate("  final   answer ", "final answer")
	if !result.Passed {
		t.Fatalf("expected normalized strings to match")
	}
}

func TestRegex(t *testing.T) {
	result := Regex{}.Evaluate("answer: 42", `\d+`)
	if !result.Passed {
		t.Fatalf("expected regex to match")
	}
}
