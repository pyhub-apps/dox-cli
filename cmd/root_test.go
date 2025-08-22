package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/i18n"
)

func TestRootCommand(t *testing.T) {
	// Initialize i18n for testing
	if err := i18n.Init("en"); err != nil {
		t.Fatal(err)
	}

	// Test Execute function doesn't panic
	t.Run("ExecuteCommand", func(t *testing.T) {
		// Capture output
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		
		// Set help flag to just print help and exit
		rootCmd.SetArgs([]string{"--help"})
		
		// Execute should not return error for help
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("rootCmd.Execute() failed: %v", err)
		}
		
		// Check if help output contains expected text
		output := buf.String()
		if !strings.Contains(output, "pyhub-docs") {
			t.Error("Help output doesn't contain 'pyhub-docs'")
		}
	})

	// Test config flag is registered
	t.Run("ConfigFlag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		if flag == nil {
			t.Error("config flag not found")
		}
	})

	// Test verbose flag
	t.Run("VerboseFlag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		if flag == nil {
			t.Error("verbose flag not found")
		}
	})

	// Test quiet flag
	t.Run("QuietFlag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("quiet")
		if flag == nil {
			t.Error("quiet flag not found")
		}
	})

	// Test lang flag
	t.Run("LangFlag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("lang")
		if flag == nil {
			t.Error("lang flag not found")
		}
	})

	// Test sub-commands are registered
	t.Run("SubCommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make(map[string]bool)
		for _, cmd := range commands {
			commandNames[cmd.Name()] = true
		}

		expectedCommands := []string{"replace", "create", "template", "generate", "version"}
		for _, expected := range expectedCommands {
			if !commandNames[expected] {
				t.Errorf("Command %s not found", expected)
			}
		}
	})
}

func TestInitConfig(t *testing.T) {
	// Test that initConfig doesn't panic
	t.Run("InitConfigNoPanic", func(t *testing.T) {
		// This should not panic
		initConfig()
	})
}