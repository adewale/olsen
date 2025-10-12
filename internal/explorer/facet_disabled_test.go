package explorer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/adewale/olsen/internal/query"
)

// TestDisabledFacetRendering tests that facet values with count=0 are properly disabled in the UI.
// This verifies Phase 2 of the state machine migration: preventing invalid state transitions at the UI level.
//
// The fundamental rule: Users must never be able to transition from a state with results (count > 0)
// to a state with zero results (count = 0).
//
// Implementation:
// - Facet values with count=0 are rendered as <span> not <a> (not clickable)
// - CSS class "disabled" is applied
// - Tooltip shows "No results with current filters"
// - Visual styling: 40% opacity, cursor: not-allowed, pointer-events: none

// emptyFacetCollection creates a FacetCollection with all facets initialized but empty
func emptyFacetCollection() *query.FacetCollection {
	return &query.FacetCollection{
		Year:              &query.Facet{Name: "year", Label: "Year", Values: []query.FacetValue{}},
		Month:             &query.Facet{Name: "month", Label: "Month", Values: []query.FacetValue{}},
		Camera:            &query.Facet{Name: "camera", Label: "Camera", Values: []query.FacetValue{}},
		Lens:              &query.Facet{Name: "lens", Label: "Lens", Values: []query.FacetValue{}},
		TimeOfDay:         &query.Facet{Name: "time_of_day", Label: "Time of Day", Values: []query.FacetValue{}},
		ColourName:        &query.Facet{Name: "colour", Label: "Colour", Values: []query.FacetValue{}},
		InBurst:           &query.Facet{Name: "in_burst", Label: "In Burst", Values: []query.FacetValue{}},
		Season:            &query.Facet{Name: "season", Label: "Season", Values: []query.FacetValue{}},
		FocalCategory:     &query.Facet{Name: "focal_category", Label: "Focal Category", Values: []query.FacetValue{}},
		ShootingCondition: &query.Facet{Name: "shooting_condition", Label: "Shooting Condition", Values: []query.FacetValue{}},
	}
}

func TestYearFacetDisabledRendering(t *testing.T) {
	// Setup: Create facet collection with a year that has 0 count
	facets := emptyFacetCollection()
	facets.Year.Values = []query.FacetValue{
		{
			Value:    "2024",
			Label:    "2024",
			Count:    50,
			Selected: false,
			URL:      "/photos?year=2024",
		},
		{
			Value:    "2025",
			Label:    "2025",
			Count:    0, // Zero count - should be disabled
			Selected: false,
			URL:      "/photos?year=2025",
		},
	}

	// Render the template
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 50,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: 2024 (count > 0) is rendered as clickable link
	if !strings.Contains(html, `<a href="/photos?year=2024"`) {
		t.Error("Expected 2024 to be rendered as clickable link")
	}

	// Verify: 2025 (count = 0) is NOT rendered as link
	if strings.Contains(html, `<a href="/photos?year=2025"`) {
		t.Error("Expected 2025 NOT to be rendered as clickable link (count=0)")
	}

	// Verify: 2025 has disabled class
	// Look for the pattern: disabled facet item containing 2025
	if !strings.Contains(html, "disabled") {
		t.Error("Expected disabled class to be present for zero-count facet")
	}

	// Verify: Tooltip message is present
	if !strings.Contains(html, "No results with current filters") {
		t.Error("Expected tooltip message for disabled facet")
	}

	// Verify: Both facet values are present in the HTML (disabled facets shown, not hidden)
	if !strings.Contains(html, "2024") {
		t.Error("Expected 2024 to be present in rendered HTML")
	}
	if !strings.Contains(html, "2025") {
		t.Error("Expected 2025 to be present in rendered HTML (visible but disabled)")
	}
}

func TestMonthFacetDisabledRendering(t *testing.T) {
	// Setup: Month facet with mixed counts
	facets := emptyFacetCollection()
	facets.Month.Values = []query.FacetValue{
		{
			Value:    "11",
			Label:    "November",
			Count:    50,
			Selected: true,
			URL:      "/photos?year=2024&month=11",
		},
		{
			Value:    "12",
			Label:    "December",
			Count:    0, // Zero count - should be disabled
			Selected: false,
			URL:      "/photos?year=2024&month=12",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 50,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: November (selected) is clickable
	if !strings.Contains(html, "November") {
		t.Error("Expected November to be present")
	}

	// Verify: December (count=0) has disabled class
	if !strings.Contains(html, "disabled") {
		t.Error("Expected disabled class for December (count=0)")
	}

	// Verify: December is not a link
	if strings.Contains(html, `<a href="/photos?year=2024&month=12"`) {
		t.Error("Expected December NOT to be clickable link (count=0)")
	}
}

func TestCameraFacetDisabledRendering(t *testing.T) {
	// Setup: Camera facet where one camera has no photos with current filters
	facets := emptyFacetCollection()
	facets.Camera.Values = []query.FacetValue{
		{
			Value:    "Canon EOS R5",
			Label:    "Canon EOS R5",
			Count:    30,
			Selected: false,
			URL:      "/photos?camera_make=Canon&camera_model=EOS+R5",
		},
		{
			Value:    "Sony A7R V",
			Label:    "Sony A7R V",
			Count:    0, // No Sony photos with current filters
			Selected: false,
			URL:      "/photos?camera_make=Sony&camera_model=A7R+V",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 30,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: Canon is clickable
	if !strings.Contains(html, "Canon EOS R5") {
		t.Error("Expected Canon EOS R5 to be present")
	}

	// Verify: Sony (count=0) is disabled
	if !strings.Contains(html, "disabled") {
		t.Error("Expected disabled class for Sony (count=0)")
	}

	// Verify: Sony is present but not clickable
	if !strings.Contains(html, "Sony A7R V") {
		t.Error("Expected Sony A7R V to be present (visible but disabled)")
	}
}

func TestColourFacetDisabledRendering(t *testing.T) {
	// Setup: Colour facet with swatch-style rendering
	facets := emptyFacetCollection()
	facets.ColourName.Values = []query.FacetValue{
		{
			Value:    "red",
			Label:    "Red",
			Count:    20,
			Selected: false,
			URL:      "/photos?colour=red",
		},
		{
			Value:    "green",
			Label:    "Green",
			Count:    0, // No green photos with current filters
			Selected: false,
			URL:      "/photos?colour=green",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 20,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: Red is clickable (color-swatch with link)
	if !strings.Contains(html, `href="/photos?colour=red"`) {
		t.Error("Expected red color swatch to be clickable")
	}

	// Verify: Green (count=0) has disabled class
	if !strings.Contains(html, "color-swatch disabled") {
		t.Error("Expected .color-swatch.disabled class for green color (count=0)")
	}

	// Verify: Green swatch is present but not clickable
	if strings.Contains(html, `href="/photos?colour=green"`) {
		t.Error("Expected green color swatch NOT to be clickable (count=0)")
	}

	// Verify: Disabled tooltip is present
	if !strings.Contains(html, "No results with current filters") {
		t.Error("Expected tooltip for disabled color swatch")
	}
}

func TestTimeOfDayChipFacetDisabledRendering(t *testing.T) {
	// Setup: Time of Day facet (chip-style rendering)
	facets := emptyFacetCollection()
	facets.TimeOfDay.Values = []query.FacetValue{
		{
			Value:    "golden_hour",
			Label:    "Golden Hour",
			Count:    15,
			Selected: false,
			URL:      "/photos?time_of_day=golden_hour",
		},
		{
			Value:    "blue_hour",
			Label:    "Blue Hour",
			Count:    0, // No blue hour photos with current filters
			Selected: false,
			URL:      "/photos?time_of_day=blue_hour",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 15,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: Golden Hour is clickable
	if !strings.Contains(html, "Golden Hour") {
		t.Error("Expected Golden Hour to be present")
	}

	// Verify: Blue Hour (count=0) has disabled class
	if !strings.Contains(html, "disabled") {
		t.Error("Expected disabled class for Blue Hour (count=0)")
	}

	// Verify: Blue Hour is present but not clickable
	if !strings.Contains(html, "Blue Hour") {
		t.Error("Expected Blue Hour to be present (visible but disabled)")
	}

	// For chip-style facets, verify it's rendered as <span> not <a>
	if strings.Contains(html, `<a href="/photos?time_of_day=blue_hour"`) {
		t.Error("Expected Blue Hour NOT to be clickable link (count=0)")
	}
}

func TestInBurstChipFacetDisabledRendering(t *testing.T) {
	// Setup: InBurst facet (chip-style, typically binary true/false)
	facets := emptyFacetCollection()
	facets.InBurst.Values = []query.FacetValue{
		{
			Value:    "true",
			Label:    "In Burst",
			Count:    10,
			Selected: false,
			URL:      "/photos?in_burst=true",
		},
		{
			Value:    "false",
			Label:    "Not in Burst",
			Count:    0, // No non-burst photos with current filters
			Selected: false,
			URL:      "/photos?in_burst=false",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 10,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: "In Burst" is clickable
	if !strings.Contains(html, "In Burst") {
		t.Error("Expected 'In Burst' to be present")
	}

	// Verify: "Not in Burst" (count=0) has disabled class
	if !strings.Contains(html, "disabled") {
		t.Error("Expected disabled class for 'Not in Burst' (count=0)")
	}
}

func TestAllFacetsDisabled_ZeroResults(t *testing.T) {
	// Setup: Extreme case - user has filtered to a state where changing any facet leads to 0 results
	// This shouldn't happen in production (we prevent invalid transitions), but test the rendering
	facets := emptyFacetCollection()
	facets.Year.Values = []query.FacetValue{
		{
			Value:    "2024",
			Label:    "2024",
			Count:    0, // All options have 0 count
			Selected: false,
			URL:      "/photos?year=2024",
		},
	}
	facets.ColourName.Values = []query.FacetValue{
		{
			Value:    "red",
			Label:    "Red",
			Count:    0, // All options have 0 count
			Selected: false,
			URL:      "/photos?colour=red",
		},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 0,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: All facet values are disabled
	disabledCount := strings.Count(html, "disabled")
	if disabledCount < 2 {
		t.Errorf("Expected at least 2 disabled facets, got %d", disabledCount)
	}

	// Verify: Tooltip message is present multiple times
	tooltipCount := strings.Count(html, "No results with current filters")
	if tooltipCount < 2 {
		t.Errorf("Expected at least 2 tooltip messages, got %d", tooltipCount)
	}
}

func TestMixedEnabledDisabledFacets(t *testing.T) {
	// Setup: Realistic scenario with mix of enabled and disabled facet values
	facets := emptyFacetCollection()
	facets.Year.Values = []query.FacetValue{
		{Value: "2022", Label: "2022", Count: 100, Selected: false, URL: "/photos?year=2022"},
		{Value: "2023", Label: "2023", Count: 120, Selected: false, URL: "/photos?year=2023"},
		{Value: "2024", Label: "2024", Count: 50, Selected: true, URL: "/photos?year=2024"},
		{Value: "2025", Label: "2025", Count: 0, Selected: false, URL: "/photos?year=2025"}, // Disabled
	}
	facets.ColourName.Values = []query.FacetValue{
		{Value: "red", Label: "Red", Count: 10, Selected: false, URL: "/photos?year=2024&colour=red"},
		{Value: "blue", Label: "Blue", Count: 20, Selected: false, URL: "/photos?year=2024&colour=blue"},
		{Value: "green", Label: "Green", Count: 0, Selected: false, URL: "/photos?year=2024&colour=green"},    // Disabled
		{Value: "yellow", Label: "Yellow", Count: 0, Selected: false, URL: "/photos?year=2024&colour=yellow"}, // Disabled
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 50,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: Enabled years are clickable
	if !strings.Contains(html, `<a href="/photos?year=2022"`) {
		t.Error("Expected 2022 to be clickable")
	}
	if !strings.Contains(html, `<a href="/photos?year=2023"`) {
		t.Error("Expected 2023 to be clickable")
	}

	// Verify: 2025 is NOT clickable
	if strings.Contains(html, `<a href="/photos?year=2025"`) {
		t.Error("Expected 2025 NOT to be clickable (count=0)")
	}

	// Verify: Enabled colors are clickable
	if !strings.Contains(html, "Red") && !strings.Contains(html, "Blue") {
		t.Error("Expected enabled colors to be present")
	}

	// Verify: We have multiple disabled elements (2025, green, yellow)
	disabledCount := strings.Count(html, "disabled")
	if disabledCount < 3 {
		t.Errorf("Expected at least 3 disabled facets (2025, green, yellow), got %d", disabledCount)
	}

	// Verify: All year values are present (disabled values are visible)
	for _, year := range []string{"2022", "2023", "2024", "2025"} {
		if !strings.Contains(html, year) {
			t.Errorf("Expected year %s to be present in HTML", year)
		}
	}

	// Verify: Color swatches are rendered (check by URL presence for enabled, class for disabled)
	// Note: URLs might be query-param style or path-style depending on template
	if !strings.Contains(html, `colour=red`) {
		t.Error("Expected red color swatch to be clickable")
	}
	if !strings.Contains(html, `colour=blue`) {
		t.Error("Expected blue color swatch to be clickable")
	}

	// Verify: Green and yellow should be disabled (check for disabled class)
	// Count color-swatch disabled instances - should be at least 2 (green + yellow)
	disabledSwatchCount := strings.Count(html, "color-swatch disabled")
	if disabledSwatchCount < 2 {
		t.Errorf("Expected at least 2 disabled color swatches (green, yellow), got %d", disabledSwatchCount)
	}
}

// TestDisabledFacetCSSClasses verifies that the correct CSS classes are applied
func TestDisabledFacetCSSClasses(t *testing.T) {
	facets := emptyFacetCollection()
	facets.Year.Values = []query.FacetValue{
		{Value: "2025", Label: "2025", Count: 0, Selected: false, URL: "/photos?year=2025"},
	}
	facets.TimeOfDay.Values = []query.FacetValue{
		{Value: "night", Label: "Night", Count: 0, Selected: false, URL: "/photos?time_of_day=night"},
	}
	facets.ColourName.Values = []query.FacetValue{
		{Value: "purple", Label: "Purple", Count: 0, Selected: false, URL: "/photos?colour=purple"},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 0,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: List-style facets use .facet-item.disabled
	if !strings.Contains(html, "facet-item disabled") && !strings.Contains(html, `class="facet-item disabled"`) {
		t.Error("Expected .facet-item.disabled class for list-style facets (Year)")
	}

	// Verify: Chip-style facets use .facet-chip.disabled
	if !strings.Contains(html, "facet-chip disabled") && !strings.Contains(html, `class="facet-chip disabled"`) {
		t.Error("Expected .facet-chip.disabled class for chip-style facets (TimeOfDay)")
	}

	// Verify: Swatch-style facets use .color-swatch.disabled
	if !strings.Contains(html, "color-swatch disabled") && !strings.Contains(html, `class="color-swatch disabled"`) {
		t.Error("Expected .color-swatch.disabled class for swatch-style facets (Colour)")
	}
}

// TestNoDisabledFacets_AllValid verifies that when all facets have count > 0, none are disabled
func TestNoDisabledFacets_AllValid(t *testing.T) {
	// Setup: All facet values have results
	facets := emptyFacetCollection()
	facets.Year.Values = []query.FacetValue{
		{Value: "2023", Label: "2023", Count: 120, Selected: false, URL: "/photos?year=2023"},
		{Value: "2024", Label: "2024", Count: 50, Selected: true, URL: "/photos?year=2024"},
	}
	facets.ColourName.Values = []query.FacetValue{
		{Value: "red", Label: "Red", Count: 10, Selected: false, URL: "/photos?year=2024&colour=red"},
		{Value: "blue", Label: "Blue", Count: 20, Selected: false, URL: "/photos?year=2024&colour=blue"},
	}

	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, "grid", map[string]interface{}{
		"Facets":     facets,
		"Photos":     []PhotoCard{},
		"TotalCount": 50,
	})
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	html := buf.String()

	// Verify: No disabled facet items (note: grid template may contain other uses of "disabled")
	// Check specifically for facet-item disabled or color-swatch disabled
	facetItemDisabled := strings.Contains(html, "facet-item disabled")
	facetChipDisabled := strings.Contains(html, "facet-chip disabled")
	colorSwatchDisabled := strings.Contains(html, "color-swatch disabled")

	if facetItemDisabled || facetChipDisabled || colorSwatchDisabled {
		t.Error("Expected no disabled facets when all have count > 0")
	}

	// Verify: All facet values are clickable links
	if !strings.Contains(html, `<a href="/photos?year=2023"`) {
		t.Error("Expected 2023 to be clickable")
	}
	if !strings.Contains(html, `<a href="/photos?year=2024"`) {
		t.Error("Expected 2024 to be clickable")
	}
}
