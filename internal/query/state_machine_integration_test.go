package query

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/pkg/models"
)

// TestStateMachineIntegration validates the temporal dimensions of the state
// machine model against a real SQLite database (see
// specs/facet_state_machine.spec): Year, Month and Day are independent
// filters, removing one preserves the others, and a facet's count is the
// number of photos its click returns.
//
// Every expected count below is derived by hand from integrationPhotoDates.
func TestStateMachineIntegration(t *testing.T) {
	engine := NewEngine(setupIntegrationTestDB(t))

	t.Run("TemporalFilters", func(t *testing.T) {
		testTemporalFilters(t, engine)
	})

	t.Run("FilterRemoval", func(t *testing.T) {
		testFilterRemoval(t, engine)
	})

	t.Run("YearFacetPreservesMonth", func(t *testing.T) {
		testYearFacetPreservesMonth(t, engine)
	})

	t.Run("MonthFacetPreservesYear", func(t *testing.T) {
		testMonthFacetPreservesYear(t, engine)
	})
}

// integrationPhotoDates is the fixture. Single-digit months and days (January,
// the 1st) are included because SQLite's strftime zero-pads them, so a filter
// that sends "1" instead of "01" matches nothing.
var integrationPhotoDates = []time.Time{
	// October across three years
	time.Date(2020, 10, 1, 12, 0, 0, 0, time.UTC),
	time.Date(2020, 10, 15, 12, 0, 0, 0, time.UTC),
	time.Date(2021, 10, 1, 12, 0, 0, 0, time.UTC),
	time.Date(2021, 10, 15, 12, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC),
	// The 15th of other months
	time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC),
	time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
	// Other dates
	time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC),
}

// testTemporalFilters checks that each temporal filter applies on its own and
// in combination, including Month and Day without Year (the state the Year
// facet is computed from).
func testTemporalFilters(t *testing.T, engine *Engine) {
	tests := []struct {
		name   string
		params QueryParams
		want   int
	}{
		{"YearOnly", QueryParams{Year: intPtr(2024)}, 7},
		{"MonthOnly_October", QueryParams{Month: intPtr(10)}, 6},
		{"MonthOnly_January", QueryParams{Month: intPtr(1)}, 2},
		{"DayOnly_15th", QueryParams{Day: intPtr(15)}, 6},
		{"DayOnly_1st", QueryParams{Day: intPtr(1)}, 4},
		{"MonthAndDay", QueryParams{Month: intPtr(10), Day: intPtr(15)}, 3},
		{"YearAndMonth", QueryParams{Year: intPtr(2024), Month: intPtr(10)}, 2},
		{"YearAndMonth_January", QueryParams{Year: intPtr(2024), Month: intPtr(1)}, 2},
		{"YearMonthDay", QueryParams{Year: intPtr(2024), Month: intPtr(10), Day: intPtr(15)}, 1},
		{"YearAndDay", QueryParams{Year: intPtr(2021), Day: intPtr(1)}, 1},
		{"NoMatch", QueryParams{Year: intPtr(2021), Month: intPtr(1)}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.params.Limit = 100
			result, err := engine.Query(tt.params)
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}
			if result.Total != tt.want {
				t.Errorf("Total = %d, want %d", result.Total, tt.want)
			}
			if len(result.Photos) != tt.want {
				t.Errorf("returned %d photos, want %d", len(result.Photos), tt.want)
			}
		})
	}
}

// testFilterRemoval verifies that removing Year keeps the Month filter:
// year=2020&month=10 (2 photos) becomes "every October" (6 photos), not
// "everything" (11).
func testFilterRemoval(t *testing.T, engine *Engine) {
	start, err := engine.Query(QueryParams{Year: intPtr(2020), Month: intPtr(10), Limit: 100})
	if err != nil {
		t.Fatalf("Initial query failed: %v", err)
	}
	if start.Total != 2 {
		t.Fatalf("year=2020&month=10: Total = %d, want 2", start.Total)
	}

	monthOnly, err := engine.Query(QueryParams{Month: intPtr(10), Limit: 100})
	if err != nil {
		t.Fatalf("Month-only query failed: %v", err)
	}
	if monthOnly.Total != 6 {
		t.Errorf("after removing year: Total = %d, want 6 (every October)", monthOnly.Total)
	}
	for _, p := range monthOnly.Photos {
		if p.DateTaken.Month() != time.October {
			t.Errorf("photo %d taken %s is not from October", p.ID, p.DateTaken.Format("2006-01-02"))
		}
	}
}

// testYearFacetPreservesMonth verifies that, with month=10 selected, the Year
// facet counts October photos per year, and that clicking each year returns
// exactly that count.
func testYearFacetPreservesMonth(t *testing.T, engine *Engine) {
	params := QueryParams{Month: intPtr(10), Limit: 100}
	facets, err := engine.ComputeFacets(params)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	want := map[string]int{"2020": 2, "2021": 2, "2024": 2}
	if got := facetCounts(facets.Year); !equalCounts(got, want) {
		t.Fatalf("Year facet with month=10 = %v, want %v", got, want)
	}

	for _, fv := range facets.Year.Values {
		year, err := strconv.Atoi(fv.Value)
		if err != nil {
			t.Fatalf("Year facet value %q is not a year: %v", fv.Value, err)
		}
		clicked, err := engine.Query(QueryParams{Year: &year, Month: intPtr(10), Limit: 100})
		if err != nil {
			t.Fatalf("Query for year %d failed: %v", year, err)
		}
		if clicked.Total != fv.Count {
			t.Errorf("Year %s shows count %d but clicking it returns %d photos", fv.Value, fv.Count, clicked.Total)
		}
	}
}

// testMonthFacetPreservesYear verifies that, with year=2024 selected, the Month
// facet counts only 2024 photos.
func testMonthFacetPreservesYear(t *testing.T, engine *Engine) {
	facets, err := engine.ComputeFacets(QueryParams{Year: intPtr(2024), Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	want := map[string]int{"01": 2, "02": 1, "03": 1, "10": 2, "12": 1}
	if got := facetCounts(facets.Month); !equalCounts(got, want) {
		t.Errorf("Month facet with year=2024 = %v, want %v", got, want)
	}
}

// facetCounts returns value -> count for the non-zero values of a facet.
func facetCounts(f *Facet) map[string]int {
	got := map[string]int{}
	if f == nil {
		return got
	}
	for _, v := range f.Values {
		if v.Count > 0 {
			got[v.Value] = v.Count
		}
	}
	return got
}

func equalCounts(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// setupIntegrationTestDB returns a database in t.TempDir() holding one photo
// per integrationPhotoDates entry, inserted through the production insert path.
func setupIntegrationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "integration.db"))
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	for _, date := range integrationPhotoDates {
		meta := &models.PhotoMetadata{
			FilePath:  "/test/" + date.Format("2006_01_02") + ".jpg",
			DateTaken: date,
			Width:     1920,
			Height:    1080,
		}
		if err := db.InsertPhoto(meta); err != nil {
			t.Fatalf("Failed to insert test photo: %v", err)
		}
	}

	return db.DB
}

// intPtr is defined in url_mapper_test.go.
