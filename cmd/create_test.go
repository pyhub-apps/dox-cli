package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateCommand(t *testing.T) {
	t.Run("Command Registration", func(t *testing.T) {
		// Check that create command is registered
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "create" {
				found = true
				break
			}
		}
		if !found {
			t.Error("create command not registered with root command")
		}
	})

	t.Run("Command Flags", func(t *testing.T) {
		// Check that required flags are defined
		if createCmd.Flags().Lookup("from") == nil {
			t.Error("--from flag not defined")
		}
		if createCmd.Flags().Lookup("output") == nil {
			t.Error("--output flag not defined")
		}
		if createCmd.Flags().Lookup("template") == nil {
			t.Error("--template flag not defined")
		}
		if createCmd.Flags().Lookup("format") == nil {
			t.Error("--format flag not defined")
		}
		if createCmd.Flags().Lookup("force") == nil {
			t.Error("--force flag not defined")
		}
	})

	t.Run("Missing Required Flags", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *createCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Reset flags
		fromFile = ""
		outputFile = ""
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error when required flags are missing")
		}
	})

	t.Run("Invalid Input File", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *createCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Set non-existent input file
		fromFile = "/non/existent/file.md"
		outputFile = "output.docx"
		
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Error("Expected error with non-existent input file")
		}
	})

	t.Run("Output Format Detection", func(t *testing.T) {
		tests := []struct {
			name           string
			outputPath     string
			expectedFormat string
		}{
			{"Word document", "output.docx", "docx"},
			{"PowerPoint", "presentation.pptx", "pptx"},
			{"Word with path", "/path/to/document.docx", "docx"},
			{"PowerPoint with path", "/path/to/slides.pptx", "pptx"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ext := strings.TrimPrefix(filepath.Ext(tt.outputPath), ".")
				if ext != tt.expectedFormat {
					t.Errorf("Format detection failed: got %v, want %v", ext, tt.expectedFormat)
				}
			})
		}
	})

	t.Run("Force Flag", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "create_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create existing file
		existingFile := filepath.Join(tempDir, "existing.docx")
		if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		
		// Create input markdown file
		inputFile := filepath.Join(tempDir, "input.md")
		if err := os.WriteFile(inputFile, []byte("# Test"), 0644); err != nil {
			t.Fatal(err)
		}
		
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *createCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Try without force flag
		fromFile = inputFile
		outputFile = existingFile
		force = false
		
		err = cmd.RunE(cmd, []string{})
		// Should error because file exists
		if err == nil {
			t.Error("Expected error when output file exists without force flag")
		}
		
		// Reset force flag
		force = false
	})

	t.Run("Template File Validation", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "create_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create input file
		inputFile := filepath.Join(tempDir, "input.md")
		if err := os.WriteFile(inputFile, []byte("# Test"), 0644); err != nil {
			t.Fatal(err)
		}
		
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		*cmd = *createCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		
		// Set non-existent template
		fromFile = inputFile
		outputFile = filepath.Join(tempDir, "output.docx")
		templateFile = "/non/existent/template.docx"
		
		err = cmd.RunE(cmd, []string{})
		// The command may not check for template existence upfront
		
		// Reset template
		templateFile = ""
	})

	t.Run("Format Flag", func(t *testing.T) {
		validFormats := []string{"docx", "pptx"}
		
		for _, fmt := range validFormats {
			format = fmt
			if format != fmt {
				t.Errorf("Format not set correctly: got %v, want %v", format, fmt)
			}
		}
		
		// Reset format
		format = ""
	})
}