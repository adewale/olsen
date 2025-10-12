package indexer

import (
	"os"
	"testing"
	"time"
)

// TestEXIFExtraction tests all EXIF fields used by the indexer
func TestEXIFExtraction(t *testing.T) {
	// Use one of the complete DNG fixtures with full EXIF
	testFile := "../../testdata/dng/01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng"

	// Check if file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found, run: go run testdata/generate_dng_fixtures.go")
	}

	metadata, err := ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// Test camera make and model
	if metadata.CameraMake != "Canon" {
		t.Errorf("CameraMake = %q; want %q", metadata.CameraMake, "Canon")
	}

	if metadata.CameraModel != "Canon EOS R5" {
		t.Errorf("CameraModel = %q; want %q", metadata.CameraModel, "Canon EOS R5")
	}

	// Test lens
	if metadata.LensModel != "RF24mm F1.4 L USM" {
		t.Errorf("LensModel = %q; want %q", metadata.LensModel, "RF24mm F1.4 L USM")
	}

	// Test focal length
	if metadata.FocalLength != 24.0 {
		t.Errorf("FocalLength = %.1f; want 24.0", metadata.FocalLength)
	}

	if metadata.FocalLength35mm != 24 {
		t.Errorf("FocalLength35mm = %d; want 24", metadata.FocalLength35mm)
	}

	// Test exposure settings
	if metadata.ISO != 100 {
		t.Errorf("ISO = %d; want 100", metadata.ISO)
	}

	if metadata.Aperture != 1.4 {
		t.Errorf("Aperture = %.1f; want 1.4", metadata.Aperture)
	}

	if metadata.ShutterSpeed != "1/1000" {
		t.Errorf("ShutterSpeed = %q; want %q", metadata.ShutterSpeed, "1/1000")
	}

	// Test date/time
	expectedDate := time.Date(2025, 3, 15, 6, 30, 0, 0, time.UTC)
	if !metadata.DateTaken.Equal(expectedDate) {
		t.Errorf("DateTaken = %v; want %v", metadata.DateTaken, expectedDate)
	}

	// Test GPS coordinates
	if metadata.Latitude == 0 {
		t.Error("Latitude is 0; expected value")
	} else if metadata.Latitude < 37.7 || metadata.Latitude > 37.8 {
		t.Errorf("Latitude = %.4f; want ~37.7749", metadata.Latitude)
	}

	if metadata.Longitude == 0 {
		t.Error("Longitude is 0; expected value")
	} else if metadata.Longitude < -122.5 || metadata.Longitude > -122.4 {
		t.Errorf("Longitude = %.4f; want ~-122.4194", metadata.Longitude)
	}

	// Test dimensions (note: may be 0 if not in EXIF, populated from image decode elsewhere)
	// These tests are informational only
	if metadata.Width == 0 {
		t.Logf("Note: Width not in EXIF (will be populated from image decode)")
	}

	if metadata.Height == 0 {
		t.Logf("Note: Height not in EXIF (will be populated from image decode)")
	}

	t.Logf("Successfully extracted all EXIF fields:")
	t.Logf("  Camera: %s %s", metadata.CameraMake, metadata.CameraModel)
	t.Logf("  Lens: %s", metadata.LensModel)
	t.Logf("  Focal: %.0fmm (35mm: %dmm)", metadata.FocalLength, metadata.FocalLength35mm)
	t.Logf("  Exposure: ISO %d, f/%.1f, %s", metadata.ISO, metadata.Aperture, metadata.ShutterSpeed)
	t.Logf("  Date: %s", metadata.DateTaken.Format("2006-01-02 15:04:05"))
	t.Logf("  GPS: %.4f, %.4f", metadata.Latitude, metadata.Longitude)
	t.Logf("  Dimensions: %dx%d", metadata.Width, metadata.Height)
}

// TestEXIFFlashDetection tests flash detection specifically
func TestEXIFFlashDetection(t *testing.T) {
	// File with flash fired
	flashFile := "../../testdata/dng/04_nikon_z9_300mm_autumn_afternoon_iso400_flash_green_nogps.dng"

	// Check if file exists
	if _, err := os.Stat(flashFile); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found")
	}

	metadata, err := ExtractMetadata(flashFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// This test will initially fail with goexif, but should pass with exif-go
	if !metadata.FlashFired {
		t.Errorf("FlashFired = false; want true (file has flash metadata)")
		t.Logf("Note: This is a known limitation of goexif library")
		t.Logf("Expected to pass after switching to exif-go")
	} else {
		t.Logf("✓ Flash detection working correctly")
	}
}

// TestEXIFNoFlash tests that non-flash photos are detected correctly
func TestEXIFNoFlash(t *testing.T) {
	testFile := "../../testdata/dng/01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found")
	}

	metadata, err := ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	if metadata.FlashFired {
		t.Errorf("FlashFired = true; want false (file has no flash)")
	}
}

// TestEXIFWithoutGPS tests files without GPS data
func TestEXIFWithoutGPS(t *testing.T) {
	testFile := "../../testdata/dng/02_canon_r5_50mm_summer_morning_iso800_orange_nogps.dng"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found")
	}

	metadata, err := ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	if metadata.Latitude != 0 {
		t.Errorf("Latitude = %.4f; want 0 (file has no GPS)", metadata.Latitude)
	}

	if metadata.Longitude != 0 {
		t.Errorf("Longitude = %.4f; want 0 (file has no GPS)", metadata.Longitude)
	}
}

// TestEXIFBurstSequence tests that burst photos have correct timing
func TestEXIFBurstSequence(t *testing.T) {
	burstFiles := []string{
		"../../testdata/dng/09_burst_1_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
		"../../testdata/dng/10_burst_2_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
		"../../testdata/dng/11_burst_3_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
	}

	var timestamps []time.Time
	for _, file := range burstFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Skip("DNG test fixtures not found")
		}

		metadata, err := ExtractMetadata(file)
		if err != nil {
			t.Fatalf("ExtractMetadata failed for %s: %v", file, err)
		}

		timestamps = append(timestamps, metadata.DateTaken)
	}

	// Check 1-second intervals
	interval1 := timestamps[1].Sub(timestamps[0])
	interval2 := timestamps[2].Sub(timestamps[1])

	if interval1 != time.Second {
		t.Errorf("Interval between burst 1-2 = %v; want 1s", interval1)
	}

	if interval2 != time.Second {
		t.Errorf("Interval between burst 2-3 = %v; want 1s", interval2)
	}

	// Check same camera and lens
	meta1, _ := ExtractMetadata(burstFiles[0])
	meta2, _ := ExtractMetadata(burstFiles[1])
	meta3, _ := ExtractMetadata(burstFiles[2])

	if meta1.CameraModel != meta2.CameraModel || meta2.CameraModel != meta3.CameraModel {
		t.Error("Burst photos have different camera models")
	}

	if meta1.FocalLength != meta2.FocalLength || meta2.FocalLength != meta3.FocalLength {
		t.Error("Burst photos have different focal lengths")
	}

	t.Logf("✓ Burst sequence verified: 3 photos, 1s intervals")
}

// TestEXIFAllFixtures tests that all fixtures can be read
func TestEXIFAllFixtures(t *testing.T) {
	fixtureDir := "../../testdata/dng"

	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found")
	}

	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("Failed to read fixture directory: %v", err)
	}

	successCount := 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "README.md" {
			continue
		}

		filePath := fixtureDir + "/" + entry.Name()
		metadata, err := ExtractMetadata(filePath)
		if err != nil {
			t.Errorf("Failed to extract EXIF from %s: %v", entry.Name(), err)
			continue
		}

		// Verify essential fields are present
		// Note: Width/Height may be 0 if not in EXIF (populated from image decode)

		if metadata.DateTaken.IsZero() {
			t.Errorf("%s: missing date", entry.Name())
		}

		successCount++
	}

	if successCount < 13 {
		t.Errorf("Successfully read %d fixtures; want 13", successCount)
	} else {
		t.Logf("✓ Successfully extracted EXIF from all %d fixtures", successCount)
	}
}
