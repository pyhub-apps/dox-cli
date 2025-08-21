package replace

import (
	"errors"
	"strings"
)

// Rule represents a text replacement rule
type Rule struct {
	Old string `yaml:"old" json:"old"`
	New string `yaml:"new" json:"new"`
}

// Validate checks if the rule is valid
func (r Rule) Validate() error {
	// Check if Old field is empty or whitespace only
	if strings.TrimSpace(r.Old) == "" {
		return errors.New("old field cannot be empty")
	}
	
	// Check if Old and New are the same
	if r.Old == r.New {
		return errors.New("old and new values cannot be the same")
	}
	
	return nil
}