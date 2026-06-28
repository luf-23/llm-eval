package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"llm-eval/internal/evaluator"
	"llm-eval/internal/runner"
)

func TestWriteHTMLIncludesConfusionMatrix(t *testing.T) {
	result := runner.EvaluationResult{
		SuiteName:   "MMLU",
		Provider:    "qwen",
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		Total:       2,
		Passed:      1,
		Failed:      1,
		Score:       0.5,
		Cases: []runner.CaseResult{
			{
				ID:        "case-a",
				Expected:  "A",
				Output:    "A",
				Evaluator: "choice_match",
				Result:    evaluator.Result{Passed: true, Score: 1, Prediction: "A"},
			},
			{
				ID:        "case-b",
				Expected:  "B",
				Output:    "A",
				Evaluator: "choice_match",
				Result:    evaluator.Result{Passed: false, Score: 0, Prediction: "A"},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "report.html")
	if err := WriteHTML(result, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, want := range []string{"Choice Confusion Matrix", "Expected \\ Predicted", "case-a", "50.00%"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected html to contain %q", want)
		}
	}
}
