package markdown

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		wantBlocks  int
		wantSections int
	}{
		{
			name: "simple markdown",
			markdown: `# Title

This is a paragraph.

## Subtitle

- Item 1
- Item 2`,
			wantBlocks:   4, // H1, paragraph, H2, list
			wantSections: 1, // One H1 section
		},
		{
			name: "multiple sections",
			markdown: `# Section 1

Content 1

# Section 2

Content 2`,
			wantBlocks:   4, // H1, para, H1, para
			wantSections: 2, // Two H1 sections
		},
		{
			name: "code block",
			markdown: "# Title\n\n```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```",
			wantBlocks:   2, // H1, code block
			wantSections: 1,
		},
		{
			name: "ordered list",
			markdown: `# Title

1. First
2. Second
3. Third`,
			wantBlocks:   2, // H1, ordered list
			wantSections: 1,
		},
		{
			name: "blockquote",
			markdown: `# Title

> This is a quote`,
			wantBlocks:   2, // H1, blockquote
			wantSections: 1,
		},
		{
			name: "no H1 sections",
			markdown: `## Subtitle

Some content here.`,
			wantBlocks:   2, // H2, paragraph
			wantSections: 1, // Default section created
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			if len(doc.Blocks) != tt.wantBlocks {
				t.Errorf("Parse() got %d blocks, want %d", len(doc.Blocks), tt.wantBlocks)
			}
			
			if len(doc.Sections) != tt.wantSections {
				t.Errorf("Parse() got %d sections, want %d", len(doc.Sections), tt.wantSections)
			}
		})
	}
}

func TestExtractText(t *testing.T) {
	markdown := `# Main Title

This is a **bold** text and *italic* text.

## Subtitle with ` + "`code`" + `

Another paragraph.`
	
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	// Check heading extraction
	if doc.Blocks[0].Type != BlockHeading || doc.Blocks[0].Content != "Main Title" {
		t.Errorf("Expected heading 'Main Title', got %v", doc.Blocks[0].Content)
	}
	
	// Check paragraph with inline formatting
	if doc.Blocks[1].Type != BlockParagraph {
		t.Errorf("Expected paragraph, got %v", doc.Blocks[1].Type)
	}
	
	// Text should be extracted without markdown formatting
	if !strings.Contains(doc.Blocks[1].Content, "bold") {
		t.Errorf("Expected text to contain 'bold', got %v", doc.Blocks[1].Content)
	}
}

func TestListExtraction(t *testing.T) {
	markdown := `# Title

- Item 1
- Item 2
- Item 3

1. First
2. Second`
	
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	// Find unordered list
	var unorderedList *Block
	for i, block := range doc.Blocks {
		if block.Type == BlockList {
			unorderedList = &doc.Blocks[i]
			break
		}
	}
	
	if unorderedList == nil {
		t.Fatal("Expected to find unordered list")
	}
	
	if len(unorderedList.Items) != 3 {
		t.Errorf("Expected 3 items in unordered list, got %d", len(unorderedList.Items))
	}
	
	// Find ordered list
	var orderedList *Block
	for i, block := range doc.Blocks {
		if block.Type == BlockOrderedList {
			orderedList = &doc.Blocks[i]
			break
		}
	}
	
	if orderedList == nil {
		t.Fatal("Expected to find ordered list")
	}
	
	if len(orderedList.Items) != 2 {
		t.Errorf("Expected 2 items in ordered list, got %d", len(orderedList.Items))
	}
}

func TestSectionSplitting(t *testing.T) {
	markdown := `# First Slide

Content for first slide.

## Subtitle in first

More content.

# Second Slide

Content for second slide.

- Bullet 1
- Bullet 2`
	
	doc, err := Parse([]byte(markdown))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	if len(doc.Sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(doc.Sections))
	}
	
	// Check first section
	if doc.Sections[0].Title != "First Slide" {
		t.Errorf("Expected first section title 'First Slide', got %v", doc.Sections[0].Title)
	}
	
	if len(doc.Sections[0].Blocks) != 3 { // paragraph, H2, paragraph
		t.Errorf("Expected 3 blocks in first section, got %d", len(doc.Sections[0].Blocks))
	}
	
	// Check second section
	if doc.Sections[1].Title != "Second Slide" {
		t.Errorf("Expected second section title 'Second Slide', got %v", doc.Sections[1].Title)
	}
	
	if len(doc.Sections[1].Blocks) != 2 { // paragraph, list
		t.Errorf("Expected 2 blocks in second section, got %d", len(doc.Sections[1].Blocks))
	}
}