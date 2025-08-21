package document

// Document defines the common interface for all document types
type Document interface {
	// GetText extracts all text from the document
	GetText() (string, error)
	
	// ReplaceText replaces all occurrences of old text with new text
	ReplaceText(old, new string) error
	
	// Save saves the modified document
	Save() error
	
	// SaveAs saves the document to a new file
	SaveAs(path string) error
	
	// Close closes the document and releases resources
	Close() error
}