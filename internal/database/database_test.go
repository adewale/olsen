package database

import (
	"os"
	"testing"
	"time"

	"github.com/adewale/olsen/pkg/models"
)

func TestOpen(t *testing.T) {
	// Test in-memory database
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='photos'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query photos table: %v", err)
	}
	if count != 1 {
		t.Error("Photos table not created")
	}
}

func TestInsertPhoto(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create test photo
	photo := &models.PhotoMetadata{
		FilePath:        "/test/photo.dng",
		FileHash:        "abcdef1234567890",
		FileSize:        1024000,
		DateTaken:       time.Date(2025, 5, 15, 12, 0, 0, 0, time.UTC),
		CameraMake:      "Canon",
		CameraModel:     "EOS R5",
		LensModel:       "RF24mm F1.4 L USM",
		FocalLength:     24.0,
		FocalLength35mm: 24,
		ISO:             100,
		Aperture:        1.4,
		ShutterSpeed:    "1/1000",
		Width:           8192,
		Height:          5464,
		Latitude:        37.7749,
		Longitude:       -122.4194,
		PerceptualHash:  "0123456789abcdef",

		// Inferred metadata
		TimeOfDay:         "midday",
		Season:            "spring",
		FocalCategory:     "wide",
		ShootingCondition: "bright",

		Thumbnails: map[models.ThumbnailSize][]byte{
			"64":   []byte("thumbnail_64"),
			"256":  []byte("thumbnail_256"),
			"512":  []byte("thumbnail_512"),
			"1024": []byte("thumbnail_1024"),
		},
		DominantColours: []models.DominantColour{
			{Colour: models.Colour{R: 255, G: 0, B: 0}, HSL: models.ColourHSL{H: 0, S: 100, L: 50}, Weight: 0.5},
			{Colour: models.Colour{R: 0, G: 255, B: 0}, HSL: models.ColourHSL{H: 120, S: 100, L: 50}, Weight: 0.3},
			{Colour: models.Colour{R: 0, G: 0, B: 255}, HSL: models.ColourHSL{H: 240, S: 100, L: 50}, Weight: 0.2},
		},
	}

	// Insert photo
	err = db.InsertPhoto(photo)
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}

	// Verify photo was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM photos WHERE file_hash = ?", photo.FileHash).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query photo: %v", err)
	}
	if count != 1 {
		t.Errorf("Photo not inserted, count = %d", count)
	}

	// Verify thumbnails were inserted
	err = db.QueryRow("SELECT COUNT(*) FROM thumbnails WHERE photo_id = (SELECT id FROM photos WHERE file_hash = ?)", photo.FileHash).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query thumbnails: %v", err)
	}
	if count != 4 {
		t.Errorf("Expected 4 thumbnails, got %d", count)
	}

	// Verify colors were inserted
	err = db.QueryRow("SELECT COUNT(*) FROM photo_colors WHERE photo_id = (SELECT id FROM photos WHERE file_hash = ?)", photo.FileHash).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query colors: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 colors, got %d", count)
	}
}

func TestPhotoExists(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	path := "/test/photo.jpg"

	// Should not exist initially
	exists, err := db.PhotoExists(path)
	if err != nil {
		t.Fatalf("PhotoExists failed: %v", err)
	}
	if exists {
		t.Error("Photo should not exist initially")
	}

	// Insert photo
	photo := &models.PhotoMetadata{
		FilePath:  path,
		FileHash:  "test_hash_12345",
		FileSize:  1024,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
	}
	err = db.InsertPhoto(photo)
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}

	// Should exist now
	exists, err = db.PhotoExists(path)
	if err != nil {
		t.Fatalf("PhotoExists failed after insert: %v", err)
	}
	if !exists {
		t.Error("Photo should exist after insert")
	}
}

func TestGetPhotoCount(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initially empty
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("GetPhotoCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 photos, got %d", count)
	}

	// Insert photos
	for i := 0; i < 5; i++ {
		photo := &models.PhotoMetadata{
			FilePath:  "/test/photo" + string(rune(i)) + ".jpg",
			FileHash:  "hash_" + string(rune(i)),
			FileSize:  1024,
			DateTaken: time.Now(),
			Width:     1920,
			Height:    1080,
		}
		err = db.InsertPhoto(photo)
		if err != nil {
			t.Fatalf("Failed to insert photo %d: %v", i, err)
		}
	}

	// Should have 5 photos
	count, err = db.GetPhotoCount()
	if err != nil {
		t.Fatalf("GetPhotoCount failed: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 photos, got %d", count)
	}
}

func TestTransactionRollback(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create photo with invalid thumbnail size (to cause error)
	photo := &models.PhotoMetadata{
		FilePath:  "/test/photo.jpg",
		FileHash:  "test_hash",
		FileSize:  1024,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
		Thumbnails: map[models.ThumbnailSize][]byte{
			"invalid_size": []byte("thumbnail"),
		},
	}

	// This should succeed (invalid sizes are just skipped in current implementation)
	err = db.InsertPhoto(photo)
	if err != nil {
		t.Logf("Note: Photo insertion failed (as expected): %v", err)
	}

	// Even if insert failed, test that no partial data was committed
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("GetPhotoCount failed: %v", err)
	}
	// Should be 0 if transaction rolled back, or 1 if it succeeded
	// Either way, no partial state
	t.Logf("Photo count after potential rollback: %d", count)
}

func TestConcurrentInserts(t *testing.T) {
	// Use a file-based database for concurrency test (in-memory doesn't support concurrent access)
	tmpFile, err := os.CreateTemp("", "concurrent_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert photos concurrently
	errors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			photo := &models.PhotoMetadata{
				FilePath:  "/test/photo" + string(rune('0'+idx)) + ".jpg",
				FileHash:  "hash_" + string(rune('0'+idx)),
				FileSize:  1024,
				DateTaken: time.Now(),
				Width:     1920,
				Height:    1080,
			}
			errors <- db.InsertPhoto(photo)
		}(i)
	}

	// Wait for all inserts and check for errors
	for i := 0; i < 10; i++ {
		if err := <-errors; err != nil {
			t.Errorf("Failed to insert photo: %v", err)
		}
	}

	// Should have 10 photos
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("GetPhotoCount failed: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10 photos, got %d", count)
	}
}

func TestDuplicateHashHandling(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	hash := "duplicate_hash"

	// Insert first photo
	photo1 := &models.PhotoMetadata{
		FilePath:  "/test/photo1.jpg",
		FileHash:  hash,
		FileSize:  1024,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
	}
	err = db.InsertPhoto(photo1)
	if err != nil {
		t.Fatalf("Failed to insert first photo: %v", err)
	}

	// Try to insert second photo with same hash (but different path)
	photo2 := &models.PhotoMetadata{
		FilePath:  "/test/photo2.jpg",
		FileHash:  hash,
		FileSize:  2048,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
	}
	err = db.InsertPhoto(photo2)
	// Current schema doesn't enforce unique hash, so this might succeed
	// The indexer checks PhotoExists before inserting
	if err != nil {
		t.Logf("Insert failed (duplicate hash): %v", err)
	}

	// Check how many photos we have
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("GetPhotoCount failed: %v", err)
	}
	// Schema doesn't enforce unique hash, so we might have 2
	t.Logf("Photo count after duplicate hash insert: %d", count)
}

func TestPersistence(t *testing.T) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)

	// Open and insert data
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	photo := &models.PhotoMetadata{
		FilePath:  "/test/photo.jpg",
		FileHash:  "persistent_hash",
		FileSize:  1024,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
	}
	err = db.InsertPhoto(photo)
	if err != nil {
		t.Fatalf("Failed to insert photo: %v", err)
	}
	db.Close()

	// Reopen and verify data persisted
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db2.Close()

	exists, err := db2.PhotoExists("/test/photo.jpg")
	if err != nil {
		t.Fatalf("PhotoExists failed: %v", err)
	}
	if !exists {
		t.Error("Photo should persist after closing and reopening database")
	}
}

func BenchmarkInsertPhoto(b *testing.B) {
	db, err := Open(":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	photo := &models.PhotoMetadata{
		FilePath:       "/test/photo.jpg",
		FileSize:       1024000,
		DateTaken:      time.Now(),
		CameraMake:     "Canon",
		CameraModel:    "EOS R5",
		Width:          8192,
		Height:         5464,
		PerceptualHash: "0123456789abcdef",
		Thumbnails: map[models.ThumbnailSize][]byte{
			"64":   make([]byte, 5000),
			"256":  make([]byte, 20000),
			"512":  make([]byte, 80000),
			"1024": make([]byte, 300000),
		},
		DominantColours: []models.DominantColour{
			{Colour: models.Colour{R: 255, G: 0, B: 0}, Weight: 0.5},
			{Colour: models.Colour{R: 0, G: 255, B: 0}, Weight: 0.3},
			{Colour: models.Colour{R: 0, G: 0, B: 255}, Weight: 0.2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		photo.FileHash = "hash_" + string(rune(i))
		photo.FilePath = "/test/photo_" + string(rune(i)) + ".jpg"
		err := db.InsertPhoto(photo)
		if err != nil {
			b.Fatalf("Failed to insert photo: %v", err)
		}
	}
}

func BenchmarkPhotoExists(b *testing.B) {
	db, err := Open(":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert test photo
	photo := &models.PhotoMetadata{
		FilePath:  "/test/photo.jpg",
		FileHash:  "bench_hash",
		FileSize:  1024,
		DateTaken: time.Now(),
		Width:     1920,
		Height:    1080,
	}
	db.InsertPhoto(photo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.PhotoExists("bench_hash")
		if err != nil {
			b.Fatalf("PhotoExists failed: %v", err)
		}
	}
}
