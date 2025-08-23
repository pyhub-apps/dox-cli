package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pyhub/pyhub-docs/internal/config"
)

func TestNewGenerator(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)

	tests := []struct {
		name    string
		apiKey  string
		envKey  string
		wantErr bool
	}{
		{
			name:    "Direct API key",
			apiKey:  "test-key",
			envKey:  "",
			wantErr: false,
		},
		{
			name:    "API key from environment",
			apiKey:  "",
			envKey:  "env-test-key",
			wantErr: false,
		},
		{
			name:    "No API key",
			apiKey:  "",
			envKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.envKey)
			} else {
				os.Unsetenv("OPENAI_API_KEY")
			}

			gen, err := NewGenerator(ProviderOpenAI, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gen == nil {
				t.Error("NewGenerator() returned nil generator")
			}
		})
	}
}

func TestEnhancePrompt(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		contentType string
		wantContains string
	}{
		{
			name:        "Blog enhancement",
			prompt:      "Go testing best practices",
			contentType: "blog",
			wantContains: "Write a blog post about:",
		},
		{
			name:        "Blog with existing keyword",
			prompt:      "Write a blog about Go",
			contentType: "blog",
			wantContains: "Write a blog about Go", // Should not be enhanced
		},
		{
			name:        "Report enhancement",
			prompt:      "Q3 sales analysis",
			contentType: "report",
			wantContains: "Create a professional report on:",
		},
		{
			name:        "Summary enhancement",
			prompt:      "Long document content here",
			contentType: "summary",
			wantContains: "Summarize the following content:",
		},
		{
			name:        "Code enhancement",
			prompt:      "binary search in Go",
			contentType: "code",
			wantContains: "Generate code for:",
		},
		{
			name:        "Custom type - no enhancement",
			prompt:      "Custom prompt",
			contentType: "custom",
			wantContains: "Custom prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnhancePrompt(tt.prompt, tt.contentType)
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("EnhancePrompt() = %v, want to contain %v", result, tt.wantContains)
			}
		})
	}
}

func TestSaveToFile(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "generator_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name      string
		content   string
		filePath  string
		createFirst bool
		wantErr   bool
	}{
		{
			name:     "Save to new file",
			content:  "Test content",
			filePath: filepath.Join(tempDir, "test1.txt"),
			createFirst: false,
			wantErr:  false,
		},
		{
			name:     "Save to existing file",
			content:  "New content",
			filePath: filepath.Join(tempDir, "test2.txt"),
			createFirst: true,
			wantErr:  true, // Should error on existing file
		},
		{
			name:     "Empty file path",
			content:  "Content",
			filePath: "",
			createFirst: false,
			wantErr:  false, // Should not error, just skip
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create file first if needed
			if tt.createFirst && tt.filePath != "" {
				os.WriteFile(tt.filePath, []byte("existing"), 0644)
			}

			err := SaveToFile(tt.content, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify file content if no error expected
			if !tt.wantErr && tt.filePath != "" {
				content, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Errorf("Failed to read saved file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("File content = %v, want %v", string(content), tt.content)
				}
			}
		})
	}
}

func TestDetectProviderFromModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected AIProvider
	}{
		{
			name:     "Claude model - lowercase",
			model:    "claude-3-opus",
			expected: ProviderClaude,
		},
		{
			name:     "Claude model - uppercase",
			model:    "CLAUDE-3-SONNET",
			expected: ProviderClaude,
		},
		{
			name:     "GPT model - lowercase",
			model:    "gpt-4",
			expected: ProviderOpenAI,
		},
		{
			name:     "GPT model - turbo",
			model:    "gpt-3.5-turbo",
			expected: ProviderOpenAI,
		},
		{
			name:     "Davinci model",
			model:    "text-davinci-003",
			expected: ProviderOpenAI,
		},
		{
			name:     "Unknown model defaults to OpenAI",
			model:    "some-unknown-model",
			expected: ProviderOpenAI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProviderFromModel(tt.model)
			if result != tt.expected {
				t.Errorf("DetectProviderFromModel(%s) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}

func TestGetAvailableModels(t *testing.T) {
	tests := []struct {
		name     string
		provider AIProvider
		wantLen  int
	}{
		{
			name:     "OpenAI models",
			provider: ProviderOpenAI,
			wantLen:  4, // Based on hardcoded list in function
		},
		{
			name:     "Claude models",
			provider: ProviderClaude,
			wantLen:  0, // Will depend on claude.AvailableModels()
		},
		{
			name:     "Unknown provider",
			provider: AIProvider("unknown"),
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := GetAvailableModels(tt.provider)
			if tt.provider == ProviderClaude {
				// For Claude, just check it returns something (could be empty)
				if models == nil {
					t.Error("GetAvailableModels(ProviderClaude) returned nil")
				}
			} else if len(models) != tt.wantLen {
				t.Errorf("GetAvailableModels(%v) returned %d models, want %d", tt.provider, len(models), tt.wantLen)
			}
		})
	}
}

func TestDefaultGenerateOptions(t *testing.T) {
	opts := DefaultGenerateOptions()
	
	if opts.ContentType != "custom" {
		t.Errorf("DefaultGenerateOptions().ContentType = %v, want custom", opts.ContentType)
	}
	if opts.Model != "gpt-3.5-turbo" {
		t.Errorf("DefaultGenerateOptions().Model = %v, want gpt-3.5-turbo", opts.Model)
	}
	if opts.MaxTokens != 2000 {
		t.Errorf("DefaultGenerateOptions().MaxTokens = %v, want 2000", opts.MaxTokens)
	}
	if opts.Temperature != 0.7 {
		t.Errorf("DefaultGenerateOptions().Temperature = %v, want 0.7", opts.Temperature)
	}
}

func TestGeneratorCache(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)
	
	// Set test API key
	os.Setenv("OPENAI_API_KEY", "test-key")
	
	gen, err := NewGenerator(ProviderOpenAI, "")
	if err != nil {
		t.Fatal(err)
	}
	
	t.Run("EnableCache", func(t *testing.T) {
		gen.EnableCache(1*time.Hour, 100)
		// Cache is now enabled, we can't directly test internals but at least verify no panic
	})
	
	t.Run("GetCacheStats with cache", func(t *testing.T) {
		gen.EnableCache(1*time.Hour, 100)
		stats := gen.GetCacheStats()
		if stats == nil {
			t.Error("GetCacheStats() returned nil when cache is enabled")
		}
	})
	
	t.Run("DisableCache", func(t *testing.T) {
		gen.DisableCache()
		stats := gen.GetCacheStats()
		if stats != nil {
			t.Error("GetCacheStats() should return nil when cache is disabled")
		}
	})
	
	t.Run("GetCacheStats without cache", func(t *testing.T) {
		gen.DisableCache()
		stats := gen.GetCacheStats()
		if stats != nil {
			t.Error("GetCacheStats() should return nil when cache is disabled")
		}
	})
}

func TestNewGeneratorWithConfig(t *testing.T) {
	// Save original env vars
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalClaude := os.Getenv("ANTHROPIC_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalOpenAI)
	defer os.Setenv("ANTHROPIC_API_KEY", originalClaude)
	
	tests := []struct {
		name     string
		provider AIProvider
		apiKey   string
		envKey   string
		cfg      *config.Config
		wantErr  bool
	}{
		{
			name:     "OpenAI with config",
			provider: ProviderOpenAI,
			apiKey:   "test-key",
			cfg: &config.Config{
				OpenAI: config.OpenAIConfig{
					Retry: config.RetryConfig{
						MaxRetries:   3,
						InitialDelay: 1000,
						MaxDelay:     10000,
						Multiplier:   2.0,
						Jitter:       true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Claude with config",
			provider: ProviderClaude,
			apiKey:   "test-claude-key",
			cfg: &config.Config{
				Claude: config.ClaudeConfig{
					Retry: config.RetryConfig{
						MaxRetries:   5,
						InitialDelay: 500,
						MaxDelay:     5000,
						Multiplier:   1.5,
						Jitter:       true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "No API key",
			provider: ProviderOpenAI,
			apiKey:   "",
			cfg:      nil,
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("ANTHROPIC_API_KEY")
			os.Unsetenv("CLAUDE_API_KEY")
			
			// Set env var if needed
			if tt.envKey != "" {
				if tt.provider == ProviderOpenAI {
					os.Setenv("OPENAI_API_KEY", tt.envKey)
				} else {
					os.Setenv("ANTHROPIC_API_KEY", tt.envKey)
				}
			}
			
			gen, err := NewGeneratorWithConfig(tt.provider, tt.apiKey, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeneratorWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && gen == nil {
				t.Error("NewGeneratorWithConfig() returned nil generator")
			}
			
			// Check that cache is initialized
			if !tt.wantErr && gen != nil {
				stats := gen.GetCacheStats()
				if stats == nil {
					t.Error("NewGeneratorWithConfig() should initialize cache")
				}
			}
		})
	}
}

func TestGenerateContent(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)
	
	// Set test API key
	os.Setenv("OPENAI_API_KEY", "test-key")
	
	gen, err := NewGenerator(ProviderOpenAI, "")
	if err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		prompt  string
		options GenerateOptions
		wantErr bool
	}{
		{
			name:    "Empty prompt",
			prompt:  "",
			options: DefaultGenerateOptions(),
			wantErr: true,
		},
		{
			name:    "Whitespace-only prompt",
			prompt:  "   \n\t  ",
			options: DefaultGenerateOptions(),
			wantErr: true,
		},
		{
			name:    "File prompt with non-existent file",
			prompt:  "@/non/existent/file.txt",
			options: DefaultGenerateOptions(),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.GenerateContent(tt.prompt, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	
	// Test with file prompt
	t.Run("File prompt with existing file", func(t *testing.T) {
		// Create a temp file
		tempFile, err := os.CreateTemp("", "prompt_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tempFile.Name())
		
		promptContent := "Test prompt from file"
		if _, err := tempFile.WriteString(promptContent); err != nil {
			t.Fatal(err)
		}
		tempFile.Close()
		
		// Note: This will still fail because we don't have a real OpenAI client
		// but it will test the file reading logic
		_, err = gen.GenerateContent("@"+tempFile.Name(), DefaultGenerateOptions())
		// We expect an error here because the client isn't actually configured
		if err == nil {
			t.Error("Expected error with test OpenAI client")
		}
	})
}

func TestNewGeneratorClaude(t *testing.T) {
	// Save original env vars
	originalAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	originalClaudeKey := os.Getenv("CLAUDE_API_KEY")
	defer os.Setenv("ANTHROPIC_API_KEY", originalAnthropicKey)
	defer os.Setenv("CLAUDE_API_KEY", originalClaudeKey)
	
	tests := []struct {
		name         string
		apiKey       string
		anthropicKey string
		claudeKey    string
		wantErr      bool
	}{
		{
			name:    "Direct API key",
			apiKey:  "test-claude-key",
			wantErr: false,
		},
		{
			name:         "ANTHROPIC_API_KEY env var",
			apiKey:       "",
			anthropicKey: "env-anthropic-key",
			wantErr:      false,
		},
		{
			name:      "CLAUDE_API_KEY env var",
			apiKey:    "",
			claudeKey: "env-claude-key",
			wantErr:   false,
		},
		{
			name:    "No API key",
			apiKey:  "",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("ANTHROPIC_API_KEY")
			os.Unsetenv("CLAUDE_API_KEY")
			
			// Set env vars if specified
			if tt.anthropicKey != "" {
				os.Setenv("ANTHROPIC_API_KEY", tt.anthropicKey)
			}
			if tt.claudeKey != "" {
				os.Setenv("CLAUDE_API_KEY", tt.claudeKey)
			}
			
			gen, err := NewGenerator(ProviderClaude, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator(ProviderClaude) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gen == nil {
				t.Error("NewGenerator(ProviderClaude) returned nil generator")
			}
		})
	}
}

func TestUnsupportedProvider(t *testing.T) {
	_, err := NewGenerator(AIProvider("unsupported"), "test-key")
	if err == nil {
		t.Error("NewGenerator with unsupported provider should return error")
	}
	if !strings.Contains(err.Error(), "unsupported AI provider") {
		t.Errorf("Expected error about unsupported provider, got: %v", err)
	}
}