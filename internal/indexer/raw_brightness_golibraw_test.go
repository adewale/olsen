//go:build cgo && !use_seppedelanghe_libraw
// +build cgo,!use_seppedelanghe_libraw

package indexer_test

import (
	"os"
	"testing"

	"github.com/adewale/olsen/internal/indexer"
)

// TestGolibrawBrightness tests brightness with inokone/golibraw
func TestGolibrawBrightness(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	t.Logf("Testing with: %s", indexer.LibRawImpl)

	// Decode RAW file
	img, err := indexer.DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("Failed to decode RAW: %v", err)
	}

	// Calculate brightness
	brightness := calculateImageBrightness(img)
	histogram := buildImageHistogram(img)

	t.Logf("Library: %s", indexer.LibRawImpl)
	t.Logf("  Brightness: %.1f/255", brightness)
	t.Logf("  Histogram: %s", formatImageHistogram(histogram))
	t.Logf("  Image bounds: %v", img.Bounds())
	t.Logf("  Image type: %T", img)

	// Check if brightness is reasonable (not completely black)
	if brightness < 1.0 {
		t.Errorf("Image is too dark (brightness %.1f/255) - likely black image bug", brightness)
	} else if brightness > 10.0 {
		t.Logf("✓ Image has reasonable brightness (%.1f/255)", brightness)
	} else {
		t.Logf("⚠ Image is very dark but not completely black (%.1f/255)", brightness)
	}
}
