// +build ignore

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
)

func main() {
	createSampleDocument()
}

func createSampleDocument() {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	// Add _rels/.rels
	rels, _ := w.Create("_rels/.rels")
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	
	// Add word/_rels/document.xml.rels
	docRels, _ := w.Create("word/_rels/document.xml.rels")
	docRels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
	
	// Add word/document.xml with sample content
	doc, _ := w.Create("word/document.xml")
	doc.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:r><w:t>Document Title - Version 1.0</w:t></w:r></w:p>
<w:p><w:r><w:t>Status: Draft</w:t></w:r></w:p>
<w:p><w:r><w:t>Year: 2023</w:t></w:r></w:p>
<w:p><w:r><w:t>This is a sample document for testing the replacement functionality.</w:t></w:r></w:p>
<w:p><w:r><w:t>It contains various placeholders that need to be replaced.</w:t></w:r></w:p>
<w:p><w:r><w:t>Copyright 2023 - All rights reserved</w:t></w:r></w:p>
</w:body>
</w:document>`))
	
	// Add [Content_Types].xml
	contentTypes, _ := w.Create("[Content_Types].xml")
	contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))
	
	w.Close()
	
	err := os.WriteFile("sample_document.docx", buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error creating sample_document.docx: %v\n", err)
	} else {
		fmt.Println("Created sample_document.docx")
	}
}