package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
			client, err := NewClient(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_GenerateContent(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"message": "Invalid API key",
				},
			})
			return
		}

		// Parse request
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock response
		response := ChatCompletionResponse{
			ID:     "test-id",
			Object: "chat.completion",
			Model:  req.Model,
			Choices: []struct {
				Index   int     `json:"index"`
				Message Message `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Generated content for: " + req.Messages[1].Content,
					},
					FinishReason: "stop",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server URL
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.apiURL = server.URL

	tests := []struct {
		name    string
		prompt  string
		options GenerateOptions
		wantErr bool
	}{
		{
			name:   "Valid generation",
			prompt: "Test prompt",
			options: GenerateOptions{
				ContentType: "blog",
				Model:       "gpt-3.5-turbo",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name:   "Custom content type",
			prompt: "Another test",
			options: GenerateOptions{
				ContentType: "custom",
				Model:       "gpt-4",
				MaxTokens:   200,
				Temperature: 0.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := client.GenerateContent(tt.prompt, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && content == "" {
				t.Error("GenerateContent() returned empty content")
			}
			if !tt.wantErr && content != "Generated content for: "+tt.prompt {
				t.Errorf("GenerateContent() = %v, want %v", content, "Generated content for: "+tt.prompt)
			}
		})
	}
}

func TestBuildSystemMessage(t *testing.T) {
	client, _ := NewClient("test-key")

	tests := []struct {
		contentType string
		wantContains string
	}{
		{"blog", "blog writer"},
		{"report", "business analyst"},
		{"summary", "summarization"},
		{"code", "programmer"},
		{"custom", "helpful assistant"},
		{"unknown", "helpful assistant"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			msg := client.buildSystemMessage(tt.contentType)
			if len(msg) == 0 {
				t.Error("buildSystemMessage() returned empty message")
			}
		})
	}
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
						"object": "chat.completion",
						"created": 1234567890,
						"model": "gpt-3.5-turbo",
						"choices": [{
							"index": 0,
							"message": {"role": "assistant", "content": "Test response"},
							"finish_reason": "stop"
						}]
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
					body: `{"error": {"message": "Rate limit exceeded", "type": "rate_limit_error", "code": "rate_limit_exceeded"}}`,
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"object": "chat.completion",
						"created": 1234567890,
						"model": "gpt-3.5-turbo",
						"choices": [{
							"index": 0,
							"message": {"role": "assistant", "content": "Test response after retry"},
							"finish_reason": "stop"
						}]
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
					body: `{"error": {"message": "Internal server error", "type": "server_error"}}`,
				},
				{
					statusCode: http.StatusInternalServerError,
					body: `{"error": {"message": "Internal server error", "type": "server_error"}}`,
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"id": "test-id",
						"object": "chat.completion",
						"created": 1234567890,
						"model": "gpt-3.5-turbo",
						"choices": [{
							"index": 0,
							"message": {"role": "assistant", "content": "Test response after retries"},
							"finish_reason": "stop"
						}]
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
					body: `{"error": {"message": "Invalid request", "type": "invalid_request_error"}}`,
				},
			},
			expectSuccess: false,
			expectedCalls: 1,
		},
		{
			name: "max retries exceeded",
			serverResponses: []struct {
				statusCode int
				body       string
			}{
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"message": "Service unavailable", "type": "server_error"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"message": "Service unavailable", "type": "server_error"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"message": "Service unavailable", "type": "server_error"}}`,
				},
				{
					statusCode: http.StatusServiceUnavailable,
					body: `{"error": {"message": "Service unavailable", "type": "server_error"}}`,
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
				RetryableCheck: isRetryableOpenAIError,
			}
			client.SetRetryConfig(retryConfig)

			// Test with context
			ctx := context.Background()
			result, err := client.GenerateContentWithContext(ctx, "test prompt", GenerateOptions{
				ContentType: "custom",
				Model:       "gpt-3.5-turbo",
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
		w.Write([]byte(`{"error": {"message": "Service unavailable", "type": "server_error"}}`))
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
		RetryableCheck: isRetryableOpenAIError,
	}
	client.SetRetryConfig(retryConfig)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = client.GenerateContentWithContext(ctx, "test prompt", GenerateOptions{
		ContentType: "custom",
		Model:       "gpt-3.5-turbo",
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

func TestIsRetryableOpenAIError(t *testing.T) {
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
			err: &OpenAIError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limit exceeded",
				Code:       "rate_limit_exceeded",
			},
			wantRetry: true,
		},
		{
			name: "server error",
			err: &OpenAIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
				Type:       "server_error",
			},
			wantRetry: true,
		},
		{
			name: "bad gateway",
			err: &OpenAIError{
				StatusCode: http.StatusBadGateway,
				Message:    "Bad gateway",
			},
			wantRetry: true,
		},
		{
			name: "service unavailable",
			err: &OpenAIError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Service unavailable",
			},
			wantRetry: true,
		},
		{
			name: "gateway timeout",
			err: &OpenAIError{
				StatusCode: http.StatusGatewayTimeout,
				Message:    "Gateway timeout",
			},
			wantRetry: true,
		},
		{
			name: "bad request",
			err: &OpenAIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid request",
			},
			wantRetry: false,
		},
		{
			name: "unauthorized",
			err: &OpenAIError{
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
			got := isRetryableOpenAIError(tt.err)
			if got != tt.wantRetry {
				t.Errorf("isRetryableOpenAIError() = %v, want %v", got, tt.wantRetry)
			}
		})
	}
}