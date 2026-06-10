package query

import (
	"database/sql"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/testsupport"
)

// TestFacetClickInvariant is a property-based test of the state machine's
// fundamental rule: every facet value displayed with count N must, when its
// URL is followed, produce a result set of exactly N photos.
//
// The historic camera facet bug (facet said "Leica Camera AG LEICA M11
// Monochrom (205)", clicking returned 0 photos) was a violation of exactly
// this invariant; this test enforces it for every facet dimension over
// randomly generated photo corpora and filter states, so any future facet
// type or WHERE-clause change that breaks count/filter agreement fails here.
func TestFacetClickInvariant(t *testing.T) {
	for seed := int64(1); seed <= 5; seed++ {
		t.Run(seedName(seed), func(t *testing.T) {
			db := setupTestDBWithSchema(t)
			defer db.Close()

			rng := rand.New(rand.NewSource(seed))
			insertRandomCorpus(t, db, rng, 40)

			engine := NewEngine(db)
			mapper := NewURLMapper()

			// Start from the empty state plus a few random filter states
			// reached by clicking facet values (i.e., states a user can reach).
			states := []QueryParams{{Limit: DefaultLimit}}
			for i := 0; i < 4; i++ {
				if next, ok := randomReachableState(t, engine, mapper, rng, states[len(states)-1]); ok {
					states = append(states, next)
				}
			}

			for _, state := range states {
				verifyFacetClickCounts(t, engine, mapper, state)
			}
		})
	}
}

func seedName(seed int64) string {
	return "seed_" + strings.Repeat("i", int(seed)) // stable, readable subtest names
}

// verifyFacetClickCounts computes facets for the state and asserts that
// following any value's URL yields exactly the advertised count.
func verifyFacetClickCounts(t *testing.T, engine *Engine, mapper *URLMapper, state QueryParams) {
	t.Helper()

	facets, err := engine.ComputeFacets(state)
	if err != nil {
		t.Fatalf("ComputeFacets(%+v) failed: %v", state, err)
	}

	for _, facet := range allFacets(facets) {
		if facet == nil {
			continue
		}
		for _, value := range facet.Values {
			if value.URL == "" {
				continue
			}
			params, err := mapper.ParsePath(splitURL(value.URL))
			if err != nil {
				t.Errorf("facet %s value %q: URL %q does not parse: %v", facet.Name, value.Value, value.URL, err)
				continue
			}
			result, err := engine.Query(params)
			if err != nil {
				t.Errorf("facet %s value %q: query for %q failed: %v", facet.Name, value.Value, value.URL, err)
				continue
			}

			// Selected values toggle the filter OFF, so their count (results
			// with the filter applied) intentionally differs from the URL's
			// results; the invariant applies to unselected values.
			if value.Selected {
				continue
			}
			if result.Total != value.Count {
				t.Errorf("INVARIANT VIOLATED: facet %s value %q shows count=%d but clicking %q returns %d photos (state %+v)",
					facet.Name, value.Value, value.Count, value.URL, result.Total, state)
			}
		}
	}
}

// randomReachableState clicks a random clickable facet value from the current
// state and returns the resulting state.
func randomReachableState(t *testing.T, engine *Engine, mapper *URLMapper, rng *rand.Rand, from QueryParams) (QueryParams, bool) {
	t.Helper()

	facets, err := engine.ComputeFacets(from)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}
	var clickable []string
	for _, facet := range allFacets(facets) {
		if facet == nil {
			continue
		}
		for _, v := range facet.Values {
			if v.URL != "" && !v.Selected && v.Count > 0 {
				clickable = append(clickable, v.URL)
			}
		}
	}
	if len(clickable) == 0 {
		return QueryParams{}, false
	}
	params, err := mapper.ParsePath(splitURL(clickable[rng.Intn(len(clickable))]))
	if err != nil {
		t.Fatalf("ParsePath failed: %v", err)
	}
	return params, true
}

func allFacets(fc *FacetCollection) []*Facet {
	return []*Facet{
		fc.Camera, fc.Lens, fc.Year, fc.Month, fc.TimeOfDay, fc.Season,
		fc.FocalCategory, fc.ShootingCondition, fc.InBurst, fc.ColourName,
	}
}

// insertRandomCorpus fills the database with n photos drawn from realistic
// vocabularies (including the multi-word camera makes from past bugs and a
// slice of photos with no capture date).
func insertRandomCorpus(t *testing.T, db *sql.DB, rng *rand.Rand, n int) {
	t.Helper()

	cameras := [][2]string{
		{"Canon", "EOS R5"},
		{"Leica Camera AG", "LEICA M11 Monochrom"},
		{"Phase One A/S", "IQ4 150MP"},
		{"Sony", "Alpha 7R V"},
	}
	lenses := []string{"RF24-70mm F2.8 L IS USM", "Summilux-M 1:1.4/50 ASPH.", ""}
	colours := []string{"red", "green", "blue", "gray", "black", "white"}
	tods := []string{"morning", "midday", "evening", "night"}
	seasons := []string{"spring", "summer", "fall", "winter"}
	focalCats := []string{"wide", "normal", "telephoto"}

	for i := 0; i < n; i++ {
		cam := cameras[rng.Intn(len(cameras))]
		b := testsupport.NewPhoto().
			WithCamera(cam[0], cam[1]).
			WithLens(lenses[rng.Intn(len(lenses))]).
			WithTimeOfDay(tods[rng.Intn(len(tods))]).
			WithSeason(seasons[rng.Intn(len(seasons))]).
			WithFocalCategory(focalCats[rng.Intn(len(focalCats))]).
			WithColourName(colours[rng.Intn(len(colours))])

		if rng.Intn(10) == 0 {
			b = b.WithoutDate() // some photos have unknown dates
		} else {
			b = b.WithDateTaken(time.Date(
				2020+rng.Intn(5), time.Month(1+rng.Intn(12)), 1+rng.Intn(28),
				rng.Intn(24), 0, 0, 0, time.UTC))
		}

		testsupport.InsertPhotoSQL(t, db, b.Build())
	}
}

// TestYearFacetMatchesGoRecount is a differential test: the SQL GROUP BY that
// computes the year facet is checked against a naive Go-side recount over the
// same rows. Two independent implementations must agree.
func TestYearFacetMatchesGoRecount(t *testing.T) {
	db := setupTestDBWithSchema(t)
	defer db.Close()

	rng := rand.New(rand.NewSource(7))
	insertRandomCorpus(t, db, rng, 60)

	engine := NewEngine(db)
	facets, err := engine.ComputeFacets(QueryParams{Limit: DefaultLimit})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Naive oracle: read every row and count years in Go.
	rows, err := db.Query(`SELECT COALESCE(strftime('%Y', date_taken), 'unknown') FROM photos`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()
	want := map[string]int{}
	for rows.Next() {
		var year string
		if err := rows.Scan(&year); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		want[year]++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	got := map[string]int{}
	for _, v := range facets.Year.Values {
		got[v.Value] = v.Count
	}

	if len(got) != len(want) {
		t.Errorf("year facet has %d values, recount has %d (facet=%v recount=%v)", len(got), len(want), got, want)
	}
	for year, count := range want {
		if got[year] != count {
			t.Errorf("year %s: facet says %d, Go recount says %d", year, got[year], count)
		}
	}
}

// TestChromaticHueRangesCoverCircle asserts the chromatic colour ranges tile
// the full hue circle with no gaps, so every hue classifies to at least one
// colour name (achromatic names cover the rest via saturation/lightness).
func TestChromaticHueRangesCoverCircle(t *testing.T) {
	t.Parallel()

	chromatic := []string{"red", "orange", "yellow", "green", "blue", "purple", "pink"}
	for hue := 0; hue <= 360; hue++ {
		covered := false
		for _, name := range chromatic {
			r := ColourNameToHueRange[name]
			if name == "red" {
				if (hue >= 0 && hue <= 15) || (hue >= 345 && hue <= 360) {
					covered = true
					break
				}
				continue
			}
			if hue >= r[0] && hue <= r[1] {
				covered = true
				break
			}
		}
		if !covered {
			t.Errorf("hue %d is not covered by any chromatic colour range", hue)
		}
	}
}
