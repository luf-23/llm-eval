package runner

import (
	"context"
	"time"

	"llm-eval/internal/cache"
	"llm-eval/internal/evaluator"
	"llm-eval/internal/provider"
	"llm-eval/internal/suite"
)

type Runner struct {
	provider provider.Provider
	cache    cache.Store
}

type EvaluationResult struct {
	SuiteName   string       `json:"suite_name"`
	Provider    string       `json:"provider"`
	StartedAt   time.Time    `json:"started_at"`
	CompletedAt time.Time    `json:"completed_at"`
	Total       int          `json:"total"`
	Passed      int          `json:"passed"`
	Failed      int          `json:"failed"`
	Score       float64      `json:"score"`
	Cases       []CaseResult `json:"cases"`
}

type CaseResult struct {
	ID        string           `json:"id"`
	Input     string           `json:"input"`
	Expected  string           `json:"expected"`
	Output    string           `json:"output"`
	Raw       string           `json:"raw"`
	Evaluator string           `json:"evaluator"`
	Cached    bool             `json:"cached"`
	Result    evaluator.Result `json:"result"`
	Duration  time.Duration    `json:"duration"`
}

func New(provider provider.Provider, cache cache.Store) Runner {
	return Runner{provider: provider, cache: cache}
}

func (r Runner) Run(ctx context.Context, s suite.Suite) (EvaluationResult, error) {
	result := EvaluationResult{
		SuiteName: s.Name,
		Provider:  r.provider.Name(),
		StartedAt: time.Now(),
		Cases:     make([]CaseResult, 0, len(s.Cases)),
	}

	for _, testCase := range s.Cases {
		started := time.Now()
		ev, err := evaluator.ByName(testCase.Evaluator)
		if err != nil {
			return EvaluationResult{}, err
		}

		key := cache.Key(r.provider.Name(), s.Prompt, testCase.ID, testCase.Input)
		resp, cached := r.cache.Get(key)
		if !cached {
			resp, err = r.provider.Generate(ctx, provider.Request{
				SuitePrompt: s.Prompt,
				CaseID:      testCase.ID,
				Input:       testCase.Input,
			})
			if err != nil {
				return EvaluationResult{}, err
			}
			if err := r.cache.Set(key, resp); err != nil {
				return EvaluationResult{}, err
			}
		}

		score := ev.Evaluate(resp.Output, testCase.Expected)
		if score.Passed {
			result.Passed++
		}
		result.Cases = append(result.Cases, CaseResult{
			ID:        testCase.ID,
			Input:     testCase.Input,
			Expected:  testCase.Expected,
			Output:    resp.Output,
			Raw:       resp.Raw,
			Evaluator: ev.Name(),
			Cached:    cached,
			Result:    score,
			Duration:  time.Since(started),
		})
	}

	result.Total = len(result.Cases)
	result.Failed = result.Total - result.Passed
	if result.Total > 0 {
		result.Score = float64(result.Passed) / float64(result.Total)
	}
	result.CompletedAt = time.Now()
	return result, nil
}
