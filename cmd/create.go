package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-documents-cli/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	fromFile     string
	templateFile string
	outputFile   string
	format       string
	force        bool
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
  pyhub-documents-cli create --from report.md --output report.docx

  # Use a template for styling
  pyhub-documents-cli create --from content.md --template company-template.docx --output final.docx

  # Create PowerPoint presentation
  pyhub-documents-cli create --from slides.md --output presentation.pptx --format pptx

  # Force overwrite existing file
  pyhub-documents-cli create --from report.md --output report.docx --force`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&fromFile, "from", "f", "", "Input markdown file (required)")
	createCmd.Flags().StringVarP(&templateFile, "template", "t", "", "Template document file")
	createCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (required)")
	createCmd.Flags().StringVar(&format, "format", "", "Output format (docx|pptx, auto-detected from extension)")
	createCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output file")

	createCmd.MarkFlagRequired("from")
	createCmd.MarkFlagRequired("output")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Check if input file exists
	if _, err := os.Stat(fromFile); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", fromFile)
	}

	// Check if output file exists and force flag is not set
	if !force {
		if _, err := os.Stat(outputFile); err == nil {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputFile)
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
			return fmt.Errorf("cannot determine output format from extension %s, please specify --format", ext)
		}
	}

	// Validate format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "docx" && outputFormat != "pptx" {
		return fmt.Errorf("unsupported format: %s (supported: docx, pptx)", outputFormat)
	}

	// Check if template is specified (not yet implemented)
	if templateFile != "" {
		cmd.PrintErrf("Warning: Template support is not yet implemented, ignoring --template flag\n")
	}

	// Create appropriate converter
	var converter markdown.Converter
	switch outputFormat {
	case "docx":
		converter = markdown.NewWordConverter()
		cmd.Printf("Converting %s to Word document...\n", fromFile)
	case "pptx":
		converter = markdown.NewPowerPointConverter()
		cmd.Printf("Converting %s to PowerPoint presentation...\n", fromFile)
	}

	// Perform conversion
	if err := markdown.ConvertFile(fromFile, converter, outputFile); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	cmd.Printf("✅ Successfully created %s\n", outputFile)
	return nil
}