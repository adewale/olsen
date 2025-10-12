//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"os"
	"testing"

	"github.com/adewale/olsen/internal/indexer"
)

// TestEmbeddedJPEGBrightness tests if embedded JPEG previews have correct brightness
func TestEmbeddedJPEGBrightness(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	// Try to extract embedded JPEG
	jpegImg, err := indexer.ExtractEmbeddedJPEG(testFile)
	if err != nil {
		t.Fatalf("Failed to extract embedded JPEG: %v", err)
	}

	brightness := calculateImageBrightness(jpegImg)
	histogram := buildImageHistogram(jpegImg)

	t.Logf("Embedded JPEG Preview:")
	t.Logf("  Brightness: %.1f/255", brightness)
	t.Logf("  Histogram: %s", formatImageHistogram(histogram))
	t.Logf("  Image bounds: %v", jpegImg.Bounds())
	t.Logf("  Image type: %T", jpegImg)

	// Check if embedded JPEG is better than LibRaw output
	if brightness > 10 {
		t.Logf("✓ Embedded JPEG has reasonable brightness (%.1f/255)", brightness)
	} else {
		t.Errorf("Embedded JPEG is also too dark (%.1f/255)", brightness)
	}
}
