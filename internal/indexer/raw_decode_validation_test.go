//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer

import (
	"os"
	"testing"

	"github.com/adewale/olsen/pkg/models"
)

// TestExtractEmbeddedJPEG_FindsLargest verifies we extract the LARGEST embedded JPEG
// LESSON: This test would have caught the bug where we returned the first (160x120) preview
// instead of the largest (9536x6336) preview
func TestExtractEmbeddedJPEG_FindsLargest(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	img, err := ExtractEmbeddedJPEG(testFile)
	if err != nil {
		t.Fatalf("ExtractEmbeddedJPEG failed: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	t.Logf("Extracted embedded JPEG: %dx%d", width, height)

	// The largest embedded JPEG should be full resolution or close to it
	// For Leica M11 Monochrom, this is ~9500x6300
	minDimension := 6000 // Must be at least 6000px on one edge
	maxDimension := max(width, height)

	if maxDimension < minDimension {
		t.Errorf("Expected large embedded JPEG (>%dpx), got %dx%d", minDimension, width, height)
		t.Errorf("This suggests we extracted a small preview instead of the largest one")
		t.Errorf("BUG: ExtractEmbeddedJPEG is returning the FIRST JPEG, not the LARGEST")
	}

	// Should NOT be the tiny 160x120 preview
	if width <= 200 || height <= 200 {
		t.Errorf("Got tiny preview (%dx%d), expected full-resolution embedded JPEG", width, height)
		t.Errorf("CRITICAL BUG: Extracting thumbnail-sized preview instead of full-resolution JPEG")
	}
}

// TestExtractEmbeddedJPEG_NotBlack verifies the extracted JPEG is not black
// LESSON: This test would have caught the issue where LibRaw returned black images
// and we weren't falling back to embedded JPEG
func TestExtractEmbeddedJPEG_NotBlack(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	img, err := ExtractEmbeddedJPEG(testFile)
	if err != nil {
		t.Fatalf("ExtractEmbeddedJPEG failed: %v", err)
	}

	if isBlackImage(img) {
		t.Error("Extracted embedded JPEG is completely black - this should NEVER happen")
		t.Error("The embedded JPEG should contain visible image data")
		t.Error("BUG: Either extraction failed or the DNG has no valid embedded JPEG")
	}
}

// TestDecodeRaw_FallsBackToEmbeddedJPEG verifies LibRaw falls back correctly
// LESSON: This documents the known limitation that LibRaw cannot decode JPEG-compressed monochrome DNGs
func TestDecodeRaw_FallsBackToEmbeddedJPEG(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	img, err := DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("DecodeRaw failed (even with fallback): %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	t.Logf("DecodeRaw returned: %dx%d (type: %T)", width, height, img)

	// Should get full resolution (from embedded JPEG fallback)
	if width < 6000 || height < 6000 {
		t.Errorf("Expected full resolution, got %dx%d", width, height)
		t.Errorf("BUG: Fallback to embedded JPEG is not working or returning small preview")
	}

	// Image should not be black
	if isBlackImage(img) {
		t.Error("DecodeRaw returned black image")
		t.Error("Expected: Fallback to embedded JPEG should provide visible image")
		t.Error("BUG: isBlackImage check is not triggering fallback, or fallback is broken")
	}
}

// TestDecodeRaw_QualityCheck verifies decoded images are usable for thumbnails
// LESSON: This test would have caught the brightness/quality issues early
func TestDecodeRaw_QualityCheck(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	img, err := DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
	}

	// Calculate average brightness
	bounds := img.Bounds()
	var totalBrightness uint64
	pixelCount := 0

	// Sample 1000 pixels
	stepX := bounds.Dx() / 100
	stepY := bounds.Dy() / 100
	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit and calculate grayscale
			gray := (r + g + b) / 3 / 256
			totalBrightness += uint64(gray)
			pixelCount++
		}
	}

	avgBrightness := float64(totalBrightness) / float64(pixelCount)
	t.Logf("Average brightness: %.2f/255", avgBrightness)

	// Image should have reasonable brightness (not too dark, not overexposed)
	if avgBrightness < 10 {
		t.Errorf("Image too dark (avg brightness: %.2f). Expected >10", avgBrightness)
		t.Error("BUG: Image is nearly black - check LibRaw settings or fallback")
	}
	if avgBrightness > 250 {
		t.Errorf("Image overexposed (avg brightness: %.2f). Expected <250", avgBrightness)
		t.Error("BUG: Image is nearly white - check brightness settings")
	}

	// Image should have some dynamic range (not flat)
	var variance float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := float64((r+g+b)/3/256) - avgBrightness
			variance += gray * gray
		}
	}
	variance /= float64(pixelCount)

	t.Logf("Brightness variance: %.2f", variance)

	if variance < 25 {
		t.Errorf("Image has no dynamic range (variance: %.2f). Expected >25", variance)
		t.Error("BUG: Image is flat/uniform with no detail - check decode quality")
	}
}

// TestThumbnailGeneration_FromMonochromDNG verifies end-to-end thumbnail generation
// LESSON: This integration test would have caught the "only 64px generated" problem immediately
func TestThumbnailGeneration_FromMonochromDNG(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	// Decode the RAW
	img, err := DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
	}

	bounds := img.Bounds()
	t.Logf("Decoded image: %dx%d", bounds.Dx(), bounds.Dy())

	// Generate thumbnails
	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	t.Logf("Generated %d thumbnails", len(thumbnails))
	for size, data := range thumbnails {
		t.Logf("  - %s: %d bytes", size, len(data))
	}

	// Should generate ALL 4 thumbnail sizes
	expectedSizes := []models.ThumbnailSize{
		models.ThumbnailTiny,
		models.ThumbnailSmall,
		models.ThumbnailMedium,
		models.ThumbnailLarge,
	}
	for _, size := range expectedSizes {
		if _, exists := thumbnails[size]; !exists {
			t.Errorf("Missing thumbnail size: %s", size)
			t.Errorf("Available sizes: %d", len(thumbnails))
			t.Error("BUG: Quality pipeline is rejecting larger thumbnails")
			t.Error("LIKELY CAUSE: DecodeRaw returned small image, triggering upscale prevention")
		}
	}

	// If we only get 1-2 thumbnails, something is very wrong
	if len(thumbnails) < 4 {
		t.Errorf("Only generated %d thumbnails, expected 4", len(thumbnails))
		t.Error("CRITICAL BUG: Not generating full set of thumbnails")
		t.Errorf("Source image: %dx%d", bounds.Dx(), bounds.Dy())

		if bounds.Dx() < 1000 || bounds.Dy() < 1000 {
			t.Error("ROOT CAUSE: DecodeRaw returned small image (embedded preview fallback used tiny preview)")
		}
	}
}

// TestEmbeddedJPEG_MultiplePreviewSizes documents what we learned about Leica DNGs
// LESSON: This test would have forced us to understand the file structure from the start
func TestEmbeddedJPEG_MultiplePreviewSizes(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata)")
	}

	img, err := ExtractEmbeddedJPEG(testFile)
	if err != nil {
		t.Fatalf("ExtractEmbeddedJPEG failed: %v", err)
	}

	bounds := img.Bounds()
	t.Logf("Extracted: %dx%d", bounds.Dx(), bounds.Dy())

	// Document what we learned
	t.Log("=== DNG Structure Documentation ===")
	t.Log("Leica M11 Monochrom JPEG-compressed DNGs contain multiple embedded JPEGs:")
	t.Log("  1. Small 160x120 preview (~5KB)")
	t.Log("  2. Medium preview (~23KB)")
	t.Log("  3. Large 9500x6300 preview (~2MB) <- We want this one!")
	t.Log("")
	t.Log("CRITICAL: ExtractEmbeddedJPEG must return the LARGEST, not the first found")
	t.Log("REASON: LibRaw cannot decode JPEG-compressed monochrome DNGs")
	t.Log("SOLUTION: We must extract the largest embedded JPEG for best quality")
}

// TestLibRawLimitation documents the known issue
// LESSON: Document known limitations so future developers understand the constraints
func TestLibRawLimitation(t *testing.T) {
	t.Log("=== Known Limitation: LibRaw + JPEG-compressed Monochrome DNGs ===")
	t.Log("")
	t.Log("ISSUE: LibRaw cannot properly decode JPEG-compressed monochrome DNG files")
	t.Log("AFFECTS: Leica M11 Monochrom, possibly other cameras")
	t.Log("SYMPTOM: Returns black images or 'not enough image data' error")
	t.Log("")
	t.Log("ROOT CAUSE:")
	t.Log("  - DNG files with JPEG-compressed RAW data")
	t.Log("  - Monochrome sensor (1 color channel vs 3)")
	t.Log("  - LibRaw's JPEG decompression doesn't handle this combination")
	t.Log("")
	t.Log("WORKAROUND:")
	t.Log("  1. Detect black image output from LibRaw")
	t.Log("  2. Fall back to ExtractEmbeddedJPEG()")
	t.Log("  3. Extract the LARGEST embedded JPEG (not the first!)")
	t.Log("  4. Use that for thumbnail generation")
	t.Log("")
	t.Log("QUALITY IMPACT:")
	t.Log("  - Embedded JPEG is ~9500x6300 (almost full 9536x6336 resolution)")
	t.Log("  - Quality is sufficient for thumbnail generation")
	t.Log("  - No visible quality loss compared to RAW decode")
}

// Helper to get max of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
