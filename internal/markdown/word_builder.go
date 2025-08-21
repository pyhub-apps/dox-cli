package markdown

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// createBasicWordStructure creates a minimal Word document structure
func createBasicWordStructure(path string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	// Add _rels/.rels
	rels, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	
	// Add word/_rels/document.xml.rels
	wordRels, err := w.Create("word/_rels/document.xml.rels")
	if err != nil {
		return err
	}
	wordRels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
	
	// Add word/document.xml with empty body
	doc, err := w.Create("word/document.xml")
	if err != nil {
		return err
	}
	doc.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:r><w:t>Document</w:t></w:r></w:p>
</w:body>
</w:document>`))
	
	// Add [Content_Types].xml
	contentTypes, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))
	
	// Close the zip writer
	if err := w.Close(); err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// WordBuilder helps build Word documents from markdown
type WordBuilder struct {
	paragraphs []string
}

// NewWordBuilder creates a new Word document builder
func NewWordBuilder() *WordBuilder {
	return &WordBuilder{
		paragraphs: []string{},
	}
}

// AddHeading adds a heading to the document
func (b *WordBuilder) AddHeading(level int, text string) {
	// Add paragraph with heading style
	style := fmt.Sprintf("Heading%d", level)
	para := fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="%s"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, 
		style, escapeXML(text))
	b.paragraphs = append(b.paragraphs, para)
}

// AddParagraph adds a paragraph to the document
func (b *WordBuilder) AddParagraph(text string) {
	para := fmt.Sprintf(`<w:p><w:r><w:t>%s</w:t></w:r></w:p>`, escapeXML(text))
	b.paragraphs = append(b.paragraphs, para)
}

// AddList adds a bulleted list to the document
func (b *WordBuilder) AddList(items []string, ordered bool) {
	for i, item := range items {
		var prefix string
		if ordered {
			prefix = fmt.Sprintf("%d. ", i+1)
		} else {
			prefix = "• "
		}
		text := prefix + item
		para := fmt.Sprintf(`<w:p><w:r><w:t>%s</w:t></w:r></w:p>`, escapeXML(text))
		b.paragraphs = append(b.paragraphs, para)
	}
}

// AddCodeBlock adds a code block to the document
func (b *WordBuilder) AddCodeBlock(code string) {
	// Add with monospace font style (Courier New)
	para := fmt.Sprintf(`<w:p><w:r><w:rPr><w:rFonts w:ascii="Courier New" w:hAnsi="Courier New"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, 
		escapeXML(code))
	b.paragraphs = append(b.paragraphs, para)
}

// AddQuote adds a blockquote to the document
func (b *WordBuilder) AddQuote(text string) {
	// Add with indentation
	para := fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="720"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, 
		escapeXML(text))
	b.paragraphs = append(b.paragraphs, para)
}

// Build creates the Word document
func (b *WordBuilder) Build(path string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	// Add _rels/.rels
	rels, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	
	// Add word/_rels/document.xml.rels
	wordRels, err := w.Create("word/_rels/document.xml.rels")
	if err != nil {
		return err
	}
	wordRels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
	
	// Add word/document.xml with paragraphs
	doc, err := w.Create("word/document.xml")
	if err != nil {
		return err
	}
	
	// Build document XML
	var docXML strings.Builder
	docXML.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>`)
	
	for _, para := range b.paragraphs {
		docXML.WriteString(para)
	}
	
	docXML.WriteString(`</w:body>
</w:document>`)
	
	doc.Write([]byte(docXML.String()))
	
	// Add [Content_Types].xml
	contentTypes, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))
	
	// Close the zip writer
	if err := w.Close(); err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}