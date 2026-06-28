package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetryRetriesRetryableErrors(t *testing.T) {
	attempts := 0
	data, err := withRetry(context.Background(), 3, time.Millisecond, func() ([]byte, bool, error) {
		attempts++
		if attempts < 3 {
			return nil, true, errors.New("temporary")
		}
		return []byte("ok"), false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "ok" {
		t.Fatalf("expected ok response, got %q", string(data))
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetryDoesNotRetryPermanentErrors(t *testing.T) {
	attempts := 0
	_, err := withRetry(context.Background(), 3, time.Millisecond, func() ([]byte, bool, error) {
		attempts++
		return nil, false, errors.New("permanent")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
