package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/i18n"
	. "github.com/pyhub/pyhub-docs/internal/ui"
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
		if !strings.Contains(output, "dox") && !strings.Contains(output, "pyhub-docs") {
			t.Error("Help output doesn't contain 'dox' or 'pyhub-docs'")
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

func TestInitI18n(t *testing.T) {
	t.Run("DefaultLanguage", func(t *testing.T) {
		// Should not panic with default language
		initI18n()
	})
	
	t.Run("WithLanguageFlag", func(t *testing.T) {
		// Test with specific language
		langFlag = "ko"
		defer func() { langFlag = "" }()
		
		initI18n()
		// Should initialize with Korean
	})
}

func TestInitUI(t *testing.T) {
	t.Run("QuietMode", func(t *testing.T) {
		quiet = true
		defer func() { quiet = false }()
		
		initUI()
		// Color should be disabled in quiet mode
		if IsColorEnabled() {
			t.Error("Color should be disabled in quiet mode")
		}
	})
	
	t.Run("NormalMode", func(t *testing.T) {
		quiet = false
		initUI()
		// Should initialize UI without issues
	})
}

func TestExecute(t *testing.T) {
	// Test the Execute function
	t.Run("ExecuteFunction", func(t *testing.T) {
		// Test that Execute() doesn't panic
		// Note: Execute() calls os.Exit on error, so we can't test it directly
		// We'll test that the command structure is set up correctly
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Execute() panicked: %v", r)
			}
		}()
		
		// Check that rootCmd is properly initialized
		if rootCmd == nil {
			t.Error("rootCmd is not initialized")
		}
		
		// Check that commands are registered
		if len(rootCmd.Commands()) == 0 {
			t.Error("No subcommands registered")
		}
	})
}