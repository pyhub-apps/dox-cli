package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialDelay   time.Duration // Initial delay between retries
	MaxDelay       time.Duration // Maximum delay between retries
	Multiplier     float64       // Exponential backoff multiplier
	Jitter         bool          // Add random jitter to delays
	RetryableCheck func(error) bool // Custom function to determine if error is retryable
}

// DefaultConfig returns default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryableCheck: DefaultRetryableCheck,
	}
}

// DefaultRetryableCheck determines if an error should trigger a retry
func DefaultRetryableCheck(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	// Check for specific error types
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return IsRetryableHTTPStatus(httpErr.StatusCode)
	}

	// Check error message for common patterns
	errMsg := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"rate limit",
		"too many requests",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
	}

	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errMsg, pattern) {
			return true
		}
	}

	return false
}

// IsRetryableHTTPStatus checks if an HTTP status code is retryable
func IsRetryableHTTPStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,      // 429
	     http.StatusInternalServerError,   // 500
	     http.StatusBadGateway,           // 502
	     http.StatusServiceUnavailable,   // 503
	     http.StatusGatewayTimeout:       // 504
		return true
	default:
		return false
	}
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string) error {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// Do executes a function with retry logic
func Do(ctx context.Context, config Config, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		
		// Success
		if err == nil {
			return nil
		}
		
		// Save the error
		lastErr = err
		
		// Check if we should retry
		if attempt >= config.MaxRetries {
			break
		}
		
		// Check if error is retryable
		if config.RetryableCheck != nil && !config.RetryableCheck(err) {
			return err
		}
		
		// Calculate delay
		delay := calculateDelay(attempt, config)
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}
	
	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}

// DoWithResult executes a function with retry logic and returns a result
func DoWithResult[T any](ctx context.Context, config Config, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		res, err := fn()
		
		// Success
		if err == nil {
			return res, nil
		}
		
		// Save the error
		lastErr = err
		
		// Check if we should retry
		if attempt >= config.MaxRetries {
			break
		}
		
		// Check if error is retryable
		if config.RetryableCheck != nil && !config.RetryableCheck(err) {
			return result, err
		}
		
		// Calculate delay
		delay := calculateDelay(attempt, config)
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}
	
	return result, fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}

// calculateDelay calculates the delay for the given attempt
func calculateDelay(attempt int, config Config) time.Duration {
	// Calculate exponential backoff
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))
	
	// Apply max delay cap
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	// Add jitter if enabled
	if config.Jitter {
		// Add random jitter between 0% and 25% of the delay
		jitter := rand.Float64() * 0.25 * delay
		delay += jitter
	}
	
	return time.Duration(delay)
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	
	// Simple case-insensitive contains check
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))
	
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		sLower[i] = c
	}
	
	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		substrLower[i] = c
	}
	
	// Search for substring
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		match := true
		for j := 0; j < len(substrLower); j++ {
			if sLower[i+j] != substrLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}