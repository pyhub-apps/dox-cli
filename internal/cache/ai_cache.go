package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AIRequest represents an AI API request for caching
type AIRequest struct {
	Provider    string  `json:"provider"`    // "openai" or "claude"
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	ContentType string  `json:"content_type"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// Hash generates a unique hash for the request
func (r *AIRequest) Hash() string {
	// Normalize the request for consistent hashing
	normalized := AIRequest{
		Provider:    strings.ToLower(r.Provider),
		Model:       strings.ToLower(r.Model),
		Prompt:      strings.TrimSpace(r.Prompt),
		ContentType: strings.ToLower(r.ContentType),
		MaxTokens:   r.MaxTokens,
		Temperature: r.Temperature,
	}
	
	// Create JSON representation
	data, _ := json.Marshal(normalized)
	
	// Generate SHA256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// AIResponse represents a cached AI response
type AIResponse struct {
	Content   string    `json:"content"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
	TokensUsed int      `json:"tokens_used,omitempty"`
}

// AICache provides specialized caching for AI responses
type AICache struct {
	cache Cache
	ttl   time.Duration
}

// NewAICache creates a new AI response cache
func NewAICache(cache Cache, ttl time.Duration) *AICache {
	if ttl <= 0 {
		ttl = 1 * time.Hour // Default TTL
	}
	
	return &AICache{
		cache: cache,
		ttl:   ttl,
	}
}

// Get retrieves a cached AI response
func (c *AICache) Get(ctx context.Context, request *AIRequest) (*AIResponse, bool) {
	key := c.buildKey(request)
	
	value, found := c.cache.Get(ctx, key)
	if !found {
		return nil, false
	}
	
	response, ok := value.(*AIResponse)
	if !ok {
		// Invalid cached type, remove it
		c.cache.Delete(ctx, key)
		return nil, false
	}
	
	return response, true
}

// Set stores an AI response in the cache
func (c *AICache) Set(ctx context.Context, request *AIRequest, response *AIResponse) error {
	key := c.buildKey(request)
	
	// Set timestamp if not already set
	if response.Timestamp.IsZero() {
		response.Timestamp = time.Now()
	}
	
	return c.cache.Set(ctx, key, response, c.ttl)
}

// Delete removes a cached AI response
func (c *AICache) Delete(ctx context.Context, request *AIRequest) error {
	key := c.buildKey(request)
	return c.cache.Delete(ctx, key)
}

// buildKey creates a cache key from the request
func (c *AICache) buildKey(request *AIRequest) string {
	return fmt.Sprintf("ai:%s:%s", request.Provider, request.Hash())
}

// Clear removes all AI responses from the cache
func (c *AICache) Clear(ctx context.Context) error {
	// If the underlying cache supports pattern-based deletion,
	// we could clear only AI-related entries.
	// For now, we'll document that this clears the entire cache.
	return c.cache.Clear(ctx)
}

// Stats returns cache statistics
func (c *AICache) Stats() *Statistics {
	return c.cache.Stats()
}

// SetTTL updates the default TTL for AI responses
func (c *AICache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// TemplateCache provides specialized caching for template processing
type TemplateCache struct {
	cache Cache
	ttl   time.Duration
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(cache Cache, ttl time.Duration) *TemplateCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour // Templates can be cached longer
	}
	
	return &TemplateCache{
		cache: cache,
		ttl:   ttl,
	}
}

// TemplateData represents cached template data
type TemplateData struct {
	Path      string                 `json:"path"`
	Content   interface{}            `json:"content"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Checksum  string                 `json:"checksum"`
	Timestamp time.Time              `json:"timestamp"`
}

// Get retrieves cached template data
func (c *TemplateCache) Get(ctx context.Context, path string, checksum string) (*TemplateData, bool) {
	key := c.buildKey(path, checksum)
	
	value, found := c.cache.Get(ctx, key)
	if !found {
		return nil, false
	}
	
	data, ok := value.(*TemplateData)
	if !ok {
		c.cache.Delete(ctx, key)
		return nil, false
	}
	
	// Verify checksum matches
	if data.Checksum != checksum {
		c.cache.Delete(ctx, key)
		return nil, false
	}
	
	return data, true
}

// Set stores template data in the cache
func (c *TemplateCache) Set(ctx context.Context, data *TemplateData) error {
	key := c.buildKey(data.Path, data.Checksum)
	
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}
	
	return c.cache.Set(ctx, key, data, c.ttl)
}

// buildKey creates a cache key for template data
func (c *TemplateCache) buildKey(path string, checksum string) string {
	return fmt.Sprintf("template:%s:%s", path, checksum)
}

// CalculateChecksum computes a checksum for file content
func CalculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}