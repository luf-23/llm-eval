package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"llm-eval/internal/runner"
)

func WriteJSON(result runner.EvaluationResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func WriteMarkdown(result runner.EvaluationResult, path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Evaluation Report\n\n")
	fmt.Fprintf(&b, "- Suite: %s\n", result.SuiteName)
	fmt.Fprintf(&b, "- Provider: %s\n", result.Provider)
	fmt.Fprintf(&b, "- Total: %d\n", result.Total)
	fmt.Fprintf(&b, "- Passed: %d\n", result.Passed)
	fmt.Fprintf(&b, "- Failed: %d\n", result.Failed)
	fmt.Fprintf(&b, "- Score: %.2f%%\n\n", result.Score*100)

	fmt.Fprintf(&b, "| Case | Evaluator | Passed | Expected | Output |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- | --- |\n")
	for _, item := range result.Cases {
		fmt.Fprintf(
			&b,
			"| %s | %s | %t | %s | %s |\n",
			escape(item.ID),
			escape(item.Evaluator),
			item.Result.Passed,
			escape(item.Expected),
			escape(item.Output),
		)
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func escape(v string) string {
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "|", "\\|")
	return v
}
