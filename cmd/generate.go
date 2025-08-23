package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/generate"
	"github.com/pyhub/pyhub-docs/internal/ui"
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
	provider     string
	claudeAPIKey string
	noCache      bool
	dryRun       bool
	jsonOutput   bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate content using AI (OpenAI or Claude)",
	Long: `Generate various types of content using AI language models.

Supported providers:
  • OpenAI (GPT-3.5, GPT-4)
  • Claude (Claude 3 Opus, Sonnet, Haiku)

Content types:
  • blog: Blog posts and articles
  • report: Business reports and summaries
  • summary: Document summarization
  • email: Professional emails
  • proposal: Business proposals
  • custom: Custom content with your prompt

Examples:
  # Generate a blog post with OpenAI
  dox generate --type blog --prompt "Best practices for Go testing" --output blog.md

  # Use Claude for a report
  dox generate --type report --prompt "Q3 sales analysis" --model claude-3-sonnet --output report.md

  # Use specific Claude model
  dox generate --provider claude --model claude-3-opus-20240229 --prompt "Complex analysis" --output analysis.md

  # Summarize a document
  dox generate --type summary --prompt "$(cat long-document.md)" --output summary.md

  # Use GPT-4 for complex content
  dox generate --type blog --prompt "Advanced Go patterns" --model gpt-4 --output article.md`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&contentType, "type", "t", "custom", "Content type (blog|report|summary|email|proposal|custom)")
	generateCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Generation prompt or file containing prompt (required)")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "Output file path")
	generateCmd.Flags().StringVar(&model, "model", "", "AI model to use (auto-detect from name)")
	generateCmd.Flags().IntVar(&maxTokens, "max-tokens", 2000, "Maximum tokens for response")
	generateCmd.Flags().Float64Var(&temperature, "temperature", 0.7, "Creativity level (0.0-2.0)")
	generateCmd.Flags().StringVar(&provider, "provider", "", "AI provider (openai|claude, auto-detect if not specified)")
	generateCmd.Flags().StringVar(&apiKey, "api-key", "", "API key (or use environment variables)")
	generateCmd.Flags().StringVar(&claudeAPIKey, "claude-api-key", "", "Claude API key (or use ANTHROPIC_API_KEY env var)")
	generateCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching of AI responses")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without making API calls")
	generateCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	generateCmd.MarkFlagRequired("prompt")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Auto-detect provider from model name if not specified
	if provider == "" && model != "" {
		provider = string(generate.DetectProviderFromModel(model))
	}
	
	// Default to OpenAI if still not specified
	if provider == "" {
		provider = "openai"
	}
	
	// Set default model based on provider
	if model == "" {
		switch provider {
		case "claude":
			model = "claude-3-sonnet-20240229"
		default:
			model = "gpt-3.5-turbo"
		}
	}
	
	// 설정 파일의 기본값 적용 (CLI 플래그가 우선)
	if appConfig != nil {
		// OpenAI API 키
		if provider == "openai" && apiKey == "" && appConfig.OpenAI.APIKey != "" {
			apiKey = appConfig.OpenAI.APIKey
		}
		
		// Claude API 키 (설정 파일에서 읽기)
		if provider == "claude" && claudeAPIKey == "" && appConfig.Claude.APIKey != "" {
			claudeAPIKey = appConfig.Claude.APIKey
		}
		
		// 다른 설정들: CLI 플래그가 설정되지 않은 경우 설정 파일 사용
		if !cmd.Flags().Changed("model") && appConfig.Generate.Model != "" {
			model = appConfig.Generate.Model
			// Re-detect provider from configured model
			if !cmd.Flags().Changed("provider") {
				provider = string(generate.DetectProviderFromModel(model))
			}
		}
		if !cmd.Flags().Changed("max-tokens") && appConfig.Generate.MaxTokens > 0 {
			maxTokens = appConfig.Generate.MaxTokens
		}
		if !cmd.Flags().Changed("temperature") {
			temperature = appConfig.Generate.Temperature
		}
		if !cmd.Flags().Changed("type") && appConfig.Generate.ContentType != "" {
			contentType = appConfig.Generate.ContentType
		}
	}
	
	// Select appropriate API key based on provider
	var selectedAPIKey string
	switch provider {
	case "claude":
		selectedAPIKey = claudeAPIKey
		if selectedAPIKey == "" {
			selectedAPIKey = apiKey // Try generic api-key flag
		}
	default:
		selectedAPIKey = apiKey
	}
	
	// Validate inputs
	if prompt == "" {
		return pkgErrors.NewValidationError("prompt", prompt, "prompt is required")
	}

	// Validate content type
	validTypes := []string{"blog", "report", "summary", "email", "proposal", "code", "custom"}
	isValid := false
	for _, t := range validTypes {
		if contentType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return pkgErrors.NewValidationError("type", contentType, "must be one of: blog, report, summary, email, proposal, code, custom")
	}

	// Check if output file exists and force flag is not set
	if genOutput != "" && !force {
		if _, err := os.Stat(genOutput); err == nil {
			return pkgErrors.NewFileError(genOutput, "creating", fmt.Errorf("%w: use --force to overwrite", pkgErrors.ErrFileAlreadyExists))
		}
	}

	// Create generator with API key and config
	if verbose {
		ui.PrintInfo("Initializing %s client...", strings.Title(provider))
		if !noCache {
			ui.PrintInfo("Cache enabled for AI responses")
		}
	}
	
	generator, err := generate.NewGeneratorWithConfig(generate.AIProvider(provider), selectedAPIKey, appConfig)
	if err != nil {
		if errors.Is(err, pkgErrors.ErrMissingAPIKey) {
			// Use new coded error with localized message and solution
			return pkgErrors.NewAPIKeyNotFoundError(provider)
		}
		return fmt.Errorf("failed to initialize generator: %w", err)
	}
	
	// Disable cache if requested
	if noCache {
		generator.DisableCache()
	}

	// Enhance prompt based on content type
	enhancedPrompt := generate.EnhancePrompt(prompt, contentType)
	
	// Handle dry-run mode
	if dryRun {
		// Create token estimator
		estimator := generate.NewTokenEstimator(model)
		
		// Estimate tokens
		promptTokens := estimator.EstimateTokens(enhancedPrompt)
		completionTokens := maxTokens // Use max tokens as estimate for completion
		
		// Calculate cost
		cost, currency := estimator.EstimateCost(promptTokens, completionTokens)
		
		// Get model info
		modelInfo := estimator.GetModelInfo()
		
		// Check if prompt fits in context window
		if promptTokens > modelInfo.ContextWindow {
			ui.PrintWarning("Prompt exceeds model's context window (%d > %d tokens)", 
				promptTokens, modelInfo.ContextWindow)
		}
		
		if jsonOutput {
			// JSON output for dry-run
			dryRunInfo := map[string]interface{}{
				"operation": "generate",
				"provider":  provider,
				"model":     model,
				"contentType": contentType,
				"temperature": temperature,
				"maxTokens":   maxTokens,
				"estimatedTokens": map[string]int{
					"prompt":     promptTokens,
					"completion": completionTokens,
					"total":      promptTokens + completionTokens,
				},
				"estimatedCost": map[string]interface{}{
					"amount":   cost,
					"currency": currency,
				},
				"modelInfo": map[string]int{
					"contextWindow": modelInfo.ContextWindow,
					"maxOutput":     modelInfo.MaxOutput,
				},
				"outputFile": genOutput,
			}
			
			jsonBytes, _ := json.MarshalIndent(dryRunInfo, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			// Human-readable output for dry-run
			ui.PrintInfo("=== DRY-RUN MODE ===")
			ui.PrintInfo("")
			ui.PrintInfo("Operation: Generate %s content", contentType)
			ui.PrintInfo("Provider:  %s", provider)
			ui.PrintInfo("")
			
			fmt.Println(generate.FormatModelInfo(modelInfo))
			fmt.Println("")
			fmt.Println(generate.FormatCostEstimate(promptTokens, completionTokens, cost, currency))
			
			if genOutput != "" {
				ui.PrintInfo("")
				ui.PrintInfo("Output will be saved to: %s", genOutput)
			}
			
			ui.PrintInfo("")
			ui.PrintInfo("No API calls were made. Remove --dry-run to execute.")
		}
		
		return nil
	}
	
	if verbose {
		ui.PrintInfo("Generating %s content with %s model %s...", contentType, provider, model)
		ui.PrintInfo("Temperature: %.2f, Max tokens: %d", temperature, maxTokens)
	}

	// Set generation options (provider-agnostic)
	options := generate.GenerateOptions{
		ContentType: contentType,
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	// Generate content
	if !quiet {
		spinner := ui.NewSpinner(fmt.Sprintf("Generating %s content with %s...", contentType, provider))
		defer spinner.Finish()
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
			ui.PrintSuccess("Content saved to: %s", genOutput)
		}
	} else {
		// Print to stdout if no output file specified
		fmt.Println("\n--- Generated Content ---")
		fmt.Println(content)
		fmt.Println("--- End of Content ---")
	}

	if verbose {
		ui.PrintSuccess("Generation completed successfully!")
		
		// Show cache statistics if cache is enabled
		if !noCache {
			if stats := generator.GetCacheStats(); stats != nil {
				ui.PrintInfo("Cache stats: Hits=%d, Misses=%d, Hit Rate=%.1f%%", 
					stats.Hits, stats.Misses, stats.HitRate())
			}
		}
	}

	return nil
}