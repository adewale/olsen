package quality

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/adewale/olsen/pkg/models"
)

// makeTestImage creates a width x height image with a simple gradient so
// resized versions remain comparable.
func makeTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	return img
}

func TestApplyOrientation(t *testing.T) {
	img := makeTestImage(100, 50)

	tests := []struct {
		orientation int
		wantApplied bool
		wantW       int
		wantH       int
	}{
		{1, false, 100, 50}, // normal: no transform
		{3, true, 100, 50},  // 180° rotation keeps dimensions
		{6, true, 50, 100},  // 90° CW swaps dimensions
		{8, true, 50, 100},  // 90° CCW swaps dimensions
		{0, false, 100, 50}, // invalid: no transform
		{9, false, 100, 50}, // out of range: no transform
	}

	for _, tt := range tests {
		got, applied := ApplyOrientation(img, tt.orientation)
		if applied != tt.wantApplied {
			t.Errorf("orientation %d: applied = %v, want %v", tt.orientation, applied, tt.wantApplied)
		}
		if got.Bounds().Dx() != tt.wantW || got.Bounds().Dy() != tt.wantH {
			t.Errorf("orientation %d: got %dx%d, want %dx%d",
				tt.orientation, got.Bounds().Dx(), got.Bounds().Dy(), tt.wantW, tt.wantH)
		}
	}
}

func TestOrientationTrackerRejectsDoubleApply(t *testing.T) {
	tracker := NewOrientationTracker()

	if err := tracker.Apply(6); err != nil {
		t.Fatalf("First Apply failed: %v", err)
	}
	if !tracker.IsApplied() {
		t.Error("IsApplied = false after Apply")
	}
	if err := tracker.Apply(6); err == nil {
		t.Error("Second Apply should return an error")
	}
}

func TestComputeSSIMIdenticalImages(t *testing.T) {
	img := makeTestImage(64, 64)

	ssim, err := ComputeSSIM(img, img)
	if err != nil {
		t.Fatalf("ComputeSSIM failed: %v", err)
	}
	if ssim < 0.99 {
		t.Errorf("SSIM of identical images = %f, want ~1.0", ssim)
	}
}

func TestComputeSSIMDifferentImages(t *testing.T) {
	img1 := makeTestImage(64, 64)
	img2 := image.NewRGBA(image.Rect(0, 0, 64, 64)) // all black

	ssim, err := ComputeSSIM(img1, img2)
	if err != nil {
		t.Fatalf("ComputeSSIM failed: %v", err)
	}
	identical, err := ComputeSSIM(img1, img1)
	if err != nil {
		t.Fatalf("ComputeSSIM failed: %v", err)
	}
	if ssim >= identical {
		t.Errorf("SSIM of different images (%f) should be below identical (%f)", ssim, identical)
	}
}

func TestComputeMSE(t *testing.T) {
	img := makeTestImage(32, 32)
	if mse := ComputeMSE(img, img); mse != 0 {
		t.Errorf("MSE of identical images = %f, want 0", mse)
	}

	black := image.NewRGBA(image.Rect(0, 0, 32, 32))
	if mse := ComputeMSE(img, black); mse == 0 {
		t.Error("MSE of different images should be > 0")
	}
}

func TestGenerateThumbnailsWithDiag(t *testing.T) {
	img := makeTestImage(2048, 1536) // large enough for all 4 sizes
	meta := ImageMetadata{
		FilePath: "/test/large.jpg",
		Width:    2048,
		Height:   1536,
	}

	thumbs, diag, err := GenerateThumbnailsWithDiag(context.Background(), img, meta, DefaultThumbnailConfig())
	if err != nil {
		t.Fatalf("GenerateThumbnailsWithDiag failed: %v", err)
	}
	if diag == nil {
		t.Fatal("Expected diagnostics, got nil")
	}

	for _, size := range []models.ThumbnailSize{models.ThumbnailTiny, models.ThumbnailSmall, models.ThumbnailMedium, models.ThumbnailLarge} {
		data, ok := thumbs[size]
		if !ok {
			t.Errorf("Missing thumbnail size %s", size)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Empty thumbnail for size %s", size)
		}
	}
}

func TestGenerateThumbnailsSkipsUpscaling(t *testing.T) {
	img := makeTestImage(200, 100) // only big enough for the 64px size
	meta := ImageMetadata{FilePath: "/test/small.jpg", Width: 200, Height: 100}

	cfg := DefaultThumbnailConfig()
	if cfg.AllowUpscale {
		t.Fatal("Default config should not allow upscaling")
	}

	thumbs, _, err := GenerateThumbnailsWithDiag(context.Background(), img, meta, cfg)
	if err != nil {
		t.Fatalf("GenerateThumbnailsWithDiag failed: %v", err)
	}

	if _, ok := thumbs[models.ThumbnailLarge]; ok {
		t.Error("1024px thumbnail generated for a 200px image (upscaling)")
	}
	if _, ok := thumbs[models.ThumbnailTiny]; !ok {
		t.Error("64px thumbnail missing for a 200px image")
	}
}

func TestGenerateThumbnailsHonorsCancelledContext(t *testing.T) {
	img := makeTestImage(1024, 768)
	meta := ImageMetadata{FilePath: "/test/cancelled.jpg", Width: 1024, Height: 768}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the pipeline runs

	if _, _, err := GenerateThumbnailsWithDiag(ctx, img, meta, DefaultThumbnailConfig()); err == nil {
		t.Error("Expected an error from a cancelled context, got nil")
	}
}

func TestCountClippedPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 1))
	for x := 0; x < 5; x++ {
		img.Set(x, 0, color.RGBA{0, 0, 0, 255}) // clipped low
	}
	for x := 5; x < 10; x++ {
		img.Set(x, 0, color.RGBA{255, 255, 255, 255}) // clipped high
	}

	low, high := CountClippedPixels(img)
	if low != 5 || high != 5 {
		t.Errorf("CountClippedPixels = (%d, %d), want (5, 5)", low, high)
	}
}
