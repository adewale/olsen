package indexer

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/adewale/olsen/internal/database"
)

func TestCalculateFileHash(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Calculate hash
	hash, err := calculateFileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("calculateFileHash failed: %v", err)
	}

	// Hash should be 64 characters (SHA-256 hex string)
	if len(hash) != 64 {
		t.Errorf("Hash length = %d; want 64", len(hash))
	}

	// Verify consistency
	hash2, err := calculateFileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("Second hash calculation failed: %v", err)
	}

	if hash != hash2 {
		t.Errorf("Hash inconsistent: %s vs %s", hash, hash2)
	}
}

func TestFindDNGFiles(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test_images_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files
	testFiles := []string{
		"photo1.dng",
		"photo2.DNG",
		"subdir/photo3.dng",
		"subdir/nested/photo4.dng",
		"image.jpg", // Should be included (JPEG support)
		"scan.bmp",  // Should be included (BMP support)
		"doc.txt",   // Should be ignored
		"file.pdf",  // Should be ignored
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create engine and find image files
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 1)
	files, err := engine.findDNGFiles(tmpDir)
	if err != nil {
		t.Fatalf("findDNGFiles failed: %v", err)
	}

	// Should find 6 image files (4 DNG + 1 JPEG + 1 BMP)
	expectedCount := 6
	if len(files) != expectedCount {
		t.Errorf("Found %d image files; want %d", len(files), expectedCount)
	}

	// Verify all found files have supported extensions
	supportedExts := map[string]bool{
		".dng":  true,
		".DNG":  true,
		".jpg":  true,
		".jpeg": true,
		".bmp":  true,
	}

	for _, file := range files {
		ext := filepath.Ext(file)
		if !supportedExts[ext] {
			t.Errorf("Found unsupported file: %s", file)
		}
	}
}

func TestNewEngine(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name          string
		workerCount   int
		expectedCount int
	}{
		{"Positive workers", 8, 8},
		{"Zero workers", 0, 4},      // Should default to 4
		{"Negative workers", -5, 4}, // Should default to 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(db, tt.workerCount)
			if engine.workerCount != tt.expectedCount {
				t.Errorf("Worker count = %d; want %d", engine.workerCount, tt.expectedCount)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)

	// Modify stats
	engine.mu.Lock()
	engine.stats.FilesFound = 100
	engine.stats.FilesProcessed = 75
	engine.stats.FilesFailed = 5
	engine.mu.Unlock()

	// Get stats
	stats := engine.GetStats()

	if stats.FilesFound != 100 {
		t.Errorf("FilesFound = %d; want 100", stats.FilesFound)
	}
	if stats.FilesProcessed != 75 {
		t.Errorf("FilesProcessed = %d; want 75", stats.FilesProcessed)
	}
	if stats.FilesFailed != 5 {
		t.Errorf("FilesFailed = %d; want 5", stats.FilesFailed)
	}
}

func TestIndexDirectoryEmpty(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "test_empty_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)
	err = engine.IndexDirectory(tmpDir)
	if err != nil {
		t.Fatalf("IndexDirectory failed: %v", err)
	}

	stats := engine.GetStats()
	if stats.FilesFound != 0 {
		t.Errorf("FilesFound = %d; want 0", stats.FilesFound)
	}
}

// Helper function to create a minimal JPEG with EXIF data
func createTestJPEGWithEXIF(t *testing.T, path string) {
	// Create a simple image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 2), uint8(y * 2), 128, 255})
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}

	// For a real test with EXIF, we'd need to add EXIF data
	// For now, just write the JPEG
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write JPEG: %v", err)
	}
}

// Integration test that would require actual DNG files
// This is marked as skipped unless DNG test files are available
func TestIndexDirectoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if testdata directory with DNG files exists
	testDataDir := filepath.Join("..", "..", "testdata", "dng")
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Testdata directory with DNG files not found")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test_*.db")
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

	// Run indexing
	engine := NewEngine(db, 2)
	err = engine.IndexDirectory(testDataDir)
	if err != nil {
		t.Fatalf("IndexDirectory failed: %v", err)
	}

	// Verify stats
	stats := engine.GetStats()
	if stats.FilesFound == 0 {
		t.Error("No files found in testdata")
	}

	if stats.FilesProcessed == 0 {
		t.Error("No files processed")
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
		t.Errorf("Database photo count (%d) != files processed (%d)", count, stats.FilesProcessed)
	}
}

// Benchmark tests
func BenchmarkCalculateFileHash(b *testing.B) {
	// Create a test file
	tmpFile, err := os.CreateTemp("", "bench_*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write 1MB of data
	data := make([]byte, 1024*1024)
	if _, err := tmpFile.Write(data); err != nil {
		b.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := calculateFileHash(tmpFile.Name())
		if err != nil {
			b.Fatalf("calculateFileHash failed: %v", err)
		}
	}
}

func BenchmarkGenerateThumbnails(b *testing.B) {
	// Create a test image
	img := createSolidColorImage(2000, 1500, color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateThumbnailsFromImage(img)
		if err != nil {
			b.Fatalf("GenerateThumbnailsFromImage failed: %v", err)
		}
	}
}

func BenchmarkExtractColourPalette(b *testing.B) {
	// Create a test image with varied colors
	img := createTestImage(256, 256, []color.RGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{255, 255, 0, 255},
		{255, 0, 255, 255},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExtractColourPalette(img, 5)
		if err != nil {
			b.Fatalf("ExtractColourPalette failed: %v", err)
		}
	}
}

func BenchmarkComputePerceptualHash(b *testing.B) {
	img := createSolidColorImage(256, 256, color.RGBA{128, 128, 128, 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ComputePerceptualHash(img)
		if err != nil {
			b.Fatalf("ComputePerceptualHash failed: %v", err)
		}
	}
}
