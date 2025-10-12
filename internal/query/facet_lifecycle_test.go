package query

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestFacetLifecycle tests the complete lifecycle of adding and removing facets
func TestFacetLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	engine := NewEngine(db)

	// Test 1: Start with no filters - all photos
	t.Run("NoFilters", func(t *testing.T) {
		params := QueryParams{Limit: 100}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		totalPhotos := result.Total
		if totalPhotos == 0 {
			t.Skip("No photos in test database")
		}

		t.Logf("Total photos: %d", totalPhotos)

		// Compute facets
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		// Verify facet counts sum correctly
		verifyFacetCounts(t, facets, totalPhotos)
	})

	// Test 2: Add color filter - narrows results
	t.Run("AddColorFilter", func(t *testing.T) {
		// First, get all photos to find a color to filter by
		params := QueryParams{Limit: 100}
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.ColourName == nil || len(facets.ColourName.Values) == 0 {
			t.Skip("No color facets available")
		}

		// Pick the first color
		testColor := facets.ColourName.Values[0].Value
		testColorCount := facets.ColourName.Values[0].Count

		// Apply color filter
		params.ColourName = []string{testColor}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query with color filter failed: %v", err)
		}

		// Result count should match facet count
		if result.Total != testColorCount {
			t.Errorf("Color filter result mismatch: got %d photos, facet said %d",
				result.Total, testColorCount)
		}

		t.Logf("Color filter '%s': %d photos (expected %d)", testColor, result.Total, testColorCount)

		// Compute facets with color filter active
		facetsWithColor, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets with color failed: %v", err)
		}

		// Verify color facet shows as selected
		if facetsWithColor.ColourName != nil {
			found := false
			for _, v := range facetsWithColor.ColourName.Values {
				if v.Value == testColor && v.Selected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Color '%s' not marked as selected in facets", testColor)
			}
		}
	})

	// Test 3: Add year filter on top of color - further narrows
	t.Run("AddYearFilter", func(t *testing.T) {
		// Start with color filter
		params := QueryParams{
			ColourName: []string{"blue"},
			Limit:      100,
		}

		resultColor, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query with color failed: %v", err)
		}

		if resultColor.Total == 0 {
			t.Skip("No blue photos in test database")
		}

		colorCount := resultColor.Total

		// Get facets with color filter
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Year == nil || len(facets.Year.Values) == 0 {
			t.Skip("No year facets available with color filter")
		}

		// Pick the first year
		testYear := 0
		testYearCount := 0
		for _, yv := range facets.Year.Values {
			if yv.Count > 0 {
				// Parse year
				var y int
				if _, err := fmt.Sscanf(yv.Value, "%d", &y); err == nil {
					testYear = y
					testYearCount = yv.Count
					break
				}
			}
		}

		if testYear == 0 {
			t.Skip("No valid year facets")
		}

		// Add year filter
		params.Year = &testYear
		resultBoth, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query with color+year failed: %v", err)
		}

		// Result count should match year facet count
		if resultBoth.Total != testYearCount {
			t.Errorf("Color+Year filter result mismatch: got %d photos, facet said %d",
				resultBoth.Total, testYearCount)
		}

		// Should be less than or equal to color-only count
		if resultBoth.Total > colorCount {
			t.Errorf("Adding year filter increased count: was %d, now %d",
				colorCount, resultBoth.Total)
		}

		t.Logf("Color+Year filter: %d photos (color only: %d)", resultBoth.Total, colorCount)
	})

	// Test 4: Add camera filter - even narrower
	t.Run("AddCameraFilter", func(t *testing.T) {
		year2024 := 2024
		params := QueryParams{
			ColourName: []string{"blue"},
			Year:       &year2024,
			Limit:      100,
		}

		resultBefore, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if resultBefore.Total == 0 {
			t.Skip("No blue 2024 photos")
		}

		beforeCount := resultBefore.Total

		// Get facets
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Camera == nil || len(facets.Camera.Values) == 0 {
			t.Skip("No camera facets available")
		}

		// Pick first camera
		testCamera := facets.Camera.Values[0].Value
		testCameraCount := facets.Camera.Values[0].Count

		// Parse camera into make/model
		parts := strings.SplitN(testCamera, " ", 2)
		if len(parts) != 2 {
			t.Skip("Invalid camera format")
		}

		params.CameraMake = []string{parts[0]}
		params.CameraModel = []string{parts[1]}

		resultAfter, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query with camera failed: %v", err)
		}

		// Should match facet count
		if resultAfter.Total != testCameraCount {
			t.Errorf("Camera filter result mismatch: got %d, facet said %d",
				resultAfter.Total, testCameraCount)
		}

		// Should not exceed previous count
		if resultAfter.Total > beforeCount {
			t.Errorf("Adding camera filter increased count: was %d, now %d",
				beforeCount, resultAfter.Total)
		}

		t.Logf("Color+Year+Camera: %d photos (before camera: %d)", resultAfter.Total, beforeCount)
	})

	// Test 5: Remove middle filter (year) - should increase count
	t.Run("RemoveMiddleFilter", func(t *testing.T) {
		year2024 := 2024
		params := QueryParams{
			ColourName:  []string{"blue"},
			Year:        &year2024,
			CameraMake:  []string{"Canon"},
			CameraModel: []string{"EOS R5"},
			Limit:       100,
		}

		resultWithYear, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if resultWithYear.Total == 0 {
			t.Skip("No photos matching all filters")
		}

		withYearCount := resultWithYear.Total

		// Remove year filter
		params.Year = nil

		resultWithoutYear, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query without year failed: %v", err)
		}

		// Count should increase or stay same (not decrease)
		if resultWithoutYear.Total < withYearCount {
			t.Errorf("Removing year filter decreased count: was %d, now %d",
				withYearCount, resultWithoutYear.Total)
		}

		t.Logf("Removed year filter: %d → %d photos", withYearCount, resultWithoutYear.Total)
	})

	// Test 6: Remove all filters - back to full count
	t.Run("RemoveAllFilters", func(t *testing.T) {
		// Start with multiple filters
		year2024 := 2024
		params := QueryParams{
			ColourName:  []string{"blue"},
			Year:        &year2024,
			CameraMake:  []string{"Canon"},
			CameraModel: []string{"EOS R5"},
			Limit:       100,
		}

		resultFiltered, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		filteredCount := resultFiltered.Total

		// Remove all filters
		paramsEmpty := QueryParams{Limit: 100}
		resultAll, err := engine.Query(paramsEmpty)
		if err != nil {
			t.Fatalf("Query without filters failed: %v", err)
		}

		// Should have more photos without filters
		if resultAll.Total < filteredCount {
			t.Errorf("Removing all filters decreased count: was %d, now %d",
				filteredCount, resultAll.Total)
		}

		t.Logf("Removed all filters: %d → %d photos", filteredCount, resultAll.Total)
	})
}

// TestFacetURLPreservation tests that facet URLs preserve other filters
func TestFacetURLPreservation(t *testing.T) {
	builder := NewFacetURLBuilder(NewURLMapper())

	// Start with multiple filters active
	year2024 := 2024
	baseParams := QueryParams{
		ColourName:  []string{"blue"},
		Year:        &year2024,
		CameraMake:  []string{"Canon"},
		CameraModel: []string{"EOS R5"},
	}

	// Create a facet collection with a TimeOfDay facet
	facets := &FacetCollection{
		TimeOfDay: &Facet{
			Name:  "time_of_day",
			Label: "Time of Day",
			Values: []FacetValue{
				{Value: "morning", Label: "Morning", Count: 10, Selected: false},
				{Value: "afternoon", Label: "Afternoon", Count: 15, Selected: false},
			},
		},
	}

	// Build URLs
	builder.BuildURLsForFacets(facets, baseParams)

	// Check that URLs preserve existing filters
	for _, v := range facets.TimeOfDay.Values {
		if v.URL == "" {
			t.Errorf("Facet value '%s' has no URL", v.Value)
			continue
		}

		// Split URL into path and query
		parts := strings.SplitN(v.URL, "?", 2)
		path := parts[0]
		query := ""
		if len(parts) == 2 {
			query = parts[1]
		}

		// Parse the URL back to params
		mapper := NewURLMapper()
		parsedParams, err := mapper.ParsePath(path, query)
		if err != nil {
			t.Errorf("Failed to parse URL '%s': %v", v.URL, err)
			continue
		}

		// Verify original filters are preserved
		if len(parsedParams.ColourName) == 0 || parsedParams.ColourName[0] != "blue" {
			t.Errorf("Color filter not preserved in URL: %s", v.URL)
		}
		if parsedParams.Year == nil || *parsedParams.Year != 2024 {
			t.Errorf("Year filter not preserved in URL: %s", v.URL)
		}
		if len(parsedParams.CameraMake) == 0 || parsedParams.CameraMake[0] != "Canon" {
			t.Errorf("Camera make not preserved in URL: %s", v.URL)
		}
		if len(parsedParams.CameraModel) == 0 || parsedParams.CameraModel[0] != "EOS R5" {
			t.Errorf("Camera model not preserved in URL: %s", v.URL)
		}

		// Verify new filter is added
		found := false
		for _, tod := range parsedParams.TimeOfDay {
			if tod == v.Value {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("New TimeOfDay filter '%s' not added in URL: %s", v.Value, v.URL)
		}

		t.Logf("✓ Facet '%s' URL preserves all filters: %s", v.Value, v.URL)
	}
}

// TestFacetRemovalURL tests that selected facets generate removal URLs
func TestFacetRemovalURL(t *testing.T) {
	builder := NewFacetURLBuilder(NewURLMapper())

	// Start with a color filter active
	baseParams := QueryParams{
		ColourName: []string{"blue"},
	}

	// Create facet with blue selected
	facets := &FacetCollection{
		ColourName: &Facet{
			Name:  "color",
			Label: "Color",
			Values: []FacetValue{
				{Value: "blue", Label: "Blue", Count: 100, Selected: true},
				{Value: "red", Label: "Red", Count: 50, Selected: false},
			},
		},
	}

	builder.BuildURLsForFacets(facets, baseParams)

	// Check blue (selected) generates removal URL
	blueValue := facets.ColourName.Values[0]
	if blueValue.Value != "blue" {
		t.Fatal("Test setup error: expected blue to be first")
	}

	// Split URL into path and query
	parts := strings.SplitN(blueValue.URL, "?", 2)
	path := parts[0]
	query := ""
	if len(parts) == 2 {
		query = parts[1]
	}

	mapper := NewURLMapper()
	parsedParams, err := mapper.ParsePath(path, query)
	if err != nil {
		t.Fatalf("Failed to parse blue URL: %v", err)
	}

	// Should have no color filter (removed)
	if len(parsedParams.ColourName) > 0 {
		t.Errorf("Selected blue facet URL should remove color filter, but has: %v", parsedParams.ColourName)
	}

	// Check red (not selected) generates addition URL
	redValue := facets.ColourName.Values[1]

	parts = strings.SplitN(redValue.URL, "?", 2)
	path = parts[0]
	query = ""
	if len(parts) == 2 {
		query = parts[1]
	}

	parsedParams, err = mapper.ParsePath(path, query)
	if err != nil {
		t.Fatalf("Failed to parse red URL: %v", err)
	}

	// Should have red color filter
	if len(parsedParams.ColourName) == 0 || parsedParams.ColourName[0] != "red" {
		t.Errorf("Unselected red facet URL should add red filter, but has: %v", parsedParams.ColourName)
	}

	t.Logf("✓ Selected facet generates removal URL")
	t.Logf("✓ Unselected facet generates addition URL")
}

// verifyFacetCounts checks that facet counts are internally consistent
func verifyFacetCounts(t *testing.T, facets *FacetCollection, totalPhotos int) {
	t.Helper()

	// For non-overlapping facets (like year), counts should sum to total
	if facets.Year != nil && len(facets.Year.Values) > 0 {
		yearSum := 0
		for _, v := range facets.Year.Values {
			yearSum += v.Count
		}
		// Year sum should equal or be less than total (some photos may have no date)
		if yearSum > totalPhotos {
			t.Errorf("Year facet counts sum (%d) exceeds total photos (%d)", yearSum, totalPhotos)
		}
		t.Logf("Year facets: %d photos across %d years", yearSum, len(facets.Year.Values))
	}

	// For overlapping facets (like color), individual counts should not exceed total
	if facets.ColourName != nil {
		for _, v := range facets.ColourName.Values {
			if v.Count > totalPhotos {
				t.Errorf("Color '%s' count (%d) exceeds total photos (%d)",
					v.Value, v.Count, totalPhotos)
			}
		}
	}

	// Camera facets should not exceed total
	if facets.Camera != nil {
		cameraSum := 0
		for _, v := range facets.Camera.Values {
			if v.Count > totalPhotos {
				t.Errorf("Camera '%s' count (%d) exceeds total photos (%d)",
					v.Value, v.Count, totalPhotos)
			}
			cameraSum += v.Count
		}
		if cameraSum > totalPhotos {
			t.Errorf("Camera facet counts sum (%d) exceeds total photos (%d)",
				cameraSum, totalPhotos)
		}
		t.Logf("Camera facets: %d photos across %d cameras", cameraSum, len(facets.Camera.Values))
	}
}

// setupTestDB creates a test database for facet testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Try to use existing test database
	testDBPath := filepath.Join("..", "..", "test.db")
	if _, err := os.Stat(testDBPath); err == nil {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		return db
	}

	// Fallback: create in-memory database (will be empty)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Log("Warning: Using empty in-memory database, tests may be skipped")
	return db
}
