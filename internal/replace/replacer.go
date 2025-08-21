package replace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyhub/pyhub-documents-cli/internal/document"
)

// ReplaceInDocument applies replacement rules to a single Word document
func ReplaceInDocument(docPath string, rules []Rule) error {
	// Validate input
	if docPath == "" {
		return fmt.Errorf("document path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("document not found: %s", docPath)
	}

	// Skip if no rules to apply
	if len(rules) == 0 {
		return nil
	}

	// Validate all rules before processing
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid rule at index %d: %w", i, err)
		}
	}

	// Open the document
	doc, err := document.OpenWordDocument(docPath)
	if err != nil {
		return fmt.Errorf("failed to open document: %w", err)
	}
	defer doc.Close()

	// Apply each replacement rule
	for _, rule := range rules {
		if err := doc.ReplaceText(rule.Old, rule.New); err != nil {
			return fmt.Errorf("failed to replace '%s' with '%s': %w", rule.Old, rule.New, err)
		}
	}

	// Save the modified document
	if err := doc.Save(); err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

// ReplaceInDirectory applies replacement rules to all Word documents in a directory
func ReplaceInDirectory(dirPath string, rules []Rule, recursive bool) error {
	// Validate input
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", dirPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Skip if no rules to apply
	if len(rules) == 0 {
		return nil
	}

	// Validate all rules before processing
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid rule at index %d: %w", i, err)
		}
	}

	// Process documents in the directory
	var processErrors []error
	
	if recursive {
		err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Record error but continue processing other files
				processErrors = append(processErrors, fmt.Errorf("error accessing %s: %w", path, err))
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Process only .docx files
			if strings.HasSuffix(strings.ToLower(path), ".docx") {
				if err := ReplaceInDocument(path, rules); err != nil {
					processErrors = append(processErrors, fmt.Errorf("failed to process %s: %w", path, err))
				}
			}

			return nil
		})
	} else {
		// Non-recursive: only process files in the top-level directory
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			// Skip directories
			if entry.IsDir() {
				continue
			}

			// Process only .docx files
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".docx") {
				path := filepath.Join(dirPath, entry.Name())
				if err := ReplaceInDocument(path, rules); err != nil {
					processErrors = append(processErrors, fmt.Errorf("failed to process %s: %w", path, err))
				}
			}
		}
	}

	// Report any errors that occurred
	if len(processErrors) > 0 {
		// Combine all errors into a single error message
		var errMsg strings.Builder
		errMsg.WriteString("some documents could not be processed:\n")
		for _, err := range processErrors {
			errMsg.WriteString("  - ")
			errMsg.WriteString(err.Error())
			errMsg.WriteString("\n")
		}
		return fmt.Errorf("%s", errMsg.String())
	}

	return nil
}

// ReplaceResult represents the result of a replacement operation
type ReplaceResult struct {
	FilePath     string
	Success      bool
	Error        error
	Replacements int
}

// ReplaceInDirectoryWithResults applies replacement rules and returns detailed results
func ReplaceInDirectoryWithResults(dirPath string, rules []Rule, recursive bool) ([]ReplaceResult, error) {
	var results []ReplaceResult

	// Validate input
	if dirPath == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found: %s", dirPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Skip if no rules to apply
	if len(rules) == 0 {
		return results, nil
	}

	// Validate all rules before processing
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return nil, fmt.Errorf("invalid rule at index %d: %w", i, err)
		}
	}

	// Process documents in the directory
	processFile := func(path string) {
		result := ReplaceResult{
			FilePath:     path,
			Replacements: len(rules), // Simplified: assume all rules are applied
		}

		if err := ReplaceInDocument(path, rules); err != nil {
			result.Success = false
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	if recursive {
		err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue processing other files
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Process only .docx files
			if strings.HasSuffix(strings.ToLower(path), ".docx") {
				processFile(path)
			}

			return nil
		})
	} else {
		// Non-recursive: only process files in the top-level directory
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			// Skip directories
			if entry.IsDir() {
				continue
			}

			// Process only .docx files
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".docx") {
				path := filepath.Join(dirPath, entry.Name())
				processFile(path)
			}
		}
	}

	return results, nil
}