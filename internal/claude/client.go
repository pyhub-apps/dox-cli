package claude

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
	defaultAPIURL = "https://api.anthropic.com/v1/messages"
	defaultModel  = "claude-3-sonnet-20240229"
	apiVersion    = "2023-06-01"
)

// Client represents a Claude API client
type Client struct {
	apiKey      string
	apiURL      string
	httpClient  *http.Client
	retryConfig retry.Config
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, pkgErrors.NewValidationError("apiKey", apiKey, "API key is required")
	}

	// Set up retry configuration
	retryConfig := retry.DefaultConfig()
	retryConfig.MaxRetries = 3
	retryConfig.InitialDelay = 1 * time.Second
	retryConfig.MaxDelay = 10 * time.Second
	retryConfig.RetryableCheck = isRetryableClaudeError

	return &Client{
		apiKey: apiKey,
		apiURL: defaultAPIURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Claude may take longer for complex requests
		},
		retryConfig: retryConfig,
	}, nil
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessagesRequest represents the request payload for the Messages API
type MessagesRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
	System      string    `json:"system,omitempty"`
}

// MessagesResponse represents the response from the Claude API
type MessagesResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Model        string `json:"model"`
	Content      []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents an error from the Claude API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
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
	req := MessagesRequest{
		Model: options.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		System:      systemMessage,
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
		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", apiVersion)

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
				return "", &ClaudeError{
					StatusCode: resp.StatusCode,
					Message:    apiError.Error.Message,
					Type:       apiError.Error.Type,
				}
			}
			return "", retry.NewHTTPError(resp.StatusCode, string(body))
		}

		// Parse the response
		var msgResp MessagesResponse
		if err := json.Unmarshal(body, &msgResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		// Check for API error in response
		if msgResp.Error != nil {
			return "", &ClaudeError{
				Message: msgResp.Error.Message,
				Type:    msgResp.Error.Type,
			}
		}

		// Extract the generated content
		if len(msgResp.Content) == 0 {
			return "", fmt.Errorf("no content generated")
		}

		// Combine all text content
		var result string
		for _, content := range msgResp.Content {
			if content.Type == "text" {
				result += content.Text
			}
		}

		if result == "" {
			return "", fmt.Errorf("no text content in response")
		}

		return result, nil
	})
}

// buildSystemMessage creates appropriate system message based on content type
func (c *Client) buildSystemMessage(contentType string) string {
	switch contentType {
	case "blog":
		return "You are a professional blog writer. Create engaging, well-structured blog posts with clear sections, compelling introductions, and actionable conclusions. Use markdown formatting."
	case "report":
		return "You are a business analyst. Create professional reports with executive summaries, detailed analysis, clear data presentation, and actionable recommendations. Use clear headings and structured format."
	case "summary":
		return "You are an expert at summarization. Create concise, accurate summaries that capture the key points, main ideas, and essential details while maintaining clarity. Focus on the most important information."
	case "email":
		return "You are a professional email writer. Create clear, concise, and professional emails with appropriate greetings, clear purpose, well-organized content, and professional closings."
	case "proposal":
		return "You are a business proposal expert. Create compelling proposals with executive summaries, clear value propositions, detailed scope, timeline, and professional formatting."
	case "code":
		return "You are an expert programmer. Generate clean, well-documented code following best practices with proper error handling, clear comments, and optimal performance considerations."
	case "custom":
		return "You are Claude, a helpful AI assistant. Provide clear, accurate, and helpful responses to the user's request. Be concise but comprehensive."
	default:
		return "You are Claude, a helpful AI assistant. Provide clear, accurate, and helpful responses to the user's request."
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

// AvailableModels returns a list of available Claude models
func AvailableModels() []string {
	return []string{
		"claude-3-opus-20240229",    // Most capable, slower
		"claude-3-sonnet-20240229",  // Balanced (default)
		"claude-3-haiku-20240307",   // Fastest, most compact
		"claude-2.1",                // Previous generation
		"claude-2.0",                // Legacy
		"claude-instant-1.2",        // Fast, older model
	}
}

// ModelInfo provides information about Claude models
type ModelInfo struct {
	Name        string
	Description string
	MaxTokens   int
	Best4       string // Best for what use case
}

// GetModelInfo returns information about available models
func GetModelInfo() []ModelInfo {
	return []ModelInfo{
		{
			Name:        "claude-3-opus-20240229",
			Description: "Most capable Claude 3 model",
			MaxTokens:   4096,
			Best4:       "Complex tasks, nuanced content, creative writing",
		},
		{
			Name:        "claude-3-sonnet-20240229",
			Description: "Balanced performance and speed",
			MaxTokens:   4096,
			Best4:       "General purpose, good balance of quality and speed",
		},
		{
			Name:        "claude-3-haiku-20240307",
			Description: "Fastest Claude 3 model",
			MaxTokens:   4096,
			Best4:       "Quick responses, simple tasks, high volume",
		},
		{
			Name:        "claude-2.1",
			Description: "Previous generation, 200K context",
			MaxTokens:   4096,
			Best4:       "Long documents, extended conversations",
		},
		{
			Name:        "claude-instant-1.2",
			Description: "Fast, lightweight model",
			MaxTokens:   4096,
			Best4:       "Quick drafts, simple content",
		},
	}
}

// ClaudeError represents an error from the Claude API with additional metadata
type ClaudeError struct {
	StatusCode int
	Message    string
	Type       string
}

func (e *ClaudeError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("Claude API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Claude API error: %s", e.Message)
}

// isRetryableClaudeError determines if a Claude error should be retried
func isRetryableClaudeError(err error) bool {
	if err == nil {
		return false
	}

	// Check for Claude specific errors
	var claudeErr *ClaudeError
	if errors.As(err, &claudeErr) {
		// Retry on rate limits and server errors
		switch claudeErr.StatusCode {
		case http.StatusTooManyRequests, // 429
		     http.StatusInternalServerError, // 500
		     http.StatusBadGateway, // 502
		     http.StatusServiceUnavailable, // 503
		     http.StatusGatewayTimeout: // 504
			return true
		}
		
		// Check for specific error types
		if claudeErr.Type == "rate_limit_error" || claudeErr.Type == "overloaded_error" {
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