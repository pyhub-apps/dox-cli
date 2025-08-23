package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestTemplateCommand(t *testing.T) {
	t.Run("Command Registration", func(t *testing.T) {
		// Check that template command is registered
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "template" {
				found = true
				break
			}
		}
		if !found {
			t.Error("template command not registered with root command")
		}
	})

	t.Run("Command Flags", func(t *testing.T) {
		// Check that required flags are defined
		if templateCmd.Flags().Lookup("template") == nil {
			t.Error("--template flag not defined")
		}
		if templateCmd.Flags().Lookup("values") == nil {
			t.Error("--values flag not defined")
		}
		if templateCmd.Flags().Lookup("output") == nil {
			t.Error("--output flag not defined")
		}
		if templateCmd.Flags().Lookup("set") == nil {
			t.Error("--set flag not defined")
		}
		if templateCmd.Flags().Lookup("force") == nil {
			t.Error("--force flag not defined")
		}
	})

	t.Run("Missing Required Flags", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *templateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Reset flags
		templatePath = ""
		templateOut = ""
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error when required flags are missing")
		}
	})

	t.Run("Invalid Template File", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *templateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Set non-existent template file
		templatePath = "/non/existent/template.docx"
		templateOut = "output.docx"
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error with non-existent template file")
		}
	})

	t.Run("Values File Validation", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "template_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create template file
		templateFile := filepath.Join(tempDir, "template.docx")
		if err := os.WriteFile(templateFile, []byte{0x50, 0x4B}, 0644); err != nil {
			t.Fatal(err)
		}
		
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *templateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Set non-existent values file
		templatePath = templateFile
		templateOut = filepath.Join(tempDir, "output.docx")
		valuesFile = "/non/existent/values.yaml"
		
		err = cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error with non-existent values file")
		}
		
		// Reset values file
		valuesFile = ""
	})

	t.Run("Force Flag", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "template_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create existing file
		existingFile := filepath.Join(tempDir, "existing.docx")
		if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		
		// Create template file
		templateFile := filepath.Join(tempDir, "template.docx")
		if err := os.WriteFile(templateFile, []byte{0x50, 0x4B}, 0644); err != nil {
			t.Fatal(err)
		}
		
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *templateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Try without force flag
		templatePath = templateFile
		templateOut = existingFile
		templateForce = false
		
		err = cmd.RunE(cmd, []string{})
		// Should error because file exists
		if err == nil {
			t.Error("Expected error when output file exists without force flag")
		}
		
		// Reset force flag
		templateForce = false
	})

	t.Run("Set Values Format", func(t *testing.T) {
		testValues := []string{
			"key=value",
			"name=John Doe",
			"year=2024",
			"title=Test Document",
		}
		
		for _, val := range testValues {
			parts := strings.SplitN(val, "=", 2)
			if len(parts) != 2 {
				t.Errorf("Invalid set value format: %s", val)
			}
		}
	})

	t.Run("Template Extension Detection", func(t *testing.T) {
		tests := []struct {
			name         string
			templateFile string
			isValid      bool
		}{
			{"Word template", "template.docx", true},
			{"PowerPoint template", "template.pptx", true},
			{"Invalid extension", "template.txt", false},
			{"No extension", "template", false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ext := filepath.Ext(tt.templateFile)
				isDocx := ext == ".docx"
				isPptx := ext == ".pptx"
				isValid := isDocx || isPptx
				
				if isValid != tt.isValid {
					t.Errorf("Template validation failed for %s: got %v, want %v", tt.templateFile, isValid, tt.isValid)
				}
			})
		}
	})

	t.Run("Values File Format Detection", func(t *testing.T) {
		tests := []struct {
			name     string
			filename string
			isYAML   bool
			isJSON   bool
		}{
			{"YAML file", "values.yaml", true, false},
			{"YAML file yml", "values.yml", true, false},
			{"JSON file", "values.json", false, true},
			{"Unknown format", "values.txt", false, false},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ext := filepath.Ext(tt.filename)
				isYAML := ext == ".yaml" || ext == ".yml"
				isJSON := ext == ".json"
				
				if isYAML != tt.isYAML {
					t.Errorf("YAML detection failed for %s: got %v, want %v", tt.filename, isYAML, tt.isYAML)
				}
				if isJSON != tt.isJSON {
					t.Errorf("JSON detection failed for %s: got %v, want %v", tt.filename, isJSON, tt.isJSON)
				}
			})
		}
	})
}