package query

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// TestFacetStateMachine tests all possible state transitions in faceted navigation
// Treats facet application as a state machine where each action transitions to a new state
func TestFacetStateMachine(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	engine := NewEngine(db)
	mapper := NewURLMapper()

	// Each state family must exercise at least one transition. Several of the
	// helpers only test transitions for facet values they find, so without
	// this check an empty or thin fixture made them pass vacuously.
	run := func(name string, fn func(*testing.T, *Engine, *URLMapper)) {
		t.Run(name, func(t *testing.T) {
			before := facetTransitionsChecked.Load()
			fn(t, engine, mapper)
			if facetTransitionsChecked.Load() == before {
				t.Fatal("no facet transition was exercised; facetFixturePhotos no longer covers this state family")
			}
		})
	}

	run("EmptyToSingleFacet", testEmptyToSingleFacetTransitions) // from empty state
	run("SingleToDualFacet", testSingleToDualFacetTransitions)   // add a second facet
	run("DualToSingleFacet", testDualToSingleFacetTransitions)   // remove one of two
	run("DualToTripleFacet", testDualToTripleFacetTransitions)   // add a third facet
	run("TripleToDualFacet", testTripleToDualFacetTransitions)   // remove one of three
	run("ReplaceSameFacetType", testReplaceSameFacetType)        // swap within a facet type
	run("AnyToEmpty", testAnyToEmptyTransitions)                 // remove all facets
}

// testEmptyToSingleFacetTransitions tests transitions from no filters to one filter
func testEmptyToSingleFacetTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Start with no filters
	baseParams := QueryParams{Limit: 100}
	baseResult, err := engine.Query(baseParams)
	if err != nil {
		t.Fatalf("Base query failed: %v", err)
	}

	if baseResult.Total == 0 {
		t.Fatal("facetFixturePhotos no longer covers this state: No photos in database")
	}

	totalPhotos := baseResult.Total
	t.Logf("Base state: %d photos", totalPhotos)

	// Get facets for base state
	facets, err := engine.ComputeFacets(baseParams)
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Test adding each color facet
	if facets.ColourName != nil {
		for _, fv := range facets.ColourName.Values {
			testFacetTransition(t, engine, mapper, "EMPTY", "Color:"+fv.Value,
				QueryParams{Limit: 100}, fv.URL, fv.Count)
		}
	}

	// Test adding each year facet
	if facets.Year != nil {
		for _, fv := range facets.Year.Values {
			testFacetTransition(t, engine, mapper, "EMPTY", "Year:"+fv.Value,
				QueryParams{Limit: 100}, fv.URL, fv.Count)
		}
	}

	// Test adding each camera facet
	if facets.Camera != nil {
		for _, fv := range facets.Camera.Values {
			testFacetTransition(t, engine, mapper, "EMPTY", "Camera:"+fv.Value,
				QueryParams{Limit: 100}, fv.URL, fv.Count)
		}
	}

	// Test adding each time of day facet
	if facets.TimeOfDay != nil {
		for _, fv := range facets.TimeOfDay.Values {
			testFacetTransition(t, engine, mapper, "EMPTY", "TimeOfDay:"+fv.Value,
				QueryParams{Limit: 100}, fv.URL, fv.Count)
		}
	}
}

// testSingleToDualFacetTransitions tests adding a second facet
func testSingleToDualFacetTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Get initial facets
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Pick first available color
	if baseFacets.ColourName != nil && len(baseFacets.ColourName.Values) > 0 {
		colorFacet := baseFacets.ColourName.Values[0]
		colorParams := QueryParams{
			ColourName: []string{colorFacet.Value},
			Limit:      100,
		}

		// Get facets with color applied
		colorFacets, err := engine.ComputeFacets(colorParams)
		if err != nil {
			t.Fatalf("ComputeFacets with color failed: %v", err)
		}

		// Try adding year to color
		if colorFacets.Year != nil && len(colorFacets.Year.Values) > 0 {
			for _, yv := range colorFacets.Year.Values {
				if yv.Count > 0 {
					testFacetTransition(t, engine, mapper,
						"Color:"+colorFacet.Value, "Color:"+colorFacet.Value+"+Year:"+yv.Value,
						colorParams, yv.URL, yv.Count)
					break // Test one transition
				}
			}
		}

		// Try adding camera to color
		if colorFacets.Camera != nil && len(colorFacets.Camera.Values) > 0 {
			for _, cv := range colorFacets.Camera.Values {
				if cv.Count > 0 {
					testFacetTransition(t, engine, mapper,
						"Color:"+colorFacet.Value, "Color:"+colorFacet.Value+"+Camera:"+cv.Value,
						colorParams, cv.URL, cv.Count)
					break // Test one transition
				}
			}
		}

		// Try adding time of day to color
		if colorFacets.TimeOfDay != nil && len(colorFacets.TimeOfDay.Values) > 0 {
			for _, tv := range colorFacets.TimeOfDay.Values {
				if tv.Count > 0 {
					testFacetTransition(t, engine, mapper,
						"Color:"+colorFacet.Value, "Color:"+colorFacet.Value+"+TimeOfDay:"+tv.Value,
						colorParams, tv.URL, tv.Count)
					break // Test one transition
				}
			}
		}
	}
}

// testDualToSingleFacetTransitions tests removing one facet from dual state
func testDualToSingleFacetTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Get initial facets
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Pick first available color
	if baseFacets.ColourName != nil && len(baseFacets.ColourName.Values) > 0 {
		colorFacet := baseFacets.ColourName.Values[0]

		// Get facets with color applied
		colorParams := QueryParams{
			ColourName: []string{colorFacet.Value},
			Limit:      100,
		}
		colorFacets, err := engine.ComputeFacets(colorParams)
		if err != nil {
			t.Fatalf("ComputeFacets with color failed: %v", err)
		}

		// Find a year with photos
		if colorFacets.Year != nil && len(colorFacets.Year.Values) > 0 {
			for _, yv := range colorFacets.Year.Values {
				if yv.Count > 0 {
					// Parse year
					var year int
					if _, err := fmt.Sscanf(yv.Value, "%d", &year); err == nil {
						// Apply both color and year
						dualParams := QueryParams{
							ColourName: []string{colorFacet.Value},
							Year:       &year,
							Limit:      100,
						}

						// Get facets with both applied
						dualFacets, err := engine.ComputeFacets(dualParams)
						if err != nil {
							t.Fatalf("ComputeFacets with dual failed: %v", err)
						}

						// Test removing color (should keep year). The target
						// state is Year alone, so the expected count is that
						// year's count with no colour filter (yv.Count is the
						// colour-filtered count and only matched when every
						// photo that year had the colour).
						yearOnlyCount := -1
						if baseFacets.Year != nil {
							for _, byv := range baseFacets.Year.Values {
								if byv.Value == yv.Value {
									yearOnlyCount = byv.Count
								}
							}
						}
						if yearOnlyCount < 0 {
							t.Fatalf("year %s is in the colour-filtered year facet but not the unfiltered one", yv.Value)
						}
						if dualFacets.ColourName != nil {
							for _, cv := range dualFacets.ColourName.Values {
								if cv.Value == colorFacet.Value && cv.Selected {
									testFacetTransition(t, engine, mapper,
										"Color:"+colorFacet.Value+"+Year:"+yv.Value,
										"Year:"+yv.Value,
										dualParams, cv.URL, yearOnlyCount)
									break
								}
							}
						}

						// Test removing year (should keep color)
						if dualFacets.Year != nil {
							for _, yv2 := range dualFacets.Year.Values {
								if yv2.Value == yv.Value && yv2.Selected {
									testFacetTransition(t, engine, mapper,
										"Color:"+colorFacet.Value+"+Year:"+yv.Value,
										"Color:"+colorFacet.Value,
										dualParams, yv2.URL, colorFacet.Count)
									break
								}
							}
						}

						break // Tested one dual combination
					}
				}
			}
		}
	}
}

// testDualToTripleFacetTransitions tests adding a third facet
func testDualToTripleFacetTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Get initial facets
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Pick first available color
	if baseFacets.ColourName != nil && len(baseFacets.ColourName.Values) > 0 {
		colorFacet := baseFacets.ColourName.Values[0]

		// Get facets with color applied
		colorParams := QueryParams{
			ColourName: []string{colorFacet.Value},
			Limit:      100,
		}
		colorFacets, err := engine.ComputeFacets(colorParams)
		if err != nil {
			t.Fatalf("ComputeFacets with color failed: %v", err)
		}

		// Find a year with photos
		if colorFacets.Year != nil && len(colorFacets.Year.Values) > 0 {
			for _, yv := range colorFacets.Year.Values {
				if yv.Count > 0 {
					// Parse year
					var year int
					if _, err := fmt.Sscanf(yv.Value, "%d", &year); err == nil {
						// Apply both color and year
						dualParams := QueryParams{
							ColourName: []string{colorFacet.Value},
							Year:       &year,
							Limit:      100,
						}

						// Get facets with both applied
						dualFacets, err := engine.ComputeFacets(dualParams)
						if err != nil {
							t.Fatalf("ComputeFacets with dual failed: %v", err)
						}

						// Try adding camera as third facet
						if dualFacets.Camera != nil && len(dualFacets.Camera.Values) > 0 {
							for _, cv := range dualFacets.Camera.Values {
								if cv.Count > 0 {
									testFacetTransition(t, engine, mapper,
										"Color:"+colorFacet.Value+"+Year:"+yv.Value,
										"Color:"+colorFacet.Value+"+Year:"+yv.Value+"+Camera:"+cv.Value,
										dualParams, cv.URL, cv.Count)
									break
								}
							}
						}

						break // Tested one triple combination
					}
				}
			}
		}
	}
}

// testTripleToDualFacetTransitions tests removing one facet from triple state
func testTripleToDualFacetTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Build triple state: Color + Year + Camera
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	if baseFacets.ColourName == nil || len(baseFacets.ColourName.Values) == 0 {
		t.Fatal("facetFixturePhotos no longer covers this state: No color facets")
	}

	colorFacet := baseFacets.ColourName.Values[0]
	colorParams := QueryParams{ColourName: []string{colorFacet.Value}, Limit: 100}
	colorFacets, err := engine.ComputeFacets(colorParams)
	if err != nil {
		t.Fatalf("ComputeFacets with color failed: %v", err)
	}

	if colorFacets.Year == nil || len(colorFacets.Year.Values) == 0 {
		t.Fatal("facetFixturePhotos no longer covers this state: No year facets with color")
	}

	var year int
	var yearValue string
	for _, yv := range colorFacets.Year.Values {
		if yv.Count > 0 {
			fmt.Sscanf(yv.Value, "%d", &year)
			yearValue = yv.Value
			break
		}
	}

	if year == 0 {
		t.Fatal("facetFixturePhotos no longer covers this state: No valid year facets")
	}

	dualParams := QueryParams{ColourName: []string{colorFacet.Value}, Year: &year, Limit: 100}
	dualFacets, err := engine.ComputeFacets(dualParams)
	if err != nil {
		t.Fatalf("ComputeFacets with dual failed: %v", err)
	}

	if dualFacets.Camera == nil || len(dualFacets.Camera.Values) == 0 {
		t.Fatal("facetFixturePhotos no longer covers this state: No camera facets with color+year")
	}

	var cameraValue string
	for _, cv := range dualFacets.Camera.Values {
		if cv.Count > 0 {
			cameraValue = cv.Value
			parts := strings.SplitN(cv.Value, " ", 2)
			if len(parts) == 2 {
				// Apply all three filters
				tripleParams := QueryParams{
					ColourName:  []string{colorFacet.Value},
					Year:        &year,
					CameraMake:  []string{parts[0]},
					CameraModel: []string{parts[1]},
					Limit:       100,
				}

				// Get facets with all three applied
				tripleFacets, err := engine.ComputeFacets(tripleParams)
				if err != nil {
					t.Fatalf("ComputeFacets with triple failed: %v", err)
				}

				// Test removing each of the three facets
				// Remove color -> keep year+camera
				if tripleFacets.ColourName != nil {
					for _, cfv := range tripleFacets.ColourName.Values {
						if cfv.Value == colorFacet.Value && cfv.Selected {
							expectedCount := cv.Count // Camera count from dual state
							testFacetTransition(t, engine, mapper,
								"Color:"+colorFacet.Value+"+Year:"+yearValue+"+Camera:"+cameraValue,
								"Year:"+yearValue+"+Camera:"+cameraValue,
								tripleParams, cfv.URL, expectedCount)
							break
						}
					}
				}

				// Remove year -> keep color+camera
				if tripleFacets.Year != nil {
					for _, yfv := range tripleFacets.Year.Values {
						if yfv.Value == yearValue && yfv.Selected {
							// Need to query for color+camera without year to get expected count
							testParams := QueryParams{
								ColourName:  []string{colorFacet.Value},
								CameraMake:  []string{parts[0]},
								CameraModel: []string{parts[1]},
								Limit:       100,
							}
							testResult, _ := engine.Query(testParams)
							testFacetTransition(t, engine, mapper,
								"Color:"+colorFacet.Value+"+Year:"+yearValue+"+Camera:"+cameraValue,
								"Color:"+colorFacet.Value+"+Camera:"+cameraValue,
								tripleParams, yfv.URL, testResult.Total)
							break
						}
					}
				}

				// Remove camera -> keep color+year
				if tripleFacets.Camera != nil {
					for _, camfv := range tripleFacets.Camera.Values {
						if camfv.Value == cameraValue && camfv.Selected {
							// Query dual state count
							dualResult, _ := engine.Query(dualParams)
							testFacetTransition(t, engine, mapper,
								"Color:"+colorFacet.Value+"+Year:"+yearValue+"+Camera:"+cameraValue,
								"Color:"+colorFacet.Value+"+Year:"+yearValue,
								tripleParams, camfv.URL, dualResult.Total)
							break
						}
					}
				}

				break // Tested one triple combination
			}
		}
	}
}

// testReplaceSameFacetType tests replacing a facet value with another value of same type
func testReplaceSameFacetType(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Get initial facets
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Test replacing color with another color
	if baseFacets.ColourName != nil && len(baseFacets.ColourName.Values) >= 2 {
		color1 := baseFacets.ColourName.Values[0]
		color2 := baseFacets.ColourName.Values[1]

		params1 := QueryParams{ColourName: []string{color1.Value}, Limit: 100}
		facets1, err := engine.ComputeFacets(params1)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		// Find color2 in the facets (it should not be selected)
		if facets1.ColourName != nil {
			for _, cv := range facets1.ColourName.Values {
				if cv.Value == color2.Value && !cv.Selected {
					testFacetTransition(t, engine, mapper,
						"Color:"+color1.Value, "Color:"+color2.Value,
						params1, cv.URL, color2.Count)
					break
				}
			}
		}
	}

	// Test replacing year with another year
	if baseFacets.Year != nil && len(baseFacets.Year.Values) >= 2 {
		year1 := baseFacets.Year.Values[0]
		year2 := baseFacets.Year.Values[1]

		var y1 int
		fmt.Sscanf(year1.Value, "%d", &y1)
		params1 := QueryParams{Year: &y1, Limit: 100}
		facets1, err := engine.ComputeFacets(params1)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		// Find year2 in the facets
		if facets1.Year != nil {
			for _, yv := range facets1.Year.Values {
				if yv.Value == year2.Value && !yv.Selected {
					testFacetTransition(t, engine, mapper,
						"Year:"+year1.Value, "Year:"+year2.Value,
						params1, yv.URL, year2.Count)
					break
				}
			}
		}
	}
}

// testAnyToEmptyTransitions tests removing all facets
func testAnyToEmptyTransitions(t *testing.T, engine *Engine, mapper *URLMapper) {
	t.Helper()

	// Get base count
	baseResult, err := engine.Query(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("Base query failed: %v", err)
	}
	baseCount := baseResult.Total

	// Get initial facets
	baseFacets, err := engine.ComputeFacets(QueryParams{Limit: 100})
	if err != nil {
		t.Fatalf("ComputeFacets failed: %v", err)
	}

	// Test removing single color facet
	if baseFacets.ColourName != nil && len(baseFacets.ColourName.Values) > 0 {
		colorFacet := baseFacets.ColourName.Values[0]
		params := QueryParams{ColourName: []string{colorFacet.Value}, Limit: 100}
		facets, err := engine.ComputeFacets(params)
		if err != nil {
			t.Fatalf("ComputeFacets failed: %v", err)
		}

		// Find the selected color and test its removal URL
		if facets.ColourName != nil {
			for _, cv := range facets.ColourName.Values {
				if cv.Value == colorFacet.Value && cv.Selected {
					testFacetTransition(t, engine, mapper,
						"Color:"+colorFacet.Value, "EMPTY",
						params, cv.URL, baseCount)
					break
				}
			}
		}
	}
}

// facetTransitionsChecked counts calls to testFacetTransition so that
// TestFacetStateMachine can tell a state family that exercised nothing.
var facetTransitionsChecked atomic.Int64

// testFacetTransition tests a single state transition
func testFacetTransition(t *testing.T, engine *Engine, mapper *URLMapper,
	fromState, toState string, fromParams QueryParams, transitionURL string, expectedCount int) {
	t.Helper()
	facetTransitionsChecked.Add(1)

	// Parse the URL
	parts := strings.SplitN(transitionURL, "?", 2)
	path := parts[0]
	query := ""
	if len(parts) == 2 {
		query = parts[1]
	}

	toParams, err := mapper.ParsePath(path, query)
	if err != nil {
		t.Errorf("Transition %s → %s: Failed to parse URL '%s': %v",
			fromState, toState, transitionURL, err)
		return
	}

	// Execute query with new params
	result, err := engine.Query(toParams)
	if err != nil {
		t.Errorf("Transition %s → %s: Query failed: %v", fromState, toState, err)
		return
	}

	// Verify count matches expectation
	if result.Total != expectedCount {
		t.Errorf("Transition %s → %s: Count mismatch: got %d, expected %d (URL: %s)",
			fromState, toState, result.Total, expectedCount, transitionURL)
		return
	}

	t.Logf("✓ Transition %s → %s: %d photos (URL: %s)",
		fromState, toState, result.Total, transitionURL)
}
