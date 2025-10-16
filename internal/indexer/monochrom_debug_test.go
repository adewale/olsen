//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer

import (
	"testing"
)

// TestMonochromRAWDecode tests RAW decoding of Monochrom DNGs to verify dimensions
func TestMonochromRAWDecode(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	t.Logf("Testing RAW decode on Monochrom DNG")
	t.Logf("Using library: %s", LibRawImpl)

	// Test DecodeRaw
	t.Run("DecodeRaw full resolution", func(t *testing.T) {
		img, err := DecodeRaw(testFile)
		if err != nil {
			t.Fatalf("DecodeRaw failed: %v", err)
		}

		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		t.Logf("Decoded image: %dx%d (type: %T)", width, height, img)

		// LibRaw fails on JPEG-compressed Monochrom DNGs, falls back to embedded JPEG
		// Full RAW: 9536x6336, Embedded JPEG: 9504x6320
		if (width == 9536 && height == 6336) || (width == 9504 && height == 6320) {
			t.Logf("✓ Got expected dimensions (RAW or embedded JPEG fallback)")
		} else {
			t.Errorf("Expected 9536x6336 (RAW) or 9504x6320 (embedded JPEG), got %dx%d", width, height)
		}
	})

	// Test ExtractEmbeddedJPEG
	t.Run("ExtractEmbeddedJPEG fallback", func(t *testing.T) {
		jpgImg, err := ExtractEmbeddedJPEG(testFile)
		if err != nil {
			t.Fatalf("ExtractEmbeddedJPEG failed: %v", err)
		}

		bounds := jpgImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		t.Logf("Embedded JPEG: %dx%d (type: %T)", width, height, jpgImg)

		// The embedded JPEG is typically smaller
		if width > 1000 || height > 1000 {
			t.Logf("Warning: Embedded JPEG is unexpectedly large: %dx%d", width, height)
		}
	})
}

// TestMonochromThumbnailGeneration tests the complete thumbnail pipeline
func TestMonochromThumbnailGeneration(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	t.Logf("Testing complete thumbnail generation pipeline for Monochrom DNG")

	// Decode the RAW
	img, err := DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
	}

	bounds := img.Bounds()
	t.Logf("Decoded image: %dx%d", bounds.Dx(), bounds.Dy())

	// Generate thumbnails using the old GenerateThumbnailsFromImage function
	// (this doesn't use the quality pipeline)
	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	t.Logf("Generated %d thumbnails", len(thumbnails))

	for size, data := range thumbnails {
		t.Logf("  - %s: %d bytes", size, len(data))
	}

	// Should generate all 4 sizes for a 9536x6336 image
	if len(thumbnails) != 4 {
		t.Errorf("Expected 4 thumbnails, got %d", len(thumbnails))
		t.Logf("This suggests the quality pipeline is skipping larger sizes")
	}
}
