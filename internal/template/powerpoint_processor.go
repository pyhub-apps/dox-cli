package template

import (
	"fmt"
	"github.com/pyhub/pyhub-docs/internal/document"
)

// PowerPointProcessor handles template processing for PowerPoint documents
type PowerPointProcessor struct {
	parser *Parser
}

// NewPowerPointProcessor creates a new PowerPoint template processor
func NewPowerPointProcessor() *PowerPointProcessor {
	return &PowerPointProcessor{
		parser: NewParser(),
	}
}

// ProcessTemplate processes a PowerPoint template with the given values
func (p *PowerPointProcessor) ProcessTemplate(templatePath string, values map[string]interface{}, outputPath string) error {
	// Open template document
	doc, err := document.OpenPowerPointDocument(templatePath)
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
	placeholders := p.parser.FindPlaceholders(text)
	
	// Replace placeholders
	for _, placeholder := range placeholders {
		value := p.getPlaceholderValue(placeholder.Name, values)
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
func (p *PowerPointProcessor) ValidateTemplate(templatePath string, values map[string]interface{}) ([]string, error) {
	// Open template document
	doc, err := document.OpenPowerPointDocument(templatePath)
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
	missing := p.parser.ValidatePlaceholders(text, values)
	return missing, nil
}

// getPlaceholderValue gets the value for a placeholder
func (p *PowerPointProcessor) getPlaceholderValue(name string, values map[string]interface{}) string {
	return p.parser.getValueForPlaceholder(name, values)
}