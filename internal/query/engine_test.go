package query

import (
	"testing"

	"github.com/adewale/olsen/internal/database"
)

func TestQueryEngine(t *testing.T) {
	// Open test database
	db, err := database.Open("../../test_query.db")
	if err != nil {
		t.Skipf("Test database not found: %v", err)
		return
	}
	defer db.Close()

	engine := NewEngine(db.DB)

	t.Run("BasicQuery", func(t *testing.T) {
		params := QueryParams{
			Limit: 10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if result.Total == 0 {
			t.Skip("No photos in test database")
		}

		if len(result.Photos) == 0 {
			t.Error("Expected photos in result")
		}

		t.Logf("Found %d total photos, returned %d", result.Total, len(result.Photos))
	})

	t.Run("FilterByYear", func(t *testing.T) {
		year := 2025
		params := QueryParams{
			Year:  &year,
			Limit: 10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos from year %d", result.Total, year)

		for _, photo := range result.Photos {
			if !photo.DateTaken.IsZero() && photo.DateTaken.Year() != year {
				t.Errorf("Expected year %d, got %d for photo %d", year, photo.DateTaken.Year(), photo.ID)
			}
		}
	})

	t.Run("FilterByISO", func(t *testing.T) {
		isoMin := 100
		isoMax := 400
		params := QueryParams{
			ISOMin: &isoMin,
			ISOMax: &isoMax,
			Limit:  10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos with ISO between %d and %d", result.Total, isoMin, isoMax)

		for _, photo := range result.Photos {
			if photo.ISO != 0 && (photo.ISO < isoMin || photo.ISO > isoMax) {
				t.Errorf("Expected ISO between %d and %d, got %d", isoMin, isoMax, photo.ISO)
			}
		}
	})

	t.Run("FilterByTimeOfDay", func(t *testing.T) {
		params := QueryParams{
			TimeOfDay: []string{"morning"},
			Limit:     10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d morning photos", result.Total)

		for _, photo := range result.Photos {
			if photo.TimeOfDay != "morning" && photo.TimeOfDay != "" {
				t.Errorf("Expected morning photo, got %s", photo.TimeOfDay)
			}
		}
	})

	t.Run("FilterByColor", func(t *testing.T) {
		params := QueryParams{
			ColourName: []string{"blue"},
			Limit:      10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos with blue colors", result.Total)
	})

	t.Run("MultipleFilters", func(t *testing.T) {
		year := 2025
		isoMin := 100
		params := QueryParams{
			Year:      &year,
			ISOMin:    &isoMin,
			TimeOfDay: []string{"morning", "afternoon"},
			Limit:     10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos matching multiple filters", result.Total)

		for _, photo := range result.Photos {
			if !photo.DateTaken.IsZero() && photo.DateTaken.Year() != year {
				t.Errorf("Expected year %d, got %d for photo %d", year, photo.DateTaken.Year(), photo.ID)
			}
			if photo.ISO != 0 && photo.ISO < isoMin {
				t.Errorf("Expected ISO >= %d, got %d", isoMin, photo.ISO)
			}
		}
	})

	t.Run("FilterByBurst", func(t *testing.T) {
		inBurst := true
		params := QueryParams{
			InBurst: &inBurst,
			Limit:   10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos in bursts", result.Total)

		for _, photo := range result.Photos {
			if !photo.InBurst {
				t.Error("Expected photo to be in burst")
			}
		}
	})

	t.Run("RangeQueries", func(t *testing.T) {
		apertureMin := 2.8
		apertureMax := 5.6
		focalMin := 24.0
		focalMax := 70.0

		params := QueryParams{
			ApertureMin:    &apertureMin,
			ApertureMax:    &apertureMax,
			FocalLengthMin: &focalMin,
			FocalLengthMax: &focalMax,
			Limit:          10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Found %d photos with aperture f/%.1f-%.1f and focal length %.0f-%.0fmm",
			result.Total, apertureMin, apertureMax, focalMin, focalMax)

		for _, photo := range result.Photos {
			if photo.Aperture != 0 && (photo.Aperture < apertureMin || photo.Aperture > apertureMax) {
				t.Errorf("Aperture %.1f outside range %.1f-%.1f", photo.Aperture, apertureMin, apertureMax)
			}
			if photo.FocalLength != 0 && (photo.FocalLength < focalMin || photo.FocalLength > focalMax) {
				t.Errorf("Focal length %.0f outside range %.0f-%.0f", photo.FocalLength, focalMin, focalMax)
			}
		}
	})

	t.Run("Sorting", func(t *testing.T) {
		params := QueryParams{
			SortBy:    "iso",
			SortOrder: "asc",
			Limit:     10,
		}
		result, err := engine.Query(params)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(result.Photos) > 1 {
			for i := 0; i < len(result.Photos)-1; i++ {
				if result.Photos[i].ISO > result.Photos[i+1].ISO && result.Photos[i+1].ISO != 0 {
					t.Errorf("Photos not sorted by ISO ascending: %d > %d",
						result.Photos[i].ISO, result.Photos[i+1].ISO)
				}
			}
		}

		t.Logf("Sorted %d photos by ISO (ascending)", len(result.Photos))
	})

	t.Run("Pagination", func(t *testing.T) {
		// Get first page
		params1 := QueryParams{
			Limit:  5,
			Offset: 0,
		}
		result1, err := engine.Query(params1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Get second page
		params2 := QueryParams{
			Limit:  5,
			Offset: 5,
		}
		result2, err := engine.Query(params2)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if result1.Total != result2.Total {
			t.Error("Total should be same across pages")
		}

		if result1.Total > 5 && !result1.HasMore {
			t.Error("Should have more results")
		}

		t.Logf("Page 1: %d photos, Page 2: %d photos, Total: %d",
			len(result1.Photos), len(result2.Photos), result1.Total)
	})
}

func TestFacets(t *testing.T) {
	db, err := database.Open("../../test_query.db")
	if err != nil {
		t.Skipf("Test database not found: %v", err)
		return
	}
	defer db.Close()

	engine := NewEngine(db.DB)

	t.Run("ComputeAllFacets", func(t *testing.T) {
		params := QueryParams{
			Limit: 10,
		}
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("Failed to compute facets: %v", err)
		}

		if facets == nil {
			t.Fatal("Facets should not be nil")
		}

		// Log facet values
		if facets.Camera != nil && len(facets.Camera.Values) > 0 {
			t.Logf("Camera facet has %d values:", len(facets.Camera.Values))
			for _, v := range facets.Camera.Values[:min(3, len(facets.Camera.Values))] {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}

		if facets.TimeOfDay != nil && len(facets.TimeOfDay.Values) > 0 {
			t.Logf("Time of day facet has %d values:", len(facets.TimeOfDay.Values))
			for _, v := range facets.TimeOfDay.Values {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}

		if facets.Year != nil && len(facets.Year.Values) > 0 {
			t.Logf("Year facet has %d values:", len(facets.Year.Values))
			for _, v := range facets.Year.Values {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}

		if facets.Season != nil && len(facets.Season.Values) > 0 {
			t.Logf("Season facet has %d values:", len(facets.Season.Values))
			for _, v := range facets.Season.Values {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}

		if facets.ColourName != nil && len(facets.ColourName.Values) > 0 {
			t.Logf("Color facet has %d values:", len(facets.ColourName.Values))
			for _, v := range facets.ColourName.Values {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}
	})

	t.Run("FacetsWithFilter", func(t *testing.T) {
		year := 2025
		params := QueryParams{
			Year:  &year,
			Limit: 10,
		}
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("Failed to compute facets: %v", err)
		}

		// Year facet should still show all years, but selected year should be marked
		if facets.Year != nil {
			for _, v := range facets.Year.Values {
				if v.Value == "2025" && !v.Selected {
					t.Error("Year 2025 should be marked as selected")
				}
				t.Logf("Year %s: %d photos (selected: %v)", v.Value, v.Count, v.Selected)
			}
		}

		// Camera facet should reflect filter
		if facets.Camera != nil && len(facets.Camera.Values) > 0 {
			t.Logf("Camera facet (filtered by year %d):", year)
			for _, v := range facets.Camera.Values[:min(3, len(facets.Camera.Values))] {
				t.Logf("  %s: %d", v.Label, v.Count)
			}
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestColorSearch(t *testing.T) {
	db, err := database.Open("../../test_query.db")
	if err != nil {
		t.Skipf("Test database not found: %v", err)
		return
	}
	defer db.Close()

	engine := NewEngine(db.DB)

	colorNames := []string{"red", "orange", "yellow", "green", "blue", "purple", "pink"}

	for _, color := range colorNames {
		t.Run("Color_"+color, func(t *testing.T) {
			params := QueryParams{
				ColourName: []string{color},
				Limit:      5,
			}
			result, err := engine.Query(params)
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			t.Logf("Found %d photos with %s colors", result.Total, color)
		})
	}
}

func BenchmarkQuery(b *testing.B) {
	db, err := database.Open("../../test_query.db")
	if err != nil {
		b.Skip("Test database not found")
		return
	}
	defer db.Close()

	engine := NewEngine(db.DB)

	b.Run("SimpleQuery", func(b *testing.B) {
		params := QueryParams{
			Limit: 50,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Query(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexQuery", func(b *testing.B) {
		year := 2025
		isoMin := 100
		isoMax := 1600
		params := QueryParams{
			Year:      &year,
			ISOMin:    &isoMin,
			ISOMax:    &isoMax,
			TimeOfDay: []string{"morning", "afternoon"},
			Limit:     50,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Query(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ColorQuery", func(b *testing.B) {
		params := QueryParams{
			ColourName: []string{"blue", "green"},
			Limit:      50,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Query(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComputeFacets", func(b *testing.B) {
		params := QueryParams{
			Limit: 50,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.ComputeFacets(params)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
