package indexer

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/adewale/olsen/pkg/models"
)

func TestGenerateThumbnailsFromImage(t *testing.T) {
	// Create a test image (landscape)
	img := createSolidColorImage(800, 600, color.RGBA{255, 0, 0, 255})

	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	// Check that all 4 sizes were generated
	expectedSizes := []models.ThumbnailSize{
		models.ThumbnailTiny,
		models.ThumbnailSmall,
		models.ThumbnailMedium,
		models.ThumbnailLarge,
	}

	for _, size := range expectedSizes {
		data, ok := thumbnails[size]
		if !ok {
			t.Errorf("Missing thumbnail size: %s", size)
			continue
		}

		if len(data) == 0 {
			t.Errorf("Thumbnail %s has zero length", size)
			continue
		}

		// Verify it's a valid JPEG
		_, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			t.Errorf("Thumbnail %s is not a valid JPEG: %v", size, err)
		}
	}
}

func TestThumbnailAspectRatioLandscape(t *testing.T) {
	// Create a landscape image (800x600 = 4:3)
	img := createSolidColorImage(800, 600, color.RGBA{100, 150, 200, 255})

	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	tests := []struct {
		size          models.ThumbnailSize
		expectedWidth int
		maxHeight     int
	}{
		{models.ThumbnailTiny, 64, 48},
		{models.ThumbnailSmall, 256, 192},
		{models.ThumbnailMedium, 512, 384},
		{models.ThumbnailLarge, 1024, 768},
	}

	for _, tt := range tests {
		data := thumbnails[tt.size]
		thumbImg, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode thumbnail %s: %v", tt.size, err)
		}

		bounds := thumbImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// For landscape, width should be constrained to max dimension
		if width != tt.expectedWidth {
			t.Errorf("Thumbnail %s width = %d; want %d", tt.size, width, tt.expectedWidth)
		}

		// Height should be proportionally scaled (within 1px due to rounding)
		if height > tt.maxHeight+1 {
			t.Errorf("Thumbnail %s height = %d; want <= %d", tt.size, height, tt.maxHeight)
		}

		// Check aspect ratio is preserved (approximately)
		originalAspect := 800.0 / 600.0
		thumbAspect := float64(width) / float64(height)
		aspectDiff := originalAspect - thumbAspect
		if aspectDiff < 0 {
			aspectDiff = -aspectDiff
		}
		if aspectDiff > 0.02 {
			t.Errorf("Thumbnail %s aspect ratio not preserved: got %.2f, want %.2f",
				tt.size, thumbAspect, originalAspect)
		}
	}
}

func TestThumbnailAspectRatioPortrait(t *testing.T) {
	// Create a portrait image (600x800 = 3:4)
	img := createSolidColorImage(600, 800, color.RGBA{100, 150, 200, 255})

	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	tests := []struct {
		size           models.ThumbnailSize
		maxWidth       int
		expectedHeight int
	}{
		{models.ThumbnailTiny, 48, 64},
		{models.ThumbnailSmall, 192, 256},
		{models.ThumbnailMedium, 384, 512},
		{models.ThumbnailLarge, 768, 1024},
	}

	for _, tt := range tests {
		data := thumbnails[tt.size]
		thumbImg, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode thumbnail %s: %v", tt.size, err)
		}

		bounds := thumbImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// For portrait, height should be constrained to max dimension
		if height != tt.expectedHeight {
			t.Errorf("Thumbnail %s height = %d; want %d", tt.size, height, tt.expectedHeight)
		}

		// Width should be proportionally scaled (within 1px due to rounding)
		if width > tt.maxWidth+1 {
			t.Errorf("Thumbnail %s width = %d; want <= %d", tt.size, width, tt.maxWidth)
		}

		// Check aspect ratio is preserved (approximately)
		originalAspect := 600.0 / 800.0
		thumbAspect := float64(width) / float64(height)
		aspectDiff := originalAspect - thumbAspect
		if aspectDiff < 0 {
			aspectDiff = -aspectDiff
		}
		if aspectDiff > 0.02 {
			t.Errorf("Thumbnail %s aspect ratio not preserved: got %.2f, want %.2f",
				tt.size, thumbAspect, originalAspect)
		}
	}
}

func TestThumbnailAspectRatioSquare(t *testing.T) {
	// Create a square image (800x800)
	img := createSolidColorImage(800, 800, color.RGBA{100, 150, 200, 255})

	thumbnails, err := GenerateThumbnailsFromImage(img)
	if err != nil {
		t.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
	}

	tests := []struct {
		size        models.ThumbnailSize
		expectedDim int
	}{
		{models.ThumbnailTiny, 64},
		{models.ThumbnailSmall, 256},
		{models.ThumbnailMedium, 512},
		{models.ThumbnailLarge, 1024},
	}

	for _, tt := range tests {
		data := thumbnails[tt.size]
		thumbImg, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to decode thumbnail %s: %v", tt.size, err)
		}

		bounds := thumbImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// For square, both dimensions should be equal (within 1px)
		if absInt(width-height) > 1 {
			t.Errorf("Thumbnail %s not square: %dx%d", tt.size, width, height)
		}

		// Both should be constrained to max dimension
		if absInt(width-tt.expectedDim) > 1 {
			t.Errorf("Thumbnail %s dimension = %d; want %d", tt.size, width, tt.expectedDim)
		}
	}
}

// Helper function to create a solid color image
func createSolidColorImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}
