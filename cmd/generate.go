package cmd

import (
	"errors"
	"fmt"
	"os"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/generate"
	"github.com/pyhub/pyhub-docs/internal/openai"
	"github.com/spf13/cobra"
)

var (
	contentType  string
	prompt       string
	genOutput    string
	model        string
	maxTokens    int
	temperature  float64
	apiKey       string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate content using AI (OpenAI)",
	Long: `Generate various types of content using OpenAI's language models.

Content types:
  • blog: Blog posts and articles
  • report: Business reports and summaries
  • summary: Document summarization
  • custom: Custom content with your prompt

Examples:
  # Generate a blog post
  pyhub-documents-cli generate --type blog --prompt "Best practices for Go testing" --output blog.md

  # Generate a report
  pyhub-documents-cli generate --type report --prompt "Q3 sales analysis" --output report.md

  # Summarize a document
  pyhub-documents-cli generate --type summary --prompt "$(cat long-document.md)" --output summary.md

  # Use GPT-4 for complex content
  pyhub-documents-cli generate --type blog --prompt "Advanced Go patterns" --model gpt-4 --output article.md`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&contentType, "type", "t", "custom", "Content type (blog|report|summary|custom)")
	generateCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Generation prompt or file containing prompt (required)")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "Output file path")
	generateCmd.Flags().StringVar(&model, "model", "gpt-3.5-turbo", "AI model to use")
	generateCmd.Flags().IntVar(&maxTokens, "max-tokens", 2000, "Maximum tokens for response")
	generateCmd.Flags().Float64Var(&temperature, "temperature", 0.7, "Creativity level (0.0-1.0)")
	generateCmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI API key (or use OPENAI_API_KEY env var)")

	generateCmd.MarkFlagRequired("prompt")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Validate inputs
	if prompt == "" {
		return pkgErrors.NewValidationError("prompt", prompt, "prompt is required")
	}

	// Validate content type
	validTypes := []string{"blog", "report", "summary", "code", "custom"}
	isValid := false
	for _, t := range validTypes {
		if contentType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return pkgErrors.NewValidationError("type", contentType, "must be one of: blog, report, summary, code, custom")
	}

	// Check if output file exists and force flag is not set
	if genOutput != "" && !force {
		if _, err := os.Stat(genOutput); err == nil {
			return pkgErrors.NewFileError(genOutput, "creating", fmt.Errorf("%w: use --force to overwrite", pkgErrors.ErrFileAlreadyExists))
		}
	}

	// Create generator with API key
	if verbose {
		fmt.Println("Initializing OpenAI client...")
	}
	
	generator, err := generate.NewGenerator(apiKey)
	if err != nil {
		if errors.Is(err, pkgErrors.ErrMissingAPIKey) {
			return fmt.Errorf("OpenAI API key is required. Set OPENAI_API_KEY environment variable or use --api-key flag")
		}
		return fmt.Errorf("failed to initialize generator: %w", err)
	}

	// Enhance prompt based on content type
	enhancedPrompt := generate.EnhancePrompt(prompt, contentType)
	
	if verbose {
		fmt.Printf("Generating %s content with model %s...\n", contentType, model)
		fmt.Printf("Temperature: %.2f, Max tokens: %d\n", temperature, maxTokens)
	}

	// Set generation options
	options := openai.GenerateOptions{
		ContentType: contentType,
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	// Generate content
	if !quiet {
		fmt.Printf("Generating %s content...\n", contentType)
	}
	
	content, err := generator.GenerateContent(enhancedPrompt, options)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Save to file if specified
	if genOutput != "" {
		// Check for force flag override for existing files
		if force {
			// Delete existing file first
			os.Remove(genOutput)
		}
		
		err = generate.SaveToFile(content, genOutput)
		if err != nil {
			if !errors.Is(err, pkgErrors.ErrFileAlreadyExists) {
				return err
			}
			// File exists, print to stdout instead
			fmt.Println("\n--- Generated Content ---")
			fmt.Println(content)
			fmt.Println("--- End of Content ---")
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", genOutput)
		}
		
		if !quiet {
			fmt.Printf("Content saved to: %s\n", genOutput)
		}
	} else {
		// Print to stdout if no output file specified
		fmt.Println("\n--- Generated Content ---")
		fmt.Println(content)
		fmt.Println("--- End of Content ---")
	}

	if verbose {
		fmt.Println("Generation completed successfully!")
	}

	return nil
}