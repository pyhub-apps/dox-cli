package template

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Placeholder represents a template placeholder
type Placeholder struct {
	Name       string
	Expression string // Full expression including {{ }}
	Position   int    // Position in text
}

// Parser handles template parsing and placeholder extraction
type Parser struct {
	placeholderPattern *regexp.Regexp
}

// NewParser creates a new template parser
func NewParser() *Parser {
	// Pattern to match {{placeholder_name}} format
	// Supports alphanumeric, underscore, dash, and dot
	pattern := regexp.MustCompile(`\{\{([a-zA-Z0-9_\-\.]+)\}\}`)
	return &Parser{
		placeholderPattern: pattern,
	}
}

// FindPlaceholders finds all placeholders in the given text
func (p *Parser) FindPlaceholders(text string) []Placeholder {
	matches := p.placeholderPattern.FindAllStringSubmatchIndex(text, -1)
	placeholders := make([]Placeholder, 0, len(matches))
	
	for _, match := range matches {
		// match[0] and match[1] are the start and end of the full match
		// match[2] and match[3] are the start and end of the first capturing group
		fullMatch := text[match[0]:match[1]]
		placeholderName := text[match[2]:match[3]]
		
		placeholders = append(placeholders, Placeholder{
			Name:       placeholderName,
			Expression: fullMatch,
			Position:   match[0],
		})
	}
	
	return placeholders
}

// ReplacePlaceholders replaces placeholders in text with provided values
func (p *Parser) ReplacePlaceholders(text string, values map[string]interface{}) string {
	result := text
	
	// Find all placeholders
	placeholders := p.FindPlaceholders(text)
	
	// Replace from end to start to maintain positions
	for i := len(placeholders) - 1; i >= 0; i-- {
		placeholder := placeholders[i]
		
		// Get value for placeholder
		value := p.getValueForPlaceholder(placeholder.Name, values)
		
		// Replace placeholder with value
		result = strings.Replace(result, placeholder.Expression, value, 1)
	}
	
	return result
}

// getValueForPlaceholder retrieves the value for a placeholder name
func (p *Parser) getValueForPlaceholder(name string, values map[string]interface{}) string {
	// Handle nested values (e.g., "author.name")
	parts := strings.Split(name, ".")
	current := values
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - get the actual value
			if val, ok := current[part]; ok {
				return p.formatValue(val)
			}
		} else {
			// Navigate nested maps
			if nested, ok := current[part].(map[string]interface{}); ok {
				current = nested
			} else {
				break
			}
		}
	}
	
	// Return placeholder unchanged if value not found
	return fmt.Sprintf("{{%s}}", name)
}

// formatValue formats a value as a string
func (p *Parser) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format("2006-01-02")
	case []interface{}:
		// For arrays, join with comma
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = p.formatValue(item)
		}
		return strings.Join(items, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValidatePlaceholders checks if all placeholders have corresponding values
func (p *Parser) ValidatePlaceholders(text string, values map[string]interface{}) []string {
	placeholders := p.FindPlaceholders(text)
	missing := make([]string, 0)
	
	for _, placeholder := range placeholders {
		value := p.getValueForPlaceholder(placeholder.Name, values)
		// If the value is still a placeholder, it means it wasn't found
		if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
			missing = append(missing, placeholder.Name)
		}
	}
	
	// Remove duplicates
	uniqueMissing := make([]string, 0)
	seen := make(map[string]bool)
	for _, name := range missing {
		if !seen[name] {
			seen[name] = true
			uniqueMissing = append(uniqueMissing, name)
		}
	}
	
	return uniqueMissing
}