package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWordConverter(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name: "simple document",
			markdown: `# Title

This is a paragraph.

## Subtitle

Another paragraph.`,
			wantErr: false,
		},
		{
			name: "document with lists",
			markdown: `# Document

## Unordered List
- Item 1
- Item 2
- Item 3

## Ordered List
1. First
2. Second
3. Third`,
			wantErr: false,
		},
		{
			name: "document with code block",
			markdown: `# Code Example

Here's some code:

` + "```go" + `
func main() {
	fmt.Println("Hello, World!")
}
` + "```",
			wantErr: false,
		},
		{
			name: "document with blockquote",
			markdown: `# Quote Example

> This is a quote
> spanning multiple lines`,
			wantErr: false,
		},
		{
			name:     "empty document",
			markdown: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test output
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "test.docx")

			// Parse markdown
			doc, err := Parse([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("Failed to parse markdown: %v", err)
			}

			// Create converter
			converter := NewWordConverter()

			// Convert document
			err = converter.Convert(doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Save document
				err = converter.SaveAs(outputPath)
				if err != nil {
					t.Errorf("SaveAs() error = %v", err)
					return
				}

				// Verify file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Output file was not created")
				}

				// Verify file is valid (has content)
				info, err := os.Stat(outputPath)
				if err != nil {
					t.Errorf("Failed to stat output file: %v", err)
				} else if info.Size() == 0 {
					t.Error("Output file is empty")
				}
			}
		})
	}
}

func TestWordConverterComplexDocument(t *testing.T) {
	markdown := `# Main Title

This is the introduction paragraph with some text.

## Section 1

First section content.

### Subsection 1.1

Some details here.

- Bullet point 1
- Bullet point 2
- Bullet point 3

## Section 2

Second section with ordered list:

1. First item
2. Second item
3. Third item

### Code Example

Here's a code block:

` + "```python" + `
def hello_world():
    print("Hello, World!")
` + "```" + `

## Conclusion

> Important quote here

Final thoughts.`

	// Parse markdown
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	// Create converter and convert
	converter := NewWordConverter()
	err = converter.Convert(doc)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Save to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "complex.docx")
	err = converter.SaveAs(outputPath)
	if err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	// Verify file
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	
	// A complex document should be larger than a simple one
	if info.Size() < 1000 {
		t.Errorf("Complex document seems too small: %d bytes", info.Size())
	}
}