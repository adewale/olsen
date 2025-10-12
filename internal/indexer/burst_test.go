package indexer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
)

func TestBurstDetection(t *testing.T) {
	// Use the DNG test fixtures which include a burst sequence (images 9-11)
	testDataPath := filepath.Join("..", "..", "testdata", "dng")

	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("DNG test fixtures not found, run: go run testdata/generate_dng_fixtures.go")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "burst_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	// Index the test data
	db, err := database.Open(tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	engine := NewEngine(db, 2)
	err = engine.IndexDirectory(testDataPath)
	if err != nil {
		t.Fatalf("Failed to index test data: %v", err)
	}

	// Detect bursts
	detector := NewBurstDetector(db)
	bursts, err := detector.DetectBursts()
	if err != nil {
		t.Fatalf("Failed to detect bursts: %v", err)
	}

	// Should find at least 1 burst (images 9-11)
	if len(bursts) == 0 {
		t.Error("Expected to find at least 1 burst group")
	}

	// Verify burst properties
	for i, burst := range bursts {
		t.Logf("Burst %d: %d photos (IDs: %v)", i+1, len(burst), burst)

		if len(burst) < 3 {
			t.Errorf("Burst %d has only %d photos, expected at least 3", i+1, len(burst))
		}
	}

	// Save bursts to database
	err = detector.SaveBursts(bursts)
	if err != nil {
		t.Fatalf("Failed to save bursts: %v", err)
	}

	// Verify saved bursts
	burstCount, photoCount, err := detector.GetBurstStats()
	if err != nil {
		t.Fatalf("Failed to get burst stats: %v", err)
	}

	t.Logf("Saved %d burst groups containing %d photos", burstCount, photoCount)

	if burstCount != len(bursts) {
		t.Errorf("Expected %d burst groups, got %d", len(bursts), burstCount)
	}
}

func TestBurstDetectorSettings(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	detector := NewBurstDetector(db)

	// Verify default settings
	if detector.maxTimeDelta.Seconds() != 2.0 {
		t.Errorf("Expected maxTimeDelta = 2s, got %v", detector.maxTimeDelta)
	}

	if detector.maxFocalDelta != 5.0 {
		t.Errorf("Expected maxFocalDelta = 5.0mm, got %.1f", detector.maxFocalDelta)
	}

	if detector.minBurstSize != 3 {
		t.Errorf("Expected minBurstSize = 3, got %d", detector.minBurstSize)
	}
}

func TestCanExtendBurst(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	detector := NewBurstDetector(db)

	// Create test photos
	baseTime := mustParseTime("2025-05-15 12:00:00")

	tests := []struct {
		name      string
		last      Photo
		candidate Photo
		want      bool
	}{
		{
			name: "Valid burst continuation",
			last: Photo{
				DateTaken:   baseTime,
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			candidate: Photo{
				DateTaken:   baseTime.Add(1 * Second),
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			want: true,
		},
		{
			name: "Time gap too large",
			last: Photo{
				DateTaken:   baseTime,
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			candidate: Photo{
				DateTaken:   baseTime.Add(5 * Second),
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			want: false,
		},
		{
			name: "Different camera",
			last: Photo{
				DateTaken:   baseTime,
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			candidate: Photo{
				DateTaken:   baseTime.Add(1 * Second),
				CameraMake:  "Nikon",
				CameraModel: "Z9",
				FocalLength: 50.0,
			},
			want: false,
		},
		{
			name: "Focal length too different",
			last: Photo{
				DateTaken:   baseTime,
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			candidate: Photo{
				DateTaken:   baseTime.Add(1 * Second),
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 60.0, // 10mm difference
			},
			want: false,
		},
		{
			name: "Focal length within tolerance",
			last: Photo{
				DateTaken:   baseTime,
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 50.0,
			},
			candidate: Photo{
				DateTaken:   baseTime.Add(1 * Second),
				CameraMake:  "Canon",
				CameraModel: "EOS R5",
				FocalLength: 53.0, // 3mm difference (< 5mm)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			photos := []Photo{tt.last, tt.candidate}
			burst := []int{0}
			got := detector.canExtendBurst(photos, burst, 1)

			if got != tt.want {
				t.Errorf("canExtendBurst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindBurstSequences(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	detector := NewBurstDetector(db)

	baseTime := mustParseTime("2025-05-15 12:00:00")

	tests := []struct {
		name      string
		photos    []Photo
		wantCount int
	}{
		{
			name: "Simple burst sequence",
			photos: []Photo{
				{ID: 1, DateTaken: baseTime, CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 2, DateTaken: baseTime.Add(1 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 3, DateTaken: baseTime.Add(2 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
			},
			wantCount: 1,
		},
		{
			name: "No burst (too few photos)",
			photos: []Photo{
				{ID: 1, DateTaken: baseTime, CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 2, DateTaken: baseTime.Add(1 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
			},
			wantCount: 0,
		},
		{
			name: "No burst (time gaps)",
			photos: []Photo{
				{ID: 1, DateTaken: baseTime, CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 2, DateTaken: baseTime.Add(5 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 3, DateTaken: baseTime.Add(10 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
			},
			wantCount: 0,
		},
		{
			name: "Multiple bursts",
			photos: []Photo{
				// First burst
				{ID: 1, DateTaken: baseTime, CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 2, DateTaken: baseTime.Add(1 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				{ID: 3, DateTaken: baseTime.Add(2 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 50},
				// Gap
				{ID: 4, DateTaken: baseTime.Add(10 * Second), CameraMake: "Canon", CameraModel: "R5", FocalLength: 85},
				// Second burst
				{ID: 5, DateTaken: baseTime.Add(20 * Second), CameraMake: "Nikon", CameraModel: "Z9", FocalLength: 85},
				{ID: 6, DateTaken: baseTime.Add(21 * Second), CameraMake: "Nikon", CameraModel: "Z9", FocalLength: 85},
				{ID: 7, DateTaken: baseTime.Add(22 * Second), CameraMake: "Nikon", CameraModel: "Z9", FocalLength: 85},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bursts := detector.findBurstSequences(tt.photos)

			if len(bursts) != tt.wantCount {
				t.Errorf("findBurstSequences() found %d bursts, want %d", len(bursts), tt.wantCount)
			}

			for i, burst := range bursts {
				t.Logf("  Burst %d: %d photos (IDs: %v)", i+1, len(burst), burst)
			}
		})
	}
}

// Helper functions

var Second = 1 * time.Second

func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}
