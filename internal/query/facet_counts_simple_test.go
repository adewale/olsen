package query

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestFacetCountsCorrect_YearPreservesMonth verifies the fix where Year facet
// computation NOW CORRECTLY preserves Month/Day filters.
//
// Expected behavior: When viewing November 2024, clicking on "2023" in the Year facet
// should navigate to November 2023, and the count shown should be for November 2023 only.
//
// This test verifies that the facet count matches the actual query result count.

func TestFacetCountsCorrect_YearPreservesMonth(t *testing.T) {
	// Create simple test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE photos (
			id INTEGER PRIMARY KEY,
			file_path TEXT NOT NULL,
			date_taken DATETIME NOT NULL,
			camera_make TEXT,
			camera_model TEXT,
			lens_model TEXT,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Insert test data:
	// November 2023: 3 photos
	// December 2023: 4 photos
	// November 2024: 5 photos
	// December 2024: 6 photos
	testData := []struct {
		id   int
		date string
	}{
		// November 2023 (3 photos)
		{1, "2023-11-15 12:00:00"},
		{2, "2023-11-16 12:00:00"},
		{3, "2023-11-17 12:00:00"},

		// December 2023 (4 photos)
		{4, "2023-12-15 12:00:00"},
		{5, "2023-12-16 12:00:00"},
		{6, "2023-12-17 12:00:00"},
		{7, "2023-12-18 12:00:00"},

		// November 2024 (5 photos)
		{8, "2024-11-15 12:00:00"},
		{9, "2024-11-16 12:00:00"},
		{10, "2024-11-17 12:00:00"},
		{11, "2024-11-18 12:00:00"},
		{12, "2024-11-19 12:00:00"},

		// December 2024 (6 photos)
		{13, "2024-12-15 12:00:00"},
		{14, "2024-12-16 12:00:00"},
		{15, "2024-12-17 12:00:00"},
		{16, "2024-12-18 12:00:00"},
		{17, "2024-12-19 12:00:00"},
		{18, "2024-12-20 12:00:00"},
	}

	for _, td := range testData {
		_, err := db.Exec(`
			INSERT INTO photos (id, file_path, date_taken, camera_make, camera_model)
			VALUES (?, ?, ?, 'Canon', 'EOS R5')
		`, td.id, "/test/photo"+string(rune('0'+td.id))+".jpg", td.date)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Create engine
	engine := NewEngine(db)

	// SCENARIO: User is viewing November 2024 (month=11, year=2024)
	year2024 := 2024
	month11 := 11
	currentParams := QueryParams{
		Year:  &year2024,
		Month: &month11,
		Limit: 100,
	}

	// Verify current query returns 5 photos
	result, err := engine.Query(currentParams)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if result.Total != 5 {
		t.Fatalf("Expected 5 photos for November 2024, got %d", result.Total)
	}

	// Compute facets
	facets, err := engine.ComputeFacets(currentParams)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Find 2023 in Year facet
	var year2023Facet *FacetValue
	for i := range facets.Year.Values {
		if facets.Year.Values[i].Value == "2023" {
			year2023Facet = &facets.Year.Values[i]
			break
		}
	}

	if year2023Facet == nil {
		t.Fatal("Expected to find 2023 in Year facet")
	}

	t.Logf("Year facet for 2023 shows count: %d", year2023Facet.Count)

	// VERIFICATION: The count MUST be 3 (November 2023 only, with month=11 preserved)
	if year2023Facet.Count != 3 {
		t.Errorf("❌ Year 2023 facet should show count=3 (November 2023 with month=11 preserved)")
		t.Errorf("  Got: %d", year2023Facet.Count)
		if year2023Facet.Count == 7 {
			t.Errorf("  BUG: This is 7 (all of 2023), meaning Month filter was cleared")
			t.Errorf("  The old hierarchical model is still active in computeYearFacet()")
		}
		t.Errorf("  This violates the state machine principle:")
		t.Errorf("    Facet counts must reflect the ACTUAL number of photos")
		t.Errorf("    that will be shown when the user clicks that facet value")
	} else {
		t.Logf("✅ Correct! Year 2023 shows count=3 (November 2023 only)")
	}

	// VERIFY: What happens when user clicks 2023?
	// The URL builder preserves month=11, so we'll transition to year=2023&month=11
	year2023 := 2023
	afterClickParams := QueryParams{
		Year:  &year2023,
		Month: &month11, // Month is preserved (correct URL builder behavior)
		Limit: 100,
	}

	afterClickResult, err := engine.Query(afterClickParams)
	if err != nil {
		t.Fatalf("Query after clicking 2023 failed: %v", err)
	}

	t.Logf("After clicking 2023, actual photos shown: %d", afterClickResult.Total)

	if afterClickResult.Total != 3 {
		t.Fatalf("Expected 3 photos for November 2023, got %d", afterClickResult.Total)
	}

	// THE CRITICAL REQUIREMENT: Facet count MUST match actual result count
	if year2023Facet.Count != afterClickResult.Total {
		t.Errorf("❌ CRITICAL BUG: Facet count (%d) doesn't match actual result count (%d)",
			year2023Facet.Count, afterClickResult.Total)
		t.Errorf("   User sees '2023 (%d)' but clicking it shows %d photos",
			year2023Facet.Count, afterClickResult.Total)
		t.Errorf("   This is confusing and violates user expectations!")
		t.Errorf("   FIX: Remove Month/Day clearing in computeYearFacet()")
	} else {
		t.Logf("✅ CORRECT: Facet count (%d) matches actual result count (%d)", year2023Facet.Count, afterClickResult.Total)
		t.Logf("✅ State machine model correctly implemented!")
	}
}

func TestFacetCountsCorrect_MonthPreservesFilters(t *testing.T) {
	t.Skip("Test demonstrates concept - full implementation would test month with other filters")
}

func TestFacetCountsCorrect_MonthPreservesDay(t *testing.T) {
	// Create simple test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE photos (
			id INTEGER PRIMARY KEY,
			file_path TEXT NOT NULL,
			date_taken DATETIME NOT NULL,
			camera_make TEXT,
			camera_model TEXT,
			lens_model TEXT,
			indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Insert test data:
	// 2024-11-15: 2 photos
	// 2024-11-16: 3 photos
	// 2024-12-15: 4 photos
	// 2024-12-16: 5 photos
	testData := []struct {
		id   int
		date string
	}{
		{1, "2024-11-15 10:00:00"},
		{2, "2024-11-15 14:00:00"},
		{3, "2024-11-16 10:00:00"},
		{4, "2024-11-16 12:00:00"},
		{5, "2024-11-16 14:00:00"},
		{6, "2024-12-15 10:00:00"},
		{7, "2024-12-15 12:00:00"},
		{8, "2024-12-15 14:00:00"},
		{9, "2024-12-15 16:00:00"},
		{10, "2024-12-16 10:00:00"},
		{11, "2024-12-16 12:00:00"},
		{12, "2024-12-16 14:00:00"},
		{13, "2024-12-16 16:00:00"},
		{14, "2024-12-16 18:00:00"},
	}

	for _, td := range testData {
		_, err := db.Exec(`
			INSERT INTO photos (id, file_path, date_taken, camera_make, camera_model)
			VALUES (?, ?, ?, 'Canon', 'EOS R5')
		`, td.id, "/test/photo"+string(rune('0'+td.id))+".jpg", td.date)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	engine := NewEngine(db)

	// SCENARIO: User is viewing November 15, 2024 (day=15, month=11, year=2024)
	year2024 := 2024
	month11 := 11
	day15 := 15
	currentParams := QueryParams{
		Year:  &year2024,
		Month: &month11,
		Day:   &day15,
		Limit: 100,
	}

	// Verify current query
	result, err := engine.Query(currentParams)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("Expected 2 photos for Nov 15, 2024, got %d", result.Total)
	}

	// Compute facets
	facets, err := engine.ComputeFacets(currentParams)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Check Month facet for December
	var december *FacetValue
	for i := range facets.Month.Values {
		if facets.Month.Values[i].Value == "12" {
			december = &facets.Month.Values[i]
			break
		}
	}

	if december == nil {
		t.Fatal("Expected to find December in Month facet")
	}

	t.Logf("December facet shows count: %d", december.Count)

	// The count should be 4 (December 15 only, with day=15 preserved)
	// But with the bug, it might show 9 (all of December)
	if december.Count != 4 {
		t.Errorf("December facet should show count=4 (Dec 15 with day=15 preserved), got %d", december.Count)
		if december.Count == 9 {
			t.Errorf("  BUG: Day filter was cleared when computing Month facet")
		}
	}
}
