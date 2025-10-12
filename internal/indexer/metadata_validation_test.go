//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/explorer"
	"github.com/adewale/olsen/internal/indexer"
)

// TestMetadataValidation verifies that displayed metadata matches original image EXIF
func TestMetadataValidation(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	// Extract metadata directly from file using our EXIF reader
	t.Logf("Extracting EXIF metadata from: %s", testFile)
	originalMetadata, err := indexer.ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Create temporary database and index the file
	dbPath := filepath.Join(t.TempDir(), "test_metadata.db")
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

	// Get the photo detail (this is what's displayed on web page)
	photos, err := repo.GetRecentPhotos(1)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}
	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	photoDetail, err := repo.GetPhotoByID(photos[0].ID)
	if err != nil {
		t.Fatalf("Failed to get photo detail: %v", err)
	}

	// Verify each metadata field matches
	t.Run("Camera Make", func(t *testing.T) {
		if photoDetail.CameraMake != originalMetadata.CameraMake {
			t.Errorf("Camera make mismatch: got %q, want %q",
				photoDetail.CameraMake, originalMetadata.CameraMake)
		}
		t.Logf("✓ Camera Make: %s", photoDetail.CameraMake)
	})

	t.Run("Camera Model", func(t *testing.T) {
		if photoDetail.CameraModel != originalMetadata.CameraModel {
			t.Errorf("Camera model mismatch: got %q, want %q",
				photoDetail.CameraModel, originalMetadata.CameraModel)
		}
		t.Logf("✓ Camera Model: %s", photoDetail.CameraModel)
	})

	t.Run("Lens Model", func(t *testing.T) {
		if photoDetail.LensModel != originalMetadata.LensModel {
			t.Errorf("Lens model mismatch: got %q, want %q",
				photoDetail.LensModel, originalMetadata.LensModel)
		}
		t.Logf("✓ Lens Model: %s", photoDetail.LensModel)
	})

	t.Run("ISO", func(t *testing.T) {
		if photoDetail.ISO != originalMetadata.ISO {
			t.Errorf("ISO mismatch: got %d, want %d",
				photoDetail.ISO, originalMetadata.ISO)
		}
		t.Logf("✓ ISO: %d", photoDetail.ISO)
	})

	t.Run("Aperture", func(t *testing.T) {
		if photoDetail.Aperture != originalMetadata.Aperture {
			t.Errorf("Aperture mismatch: got %.1f, want %.1f",
				photoDetail.Aperture, originalMetadata.Aperture)
		}
		t.Logf("✓ Aperture: f/%.1f", photoDetail.Aperture)
	})

	t.Run("Shutter Speed", func(t *testing.T) {
		if photoDetail.ShutterSpeed != originalMetadata.ShutterSpeed {
			t.Errorf("Shutter speed mismatch: got %q, want %q",
				photoDetail.ShutterSpeed, originalMetadata.ShutterSpeed)
		}
		t.Logf("✓ Shutter Speed: %s", photoDetail.ShutterSpeed)
	})

	t.Run("Focal Length", func(t *testing.T) {
		if photoDetail.FocalLength != originalMetadata.FocalLength {
			t.Errorf("Focal length mismatch: got %.1f, want %.1f",
				photoDetail.FocalLength, originalMetadata.FocalLength)
		}
		t.Logf("✓ Focal Length: %.1fmm", photoDetail.FocalLength)
	})

	t.Run("Date Taken", func(t *testing.T) {
		// Compare dates with 1-second tolerance (database may truncate)
		diff := photoDetail.DateTaken.Sub(originalMetadata.DateTaken)
		if diff < -1*time.Second || diff > 1*time.Second {
			t.Errorf("Date taken mismatch: got %v, want %v (diff: %v)",
				photoDetail.DateTaken, originalMetadata.DateTaken, diff)
		}
		t.Logf("✓ Date Taken: %s", photoDetail.DateTaken.Format("2006-01-02 15:04:05"))
	})

	t.Run("Image Dimensions", func(t *testing.T) {
		if photoDetail.Width != originalMetadata.Width {
			t.Errorf("Width mismatch: got %d, want %d",
				photoDetail.Width, originalMetadata.Width)
		}
		if photoDetail.Height != originalMetadata.Height {
			t.Errorf("Height mismatch: got %d, want %d",
				photoDetail.Height, originalMetadata.Height)
		}
		t.Logf("✓ Dimensions: %dx%d", photoDetail.Width, photoDetail.Height)
	})

	t.Run("GPS Coordinates", func(t *testing.T) {
		if originalMetadata.Latitude != 0 && originalMetadata.Longitude != 0 {
			if photoDetail.Latitude != originalMetadata.Latitude {
				t.Errorf("Latitude mismatch: got %.6f, want %.6f",
					photoDetail.Latitude, originalMetadata.Latitude)
			}
			if photoDetail.Longitude != originalMetadata.Longitude {
				t.Errorf("Longitude mismatch: got %.6f, want %.6f",
					photoDetail.Longitude, originalMetadata.Longitude)
			}
			t.Logf("✓ GPS: %.6f, %.6f", photoDetail.Latitude, photoDetail.Longitude)
		} else {
			t.Logf("✓ GPS: Not available in this image")
		}
	})

	t.Logf("\n✓ All metadata fields validated successfully")
}

// TestMetadataBatchValidation validates metadata for multiple files
func TestMetadataBatchValidation(t *testing.T) {
	testDir := "../../private-testdata/2024-12-18"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test directory not found (requires private-testdata)")
	}

	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_batch_metadata.db")
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
	t.Logf("Indexed %d files", stats.FilesProcessed)

	// Create repository
	repo := explorer.NewRepository(db)

	// Get all photos
	photos, err := repo.GetRecentPhotos(1000)
	if err != nil {
		t.Fatalf("Failed to get photos: %v", err)
	}

	if len(photos) == 0 {
		t.Fatal("No photos found in database")
	}

	// Verify key fields are populated for all photos
	missingCamera := 0
	missingLens := 0
	missingDate := 0
	missingDimensions := 0

	for _, photo := range photos {
		if photo.CameraMake == "" || photo.CameraModel == "" {
			missingCamera++
		}

		// Get full photo detail
		detail, err := repo.GetPhotoByID(photo.ID)
		if err != nil {
			t.Errorf("Failed to get photo %d: %v", photo.ID, err)
			continue
		}

		if detail.LensModel == "" {
			missingLens++
		}
		if detail.DateTaken.IsZero() {
			missingDate++
		}
		if detail.Width == 0 || detail.Height == 0 {
			missingDimensions++
		}
	}

	// Report statistics
	totalPhotos := len(photos)
	t.Logf("\nMetadata Coverage Report:")
	t.Logf("  Total photos: %d", totalPhotos)
	t.Logf("  Camera info: %d/%d (%.1f%%)", totalPhotos-missingCamera, totalPhotos,
		float64(totalPhotos-missingCamera)/float64(totalPhotos)*100)
	t.Logf("  Lens info: %d/%d (%.1f%%)", totalPhotos-missingLens, totalPhotos,
		float64(totalPhotos-missingLens)/float64(totalPhotos)*100)
	t.Logf("  Date taken: %d/%d (%.1f%%)", totalPhotos-missingDate, totalPhotos,
		float64(totalPhotos-missingDate)/float64(totalPhotos)*100)
	t.Logf("  Dimensions: %d/%d (%.1f%%)", totalPhotos-missingDimensions, totalPhotos,
		float64(totalPhotos-missingDimensions)/float64(totalPhotos)*100)

	// Fail if any critical metadata is missing
	if missingCamera > 0 {
		t.Errorf("Missing camera info for %d photos", missingCamera)
	}
	if missingDate > 0 {
		t.Errorf("Missing date taken for %d photos", missingDate)
	}
	if missingDimensions > 0 {
		t.Errorf("Missing dimensions for %d photos", missingDimensions)
	}
}
