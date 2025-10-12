package query

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// FacetTransitionLog contains structured logging information about available state transitions
type FacetTransitionLog struct {
	CurrentState  StateInfo        `json:"current_state"`
	Transitions   []TransitionInfo `json:"transitions"`
	TotalResults  int              `json:"total_results"`
	DisabledCount int              `json:"disabled_count"`
	EnabledCount  int              `json:"enabled_count"`
}

// StateInfo describes the current query state
type StateInfo struct {
	Year        *int     `json:"year,omitempty"`
	Month       *int     `json:"month,omitempty"`
	Day         *int     `json:"day,omitempty"`
	ColourName  []string `json:"colour_name,omitempty"`
	CameraMake  []string `json:"camera_make,omitempty"`
	CameraModel []string `json:"camera_model,omitempty"`
	LensModel   []string `json:"lens_model,omitempty"`
	TimeOfDay   []string `json:"time_of_day,omitempty"`
	Season      []string `json:"season,omitempty"`
	InBurst     *bool    `json:"in_burst,omitempty"`
	FilterCount int      `json:"filter_count"`
}

// TransitionInfo describes a possible state transition
type TransitionInfo struct {
	FacetType   string `json:"facet_type"`   // "year", "month", "colour", etc.
	FacetValue  string `json:"facet_value"`  // "2024", "November", "red", etc.
	FacetLabel  string `json:"facet_label"`  // Display label
	ResultCount int    `json:"result_count"` // Expected number of photos
	IsEnabled   bool   `json:"is_enabled"`   // true if count > 0
	IsSelected  bool   `json:"is_selected"`  // true if currently active
	TargetURL   string `json:"target_url"`   // URL for this transition
}

// BuildStateInfo creates StateInfo from QueryParams
func BuildStateInfo(params QueryParams) StateInfo {
	filterCount := 0

	if params.Year != nil {
		filterCount++
	}
	if params.Month != nil {
		filterCount++
	}
	if params.Day != nil {
		filterCount++
	}
	if len(params.ColourName) > 0 {
		filterCount++
	}
	if len(params.CameraMake) > 0 {
		filterCount++
	}
	if len(params.CameraModel) > 0 {
		filterCount++
	}
	if len(params.LensModel) > 0 {
		filterCount++
	}
	if len(params.TimeOfDay) > 0 {
		filterCount++
	}
	if len(params.Season) > 0 {
		filterCount++
	}
	if params.InBurst != nil {
		filterCount++
	}

	return StateInfo{
		Year:        params.Year,
		Month:       params.Month,
		Day:         params.Day,
		ColourName:  params.ColourName,
		CameraMake:  params.CameraMake,
		CameraModel: params.CameraModel,
		LensModel:   params.LensModel,
		TimeOfDay:   params.TimeOfDay,
		Season:      params.Season,
		InBurst:     params.InBurst,
		FilterCount: filterCount,
	}
}

// BuildTransitionLog creates a structured log of all possible transitions from the current state
func BuildTransitionLog(params QueryParams, facets *FacetCollection, totalResults int) *FacetTransitionLog {
	if facets == nil {
		return nil
	}

	log := &FacetTransitionLog{
		CurrentState: BuildStateInfo(params),
		Transitions:  []TransitionInfo{},
		TotalResults: totalResults,
	}

	// Extract transitions from each facet
	if facets.Year != nil {
		for _, v := range facets.Year.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "year",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.Month != nil {
		for _, v := range facets.Month.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "month",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.Camera != nil {
		for _, v := range facets.Camera.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "camera",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.Lens != nil {
		for _, v := range facets.Lens.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "lens",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.ColourName != nil {
		for _, v := range facets.ColourName.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "colour",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.TimeOfDay != nil {
		for _, v := range facets.TimeOfDay.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "time_of_day",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.Season != nil {
		for _, v := range facets.Season.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "season",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	if facets.InBurst != nil {
		for _, v := range facets.InBurst.Values {
			log.Transitions = append(log.Transitions, TransitionInfo{
				FacetType:   "in_burst",
				FacetValue:  v.Value,
				FacetLabel:  v.Label,
				ResultCount: v.Count,
				IsEnabled:   v.Count > 0,
				IsSelected:  v.Selected,
				TargetURL:   v.URL,
			})
			if v.Count == 0 {
				log.DisabledCount++
			} else {
				log.EnabledCount++
			}
		}
	}

	return log
}

// LogTransitions logs the facet transition information in structured format
func LogTransitions(params QueryParams, facets *FacetCollection, totalResults int) {
	transitionLog := BuildTransitionLog(params, facets, totalResults)
	if transitionLog == nil {
		return
	}

	// Log in JSON format for easy parsing
	jsonData, err := json.Marshal(transitionLog)
	if err != nil {
		log.Printf("FACET_TRANSITIONS_ERROR: Failed to marshal transition log: %v", err)
		return
	}

	log.Printf("FACET_TRANSITIONS: %s", string(jsonData))
}

// LogTransitionsSummary logs a compact summary of available transitions
func LogTransitionsSummary(params QueryParams, facets *FacetCollection, totalResults int) {
	transitionLog := BuildTransitionLog(params, facets, totalResults)
	if transitionLog == nil {
		return
	}

	// Build compact summary
	var parts []string

	// Current state
	stateDesc := buildStateDescription(transitionLog.CurrentState)
	parts = append(parts, "state="+stateDesc)

	// Results
	parts = append(parts, fmt.Sprintf("results=%d", totalResults))

	// Enabled vs disabled transitions
	parts = append(parts, fmt.Sprintf("enabled=%d", transitionLog.EnabledCount))
	parts = append(parts, fmt.Sprintf("disabled=%d", transitionLog.DisabledCount))

	// Log disabled transitions (these are the critical ones - should not be clickable)
	var disabledTransitions []string
	for _, t := range transitionLog.Transitions {
		if !t.IsEnabled {
			disabledTransitions = append(disabledTransitions, t.FacetType+":"+t.FacetValue)
		}
	}

	if len(disabledTransitions) > 0 {
		parts = append(parts, "disabled_facets=["+strings.Join(disabledTransitions, ",")+"]")
	}

	log.Printf("FACET_STATE: %s", strings.Join(parts, " "))
}

// buildStateDescription creates a compact description of the current state
func buildStateDescription(state StateInfo) string {
	if state.FilterCount == 0 {
		return "all_photos"
	}

	var parts []string
	if state.Year != nil {
		parts = append(parts, fmt.Sprintf("year=%d", *state.Year))
	}
	if state.Month != nil {
		parts = append(parts, fmt.Sprintf("month=%d", *state.Month))
	}
	if state.Day != nil {
		parts = append(parts, fmt.Sprintf("day=%d", *state.Day))
	}
	if len(state.ColourName) > 0 {
		parts = append(parts, "colour="+strings.Join(state.ColourName, ","))
	}
	if len(state.CameraMake) > 0 {
		parts = append(parts, "camera="+strings.Join(state.CameraMake, ","))
	}
	if len(state.TimeOfDay) > 0 {
		parts = append(parts, "time="+strings.Join(state.TimeOfDay, ","))
	}

	return strings.Join(parts, "&")
}

// ValidateTransition checks if a transition from prevState to currentState is valid
// Returns true if the transition is valid (count > 0 or direct URL entry)
func ValidateTransition(prevParams, currentParams QueryParams, prevFacets *FacetCollection) (valid bool, expectedCount int, message string) {
	// If no previous facets, this is a direct entry (always valid)
	if prevFacets == nil {
		return true, -1, "direct_entry"
	}

	// Determine which facet changed
	changedFacet, changedValue := detectTransition(prevParams, currentParams)
	if changedFacet == "" {
		return true, -1, "no_change"
	}

	// Find the expected count for this transition
	expectedCount = findExpectedCount(changedFacet, changedValue, prevFacets)

	if expectedCount == 0 {
		return false, 0, "invalid_transition: " + changedFacet + "=" + changedValue + " had count=0"
	}

	return true, expectedCount, "valid_transition"
}

// LogSuspiciousZeroResults logs additional information when we get zero results
// to help detect if this was due to a bug (facet count mismatch) or expected (direct URL)
func LogSuspiciousZeroResults(params QueryParams, facets *FacetCollection) {
	// When we have zero results, check if all enabled facets have count > 0
	// If we have facets with count=0 that are marked as enabled, that's a BUG

	if facets == nil {
		log.Printf("  Note: No facets available for validation (direct URL entry or error)")
		return
	}

	suspiciousCount := 0

	// Check all facet types for count=0 values that are marked as enabled
	checkFacet := func(facet *Facet, facetType string) {
		if facet == nil {
			return
		}
		for _, v := range facet.Values {
			// If count=0 but not explicitly marked as disabled, that's suspicious
			// (In our model, count=0 should mean IsEnabled=false)
			if v.Count == 0 && !strings.Contains(v.URL, "disabled") {
				log.Printf("  SUSPICIOUS: %s facet '%s' has count=0 but appears enabled",
					facetType, v.Label)
				suspiciousCount++
			}
		}
	}

	checkFacet(facets.Year, "Year")
	checkFacet(facets.Month, "Month")
	checkFacet(facets.Camera, "Camera")
	checkFacet(facets.Lens, "Lens")
	checkFacet(facets.ColourName, "Colour")
	checkFacet(facets.TimeOfDay, "TimeOfDay")
	checkFacet(facets.Season, "Season")
	checkFacet(facets.InBurst, "InBurst")

	if suspiciousCount > 0 {
		log.Printf("  WARNING: Found %d facet values with count=0 - UI should render these as disabled", suspiciousCount)
	}

	// Provide diagnostic hint about the zero results
	filterCount := 0
	if params.Year != nil {
		filterCount++
	}
	if params.Month != nil {
		filterCount++
	}
	if params.Day != nil {
		filterCount++
	}
	if len(params.ColourName) > 0 {
		filterCount++
	}
	if len(params.CameraMake) > 0 {
		filterCount++
	}

	if filterCount == 0 {
		log.Printf("  Note: No filters active - database may be empty")
	} else if filterCount == 1 {
		log.Printf("  Note: Single filter active - might be direct URL entry")
	} else {
		log.Printf("  Note: %d filters active - check if this combination exists in data", filterCount)
	}
}

// detectTransition identifies which facet changed between two states
func detectTransition(prev, current QueryParams) (facetType string, value string) {
	// Check Year
	if (prev.Year == nil && current.Year != nil) || (prev.Year != nil && current.Year != nil && *prev.Year != *current.Year) {
		if current.Year != nil {
			return "year", fmt.Sprintf("%d", *current.Year)
		}
	}

	// Check Month
	if (prev.Month == nil && current.Month != nil) || (prev.Month != nil && current.Month != nil && *prev.Month != *current.Month) {
		if current.Month != nil {
			return "month", fmt.Sprintf("%d", *current.Month)
		}
	}

	// Check Colour
	if len(current.ColourName) > 0 && (len(prev.ColourName) == 0 || prev.ColourName[0] != current.ColourName[0]) {
		return "colour", current.ColourName[0]
	}

	// Check Camera
	if len(current.CameraMake) > 0 && (len(prev.CameraMake) == 0 || prev.CameraMake[0] != current.CameraMake[0]) {
		return "camera", current.CameraMake[0]
	}

	return "", ""
}

// findExpectedCount looks up the expected count for a facet value in the previous facets
func findExpectedCount(facetType, value string, facets *FacetCollection) int {
	switch facetType {
	case "year":
		if facets.Year != nil {
			for _, v := range facets.Year.Values {
				if v.Value == value {
					return v.Count
				}
			}
		}
	case "month":
		if facets.Month != nil {
			for _, v := range facets.Month.Values {
				if v.Value == value {
					return v.Count
				}
			}
		}
	case "colour":
		if facets.ColourName != nil {
			for _, v := range facets.ColourName.Values {
				if v.Value == value {
					return v.Count
				}
			}
		}
	case "camera":
		if facets.Camera != nil {
			for _, v := range facets.Camera.Values {
				if v.Value == value {
					return v.Count
				}
			}
		}
	}
	return -1 // Not found (might be direct URL entry)
}
