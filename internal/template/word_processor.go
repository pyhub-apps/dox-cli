package template

import (
	"fmt"
	"github.com/pyhub/pyhub-documents-cli/internal/document"
)

// WordProcessor handles template processing for Word documents
type WordProcessor struct {
	parser *Parser
}

// NewWordProcessor creates a new Word template processor
func NewWordProcessor() *WordProcessor {
	return &WordProcessor{
		parser: NewParser(),
	}
}

// ProcessTemplate processes a Word template with the given values
func (w *WordProcessor) ProcessTemplate(templatePath string, values map[string]interface{}, outputPath string) error {
	// Open template document
	doc, err := document.OpenWordDocument(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()
	
	// Get document text
	text, err := doc.GetText()
	if err != nil {
		return fmt.Errorf("failed to get text from template: %w", err)
	}
	
	// Find all placeholders
	placeholders := w.parser.FindPlaceholders(text)
	
	// Replace placeholders
	for _, placeholder := range placeholders {
		value := w.getPlaceholderValue(placeholder.Name, values)
		err = doc.ReplaceText(placeholder.Expression, value)
		if err != nil {
			return fmt.Errorf("failed to replace placeholder %s: %w", placeholder.Name, err)
		}
	}
	
	// Save the processed document
	if err := doc.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save processed document: %w", err)
	}
	
	return nil
}

// ValidateTemplate checks if all placeholders in the template have values
func (w *WordProcessor) ValidateTemplate(templatePath string, values map[string]interface{}) ([]string, error) {
	// Open template document
	doc, err := document.OpenWordDocument(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()
	
	// Get document text
	text, err := doc.GetText()
	if err != nil {
		return nil, fmt.Errorf("failed to get text from template: %w", err)
	}
	
	// Validate placeholders
	missing := w.parser.ValidatePlaceholders(text, values)
	return missing, nil
}

// getPlaceholderValue gets the value for a placeholder
func (w *WordProcessor) getPlaceholderValue(name string, values map[string]interface{}) string {
	return w.parser.getValueForPlaceholder(name, values)
}