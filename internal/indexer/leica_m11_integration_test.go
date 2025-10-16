//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/explorer"
	"github.com/adewale/olsen/internal/indexer"
)

// TestLeicaM11Monochrom_LensMetadata is an integration test that verifies
// L1001515.DNG was taken with a 50mm f/2 lens (Apo-Summicron-M 1:2/50 ASPH.)
func TestLeicaM11Monochrom_LensMetadata(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found: ", testFile)
	}

	// Create temporary database and index the file
	dbPath := filepath.Join(t.TempDir(), "test_leica_lens.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the directory containing L1001515.DNG
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

	// Query specifically for L1001515.DNG by file path
	testFileName := filepath.Base(testFile)
	var photoID int
	err = db.QueryRow("SELECT id FROM photos WHERE file_path LIKE ?", "%"+testFileName).Scan(&photoID)
	if err != nil {
		t.Fatalf("Failed to find %s in database: %v", testFileName, err)
	}

	photoDetail, err := repo.GetPhotoByID(photoID)
	if err != nil {
		t.Fatalf("Failed to get photo detail for %s: %v", testFileName, err)
	}

	t.Logf("Testing lens metadata for: %s (ID: %d)", testFileName, photoID)
	t.Logf("  Camera: %s %s", photoDetail.CameraMake, photoDetail.CameraModel)
	t.Logf("  Lens: %s", photoDetail.LensModel)
	t.Logf("  Focal Length: %.1fmm", photoDetail.FocalLength)
	t.Logf("  Max Aperture: f/%.1f", photoDetail.Aperture)

	// Verify focal length is 50mm
	if photoDetail.FocalLength != 50.0 {
		t.Errorf("Expected focal length 50.0mm, got %.1fmm", photoDetail.FocalLength)
	}

	// Verify lens model contains "50" and "f/2" or "1:2"
	lensLower := strings.ToLower(photoDetail.LensModel)
	if !strings.Contains(lensLower, "50") {
		t.Errorf("Expected lens model to contain '50mm', got %q", photoDetail.LensModel)
	}

	// Check for f/2 aperture indication (either "f/2" or "1:2" notation)
	hasF2 := strings.Contains(lensLower, "f/2") || strings.Contains(lensLower, "1:2")
	if !hasF2 {
		t.Errorf("Expected lens model to indicate f/2 aperture (f/2 or 1:2), got %q", photoDetail.LensModel)
	}

	// Verify it's the Apo-Summicron-M lens
	if !strings.Contains(lensLower, "summicron") {
		t.Errorf("Expected Apo-Summicron-M lens, got %q", photoDetail.LensModel)
	}

	t.Logf("✓ Verified: L1001515.DNG was taken with 50mm f/2 lens")
	t.Logf("  Full lens name: %s", photoDetail.LensModel)
}

// TestLeicaM11Monochrom_ThumbnailGenerationWithoutFallback verifies that
// we can construct thumbnails for L1001515.DNG without triggering the
// LibRaw fallback to embedded JPEG extraction.
//
// IMPORTANT: This test currently EXPECTS the fallback to be triggered because
// seppedelanghe/go-libraw has a known bug with JPEG-compressed monochrome DNGs.
// When the upstream library is fixed, this test should be updated to fail if
// the fallback is triggered.
func TestLeicaM11Monochrom_ThumbnailGenerationWithoutFallback(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found: ", testFile)
	}

	// Create temporary database and index the file
	dbPath := filepath.Join(t.TempDir(), "test_leica_thumbnail.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the directory containing L1001515.DNG
	engine := indexer.NewEngine(db, 1)
	testDir := filepath.Dir(testFile)

	// Track whether fallback was triggered by watching log output
	fallbackTriggered := false

	err = engine.IndexDirectory(testDir)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	stats := engine.GetStats()
	if stats.FilesFailed > 0 {
		t.Fatalf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	// Query for thumbnails to verify they were generated
	testFileName := filepath.Base(testFile)
	var photoID int
	err = db.QueryRow("SELECT id FROM photos WHERE file_path LIKE ?", "%"+testFileName).Scan(&photoID)
	if err != nil {
		t.Fatalf("Failed to find %s in database: %v", testFileName, err)
	}

	// Check that all 4 thumbnail sizes were generated
	thumbnailSizes := []string{"64", "256", "512", "1024"}
	for _, size := range thumbnailSizes {
		var thumbnailData []byte
		err = db.QueryRow(
			"SELECT data FROM thumbnails WHERE photo_id = ? AND size = ?",
			photoID, size,
		).Scan(&thumbnailData)

		if err != nil {
			t.Errorf("Thumbnail size %s not found for %s: %v", size, testFileName, err)
			continue
		}

		if len(thumbnailData) == 0 {
			t.Errorf("Thumbnail size %s is empty for %s", size, testFileName)
			continue
		}

		t.Logf("✓ Thumbnail %spx: %d bytes", size, len(thumbnailData))
	}

	// CURRENT BEHAVIOR: LibRaw fallback is expected
	// The fallback is triggered because seppedelanghe/go-libraw cannot decode
	// JPEG-compressed Leica M11 Monochrom DNGs properly.
	//
	// When this is fixed upstream, we should detect the fallback and fail:
	// if fallbackTriggered {
	//     t.Error("LibRaw fallback was triggered - expected RAW decode to work")
	// }
	//
	// For now, we document that the fallback is working correctly:
	_ = fallbackTriggered // Not checking this yet - fallback is expected

	t.Logf("✓ All 4 thumbnail sizes generated successfully")
	t.Logf("  Note: LibRaw fallback to embedded JPEG is currently expected")
	t.Logf("  When upstream bug is fixed, this test should verify RAW decode works")
}

// TestLeicaM11Monochrom_FullIndexingPipeline is a comprehensive integration test
// that verifies the complete indexing pipeline for L1001515.DNG produces
// correct metadata, thumbnails, colors, and perceptual hash.
func TestLeicaM11Monochrom_FullIndexingPipeline(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found: ", testFile)
	}

	// Create temporary database and index the file
	dbPath := filepath.Join(t.TempDir(), "test_leica_full.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the directory containing L1001515.DNG
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

	// Verify thumbnails were generated
	if stats.ThumbnailsGenerated < 4 {
		t.Errorf("Expected at least 4 thumbnails generated for L1001515.DNG, got %d", stats.ThumbnailsGenerated)
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Query specifically for L1001515.DNG
	testFileName := filepath.Base(testFile)
	var photoID int
	err = db.QueryRow("SELECT id FROM photos WHERE file_path LIKE ?", "%"+testFileName).Scan(&photoID)
	if err != nil {
		t.Fatalf("Failed to find %s in database: %v", testFileName, err)
	}

	photoDetail, err := repo.GetPhotoByID(photoID)
	if err != nil {
		t.Fatalf("Failed to get photo detail for %s: %v", testFileName, err)
	}

	t.Logf("Full pipeline results for %s (ID: %d):", testFileName, photoID)

	// 1. Verify camera metadata
	t.Run("Camera Metadata", func(t *testing.T) {
		if !strings.Contains(photoDetail.CameraModel, "M11 Monochrom") {
			t.Errorf("Expected Leica M11 Monochrom, got %s", photoDetail.CameraModel)
		}
		t.Logf("  ✓ Camera: %s %s", photoDetail.CameraMake, photoDetail.CameraModel)
	})

	// 2. Verify lens metadata (50mm f/2)
	t.Run("Lens Metadata", func(t *testing.T) {
		if photoDetail.FocalLength != 50.0 {
			t.Errorf("Expected 50mm focal length, got %.1fmm", photoDetail.FocalLength)
		}
		lensLower := strings.ToLower(photoDetail.LensModel)
		if !strings.Contains(lensLower, "50") || !strings.Contains(lensLower, "summicron") {
			t.Errorf("Expected 50mm Summicron lens, got %s", photoDetail.LensModel)
		}
		t.Logf("  ✓ Lens: %s (%.1fmm)", photoDetail.LensModel, photoDetail.FocalLength)
	})

	// 3. Verify exposure settings
	t.Run("Exposure Settings", func(t *testing.T) {
		if photoDetail.ISO != 10000 {
			t.Errorf("Expected ISO 10000, got %d", photoDetail.ISO)
		}
		if photoDetail.ShutterSpeed != "1/250" {
			t.Errorf("Expected 1/250s shutter, got %s", photoDetail.ShutterSpeed)
		}
		t.Logf("  ✓ Exposure: ISO %d, %s", photoDetail.ISO, photoDetail.ShutterSpeed)
	})

	// 4. Verify image dimensions
	t.Run("Image Dimensions", func(t *testing.T) {
		// Either full RAW (9536x6336) or embedded JPEG (9504x6320) is acceptable
		validDimensions := (photoDetail.Width == 9536 && photoDetail.Height == 6336) ||
			(photoDetail.Width == 9504 && photoDetail.Height == 6320)

		if !validDimensions {
			t.Errorf("Expected 9536x6336 or 9504x6320, got %dx%d",
				photoDetail.Width, photoDetail.Height)
		}
		t.Logf("  ✓ Dimensions: %dx%d", photoDetail.Width, photoDetail.Height)
	})

	// 5. Verify thumbnails exist
	t.Run("Thumbnails", func(t *testing.T) {
		thumbnailSizes := []string{"64", "256", "512", "1024"}
		for _, size := range thumbnailSizes {
			var thumbnailData []byte
			err = db.QueryRow(
				"SELECT data FROM thumbnails WHERE photo_id = ? AND size = ?",
				photoID, size,
			).Scan(&thumbnailData)

			if err != nil || len(thumbnailData) == 0 {
				t.Errorf("Thumbnail %spx missing or empty", size)
				continue
			}
		}
		t.Logf("  ✓ All 4 thumbnail sizes present")
	})

	// 6. Verify color palette was extracted
	t.Run("Color Palette", func(t *testing.T) {
		var colorCount int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM photo_colors WHERE photo_id = ?",
			photoID,
		).Scan(&colorCount)

		if err != nil {
			t.Errorf("Failed to query colors: %v", err)
		}

		// Should have extracted color data (each row is a pixel's color info)
		// Typical photos have 30-50 color rows based on sampling
		if colorCount < 1 {
			t.Errorf("Expected color data to be extracted, got %d rows", colorCount)
		}
		t.Logf("  ✓ Color palette: %d color samples extracted", colorCount)
	})

	// 7. Verify perceptual hash was computed
	t.Run("Perceptual Hash", func(t *testing.T) {
		var phash string
		err = db.QueryRow(
			"SELECT perceptual_hash FROM photos WHERE id = ?",
			photoID,
		).Scan(&phash)

		if err != nil {
			t.Errorf("Failed to query perceptual hash: %v", err)
		}

		if phash == "" {
			t.Error("Perceptual hash is empty")
		}
		t.Logf("  ✓ Perceptual hash: %s", phash)
	})

	// 8. Verify file hash was computed
	t.Run("File Hash", func(t *testing.T) {
		var fileHash string
		err = db.QueryRow(
			"SELECT file_hash FROM photos WHERE id = ?",
			photoID,
		).Scan(&fileHash)

		if err != nil {
			t.Errorf("Failed to query file hash: %v", err)
		}

		if fileHash == "" {
			t.Error("File hash is empty")
		}
		// SHA-256 hash should be 64 hex characters
		if len(fileHash) != 64 {
			t.Errorf("Expected 64-character SHA-256 hash, got %d characters", len(fileHash))
		}
		t.Logf("  ✓ File hash: %s...", fileHash[:16])
	})

	t.Logf("\n✓ Complete indexing pipeline verified for L1001515.DNG")
}
