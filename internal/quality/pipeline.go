// Package quality provides thumbnail quality assurance and diagnostic tools.
//
// It implements instrumented thumbnail generation with detailed tracking of orientation
// correction, color space handling, resizing operations, and quality metrics. The package
// supports sampling-based quality validation and artifact generation for debugging.
package quality

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	"github.com/nfnt/resize"

	"github.com/adewale/olsen/pkg/models"
)

// ThumbnailConfig contains configuration for thumbnail generation
type ThumbnailConfig struct {
	// Quality settings per size
	QualityTiers map[models.ThumbnailSize]int // e.g., {ThumbnailSmall: 80, ThumbnailMedium: 85}

	// Resize filter
	Filter resize.InterpolationFunction

	// Sharpening
	PostSharpen   bool
	SharpenAmount float64
	SharpenRadius float64

	// Policies
	AllowUpscale bool
	LinearResize bool // Gamma-correct resizing

	// QA/Sampling
	QASample           float64 // 0.01 = 1%
	QADir              string  // Where to store artifacts
	QADisableArtifacts bool
}

// DefaultThumbnailConfig returns the default configuration
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		QualityTiers: map[models.ThumbnailSize]int{
			models.ThumbnailTiny:   80,
			models.ThumbnailSmall:  85,
			models.ThumbnailMedium: 90,
			models.ThumbnailLarge:  92,
		},
		Filter:       resize.Lanczos3,
		PostSharpen:  false,
		AllowUpscale: false,
		LinearResize: false,
		QASample:     0.0, // Disabled by default
	}
}

// ImageMetadata contains metadata needed for thumbnail generation
type ImageMetadata struct {
	FilePath       string
	Orientation    int    // EXIF orientation (1-8)
	ColorSpace     string // "sRGB", "AdobeRGB", etc.
	HasICCProfile  bool
	ICCDescription string
	Width          int
	Height         int
}

// ThumbnailResult contains the generated thumbnail and diagnostics
type ThumbnailResult struct {
	Size        models.ThumbnailSize
	Data        []byte
	Diagnostics *ImageDiag
}

// GenerateThumbnailsWithDiag generates thumbnails with full instrumentation
func GenerateThumbnailsWithDiag(ctx context.Context, img image.Image, meta ImageMetadata, cfg ThumbnailConfig) (map[models.ThumbnailSize][]byte, *ImageDiag, error) {
	t0 := time.Now()

	// Initialize diagnostics
	imgID := computeImageID(meta.FilePath)
	diag := NewImageDiag(imgID)
	diag.Source.Format = "processed" // Caller should override if known
	diag.Source.InputW = img.Bounds().Dx()
	diag.Source.InputH = img.Bounds().Dy()
	diag.Source.EXIFOrientation = meta.Orientation
	diag.Source.HasICC = meta.HasICCProfile
	diag.Source.ICCDesc = meta.ICCDescription

	defer func() {
		diag.TimingMS.Total = msSince(t0)
	}()

	// Stage 1: Orientation
	orientStart := time.Now()
	orientTracker := NewOrientationTracker()

	if meta.Orientation > 1 && meta.Orientation <= 8 {
		var applied bool
		img, applied = ApplyOrientation(img, meta.Orientation)
		if applied {
			if err := orientTracker.Apply(meta.Orientation); err != nil {
				diag.AddWarning(fmt.Sprintf("orientation_double_apply: %v", err))
			}
			diag.Pipeline.OrientationApplied = true
		}
	}
	diag.TimingMS.Orient = msSince(orientStart)

	// Stage 2: Color space (stub for now - ICC handling would go here)
	colorStart := time.Now()
	diag.Pipeline.ColorspaceIn = meta.ColorSpace
	diag.Pipeline.ColorspaceOut = "sRGB" // Assume sRGB output for now
	if !meta.HasICCProfile && meta.ColorSpace == "" {
		diag.AddWarning("icc_missing_assumed_srgb")
	}
	diag.TimingMS.Color = msSince(colorStart)

	// Stage 3: Generate thumbnails for each size
	thumbnails := make(map[models.ThumbnailSize][]byte)
	sizes := []struct {
		name         models.ThumbnailSize
		maxDimension uint
	}{
		{models.ThumbnailTiny, 64},
		{models.ThumbnailSmall, 256},
		{models.ThumbnailMedium, 512},
		{models.ThumbnailLarge, 1024},
	}

	// We'll track metrics for the medium size as representative
	var mediumThumb image.Image

	for _, size := range sizes {
		resizeStart := time.Now()

		// Check for upscaling
		bounds := img.Bounds()
		width := uint(bounds.Dx())
		height := uint(bounds.Dy())
		longEdge := max(width, height)

		if longEdge < size.maxDimension {
			if !cfg.AllowUpscale {
				diag.AddWarning(fmt.Sprintf("upscale_detected: %dx%d -> %d (skipped)", width, height, size.maxDimension))
				// Skip this size or use original
				continue
			}
			diag.AddWarning(fmt.Sprintf("upscale_detected: %dx%d -> %d", width, height, size.maxDimension))
			if size.name == models.ThumbnailMedium {
				diag.Pipeline.Resize.Upscale = true
			}
		}

		// Calculate dimensions preserving aspect ratio
		var newWidth, newHeight uint
		if width > height {
			newWidth = size.maxDimension
			newHeight = 0 // resize library calculates
		} else {
			newWidth = 0
			newHeight = size.maxDimension
		}

		// Resize
		thumb := resize.Resize(newWidth, newHeight, img, cfg.Filter)

		resizeTime := msSince(resizeStart)
		if size.name == models.ThumbnailMedium {
			diag.TimingMS.Resize = resizeTime
			diag.Pipeline.Resize.TargetLongEdge = int(size.maxDimension)
			diag.Pipeline.Resize.Filter = filterName(cfg.Filter)
			mediumThumb = thumb
		}

		// Stage 4: Post-sharpening (if enabled)
		sharpenStart := time.Now()
		if cfg.PostSharpen {
			// TODO: Implement actual unsharp mask
			// For now, just record the configuration
			if size.name == models.ThumbnailMedium {
				diag.Pipeline.Resize.PostSharpen.Enabled = true
				diag.Pipeline.Resize.PostSharpen.Amount = cfg.SharpenAmount
				diag.Pipeline.Resize.PostSharpen.Radius = cfg.SharpenRadius
			}
		}
		sharpenTime := msSince(sharpenStart)
		if size.name == models.ThumbnailMedium {
			diag.TimingMS.Sharpen = sharpenTime
		}

		// Stage 5: Encode
		encodeStart := time.Now()
		quality := cfg.QualityTiers[size.name]
		if quality == 0 {
			quality = 85 // Default
		}

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: quality}); err != nil {
			return nil, diag, fmt.Errorf("failed to encode thumbnail %s: %w", size.name, err)
		}

		thumbnailData := buf.Bytes()
		thumbnails[size.name] = thumbnailData

		encodeTime := msSince(encodeStart)
		if size.name == models.ThumbnailMedium {
			diag.TimingMS.Encode = encodeTime
			diag.Pipeline.Encode.Format = "jpeg"
			diag.Pipeline.Encode.Quality = quality
			diag.Pipeline.Encode.Chroma = "420" // JPEG default
			diag.Pipeline.Encode.Progressive = false
			diag.Pipeline.Encode.Bytes = len(thumbnailData)
		}
	}

	// Stage 6: Compute metrics (if sampling enabled)
	if cfg.QASample > 0 && mediumThumb != nil {
		// TODO: Implement reference generation and metric computation
		// For now, just record that sampling would happen
		if shouldSample(cfg.QASample) {
			// Generate reference thumbnail
			refStart := time.Now()
			reference := generateReferenceThumbnail(img, 512)

			// Compute metrics
			metrics, err := ComputeAllMetrics(reference, mediumThumb)
			if err == nil {
				diag.Metrics.SSIMVsRef = metrics.SSIM
				diag.Metrics.PSNRVsRefDB = metrics.PSNR
				diag.Metrics.LapVar = metrics.Sharpness
			}

			// Compute Delta-E
			diag.Metrics.DeltaEMean = ComputeDeltaE(reference, mediumThumb)

			// Count clipped pixels
			low, high := CountClippedPixels(mediumThumb)
			diag.Metrics.ClippedPixelsLow = low
			diag.Metrics.ClippedPixelsHigh = high

			_ = msSince(refStart) // Metrics computation time (not logged separately)
		}
	}

	return thumbnails, diag, nil
}

// Helper functions

func computeImageID(filePath string) string {
	h := sha256.Sum256([]byte(filePath))
	return fmt.Sprintf("sha256:%x", h[:8]) // First 8 bytes
}

func msSince(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000.0
}

func filterName(f resize.InterpolationFunction) string {
	// This is a bit of a hack since InterpolationFunction doesn't expose its name
	// We could use a map or switch if needed
	return "Lanczos3" // Default assumption
}

func shouldSample(rate float64) bool {
	if rate <= 0 || rate > 1 {
		return false
	}
	// Simple deterministic sampling based on time
	// In production, use a better random source
	return time.Now().UnixNano()%100 < int64(rate*100)
}

func generateReferenceThumbnail(img image.Image, targetSize uint) image.Image {
	// Generate a high-quality reference using best-known settings
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newWidth, newHeight uint
	if width > height {
		newWidth = targetSize
		newHeight = 0
	} else {
		newWidth = 0
		newHeight = targetSize
	}

	// Use Lanczos3 for reference
	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}
