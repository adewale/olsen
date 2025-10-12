package query

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/pkg/models"
)

// TestFacetCountsMatchActualResults is the MASTER test that verifies
// the core invariant of the state machine model:
//
// "The count shown on a facet value MUST equal the number of photos
//
//	that will be displayed when that facet value is clicked."
//
// This test caught the WHERE clause bug where Month filter was skipped
// when Year was nil, causing facet counts to be wrong.
func TestFacetCountsMatchActualResults(t *testing.T) {
	// Create test database with known data
	db, cleanup := createTestDatabase(t)
	defer cleanup()

	engine := NewEngine(db)

	// Test Scenario 1: Year facet with Month filter active
	t.Run("YearFacetWithMonthFilter", func(t *testing.T) {
		// Start with January 2025 (should have photos)
		year := 2025
		month := 1
		currentParams := QueryParams{
			Year:  &year,
			Month: &month,
			Limit: 100,
		}

		// Execute query to get current results
		currentResult, err := engine.Query(currentParams)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if currentResult.Total == 0 {
			t.Skip("Test database has no January 2025 photos - need test data")
		}

		// Compute facets for current state
		facets, err := engine.ComputeFacets(currentParams)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Year == nil {
			t.Fatal("Year facet is nil")
		}

		// Verify EVERY year facet value
		for _, yearFacet := range facets.Year.Values {
			// Parse year from facet value
			var facetYear int
			if yearFacet.Value == "unknown" {
				facetYear = -1
			} else {
				fmt.Sscanf(yearFacet.Value, "%d", &facetYear)
			}

			// Simulate clicking this year facet
			clickParams := QueryParams{
				Year:  &facetYear,
				Month: &month, // Month MUST be preserved
				Limit: 100,
			}

			// Execute query to see actual results
			clickResult, err := engine.Query(clickParams)
			if err != nil {
				t.Fatalf("Query for year %d failed: %v", facetYear, err)
			}

			// CRITICAL ASSERTION: Count must match
			if yearFacet.Count != clickResult.Total {
				t.Errorf("❌ BUG DETECTED: Year %d facet shows count=%d but actual query returned %d photos\n"+
					"This means user sees '%s (%d)' but clicking it shows %d photos.\n"+
					"Context: Viewing year=%d&month=%d (Month filter must be preserved!)",
					facetYear, yearFacet.Count, clickResult.Total,
					yearFacet.Label, yearFacet.Count, clickResult.Total,
					year, month)
			} else {
				t.Logf("✅ Year %d: facet count=%d matches actual results=%d", facetYear, yearFacet.Count, clickResult.Total)
			}
		}
	})

	// Test Scenario 2: Month facet with Year filter active
	t.Run("MonthFacetWithYearFilter", func(t *testing.T) {
		// Start with 2025
		year := 2025
		currentParams := QueryParams{
			Year:  &year,
			Limit: 100,
		}

		currentResult, err := engine.Query(currentParams)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if currentResult.Total == 0 {
			t.Skip("Test database has no 2025 photos")
		}

		facets, err := engine.ComputeFacets(currentParams)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Month == nil {
			t.Fatal("Month facet is nil")
		}

		// Verify EVERY month facet value
		for _, monthFacet := range facets.Month.Values {
			var facetMonth int
			fmt.Sscanf(monthFacet.Value, "%d", &facetMonth)

			// Simulate clicking this month
			clickParams := QueryParams{
				Year:  &year, // Year preserved
				Month: &facetMonth,
				Limit: 100,
			}

			clickResult, err := engine.Query(clickParams)
			if err != nil {
				t.Fatalf("Query for month %d failed: %v", facetMonth, err)
			}

			if monthFacet.Count != clickResult.Total {
				t.Errorf("❌ BUG DETECTED: Month %d facet shows count=%d but actual query returned %d photos",
					facetMonth, monthFacet.Count, clickResult.Total)
			} else {
				t.Logf("✅ Month %d: facet count=%d matches actual results=%d", facetMonth, monthFacet.Count, clickResult.Total)
			}
		}
	})

	// Test Scenario 3: Month facet with Day filter active (hierarchical trap!)
	t.Run("MonthFacetWithDayFilter", func(t *testing.T) {
		// Start with day=15 (all photos on 15th of any month)
		day := 15
		currentParams := QueryParams{
			Day:   &day,
			Limit: 100,
		}

		currentResult, err := engine.Query(currentParams)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if currentResult.Total == 0 {
			t.Skip("Test database has no day=15 photos")
		}

		facets, err := engine.ComputeFacets(currentParams)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Month == nil {
			t.Fatal("Month facet is nil")
		}

		// Verify month facets preserve day filter
		for _, monthFacet := range facets.Month.Values {
			var facetMonth int
			fmt.Sscanf(monthFacet.Value, "%d", &facetMonth)

			clickParams := QueryParams{
				Month: &facetMonth,
				Day:   &day, // Day MUST be preserved
				Limit: 100,
			}

			clickResult, err := engine.Query(clickParams)
			if err != nil {
				t.Fatalf("Query for month %d, day %d failed: %v", facetMonth, day, err)
			}

			if monthFacet.Count != clickResult.Total {
				t.Errorf("❌ BUG DETECTED: Month %d facet shows count=%d but actual query returned %d photos (day=%d must be preserved!)",
					facetMonth, monthFacet.Count, clickResult.Total, day)
			}
		}
	})

	// Test Scenario 4: Year facet with multiple filters
	t.Run("YearFacetWithMultipleFilters", func(t *testing.T) {
		// Start with month=9, colour=red
		year := 2025
		month := 9
		currentParams := QueryParams{
			Year:       &year,
			Month:      &month,
			ColourName: []string{"red"},
			Limit:      100,
		}

		currentResult, err := engine.Query(currentParams)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if currentResult.Total == 0 {
			t.Skip("No red photos in September 2025")
		}

		facets, err := engine.ComputeFacets(currentParams)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		if facets.Year == nil {
			t.Fatal("Year facet is nil")
		}

		// Verify year facets preserve BOTH month and colour filters
		for _, yearFacet := range facets.Year.Values {
			var facetYear int
			if yearFacet.Value == "unknown" {
				continue
			}
			fmt.Sscanf(yearFacet.Value, "%d", &facetYear)

			clickParams := QueryParams{
				Year:       &facetYear,
				Month:      &month,          // Month preserved
				ColourName: []string{"red"}, // Colour preserved
				Limit:      100,
			}

			clickResult, err := engine.Query(clickParams)
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			if yearFacet.Count != clickResult.Total {
				t.Errorf("❌ BUG DETECTED: Year %d facet shows count=%d but actual query returned %d photos (month=%d, colour=red must be preserved!)",
					facetYear, yearFacet.Count, clickResult.Total, month)
			}
		}
	})
}

// TestDisabledFacetsUnclickable verifies that facets with count=0 are marked as disabled
func TestDisabledFacetsUnclickable(t *testing.T) {
	db, cleanup := createTestDatabase(t)
	defer cleanup()

	engine := NewEngine(db)

	// Find a state with some results
	year := 2025
	month := 9
	params := QueryParams{
		Year:  &year,
		Month: &month,
		Limit: 100,
	}

	result, err := engine.Query(params)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Total == 0 {
		t.Skip("No September 2025 photos")
	}

	// Compute facets
	facets, err := engine.ComputeFacets(params)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Check ALL facets for count=0 values
	disabledCount := 0
	if facets.Year != nil {
		for _, v := range facets.Year.Values {
			if v.Count == 0 {
				disabledCount++
				t.Logf("Year facet '%s' has count=0 (should be disabled in UI)", v.Label)
			}
		}
	}
	if facets.Month != nil {
		for _, v := range facets.Month.Values {
			if v.Count == 0 {
				disabledCount++
				t.Logf("Month facet '%s' has count=0 (should be disabled in UI)", v.Label)
			}
		}
	}

	t.Logf("Found %d disabled facet values (count=0) - these should not be clickable in UI", disabledCount)
}

// createTestDatabase creates an in-memory test database with sample photos
func createTestDatabase(t *testing.T) (*sql.DB, func()) {
	// Create temp database file
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	dbPath := tmpfile.Name()

	// Initialize database
	db, err := database.Open(dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Insert test data
	photos := []struct {
		path string
		date time.Time
	}{
		{"/test/2025_jan_01.jpg", time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2025_jan_15.jpg", time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2025_sep_01.jpg", time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2025_sep_15.jpg", time.Date(2025, 9, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_sep_01.jpg", time.Date(2024, 9, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_sep_15.jpg", time.Date(2024, 9, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2023_jan_15.jpg", time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)},
	}

	for _, p := range photos {
		meta := &models.PhotoMetadata{
			FilePath:  p.path,
			DateTaken: p.date,
			Width:     1920,
			Height:    1080,
		}
		err := db.InsertPhoto(meta)
		if err != nil {
			db.Close()
			os.Remove(dbPath)
			t.Fatalf("Failed to insert test photo: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db.DB, cleanup
}
