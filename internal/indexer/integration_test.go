package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adewale/olsen/internal/database"
)

// TestIntegrationIndexTestData tests indexing the testdata directory
func TestIntegrationIndexTestData(t *testing.T) {
	// Get path to testdata
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	// Check if testdata exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata directory not found, run: go run testdata/generate_fixtures.go")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "integration_test_*.db")
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

	// Create engine with 2 workers
	engine := NewEngine(db, 2)

	// Run indexing
	t.Logf("Indexing %s", testDataPath)
	err = engine.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("IndexDirectory failed: %v", err)
	}

	// Get stats
	stats := engine.GetStats()

	// Verify results
	if stats.FilesFound == 0 {
		t.Error("No files found in testdata")
	}

	if stats.FilesProcessed == 0 {
		t.Error("No files processed")
	}

	t.Logf("Found: %d, Processed: %d, Failed: %d", stats.FilesFound, stats.FilesProcessed, stats.FilesFailed)

	// Expected 4 files: test1.jpg, test2.jpg, scan1.bmp, subfolder/test3.jpg
	expectedFiles := 4
	if stats.FilesFound != expectedFiles {
		t.Errorf("Expected to find %d files, found %d", expectedFiles, stats.FilesFound)
	}

	// All files should be processed successfully
	if stats.FilesProcessed != stats.FilesFound {
		t.Errorf("Not all files processed: %d/%d", stats.FilesProcessed, stats.FilesFound)
	}

	if stats.FilesFailed != 0 {
		t.Errorf("Files failed: %d (expected 0)", stats.FilesFailed)
	}

	// Verify database has photos
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("Failed to get photo count: %v", err)
	}

	if count != stats.FilesProcessed {
		t.Errorf("Database photo count (%d) != files processed (%d)", count, stats.FilesProcessed)
	}

	t.Logf("Successfully indexed %d photos", count)

	// Verify thumbnails were generated
	if stats.ThumbnailsGenerated == 0 {
		t.Error("No thumbnails generated")
	}

	expectedThumbnails := stats.FilesProcessed * 4 // 4 sizes per photo
	if stats.ThumbnailsGenerated != expectedThumbnails {
		t.Errorf("Expected %d thumbnails, got %d", expectedThumbnails, stats.ThumbnailsGenerated)
	}

	// Verify hashes were computed
	if stats.HashesComputed != stats.FilesProcessed {
		t.Errorf("Expected %d hashes, got %d", stats.FilesProcessed, stats.HashesComputed)
	}

	// Performance check
	duration := stats.Duration()
	rate := stats.PhotosPerSecond()
	t.Logf("Duration: %v, Rate: %.2f photos/second", duration, rate)

	if rate == 0 {
		t.Error("Invalid processing rate")
	}
}

// TestIntegrationReIndexing tests that re-indexing skips existing photos
func TestIntegrationReIndexing(t *testing.T) {
	// Get path to testdata
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	// Check if testdata exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata directory not found")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "reindex_test_*.db")
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

	// First indexing pass
	engine1 := NewEngine(db, 2)
	err = engine1.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("First indexing failed: %v", err)
	}

	stats1 := engine1.GetStats()
	t.Logf("First pass: Found %d, Processed %d", stats1.FilesFound, stats1.FilesProcessed)

	// Second indexing pass (should skip all existing photos)
	engine2 := NewEngine(db, 2)
	err = engine2.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("Second indexing failed: %v", err)
	}

	stats2 := engine2.GetStats()
	t.Logf("Second pass: Found %d, Processed %d", stats2.FilesFound, stats2.FilesProcessed)

	// Should find same number of files
	if stats2.FilesFound != stats1.FilesFound {
		t.Errorf("File count changed: %d -> %d", stats1.FilesFound, stats2.FilesFound)
	}

	// Second pass processes files but skips database insert (files already exist)
	// The FilesProcessed counter still increments because processFile returns nil (no error)
	if stats2.FilesProcessed != stats2.FilesFound {
		t.Errorf("Second pass: processed %d, found %d", stats2.FilesProcessed, stats2.FilesFound)
	}

	// Verify no thumbnails or hashes were regenerated
	if stats2.ThumbnailsGenerated > 0 {
		t.Errorf("Second pass generated %d thumbnails, expected 0", stats2.ThumbnailsGenerated)
	}

	// Database count should remain the same
	count, err := db.GetPhotoCount()
	if err != nil {
		t.Fatalf("Failed to get photo count: %v", err)
	}

	if count != stats1.FilesProcessed {
		t.Errorf("Database count changed: expected %d, got %d", stats1.FilesProcessed, count)
	}
}

// TestIntegrationFileTypeSupport tests that all supported formats are indexed
func TestIntegrationFileTypeSupport(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata directory not found")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "filetype_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := database.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)

	// Find all files
	files, err := engine.findDNGFiles(testDataPath)
	if err != nil {
		t.Fatalf("Failed to find files: %v", err)
	}

	// Check that we have files of each type
	hasJPEG := false
	hasBMP := false

	for _, file := range files {
		ext := filepath.Ext(file)
		switch ext {
		case ".jpg", ".jpeg":
			hasJPEG = true
		case ".bmp":
			hasBMP = true
		}
	}

	if !hasJPEG {
		t.Error("No JPEG files found in testdata")
	}

	if !hasBMP {
		t.Error("No BMP files found in testdata")
	}

	// Index all files
	err = engine.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	stats := engine.GetStats()

	// All files should be processed successfully regardless of type
	if stats.FilesFailed > 0 {
		t.Errorf("Some files failed to index: %d", stats.FilesFailed)
	}

	t.Logf("Successfully indexed JPEG: %v, BMP: %v", hasJPEG, hasBMP)
}

// TestIntegrationThumbnailGeneration verifies thumbnails are stored in database
func TestIntegrationThumbnailGeneration(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata directory not found")
	}

	tmpDB, err := os.CreateTemp("", "thumbnail_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := database.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)
	err = engine.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	// Query thumbnails table
	var thumbnailCount int
	err = db.QueryRow("SELECT COUNT(*) FROM thumbnails").Scan(&thumbnailCount)
	if err != nil {
		t.Fatalf("Failed to query thumbnails: %v", err)
	}

	stats := engine.GetStats()
	expectedThumbnails := stats.FilesProcessed * 4 // 4 sizes

	if thumbnailCount != expectedThumbnails {
		t.Errorf("Expected %d thumbnail rows, got %d", expectedThumbnails, thumbnailCount)
	}

	// Verify each size exists
	sizes := []string{"64", "256", "512", "1024"}
	for _, size := range sizes {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM thumbnails WHERE size = ?", size).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query size %s: %v", size, err)
		}

		if count != stats.FilesProcessed {
			t.Errorf("Size %s: expected %d thumbnails, got %d", size, stats.FilesProcessed, count)
		}
	}

	t.Logf("All %d thumbnails (4 sizes × %d photos) verified in database", thumbnailCount, stats.FilesProcessed)
}

// TestIntegrationColorExtraction verifies colors are extracted and stored
func TestIntegrationColorExtraction(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Testdata directory not found")
	}

	tmpDB, err := os.CreateTemp("", "color_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	db, err := database.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)
	err = engine.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("Indexing failed: %v", err)
	}

	// Query colors table
	var colorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM photo_colors").Scan(&colorCount)
	if err != nil {
		t.Fatalf("Failed to query colors: %v", err)
	}

	stats := engine.GetStats()

	// Verify we have colors stored
	if colorCount == 0 {
		t.Error("No colors extracted")
	}

	// Log the actual count for debugging
	t.Logf("Color count: %d for %d photos (%.1f colors/photo)",
		colorCount, stats.FilesProcessed, float64(colorCount)/float64(stats.FilesProcessed))

	// Verify HSL values are populated
	var hslCount int
	err = db.QueryRow("SELECT COUNT(*) FROM photo_colors WHERE hue IS NOT NULL AND saturation IS NOT NULL AND lightness IS NOT NULL").Scan(&hslCount)
	if err != nil {
		t.Fatalf("Failed to query HSL values: %v", err)
	}

	if hslCount != colorCount {
		t.Errorf("Not all colors have HSL values: %d/%d", hslCount, colorCount)
	}

	t.Logf("All %d colors with HSL values verified", colorCount)
}

// BenchmarkIntegrationIndexing benchmarks the full indexing pipeline
func BenchmarkIntegrationIndexing(b *testing.B) {
	testDataPath := filepath.Join("..", "..", "testdata", "photos")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		b.Skip("Testdata directory not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpDB, err := os.CreateTemp("", "bench_*.db")
		if err != nil {
			b.Fatalf("Failed to create temp database: %v", err)
		}
		tmpDB.Close()
		defer os.Remove(tmpDB.Name())

		db, err := database.Open(tmpDB.Name())
		if err != nil {
			b.Fatalf("Failed to open database: %v", err)
		}

		engine := NewEngine(db, 4)
		err = engine.IndexDirectory(testDataPath)
		if err != nil {
			b.Fatalf("Indexing failed: %v", err)
		}

		db.Close()
	}
}
