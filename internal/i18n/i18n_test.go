package i18n

import (
	"os"
	"testing"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en_US.UTF-8", "en"},
		{"ko_KR.UTF-8", "ko"},
		{"en_US", "en"},
		{"ko_KR", "ko"},
		{"en", "en"},
		{"ko", "ko"},
		{"korean", "ko"},
		{"english", "en"},
		{"kor", "ko"},
		{"eng", "en"},
		{"kr", "ko"},
		{"us", "en"},
		{"EN_US.UTF-8", "en"},
		{"KO_KR.UTF-8", "ko"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeLanguage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetermineLanguage(t *testing.T) {
	// Save original env vars
	origLang := os.Getenv("LANG")
	origLCAll := os.Getenv("LC_ALL")
	defer func() {
		os.Setenv("LANG", origLang)
		os.Setenv("LC_ALL", origLCAll)
	}()

	tests := []struct {
		name         string
		explicitLang string
		envLang      string
		envLCAll     string
		expected     string
	}{
		{
			name:         "explicit lang takes priority",
			explicitLang: "ko",
			envLang:      "en_US.UTF-8",
			envLCAll:     "en_US.UTF-8",
			expected:     "ko",
		},
		{
			name:         "LANG env var used when no explicit",
			explicitLang: "",
			envLang:      "ko_KR.UTF-8",
			envLCAll:     "en_US.UTF-8",
			expected:     "ko",
		},
		{
			name:         "LC_ALL used when no LANG",
			explicitLang: "",
			envLang:      "",
			envLCAll:     "ko_KR.UTF-8",
			expected:     "ko",
		},
		{
			name:         "default to en when nothing set",
			explicitLang: "",
			envLang:      "",
			envLCAll:     "",
			expected:     "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LANG", tt.envLang)
			os.Setenv("LC_ALL", tt.envLCAll)

			result := determineLanguage(tt.explicitLang)
			if result != tt.expected {
				t.Errorf("determineLanguage(%q) = %q, want %q", tt.explicitLang, result, tt.expected)
			}
		})
	}
}

func TestInitWithFiles(t *testing.T) {
	// Test initializing with locale files from directory
	err := InitWithFiles("locales", "")
	if err != nil {
		t.Fatalf("InitWithFiles failed: %v", err)
	}

	// Test that localizer is set
	if GetLocalizer() == nil {
		t.Error("Localizer not initialized")
	}
}

func TestTranslation(t *testing.T) {
	// Initialize with test locale files
	err := InitWithFiles("locales", "en")
	if err != nil {
		t.Fatalf("InitWithFiles failed: %v", err)
	}

	// Test simple translation
	msg := T(MsgCmdRootShort)
	if msg == "" || msg == MsgCmdRootShort {
		t.Errorf("Translation failed for %s", MsgCmdRootShort)
	}

	// Test translation with template data
	msg = T(MsgSuccessCreated, map[string]interface{}{
		"File": "test.docx",
	})
	if msg == "" || msg == MsgSuccessCreated {
		t.Errorf("Translation with template failed for %s", MsgSuccessCreated)
	}

	// Test fallback for non-existent message
	msg = T("non.existent.message")
	if msg != "non.existent.message" {
		t.Errorf("Fallback failed, got %q", msg)
	}
}

func TestSetLanguage(t *testing.T) {
	// Initialize first
	err := InitWithFiles("locales", "en")
	if err != nil {
		t.Fatalf("InitWithFiles failed: %v", err)
	}

	// Switch to Korean
	SetLanguage("ko")

	// Get Korean translation
	msg := T(MsgCmdRootShort)
	// Since we know the Korean translation exists, check it's different from English
	if msg == "" || msg == MsgCmdRootShort {
		t.Errorf("Language switch to Korean failed")
	}

	// Switch back to English
	SetLanguage("en")

	// Verify we're back to English
	msg = T(MsgCmdRootShort)
	if msg != "Document automation and AI-powered content generation CLI" {
		t.Errorf("Language switch back to English failed")
	}
}