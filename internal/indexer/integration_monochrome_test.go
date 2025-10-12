//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/explorer"
	"github.com/adewale/olsen/internal/indexer"
)

// TestIntegrationMonochromeDNG tests the complete pipeline for monochrome JPEG-compressed DNG files
// This ensures that:
// 1. RAW decoding works (no buffer overflow)
// 2. Thumbnail generation works (JPEG encoding of Gray images)
// 3. Color extraction works
// 4. Database storage works
func TestIntegrationMonochromeDNG(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_mono.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the file
	engine := indexer.NewEngine(db, 1)

	// Get directory containing test file
	testDir := filepath.Dir(testFile)
	err = engine.IndexDirectory(testDir)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	// Verify no files failed
	stats := engine.GetStats()
	if stats.FilesFailed > 0 {
		t.Errorf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	// Verify files were processed
	if stats.FilesProcessed == 0 {
		t.Error("No files were processed")
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Get recent photos
	photos, err := repo.GetRecentPhotos(100)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}

	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	photoID := photos[0].ID

	// Verify thumbnail was generated (try all sizes in case some were skipped)
	var thumbData []byte
	var thumbSize string
	for _, size := range []string{"64", "256", "512", "1024"} {
		data, err := repo.GetThumbnail(photoID, size)
		if err == nil && len(data) > 0 {
			thumbData = data
			thumbSize = size
			break
		}
	}

	if len(thumbData) == 0 {
		t.Error("No thumbnails found in database (tried all sizes)")
	} else {
		// Verify thumbnail is valid JPEG by checking header
		if len(thumbData) < 2 || thumbData[0] != 0xFF || thumbData[1] != 0xD8 {
			t.Error("Thumbnail is not a valid JPEG (missing JPEG header)")
		} else {
			t.Logf("✓ Found valid JPEG thumbnail at size %s", thumbSize)
		}
	}

	// Verify colors were extracted (query database directly)
	var colorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM photo_colors WHERE photo_id = ?", photoID).Scan(&colorCount)
	if err != nil {
		t.Errorf("Failed to count colors: %v", err)
	}
	if colorCount == 0 {
		t.Error("No colors extracted from photo")
	}

	t.Logf("✓ Complete pipeline test passed for monochrome DNG")
	t.Logf("  Photos indexed: %d", len(photos))
	t.Logf("  Thumbnail size: %d bytes", len(thumbData))
	t.Logf("  Colors extracted: %d", colorCount)
}

// TestIntegrationMonochromeBatch tests batch processing of monochrome DNGs
func TestIntegrationMonochromeBatch(t *testing.T) {
	testDir := "../../private-testdata/2024-12-18"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test directory not found (requires private-testdata)")
	}

	// Count DNG files
	var fileCount int
	filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".DNG" || filepath.Ext(path) == ".dng" {
			fileCount++
		}
		return nil
	})

	if fileCount == 0 {
		t.Skip("No DNG files found in test directory")
	}

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_mono_batch.db")
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

	// Verify results
	stats := engine.GetStats()
	t.Logf("Batch indexing results:")
	t.Logf("  Files found: %d", stats.FilesFound)
	t.Logf("  Files processed: %d", stats.FilesProcessed)
	t.Logf("  Files failed: %d", stats.FilesFailed)

	if stats.FilesFailed > 0 {
		t.Errorf("Expected 0 failed files, got %d", stats.FilesFailed)
	}

	if stats.FilesProcessed != fileCount {
		t.Errorf("Expected %d processed files, got %d", fileCount, stats.FilesProcessed)
	}

	// Create repository to query database
	repo := explorer.NewRepository(db)

	// Verify all photos have thumbnails and colors
	photos, err := repo.GetRecentPhotos(1000)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}

	for _, photo := range photos {
		// Check thumbnail exists
		thumbData, err := repo.GetThumbnail(photo.ID, "256")
		if err != nil || len(thumbData) == 0 {
			t.Errorf("Photo %d missing thumbnail: %v", photo.ID, err)
		}

		// Check colors extracted (query database directly)
		var colorCount int
		err = db.QueryRow("SELECT COUNT(*) FROM photo_colors WHERE photo_id = ?", photo.ID).Scan(&colorCount)
		if err != nil || colorCount == 0 {
			t.Errorf("Photo %d missing colors: %v", photo.ID, err)
		}
	}

	t.Logf("✓ All %d photos have thumbnails and colors", len(photos))
}
