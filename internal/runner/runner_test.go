package runner

import (
	"context"
	"testing"

	"llm-eval/internal/cache"
	"llm-eval/internal/provider"
	"llm-eval/internal/suite"
)

type echoProvider struct{}

func (echoProvider) Name() string { return "echo" }

func (echoProvider) Generate(_ context.Context, req provider.Request) (provider.Response, error) {
	return provider.Response{Output: req.Input, Raw: req.Input}, nil
}

func TestRunWithConcurrencyPreservesCaseOrder(t *testing.T) {
	evalSuite := suite.Suite{
		Name: "concurrent",
		Cases: []suite.TestCase{
			{ID: "case-1", Input: "A", Expected: "A", Evaluator: "exact_match"},
			{ID: "case-2", Input: "B", Expected: "B", Evaluator: "exact_match"},
			{ID: "case-3", Input: "C", Expected: "C", Evaluator: "exact_match"},
			{ID: "case-4", Input: "D", Expected: "D", Evaluator: "exact_match"},
		},
	}

	result, err := New(echoProvider{}, cache.NoopStore{}).WithConcurrency(3).Run(context.Background(), evalSuite)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed != 4 {
		t.Fatalf("expected all cases to pass, got %d", result.Passed)
	}
	for i, item := range result.Cases {
		if item.ID != evalSuite.Cases[i].ID {
			t.Fatalf("case order changed at index %d: got %s want %s", i, item.ID, evalSuite.Cases[i].ID)
		}
	}
}
