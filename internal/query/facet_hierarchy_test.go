package query

import (
	"strings"
	"testing"
)

// NOTE: This file used to test "hierarchical" behavior where changing year cleared month/day.
// That was INCORRECT. Facets are not hierarchical - they're a state machine model where
// ALL filters are preserved during transitions. The facet computation determines which
// transitions are valid (have results > 0).
//
// See specs/facet_state_machine.spec for the correct mental model.

// TestYearFacetPreservesMonthAndDay verifies that selecting a different year
// PRESERVES month and day filters (state machine model, not hierarchical)
func TestYearFacetPreservesMonthAndDay(t *testing.T) {
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	// Setup: User is viewing November 2024
	year2024 := 2024
	month11 := 11
	baseParams := QueryParams{
		Year:  &year2024,
		Month: &month11,
		Limit: 100,
	}

	// Create year facet with 2025 option
	yearFacet := &Facet{
		Name:  "year",
		Label: "Year",
		Values: []FacetValue{
			{Value: "2025", Label: "2025", Count: 100, Selected: false},
		},
	}

	// Act: Build URLs
	builder.buildYearURLs(yearFacet, baseParams)

	// Assert: URL should have year=2025 AND month=11
	// The facet computation will determine if this state has results
	url := yearFacet.Values[0].URL

	if !strings.Contains(url, "year=2025") {
		t.Errorf("Expected URL to contain year=2025, got: %s", url)
	}

	if !strings.Contains(url, "month=11") {
		t.Errorf("Expected month to be PRESERVED when year changes, got: %s", url)
	}

	t.Logf("✅ Year change preserves month filter: %s", url)
}

// TestMonthFacetPreservesDay verifies that selecting a different month
// PRESERVES day filter (state machine model)
func TestMonthFacetPreservesDay(t *testing.T) {
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	// Setup: User is viewing Nov 15, 2024
	year2024 := 2024
	month11 := 11
	day15 := 15
	baseParams := QueryParams{
		Year:  &year2024,
		Month: &month11,
		Day:   &day15,
		Limit: 100,
	}

	// Create month facet with December option
	monthFacet := &Facet{
		Name:  "month",
		Label: "Month",
		Values: []FacetValue{
			{Value: "12", Label: "December", Count: 50, Selected: false},
		},
	}

	// Act: Build URLs
	builder.buildMonthURLs(monthFacet, baseParams)

	// Assert: URL should have year=2024, month=12 AND day=15
	url := monthFacet.Values[0].URL

	if !strings.Contains(url, "year=2024") {
		t.Errorf("Expected URL to preserve year=2024, got: %s", url)
	}

	if !strings.Contains(url, "month=12") {
		t.Errorf("Expected URL to contain month=12, got: %s", url)
	}

	if !strings.Contains(url, "day=15") {
		t.Errorf("Expected day to be PRESERVED when month changes, got: %s", url)
	}

	t.Logf("✅ Month change preserves day and year: %s", url)
}

// TestRemovingYearPreservesMonth verifies that deselecting year
// PRESERVES month and day filters (state machine model)
func TestRemovingYearPreservesMonth(t *testing.T) {
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	// Setup: User is viewing November 2024
	year2024 := 2024
	month11 := 11
	baseParams := QueryParams{
		Year:  &year2024,
		Month: &month11,
		Limit: 100,
	}

	// Create year facet with 2024 selected
	yearFacet := &Facet{
		Name:  "year",
		Label: "Year",
		Values: []FacetValue{
			{Value: "2024", Label: "2024", Count: 200, Selected: true},
		},
	}

	// Act: Build URLs (clicking selected year removes it)
	builder.buildYearURLs(yearFacet, baseParams)

	// Assert: URL should NOT have year, but SHOULD have month
	// This means "all Novembers across all years"
	url := yearFacet.Values[0].URL

	if strings.Contains(url, "year=") {
		t.Errorf("Expected year to be removed, got: %s", url)
	}

	if !strings.Contains(url, "month=11") {
		t.Errorf("Expected month to be PRESERVED when year is removed, got: %s", url)
	}

	t.Logf("✅ Removing year preserves month filter: %s", url)
}
