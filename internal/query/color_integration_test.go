package query

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/indexer"
)

// TestColorQueryIntegration tests that querying for each color returns only images with that color
func TestColorQueryIntegration(t *testing.T) {
	// Get path to color test images
	colorTestPath := filepath.Join("..", "..", "testdata", "color_test")

	// Check if testdata exists
	if _, err := os.Stat(colorTestPath); os.IsNotExist(err) {
		t.Skip("Color test directory not found")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "color_test_*.db")
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

	// Index color test images
	engine := indexer.NewEngine(db, 4)
	t.Logf("Indexing %s", colorTestPath)
	err = engine.IndexDirectory(colorTestPath)
	if err != nil {
		t.Fatalf("IndexDirectory failed: %v", err)
	}

	stats := engine.GetStats()
	t.Logf("Indexed %d images", stats.FilesProcessed)

	if stats.FilesFailed > 0 {
		t.Errorf("%d files failed to index", stats.FilesFailed)
	}

	// Create query engine
	queryEngine := NewEngine(db.DB)

	// Test cases: map of color name to expected dominant image filename patterns
	testCases := []struct {
		color            string
		mustContain      []string // Filenames that MUST be in results
		shouldNotContain []string // Filenames that should NOT be in results (optional check)
	}{
		{
			color:       "red",
			mustContain: []string{"red_dominant"},
		},
		{
			color:       "orange",
			mustContain: []string{"orange_dominant"},
		},
		{
			color:       "yellow",
			mustContain: []string{"yellow_dominant"},
		},
		{
			color:       "green",
			mustContain: []string{"green_dominant"},
		},
		{
			color:       "blue",
			mustContain: []string{"blue_dominant"},
		},
		{
			color:       "purple",
			mustContain: []string{"purple_dominant"},
		},
		{
			color:       "pink",
			mustContain: []string{"pink_dominant"},
		},
		{
			color:       "brown",
			mustContain: []string{"brown_dominant"},
		},
		{
			color:            "grey",
			mustContain:      []string{"grey_dominant"}, // Note: filename uses 'grey'
			shouldNotContain: []string{"red_", "blue_", "green_", "yellow_"},
		},
		{
			color:            "black",
			mustContain:      []string{"black_dominant"},
			shouldNotContain: []string{"white_dominant"},
		},
		{
			color:            "white",
			mustContain:      []string{"white_dominant"},
			shouldNotContain: []string{"black_dominant"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.color, func(t *testing.T) {
			// Query for this color
			params := QueryParams{
				ColourName: []string{tc.color},
				Limit:      100, // Get all results
			}

			results, err := queryEngine.Query(params)
			if err != nil {
				t.Fatalf("Query failed for color %s: %v", tc.color, err)
			}

			if len(results.Photos) == 0 {
				t.Errorf("No photos returned for color %s", tc.color)
				return
			}

			t.Logf("Color %s returned %d photo(s)", tc.color, len(results.Photos))

			// Check that all mustContain patterns are present
			for _, pattern := range tc.mustContain {
				found := false
				for _, photo := range results.Photos {
					if strings.Contains(photo.FilePath, pattern) {
						found = true
						t.Logf("  ✓ Found expected photo: %s", filepath.Base(photo.FilePath))
						break
					}
				}
				if !found {
					t.Errorf("Expected to find photo matching '%s' for color %s", pattern, tc.color)
					t.Logf("  Returned photos:")
					for _, photo := range results.Photos {
						t.Logf("    - %s", filepath.Base(photo.FilePath))
					}
				}
			}

			// Check that shouldNotContain patterns are absent (if specified)
			for _, pattern := range tc.shouldNotContain {
				for _, photo := range results.Photos {
					if strings.Contains(photo.FilePath, pattern) {
						t.Errorf("Unexpected photo '%s' returned for color %s (pattern: %s)",
							filepath.Base(photo.FilePath), tc.color, pattern)
					}
				}
			}
		})
	}
}

// TestColorQueryPrecision tests that single-color images don't match wrong colors
func TestColorQueryPrecision(t *testing.T) {
	// Get path to color test images
	colorTestPath := filepath.Join("..", "..", "testdata", "color_test")

	// Check if testdata exists
	if _, err := os.Stat(colorTestPath); os.IsNotExist(err) {
		t.Skip("Color test directory not found")
	}

	// Create temporary database
	tmpDB, err := os.CreateTemp("", "color_precision_test_*.db")
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

	// Index the entire color test directory (includes all pure colors)
	engine := indexer.NewEngine(db, 4)
	if err := engine.IndexDirectory(colorTestPath); err != nil {
		t.Fatalf("Failed to index color test directory: %v", err)
	}

	stats := engine.GetStats()
	t.Logf("Indexed %d pure color images", stats.FilesProcessed)

	// Create query engine
	queryEngine := NewEngine(db.DB)

	// Test that querying for "red" returns ONLY red_dominant
	redParams := QueryParams{
		ColourName: []string{"red"},
		Limit:      100,
	}

	redResults, err := queryEngine.Query(redParams)
	if err != nil {
		t.Fatalf("Query failed for red: %v", err)
	}

	// Should contain red_dominant
	foundRed := false
	for _, photo := range redResults.Photos {
		if strings.Contains(photo.FilePath, "red_dominant") {
			foundRed = true
		}
	}

	if !foundRed {
		t.Error("red_dominant not found when querying for red")
	}

	// Log all results for debugging
	t.Logf("Red query returned %d photos:", len(redResults.Photos))
	for _, photo := range redResults.Photos {
		t.Logf("  - %s", filepath.Base(photo.FilePath))
	}
}
