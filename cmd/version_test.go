package cmd

import (
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Test version command execution
	t.Run("VersionOutput", func(t *testing.T) {
		// The version command uses Run, not RunE, so it doesn't return an error
		// Just verify it doesn't panic
		args := []string{}
		
		// This should not panic
		versionCmd.Run(versionCmd, args)
	})

	// Test that version command is registered with root
	t.Run("VersionCommandRegistered", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "version" {
				found = true
				break
			}
		}
		if !found {
			t.Error("version command not registered with root command")
		}
	})
}