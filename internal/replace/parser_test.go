package replace

import (
	"testing"
)

func TestParseYAMLRules(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Rule
		wantErr bool
	}{
		{
			name: "valid single rule",
			input: `- old: "version 1.0"
  new: "version 2.0"`,
			want: []Rule{
				{Old: "version 1.0", New: "version 2.0"},
			},
			wantErr: false,
		},
		{
			name: "valid multiple rules",
			input: `- old: "2023"
  new: "2024"
- old: "v1.0.0"
  new: "v2.0.0"
- old: "old company"
  new: "new company"`,
			want: []Rule{
				{Old: "2023", New: "2024"},
				{Old: "v1.0.0", New: "v2.0.0"},
				{Old: "old company", New: "new company"},
			},
			wantErr: false,
		},
		{
			name: "empty YAML",
			input: ``,
			want: []Rule{},
			wantErr: false,
		},
		{
			name: "YAML with comments",
			input: `# This is a comment
- old: "foo"
  new: "bar"
# Another comment
- old: "baz"
  new: "qux"`,
			want: []Rule{
				{Old: "foo", New: "bar"},
				{Old: "baz", New: "qux"},
			},
			wantErr: false,
		},
		{
			name: "invalid YAML structure",
			input: `this is not valid yaml: [`,
			want: nil,
			wantErr: true,
		},
		{
			name: "missing required field 'old'",
			input: `- new: "something"`,
			want: nil,
			wantErr: true,
		},
		{
			name: "missing required field 'new'",
			input: `- old: "something"`,
			want: nil,
			wantErr: true,
		},
		{
			name: "empty old field",
			input: `- old: ""
  new: "something"`,
			want: nil,
			wantErr: true,
		},
		{
			name: "rule with extra fields (should be ignored)",
			input: `- old: "test"
  new: "replaced"
  comment: "this is extra"
  priority: 1`,
			want: []Rule{
				{Old: "test", New: "replaced"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYAMLRules([]byte(tt.input))
			
			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYAMLRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If we expect an error, no need to check the result
			if tt.wantErr {
				return
			}
			
			// Check result length
			if len(got) != len(tt.want) {
				t.Errorf("ParseYAMLRules() returned %d rules, want %d", len(got), len(tt.want))
				return
			}
			
			// Check each rule
			for i := range got {
				if got[i].Old != tt.want[i].Old || got[i].New != tt.want[i].New {
					t.Errorf("ParseYAMLRules() rule[%d] = {Old: %q, New: %q}, want {Old: %q, New: %q}",
						i, got[i].Old, got[i].New, tt.want[i].Old, tt.want[i].New)
				}
			}
		})
	}
}

func TestRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name:    "valid rule",
			rule:    Rule{Old: "old text", New: "new text"},
			wantErr: false,
		},
		{
			name:    "empty old field",
			rule:    Rule{Old: "", New: "new text"},
			wantErr: true,
		},
		{
			name:    "empty new field allowed",
			rule:    Rule{Old: "old text", New: ""},
			wantErr: false,
		},
		{
			name:    "whitespace only in old field",
			rule:    Rule{Old: "   ", New: "new text"},
			wantErr: true,
		},
		{
			name:    "same old and new values",
			rule:    Rule{Old: "same", New: "same"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadRulesFromFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		rules, err := LoadRulesFromFile("testdata/valid_rules.yml")
		if err != nil {
			t.Errorf("LoadRulesFromFile() unexpected error: %v", err)
		}
		if len(rules) != 3 {
			t.Errorf("LoadRulesFromFile() returned %d rules, want 3", len(rules))
		}
		// Check first rule
		if rules[0].Old != "test1" || rules[0].New != "replacement1" {
			t.Errorf("LoadRulesFromFile() first rule = {Old: %q, New: %q}, want {Old: \"test1\", New: \"replacement1\"}", 
				rules[0].Old, rules[0].New)
		}
	})
	
	t.Run("invalid file", func(t *testing.T) {
		rules, err := LoadRulesFromFile("testdata/invalid_rules.yml")
		if err == nil {
			t.Errorf("LoadRulesFromFile() expected error for invalid YAML file")
		}
		if rules != nil {
			t.Errorf("LoadRulesFromFile() returned non-nil rules for invalid file")
		}
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		rules, err := LoadRulesFromFile("testdata/non_existent.yml")
		if err == nil {
			t.Errorf("LoadRulesFromFile() expected error for non-existent file")
		}
		if rules != nil {
			t.Errorf("LoadRulesFromFile() returned non-nil rules for non-existent file")
		}
	})
}