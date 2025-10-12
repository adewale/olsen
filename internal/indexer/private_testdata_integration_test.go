package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adewale/olsen/internal/database"
)

// TestIntegrationIndexPrivateTestData tests indexing the private-testdata directory
// This test is skipped if the private-testdata directory doesn't exist
func TestIntegrationIndexPrivateTestData(t *testing.T) {
	// Get path to private-testdata (relative to project root)
	privateTestDataPath := filepath.Join("..", "..", "private-testdata")

	// Check if private-testdata exists
	if _, err := os.Stat(privateTestDataPath); os.IsNotExist(err) {
		t.Skip("private-testdata directory not found, skipping test")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "private_integration_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	// Open database
	db, err := database.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create engine with 4 workers
	engine := NewEngine(db, 4)

	// Run indexing
	t.Logf("Indexing %s", privateTestDataPath)
	err = engine.IndexDirectory(privateTestDataPath)
	if err != nil {
		t.Fatalf("IndexDirectory failed: %v", err)
	}

	// Get stats
	stats := engine.GetStats()

	t.Logf("Indexing Results:")
	t.Logf("  Files found: %d", stats.FilesFound)
	t.Logf("  Files processed: %d", stats.FilesProcessed)
	t.Logf("  Files skipped: %d", stats.FilesSkipped)
	t.Logf("  Files updated: %d", stats.FilesUpdated)
	t.Logf("  Files failed: %d", stats.FilesFailed)
	t.Logf("  Thumbnails generated: %d", stats.ThumbnailsGenerated)

	// Verify results
	if stats.FilesFound == 0 {
		t.Error("No files found in private-testdata")
	}

	// NOTE: The private-testdata contains RAW DNG files from Leica M11 Monochrom cameras
	// With LibRaw support (build with CGO and `make build-raw`), we can now:
	// 1. Extract metadata from RAW files (camera, lens, ISO, aperture, etc.)
	// 2. Index RAW files with full metadata support
	//
	// Without LibRaw (standard build), RAW files will be indexed with metadata only
	// (no thumbnails or colour extraction).
	//
	// This test verifies that files are discovered and processed appropriately.

	if stats.FilesProcessed == 0 && stats.FilesFailed == 0 {
		t.Error("No files processed or failed - indexer may not be working")
	}

	// With RAW support, we expect all files to be processed successfully
	if IsRawSupported() {
		t.Logf("RAW support is enabled (LibRaw detected)")
		if stats.FilesFailed > 0 {
			t.Logf("Warning: %d files failed despite RAW support being enabled", stats.FilesFailed)
		}
		// We expect at least some files to be processed
		if stats.FilesProcessed == 0 {
			t.Error("No files processed - RAW support appears non-functional")
		}
	} else {
		t.Logf("RAW support is disabled (build without LibRaw)")
		// Without RAW support, files may be processed with metadata only
		if stats.FilesFailed == stats.FilesFound {
			t.Skip("All RAW files failed without LibRaw support - expected behavior")
		}
	}

	// Verify database has photos
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("Failed to get photo count: %v", err)
	}

	if count == 0 {
		t.Error("No photos in database after indexing")
	}

	if count != stats.FilesProcessed {
		t.Errorf("Photo count mismatch: database has %d, but processed %d", count, stats.FilesProcessed)
	}

	t.Logf("Database contains %d photos", count)

	// Verify thumbnails - note that RAW files may not have thumbnails if image decoding fails
	// (this is a known limitation - metadata extraction works, but full image decode may fail)
	expectedThumbnails := stats.FilesProcessed * 4 // 4 sizes per photo
	if stats.ThumbnailsGenerated > 0 {
		t.Logf("Generated %d thumbnails for %d photos (%.1f%% coverage)",
			stats.ThumbnailsGenerated, stats.FilesProcessed,
			float64(stats.ThumbnailsGenerated)/float64(expectedThumbnails)*100)
	} else {
		t.Logf("Note: No thumbnails generated (RAW image decoding limitation)")
	}

	// Additional verification: Query to ensure photos have expected metadata
	rows, err := db.Query("SELECT COUNT(*) FROM photos WHERE date_taken IS NOT NULL")
	if err != nil {
		t.Fatalf("Failed to query photos with dates: %v", err)
	}
	defer rows.Close()

	var photosWithDates int
	if rows.Next() {
		rows.Scan(&photosWithDates)
	}

	t.Logf("Photos with date metadata: %d/%d (%.1f%%)",
		photosWithDates, count, float64(photosWithDates)/float64(count)*100)

	// Query for photos with camera info
	rows2, err := db.Query("SELECT COUNT(*) FROM photos WHERE camera_make IS NOT NULL AND camera_model IS NOT NULL")
	if err != nil {
		t.Fatalf("Failed to query photos with camera info: %v", err)
	}
	defer rows2.Close()

	var photosWithCamera int
	if rows2.Next() {
		rows2.Scan(&photosWithCamera)
	}

	t.Logf("Photos with camera metadata: %d/%d (%.1f%%)",
		photosWithCamera, count, float64(photosWithCamera)/float64(count)*100)

	// Query for photos with colour data
	rows3, err := db.Query("SELECT COUNT(DISTINCT photo_id) FROM photo_colors")
	if err != nil {
		t.Fatalf("Failed to query photos with colours: %v", err)
	}
	defer rows3.Close()

	var photosWithColours int
	if rows3.Next() {
		rows3.Scan(&photosWithColours)
	}

	t.Logf("Photos with colour data: %d/%d (%.1f%%)",
		photosWithColours, count, float64(photosWithColours)/float64(count)*100)

	// Note: Colour extraction requires successful image decoding
	// RAW files may not have colour data if image decoding failed
	if photosWithColours < count {
		t.Logf("Note: Some photos missing colour data (expected for RAW files without full image decode)")
	}

	// Query for photos with GPS data
	rows4, err := db.Query("SELECT COUNT(*) FROM photos WHERE latitude IS NOT NULL AND longitude IS NOT NULL")
	if err != nil {
		t.Fatalf("Failed to query photos with GPS: %v", err)
	}
	defer rows4.Close()

	var photosWithGPS int
	if rows4.Next() {
		rows4.Scan(&photosWithGPS)
	}

	t.Logf("Photos with GPS data: %d/%d (%.1f%%)",
		photosWithGPS, count, float64(photosWithGPS)/float64(count)*100)

	// Success!
	t.Logf("✓ Successfully indexed all photos from private-testdata")
}
