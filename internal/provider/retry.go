package provider

import (
	"context"
	"fmt"
	"time"
)

func withRetry(ctx context.Context, attempts int, baseDelay time.Duration, fn func() ([]byte, bool, error)) ([]byte, error) {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		data, retryable, err := fn()
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !retryable || attempt == attempts {
			break
		}

		delay := baseDelay * time.Duration(1<<(attempt-1))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("request failed after %d attempt(s): %w", attempts, lastErr)
}
