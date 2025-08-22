package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// OpenAI configuration
	OpenAI OpenAIConfig `yaml:"openai"`
	
	// Default command options
	Replace  ReplaceConfig  `yaml:"replace"`
	Create   CreateConfig   `yaml:"create"`
	Generate GenerateConfig `yaml:"generate"`
	Template TemplateConfig `yaml:"template"`
	
	// Global options
	Global GlobalConfig `yaml:"global"`
}

// OpenAIConfig contains OpenAI API settings
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// ReplaceConfig contains default settings for replace command
type ReplaceConfig struct {
	Backup    bool   `yaml:"backup"`
	Recursive bool   `yaml:"recursive"`
	DryRun    bool   `yaml:"dry_run"`
	Exclude   string `yaml:"exclude"`
	Concurrent bool  `yaml:"concurrent"`
	MaxWorkers int   `yaml:"max_workers"`
}

// CreateConfig contains default settings for create command
type CreateConfig struct {
	Force  bool   `yaml:"force"`
	Format string `yaml:"format"`
}

// GenerateConfig contains default settings for generate command
type GenerateConfig struct {
	ContentType string  `yaml:"content_type"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// TemplateConfig contains default settings for template command
type TemplateConfig struct {
	Force bool `yaml:"force"`
}

// GlobalConfig contains global application settings
type GlobalConfig struct {
	Verbose bool   `yaml:"verbose"`
	Quiet   bool   `yaml:"quiet"`
	Lang    string `yaml:"lang"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		OpenAI: OpenAIConfig{
			Model:       "gpt-3.5-turbo",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		Replace: ReplaceConfig{
			Backup:    false,
			Recursive: true,
			DryRun:    false,
			Concurrent: false,
			MaxWorkers: 0, // Will use runtime.NumCPU() if 0
		},
		Create: CreateConfig{
			Force: false,
		},
		Generate: GenerateConfig{
			ContentType: "custom",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   2000,
			Temperature: 0.7,
		},
		Template: TemplateConfig{
			Force: false,
		},
		Global: GlobalConfig{
			Verbose: false,
			Quiet:   false,
			Lang:    "en",
		},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, return default config
		return cfg, nil
	}
	
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	// Check if custom path is set via environment variable
	if path := os.Getenv("PYHUB_CONFIG"); path != "" {
		return path
	}
	
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".pyhub/config.yml"
	}
	
	return filepath.Join(home, ".pyhub", "config.yml")
}

// Merge merges CLI flags with config file settings
// CLI flags take precedence over config file
func (c *Config) Merge(flags map[string]interface{}) {
	// This is a helper method that commands can use to merge flags
	// Implementation depends on specific command needs
	// CLI flags should override config file settings
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate OpenAI settings
	if c.OpenAI.Model != "" {
		validModels := []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo-preview"}
		valid := false
		for _, m := range validModels {
			if c.OpenAI.Model == m {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid OpenAI model: %s", c.OpenAI.Model)
		}
	}
	
	// Validate temperature
	if c.OpenAI.Temperature < 0 || c.OpenAI.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}
	
	// Validate max tokens
	if c.OpenAI.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	
	// Validate global settings
	if c.Global.Verbose && c.Global.Quiet {
		return fmt.Errorf("verbose and quiet cannot both be true")
	}
	
	// Validate language
	if c.Global.Lang != "" {
		validLangs := []string{"en", "ko"}
		valid := false
		for _, l := range validLangs {
			if c.Global.Lang == l {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid language: %s (must be 'en' or 'ko')", c.Global.Lang)
		}
	}
	
	return nil
}