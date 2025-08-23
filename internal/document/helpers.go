package document

import (
	"archive/zip"
	"io"
	"os"
)

// CleanupTempFile safely removes a temporary file if it exists
func CleanupTempFile(path string) {
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
}

// CopyZipFileWithCompression copies a file from source zip to destination zip with consistent compression
func CopyZipFileWithCompression(src *zip.File, dst *zip.Writer, bufferPool []byte) error {
	reader, err := src.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Clone the header to avoid modifying the original
	header := src.FileHeader
	// Use the original compression method if it's valid, otherwise use Deflate
	if header.Method != zip.Store && header.Method != zip.Deflate {
		header.Method = zip.Deflate
	}
	
	writer, err := dst.CreateHeader(&header)
	if err != nil {
		return err
	}

	// Stream copy with buffer
	if bufferPool != nil {
		_, err = io.CopyBuffer(writer, reader, bufferPool)
	} else {
		_, err = io.Copy(writer, reader)
	}
	
	return err
}

// File size thresholds for adaptive streaming
const (
	SmallFileThreshold   = 10 * 1024 * 1024  // 10MB
	MediumFileThreshold  = 50 * 1024 * 1024  // 50MB
	LargeFileThreshold   = 100 * 1024 * 1024 // 100MB
	
	// Chunk sizes for different file sizes
	SmallFileChunkSize   = 32 * 1024   // 32KB
	MediumFileChunkSize  = 64 * 1024   // 64KB
	LargeFileChunkSize   = 128 * 1024  // 128KB
	VeryLargeFileChunkSize = 256 * 1024 // 256KB
	
	// Memory limits for different file sizes
	SmallFileMemoryLimit   = 50 * 1024 * 1024   // 50MB
	MediumFileMemoryLimit  = 100 * 1024 * 1024  // 100MB
	LargeFileMemoryLimit   = 200 * 1024 * 1024  // 200MB
	VeryLargeFileMemoryLimit = 500 * 1024 * 1024 // 500MB
)