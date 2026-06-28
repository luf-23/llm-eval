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

func TestChoiceMatchExtractsFinalOption(t *testing.T) {
	result := ChoiceMatch{}.Evaluate("After checking the options, the answer is B.", "B")
	if !result.Passed {
		t.Fatalf("expected extracted choice to match")
	}
}

func TestNumberMatchExtractsFinalNumber(t *testing.T) {
	result := NumberMatch{}.Evaluate("Tom has 12 apples, so the final answer is 40.", "40")
	if !result.Passed {
		t.Fatalf("expected extracted number to match")
	}
}

func TestNumberMatchComparesEquivalentFractions(t *testing.T) {
	result := NumberMatch{}.Evaluate("The answer is 0.5", "1/2")
	if !result.Passed {
		t.Fatalf("expected decimal and fraction to match")
	}
}

func TestMathMatchExtractsBoxedAnswer(t *testing.T) {
	result := MathMatch{}.Evaluate(`Therefore \boxed{13}.`, "13")
	if !result.Passed {
		t.Fatalf("expected boxed answer to match")
	}
}
