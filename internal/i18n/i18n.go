package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Embed locale files
//
//go:embed locales/*.json
var localesFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
)

// Init initializes the i18n system with the specified language
func Init(lang string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load embedded locale files
	files, err := localesFS.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			data, err := localesFS.ReadFile("locales/" + file.Name())
			if err != nil {
				return fmt.Errorf("failed to read locale file %s: %w", file.Name(), err)
			}
			
			bundle.ParseMessageFileBytes(data, file.Name())
		}
	}

	// Determine language preference
	userLang := determineLanguage(lang)
	localizer = i18n.NewLocalizer(bundle, userLang)

	return nil
}

// determineLanguage determines the language to use based on priority:
// 1. Explicit lang parameter
// 2. LANG environment variable
// 3. LC_ALL environment variable
// 4. Default to English
func determineLanguage(explicitLang string) string {
	if explicitLang != "" {
		return normalizeLanguage(explicitLang)
	}

	if lang := os.Getenv("LANG"); lang != "" {
		return normalizeLanguage(lang)
	}

	if lcAll := os.Getenv("LC_ALL"); lcAll != "" {
		return normalizeLanguage(lcAll)
	}

	return "en"
}

// normalizeLanguage extracts the language code from locale strings
// e.g., "en_US.UTF-8" -> "en", "ko_KR.UTF-8" -> "ko"
func normalizeLanguage(lang string) string {
	// Remove encoding suffix (e.g., .UTF-8)
	if idx := strings.Index(lang, "."); idx > 0 {
		lang = lang[:idx]
	}

	// Extract language code (e.g., en_US -> en)
	if idx := strings.Index(lang, "_"); idx > 0 {
		lang = lang[:idx]
	}

	// Convert to lowercase
	lang = strings.ToLower(lang)

	// Map common variations
	switch lang {
	case "korean", "kor", "kr":
		return "ko"
	case "english", "eng", "us", "uk":
		return "en"
	default:
		return lang
	}
}

// T translates a message with optional template data
func T(messageID string, templateData ...map[string]interface{}) string {
	if localizer == nil {
		// Fallback if i18n is not initialized
		return messageID
	}

	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(templateData) > 0 {
		config.TemplateData = templateData[0]
	}

	msg, err := localizer.Localize(config)
	if err != nil {
		// Fallback to message ID if translation not found
		return messageID
	}

	return msg
}

// GetLocalizer returns the current localizer
func GetLocalizer() *i18n.Localizer {
	return localizer
}

// SetLanguage changes the current language
func SetLanguage(lang string) {
	userLang := determineLanguage(lang)
	localizer = i18n.NewLocalizer(bundle, userLang)
}

// GetCurrentLanguage returns the current language code
func GetCurrentLanguage() string {
	if localizer == nil {
		return "en"
	}
	
	// Try to get the current language from environment
	lang := determineLanguage("")
	return lang
}