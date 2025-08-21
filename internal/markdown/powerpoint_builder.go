package markdown

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Slide represents a PowerPoint slide
type Slide struct {
	Title   string
	Content []string
	Bullets []string
}

// PowerPointBuilder helps build PowerPoint presentations from markdown
type PowerPointBuilder struct {
	slides []Slide
}

// NewPowerPointBuilder creates a new PowerPoint builder
func NewPowerPointBuilder() *PowerPointBuilder {
	return &PowerPointBuilder{
		slides: []Slide{},
	}
}

// AddTitleSlide adds a title slide
func (b *PowerPointBuilder) AddTitleSlide(title, subtitle string) {
	slide := Slide{
		Title: title,
	}
	if subtitle != "" {
		slide.Content = []string{subtitle}
	}
	b.slides = append(b.slides, slide)
}

// AddContentSlide adds a content slide
func (b *PowerPointBuilder) AddContentSlide(slide *Slide) {
	if slide != nil {
		b.slides = append(b.slides, *slide)
	}
}

// Build creates the PowerPoint presentation
func (b *PowerPointBuilder) Build(path string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	// Add _rels/.rels
	rels, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
</Relationships>`))
	
	// Add ppt/_rels/presentation.xml.rels
	pptRels, err := w.Create("ppt/_rels/presentation.xml.rels")
	if err != nil {
		return err
	}
	
	// Build relationships for slides
	var relsBuilder strings.Builder
	relsBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	
	for i := range b.slides {
		relsBuilder.WriteString(fmt.Sprintf(`
<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`,
			i+1, i+1))
	}
	
	relsBuilder.WriteString(`
</Relationships>`)
	pptRels.Write([]byte(relsBuilder.String()))
	
	// Add ppt/presentation.xml
	presentation, err := w.Create("ppt/presentation.xml")
	if err != nil {
		return err
	}
	
	var presBuilder strings.Builder
	presBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:sldIdLst>`)
	
	for i := range b.slides {
		presBuilder.WriteString(fmt.Sprintf(`
<p:sldId id="%d" r:id="rId%d"/>`, 256+i, i+1))
	}
	
	presBuilder.WriteString(`
</p:sldIdLst>
</p:presentation>`)
	presentation.Write([]byte(presBuilder.String()))
	
	// Add slides
	for i, slide := range b.slides {
		slideNum := i + 1
		
		// Add slide relationship
		slideRels, err := w.Create(fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideNum))
		if err != nil {
			return err
		}
		slideRels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
		
		// Add slide content
		slideFile, err := w.Create(fmt.Sprintf("ppt/slides/slide%d.xml", slideNum))
		if err != nil {
			return err
		}
		
		slideXML := b.buildSlideXML(slide)
		slideFile.Write([]byte(slideXML))
	}
	
	// Add [Content_Types].xml
	contentTypes, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	
	var ctBuilder strings.Builder
	ctBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>`)
	
	for i := range b.slides {
		ctBuilder.WriteString(fmt.Sprintf(`
<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`,
			i+1))
	}
	
	ctBuilder.WriteString(`
</Types>`)
	contentTypes.Write([]byte(ctBuilder.String()))
	
	// Close the zip writer
	if err := w.Close(); err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// buildSlideXML builds XML for a slide
func (b *PowerPointBuilder) buildSlideXML(slide Slide) string {
	var xml strings.Builder
	
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>`)
	
	// Add title if present
	if slide.Title != "" {
		xml.WriteString(`
<p:sp>
<p:txBody>
<a:p>
<a:r>
<a:t>`)
		xml.WriteString(escapeXML(slide.Title))
		xml.WriteString(`</a:t>
</a:r>
</a:p>
</p:txBody>
</p:sp>`)
	}
	
	// Add content paragraphs
	if len(slide.Content) > 0 || len(slide.Bullets) > 0 {
		xml.WriteString(`
<p:sp>
<p:txBody>`)
		
		// Add regular content
		for _, content := range slide.Content {
			xml.WriteString(`
<a:p>
<a:r>
<a:t>`)
			xml.WriteString(escapeXML(content))
			xml.WriteString(`</a:t>
</a:r>
</a:p>`)
		}
		
		// Add bullet points
		for _, bullet := range slide.Bullets {
			xml.WriteString(`
<a:p>
<a:r>
<a:t>`)
			xml.WriteString(escapeXML(bullet))
			xml.WriteString(`</a:t>
</a:r>
</a:p>`)
		}
		
		xml.WriteString(`
</p:txBody>
</p:sp>`)
	}
	
	xml.WriteString(`
</p:spTree>
</p:cSld>
</p:sld>`)
	
	return xml.String()
}