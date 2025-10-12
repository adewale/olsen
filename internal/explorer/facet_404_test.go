package explorer

import (
	"fmt"
	"testing"

	"github.com/adewale/olsen/internal/query"
)

// TestFacet404Logging demonstrates how to write tests based on FACET_404 log entries.
//
// When the explorer server logs a FACET_404 entry, it includes:
// - path: The requested URL path
// - query: The query string (if any)
// - params: The parsed QueryParams (if parsing succeeded)
//
// Example log entries:
//   FACET_404: No route matched - path=/invalid_path query=
//   FACET_404: No results found - path=/color/red query= params={ColorName:[red] Limit:100 ...}
//   FACET_404: URL parse failed - path=/color/red/year/9999 query= error=...
//
// To create a test from a log entry:
// 1. Copy the path and query from the log
// 2. Create a test that verifies the expected behavior
// 3. For "No results found", ensure the query is valid but legitimately returns 0 results
// 4. For "No route matched", ensure it properly returns 404
// 5. For "URL parse failed", ensure the URL format is invalid

func TestFacet404Example_NoRouteMatched(t *testing.T) {
	// Based on log: FACET_404: No route matched - path=/invalid_path query=

	// This test would verify that invalid paths return 404
	// In a real test, you would:
	// 1. Create a test server
	// 2. Make a request to /invalid_path
	// 3. Verify it returns 404

	t.Skip("Example test - demonstrates test structure from log entry")
}

func TestFacet404Example_NoResults(t *testing.T) {
	// Based on log: FACET_404: No results found - path=/color/purple query=
	//   params={ColorName:[purple] Year:<nil> Month:<nil> ... Limit:100}

	// This test would verify that valid queries with no matching photos work correctly
	// In a real test, you would:
	// 1. Create a database with known data (e.g., photos without purple color)
	// 2. Execute a query for purple
	// 3. Verify it returns 0 results gracefully (not an error)
	// 4. Verify facets are still computed correctly

	mapper := query.NewURLMapper()
	params, err := mapper.ParsePath("/color/purple", "")
	if err != nil {
		t.Fatalf("URL should parse correctly: %v", err)
	}

	if len(params.ColourName) != 1 || params.ColourName[0] != "purple" {
		t.Errorf("Expected color=purple, got %v", params.ColourName)
	}

	// In real test: Execute query and verify Total == 0
}

func TestFacet404Example_InvalidCombination(t *testing.T) {
	// Based on log: FACET_404: No results found - path=/2025 query=color=red
	//   params={ColorName:[red] Year:2025 ... Limit:100}

	// This test would verify that valid but impossible combinations are handled
	// E.g., a year with no photos in that color

	mapper := query.NewURLMapper()
	params, err := mapper.ParsePath("/2025", "color=red")
	if err != nil {
		t.Fatalf("URL should parse correctly: %v", err)
	}

	// Verify the params were parsed correctly
	if params.Year == nil || *params.Year != 2025 {
		t.Errorf("Expected year=2025, got %v", params.Year)
	}
	if len(params.ColourName) != 1 || params.ColourName[0] != "red" {
		t.Errorf("Expected color=red, got %v", params.ColourName)
	}

	// In real test: Create DB with photos from 2025 but none red,
	// verify query returns 0 results but doesn't error
}

func TestFacet404Example_ParseError(t *testing.T) {
	// Based on log: FACET_404: URL parse failed - path=/color/red/year/9999 query= error=...

	// This test would verify that malformed URLs are rejected
	// In this case, the path structure is invalid (year value embedded in path incorrectly)

	mapper := query.NewURLMapper()
	_, err := mapper.ParsePath("/color/red/year/9999", "")

	// Depending on URL mapper behavior, this might parse or might fail
	// The log tells us it failed, so we can write a test to ensure it fails consistently
	if err == nil {
		fmt.Printf("Note: This URL now parses successfully. Params might have changed.\n")
		// In this case, you'd verify the parsed params match expectations
	}
}

// Guidelines for writing tests from FACET_404 logs:
//
// 1. NO ROUTE MATCHED logs indicate:
//    - Test that the URL properly returns 404
//    - Consider if this URL should be supported
//    - If it should be supported, add a route
//
// 2. NO RESULTS FOUND logs indicate:
//    - Test that valid queries with 0 results work correctly
//    - Verify no errors are thrown
//    - Verify facets are still computed
//    - Consider if this is expected (e.g., filtering by nonexistent value)
//
// 3. URL PARSE FAILED logs indicate:
//    - Test that the URL format is properly rejected
//    - Verify appropriate error message
//    - Consider if this URL format should be supported
//    - If it should be supported, update URLMapper
//
// 4. QUERY EXECUTION FAILED logs indicate:
//    - Test that the query construction is correct
//    - Verify database constraints are met
//    - This usually indicates a bug in query building
