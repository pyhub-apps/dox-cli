package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/export"
	"github.com/pyhub/pyhub-docs/internal/pdf"
	"github.com/spf13/cobra"
)

var (
	extractFormat     string
	extractOutput     string
	extractDebug      bool
	extractStrict     bool
	extractMinQuality float64
	extractIgnoreQual bool
)

var extractCmd = &cobra.Command{
	Use:   "extract [pdf-file]",
	Short: "Extract content from PDF documents",
	Long: `Extract structured content from PDF documents including text, tables, and layout.
	
Preserves document structure including:
  • Headings and paragraphs
  • Tables with proper formatting
  • Lists and hierarchical content
  • Metadata (title, author, etc.)

Supports export to HTML and Markdown formats.`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func init() {
	rootCmd.AddCommand(extractCmd)

	extractCmd.Flags().StringVarP(&extractFormat, "format", "f", "markdown", "Output format (html|markdown)")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "Output file path (default: stdout)")
	extractCmd.Flags().BoolVarP(&extractDebug, "debug", "d", false, "Enable debug output")
	extractCmd.Flags().BoolVarP(&extractStrict, "strict", "s", false, "Strict quality mode - fail on low quality")
	extractCmd.Flags().Float64VarP(&extractMinQuality, "min-quality", "m", 0.2, "Minimum quality threshold (0.0-1.0)")
	extractCmd.Flags().BoolVar(&extractIgnoreQual, "ignore-quality", false, "Ignore quality checks and force extraction")
}

func runExtract(cmd *cobra.Command, args []string) error {
	pdfPath := args[0]

	// Verify PDF file exists
	if _, err := os.Stat(pdfPath); err != nil {
		return fmt.Errorf("PDF file not found: %s", pdfPath)
	}

	// Create extractor with options
	options := pdf.ExtractorOptions{
		Debug:         extractDebug,
		Strict:        extractStrict,
		MinQuality:    extractMinQuality,
		IgnoreQuality: extractIgnoreQual,
	}
	
	extractor, err := pdf.NewExtractor(options)
	if err != nil {
		// Check if it's a dependency issue
		if strings.Contains(err.Error(), "Python not found") {
			fmt.Fprintln(os.Stderr, "Error: Python 3 is required for PDF extraction")
			fmt.Fprintln(os.Stderr, "Please install Python 3 from https://www.python.org/")
			return err
		}
		if strings.Contains(err.Error(), "script not found") {
			fmt.Fprintln(os.Stderr, "Error: PDF extraction script not found")
			fmt.Fprintln(os.Stderr, "Please ensure scripts/pdf_extract.py exists")
			return err
		}
		return err
	}

	// Check dependencies
	if err := extractor.CheckDependencies(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Missing dependencies")
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "\nTo install required Python libraries:")
		fmt.Fprintln(os.Stderr, "  pip install pdfplumber")
		return err
	}

	// Extract PDF content
	if extractDebug {
		fmt.Fprintf(os.Stderr, "Extracting content from: %s\n", pdfPath)
	}

	result, err := extractor.Extract(pdfPath)
	if err != nil {
		// Check if it's a quality issue
		if strings.Contains(err.Error(), "quality") {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "❌ PDF Quality Issue Detected")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "The PDF appears to have tables with mostly empty cells,")
			fmt.Fprintln(os.Stderr, "which suggests the content may be in image format rather than text.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Possible solutions:")
			fmt.Fprintln(os.Stderr, "  1. Re-generate the PDF from the original document using 'Export to PDF'")
			fmt.Fprintln(os.Stderr, "  2. Use --ignore-quality flag to force extraction (may produce poor results)")
			fmt.Fprintln(os.Stderr, "  3. Use OCR tools if the content is in image format")
			fmt.Fprintln(os.Stderr, "")
		}
		return fmt.Errorf("extraction failed: %w", err)
	}

	if extractDebug {
		fmt.Fprintf(os.Stderr, "Extracted %d pages\n", len(result.Pages))
		for _, page := range result.Pages {
			if len(page.Tables) > 0 {
				fmt.Fprintf(os.Stderr, "  Page %d: %d tables found\n", page.Number, len(page.Tables))
			}
		}
	}

	// Convert to desired format
	converter := export.NewConverter(result)
	
	var format export.Format
	switch strings.ToLower(extractFormat) {
	case "html":
		format = export.FormatHTML
	case "markdown", "md":
		format = export.FormatMarkdown
	default:
		return fmt.Errorf("unsupported format: %s (use 'html' or 'markdown')", extractFormat)
	}

	output, err := converter.Convert(format)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Write output
	if extractOutput == "" {
		// Write to stdout
		fmt.Print(output)
	} else {
		// Ensure output directory exists
		outputDir := filepath.Dir(extractOutput)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write to file
		if err := os.WriteFile(extractOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✅ Successfully extracted to: %s\n", extractOutput)
	}

	return nil
}