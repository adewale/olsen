package query

import (
	"strings"
	"testing"
)

// TestFacetStateTransitions verifies all supported state transitions between facet combinations
// This ensures that facet URL generation behaves correctly when transitioning from one filter
// combination to another.

// Test suite structure - STATE MACHINE MODEL (not hierarchical):
// ALL facets are independent and preserve each other's values during transitions.
// The facet computation layer determines which transitions are valid (have results > 0).
//
// 1. Temporal facets (Year, Month, Day) - INDEPENDENT (preserved during transitions)
// 2. Visual facets (Color, TimeOfDay, Season) - independent
// 3. Equipment facets (Camera, Lens) - independent
// 4. Special facets (InBurst) - independent
// 5. Cross-category transitions - all preserve each other

func TestTransition_Empty_To_Year(t *testing.T) {
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	baseParams := QueryParams{Limit: 100}

	yearFacet := &Facet{
		Name:   "year",
		Label:  "Year",
		Values: []FacetValue{{Value: "2024", Label: "2024", Count: 100, Selected: false}},
	}

	builder.buildYearURLs(yearFacet, baseParams)

	url := yearFacet.Values[0].URL
	if !strings.Contains(url, "year=2024") {
		t.Errorf("Expected year=2024 in URL, got: %s", url)
	}
	if strings.Contains(url, "month=") || strings.Contains(url, "day=") {
		t.Errorf("URL should not contain month or day, got: %s", url)
	}
}

func TestTransition_Year_To_Color(t *testing.T) {
	// Adding a color filter should preserve year
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	baseParams := QueryParams{Year: &year2024, Limit: 100}

	colorFacet := &Facet{
		Name:   "color",
		Label:  "Color",
		Values: []FacetValue{{Value: "red", Label: "Red", Count: 50, Selected: false}},
	}

	builder.buildColourURLs(colorFacet, baseParams)

	url := colorFacet.Values[0].URL
	if !strings.Contains(url, "year=2024") {
		t.Errorf("Expected year to be preserved, got: %s", url)
	}
	if !strings.Contains(url, "color=red") {
		t.Errorf("Expected color=red in URL, got: %s", url)
	}
}

func TestTransition_YearMonth_To_DifferentYear(t *testing.T) {
	// STATE MACHINE MODEL: changing year PRESERVES month
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	month11 := 11
	baseParams := QueryParams{Year: &year2024, Month: &month11, Limit: 100}

	yearFacet := &Facet{
		Name:   "year",
		Label:  "Year",
		Values: []FacetValue{{Value: "2025", Label: "2025", Count: 80, Selected: false}},
	}

	builder.buildYearURLs(yearFacet, baseParams)

	url := yearFacet.Values[0].URL
	if !strings.Contains(url, "year=2025") {
		t.Errorf("Expected year=2025, got: %s", url)
	}
	if !strings.Contains(url, "month=11") {
		t.Errorf("Expected month to be PRESERVED when year changes, got: %s", url)
	}
}

func TestTransition_YearColorCamera_To_DifferentYear(t *testing.T) {
	// STATE MACHINE MODEL: changing year preserves ALL other filters including month
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	month11 := 11
	baseParams := QueryParams{
		Year:       &year2024,
		Month:      &month11,
		ColourName: []string{"red"},
		CameraMake: []string{"Canon"},
		Limit:      100,
	}

	yearFacet := &Facet{
		Name:   "year",
		Label:  "Year",
		Values: []FacetValue{{Value: "2025", Label: "2025", Count: 50, Selected: false}},
	}

	builder.buildYearURLs(yearFacet, baseParams)

	url := yearFacet.Values[0].URL
	if !strings.Contains(url, "year=2025") {
		t.Errorf("Expected year=2025, got: %s", url)
	}
	if !strings.Contains(url, "month=11") {
		t.Errorf("Expected month to be PRESERVED, got: %s", url)
	}
	if !strings.Contains(url, "color=red") {
		t.Errorf("Expected color to be preserved, got: %s", url)
	}
	if !strings.Contains(url, "camera_make=Canon") {
		t.Errorf("Expected camera to be preserved, got: %s", url)
	}
}

func TestTransition_Color_To_DifferentColor(t *testing.T) {
	// Changing color should replace the color filter
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	baseParams := QueryParams{ColourName: []string{"red"}, Limit: 100}

	colorFacet := &Facet{
		Name:  "color",
		Label: "Color",
		Values: []FacetValue{
			{Value: "red", Label: "Red", Count: 50, Selected: true},
			{Value: "blue", Label: "Blue", Count: 40, Selected: false},
		},
	}

	builder.buildColourURLs(colorFacet, baseParams)

	// Red is selected, clicking it should remove the filter
	redURL := colorFacet.Values[0].URL
	if strings.Contains(redURL, "color=") {
		t.Errorf("Clicking selected color should remove filter, got: %s", redURL)
	}

	// Blue is not selected, clicking it should set it
	blueURL := colorFacet.Values[1].URL
	if !strings.Contains(blueURL, "color=blue") {
		t.Errorf("Expected color=blue in URL, got: %s", blueURL)
	}
}

func TestTransition_RemoveYearWithColorAndCamera(t *testing.T) {
	// STATE MACHINE MODEL: removing year preserves ALL other filters including month
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	month11 := 11
	baseParams := QueryParams{
		Year:       &year2024,
		Month:      &month11,
		ColourName: []string{"red"},
		CameraMake: []string{"Canon"},
		Limit:      100,
	}

	yearFacet := &Facet{
		Name:   "year",
		Label:  "Year",
		Values: []FacetValue{{Value: "2024", Label: "2024", Count: 100, Selected: true}},
	}

	builder.buildYearURLs(yearFacet, baseParams)

	url := yearFacet.Values[0].URL
	if strings.Contains(url, "year=") {
		t.Errorf("Expected year to be removed, got: %s", url)
	}
	if !strings.Contains(url, "month=11") {
		t.Errorf("Expected month to be PRESERVED, got: %s", url)
	}
	if !strings.Contains(url, "color=red") {
		t.Errorf("Expected color to be preserved, got: %s", url)
	}
	if !strings.Contains(url, "camera_make=Canon") {
		t.Errorf("Expected camera to be preserved, got: %s", url)
	}
}

func TestTransition_MonthDay_To_DifferentMonth(t *testing.T) {
	// STATE MACHINE MODEL: changing month preserves day and year
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	month11 := 11
	day15 := 15
	baseParams := QueryParams{Year: &year2024, Month: &month11, Day: &day15, Limit: 100}

	monthFacet := &Facet{
		Name:   "month",
		Label:  "Month",
		Values: []FacetValue{{Value: "12", Label: "December", Count: 50, Selected: false}},
	}

	builder.buildMonthURLs(monthFacet, baseParams)

	url := monthFacet.Values[0].URL
	if !strings.Contains(url, "year=2024") {
		t.Errorf("Expected year to be preserved, got: %s", url)
	}
	if !strings.Contains(url, "month=12") {
		t.Errorf("Expected month=12, got: %s", url)
	}
	if !strings.Contains(url, "day=15") {
		t.Errorf("Expected day to be PRESERVED when month changes, got: %s", url)
	}
}

func TestTransition_CameraAndLens_To_DifferentCamera(t *testing.T) {
	// Changing camera should work independently (no hierarchy)
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	baseParams := QueryParams{
		CameraMake:  []string{"Canon"},
		CameraModel: []string{"EOS R5"},
		LensModel:   []string{"RF 24-70mm"},
		Limit:       100,
	}

	cameraFacet := &Facet{
		Name:  "camera",
		Label: "Camera",
		Values: []FacetValue{
			{Value: "Canon EOS R5", Label: "Canon EOS R5", Count: 100, Selected: true},
			{Value: "Sony A7 IV", Label: "Sony A7 IV", Count: 80, Selected: false},
		},
	}

	builder.buildCameraURLs(cameraFacet, baseParams)

	// Canon is selected, clicking should remove it (and preserve lens)
	canonURL := cameraFacet.Values[0].URL
	if strings.Contains(canonURL, "camera_make=") || strings.Contains(canonURL, "camera_model=") {
		t.Errorf("Clicking selected camera should remove filter, got: %s", canonURL)
	}
	if !strings.Contains(canonURL, "lens=RF") {
		t.Errorf("Expected lens to be preserved, got: %s", canonURL)
	}

	// Sony not selected, clicking should set it (and preserve lens)
	sonyURL := cameraFacet.Values[1].URL
	if !strings.Contains(sonyURL, "camera_make=Sony") {
		t.Errorf("Expected camera_make=Sony, got: %s", sonyURL)
	}
	if !strings.Contains(sonyURL, "lens=RF") {
		t.Errorf("Expected lens to be preserved, got: %s", sonyURL)
	}
}

func TestTransition_MultipleIndependentFilters(t *testing.T) {
	// Test that independent facets don't affect each other
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	baseParams := QueryParams{
		Year:       &year2024,
		ColourName: []string{"red"},
		TimeOfDay:  []string{"golden_am"},
		Limit:      100,
	}

	// Add another time of day
	timeOfDayFacet := &Facet{
		Name:  "time_of_day",
		Label: "Time of Day",
		Values: []FacetValue{
			{Value: "golden_am", Label: "Golden Hour AM", Count: 50, Selected: true},
			{Value: "midday", Label: "Midday", Count: 40, Selected: false},
		},
	}

	builder.buildTimeOfDayURLs(timeOfDayFacet, baseParams)

	// Removing golden_am
	removeURL := timeOfDayFacet.Values[0].URL
	if strings.Contains(removeURL, "time_of_day=golden_am") {
		t.Errorf("Expected golden_am to be removed, got: %s", removeURL)
	}
	if !strings.Contains(removeURL, "year=2024") {
		t.Errorf("Expected year to be preserved, got: %s", removeURL)
	}
	if !strings.Contains(removeURL, "color=red") {
		t.Errorf("Expected color to be preserved, got: %s", removeURL)
	}

	// Adding midday
	addURL := timeOfDayFacet.Values[1].URL
	if !strings.Contains(addURL, "time_of_day=midday") {
		t.Errorf("Expected time_of_day=midday, got: %s", addURL)
	}
	if !strings.Contains(addURL, "time_of_day=golden_am") {
		t.Errorf("Expected existing time_of_day to be preserved (multi-select), got: %s", addURL)
	}
}

func TestTransition_ComplexState(t *testing.T) {
	// STATE MACHINE MODEL: Test a complex state with multiple filters across all categories
	// Changing year preserves EVERYTHING including month
	mapper := NewURLMapper()
	builder := NewFacetURLBuilder(mapper)

	year2024 := 2024
	month6 := 6
	inBurst := true
	baseParams := QueryParams{
		Year:              &year2024,
		Month:             &month6,
		ColourName:        []string{"blue"},
		CameraMake:        []string{"Canon"},
		CameraModel:       []string{"EOS R5"},
		LensModel:         []string{"RF 24-70mm"},
		TimeOfDay:         []string{"golden_am", "golden_pm"},
		Season:            []string{"summer"},
		FocalCategory:     []string{"wide"},
		ShootingCondition: []string{"sunny"},
		InBurst:           &inBurst,
		Limit:             100,
	}

	// Try changing the year - should preserve EVERYTHING including month
	yearFacet := &Facet{
		Name:   "year",
		Label:  "Year",
		Values: []FacetValue{{Value: "2023", Label: "2023", Count: 150, Selected: false}},
	}

	builder.buildYearURLs(yearFacet, baseParams)
	url := yearFacet.Values[0].URL

	// Check year changed
	if !strings.Contains(url, "year=2023") {
		t.Errorf("Expected year=2023, got: %s", url)
	}
	// Check ALL filters preserved including month
	requiredParams := []string{
		"month=6", // PRESERVED in state machine model
		"color=blue",
		"camera_make=Canon",
		"lens=RF",
		"time_of_day=golden_am",
		"season=summer",
		"focal_category=wide",
		"shooting_condition=sunny",
		"in_burst=true",
	}

	for _, param := range requiredParams {
		if !strings.Contains(url, param) {
			t.Errorf("Expected %s to be preserved, got: %s", param, url)
		}
	}
}
