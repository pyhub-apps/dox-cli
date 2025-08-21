package template

import (
	"reflect"
	"testing"
)

func TestFindPlaceholders(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		text     string
		expected []string // placeholder names
	}{
		{
			name:     "simple placeholders",
			text:     "Hello {{name}}, welcome to {{company}}!",
			expected: []string{"name", "company"},
		},
		{
			name:     "nested placeholders",
			text:     "Author: {{author.name}}, Email: {{author.email}}",
			expected: []string{"author.name", "author.email"},
		},
		{
			name:     "repeated placeholders",
			text:     "{{title}} - Content - {{title}}",
			expected: []string{"title", "title"},
		},
		{
			name:     "no placeholders",
			text:     "This is plain text without any placeholders.",
			expected: []string{},
		},
		{
			name:     "placeholders with underscores and dashes",
			text:     "{{first_name}} {{last-name}} {{user.full_name}}",
			expected: []string{"first_name", "last-name", "user.full_name"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placeholders := parser.FindPlaceholders(tt.text)
			
			if len(placeholders) != len(tt.expected) {
				t.Errorf("Expected %d placeholders, got %d", len(tt.expected), len(placeholders))
				return
			}
			
			for i, placeholder := range placeholders {
				if placeholder.Name != tt.expected[i] {
					t.Errorf("Expected placeholder %d to be %s, got %s", i, tt.expected[i], placeholder.Name)
				}
			}
		})
	}
}

func TestReplacePlaceholders(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name     string
		text     string
		values   map[string]interface{}
		expected string
	}{
		{
			name: "simple replacement",
			text: "Hello {{name}}, welcome to {{company}}!",
			values: map[string]interface{}{
				"name":    "John Doe",
				"company": "TechCorp",
			},
			expected: "Hello John Doe, welcome to TechCorp!",
		},
		{
			name: "nested values",
			text: "Author: {{author.name}}, Email: {{author.email}}",
			values: map[string]interface{}{
				"author": map[string]interface{}{
					"name":  "Jane Smith",
					"email": "jane@example.com",
				},
			},
			expected: "Author: Jane Smith, Email: jane@example.com",
		},
		{
			name: "missing values",
			text: "Hello {{name}}, your ID is {{id}}",
			values: map[string]interface{}{
				"name": "Alice",
			},
			expected: "Hello Alice, your ID is {{id}}",
		},
		{
			name: "numeric values",
			text: "Year: {{year}}, Score: {{score}}",
			values: map[string]interface{}{
				"year":  2024,
				"score": 95.5,
			},
			expected: "Year: 2024, Score: 95.5",
		},
		{
			name: "array values",
			text: "Items: {{items}}",
			values: map[string]interface{}{
				"items": []interface{}{"apple", "banana", "orange"},
			},
			expected: "Items: apple, banana, orange",
		},
		{
			name: "boolean values",
			text: "Active: {{active}}, Verified: {{verified}}",
			values: map[string]interface{}{
				"active":   true,
				"verified": false,
			},
			expected: "Active: true, Verified: false",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ReplacePlaceholders(tt.text, tt.values)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidatePlaceholders(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		name    string
		text    string
		values  map[string]interface{}
		missing []string
	}{
		{
			name: "all placeholders have values",
			text: "{{name}} - {{email}}",
			values: map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
			missing: []string{},
		},
		{
			name: "some placeholders missing",
			text: "{{name}} - {{email}} - {{phone}}",
			values: map[string]interface{}{
				"name": "John",
			},
			missing: []string{"email", "phone"},
		},
		{
			name: "nested placeholder missing",
			text: "{{user.name}} - {{user.email}}",
			values: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Jane",
				},
			},
			missing: []string{"user.email"},
		},
		{
			name: "repeated placeholder counted once",
			text: "{{title}} content {{title}} footer {{title}}",
			values: map[string]interface{}{},
			missing: []string{"title"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := parser.ValidatePlaceholders(tt.text, tt.values)
			
			if !reflect.DeepEqual(missing, tt.missing) {
				t.Errorf("Expected missing %v, got %v", tt.missing, missing)
			}
		})
	}
}