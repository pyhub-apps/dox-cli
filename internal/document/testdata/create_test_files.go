// +build ignore

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
)

func main() {
	// Create a minimal valid .docx file
	createSampleDocx()
	// Create a corrupted .docx file  
	createCorruptedDocx()
	// Create a text file with wrong extension
	createTextFile()
	// Create an empty .docx file
	createEmptyDocx()
	// Create a unicode .docx file
	createUnicodeDocx()
}

func createSampleDocx() {
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
<w:p><w:r><w:t>This is a sample document</w:t></w:r></w:p>
<w:p><w:r><w:t>Second paragraph with some text</w:t></w:r></w:p>
<w:p><w:r><w:t>Third paragraph</w:t></w:r></w:p>
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
	
	err := os.WriteFile("sample.docx", buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error creating sample.docx: %v\n", err)
	} else {
		fmt.Println("Created sample.docx")
	}
}

func createCorruptedDocx() {
	// Create a file that looks like a zip but has invalid structure
	err := os.WriteFile("corrupted.docx", []byte("This is not a valid docx file"), 0644)
	if err != nil {
		fmt.Printf("Error creating corrupted.docx: %v\n", err)
	} else {
		fmt.Println("Created corrupted.docx")
	}
}

func createTextFile() {
	err := os.WriteFile("sample.txt", []byte("This is a text file"), 0644)
	if err != nil {
		fmt.Printf("Error creating sample.txt: %v\n", err)
	} else {
		fmt.Println("Created sample.txt")
	}
}

func createEmptyDocx() {
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
	
	// Add word/document.xml with empty body
	doc, _ := w.Create("word/document.xml")
	doc.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
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
	
	err := os.WriteFile("empty.docx", buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error creating empty.docx: %v\n", err)
	} else {
		fmt.Println("Created empty.docx")
	}
}

func createUnicodeDocx() {
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
	
	// Add word/document.xml with unicode content
	doc, _ := w.Create("word/document.xml")
	doc.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:r><w:t>Hello, 世界! 👋</w:t></w:r></w:p>
<w:p><w:r><w:t>한글 텍스트 테스트</w:t></w:r></w:p>
<w:p><w:r><w:t>Café naïve fiancé</w:t></w:r></w:p>
<w:p><w:r><w:t>Emoji: 😃🚀🌟</w:t></w:r></w:p>
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
	
	err := os.WriteFile("unicode.docx", buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error creating unicode.docx: %v\n", err)
	} else {
		fmt.Println("Created unicode.docx")
	}
}