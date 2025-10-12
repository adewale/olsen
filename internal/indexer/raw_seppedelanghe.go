//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	golibraw "github.com/seppedelanghe/go-libraw"
)

// LibRawImpl identifies which library implementation is in use
const LibRawImpl = "seppedelanghe/go-libraw"

// DecodeRaw attempts to decode a RAW image file using LibRaw (seppedelanghe/go-libraw)
// Returns image.Image on success, error on failure
// Automatically falls back to embedded JPEG if LibRaw produces black images
func DecodeRaw(path string) (image.Image, error) {
	basename := filepath.Base(path)

	// Use high-quality settings for RAW decode
	processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
		UserQual:    3, // AHD demosaicing (highest quality)
		OutputBps:   8, // 8-bit output (sufficient for thumbnails)
		OutputColor: golibraw.SRGB,
		UseCameraWb: true, // Use camera white balance
	})

	img, _, err := processor.ProcessRaw(path)
	if err != nil {
		// LibRaw decode failed, try embedded JPEG as fallback
		log.Printf("[RAW] LibRaw decode failed for %s: %v (falling back to embedded JPEG)", basename, err)
		jpegImg, jpegErr := ExtractEmbeddedJPEG(path)
		if jpegErr == nil {
			log.Printf("[RAW] Successfully used embedded JPEG fallback for %s", basename)
			return jpegImg, nil
		}
		log.Printf("[RAW] ERROR: Both LibRaw and embedded JPEG extraction failed for %s", basename)
		return nil, fmt.Errorf("libraw decode failed: %w (embedded jpeg: %v)", err, jpegErr)
	}

	// Log successful LibRaw decode
	bounds := img.Bounds()
	log.Printf("[RAW] LibRaw decoded %s: %dx%d (type: %T)", basename, bounds.Dx(), bounds.Dy(), img)

	// Check if LibRaw produced a black/unusable image
	// This happens with JPEG-compressed monochrome DNGs
	if isBlackImage(img) {
		log.Printf("[RAW] WARNING: LibRaw returned black image for %s (known issue with JPEG-compressed monochrome DNGs)", basename)
		jpegImg, jpegErr := ExtractEmbeddedJPEG(path)
		if jpegErr == nil {
			log.Printf("[RAW] Successfully used embedded JPEG fallback after black image detection for %s", basename)
			return jpegImg, nil
		}
		log.Printf("[RAW] WARNING: Embedded JPEG extraction also failed for %s, returning black image", basename)
		// If embedded JPEG fails, return the black image rather than error
		// (at least we have dimensions)
	}

	return img, nil
}

// isBlackImage checks if an image is completely black (all pixels near 0)
// Used to detect the JPEG-compressed monochrome DNG bug
func isBlackImage(img image.Image) bool {
	bounds := img.Bounds()
	// Sample 100 pixels across the image
	sampleCount := 0
	brightPixels := 0

	stepX := bounds.Dx() / 10
	stepY := bounds.Dy() / 10
	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y && sampleCount < 100; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X && sampleCount < 100; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit
			gray := (r + g + b) / 3 / 256
			if gray > 5 { // Any pixel brighter than 5/255
				brightPixels++
			}
			sampleCount++
		}
	}

	// Image is "black" if fewer than 5% of sampled pixels are bright
	return brightPixels < 5
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
	var largestDimensions string
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
						cfg, err := jpeg.DecodeConfig(bytes.NewReader(jpegData))
						if err == nil {
							largestJPEG = jpegData
							largestSize = jpegSize
							largestDimensions = fmt.Sprintf("%dx%d", cfg.Width, cfg.Height)
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
	log.Printf("[EMBED] Extracted largest embedded JPEG: %s (%d bytes) from %d previews in %s",
		largestDimensions, largestSize, jpegCount, filepath.Base(path))

	return img, nil
}
