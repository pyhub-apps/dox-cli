package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	// OpenAI 기본값 확인
	if cfg.OpenAI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected default model to be gpt-3.5-turbo, got %s", cfg.OpenAI.Model)
	}
	if cfg.OpenAI.MaxTokens != 2000 {
		t.Errorf("Expected default max tokens to be 2000, got %d", cfg.OpenAI.MaxTokens)
	}
	if cfg.OpenAI.Temperature != 0.7 {
		t.Errorf("Expected default temperature to be 0.7, got %f", cfg.OpenAI.Temperature)
	}
	
	// Replace 기본값 확인
	if cfg.Replace.Recursive != true {
		t.Error("Expected default recursive to be true")
	}
	
	// Global 기본값 확인
	if cfg.Global.Lang != "en" {
		t.Errorf("Expected default language to be 'en', got %s", cfg.Global.Lang)
	}
}

func TestLoadConfig(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// 테스트 설정 파일 생성
	testConfig := &Config{
		OpenAI: OpenAIConfig{
			APIKey:      "test-key",
			Model:       "gpt-4",
			MaxTokens:   3000,
			Temperature: 0.5,
		},
		Global: GlobalConfig{
			Verbose: true,
			Lang:    "ko",
		},
	}
	
	configPath := filepath.Join(tempDir, "test_config.yml")
	data, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	
	// 설정 파일 로드
	loadedConfig, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// 로드된 값 확인
	if loadedConfig.OpenAI.APIKey != "test-key" {
		t.Errorf("Expected API key to be 'test-key', got %s", loadedConfig.OpenAI.APIKey)
	}
	if loadedConfig.OpenAI.Model != "gpt-4" {
		t.Errorf("Expected model to be 'gpt-4', got %s", loadedConfig.OpenAI.Model)
	}
	if loadedConfig.Global.Verbose != true {
		t.Error("Expected verbose to be true")
	}
	if loadedConfig.Global.Lang != "ko" {
		t.Errorf("Expected language to be 'ko', got %s", loadedConfig.Global.Lang)
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	// 존재하지 않는 파일 로드
	cfg, err := Load("/non/existent/path/config.yml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got %v", err)
	}
	
	// 기본 설정이 반환되어야 함
	if cfg.OpenAI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected default model, got %s", cfg.OpenAI.Model)
	}
}

func TestSaveConfig(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir, err := os.MkdirTemp("", "config_save_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// 설정 생성 및 저장
	cfg := &Config{
		OpenAI: OpenAIConfig{
			APIKey: "save-test-key",
			Model:  "gpt-3.5-turbo",
		},
		Global: GlobalConfig{
			Lang: "ko",
		},
	}
	
	configPath := filepath.Join(tempDir, "saved_config.yml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// 파일이 생성되었는지 확인
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
	
	// 저장된 파일 로드하여 확인
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	if loadedCfg.OpenAI.APIKey != "save-test-key" {
		t.Errorf("Expected API key to be 'save-test-key', got %s", loadedCfg.OpenAI.APIKey)
	}
	if loadedCfg.Global.Lang != "ko" {
		t.Errorf("Expected language to be 'ko', got %s", loadedCfg.Global.Lang)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid config",
			config: &Config{
				OpenAI: OpenAIConfig{
					Model:       "gpt-3.5-turbo",
					Temperature: 0.5,
					MaxTokens:   1000,
				},
				Global: GlobalConfig{
					Lang:    "en",
					Verbose: false,
					Quiet:   false,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid model",
			config: &Config{
				OpenAI: OpenAIConfig{
					Model: "invalid-model",
				},
			},
			wantErr: true,
			errMsg:  "invalid OpenAI model",
		},
		{
			name: "Invalid temperature (too high)",
			config: &Config{
				OpenAI: OpenAIConfig{
					Model:       "gpt-3.5-turbo",
					Temperature: 1.5,
				},
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
		{
			name: "Invalid temperature (negative)",
			config: &Config{
				OpenAI: OpenAIConfig{
					Model:       "gpt-3.5-turbo",
					Temperature: -0.1,
				},
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
		{
			name: "Invalid max tokens",
			config: &Config{
				OpenAI: OpenAIConfig{
					Model:     "gpt-3.5-turbo",
					MaxTokens: -100,
				},
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
		{
			name: "Verbose and quiet both true",
			config: &Config{
				Global: GlobalConfig{
					Verbose: true,
					Quiet:   true,
				},
			},
			wantErr: true,
			errMsg:  "verbose and quiet cannot both be true",
		},
		{
			name: "Invalid language",
			config: &Config{
				Global: GlobalConfig{
					Lang: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid language",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 기본값 설정
			if tt.config.OpenAI.Temperature == 0 && !tt.wantErr {
				tt.config.OpenAI.Temperature = 0.7
			}
			
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// 환경 변수 백업
	originalEnv := os.Getenv("PYHUB_CONFIG")
	defer os.Setenv("PYHUB_CONFIG", originalEnv)
	
	// 환경 변수가 설정된 경우
	os.Setenv("PYHUB_CONFIG", "/custom/path/config.yml")
	path := GetConfigPath()
	if path != "/custom/path/config.yml" {
		t.Errorf("Expected custom path, got %s", path)
	}
	
	// 환경 변수가 없는 경우
	os.Unsetenv("PYHUB_CONFIG")
	path = GetConfigPath()
	
	// 홈 디렉토리 기반 경로이거나 현재 디렉토리 기반 경로여야 함
	if !contains(path, ".pyhub/config.yml") {
		t.Errorf("Expected path to contain '.pyhub/config.yml', got %s", path)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   (len(substr) > 0 && len(s) > 0 && s[:len(substr)] == substr) ||
		   (len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}