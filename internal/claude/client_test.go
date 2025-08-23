package claude

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pyhub/pyhub-docs/internal/retry"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "Valid API key",
			apiKey:  "test-api-key",
			wantErr: false,
		},
		{
			name:    "Empty API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildSystemMessage(t *testing.T) {
	client := &Client{}
	
	tests := []struct {
		contentType string
		wantPrefix  string
	}{
		{
			contentType: "blog",
			wantPrefix:  "You are a professional blog writer",
		},
		{
			contentType: "report",
			wantPrefix:  "You are a business analyst",
		},
		{
			contentType: "summary",
			wantPrefix:  "You are an expert at summarization",
		},
		{
			contentType: "email",
			wantPrefix:  "You are a professional email writer",
		},
		{
			contentType: "proposal",
			wantPrefix:  "You are a business proposal expert",
		},
		{
			contentType: "code",
			wantPrefix:  "You are an expert programmer",
		},
		{
			contentType: "custom",
			wantPrefix:  "You are Claude",
		},
		{
			contentType: "unknown",
			wantPrefix:  "You are Claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := client.buildSystemMessage(tt.contentType)
			if len(result) == 0 {
				t.Error("buildSystemMessage() returned empty string")
			}
			// Check if the result starts with expected prefix
			if len(result) < len(tt.wantPrefix) || result[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("buildSystemMessage() = %v, want prefix %v", result[:50], tt.wantPrefix)
			}
		})
	}
}

func TestAvailableModels(t *testing.T) {
	models := AvailableModels()
	
	if len(models) == 0 {
		t.Error("AvailableModels() returned empty slice")
	}
	
	// Check for expected models
	expectedModels := []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
	
	for _, expected := range expectedModels {
		found := false
		for _, model := range models {
			if model == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found in AvailableModels()", expected)
		}
	}
}

func TestGetModelInfo(t *testing.T) {
	infos := GetModelInfo()
	
	if len(infos) == 0 {
		t.Error("GetModelInfo() returned empty slice")
	}
	
	// Check first model info
	if len(infos) > 0 {
		first := infos[0]
		if first.Name == "" {
			t.Error("Model info has empty name")
		}
		if first.Description == "" {
			t.Error("Model info has empty description")
		}
		if first.MaxTokens <= 0 {
			t.Error("Model info has invalid MaxTokens")
		}
		if first.Best4 == "" {
			t.Error("Model info has empty Best4")
		}
	}
}

func TestDefaultGenerateOptions(t *testing.T) {
	opts := DefaultGenerateOptions()
	
	if opts.ContentType != "custom" {
		t.Errorf("DefaultGenerateOptions() ContentType = %v, want custom", opts.ContentType)
	}
	if opts.Model != "claude-3-sonnet-20240229" {
		t.Errorf("DefaultGenerateOptions() Model = %v, want claude-3-sonnet-20240229", opts.Model)
	}
	if opts.MaxTokens != 2000 {
		t.Errorf("DefaultGenerateOptions() MaxTokens = %v, want 2000", opts.MaxTokens)
	}
	if opts.Temperature != 0.7 {
		t.Errorf("DefaultGenerateOptions() Temperature = %v, want 0.7", opts.Temperature)
	}
}

// Integration test - only run if API key is available
func TestGenerateContent(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CLAUDE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY or CLAUDE_API_KEY not set")
	}
	
	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test with a simple prompt
	options := GenerateOptions{
		ContentType: "custom",
		Model:       "claude-3-haiku-20240307", // Use the fastest/cheapest model for testing
		MaxTokens:   100,
		Temperature: 0.5,
	}
	
	content, err := client.GenerateContent("Say 'Hello, World!' and nothing else.", options)
	if err != nil {
		t.Fatalf("GenerateContent() error = %v", err)
	}
	
	if content == "" {
		t.Error("GenerateContent() returned empty content")
	}
	
	t.Logf("Generated content: %s", content)
}

func TestGenerateContentWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		serverResponses []struct {
			statusCode int
			body       string
		}
		expectSuccess  bool
		expectedCalls  int
	}{
		{
			name: "success on first attempt",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"type": "message",
						"role": "assistant",
						"model": "claude-3-sonnet-20240229",
						"content": [{"type": "text", "text": "Test response"}],
						"stop_reason": "end_turn"
					}`,
				},
			},
			expectSuccess: true,
			expectedCalls: 1,
		},
		{
			name: "retry on 429 rate limit",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusTooManyRequests,
					body: `{"error": {"type": "rate_limit_error", "message": "Rate limit exceeded"}}`,
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"type": "message",
						"role": "assistant",
						"model": "claude-3-sonnet-20240229",
						"content": [{"type": "text", "text": "Test response after retry"}],
						"stop_reason": "end_turn"
					}`,
				},
			},
			expectSuccess: true,
			expectedCalls: 2,
		},
		{
			name: "retry on 500 server error",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusInternalServerError,
					body: `{"error": {"type": "server_error", "message": "Internal server error"}}`,
				},
				{
					statusCode: http.StatusInternalServerError,
					body: `{"error": {"type": "server_error", "message": "Internal server error"}}`,
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"type": "message",
						"role": "assistant",
						"model": "claude-3-sonnet-20240229",
						"content": [{"type": "text", "text": "Test response after retries"}],
						"stop_reason": "end_turn"
					}`,
				},
			},
			expectSuccess: true,
			expectedCalls: 3,
		},
		{
			name: "no retry on 400 bad request",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusBadRequest,
					body: `{"error": {"type": "invalid_request_error", "message": "Invalid request"}}`,
				},
			},
			expectSuccess: false,
			expectedCalls: 1,
		},
		{
			name: "retry on overloaded error",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"type": "overloaded_error", "message": "Service overloaded"}}`,
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"type": "message",
						"role": "assistant",
						"model": "claude-3-sonnet-20240229",
						"content": [{"type": "text", "text": "Test response after overload"}],
						"stop_reason": "end_turn"
					}`,
				},
			},
			expectSuccess: true,
			expectedCalls: 2,
		},
		{
			name: "max retries exceeded",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"type": "server_error", "message": "Service unavailable"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"type": "server_error", "message": "Service unavailable"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"type": "server_error", "message": "Service unavailable"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"type": "server_error", "message": "Service unavailable"}}`,
				},
			},
			expectSuccess: false,
			expectedCalls: 4, // initial + 3 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if callCount >= len(tt.serverResponses) {
					// If we run out of responses, return the last one
					response := tt.serverResponses[len(tt.serverResponses)-1]
					w.WriteHeader(response.statusCode)
					w.Write([]byte(response.body))
				} else {
					response := tt.serverResponses[callCount]
					w.WriteHeader(response.statusCode)
					w.Write([]byte(response.body))
				}
				callCount++
			}))
			defer server.Close()

			// Create client with test server URL
			client, err := NewClient("test-api-key")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			client.apiURL = server.URL

			// Configure retry with shorter delays for testing
			retryConfig := retry.Config{
				MaxRetries:     3,
				InitialDelay:   10 * time.Millisecond,
				MaxDelay:       100 * time.Millisecond,
				Multiplier:     2.0,
				Jitter:         false,
				RetryableCheck: isRetryableClaudeError,
			}
			client.SetRetryConfig(retryConfig)

			// Test with context
			ctx := context.Background()
			result, err := client.GenerateContentWithContext(ctx, "test prompt", GenerateOptions{
				ContentType: "custom",
				Model:       "claude-3-sonnet-20240229",
				MaxTokens:   100,
				Temperature: 0.7,
			})

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if result == "" {
					t.Errorf("Expected non-empty result")
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			}

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d calls but got %d", tt.expectedCalls, callCount)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return 503 to trigger retries
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": {"type": "server_error", "message": "Service unavailable"}}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.apiURL = server.URL

	// Configure retry with longer delays
	retryConfig := retry.Config{
		MaxRetries:     5,
		InitialDelay:   1 * time.Second,
		MaxDelay:       10 * time.Second,
		Multiplier:     2.0,
		Jitter:         false,
		RetryableCheck: isRetryableClaudeError,
	}
	client.SetRetryConfig(retryConfig)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = client.GenerateContentWithContext(ctx, "test prompt", GenerateOptions{
		ContentType: "custom",
		Model:       "claude-3-sonnet-20240229",
		MaxTokens:   100,
		Temperature: 0.7,
	})

	elapsed := time.Since(start)

	if err == nil {
		t.Errorf("Expected error due to context cancellation")
	}

	// Should fail quickly due to context timeout, not wait for all retries
	if elapsed > 200*time.Millisecond {
		t.Errorf("Context cancellation took too long: %v", elapsed)
	}

	// Check that the error is related to context
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

func TestIsRetryableClaudeError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
		{
			name: "rate limit error",
			err: &ClaudeError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limit exceeded",
				Type:       "rate_limit_error",
			},
			wantRetry: true,
		},
		{
			name: "overloaded error",
			err: &ClaudeError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Service overloaded",
				Type:       "overloaded_error",
			},
			wantRetry: true,
		},
		{
			name: "server error",
			err: &ClaudeError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
				Type:       "server_error",
			},
			wantRetry: true,
		},
		{
			name: "bad gateway",
			err: &ClaudeError{
				StatusCode: http.StatusBadGateway,
				Message:    "Bad gateway",
			},
			wantRetry: true,
		},
		{
			name: "service unavailable",
			err: &ClaudeError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Service unavailable",
			},
			wantRetry: true,
		},
		{
			name: "gateway timeout",
			err: &ClaudeError{
				StatusCode: http.StatusGatewayTimeout,
				Message:    "Gateway timeout",
			},
			wantRetry: true,
		},
		{
			name: "bad request",
			err: &ClaudeError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request",
				Type:       "invalid_request_error",
			},
			wantRetry: false,
		},
		{
			name: "unauthorized",
			err: &ClaudeError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid API key",
			},
			wantRetry: false,
		},
		{
			name:      "generic error",
			err:       errors.New("some error"),
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableClaudeError(tt.err)
			if got != tt.wantRetry {
				t.Errorf("isRetryableClaudeError() = %v, want %v", got, tt.wantRetry)
			}
		})
	}
}