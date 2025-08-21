package replace_test

import (
	"fmt"
	"log"

	"github.com/pyhub/pyhub-documents-cli/internal/replace"
)

func ExampleParseYAMLRules() {
	yamlContent := `
- old: "foo"
  new: "bar"
- old: "version 1.0"
  new: "version 2.0"
`
	rules, err := replace.ParseYAMLRules([]byte(yamlContent))
	if err != nil {
		log.Fatal(err)
	}
	
	for _, rule := range rules {
		fmt.Printf("Replace %q with %q\n", rule.Old, rule.New)
	}
	// Output:
	// Replace "foo" with "bar"
	// Replace "version 1.0" with "version 2.0"
}

func ExampleLoadRulesFromFile() {
	// This example assumes you have a file at testdata/valid_rules.yml
	rules, err := replace.LoadRulesFromFile("testdata/valid_rules.yml")
	if err != nil {
		// Handle error in real code
		return
	}
	
	fmt.Printf("Loaded %d replacement rules\n", len(rules))
	// Output would be: Loaded 3 replacement rules
}