package generate

import (
	"fmt"
	"os"
	"strings"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/openai"
)

// Generator handles content generation using AI
type Generator struct {
	client *openai.Client
}

// NewGenerator creates a new content generator
func NewGenerator(apiKey string) (*Generator, error) {
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

	return &Generator{
		client: client,
	}, nil
}

// GenerateContent generates content based on the provided options
func (g *Generator) GenerateContent(prompt string, options openai.GenerateOptions) (string, error) {
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

	// Generate content using OpenAI client
	content, err := g.client.GenerateContent(prompt, options)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return content, nil
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
	case "code":
		if !strings.Contains(strings.ToLower(prompt), "code") && !strings.Contains(strings.ToLower(prompt), "function") {
			return fmt.Sprintf("Generate code for: %s\n\nInclude proper error handling, comments, and follow best practices.", prompt)
		}
	}
	return prompt
}