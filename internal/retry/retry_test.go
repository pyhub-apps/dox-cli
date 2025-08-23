package retry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestDefaultRetryableCheck(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantRetry bool
	}{
		{
			name:     "nil error",
			err:      nil,
			wantRetry: false,
		},
		{
			name:     "timeout error",
			err:      context.DeadlineExceeded,
			wantRetry: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			wantRetry: true,
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded"),
			wantRetry: true,
		},
		{
			name:     "HTTP 429",
			err:      NewHTTPError(http.StatusTooManyRequests, "too many requests"),
			wantRetry: true,
		},
		{
			name:     "HTTP 500",
			err:      NewHTTPError(http.StatusInternalServerError, "internal server error"),
			wantRetry: true,
		},
		{
			name:     "HTTP 503",
			err:      NewHTTPError(http.StatusServiceUnavailable, "service unavailable"),
			wantRetry: true,
		},
		{
			name:     "HTTP 400",
			err:      NewHTTPError(http.StatusBadRequest, "bad request"),
			wantRetry: false,
		},
		{
			name:     "HTTP 404",
			err:      NewHTTPError(http.StatusNotFound, "not found"),
			wantRetry: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some random error"),
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultRetryableCheck(tt.err)
			if got != tt.wantRetry {
				t.Errorf("DefaultRetryableCheck() = %v, want %v", got, tt.wantRetry)
			}
		})
	}
}

func TestDo(t *testing.T) {
	tests := []struct {
		name          string
		attempts      int
		failUntil     int
		retryable     bool
		wantErr       bool
		wantAttempts  int
	}{
		{
			name:          "success on first attempt",
			attempts:      3,
			failUntil:     0,
			retryable:     true,
			wantErr:       false,
			wantAttempts:  1,
		},
		{
			name:          "success after 2 retries",
			attempts:      3,
			failUntil:     2,
			retryable:     true,
			wantErr:       false,
			wantAttempts:  3,
		},
		{
			name:          "max retries exceeded",
			attempts:      3,
			failUntil:     10,
			retryable:     true,
			wantErr:       true,
			wantAttempts:  4, // initial + 3 retries
		},
		{
			name:          "non-retryable error",
			attempts:      3,
			failUntil:     10,
			retryable:     false,
			wantErr:       true,
			wantAttempts:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attemptCount := 0
			config := Config{
				MaxRetries:   tt.attempts,
				InitialDelay: 1 * time.Millisecond,
				MaxDelay:     10 * time.Millisecond,
				Multiplier:   2.0,
				Jitter:       false,
				RetryableCheck: func(err error) bool {
					return tt.retryable
				},
			}

			err := Do(context.Background(), config, func() error {
				attemptCount++
				if attemptCount <= tt.failUntil {
					return fmt.Errorf("attempt %d failed", attemptCount)
				}
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if attemptCount != tt.wantAttempts {
				t.Errorf("Do() attemptCount = %v, want %v", attemptCount, tt.wantAttempts)
			}
		})
	}
}

func TestDoWithResult(t *testing.T) {
	config := Config{
		MaxRetries:   2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
		RetryableCheck: DefaultRetryableCheck,
	}

	t.Run("success with result", func(t *testing.T) {
		attemptCount := 0
		result, err := DoWithResult(context.Background(), config, func() (string, error) {
			attemptCount++
			if attemptCount < 2 {
				return "", errors.New("connection refused")  // Use a retryable error
			}
			return "success", nil
		})

		if err != nil {
			t.Errorf("DoWithResult() unexpected error: %v", err)
		}
		if result != "success" {
			t.Errorf("DoWithResult() = %v, want success", result)
		}
		if attemptCount != 2 {
			t.Errorf("DoWithResult() attemptCount = %v, want 2", attemptCount)
		}
	})

	t.Run("failure with max retries", func(t *testing.T) {
		attemptCount := 0
		result, err := DoWithResult(context.Background(), config, func() (string, error) {
			attemptCount++
			return "", errors.New("connection refused")
		})

		if err == nil {
			t.Errorf("DoWithResult() expected error, got nil")
		}
		if result != "" {
			t.Errorf("DoWithResult() = %v, want empty string", result)
		}
		if attemptCount != 3 { // initial + 2 retries
			t.Errorf("DoWithResult() attemptCount = %v, want 3", attemptCount)
		}
	})
}

func TestDoWithContext(t *testing.T) {
	config := Config{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
		RetryableCheck: DefaultRetryableCheck,
	}

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		attemptCount := 0
		err := Do(ctx, config, func() error {
			attemptCount++
			return errors.New("connection refused")  // Use a retryable error
		})

		if err == nil {
			t.Errorf("Do() expected error due to context cancellation")
		}
		if attemptCount > 2 {
			t.Errorf("Do() attemptCount = %v, expected <= 2 due to timeout", attemptCount)
		}
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Errorf("Do() expected context error, got: %v", err)
		}
	})
}

func TestCalculateDelay(t *testing.T) {
	config := Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // capped at max delay
		{5, 1 * time.Second}, // capped at max delay
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			got := calculateDelay(tt.attempt, config)
			if got != tt.want {
				t.Errorf("calculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	config := Config{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Test that jitter adds variability
	delay1 := calculateDelay(1, config)
	delay2 := calculateDelay(1, config)
	
	// Base delay should be 200ms, with jitter it should be between 200ms and 250ms
	baseDelay := 200 * time.Millisecond
	maxJitterDelay := 250 * time.Millisecond
	
	if delay1 < baseDelay || delay1 > maxJitterDelay {
		t.Errorf("calculateDelay with jitter = %v, expected between %v and %v", delay1, baseDelay, maxJitterDelay)
	}
	if delay2 < baseDelay || delay2 > maxJitterDelay {
		t.Errorf("calculateDelay with jitter = %v, expected between %v and %v", delay2, baseDelay, maxJitterDelay)
	}
}