//go:build cgo && !use_seppedelanghe_libraw
// +build cgo,!use_seppedelanghe_libraw

package indexer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	golibraw "github.com/inokone/golibraw"
)

// LibRawImpl identifies which library implementation is in use
const LibRawImpl = "inokone/golibraw"

// DecodeRaw attempts to decode a RAW image file using LibRaw (inokone/golibraw)
// Returns image.Image on success, error on failure
func DecodeRaw(path string) (image.Image, error) {
	img, err := golibraw.ImportRaw(path)
	if err != nil {
		return nil, fmt.Errorf("libraw decode failed: %w", err)
	}
	return img, nil
}

// IsRawSupported returns true if LibRaw support is compiled in
func IsRawSupported() bool {
	return true
}

// ExtractEmbeddedJPEG extracts the embedded JPEG preview from a DNG file
// DNG files (which are TIFF-based) often contain embedded JPEG previews
// This function finds the LARGEST valid embedded JPEG (best quality)
func ExtractEmbeddedJPEG(path string) (image.Image, error) {
	// Read the entire file for JPEG marker search
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Look for JPEG markers (0xFFD8 = JPEG Start of Image, 0xFFD9 = End of Image)
	// DNG files typically contain multiple embedded JPEG previews
	// We want the largest one for best quality

	var largestJPEG []byte
	var largestSize int
	jpegCount := 0

	for i := 0; i < len(data)-1; i++ {
		// Look for JPEG SOI (Start of Image) marker
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			jpegStart := i
			// Look for corresponding EOI (End of Image) marker
			for j := jpegStart + 2; j < len(data)-1; j++ {
				if data[j] == 0xFF && data[j+1] == 0xD9 {
					jpegEnd := j + 2
					jpegSize := jpegEnd - jpegStart
					jpegCount++

					// Keep track of the largest JPEG found
					if jpegSize > largestSize {
						jpegData := data[jpegStart:jpegEnd]
						// Verify it's valid by attempting to decode
						_, err := jpeg.DecodeConfig(bytes.NewReader(jpegData))
						if err == nil {
							largestJPEG = jpegData
							largestSize = jpegSize
						}
					}

					i = jpegEnd - 1 // Skip past this JPEG
					break
				}
			}
		}
	}

	if largestJPEG == nil {
		return nil, fmt.Errorf("no valid embedded JPEG preview found in DNG file")
	}

	// Decode the largest JPEG
	img, err := jpeg.Decode(bytes.NewReader(largestJPEG))
	if err != nil {
		return nil, fmt.Errorf("failed to decode largest embedded JPEG: %w", err)
	}

	// Log diagnostic information
	cfg, _ := jpeg.DecodeConfig(bytes.NewReader(largestJPEG))
	log.Printf("[EMBED] Extracted largest embedded JPEG: %dx%d (%d bytes) from %d previews in %s",
		cfg.Width, cfg.Height, largestSize, jpegCount, filepath.Base(path))

	return img, nil
}
