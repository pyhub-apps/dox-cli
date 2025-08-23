package generate

import (
	"fmt"
	"strings"
)

// TokenEstimator provides token counting and cost estimation for AI models
type TokenEstimator struct {
	model string
}

// NewTokenEstimator creates a new token estimator for the given model
func NewTokenEstimator(model string) *TokenEstimator {
	return &TokenEstimator{
		model: model,
	}
}

// EstimateTokens estimates the number of tokens in a text
// This is a simplified estimation - actual token count may vary
func (te *TokenEstimator) EstimateTokens(text string) int {
	// Simple estimation: ~4 characters per token for English text
	// This is a rough approximation; actual tokenization is more complex
	
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	
	// Count words and characters
	words := strings.Fields(text)
	wordCount := len(words)
	charCount := len(text)
	
	// Estimate based on both word and character count
	// Average English word is ~4-5 characters, ~1.3 tokens
	tokensByWords := int(float64(wordCount) * 1.3)
	tokensByChars := charCount / 4
	
	// Use the average of both methods
	estimatedTokens := (tokensByWords + tokensByChars) / 2
	
	// Add buffer for special tokens and formatting
	estimatedTokens = int(float64(estimatedTokens) * 1.1)
	
	return estimatedTokens
}

// EstimateCost calculates the estimated cost based on model pricing
func (te *TokenEstimator) EstimateCost(promptTokens, completionTokens int) (float64, string) {
	// Pricing per 1M tokens (as of 2024)
	// These are example prices and should be updated based on current pricing
	
	var promptPrice, completionPrice float64
	var currency = "USD"
	
	switch {
	// OpenAI Models
	case strings.HasPrefix(te.model, "gpt-4-turbo"), strings.HasPrefix(te.model, "gpt-4-1106"):
		promptPrice = 10.00       // $10 per 1M input tokens
		completionPrice = 30.00    // $30 per 1M output tokens
	case strings.HasPrefix(te.model, "gpt-4"):
		promptPrice = 30.00       // $30 per 1M input tokens
		completionPrice = 60.00    // $60 per 1M output tokens
	case strings.HasPrefix(te.model, "gpt-3.5-turbo"):
		promptPrice = 0.50        // $0.50 per 1M input tokens
		completionPrice = 1.50     // $1.50 per 1M output tokens
		
	// Claude Models
	case strings.Contains(te.model, "claude-3-opus"):
		promptPrice = 15.00       // $15 per 1M input tokens
		completionPrice = 75.00    // $75 per 1M output tokens
	case strings.Contains(te.model, "claude-3-sonnet"):
		promptPrice = 3.00        // $3 per 1M input tokens
		completionPrice = 15.00    // $15 per 1M output tokens
	case strings.Contains(te.model, "claude-3-haiku"):
		promptPrice = 0.25        // $0.25 per 1M input tokens
		completionPrice = 1.25     // $1.25 per 1M output tokens
		
	default:
		// Default pricing for unknown models
		promptPrice = 1.00
		completionPrice = 2.00
	}
	
	// Calculate cost (price is per 1M tokens)
	promptCost := (float64(promptTokens) / 1_000_000) * promptPrice
	completionCost := (float64(completionTokens) / 1_000_000) * completionPrice
	totalCost := promptCost + completionCost
	
	return totalCost, currency
}

// GetModelInfo returns information about the model's capabilities
func (te *TokenEstimator) GetModelInfo() ModelInfo {
	info := ModelInfo{
		Model: te.model,
	}
	
	// Set context window sizes
	switch {
	case strings.HasPrefix(te.model, "gpt-4-turbo"), strings.HasPrefix(te.model, "gpt-4-1106"):
		info.ContextWindow = 128000
		info.MaxOutput = 4096
	case strings.HasPrefix(te.model, "gpt-4"):
		info.ContextWindow = 8192
		info.MaxOutput = 4096
	case strings.HasPrefix(te.model, "gpt-3.5-turbo-16k"):
		info.ContextWindow = 16384
		info.MaxOutput = 4096
	case strings.HasPrefix(te.model, "gpt-3.5-turbo"):
		info.ContextWindow = 4096
		info.MaxOutput = 4096
		
	case strings.Contains(te.model, "claude-3"):
		info.ContextWindow = 200000
		info.MaxOutput = 4096
		
	default:
		info.ContextWindow = 4096
		info.MaxOutput = 2048
	}
	
	return info
}

// ModelInfo contains information about a model's capabilities
type ModelInfo struct {
	Model         string
	ContextWindow int // Maximum input tokens
	MaxOutput     int // Maximum output tokens
}

// FormatCostEstimate formats the cost estimate for display
func FormatCostEstimate(promptTokens, completionTokens int, cost float64, currency string) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Token Estimate:\n"))
	sb.WriteString(fmt.Sprintf("  Input tokens:  ~%d\n", promptTokens))
	sb.WriteString(fmt.Sprintf("  Output tokens: ~%d (max)\n", completionTokens))
	sb.WriteString(fmt.Sprintf("  Total tokens:  ~%d\n", promptTokens+completionTokens))
	sb.WriteString(fmt.Sprintf("\nEstimated Cost:\n"))
	sb.WriteString(fmt.Sprintf("  ~$%.4f %s\n", cost, currency))
	
	return sb.String()
}

// FormatModelInfo formats model information for display
func FormatModelInfo(info ModelInfo) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Model Information:\n"))
	sb.WriteString(fmt.Sprintf("  Model:         %s\n", info.Model))
	sb.WriteString(fmt.Sprintf("  Context:       %d tokens\n", info.ContextWindow))
	sb.WriteString(fmt.Sprintf("  Max output:    %d tokens\n", info.MaxOutput))
	
	return sb.String()
}