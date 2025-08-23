package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pyhub/pyhub-docs/internal/config"
	"gopkg.in/yaml.v3"
)

func TestConfigCommand(t *testing.T) {
	// Test config command is registered
	t.Run("CommandRegistered", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "config" {
				found = true
				break
			}
		}
		if !found {
			t.Error("config command not found")
		}
	})
}

func TestInitConfigFile(t *testing.T) {
	t.Run("CreateNewConfig", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		configPath := filepath.Join(tempDir, "config.yml")
		
		// Create config file
		err = initConfigFile(configPath)
		if err != nil {
			t.Errorf("initConfigFile failed: %v", err)
		}
		
		// Check file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
		
		// Load and verify config
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Errorf("Failed to load created config: %v", err)
		}
		
		// Check default values
		if cfg.Global.Lang != "en" {
			t.Errorf("Expected default lang 'en', got %s", cfg.Global.Lang)
		}
	})
	
	t.Run("ExistingConfigWithoutForce", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		configPath := filepath.Join(tempDir, "config.yml")
		
		// Create existing file
		if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		
		// Try to create without force flag
		force = false
		err = initConfigFile(configPath)
		if err == nil {
			t.Error("Expected error for existing file without force")
		}
	})
	
	t.Run("ExistingConfigWithForce", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		configPath := filepath.Join(tempDir, "config.yml")
		
		// Create existing file
		if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		
		// Try to create with force flag
		force = true
		defer func() { force = false }()
		
		err = initConfigFile(configPath)
		if err != nil {
			t.Errorf("initConfigFile with force failed: %v", err)
		}
	})
}

func TestListConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		// Create temp config file
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		
		configPath := filepath.Join(tempDir, "config.yml")
		
		// Create config
		cfg := config.DefaultConfig()
		data, _ := yaml.Marshal(cfg)
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatal(err)
		}
		
		// List config
		err = listConfig(configPath)
		if err != nil {
			t.Errorf("listConfig failed: %v", err)
		}
	})
	
	t.Run("NonExistentConfig", func(t *testing.T) {
		// Note: config.Load returns default config for non-existent files
		// This is expected behavior, so we just test that it doesn't panic
		err := listConfig("/nonexistent/config.yml")
		// If it returns a default config, that's okay
		_ = err
	})
}

func TestGetConfig(t *testing.T) {
	// Create temp config file
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.yml")
	
	// Create config with test values
	cfg := config.DefaultConfig()
	cfg.OpenAI.Model = "gpt-4-test"
	cfg.Global.Lang = "ko"
	
	data, _ := yaml.Marshal(cfg)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		key      string
		hasError bool
	}{
		{"openai.model", false},
		{"global.lang", false},
		{"global.verbose", false},
		{"replace.backup", false},
		{"unknown.key", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := getConfig(configPath, tt.key)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSetConfig(t *testing.T) {
	// Create temp config file
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.yml")
	
	// Create initial config
	cfg := config.DefaultConfig()
	data, _ := yaml.Marshal(cfg)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		keyValue string
		hasError bool
	}{
		{"openai.model=gpt-4", false},
		{"global.verbose=true", false},
		{"global.quiet=false", false},
		{"global.lang=ko", false},
		{"replace.backup=true", false},
		{"invalid", true}, // No = sign
		{"unknown.key=value", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.keyValue, func(t *testing.T) {
			err := setConfig(configPath, tt.keyValue)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRunConfig(t *testing.T) {
	t.Run("ShowHelp", func(t *testing.T) {
		// Reset flags
		configInit = false
		configList = false
		configSet = ""
		configGet = ""
		
		// runConfig should show help when no flags are set
		err := runConfig(configCmd, []string{})
		if err != nil {
			t.Errorf("runConfig failed: %v", err)
		}
	})
}