package explorer

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
)

// TestThumbnailFallback tests that GetThumbnail falls back to smaller sizes
func TestThumbnailFallback(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_fallback.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert a test photo
	tx, _ := db.Begin()
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, indexed_at, last_modified, date_taken)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/test/photo.dng", "abc123", 1000000, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()

	// Insert only 64px thumbnail (simulating upscale-skipped case)
	_, err = tx.Exec(`
		INSERT INTO thumbnails (photo_id, size, data)
		VALUES (?, '64', ?)
	`, photoID, []byte("fake_64px_thumbnail_data"))
	if err != nil {
		t.Fatalf("Failed to insert thumbnail: %v", err)
	}
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test 1: Request 1024px, should fall back to 64px
	t.Run("Request 1024 falls back to 64", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "1024")
		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "fake_64px_thumbnail_data" {
			t.Errorf("Expected 64px thumbnail data, got: %s", string(data))
		}
	})

	// Test 2: Request 512px, should fall back to 64px
	t.Run("Request 512 falls back to 64", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "512")
		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "fake_64px_thumbnail_data" {
			t.Errorf("Expected 64px thumbnail data, got: %s", string(data))
		}
	})

	// Test 3: Request 256px, should fall back to 64px
	t.Run("Request 256 falls back to 64", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "256")
		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "fake_64px_thumbnail_data" {
			t.Errorf("Expected 64px thumbnail data, got: %s", string(data))
		}
	})

	// Test 4: Request 64px directly, should succeed
	t.Run("Request 64 directly", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "64")
		if err != nil {
			t.Errorf("Expected direct request to succeed, got error: %v", err)
		}
		if string(data) != "fake_64px_thumbnail_data" {
			t.Errorf("Expected 64px thumbnail data, got: %s", string(data))
		}
	})
}

// TestThumbnailFallbackWithMultipleSizes tests fallback with partial thumbnail sets
func TestThumbnailFallbackWithMultipleSizes(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_fallback_multi.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert a test photo
	tx, _ := db.Begin()
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, indexed_at, last_modified, date_taken)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/test/photo2.dng", "def456", 2000000, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()

	// Insert 64px and 256px thumbnails (but not 512px or 1024px)
	_, err = tx.Exec(`
		INSERT INTO thumbnails (photo_id, size, data) VALUES
		(?, '64', ?),
		(?, '256', ?)
	`, photoID, []byte("fake_64px"), photoID, []byte("fake_256px"))
	if err != nil {
		t.Fatalf("Failed to insert thumbnails: %v", err)
	}
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test 1: Request 1024px, should fall back to 256px
	t.Run("Request 1024 falls back to 256", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "1024")
		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "fake_256px" {
			t.Errorf("Expected 256px thumbnail data, got: %s", string(data))
		}
	})

	// Test 2: Request 512px, should fall back to 256px
	t.Run("Request 512 falls back to 256", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "512")
		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "fake_256px" {
			t.Errorf("Expected 256px thumbnail data, got: %s", string(data))
		}
	})

	// Test 3: Request 256px directly, should get 256px
	t.Run("Request 256 directly", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "256")
		if err != nil {
			t.Errorf("Expected direct request to succeed, got error: %v", err)
		}
		if string(data) != "fake_256px" {
			t.Errorf("Expected 256px thumbnail data, got: %s", string(data))
		}
	})

	// Test 4: Request 64px directly, should get 64px
	t.Run("Request 64 directly", func(t *testing.T) {
		data, err := repo.GetThumbnail(int(photoID), "64")
		if err != nil {
			t.Errorf("Expected direct request to succeed, got error: %v", err)
		}
		if string(data) != "fake_64px" {
			t.Errorf("Expected 64px thumbnail data, got: %s", string(data))
		}
	})
}

// TestThumbnailFallbackWithNoThumbnails tests error handling when no thumbnails exist
func TestThumbnailFallbackWithNoThumbnails(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_fallback_none.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert a test photo WITHOUT thumbnails
	tx, _ := db.Begin()
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, indexed_at, last_modified, date_taken)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/test/photo3.dng", "ghi789", 3000000, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test: Request any size, should fail gracefully
	t.Run("Request with no thumbnails returns error", func(t *testing.T) {
		_, err := repo.GetThumbnail(int(photoID), "256")
		if err == nil {
			t.Error("Expected error when no thumbnails exist, got nil")
		}
	})
}

// TestGetThumbnailWithTimestamp tests the timestamp variant of GetThumbnail
func TestGetThumbnailWithTimestamp(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_timestamp.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert a test photo with known timestamp
	indexedTime := time.Date(2025, 3, 25, 14, 30, 0, 0, time.UTC)
	tx, _ := db.Begin()
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, indexed_at, last_modified, date_taken)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/test/photo.dng", "abc123", 1000000, indexedTime.Format(time.RFC3339), time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()

	// Insert only 64px thumbnail
	_, err = tx.Exec(`
		INSERT INTO thumbnails (photo_id, size, data)
		VALUES (?, '64', ?)
	`, photoID, []byte("thumbnail_data"))
	if err != nil {
		t.Fatalf("Failed to insert thumbnail: %v", err)
	}
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test: Request 256px, should fall back to 64px AND return correct timestamp
	t.Run("Fallback returns correct timestamp", func(t *testing.T) {
		data, timestamp, err := repo.GetThumbnailWithTimestamp(int(photoID), "256")
		if err != nil {
			t.Fatalf("Expected fallback to succeed, got error: %v", err)
		}
		if string(data) != "thumbnail_data" {
			t.Errorf("Expected thumbnail data, got: %s", string(data))
		}

		// Check timestamp is approximately correct (within 1 second tolerance)
		if timestamp.Sub(indexedTime).Abs() > time.Second {
			t.Errorf("Expected timestamp %v, got %v", indexedTime, timestamp)
		}
	})
}

// TestDateParsing tests RFC3339 date parsing from database
func TestDateParsing(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_dates.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert photos with various date formats (as SQLite would store them)
	testDate := time.Date(2025, 3, 25, 13, 16, 5, 0, time.UTC)
	tx, _ := db.Begin()

	// SQLite stores dates in RFC3339 format (with 'T' and 'Z')
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, last_modified, date_taken, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "/test/photo1.dng", "hash1", 1000000, testDate.Format(time.RFC3339), testDate.Format(time.RFC3339), testDate.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test: GetRecentPhotos should parse dates correctly
	t.Run("GetRecentPhotos parses RFC3339 dates", func(t *testing.T) {
		photos, err := repo.GetRecentPhotos(10)
		if err != nil {
			t.Fatalf("Failed to get recent photos: %v", err)
		}

		if len(photos) == 0 {
			t.Fatal("Expected at least one photo")
		}

		photo := photos[0]

		// Check the date was parsed correctly (not zero)
		if photo.DateTaken.IsZero() {
			t.Error("DateTaken is zero - RFC3339 parsing failed")
		}

		// Check the date matches what we inserted (within 1 second tolerance)
		if photo.DateTaken.Sub(testDate).Abs() > time.Second {
			t.Errorf("Expected date %v, got %v", testDate, photo.DateTaken)
		}
	})

	// Test: GetPhotoByID should parse dates correctly
	t.Run("GetPhotoByID parses RFC3339 dates", func(t *testing.T) {
		photo, err := repo.GetPhotoByID(int(photoID))
		if err != nil {
			t.Fatalf("Failed to get photo by ID: %v", err)
		}

		// Check the date was parsed correctly (not zero)
		if photo.DateTaken.IsZero() {
			t.Error("DateTaken is zero - RFC3339 parsing failed")
		}

		// Check the date matches what we inserted
		if photo.DateTaken.Sub(testDate).Abs() > time.Second {
			t.Errorf("Expected date %v, got %v", testDate, photo.DateTaken)
		}
	})
}

// TestDateParsingWithNullDates tests handling of NULL dates
func TestDateParsingWithNullDates(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "test_null_dates.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert photo with NULL date_taken
	tx, _ := db.Begin()
	result, err := tx.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, last_modified, date_taken, indexed_at)
		VALUES (?, ?, ?, ?, NULL, ?)
	`, "/test/no_date.jpg", "hash2", 500000, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	photoID, _ := result.LastInsertId()
	tx.Commit()

	// Create repository
	repo := NewRepository(db)

	// Test: GetPhotoByID should handle NULL dates gracefully
	t.Run("GetPhotoByID handles NULL dates", func(t *testing.T) {
		photo, err := repo.GetPhotoByID(int(photoID))
		if err != nil {
			t.Fatalf("Failed to get photo by ID: %v", err)
		}

		// DateTaken should be zero value for NULL dates
		if !photo.DateTaken.IsZero() {
			t.Errorf("Expected zero time for NULL date, got %v", photo.DateTaken)
		}
	})
}
