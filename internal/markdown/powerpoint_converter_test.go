package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPowerPointConverter(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
	}{
		{
			name: "simple presentation",
			markdown: `# Slide 1

Content for slide 1

# Slide 2

Content for slide 2`,
			wantErr: false,
		},
		{
			name: "presentation with subsections",
			markdown: `# Main Title

## Subtitle 1

Content under subtitle

## Subtitle 2

More content

# Second Slide

## Another subtitle

Final content`,
			wantErr: false,
		},
		{
			name: "presentation with lists",
			markdown: `# Bullet Points

- Point 1
- Point 2
- Point 3

# Numbered List

1. First
2. Second
3. Third`,
			wantErr: false,
		},
		{
			name: "presentation with code",
			markdown: `# Code Slide

Here's some code:

` + "```python" + `
print("Hello")
` + "```",
			wantErr: false,
		},
		{
			name: "single slide",
			markdown: `# Only One Slide

With some content`,
			wantErr: false,
		},
		{
			name: "no H1 headers",
			markdown: `## Subtitle Only

Content without H1`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test output
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "test.pptx")

			// Parse markdown
			doc, err := Parse([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("Failed to parse markdown: %v", err)
			}

			// Create converter
			converter := NewPowerPointConverter()

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

func TestPowerPointConverterH2Handling(t *testing.T) {
	// Test specifically for H2 -> slide title mapping
	markdown := `# Presentation Title

Introduction content

## First Topic

Details about first topic

## Second Topic  

Details about second topic

# Next Slide

## Subtitle in Next Slide

Content under subtitle`

	// Parse markdown
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	// Create converter and convert
	converter := NewPowerPointConverter()
	err = converter.Convert(doc)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Verify we have the expected number of slides
	// H1 creates slides, H2 becomes content within slides
	// We have 2 H1 headers, so we should have 2 main slides
	// Plus any additional content slides from convertBlocksToSlides
	if len(converter.builder.slides) < 2 {
		t.Errorf("Expected at least 2 slides, got %d", len(converter.builder.slides))
	}

	// Verify first slide has title from H1
	if converter.builder.slides[0].Title != "Presentation Title" {
		t.Errorf("Expected first slide title 'Presentation Title', got '%s'", converter.builder.slides[0].Title)
	}

	// Save to temp file for validation
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "h2_test.pptx")
	err = converter.SaveAs(outputPath)
	if err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}
}

func TestPowerPointConverterComplexPresentation(t *testing.T) {
	markdown := `# Title Slide

Welcome to the presentation

# Agenda

## Topics to Cover

- Introduction
- Main Content
- Conclusion

# Introduction

## What is this about?

This presentation covers:
1. First point
2. Second point
3. Third point

# Main Content

## Key Features

- Feature A
- Feature B
- Feature C

## Code Example

` + "```go" + `
func example() {
    fmt.Println("Demo")
}
` + "```" + `

# Conclusion

## Summary

> Important takeaway

Thank you!`

	// Parse markdown
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	// Create converter and convert
	converter := NewPowerPointConverter()
	err = converter.Convert(doc)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Should have at least 5 slides (5 H1 headers create main slides)
	// Additional slides may be created from content blocks
	if len(converter.builder.slides) < 5 {
		t.Errorf("Expected at least 5 slides, got %d", len(converter.builder.slides))
	}

	// Save to temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "complex.pptx")
	err = converter.SaveAs(outputPath)
	if err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}

	// Verify file
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	
	// A complex presentation should be reasonably sized (at least 4KB)
	if info.Size() < 4000 {
		t.Errorf("Complex presentation seems too small: %d bytes", info.Size())
	}
}