package cache

import (
	"context"
	"testing"
	"time"
)

func TestAIRequest_Hash(t *testing.T) {
	req1 := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Prompt:      "Hello, world!",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	req2 := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Prompt:      "Hello, world!",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	req3 := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Prompt:      "Different prompt",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	// Same requests should have same hash
	hash1 := req1.Hash()
	hash2 := req2.Hash()
	if hash1 != hash2 {
		t.Errorf("Same requests have different hashes: %s != %s", hash1, hash2)
	}

	// Different requests should have different hash
	hash3 := req3.Hash()
	if hash1 == hash3 {
		t.Error("Different requests have same hash")
	}

	// Hash should be consistent
	for i := 0; i < 10; i++ {
		if req1.Hash() != hash1 {
			t.Error("Hash is not consistent")
		}
	}

	// Test normalization
	req4 := &AIRequest{
		Provider:    "OPENAI", // Different case
		Model:       "GPT-3.5-TURBO",
		Prompt:      "  Hello, world!  ", // Extra whitespace
		ContentType: "CUSTOM",
		MaxTokens:   100,
		Temperature: 0.7,
	}
	
	// Should normalize and match req1
	if req4.Hash() != req1.Hash() {
		t.Error("Normalization not working correctly")
	}
}

func TestAICache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(DefaultOptions())
	defer lruCache.Close()
	
	aiCache := NewAICache(lruCache, 1*time.Hour)

	request := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Prompt:      "Test prompt",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	response := &AIResponse{
		Content:    "Test response",
		Provider:   "openai",
		Model:      "gpt-3.5-turbo",
		TokensUsed: 50,
	}

	// Set response
	err := aiCache.Set(ctx, request, response)
	if err != nil {
		t.Fatalf("Failed to set response: %v", err)
	}

	// Get response
	cachedResp, found := aiCache.Get(ctx, request)
	if !found {
		t.Fatal("Expected to find cached response")
	}

	if cachedResp.Content != response.Content {
		t.Errorf("Content mismatch: got %s, want %s", cachedResp.Content, response.Content)
	}

	if cachedResp.TokensUsed != response.TokensUsed {
		t.Errorf("TokensUsed mismatch: got %d, want %d", cachedResp.TokensUsed, response.TokensUsed)
	}

	// Check timestamp was set
	if cachedResp.Timestamp.IsZero() {
		t.Error("Timestamp was not set")
	}
}

func TestAICache_CacheMiss(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(DefaultOptions())
	defer lruCache.Close()
	
	aiCache := NewAICache(lruCache, 1*time.Hour)

	request := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Prompt:      "Uncached prompt",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	// Should not find uncached request
	_, found := aiCache.Get(ctx, request)
	if found {
		t.Error("Expected cache miss for uncached request")
	}
}

func TestAICache_Delete(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(DefaultOptions())
	defer lruCache.Close()
	
	aiCache := NewAICache(lruCache, 1*time.Hour)

	request := &AIRequest{
		Provider:    "claude",
		Model:       "claude-3-sonnet",
		Prompt:      "Test prompt",
		ContentType: "blog",
		MaxTokens:   200,
		Temperature: 0.8,
	}

	response := &AIResponse{
		Content:  "Test blog content",
		Provider: "claude",
		Model:    "claude-3-sonnet",
	}

	// Set and delete
	aiCache.Set(ctx, request, response)
	err := aiCache.Delete(ctx, request)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should not find deleted response
	_, found := aiCache.Get(ctx, request)
	if found {
		t.Error("Expected response to be deleted")
	}
}

func TestAICache_TTL(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(Options{
		MaxSize:         10,
		CleanupInterval: 50 * time.Millisecond,
	})
	defer lruCache.Close()
	
	// Use very short TTL for testing
	aiCache := NewAICache(lruCache, 100*time.Millisecond)

	request := &AIRequest{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "Expiring prompt",
		ContentType: "custom",
		MaxTokens:   100,
		Temperature: 0.5,
	}

	response := &AIResponse{
		Content: "This will expire",
	}

	// Set response
	aiCache.Set(ctx, request, response)

	// Should find immediately
	_, found := aiCache.Get(ctx, request)
	if !found {
		t.Error("Expected to find response immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = aiCache.Get(ctx, request)
	if found {
		t.Error("Expected response to be expired")
	}
}

func TestTemplateCache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(DefaultOptions())
	defer lruCache.Close()
	
	templateCache := NewTemplateCache(lruCache, 24*time.Hour)

	content := []byte("template content")
	checksum := CalculateChecksum(content)

	templateData := &TemplateData{
		Path:     "/path/to/template.docx",
		Content:  "parsed template",
		Checksum: checksum,
		Variables: map[string]interface{}{
			"var1": "value1",
			"var2": 42,
		},
	}

	// Set template data
	err := templateCache.Set(ctx, templateData)
	if err != nil {
		t.Fatalf("Failed to set template data: %v", err)
	}

	// Get template data
	cached, found := templateCache.Get(ctx, templateData.Path, checksum)
	if !found {
		t.Fatal("Expected to find cached template")
	}

	if cached.Path != templateData.Path {
		t.Errorf("Path mismatch: got %s, want %s", cached.Path, templateData.Path)
	}

	if cached.Variables["var1"] != "value1" {
		t.Error("Variable mismatch")
	}
}

func TestTemplateCache_ChecksumValidation(t *testing.T) {
	ctx := context.Background()
	lruCache := NewLRUCache(DefaultOptions())
	defer lruCache.Close()
	
	templateCache := NewTemplateCache(lruCache, 24*time.Hour)

	content1 := []byte("original content")
	checksum1 := CalculateChecksum(content1)

	templateData := &TemplateData{
		Path:     "/path/to/template.docx",
		Content:  "parsed template",
		Checksum: checksum1,
	}

	// Set template data
	templateCache.Set(ctx, templateData)

	// Try to get with different checksum (file changed)
	content2 := []byte("modified content")
	checksum2 := CalculateChecksum(content2)

	_, found := templateCache.Get(ctx, templateData.Path, checksum2)
	if found {
		t.Error("Should not find template with different checksum")
	}

	// Should find with correct checksum
	_, found = templateCache.Get(ctx, templateData.Path, checksum1)
	if !found {
		t.Error("Should find template with correct checksum")
	}
}

func TestCalculateChecksum(t *testing.T) {
	content1 := []byte("test content")
	content2 := []byte("test content")
	content3 := []byte("different content")

	checksum1 := CalculateChecksum(content1)
	checksum2 := CalculateChecksum(content2)
	checksum3 := CalculateChecksum(content3)

	// Same content should have same checksum
	if checksum1 != checksum2 {
		t.Error("Same content has different checksums")
	}

	// Different content should have different checksum
	if checksum1 == checksum3 {
		t.Error("Different content has same checksum")
	}

	// Checksum should be hex string
	if len(checksum1) != 64 { // SHA256 produces 32 bytes = 64 hex chars
		t.Errorf("Invalid checksum length: %d", len(checksum1))
	}
}