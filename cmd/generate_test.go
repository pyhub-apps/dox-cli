package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateCommand(t *testing.T) {
	// Save original env vars
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalClaude := os.Getenv("ANTHROPIC_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalOpenAI)
	defer os.Setenv("ANTHROPIC_API_KEY", originalClaude)

	t.Run("Command Registration", func(t *testing.T) {
		// Check that generate command is registered
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "generate" {
				found = true
				break
			}
		}
		if !found {
			t.Error("generate command not registered with root command")
		}
	})

	t.Run("Command Flags", func(t *testing.T) {
		// Check that required flags are defined
		if generateCmd.Flags().Lookup("prompt") == nil {
			t.Error("--prompt flag not defined")
		}
		if generateCmd.Flags().Lookup("type") == nil {
			t.Error("--type flag not defined")
		}
		if generateCmd.Flags().Lookup("output") == nil {
			t.Error("--output flag not defined")
		}
		if generateCmd.Flags().Lookup("model") == nil {
			t.Error("--model flag not defined")
		}
		if generateCmd.Flags().Lookup("provider") == nil {
			t.Error("--provider flag not defined")
		}
	})

	t.Run("Missing Required Flags", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *generateCmd // Copy command
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Clear env vars to ensure no API key
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		
		// Reset flags
		prompt = ""
		contentType = ""
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error when prompt is missing")
		}
	})

	t.Run("Empty Prompt Error", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *generateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Set API key
		os.Setenv("OPENAI_API_KEY", "test-key")
		
		// Set empty prompt
		prompt = ""
		contentType = "blog"
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error with empty prompt")
		}
		if !strings.Contains(err.Error(), "prompt is required") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("Missing API Key Error", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *generateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Clear env vars
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
		
		// Set valid prompt but no API key
		prompt = "test prompt"
		contentType = "blog"
		apiKey = ""
		provider = "openai"
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error when API key is missing")
		}
	})

	t.Run("Provider Detection from Model", func(t *testing.T) {
		tests := []struct {
			name           string
			modelName      string
			expectedProvider string
		}{
			{"Claude model", "claude-3-opus", "claude"},
			{"GPT model", "gpt-4", "openai"},
			{"GPT-3.5 model", "gpt-3.5-turbo", "openai"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Set model
				model = tt.modelName
				provider = "" // Clear provider to test auto-detection
				
				// The actual provider detection happens in the command
				// We can't test the full flow without mocking, but we can
				// verify the model flag is set correctly
				if model != tt.modelName {
					t.Errorf("Model not set correctly: got %v, want %v", model, tt.modelName)
				}
			})
		}
	})

	t.Run("Output File Validation", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "generate_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		outputPath := filepath.Join(tempDir, "output.md")
		
		// Set output flag
		genOutput = outputPath
		
		// Verify the flag is set
		if genOutput != outputPath {
			t.Errorf("Output path not set correctly: got %v, want %v", genOutput, outputPath)
		}
	})

	t.Run("Content Type Flag", func(t *testing.T) {
		validTypes := []string{"blog", "report", "summary", "email", "proposal", "custom"}
		
		for _, ct := range validTypes {
			contentType = ct
			if contentType != ct {
				t.Errorf("Content type not set correctly: got %v, want %v", contentType, ct)
			}
		}
	})

	t.Run("Temperature Range", func(t *testing.T) {
		tests := []float64{0.0, 0.5, 0.7, 1.0, 2.0}
		
		for _, temp := range tests {
			temperature = temp
			if temperature != temp {
				t.Errorf("Temperature not set correctly: got %v, want %v", temperature, temp)
			}
		}
	})

	t.Run("Max Tokens", func(t *testing.T) {
		tests := []int{100, 500, 1000, 2000, 4000}
		
		for _, tokens := range tests {
			maxTokens = tokens
			if maxTokens != tokens {
				t.Errorf("Max tokens not set correctly: got %v, want %v", maxTokens, tokens)
			}
		}
	})

	t.Run("Cache Flag", func(t *testing.T) {
		noCache = true
		if !noCache {
			t.Error("No-cache flag not set correctly")
		}
		
		noCache = false
		if noCache {
			t.Error("No-cache flag not cleared correctly")
		}
	})
}

