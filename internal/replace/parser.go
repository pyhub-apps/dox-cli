package replace

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAMLRules parses YAML data into a slice of Rules
func ParseYAMLRules(data []byte) ([]Rule, error) {
	// Handle empty data
	if len(data) == 0 {
		return []Rule{}, nil
	}

	// First, parse as generic interface to check structure
	var rawRules []map[string]interface{}
	err := yaml.Unmarshal(data, &rawRules)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Validate and convert each rule
	rules := make([]Rule, 0, len(rawRules))
	for i, rawRule := range rawRules {
		// Check for required fields
		if _, hasOld := rawRule["old"]; !hasOld {
			return nil, fmt.Errorf("rule at index %d: missing required field 'old'", i)
		}
		if _, hasNew := rawRule["new"]; !hasNew {
			return nil, fmt.Errorf("rule at index %d: missing required field 'new'", i)
		}
		
		// Convert to Rule struct
		rule := Rule{
			Old: fmt.Sprintf("%v", rawRule["old"]),
			New: fmt.Sprintf("%v", rawRule["new"]),
		}
		
		// Use the Validate method for additional validation
		if err := rule.Validate(); err != nil {
			return nil, fmt.Errorf("rule at index %d: %w", i, err)
		}
		
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// LoadRulesFromFile loads replacement rules from a YAML file
func LoadRulesFromFile(filename string) ([]Rule, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	// Parse YAML
	rules, err := ParseYAMLRules(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rules from %s: %w", filename, err)
	}
	
	return rules, nil
}