package quality

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	"github.com/nfnt/resize"
)

// ApproachConfig defines a thumbnail generation approach to test
type ApproachConfig struct {
	Name          string                       // Human-readable name
	ResizeMethod  resize.InterpolationFunction // Resize algorithm
	JPEGQuality   int                          // JPEG quality (1-100)
	PreSharpen    bool                         // Apply sharpening before resize
	PostSharpen   bool                         // Apply sharpening after resize
	SharpenAmount float64                      // Sharpening strength (0.0-1.0)
}

// ComparisonResult holds the results of comparing a thumbnail approach
type ComparisonResult struct {
	Config         ApproachConfig
	Metrics        Metrics
	ThumbnailData  []byte        // The actual thumbnail bytes
	ProcessingTime time.Duration // Time to generate thumbnail
	ThumbnailSize  int           // Size in bytes
	WidthPx        int           // Actual width in pixels
	HeightPx       int           // Actual height in pixels
}

// TestApproach generates a thumbnail using a specific approach and measures quality
// reference: the full-resolution source image
// config: the approach configuration to test
// targetSize: the longest edge size for the thumbnail
func TestApproach(reference image.Image, config ApproachConfig, targetSize uint) (*ComparisonResult, error) {
	startTime := time.Now()

	img := reference

	// Pre-sharpening if enabled
	if config.PreSharpen {
		img = applyUnsharpMask(img, config.SharpenAmount)
	}

	// Calculate dimensions preserving aspect ratio
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newWidth, newHeight uint
	if width > height {
		newWidth = targetSize
		newHeight = 0 // resize library calculates to preserve aspect ratio
	} else {
		newWidth = 0
		newHeight = targetSize
	}

	// Resize
	thumbnail := resize.Resize(newWidth, newHeight, img, config.ResizeMethod)

	// Post-sharpening if enabled
	if config.PostSharpen {
		thumbnail = applyUnsharpMask(thumbnail, config.SharpenAmount)
	}

	// Encode as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: config.JPEGQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	thumbnailData := buf.Bytes()
	processingTime := time.Since(startTime)

	// Decode the thumbnail back for quality comparison
	decodedThumbnail, _, err := image.Decode(bytes.NewReader(thumbnailData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode thumbnail for comparison: %w", err)
	}

	// Resize reference image to same size as thumbnail for fair comparison
	thumbBounds := decodedThumbnail.Bounds()
	resizedReference := resize.Resize(
		uint(thumbBounds.Dx()),
		uint(thumbBounds.Dy()),
		reference,
		resize.Lanczos3, // Use high-quality resize for reference
	)

	// Compute quality metrics
	metrics, err := ComputeAllMetrics(resizedReference, decodedThumbnail)
	if err != nil {
		return nil, fmt.Errorf("failed to compute metrics: %w", err)
	}

	result := &ComparisonResult{
		Config:         config,
		Metrics:        metrics,
		ThumbnailData:  thumbnailData,
		ProcessingTime: processingTime,
		ThumbnailSize:  len(thumbnailData),
		WidthPx:        thumbBounds.Dx(),
		HeightPx:       thumbBounds.Dy(),
	}

	return result, nil
}

// CompareApproaches tests multiple approaches on a single image
func CompareApproaches(reference image.Image, configs []ApproachConfig, targetSize uint) ([]*ComparisonResult, error) {
	results := make([]*ComparisonResult, 0, len(configs))

	for _, config := range configs {
		result, err := TestApproach(reference, config, targetSize)
		if err != nil {
			return nil, fmt.Errorf("failed to test approach %s: %w", config.Name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetStandardApproaches returns a set of standard approaches to compare
func GetStandardApproaches() []ApproachConfig {
	return []ApproachConfig{
		{
			Name:         "Current (Lanczos3, JPEG 85%)",
			ResizeMethod: resize.Lanczos3,
			JPEGQuality:  85,
		},
		{
			Name:          "Lanczos3 + Post-sharpen",
			ResizeMethod:  resize.Lanczos3,
			JPEGQuality:   85,
			PostSharpen:   true,
			SharpenAmount: 0.4,
		},
		{
			Name:         "Lanczos2, JPEG 85%",
			ResizeMethod: resize.Lanczos2,
			JPEGQuality:  85,
		},
		{
			Name:         "Mitchell-Netravali, JPEG 85%",
			ResizeMethod: resize.MitchellNetravali,
			JPEGQuality:  85,
		},
		{
			Name:         "Bilinear, JPEG 85%",
			ResizeMethod: resize.Bilinear,
			JPEGQuality:  85,
		},
		{
			Name:         "Lanczos3, JPEG 90%",
			ResizeMethod: resize.Lanczos3,
			JPEGQuality:  90,
		},
		{
			Name:         "Lanczos3, JPEG 80%",
			ResizeMethod: resize.Lanczos3,
			JPEGQuality:  80,
		},
		{
			Name:         "Lanczos3, JPEG 95% (high quality)",
			ResizeMethod: resize.Lanczos3,
			JPEGQuality:  95,
		},
	}
}

// applyUnsharpMask applies basic unsharp masking for sharpening
// This is a simplified version - a production implementation might use Gaussian blur
func applyUnsharpMask(img image.Image, amount float64) image.Image {
	// For now, return the original image
	// TODO: Implement proper unsharp masking when needed
	// This would require:
	// 1. Gaussian blur of the original
	// 2. Subtract blurred from original to get mask
	// 3. Add mask back to original with amount as weight
	return img
}

// BenchmarkSummary provides aggregate statistics across multiple images
type BenchmarkSummary struct {
	ApproachName      string
	ImageCount        int
	AvgSSIM           float64
	AvgPSNR           float64
	AvgSharpness      float64
	AvgProcessingTime time.Duration
	AvgFileSize       int
	TotalSize         int64
}

// SummarizeResults creates aggregate statistics for an approach across multiple images
func SummarizeResults(results []*ComparisonResult) BenchmarkSummary {
	if len(results) == 0 {
		return BenchmarkSummary{}
	}

	summary := BenchmarkSummary{
		ApproachName: results[0].Config.Name,
		ImageCount:   len(results),
	}

	var totalSSIM, totalPSNR, totalSharpness float64
	var totalProcessingTime time.Duration
	var totalSize int64

	for _, result := range results {
		totalSSIM += result.Metrics.SSIM
		totalPSNR += result.Metrics.PSNR
		totalSharpness += result.Metrics.Sharpness
		totalProcessingTime += result.ProcessingTime
		totalSize += int64(result.ThumbnailSize)
	}

	n := float64(len(results))
	summary.AvgSSIM = totalSSIM / n
	summary.AvgPSNR = totalPSNR / n
	summary.AvgSharpness = totalSharpness / n
	summary.AvgProcessingTime = totalProcessingTime / time.Duration(len(results))
	summary.AvgFileSize = int(totalSize / int64(len(results)))
	summary.TotalSize = totalSize

	return summary
}
