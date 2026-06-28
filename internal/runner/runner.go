package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"llm-eval/internal/cache"
	"llm-eval/internal/evaluator"
	"llm-eval/internal/provider"
	"llm-eval/internal/suite"
)

type Runner struct {
	provider    provider.Provider
	cache       cache.Store
	concurrency int
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
	return Runner{provider: provider, cache: cache, concurrency: 1}
}

func (r Runner) WithConcurrency(concurrency int) Runner {
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 5 {
		concurrency = 5
	}
	r.concurrency = concurrency
	return r
}

func (r Runner) Run(ctx context.Context, s suite.Suite) (EvaluationResult, error) {
	result := EvaluationResult{
		SuiteName: s.Name,
		Provider:  r.provider.Name(),
		StartedAt: time.Now(),
		Cases:     make([]CaseResult, 0, len(s.Cases)),
	}

	cases, err := r.runCases(ctx, s)
	if err != nil {
		return EvaluationResult{}, err
	}
	for _, item := range cases {
		if item.Result.Passed {
			result.Passed++
		}
	}
	result.Cases = cases

	result.Total = len(result.Cases)
	result.Failed = result.Total - result.Passed
	if result.Total > 0 {
		result.Score = float64(result.Passed) / float64(result.Total)
	}
	result.CompletedAt = time.Now()
	return result, nil
}

func (r Runner) runCases(ctx context.Context, s suite.Suite) ([]CaseResult, error) {
	if len(s.Cases) == 0 {
		return nil, nil
	}

	concurrency := r.concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(s.Cases) {
		concurrency = len(s.Cases)
	}

	cases := make([]CaseResult, len(s.Cases))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var runErr error

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for index := worker; index < len(s.Cases); index += concurrency {
				if ctx.Err() != nil {
					return
				}

				testCase := s.Cases[index]
				result, err := r.runCase(ctx, s.Prompt, testCase)
				if err != nil {
					errMu.Lock()
					if runErr == nil {
						runErr = fmt.Errorf("case %s: %w", testCase.ID, err)
					}
					errMu.Unlock()
					return
				}
				cases[index] = result
			}
		}(worker)
	}

	wg.Wait()
	if runErr != nil {
		return nil, runErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return cases, nil
}

func (r Runner) runCase(ctx context.Context, prompt string, testCase suite.TestCase) (CaseResult, error) {
	started := time.Now()
	ev, err := evaluator.ByName(testCase.Evaluator)
	if err != nil {
		return CaseResult{ID: testCase.ID}, err
	}

	key := cache.Key(r.provider.Name(), prompt, testCase.ID, testCase.Input)
	resp, cached := r.cache.Get(key)
	if !cached {
		resp, err = r.provider.Generate(ctx, provider.Request{
			SuitePrompt: prompt,
			CaseID:      testCase.ID,
			Input:       testCase.Input,
		})
		if err != nil {
			return CaseResult{ID: testCase.ID}, err
		}
		if err := r.cache.Set(key, resp); err != nil {
			return CaseResult{ID: testCase.ID}, err
		}
	}

	score := ev.Evaluate(resp.Output, testCase.Expected)
	return CaseResult{
		ID:        testCase.ID,
		Input:     testCase.Input,
		Expected:  testCase.Expected,
		Output:    resp.Output,
		Raw:       resp.Raw,
		Evaluator: ev.Name(),
		Cached:    cached,
		Result:    score,
		Duration:  time.Since(started),
	}, nil
}
