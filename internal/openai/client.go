package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/retry"
)

const (
	defaultAPIURL = "https://api.openai.com/v1/chat/completions"
	defaultModel  = "gpt-3.5-turbo"
)

// Client represents an OpenAI API client
type Client struct {
	apiKey      string
	apiURL      string
	httpClient  *http.Client
	retryConfig retry.Config
}

// NewClient creates a new OpenAI API client
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, pkgErrors.NewValidationError("apiKey", apiKey, "API key is required")
	}

	// Set up retry configuration
	retryConfig := retry.DefaultConfig()
	retryConfig.MaxRetries = 3
	retryConfig.InitialDelay = 1 * time.Second
	retryConfig.MaxDelay = 10 * time.Second
	retryConfig.RetryableCheck = isRetryableOpenAIError

	return &Client{
		apiKey: apiKey,
		apiURL: defaultAPIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryConfig: retryConfig,
	}, nil
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request payload for chat completions
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// ChatCompletionResponse represents the response from the API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int     `json:"index"`
		Message Message `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents an error from the OpenAI API
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GenerateContent generates content based on the given prompt
func (c *Client) GenerateContent(prompt string, options GenerateOptions) (string, error) {
	// Use GenerateContentWithContext with a default context
	ctx := context.Background()
	return c.GenerateContentWithContext(ctx, prompt, options)
}

// GenerateContentWithContext generates content with context and retry support
func (c *Client) GenerateContentWithContext(ctx context.Context, prompt string, options GenerateOptions) (string, error) {
	// Build system message based on content type
	systemMessage := c.buildSystemMessage(options.ContentType)
	
	// Create the request
	req := ChatCompletionRequest{
		Model: options.Model,
		Messages: []Message{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// Marshal the request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Execute with retry logic
	return retry.DoWithResult(ctx, c.retryConfig, func() (string, error) {
		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

		// Send the request
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		// Check for HTTP errors
		if resp.StatusCode != http.StatusOK {
			var apiError struct {
				Error APIError `json:"error"`
			}
			if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error.Message != "" {
				// Return error with status code for retry logic
				return "", &OpenAIError{
					StatusCode: resp.StatusCode,
					Message:    apiError.Error.Message,
					Type:       apiError.Error.Type,
					Code:       apiError.Error.Code,
				}
			}
			return "", retry.NewHTTPError(resp.StatusCode, string(body))
		}

		// Parse the response
		var chatResp ChatCompletionResponse
		if err := json.Unmarshal(body, &chatResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		// Check for API error in response
		if chatResp.Error != nil {
			return "", fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
		}

		// Extract the generated content
		if len(chatResp.Choices) == 0 {
			return "", fmt.Errorf("no content generated")
		}

		return chatResp.Choices[0].Message.Content, nil
	})
}

// buildSystemMessage creates appropriate system message based on content type
func (c *Client) buildSystemMessage(contentType string) string {
	switch contentType {
	case "blog":
		return "You are a professional blog writer. Create engaging, well-structured blog posts with clear sections, compelling introductions, and actionable conclusions."
	case "report":
		return "You are a business analyst. Create professional reports with executive summaries, detailed analysis, clear data presentation, and actionable recommendations."
	case "summary":
		return "You are an expert at summarization. Create concise, accurate summaries that capture the key points, main ideas, and essential details while maintaining clarity."
	case "code":
		return "You are an expert programmer. Generate clean, well-documented code following best practices with proper error handling and clear comments."
	default:
		return "You are a helpful assistant. Provide clear, accurate, and helpful responses to the user's request."
	}
}

// GenerateOptions contains options for content generation
type GenerateOptions struct {
	ContentType string
	Model       string
	MaxTokens   int
	Temperature float64
}

// DefaultGenerateOptions returns default generation options
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		ContentType: "custom",
		Model:       defaultModel,
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}

// OpenAIError represents an error from the OpenAI API with additional metadata
type OpenAIError struct {
	StatusCode int
	Message    string
	Type       string
	Code       string
}

func (e *OpenAIError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("OpenAI API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("OpenAI API error: %s", e.Message)
}

// isRetryableOpenAIError determines if an OpenAI error should be retried
func isRetryableOpenAIError(err error) bool {
	if err == nil {
		return false
	}

	// Check for OpenAI specific errors
	var openAIErr *OpenAIError
	if errors.As(err, &openAIErr) {
		// Retry on rate limits and server errors
		switch openAIErr.StatusCode {
		case http.StatusTooManyRequests, // 429
		     http.StatusInternalServerError, // 500
		     http.StatusBadGateway, // 502
		     http.StatusServiceUnavailable, // 503
		     http.StatusGatewayTimeout: // 504
			return true
		}
		
		// Check for specific error codes
		if openAIErr.Code == "rate_limit_exceeded" || openAIErr.Type == "server_error" {
			return true
		}
	}

	// Fall back to default retry logic
	return retry.DefaultRetryableCheck(err)
}

// SetRetryConfig allows customizing the retry configuration
func (c *Client) SetRetryConfig(config retry.Config) {
	c.retryConfig = config
}