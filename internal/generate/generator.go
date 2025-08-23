package generate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pyhub/pyhub-docs/internal/cache"
	"github.com/pyhub/pyhub-docs/internal/claude"
	"github.com/pyhub/pyhub-docs/internal/config"
	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/openai"
	"github.com/pyhub/pyhub-docs/internal/retry"
	"github.com/pyhub/pyhub-docs/internal/ui"
)

// AIProvider represents the AI provider type
type AIProvider string

const (
	ProviderOpenAI AIProvider = "openai"
	ProviderClaude AIProvider = "claude"
)

// Generator handles content generation using AI
type Generator struct {
	provider      AIProvider
	openaiClient  *openai.Client
	claudeClient  *claude.Client
	cache         *cache.AICache
}

// GenerateOptions contains options for content generation (provider-agnostic)
type GenerateOptions struct {
	ContentType string
	Model       string
	MaxTokens   int
	Temperature float64
}

// NewGenerator creates a new content generator
func NewGenerator(provider AIProvider, apiKey string) (*Generator, error) {
	gen := &Generator{
		provider: provider,
	}

	switch provider {
	case ProviderOpenAI:
		// Check for API key from environment if not provided
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		
		if apiKey == "" {
			return nil, pkgErrors.NewConfigError("", "OpenAI API key not found", pkgErrors.ErrMissingAPIKey)
		}

		client, err := openai.NewClient(apiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
		}
		gen.openaiClient = client

	case ProviderClaude:
		// Check for API key from environment if not provided
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
			if apiKey == "" {
				apiKey = os.Getenv("CLAUDE_API_KEY") // Alternative env var
			}
		}
		
		if apiKey == "" {
			return nil, pkgErrors.NewConfigError("", "Claude API key not found", pkgErrors.ErrMissingAPIKey)
		}

		client, err := claude.NewClient(apiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create Claude client: %w", err)
		}
		gen.claudeClient = client

	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}

	return gen, nil
}

// NewGeneratorWithConfig creates a new content generator with retry configuration and caching
func NewGeneratorWithConfig(provider AIProvider, apiKey string, cfg *config.Config) (*Generator, error) {
	gen, err := NewGenerator(provider, apiKey)
	if err != nil {
		return nil, err
	}

	// Setup caching
	lruCache := cache.NewLRUCache(cache.Options{
		MaxSize:         100,
		MaxBytes:        50 * 1024 * 1024, // 50MB for AI responses
		DefaultTTL:      1 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	})
	gen.cache = cache.NewAICache(lruCache, 1*time.Hour)

	// Apply retry configuration based on provider
	switch provider {
	case ProviderOpenAI:
		if gen.openaiClient != nil && cfg != nil {
			retryConfig := retry.Config{
				MaxRetries:   cfg.OpenAI.Retry.MaxRetries,
				InitialDelay: time.Duration(cfg.OpenAI.Retry.InitialDelay) * time.Millisecond,
				MaxDelay:     time.Duration(cfg.OpenAI.Retry.MaxDelay) * time.Millisecond,
				Multiplier:   cfg.OpenAI.Retry.Multiplier,
				Jitter:       cfg.OpenAI.Retry.Jitter,
				RetryableCheck: nil, // Will use the default retryable check
			}
			gen.openaiClient.SetRetryConfig(retryConfig)
		}

	case ProviderClaude:
		if gen.claudeClient != nil && cfg != nil {
			retryConfig := retry.Config{
				MaxRetries:   cfg.Claude.Retry.MaxRetries,
				InitialDelay: time.Duration(cfg.Claude.Retry.InitialDelay) * time.Millisecond,
				MaxDelay:     time.Duration(cfg.Claude.Retry.MaxDelay) * time.Millisecond,
				Multiplier:   cfg.Claude.Retry.Multiplier,
				Jitter:       cfg.Claude.Retry.Jitter,
				RetryableCheck: nil, // Will use the default retryable check
			}
			gen.claudeClient.SetRetryConfig(retryConfig)
		}
	}

	return gen, nil
}

// EnableCache enables caching with custom settings
func (g *Generator) EnableCache(ttl time.Duration, maxSize int) {
	lruCache := cache.NewLRUCache(cache.Options{
		MaxSize:         maxSize,
		MaxBytes:        100 * 1024 * 1024, // 100MB
		DefaultTTL:      ttl,
		CleanupInterval: 5 * time.Minute,
	})
	g.cache = cache.NewAICache(lruCache, ttl)
}

// DisableCache disables caching
func (g *Generator) DisableCache() {
	g.cache = nil
}

// GenerateContent generates content based on the provided options
func (g *Generator) GenerateContent(prompt string, options GenerateOptions) (string, error) {
	// Validate prompt
	if strings.TrimSpace(prompt) == "" {
		return "", pkgErrors.NewValidationError("prompt", prompt, "prompt cannot be empty")
	}

	// Check if prompt is a file path (starts with @ or looks like a file)
	if strings.HasPrefix(prompt, "@") {
		filePath := strings.TrimPrefix(prompt, "@")
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", pkgErrors.NewFileError(filePath, "reading prompt file", err)
		}
		prompt = string(content)
	}

	ctx := context.Background()

	// Create cache request
	cacheRequest := &cache.AIRequest{
		Provider:    string(g.provider),
		Model:       options.Model,
		Prompt:      prompt,
		ContentType: options.ContentType,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	// Check cache if enabled
	if g.cache != nil {
		if cachedResponse, found := g.cache.Get(ctx, cacheRequest); found {
			ui.PrintInfo("Using cached response (cache hit)")
			return cachedResponse.Content, nil
		}
	}

	// Generate content based on provider
	var content string
	var err error

	switch g.provider {
	case ProviderOpenAI:
		if g.openaiClient == nil {
			return "", fmt.Errorf("OpenAI client not initialized")
		}
		openaiOpts := openai.GenerateOptions{
			ContentType: options.ContentType,
			Model:       options.Model,
			MaxTokens:   options.MaxTokens,
			Temperature: options.Temperature,
		}
		content, err = g.openaiClient.GenerateContent(prompt, openaiOpts)

	case ProviderClaude:
		if g.claudeClient == nil {
			return "", fmt.Errorf("Claude client not initialized")
		}
		claudeOpts := claude.GenerateOptions{
			ContentType: options.ContentType,
			Model:       options.Model,
			MaxTokens:   options.MaxTokens,
			Temperature: options.Temperature,
		}
		content, err = g.claudeClient.GenerateContent(prompt, claudeOpts)

	default:
		return "", fmt.Errorf("unsupported provider: %s", g.provider)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Cache the response if cache is enabled
	if g.cache != nil && content != "" {
		cacheResponse := &cache.AIResponse{
			Content:   content,
			Provider:  string(g.provider),
			Model:     options.Model,
			Timestamp: time.Now(),
		}
		if err := g.cache.Set(ctx, cacheRequest, cacheResponse); err != nil {
			// Log cache error but don't fail the request
			ui.PrintWarning("Failed to cache response: %v", err)
		}
	}

	return content, nil
}

// GetCacheStats returns cache statistics if cache is enabled
func (g *Generator) GetCacheStats() *cache.Statistics {
	if g.cache != nil {
		return g.cache.Stats()
	}
	return nil
}

// SaveToFile saves the generated content to a file
func SaveToFile(content string, filePath string) error {
	if filePath == "" {
		return nil // No file specified, skip saving
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, ask for confirmation or use --force flag
		return pkgErrors.NewFileError(filePath, "writing output", pkgErrors.ErrFileAlreadyExists)
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return pkgErrors.NewFileError(filePath, "writing output", err)
	}

	return nil
}

// EnhancePrompt adds context or improvements to the user's prompt based on content type
func EnhancePrompt(prompt string, contentType string) string {
	switch contentType {
	case "blog":
		if !strings.Contains(strings.ToLower(prompt), "blog") && !strings.Contains(strings.ToLower(prompt), "article") {
			return fmt.Sprintf("Write a blog post about: %s\n\nInclude an engaging title, introduction, main sections with subheadings, and a conclusion.", prompt)
		}
	case "report":
		if !strings.Contains(strings.ToLower(prompt), "report") {
			return fmt.Sprintf("Create a professional report on: %s\n\nInclude an executive summary, detailed analysis, key findings, and recommendations.", prompt)
		}
	case "summary":
		if !strings.Contains(strings.ToLower(prompt), "summar") {
			return fmt.Sprintf("Summarize the following content:\n\n%s\n\nProvide a clear and concise summary highlighting the main points.", prompt)
		}
	case "email":
		if !strings.Contains(strings.ToLower(prompt), "email") {
			return fmt.Sprintf("Write a professional email about: %s\n\nInclude appropriate greeting, clear purpose, organized content, and professional closing.", prompt)
		}
	case "proposal":
		if !strings.Contains(strings.ToLower(prompt), "proposal") {
			return fmt.Sprintf("Create a business proposal for: %s\n\nInclude executive summary, objectives, scope, timeline, and next steps.", prompt)
		}
	case "code":
		if !strings.Contains(strings.ToLower(prompt), "code") && !strings.Contains(strings.ToLower(prompt), "function") {
			return fmt.Sprintf("Generate code for: %s\n\nInclude proper error handling, comments, and follow best practices.", prompt)
		}
	}
	return prompt
}

// DetectProviderFromModel detects the AI provider based on the model name
func DetectProviderFromModel(model string) AIProvider {
	modelLower := strings.ToLower(model)
	
	// Check for Claude models
	if strings.Contains(modelLower, "claude") {
		return ProviderClaude
	}
	
	// Check for OpenAI models
	if strings.Contains(modelLower, "gpt") || strings.Contains(modelLower, "davinci") || strings.Contains(modelLower, "turbo") {
		return ProviderOpenAI
	}
	
	// Default to OpenAI for backward compatibility
	return ProviderOpenAI
}

// GetAvailableModels returns available models for the specified provider
func GetAvailableModels(provider AIProvider) []string {
	switch provider {
	case ProviderClaude:
		return claude.AvailableModels()
	case ProviderOpenAI:
		return []string{
			"gpt-4-turbo-preview",
			"gpt-4",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
		}
	default:
		return []string{}
	}
}

// DefaultGenerateOptions returns default generation options
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		ContentType: "custom",
		Model:       "gpt-3.5-turbo", // Default to OpenAI for backward compatibility
		MaxTokens:   2000,
		Temperature: 0.7,
	}
}