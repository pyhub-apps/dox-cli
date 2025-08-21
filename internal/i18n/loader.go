package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// LoadFromFiles loads locale files from the file system instead of embedded files
// This is useful for development and when locale files are external
func LoadFromFiles(localesDir string) error {
	if bundle == nil {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	}

	// Find all JSON files in the locales directory
	files, err := filepath.Glob(filepath.Join(localesDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to find locale files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no locale files found in %s", localesDir)
	}

	// Load each locale file
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read locale file %s: %w", file, err)
		}

		_, err = bundle.ParseMessageFileBytes(data, filepath.Base(file))
		if err != nil {
			return fmt.Errorf("failed to parse locale file %s: %w", file, err)
		}
	}

	return nil
}

// InitWithFiles initializes i18n using external locale files
func InitWithFiles(localesDir string, lang string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load locale files from directory
	if err := LoadFromFiles(localesDir); err != nil {
		return err
	}

	// Set up localizer
	userLang := determineLanguage(lang)
	localizer = i18n.NewLocalizer(bundle, userLang)

	return nil
}