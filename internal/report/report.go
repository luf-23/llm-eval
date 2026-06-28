package report

import (
	"encoding/json"
	"fmt"
	"html/template"
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

	fmt.Fprintf(&b, "| Case | Evaluator | Passed | Expected | Prediction | Output |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- | --- | --- |\n")
	for _, item := range result.Cases {
		fmt.Fprintf(
			&b,
			"| %s | %s | %t | %s | %s | %s |\n",
			escape(item.ID),
			escape(item.Evaluator),
			item.Result.Passed,
			escape(item.Expected),
			escape(item.Result.Prediction),
			escape(item.Output),
		)
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func WriteHTML(result runner.EvaluationResult, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	view := htmlView{
		Result:          result,
		ScorePercent:    fmt.Sprintf("%.2f%%", result.Score*100),
		EvaluatorStats:  buildEvaluatorStats(result),
		ConfusionMatrix: buildChoiceConfusionMatrix(result),
	}
	return htmlReportTemplate.Execute(file, view)
}

func escape(v string) string {
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "|", "\\|")
	return v
}

type htmlView struct {
	Result          runner.EvaluationResult
	ScorePercent    string
	EvaluatorStats  []evaluatorStat
	ConfusionMatrix choiceConfusionMatrix
}

type evaluatorStat struct {
	Name         string
	Total        int
	Passed       int
	ScorePercent string
}

type choiceConfusionMatrix struct {
	Enabled bool
	Labels  []string
	Rows    []choiceConfusionRow
}

type choiceConfusionRow struct {
	Expected string
	Counts   []int
}

func buildEvaluatorStats(result runner.EvaluationResult) []evaluatorStat {
	type agg struct {
		total  int
		passed int
	}
	stats := make(map[string]agg)
	order := make([]string, 0)
	for _, item := range result.Cases {
		if _, ok := stats[item.Evaluator]; !ok {
			order = append(order, item.Evaluator)
		}
		current := stats[item.Evaluator]
		current.total++
		if item.Result.Passed {
			current.passed++
		}
		stats[item.Evaluator] = current
	}

	items := make([]evaluatorStat, 0, len(order))
	for _, name := range order {
		current := stats[name]
		score := 0.0
		if current.total > 0 {
			score = float64(current.passed) / float64(current.total) * 100
		}
		items = append(items, evaluatorStat{
			Name:         name,
			Total:        current.total,
			Passed:       current.passed,
			ScorePercent: fmt.Sprintf("%.2f%%", score),
		})
	}
	return items
}

func buildChoiceConfusionMatrix(result runner.EvaluationResult) choiceConfusionMatrix {
	labels := []string{"A", "B", "C", "D", "Unknown"}
	index := make(map[string]int, len(labels))
	for i, label := range labels {
		index[label] = i
	}

	rows := make([]choiceConfusionRow, 0, 4)
	rowIndex := make(map[string]int)
	enabled := false
	for _, item := range result.Cases {
		if item.Evaluator != "choice_match" {
			continue
		}
		enabled = true
		expected := strings.ToUpper(strings.TrimSpace(item.Expected))
		prediction := strings.ToUpper(strings.TrimSpace(item.Result.Prediction))
		if _, ok := index[prediction]; !ok {
			prediction = "Unknown"
		}
		if _, ok := rowIndex[expected]; !ok {
			rowIndex[expected] = len(rows)
			rows = append(rows, choiceConfusionRow{
				Expected: expected,
				Counts:   make([]int, len(labels)),
			})
		}
		rows[rowIndex[expected]].Counts[index[prediction]]++
	}

	return choiceConfusionMatrix{
		Enabled: enabled,
		Labels:  labels,
		Rows:    rows,
	}
}

var htmlReportTemplate = template.Must(template.New("html-report").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Result.SuiteName}} Evaluation Report</title>
  <style>
    * { box-sizing: border-box; }
    html, body { width: 100%; height: 100%; overflow: hidden; }
    body { margin: 0; font-family: Arial, "Microsoft YaHei", sans-serif; color: #1f2937; background: #f6f7f9; }
    main { height: 100vh; max-width: 1360px; margin: 0 auto; padding: 12px; display: flex; flex-direction: column; gap: 8px; overflow: hidden; }
    h1 { margin: 0; font-size: 22px; line-height: 1.15; }
    h2 { margin: 0 0 6px; font-size: 15px; line-height: 1.2; }
    .meta { color: #6b7280; font-size: 11px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .summary { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 8px; flex: 0 0 auto; }
    .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 6px; padding: 8px 10px; min-width: 0; }
    .label { color: #6b7280; font-size: 11px; }
    .value { margin-top: 2px; font-size: 22px; line-height: 1; font-weight: 700; }
    .top-grid { display: grid; grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr); gap: 8px; flex: 0 0 auto; min-height: 0; }
    .panel { min-width: 0; overflow: hidden; }
    .details { flex: 1 1 auto; min-height: 0; overflow: hidden; }
    table { width: 100%; table-layout: fixed; border-collapse: collapse; background: #fff; border: 1px solid #e5e7eb; }
    th, td { padding: 4px 6px; border-bottom: 1px solid #e5e7eb; text-align: left; vertical-align: middle; font-size: 11px; line-height: 1.2; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    th { background: #f3f4f6; font-weight: 700; }
    tr:last-child td { border-bottom: 0; }
    .pass { color: #047857; font-weight: 700; }
    .fail { color: #b91c1c; font-weight: 700; }
    .case-table th:nth-child(1), .case-table td:nth-child(1) { width: 18%; }
    .case-table th:nth-child(2), .case-table td:nth-child(2) { width: 13%; }
    .case-table th:nth-child(3), .case-table td:nth-child(3) { width: 8%; }
    .case-table th:nth-child(4), .case-table td:nth-child(4) { width: 11%; }
    .case-table th:nth-child(5), .case-table td:nth-child(5) { width: 11%; }
    .case-table th:nth-child(6), .case-table td:nth-child(6) { width: 39%; }
    .matrix td, .matrix th { text-align: center; }
    .matrix th:first-child, .matrix td:first-child { text-align: left; width: 32%; }
    @media print {
      @page { size: A4 landscape; margin: 8mm; }
      html, body { overflow: hidden; }
      main { height: auto; max-width: none; padding: 0; }
    }
  </style>
</head>
<body>
<main>
  <h1>{{.Result.SuiteName}} Evaluation Report</h1>
  <div class="meta">Provider: {{.Result.Provider}} | Started: {{.Result.StartedAt}} | Completed: {{.Result.CompletedAt}}</div>

  <section class="summary">
    <div class="card"><div class="label">Total</div><div class="value">{{.Result.Total}}</div></div>
    <div class="card"><div class="label">Passed</div><div class="value">{{.Result.Passed}}</div></div>
    <div class="card"><div class="label">Failed</div><div class="value">{{.Result.Failed}}</div></div>
    <div class="card"><div class="label">Score</div><div class="value">{{.ScorePercent}}</div></div>
  </section>

  <section class="top-grid">
    <div class="panel">
      <h2>Evaluator Summary</h2>
      <table>
        <thead><tr><th>Evaluator</th><th>Total</th><th>Passed</th><th>Score</th></tr></thead>
        <tbody>
          {{range .EvaluatorStats}}
          <tr><td>{{.Name}}</td><td>{{.Total}}</td><td>{{.Passed}}</td><td>{{.ScorePercent}}</td></tr>
          {{end}}
        </tbody>
      </table>
    </div>

    {{if .ConfusionMatrix.Enabled}}
    <div class="panel">
      <h2>Choice Confusion Matrix</h2>
      <table class="matrix">
        <thead>
          <tr><th>Expected \ Predicted</th>{{range .ConfusionMatrix.Labels}}<th>{{.}}</th>{{end}}</tr>
        </thead>
        <tbody>
          {{range .ConfusionMatrix.Rows}}
          <tr><td>{{.Expected}}</td>{{range .Counts}}<td>{{.}}</td>{{end}}</tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{end}}
  </section>

  <section class="details">
    <h2>Case Details</h2>
    <table class="case-table">
      <thead><tr><th>Case</th><th>Evaluator</th><th>Status</th><th>Expected</th><th>Prediction</th><th>Output</th></tr></thead>
      <tbody>
        {{range .Result.Cases}}
        <tr>
          <td>{{.ID}}</td>
          <td>{{.Evaluator}}</td>
          <td>{{if .Result.Passed}}<span class="pass">pass</span>{{else}}<span class="fail">fail</span>{{end}}</td>
          <td>{{.Expected}}</td>
          <td>{{.Result.Prediction}}</td>
          <td>{{.Output}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </section>
</main>
</body>
</html>`))
