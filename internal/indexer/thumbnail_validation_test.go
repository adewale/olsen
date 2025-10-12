//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/explorer"
	"github.com/adewale/olsen/internal/indexer"
)

// TestThumbnailVisualFidelity verifies that thumbnails visually resemble the original images
func TestThumbnailVisualFidelity(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	// Load original RAW file for comparison
	t.Logf("Loading original RAW file: %s", testFile)
	originalImg, err := indexer.DecodeRaw(testFile)
	if err != nil {
		t.Fatalf("Failed to load original RAW: %v", err)
	}

	// Analyze original image
	originalBrightness := calculateAverageBrightness(originalImg)
	originalHistogram := buildHistogram(originalImg)
	t.Logf("Original RAW brightness: %.1f/255", originalBrightness)
	t.Logf("Original RAW histogram: %s", formatHistogram(originalHistogram))

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_visual.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the file
	engine := indexer.NewEngine(db, 1)
	testDir := filepath.Dir(testFile)
	err = engine.IndexDirectory(testDir)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	stats := engine.GetStats()
	if stats.FilesFailed > 0 {
		t.Fatalf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Get recent photos
	photos, err := repo.GetRecentPhotos(1)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}
	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	photoID := photos[0].ID

	// Test each thumbnail size
	sizes := []string{"64", "256", "512", "1024"}
	expectedMaxDimensions := map[string]int{
		"64":   64,
		"256":  256,
		"512":  512,
		"1024": 1024,
	}

	for _, size := range sizes {
		thumbData, err := repo.GetThumbnail(photoID, size)
		if err != nil {
			t.Errorf("Failed to get thumbnail %s: %v", size, err)
			continue
		}

		// Decode thumbnail
		thumbImg, err := jpeg.Decode(bytes.NewReader(thumbData))
		if err != nil {
			t.Errorf("Failed to decode thumbnail %s: %v", size, err)
			continue
		}

		bounds := thumbImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// 1. Check dimensions: longest edge should be constrained to max dimension
		maxDim := expectedMaxDimensions[size]
		longestEdge := max(width, height)
		if longestEdge > maxDim {
			t.Errorf("Thumbnail %s longest edge %d exceeds max %d", size, longestEdge, maxDim)
		}

		// 2. Check aspect ratio preservation (original is 9536×6336 = 1.505:1)
		expectedRatio := 9536.0 / 6336.0
		actualRatio := float64(width) / float64(height)
		ratioDiff := math.Abs(expectedRatio - actualRatio)
		if ratioDiff > 0.01 { // Allow 1% tolerance
			t.Errorf("Thumbnail %s aspect ratio %.3f differs from original %.3f (diff: %.3f)",
				size, actualRatio, expectedRatio, ratioDiff)
		}

		// 3. Analyze brightness and histogram
		thumbBrightness := calculateAverageBrightness(thumbImg)
		thumbHistogram := buildHistogram(thumbImg)

		// Compare brightness with original (allow some variation due to resampling)
		brightnessDiff := originalBrightness - thumbBrightness
		brightnessDiffPct := (brightnessDiff / originalBrightness) * 100

		t.Logf("Thumbnail %s:", size)
		t.Logf("  Dimensions: %dx%d (ratio %.3f)", width, height, actualRatio)
		t.Logf("  Brightness: %.1f/255 (original: %.1f, diff: %.1f%%)",
			thumbBrightness, originalBrightness, brightnessDiffPct)
		t.Logf("  Histogram: %s", formatHistogram(thumbHistogram))

		// 4. Check brightness similarity (thumbnails shouldn't be drastically different)
		if math.Abs(brightnessDiffPct) > 30 {
			t.Errorf("Thumbnail %s brightness differs by %.1f%% from original (threshold: 30%%)",
				size, brightnessDiffPct)
		}

		// 5. Check that image is grayscale for monochrome DNGs
		if !isGrayscaleImage(thumbImg) {
			t.Errorf("Thumbnail %s is not grayscale (expected for monochrome DNG)", size)
		}

		t.Logf("✓ Thumbnail %s: brightness within %.1f%% of original, is grayscale",
			size, brightnessDiffPct)
	}
}

// TestThumbnailContentSimilarity tests that thumbnails preserve visual content
func TestThumbnailContentSimilarity(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_similarity.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the file
	engine := indexer.NewEngine(db, 1)
	testDir := filepath.Dir(testFile)
	err = engine.IndexDirectory(testDir)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	stats := engine.GetStats()
	if stats.FilesFailed > 0 {
		t.Fatalf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Get recent photos
	photos, err := repo.GetRecentPhotos(1)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}
	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	photoID := photos[0].ID

	// Get small and medium thumbnails for comparison
	smallData, err := repo.GetThumbnail(photoID, "256")
	if err != nil {
		t.Fatalf("Failed to get small thumbnail: %v", err)
	}

	mediumData, err := repo.GetThumbnail(photoID, "512")
	if err != nil {
		t.Fatalf("Failed to get medium thumbnail: %v", err)
	}

	smallImg, _ := jpeg.Decode(bytes.NewReader(smallData))
	mediumImg, _ := jpeg.Decode(bytes.NewReader(mediumData))

	// Calculate average brightness for both thumbnails
	// They should be similar (both from same original image)
	smallBrightness := calculateAverageBrightness(smallImg)
	mediumBrightness := calculateAverageBrightness(mediumImg)

	brightnessDiff := math.Abs(smallBrightness - mediumBrightness)
	if brightnessDiff > 10.0 { // Allow 10/255 tolerance
		t.Errorf("Brightness differs too much between small (%.1f) and medium (%.1f) thumbnails (diff: %.1f)",
			smallBrightness, mediumBrightness, brightnessDiff)
	}

	t.Logf("✓ Thumbnail brightness similarity: small=%.1f, medium=%.1f, diff=%.1f",
		smallBrightness, mediumBrightness, brightnessDiff)

	// Calculate histogram correlation (basic structural similarity)
	correlation := calculateHistogramCorrelation(smallImg, mediumImg)
	if correlation < 0.90 { // Expect >90% correlation
		t.Errorf("Histogram correlation too low: %.3f (expected >0.90)", correlation)
	}

	t.Logf("✓ Histogram correlation: %.3f", correlation)
}

// TestThumbnailBatchConsistency tests that all thumbnails in a batch are valid
func TestThumbnailBatchConsistency(t *testing.T) {
	testDir := "../../private-testdata/2024-12-18"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test directory not found (requires private-testdata)")
	}

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_consistency.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index directory
	engine := indexer.NewEngine(db, 4)
	err = engine.IndexDirectory(testDir)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	stats := engine.GetStats()
	if stats.FilesFailed > 0 {
		t.Fatalf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Get all photos (using a large limit)
	photos, err := repo.GetRecentPhotos(1000)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}

	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	// Check each photo has valid thumbnails with correct properties
	for _, photo := range photos {
		// Get small thumbnail (most commonly used)
		thumbData, err := repo.GetThumbnail(photo.ID, "256")
		if err != nil {
			t.Errorf("Photo %d: failed to get thumbnail: %v", photo.ID, err)
			continue
		}

		if len(thumbData) == 0 {
			t.Errorf("Photo %d: thumbnail data is empty", photo.ID)
			continue
		}

		// Decode thumbnail
		thumbImg, err := jpeg.Decode(bytes.NewReader(thumbData))
		if err != nil {
			t.Errorf("Photo %d: failed to decode thumbnail: %v", photo.ID, err)
			continue
		}

		bounds := thumbImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// Check dimensions are reasonable
		if width < 10 || height < 10 {
			t.Errorf("Photo %d: thumbnail too small: %dx%d", photo.ID, width, height)
		}

		if width > 256 && height > 256 {
			t.Errorf("Photo %d: thumbnail too large for ThumbnailSmall: %dx%d", photo.ID, width, height)
		}

		// Check image has content
		if isBlankImage(thumbImg) {
			t.Errorf("Photo %d: thumbnail appears blank", photo.ID)
		}
	}

	t.Logf("✓ All %d photos have valid thumbnails", len(photos))
}

// Helper: Check if image is blank (all pixels similar)
func isBlankImage(img image.Image) bool {
	bounds := img.Bounds()
	if bounds.Dx() < 2 || bounds.Dy() < 2 {
		return true
	}

	// Sample 10 pixels and check for variation
	samples := make([]uint32, 0, 10)
	step := bounds.Dx() / 10
	if step < 1 {
		step = 1
	}

	for x := bounds.Min.X; x < bounds.Max.X && len(samples) < 10; x += step {
		y := bounds.Min.Y + bounds.Dy()/2 // Sample middle row
		r, g, b, _ := img.At(x, y).RGBA()
		// Convert to 8-bit grayscale
		gray := (r + g + b) / 3 / 256
		samples = append(samples, gray)
	}

	// Check for variation (max - min should be > 5)
	minVal := samples[0]
	maxVal := samples[0]
	for _, val := range samples {
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	variation := maxVal - minVal
	return variation < 5 // Less than 5/255 variation = blank
}

// Helper: Check if image is grayscale (R=G=B for most pixels)
func isGrayscaleImage(img image.Image) bool {
	bounds := img.Bounds()
	sampleCount := 0
	grayscaleCount := 0

	// Sample 100 pixels
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
			r8 := r / 256
			g8 := g / 256
			b8 := b / 256

			// Check if R ≈ G ≈ B (within 2/255 tolerance)
			if math.Abs(float64(r8)-float64(g8)) < 2 &&
				math.Abs(float64(g8)-float64(b8)) < 2 {
				grayscaleCount++
			}
			sampleCount++
		}
	}

	// Image is grayscale if >95% of samples are grayscale
	return float64(grayscaleCount)/float64(sampleCount) > 0.95
}

// Helper: Calculate average brightness (0-255)
func calculateAverageBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	var sum uint64
	var count uint64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit grayscale
			gray := (r + g + b) / 3 / 256
			sum += uint64(gray)
			count++
		}
	}

	return float64(sum) / float64(count)
}

// Helper: Calculate histogram correlation between two images
func calculateHistogramCorrelation(img1, img2 image.Image) float64 {
	// Build histograms (256 bins for grayscale)
	hist1 := make([]int, 256)
	hist2 := make([]int, 256)

	bounds1 := img1.Bounds()
	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r, g, b, _ := img1.At(x, y).RGBA()
			gray := (r + g + b) / 3 / 256
			if gray < 256 {
				hist1[gray]++
			}
		}
	}

	bounds2 := img2.Bounds()
	for y := bounds2.Min.Y; y < bounds2.Max.Y; y++ {
		for x := bounds2.Min.X; x < bounds2.Max.X; x++ {
			r, g, b, _ := img2.At(x, y).RGBA()
			gray := (r + g + b) / 3 / 256
			if gray < 256 {
				hist2[gray]++
			}
		}
	}

	// Normalize histograms
	total1 := float64(bounds1.Dx() * bounds1.Dy())
	total2 := float64(bounds2.Dx() * bounds2.Dy())

	// Calculate correlation coefficient
	var sum1, sum2, sum1Sq, sum2Sq, productSum float64
	for i := 0; i < 256; i++ {
		val1 := float64(hist1[i]) / total1
		val2 := float64(hist2[i]) / total2

		sum1 += val1
		sum2 += val2
		sum1Sq += val1 * val1
		sum2Sq += val2 * val2
		productSum += val1 * val2
	}

	n := 256.0
	numerator := n*productSum - sum1*sum2
	denominator := math.Sqrt((n*sum1Sq - sum1*sum1) * (n*sum2Sq - sum2*sum2))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// buildHistogram creates a 256-bin histogram for grayscale images
func buildHistogram(img image.Image) [256]int {
	var histogram [256]int
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (r + g + b) / 3 / 256
			if gray < 256 {
				histogram[gray]++
			}
		}
	}

	return histogram
}

// formatHistogram creates a visual representation of the histogram
// Shows distribution across 10 bins (0-25, 26-50, 51-75, etc.)
func formatHistogram(histogram [256]int) string {
	// Aggregate into 10 bins
	bins := make([]int, 10)
	for i := 0; i < 256; i++ {
		bin := i / 26 // 0-25->0, 26-51->1, etc.
		if bin > 9 {
			bin = 9
		}
		bins[bin] += histogram[i]
	}

	// Find max for scaling
	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	// Create visual bars
	result := "["
	for i, count := range bins {
		if i > 0 {
			result += " "
		}
		pct := float64(count) / float64(maxCount) * 100
		if pct < 10 {
			result += "_"
		} else if pct < 30 {
			result += "▁"
		} else if pct < 50 {
			result += "▃"
		} else if pct < 70 {
			result += "▅"
		} else if pct < 90 {
			result += "▇"
		} else {
			result += "█"
		}
	}
	result += "]"

	// Add range labels
	result += " (0=black → 9=white)"

	return result
}

// getHistogramStats returns key statistics from a histogram
func getHistogramStats(histogram [256]int) (min, max, median, mean int) {
	totalPixels := 0
	sum := 0
	min = 255
	max = 0

	for value, count := range histogram {
		if count > 0 {
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
			totalPixels += count
			sum += value * count
		}
	}

	if totalPixels > 0 {
		mean = sum / totalPixels

		// Find median
		halfPixels := totalPixels / 2
		cumulative := 0
		for value, count := range histogram {
			cumulative += count
			if cumulative >= halfPixels {
				median = value
				break
			}
		}
	}

	return
}
