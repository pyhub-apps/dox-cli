package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/i18n"
	"github.com/pyhub/pyhub-docs/internal/template"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	templatePath string
	valuesFile   string
	setValues    []string
	templateOut  string
	templateForce bool
	templateDryRun bool
	templateJsonOutput bool
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
  dox template --template report.docx --output final.docx --set title="Q4 Report" --set year="2024"

  # Process template with values from YAML file
  dox template --template presentation.pptx --values data.yaml --output final.pptx

  # Force overwrite existing file
  dox template --template template.docx --values values.yaml --output output.docx --force

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
	templateCmd.Flags().BoolVar(&templateDryRun, "dry-run", false, "Preview operation without creating files")
	templateCmd.Flags().BoolVar(&templateJsonOutput, "json", false, "Output in JSON format")

	templateCmd.MarkFlagRequired("template")
	templateCmd.MarkFlagRequired("output")
	
	// Update descriptions after i18n initialization
	cobra.OnInitialize(func() {
		templateCmd.Short = i18n.T(i18n.MsgCmdTemplateShort)
		templateCmd.Long = i18n.T(i18n.MsgCmdTemplateLong)
	})
}

func runTemplate(cmd *cobra.Command, args []string) error {
	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("%s", i18n.T(i18n.MsgErrorFileNotFound, map[string]interface{}{
			"Type": "Template",
			"Path": templatePath,
		}))
	}

	// Check if output file exists and force flag is not set
	if !templateForce {
		if _, err := os.Stat(templateOut); err == nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorFileExists, map[string]interface{}{
				"Path": templateOut,
			}))
		}
	}

	// Load values
	values := make(map[string]interface{})

	// Load values from file if provided
	if valuesFile != "" {
		fileValues, err := loadValuesFromFile(valuesFile)
		if err != nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorLoadValues, map[string]interface{}{
				"Error": err.Error(),
			}))
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
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorInvalidSet, map[string]interface{}{
				"Value": setValue,
			}))
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
	
	// Handle dry-run mode
	if templateDryRun {
		// Get template information
		var placeholders []string
		var templateType string
		
		switch ext {
		case ".docx":
			processor := template.NewWordProcessor()
			foundPlaceholders, err := processor.ExtractPlaceholders(templatePath)
			if err != nil {
				return fmt.Errorf("failed to extract placeholders: %w", err)
			}
			placeholders = foundPlaceholders
			templateType = "Word Document"
		case ".pptx":
			processor := template.NewPowerPointProcessor()
			foundPlaceholders, err := processor.ExtractPlaceholders(templatePath)
			if err != nil {
				return fmt.Errorf("failed to extract placeholders: %w", err)
			}
			placeholders = foundPlaceholders
			templateType = "PowerPoint Presentation"
		default:
			return fmt.Errorf("unsupported template format: %s", ext)
		}
		
		// Check which placeholders will be replaced
		replaced := make([]string, 0)
		missing := make([]string, 0)
		
		for _, placeholder := range placeholders {
			if _, exists := values[placeholder]; exists {
				replaced = append(replaced, placeholder)
			} else {
				missing = append(missing, placeholder)
			}
		}
		
		if templateJsonOutput {
			// JSON output for dry-run
			dryRunInfo := map[string]interface{}{
				"operation": "template",
				"template": map[string]interface{}{
					"path": templatePath,
					"type": templateType,
				},
				"placeholders": map[string]interface{}{
					"found":    placeholders,
					"replaced": replaced,
					"missing":  missing,
				},
				"values": values,
				"output": templateOut,
			}
			
			jsonBytes, _ := json.MarshalIndent(dryRunInfo, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			// Human-readable output for dry-run
			fmt.Println("=== DRY-RUN MODE ===")
			fmt.Println()
			fmt.Printf("Template: %s (%s)\n", templatePath, templateType)
			fmt.Printf("Output:   %s\n", templateOut)
			fmt.Println()
			
			fmt.Printf("Placeholders found: %d\n", len(placeholders))
			if len(placeholders) > 0 {
				fmt.Println("  " + strings.Join(placeholders, ", "))
			}
			fmt.Println()
			
			fmt.Printf("Values to be replaced: %d\n", len(replaced))
			if len(replaced) > 0 {
				for _, key := range replaced {
					fmt.Printf("  {{%s}} → %v\n", key, values[key])
				}
			}
			fmt.Println()
			
			if len(missing) > 0 {
				fmt.Printf("Missing values: %d\n", len(missing))
				fmt.Println("  " + strings.Join(missing, ", "))
				fmt.Println()
			}
			
			fmt.Println("No files were created. Remove --dry-run to execute.")
		}
		
		return nil
	}
	
	switch ext {
	case ".docx":
		processor := template.NewWordProcessor()
		
		// Validate template
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err != nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorValidate, map[string]interface{}{
				"Type":  "template",
				"Error": err.Error(),
			}))
		}
		
		if len(missing) > 0 {
			cmd.PrintErrf("%s\n", i18n.T(i18n.MsgWarningNoValues, map[string]interface{}{
				"Placeholders": fmt.Sprintf("%v", missing),
			}))
		}
		
		// Process template
		cmd.Printf("%s\n", i18n.T(i18n.MsgProgressProcessing, map[string]interface{}{
			"Type": "Word",
		}))
		if err := processor.ProcessTemplate(templatePath, values, templateOut); err != nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorProcess, map[string]interface{}{
				"Type":  "template",
				"Error": err.Error(),
			}))
		}
		
	case ".pptx":
		processor := template.NewPowerPointProcessor()
		
		// Validate template
		missing, err := processor.ValidateTemplate(templatePath, values)
		if err != nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorValidate, map[string]interface{}{
				"Type":  "template",
				"Error": err.Error(),
			}))
		}
		
		if len(missing) > 0 {
			cmd.PrintErrf("%s\n", i18n.T(i18n.MsgWarningNoValues, map[string]interface{}{
				"Placeholders": fmt.Sprintf("%v", missing),
			}))
		}
		
		// Process template
		cmd.Printf("%s\n", i18n.T(i18n.MsgProgressProcessing, map[string]interface{}{
			"Type": "PowerPoint",
		}))
		if err := processor.ProcessTemplate(templatePath, values, templateOut); err != nil {
			return fmt.Errorf("%s", i18n.T(i18n.MsgErrorProcess, map[string]interface{}{
				"Type":  "template",
				"Error": err.Error(),
			}))
		}
		
	default:
		return fmt.Errorf("%s", i18n.T(i18n.MsgErrorUnsupported, map[string]interface{}{
			"Type":      "template format",
			"Value":     ext,
			"Supported": ".docx, .pptx",
		}))
	}

	cmd.Printf("%s\n", i18n.T(i18n.MsgSuccessCreated, map[string]interface{}{
		"File": templateOut,
	}))
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
			return nil, fmt.Errorf("%s", i18n.T(i18n.MsgErrorParseYAML, map[string]interface{}{
				"Error": err.Error(),
			}))
		}
	case ".json":
		if err := json.Unmarshal(data, &values); err != nil {
			return nil, fmt.Errorf("%s", i18n.T(i18n.MsgErrorParseJSON, map[string]interface{}{
				"Error": err.Error(),
			}))
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &values); err != nil {
			if err := json.Unmarshal(data, &values); err != nil {
				return nil, fmt.Errorf("%s", i18n.T(i18n.MsgErrorParseFile))
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