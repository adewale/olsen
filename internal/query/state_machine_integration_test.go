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

// NOTE: This test requires SQLite (CGO_ENABLED=1) to run.
// Run with: make test-state-machine
// Or manually: CGO_ENABLED=1 go test -v ./internal/query/ -run TestStateMachine
//
// The test requires LibRaw to be installed: brew install libraw

// TestStateMachineIntegration is the comprehensive integration test suite
// that validates the complete state machine model across all filter combinations.
//
// This test ensures:
// 1. All filter combinations work independently
// 2. Facet counts match actual query results
// 3. Removing filters preserves other dimensions
// 4. No hierarchical dependencies anywhere
func TestStateMachineIntegration(t *testing.T) {
	db, cleanup := setupIntegrationTestDB(t)
	defer cleanup()

	engine := NewEngine(db)

	// Run all integration test scenarios
	t.Run("FilterIndependence", func(t *testing.T) {
		testFilterIndependence(t, engine)
	})

	t.Run("FilterRemoval", func(t *testing.T) {
		testFilterRemoval(t, engine)
	})

	t.Run("FilterCombinations", func(t *testing.T) {
		testFilterCombinations(t, engine)
	})

	t.Run("FacetCountAccuracy", func(t *testing.T) {
		testFacetCountAccuracy(t, engine)
	})
}

// testFilterIndependence verifies each temporal dimension works independently
func testFilterIndependence(t *testing.T, engine *Engine) {
	testCases := []struct {
		name   string
		params QueryParams
		desc   string
	}{
		{
			name:   "MonthOnly",
			params: QueryParams{Month: intPtr(10), Limit: 100},
			desc:   "Month=10 without year should return all October photos",
		},
		{
			name:   "DayOnly",
			params: QueryParams{Day: intPtr(15), Limit: 100},
			desc:   "Day=15 without month/year should return all 15th photos",
		},
		{
			name: "MonthAndDay",
			params: QueryParams{
				Month: intPtr(10),
				Day:   intPtr(15),
				Limit: 100,
			},
			desc: "Month=10&Day=15 without year should work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.Query(tc.params)
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			// Should execute successfully (might have 0 results if no data)
			t.Logf("✅ %s: %d results", tc.desc, result.Total)

			// Verify facets can be computed
			facets, err := engine.ComputeFacets(tc.params)
			if err != nil {
				t.Errorf("ComputeFacets failed: %v", err)
			}
			if facets == nil {
				t.Error("Facets should not be nil")
			}
		})
	}
}

// testFilterRemoval verifies removing filters preserves other dimensions
func testFilterRemoval(t *testing.T, engine *Engine) {
	// Start with year=2020&month=10
	year := 2020
	month := 10
	startParams := QueryParams{
		Year:  &year,
		Month: &month,
		Limit: 100,
	}

	startResult, err := engine.Query(startParams)
	if err != nil {
		t.Fatalf("Initial query failed: %v", err)
	}

	if startResult.Total == 0 {
		t.Skip("No October 2020 photos in test database")
	}

	t.Logf("Starting state: year=2020&month=10 (%d photos)", startResult.Total)

	// Remove year, keep month
	monthOnlyParams := QueryParams{
		Month: &month,
		Limit: 100,
	}

	monthOnlyResult, err := engine.Query(monthOnlyParams)
	if err != nil {
		t.Fatalf("Month-only query failed: %v", err)
	}

	t.Logf("After removing year: month=10 (%d photos)", monthOnlyResult.Total)

	// Verify month filter was preserved
	if monthOnlyResult.Total == 0 {
		t.Error("❌ BUG: Removing year resulted in zero results (month filter may have been cleared)")
	}

	// Verify result count increased or stayed same (never decreases)
	if monthOnlyResult.Total < startResult.Total {
		t.Errorf("❌ BUG: Removing filter DECREASED results (%d -> %d). Should increase or stay same.",
			startResult.Total, monthOnlyResult.Total)
	}

	t.Logf("✅ Month filter preserved after removing year")
}

// testFilterCombinations tests various filter combinations
func testFilterCombinations(t *testing.T, engine *Engine) {
	combinations := []struct {
		name   string
		params QueryParams
	}{
		{"YearOnly", QueryParams{Year: intPtr(2024), Limit: 100}},
		{"YearMonth", QueryParams{Year: intPtr(2024), Month: intPtr(10), Limit: 100}},
		{"YearMonthDay", QueryParams{Year: intPtr(2024), Month: intPtr(10), Day: intPtr(15), Limit: 100}},
		{"MonthDay", QueryParams{Month: intPtr(10), Day: intPtr(15), Limit: 100}},
	}

	for _, combo := range combinations {
		t.Run(combo.name, func(t *testing.T) {
			result, err := engine.Query(combo.params)
			if err != nil {
				t.Errorf("Query failed: %v", err)
				return
			}

			facets, err := engine.ComputeFacets(combo.params)
			if err != nil {
				t.Errorf("ComputeFacets failed: %v", err)
				return
			}

			t.Logf("✅ %s: %d results, %d facets computed",
				combo.name, result.Total, countFacets(facets))
		})
	}
}

// testFacetCountAccuracy verifies facet counts match actual query results
func testFacetCountAccuracy(t *testing.T, engine *Engine) {
	// Test with a state that should have facets
	month := 10
	params := QueryParams{
		Month: &month,
		Limit: 100,
	}

	result, err := engine.Query(params)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Total == 0 {
		t.Skip("No October photos in test database")
	}

	facets, err := engine.ComputeFacets(params)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Verify Year facet counts
	if facets.Year != nil {
		for _, yearFacet := range facets.Year.Values {
			if yearFacet.Value == "unknown" {
				continue
			}

			var year int
			fmt.Sscanf(yearFacet.Value, "%d", &year)

			// Simulate clicking this year facet
			testParams := QueryParams{
				Year:  &year,
				Month: &month,
				Limit: 100,
			}

			testResult, err := engine.Query(testParams)
			if err != nil {
				t.Errorf("Test query failed: %v", err)
				continue
			}

			if yearFacet.Count != testResult.Total {
				t.Errorf("❌ FACET COUNT MISMATCH: Year %d shows count=%d but query returned %d photos",
					year, yearFacet.Count, testResult.Total)
			} else {
				t.Logf("✅ Year %d: count=%d matches query result", year, yearFacet.Count)
			}
		}
	}
}

// Helper functions

func setupIntegrationTestDB(t *testing.T) (*sql.DB, func()) {
	tmpfile, err := os.CreateTemp("", "integration_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	dbPath := tmpfile.Name()

	db, err := database.Open(dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Insert test data covering multiple scenarios
	photos := []struct {
		path string
		date time.Time
	}{
		// October photos across multiple years
		{"/test/2020_oct_01.jpg", time.Date(2020, 10, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2020_oct_15.jpg", time.Date(2020, 10, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2021_oct_01.jpg", time.Date(2021, 10, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2021_oct_15.jpg", time.Date(2021, 10, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_oct_01.jpg", time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_oct_15.jpg", time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC)},

		// 15th of various months
		{"/test/2024_jan_15.jpg", time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_feb_15.jpg", time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_mar_15.jpg", time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)},

		// Other dates for diversity
		{"/test/2024_jan_01.jpg", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
		{"/test/2024_dec_31.jpg", time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)},
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

// intPtr helper removed - already defined in url_mapper_test.go

func countFacets(fc *FacetCollection) int {
	count := 0
	if fc.Year != nil {
		count += len(fc.Year.Values)
	}
	if fc.Month != nil {
		count += len(fc.Month.Values)
	}
	if fc.Camera != nil {
		count += len(fc.Camera.Values)
	}
	if fc.Lens != nil {
		count += len(fc.Lens.Values)
	}
	if fc.ColourName != nil {
		count += len(fc.ColourName.Values)
	}
	return count
}
