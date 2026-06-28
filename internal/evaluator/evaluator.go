package evaluator

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

type Result struct {
	Passed     bool    `json:"passed"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
	Prediction string  `json:"prediction"`
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
	case "choice_match":
		return ChoiceMatch{}, nil
	case "number_match":
		return NumberMatch{}, nil
	case "math_match":
		return MathMatch{}, nil
	default:
		return nil, fmt.Errorf("unknown evaluator %q", name)
	}
}

type ExactMatch struct{}

func (ExactMatch) Name() string { return "exact_match" }

func (ExactMatch) Evaluate(output string, expected string) Result {
	prediction := normalize(output)
	passed := prediction == normalize(expected)
	return boolResult(passed, "normalized output must equal expected answer", prediction)
}

type Contains struct{}

func (Contains) Name() string { return "contains" }

func (Contains) Evaluate(output string, expected string) Result {
	passed := strings.Contains(strings.ToLower(output), strings.ToLower(expected))
	return boolResult(passed, "output must contain expected answer", normalize(output))
}

type Regex struct{}

func (Regex) Name() string { return "regex" }

func (Regex) Evaluate(output string, expected string) Result {
	re, err := regexp.Compile(expected)
	if err != nil {
		return Result{Passed: false, Score: 0, Reason: "invalid expected regex: " + err.Error()}
	}
	passed := re.MatchString(output)
	return boolResult(passed, "output must match expected regex", normalize(output))
}

type ChoiceMatch struct{}

func (ChoiceMatch) Name() string { return "choice_match" }

func (ChoiceMatch) Evaluate(output string, expected string) Result {
	actual, ok := extractChoice(output)
	if !ok {
		return Result{Passed: false, Score: 0, Reason: "no choice option A/B/C/D found in output"}
	}
	want, ok := extractChoice(expected)
	if !ok {
		want = strings.ToUpper(strings.TrimSpace(expected))
	}
	passed := actual == want
	return boolResult(passed, fmt.Sprintf("extracted choice %q must equal expected choice %q", actual, want), actual)
}

type NumberMatch struct{}

func (NumberMatch) Name() string { return "number_match" }

func (NumberMatch) Evaluate(output string, expected string) Result {
	actual, ok := extractNumber(output)
	if !ok {
		return Result{Passed: false, Score: 0, Reason: "no numeric answer found in output"}
	}
	want, ok := extractNumber(expected)
	if !ok {
		want = strings.TrimSpace(expected)
	}
	passed := numericEqual(actual, want)
	return boolResult(passed, fmt.Sprintf("extracted number %q must equal expected number %q", actual, want), actual)
}

type MathMatch struct{}

func (MathMatch) Name() string { return "math_match" }

func (MathMatch) Evaluate(output string, expected string) Result {
	actual := extractBoxed(output)
	if actual == "" {
		if number, ok := extractNumber(output); ok {
			actual = number
		} else {
			actual = output
		}
	}
	want := extractBoxed(expected)
	if want == "" {
		want = expected
	}

	passed := numericEqual(actual, want)
	if !passed {
		passed = normalizeMath(actual) == normalizeMath(want)
	}
	return boolResult(passed, fmt.Sprintf("extracted math answer %q must equal expected answer %q", actual, want), actual)
}

func normalize(v string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(v)), " ")
}

func boolResult(passed bool, reason string, prediction string) Result {
	if passed {
		return Result{Passed: true, Score: 1, Reason: reason, Prediction: prediction}
	}
	return Result{Passed: false, Score: 0, Reason: reason, Prediction: prediction}
}

func extractChoice(v string) (string, bool) {
	re := regexp.MustCompile(`(?i)(?:^|[^A-Z])([ABCD])(?:[^A-Z]|$)`)
	matches := re.FindAllStringSubmatch(v, -1)
	if len(matches) == 0 {
		return "", false
	}
	return strings.ToUpper(matches[len(matches)-1][1]), true
}

func extractNumber(v string) (string, bool) {
	if boxed := extractBoxed(v); boxed != "" {
		if number, ok := extractNumberToken(boxed); ok {
			return number, true
		}
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)####\s*([-+]?\d[\d,]*(?:\.\d+)?(?:/\d[\d,]*)?)`),
		regexp.MustCompile(`(?i)(?:final answer|answer is|answer:)\s*([-+]?\d[\d,]*(?:\.\d+)?(?:/\d[\d,]*)?)`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(v, -1)
		if len(matches) > 0 {
			return cleanNumber(matches[len(matches)-1][1]), true
		}
	}

	return extractNumberToken(v)
}

func extractNumberToken(v string) (string, bool) {
	re := regexp.MustCompile(`[-+]?\d[\d,]*(?:\.\d+)?(?:/\d[\d,]*)?`)
	matches := re.FindAllString(v, -1)
	if len(matches) == 0 {
		return "", false
	}
	return cleanNumber(matches[len(matches)-1]), true
}

func extractBoxed(v string) string {
	re := regexp.MustCompile(`\\boxed\{([^{}]+)\}`)
	matches := re.FindAllStringSubmatch(v, -1)
	if len(matches) == 0 {
		return ""
	}
	return strings.TrimSpace(matches[len(matches)-1][1])
}

func numericEqual(actual string, expected string) bool {
	left, okLeft := parseNumber(actual)
	right, okRight := parseNumber(expected)
	if !okLeft || !okRight {
		return false
	}
	return left.Cmp(right) == 0
}

func parseNumber(v string) (*big.Rat, bool) {
	cleaned := cleanNumber(v)
	if cleaned == "" {
		return nil, false
	}
	rat := new(big.Rat)
	if _, ok := rat.SetString(cleaned); ok {
		return rat, true
	}
	return nil, false
}

func cleanNumber(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, " .。,:;，；")
	v = strings.ReplaceAll(v, ",", "")
	return v
}

func normalizeMath(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, " .。,:;，；")
	v = strings.ReplaceAll(v, `\left`, "")
	v = strings.ReplaceAll(v, `\right`, "")
	v = strings.ReplaceAll(v, "$", "")
	v = strings.ReplaceAll(v, ",", "")
	var b strings.Builder
	for _, r := range strings.ToLower(v) {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
