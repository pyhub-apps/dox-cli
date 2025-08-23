package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pkgErrors "github.com/pyhub/pyhub-docs/internal/errors"
	"github.com/pyhub/pyhub-docs/internal/i18n"
	"github.com/pyhub/pyhub-docs/internal/markdown"
	"github.com/pyhub/pyhub-docs/internal/ui"
	"github.com/spf13/cobra"
)

var (
	fromFile     string
	templateFile string
	outputFile   string
	format       string
	force        bool
	createDryRun bool
	createJsonOutput bool
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create documents from markdown files",
	Long: `Create Word or PowerPoint documents from markdown files.

Supports:
  • Direct markdown to document conversion
  • Template-based document generation
  • Style and format preservation

Examples:
  # Create Word document from markdown
  dox create --from report.md --output report.docx

  # Use a template for styling
  dox create --from content.md --template company-template.docx --output final.docx

  # Create PowerPoint presentation
  dox create --from slides.md --output presentation.pptx --format pptx

  # Force overwrite existing file
  dox create --from report.md --output report.docx --force`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&fromFile, "from", "f", "", "Input markdown file (required)")
	createCmd.Flags().StringVarP(&templateFile, "template", "t", "", "Template document file")
	createCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	createCmd.Flags().StringVar(&format, "format", "", "Output format (docx|pptx, auto-detected from extension)")
	createCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output file")
	createCmd.Flags().BoolVar(&createDryRun, "dry-run", false, "Preview operation without creating files")
	createCmd.Flags().BoolVar(&createJsonOutput, "json", false, "Output in JSON format")

	createCmd.MarkFlagRequired("from")
	createCmd.MarkFlagRequired("output")
	
	// Update descriptions after i18n initialization
	cobra.OnInitialize(func() {
		createCmd.Short = i18n.T(i18n.MsgCmdCreateShort)
		createCmd.Long = i18n.T(i18n.MsgCmdCreateLong)
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Validate inputs
	if fromFile == "" {
		return pkgErrors.NewValidationError("from", fromFile, "input file is required")
	}
	if outputFile == "" {
		return pkgErrors.NewValidationError("output", outputFile, "output file is required")
	}

	// Check if input file exists
	if _, err := os.Stat(fromFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return pkgErrors.NewFileError(fromFile, "reading input", pkgErrors.ErrFileNotFound)
		}
		if errors.Is(err, os.ErrPermission) {
			return pkgErrors.NewFileError(fromFile, "reading input", pkgErrors.ErrPermissionDenied)
		}
		return pkgErrors.NewFileError(fromFile, "reading input", err)
	}

	// Check if output file exists and force flag is not set
	if !force {
		if _, err := os.Stat(outputFile); err == nil {
			return pkgErrors.NewFileError(outputFile, "creating", fmt.Errorf("%w: use --force to overwrite", pkgErrors.ErrFileAlreadyExists))
		}
	}

	// Determine output format
	outputFormat := format
	if outputFormat == "" {
		// Auto-detect from file extension
		ext := strings.ToLower(filepath.Ext(outputFile))
		switch ext {
		case ".docx":
			outputFormat = "docx"
		case ".pptx":
			outputFormat = "pptx"
		default:
			return pkgErrors.NewDocumentError(outputFile, ext, "unsupported format (use .docx or .pptx)", pkgErrors.ErrUnsupportedFormat)
		}
	}

	// Validate format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "docx" && outputFormat != "pptx" {
		return pkgErrors.NewValidationError("format", outputFormat, "must be 'docx' or 'pptx'")
	}

	// Handle dry-run mode
	if createDryRun {
		// Get file information
		inputInfo, err := os.Stat(fromFile)
		if err != nil {
			return fmt.Errorf("failed to get input file info: %w", err)
		}
		
		outputExists := false
		var outputInfo os.FileInfo
		if info, err := os.Stat(outputFile); err == nil {
			outputExists = true
			outputInfo = info
		}
		
		// Prepare dry-run information
		if createJsonOutput {
			// JSON output for dry-run
			dryRunInfo := map[string]interface{}{
				"operation": "create",
				"input": map[string]interface{}{
					"path":   fromFile,
					"size":   inputInfo.Size(),
					"format": "markdown",
				},
				"output": map[string]interface{}{
					"path":   outputFile,
					"format": outputFormat,
					"exists": outputExists,
				},
			}
			
			if templateFile != "" {
				dryRunInfo["template"] = templateFile
			}
			
			if outputExists {
				dryRunInfo["output"].(map[string]interface{})["currentSize"] = outputInfo.Size()
				if !force {
					dryRunInfo["warning"] = "Output file exists. Use --force to overwrite"
				}
			}
			
			jsonBytes, _ := json.MarshalIndent(dryRunInfo, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			// Human-readable output for dry-run
			fmt.Println("=== DRY-RUN MODE ===")
			fmt.Println()
			fmt.Printf("Input:    %s (%.2f KB)\n", fromFile, float64(inputInfo.Size())/1024)
			fmt.Printf("Output:   %s (%s format)\n", outputFile, strings.ToUpper(outputFormat))
			
			if templateFile != "" {
				fmt.Printf("Template: %s\n", templateFile)
			}
			
			fmt.Println()
			
			if outputExists {
				fmt.Printf("⚠️  Output file exists (%.2f KB)\n", float64(outputInfo.Size())/1024)
				if !force {
					fmt.Println("   Use --force flag to overwrite")
				} else {
					fmt.Println("   Will be overwritten (--force is set)")
				}
			} else {
				fmt.Println("✓ Output file will be created")
			}
			
			fmt.Println()
			fmt.Println("No files were created. Remove --dry-run to execute.")
		}
		
		return nil
	}

	// Check if template is specified
	if templateFile != "" {
		if verbose {
			ui.PrintInfo("Using template: %s", templateFile)
		}
	}

	// Create appropriate converter
	var converter markdown.Converter
	switch outputFormat {
	case "docx":
		converter = markdown.NewWordConverter()
		if !quiet {
			ui.PrintInfo("Converting %s to Word document...", fromFile)
		}
	case "pptx":
		converter = markdown.NewPowerPointConverter()
		if !quiet {
			ui.PrintInfo("Converting %s to PowerPoint presentation...", fromFile)
		}
	}

	// Perform conversion with spinner
	var spinner *ui.ProgressBar
	if !quiet {
		spinner = ui.NewSpinner("Processing...")
	}
	
	if err := markdown.ConvertFile(fromFile, converter, outputFile); err != nil {
		if spinner != nil {
			spinner.Clear()
		}
		ui.PrintError("Conversion failed: %v", err)
		return fmt.Errorf("%s", i18n.T(i18n.MsgErrorConversion, map[string]interface{}{
			"Error": err.Error(),
		}))
	}
	
	if spinner != nil {
		spinner.Finish()
	}
	
	ui.PrintSuccess("Document created: %s", outputFile)
	return nil
}