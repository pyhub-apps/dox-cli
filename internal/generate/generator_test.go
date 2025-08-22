package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

			gen, err := NewGenerator(tt.apiKey)
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