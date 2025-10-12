package query

import (
	"strings"
	"testing"
)

// TestYearRemovalPreservesOtherFilters tests that removing a year filter
// preserves other filters like color
func TestYearRemovalPreservesOtherFilters(t *testing.T) {
	builder := NewFacetURLBuilder(NewURLMapper())

	// Start with Color: red + Year: 2025
	year2025 := 2025
	baseParams := QueryParams{
		ColourName: []string{"red"},
		Year:       &year2025,
		Limit:      1000,
	}

	// Create facet with year 2025 selected
	yearFacet := &Facet{
		Name:  "year",
		Label: "Year",
		Values: []FacetValue{
			{Value: "2025", Label: "2025", Count: 4, Selected: true},
		},
	}

	// Build URLs
	builder.buildYearURLs(yearFacet, baseParams)

	// Check the URL for the selected year (clicking it should remove year, keep color)
	url := yearFacet.Values[0].URL
	t.Logf("Generated URL: %s", url)

	// Parse the URL back
	parts := strings.SplitN(url, "?", 2)
	path := parts[0]
	query := ""
	if len(parts) == 2 {
		query = parts[1]
	}

	mapper := NewURLMapper()
	parsedParams, err := mapper.ParsePath(path, query)
	if err != nil {
		t.Fatalf("Failed to parse URL '%s': %v", url, err)
	}

	// Should have color but NO year
	if len(parsedParams.ColourName) == 0 || parsedParams.ColourName[0] != "red" {
		t.Errorf("Color filter not preserved. Expected [red], got %v. URL: %s", parsedParams.ColourName, url)
	}

	if parsedParams.Year != nil {
		t.Errorf("Year should be removed. Expected nil, got %d. URL: %s", *parsedParams.Year, url)
	}
}

// TestTimeOfDayURLParsing tests that extended time-of-day values parse correctly
func TestTimeOfDayURLParsing(t *testing.T) {
	mapper := NewURLMapper()

	testCases := []struct {
		path     string
		expected string
	}{
		{"/blue_hour", "blue_hour"},
		{"/golden_hour_morning", "golden_hour_morning"},
		{"/golden_hour_evening", "golden_hour_evening"},
		{"/midday", "midday"},
		{"/morning", "morning"},
		{"/afternoon", "afternoon"},
		{"/evening", "evening"},
		{"/night", "night"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			params, err := mapper.ParsePath(tc.path, "")
			if err != nil {
				t.Fatalf("Failed to parse path '%s': %v", tc.path, err)
			}

			if len(params.TimeOfDay) == 0 {
				t.Errorf("TimeOfDay not set for path '%s'", tc.path)
				return
			}

			if params.TimeOfDay[0] != tc.expected {
				t.Errorf("Wrong TimeOfDay. Expected '%s', got '%s'", tc.expected, params.TimeOfDay[0])
			}
		})
	}
}

// TestColorRemovalPreservesYear tests that removing color preserves year
func TestColorRemovalPreservesYear(t *testing.T) {
	builder := NewFacetURLBuilder(NewURLMapper())

	// Start with Color: red + Year: 2025
	year2025 := 2025
	baseParams := QueryParams{
		ColourName: []string{"red"},
		Year:       &year2025,
		Limit:      1000,
	}

	// Create facet with red selected
	colorFacet := &Facet{
		Name:  "color",
		Label: "Color",
		Values: []FacetValue{
			{Value: "red", Label: "Red", Count: 4, Selected: true},
		},
	}

	// Build URLs
	builder.buildColourURLs(colorFacet, baseParams)

	// Check the URL for the selected color (clicking it should remove color, keep year)
	url := colorFacet.Values[0].URL
	t.Logf("Generated URL: %s", url)

	// Parse the URL back
	parts := strings.SplitN(url, "?", 2)
	path := parts[0]
	query := ""
	if len(parts) == 2 {
		query = parts[1]
	}

	mapper := NewURLMapper()
	parsedParams, err := mapper.ParsePath(path, query)
	if err != nil {
		t.Fatalf("Failed to parse URL '%s': %v", url, err)
	}

	// Should have year but NO color
	if parsedParams.Year == nil || *parsedParams.Year != 2025 {
		t.Errorf("Year filter not preserved. Expected 2025, got %v. URL: %s", parsedParams.Year, url)
	}

	if len(parsedParams.ColourName) > 0 {
		t.Errorf("Color should be removed. Expected [], got %v. URL: %s", parsedParams.ColourName, url)
	}
}
