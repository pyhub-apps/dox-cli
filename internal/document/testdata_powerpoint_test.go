package document

import (
	"archive/zip"
	"bytes"
	"os"
)

// createTestPowerPoint creates a simple PowerPoint file for testing
func createTestPowerPoint(path string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	// Add _rels/.rels
	rels, _ := w.Create("_rels/.rels")
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
</Relationships>`))
	
	// Add ppt/_rels/presentation.xml.rels
	pptRels, _ := w.Create("ppt/_rels/presentation.xml.rels")
	pptRels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml"/>
</Relationships>`))
	
	// Add ppt/presentation.xml
	presentation, _ := w.Create("ppt/presentation.xml")
	presentation.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:sldIdLst>
<p:sldId id="256" r:id="rId1"/>
<p:sldId id="257" r:id="rId2"/>
</p:sldIdLst>
</p:presentation>`))
	
	// Add ppt/slides/_rels/slide1.xml.rels
	slide1Rels, _ := w.Create("ppt/slides/_rels/slide1.xml.rels")
	slide1Rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
	
	// Add ppt/slides/slide1.xml with sample content
	slide1, _ := w.Create("ppt/slides/slide1.xml")
	slide1.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:sp>
<p:txBody>
<a:p>
<a:r>
<a:t>Presentation Title - Version 1.0</a:t>
</a:r>
</a:p>
<a:p>
<a:r>
<a:t>Status: Draft</a:t>
</a:r>
</a:p>
</p:txBody>
</p:sp>
</p:spTree>
</p:cSld>
</p:sld>`))

	// Add ppt/slides/_rels/slide2.xml.rels
	slide2Rels, _ := w.Create("ppt/slides/_rels/slide2.xml.rels")
	slide2Rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))
	
	// Add ppt/slides/slide2.xml with sample content
	slide2, _ := w.Create("ppt/slides/slide2.xml")
	slide2.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:sp>
<p:txBody>
<a:p>
<a:r>
<a:t>Year: 2023</a:t>
</a:r>
</a:p>
<a:p>
<a:r>
<a:t>Copyright 2023 - All rights reserved</a:t>
</a:r>
</a:p>
</p:txBody>
</p:sp>
</p:spTree>
</p:cSld>
</p:sld>`))
	
	// Add [Content_Types].xml
	contentTypes, _ := w.Create("[Content_Types].xml")
	contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
<Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
<Override PartName="/ppt/slides/slide2.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
</Types>`))
	
	w.Close()
	
	return os.WriteFile(path, buf.Bytes(), 0644)
}