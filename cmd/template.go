package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-documents-cli/internal/template"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	templatePath string
	valuesFile   string
	setValues    []string
	templateOut  string
	templateForce bool
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Process documents using templates",
	Long: `Process Word or PowerPoint documents using templates with placeholders.

Placeholders use the {{placeholder_name}} format and can be replaced with values
provided via command-line flags or from a YAML/JSON file.

Examples:
  # Process template with inline values
  pyhub-documents-cli template --template report.docx --output final.docx --set title="Q4 Report" --set year="2024"

  # Process template with values from YAML file
  pyhub-documents-cli template --template presentation.pptx --values data.yaml --output final.pptx

  # Force overwrite existing file
  pyhub-documents-cli template --template template.docx --values values.yaml --output output.docx --force

Values file format (YAML):
  title: "Annual Report"
  author: "John Doe"
  year: 2024
  items:
    - "Achievement 1"
    - "Achievement 2"`,
	RunE: runTemplate,
}

func init() {
	rootCmd.AddCommand(templateCmd)

	templateCmd.Flags().StringVarP(&templatePath, "template", "t", "", "Template file path (required)")
	templateCmd.Flags().StringVar(&valuesFile, "values", "", "Values file (YAML or JSON)")
	templateCmd.Flags().StringArrayVar(&setValues, "set", []string{}, "Set individual values (format: key=value)")
	templateCmd.Flags().StringVarP(&templateOut, "output", "o", "", "Output file path (required)")
	templateCmd.Flags().BoolVar(&templateForce, "force", false, "Overwrite existing output file")

	templateCmd.MarkFlagRequired("template")
	templateCmd.MarkFlagRequired("output")
}

func runTemplate(cmd *cobra.Command, args []string) error {
	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// Check if output file exists and force flag is not set
	if !templateForce {
		if _, err := os.Stat(templateOut); err == nil {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", templateOut)
		}
	}

	// Load values
	values := make(map[string]interface{})

	// Load values from file if provided
	if valuesFile != "" {
		fileValues, err := loadValuesFromFile(valuesFile)
		if err != nil {
			return fmt.Errorf("failed to load values from file: %w", err)
		}
		// Merge file values
		for k, v := range fileValues {
			values[k] = v
		}
	}

	// Parse and apply --set values (these override file values)
	for _, setValue := range setValues {
		parts := strings.SplitN(setValue, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set format: %s (expected key=value)", setValue)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Try to parse value as number or boolean
		if v, err := parseValue(value); err == nil {
			values[key] = v
		} else {
			values[key] = value
		}
	}

	// Determine document type from template extension
	ext := strings.ToLower(filepath.Ext(templatePath))
	
	switch ext {
	case ".docx":
		processor := template.NewWordProcessor()
		
		// Validate template
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err != nil {
			return fmt.Errorf("failed to validate template: %w", err)
		}
		
		if len(missing) > 0 {
			cmd.PrintErrf("Warning: The following placeholders have no values: %v\n", missing)
		}
		
		// Process template
		cmd.Printf("Processing Word template...\n")
		if err := processor.ProcessTemplate(templatePath, values, templateOut); err != nil {
			return fmt.Errorf("failed to process template: %w", err)
		}
		
	case ".pptx":
		processor := template.NewPowerPointProcessor()
		
		// Validate template
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err != nil {
			return fmt.Errorf("failed to validate template: %w", err)
		}
		
		if len(missing) > 0 {
			cmd.PrintErrf("Warning: The following placeholders have no values: %v\n", missing)
		}
		
		// Process template
		cmd.Printf("Processing PowerPoint template...\n")
		if err := processor.ProcessTemplate(templatePath, values, templateOut); err != nil {
			return fmt.Errorf("failed to process template: %w", err)
		}
		
	default:
		return fmt.Errorf("unsupported template format: %s (only .docx and .pptx are supported)", ext)
	}

	cmd.Printf("✅ Successfully created %s from template\n", templateOut)
	return nil
}

// loadValuesFromFile loads values from a YAML or JSON file
func loadValuesFromFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	values := make(map[string]interface{})

	// Try to determine format from extension
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &values); err != nil {
			if err := json.Unmarshal(data, &values); err != nil {
				return nil, fmt.Errorf("failed to parse file as YAML or JSON")
			}
		}
	}

	return values, nil
}

// parseValue tries to parse a string value as a number or boolean
func parseValue(s string) (interface{}, error) {
	// Try boolean
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Try integer
	var intVal int
	if _, err := fmt.Sscanf(s, "%d", &intVal); err == nil {
		return intVal, nil
	}

	// Try float
	var floatVal float64
	if _, err := fmt.Sscanf(s, "%f", &floatVal); err == nil {
		return floatVal, nil
	}

	return nil, fmt.Errorf("not a special value")
}