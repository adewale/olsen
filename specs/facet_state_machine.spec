# Faceted Navigation as a State Machine

## Core Insight

**Faceted navigation is a state machine where users explore a dataset through valid state transitions.**

The fundamental rule: **Users should never be able to transition from a state with results to a state with zero results.**

This is NOT about hierarchical relationships between facets. It's about maintaining a valid exploration path through the data.

## The Problem with Hierarchical Thinking

### What We Thought (WRONG)

> "Year contains Month, so changing Year should clear Month because of the hierarchical relationship."

**Example of broken behavior:**
```
State A: year=2024&month=11 (has 50 photos)
User clicks: Year 2025
State B: year=2025 (cleared month=11 due to "hierarchy")
```

**Why this is wrong:**
- The system made an assumption about what the user wanted
- If the user had photos from November 2025, they can't get to them
- The "hierarchy" rule is arbitrary - it's based on how we think about calendars, not about the data

### What Actually Matters (CORRECT)

> "Users should only be able to select facet values that will return results given their current filter state."

**Example of correct behavior:**
```
State A: year=2024&month=11 (has 50 photos)

Facet display shows:
  Year:
    - 2023 (120 photos in Nov 2023) ← ENABLED
    - 2024 (50 photos) ← SELECTED
    - 2025 (0 photos in Nov 2025) ← DISABLED

  Month:
    - October (80 photos in Oct 2024) ← ENABLED
    - November (50 photos) ← SELECTED
    - December (30 photos in Dec 2024) ← ENABLED
```

If user clicks "Year 2025" (disabled), nothing happens. They cannot make an invalid transition.

If user clicks "Year 2023", they get:
```
State B: year=2023&month=11 (has 120 photos)
```

The month filter is preserved because "November 2023" is a **valid state with results**.

## State Machine Principles

### 1. Every Filter Combination is a State

```
State = {
  year?: number,
  month?: number,
  day?: number,
  colour?: string[],
  camera?: string,
  lens?: string,
  timeOfDay?: string[],
  season?: string[],
  // ... all other facets
}
```

Each state has a **result count**: `count(photos WHERE state conditions)`

### 2. Valid Transitions

A transition from State A to State B is **valid** if and only if:

```
count(State B) > 0
```

### 3. Facet Values Show Valid Transitions

When computing facet values, the system MUST:

1. **Calculate counts** for each potential next state
2. **Enable** facet values where count > 0
3. **Disable** facet values where count = 0
4. **Show counts** so users understand the size of each transition

### 4. No Assumptions About Relationships

The system should NOT assume:
- Year "contains" Month
- Camera "contains" Lens
- Season "relates to" Month

Instead, the system should COMPUTE which combinations exist in the data.

## Implementation Strategy

### Current Behavior (INCORRECT)

**File:** `internal/query/facet_url_builder.go`

```go
func (b *FacetURLBuilder) buildYearURLs(facet *Facet, baseParams QueryParams) {
    // ...
    if facet.Values[i].Selected {
        p.Year = nil
        p.Month = nil  // ← WRONG: Assumes hierarchy
        p.Day = nil    // ← WRONG: Assumes hierarchy
    } else {
        p.Year = &year
        p.Month = nil  // ← WRONG: Assumes hierarchy
        p.Day = nil    // ← WRONG: Assumes hierarchy
    }
}
```

**Problem:** Clears Month/Day based on assumed hierarchy, not based on whether the resulting state has results.

### Correct Behavior

**File:** `internal/query/facet_url_builder.go`

```go
func (b *FacetURLBuilder) buildYearURLs(facet *Facet, baseParams QueryParams) {
    // ...
    if facet.Values[i].Selected {
        // Remove year filter, PRESERVE all other filters
        p.Year = nil
    } else {
        // Add year filter, PRESERVE all other filters
        p.Year = &year
    }
    // Let the facet computation determine if this state is valid
}
```

**File:** `internal/query/facets.go`

The facet computation already calculates counts correctly:

```go
// When computing Year facet with month=11 already selected:
SELECT
    CAST(strftime('%Y', date_taken) AS INTEGER) as year,
    COUNT(DISTINCT p.id) as count
FROM photos p
WHERE strftime('%m', date_taken) = '11'  -- ← Month filter PRESERVED
GROUP BY year
```

This query will return:
- 2023: 120 (if there are 120 photos from Nov 2023)
- 2024: 50 (if there are 50 photos from Nov 2024)
- (2025 not returned if count = 0)

### UI Rendering

**File:** `internal/explorer/templates/grid.html`

```html
{{range .Values}}
  <div class="facet-value {{if eq .Count 0}}disabled{{end}}">
    {{if gt .Count 0}}
      <a href="{{.URL}}">
        {{.Label}} <span class="count">({{.Count}})</span>
      </a>
    {{else}}
      <span class="disabled-label">
        {{.Label}} <span class="count">(0)</span>
      </span>
    {{end}}
  </div>
{{end}}
```

With CSS:
```css
.facet-value.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.facet-value.disabled .disabled-label {
  color: #999;
  text-decoration: none;
  pointer-events: none;
}
```

## Real-World Examples

### Example 1: Temporal Filters

**Current State:** `year=2024&month=11` (50 photos)

**Year Facet Display:**
```
2023 (120) ← clicking goes to year=2023&month=11
2024 (50)  ← SELECTED
2025 (0)   ← DISABLED (no photos from Nov 2025)
```

**User Action:** Click "2023"

**New State:** `year=2023&month=11` (120 photos)

**Why this works:** The system computed that November 2023 has 120 photos, so it's a valid transition. Month filter was PRESERVED, not cleared.

### Example 2: Camera + Lens Filters

**Current State:** `camera=Canon&lens=RF 50mm` (30 photos)

**Camera Facet Display:**
```
Canon (30)  ← SELECTED
Nikon (15)  ← clicking goes to camera=Nikon&lens=RF 50mm
Sony (0)    ← DISABLED (Sony doesn't make RF mount lenses)
```

**Why this works:** The system knows Sony cameras can't use RF 50mm lens because the data doesn't contain that combination. It's not about "hierarchy" (camera contains lens), it's about valid data combinations.

### Example 3: Color + Season

**Current State:** `season=winter` (200 photos)

**Color Facet Display:**
```
white (80)  ← snow scenes
blue (60)   ← blue hour
brown (40)  ← trees
red (10)    ← winter sunsets
orange (3)  ← rare orange in winter
green (0)   ← DISABLED (no green in winter photos)
```

**Why this works:** The data happens to have no green-dominant photos from winter. This isn't a "rule" - if the photographer shoots evergreen trees in winter next year, green will become enabled.

## Migration Plan

### Phase 1: Remove Hierarchical Clearing (This Change)

**Goal:** Stop clearing dependent filters

**Changes:**
1. Update `facet_url_builder.go` to preserve all filters
2. Remove `p.Month = nil` and `p.Day = nil` from year facet builder
3. Remove `p.Day = nil` from month facet builder
4. Update camera facet to preserve lens (currently it clears lens)

**Expected Behavior:**
- Some facet values will have count=0
- These will appear as links (for now)
- Clicking them leads to zero-result states (temporarily broken)

### Phase 2: Disable Zero-Result Facet Values (Next Change)

**Goal:** Hide or disable facet values with count=0

**Changes:**
1. Update template to check `{{if gt .Count 0}}`
2. Render disabled state for zero-count values
3. Add CSS for disabled styling
4. Add tooltip: "No results with current filters"

**Expected Behavior:**
- Facet values with count=0 are visible but not clickable
- Users understand why certain options are unavailable
- No way to reach zero-result states

### Phase 3: Progressive Disclosure (Future Enhancement)

**Goal:** Only show relevant facets based on current state

**Changes:**
1. Hide Month facet until Year is selected
2. Hide Day facet until Month is selected
3. Add "Show more filters" for contextually irrelevant facets

**Expected Behavior:**
- Cleaner UI with less clutter
- Guided exploration path
- Still based on data, not assumed hierarchy

## Testing Strategy

### Current Tests (OUTDATED)

**File:** `internal/query/facet_hierarchy_test.go`

These tests verify hierarchical clearing:
```go
func TestYearFacetClearsMonthAndDay(t *testing.T) {
    // Expects month to be cleared when year changes
    // THIS IS THE OLD MODEL - NEEDS UPDATE
}
```

### New Tests (STATE MACHINE MODEL)

**File:** `internal/query/facet_state_transitions_test.go`

These tests should verify valid state transitions:

```go
func TestFacetValuesOnlyShowValidTransitions(t *testing.T) {
    // Given: Photos in Nov 2023 and Nov 2024 (not Nov 2025)
    // When: Viewing state year=2024&month=11
    // Then: Year facet shows 2023 (count > 0), 2024 (selected), but NOT 2025
}

func TestTransitionPreservesAllFilters(t *testing.T) {
    // Given: State year=2024&month=11&colour=red
    // When: Changing year to 2023
    // Then: New state is year=2023&month=11&colour=red
    //       (all filters preserved)
}

func TestZeroResultTransitionsAreInvalid(t *testing.T) {
    // Given: Photos only in Nov 2024 (not Nov 2025)
    // When: Computing facets for state month=11
    // Then: Year 2025 should have count=0 or not appear
}
```

## Why This Matters

### User Experience

**Old model (hierarchy):**
- System makes assumptions about what user wants
- Surprising behavior when filters disappear
- Users lose context during exploration
- "Smart" behavior that feels unpredictable

**New model (state machine):**
- System guides users through valid data paths
- No surprises - disabled options are visible but not clickable
- Users maintain full context
- Transparent behavior based on actual data

### Data-Driven Design

The state machine model means:
- No hardcoded rules about facet relationships
- Behavior emerges from the actual data
- Works with any dataset, any facet combination
- Scales to new facets without special cases

### Consistency

Every facet follows the same rule:
1. Compute counts with current filters applied
2. Enable if count > 0, disable if count = 0
3. Show counts to inform users

No special cases for "temporal hierarchies" or "equipment hierarchies".

## References

- **Faceted Search UX Best Practices** - Nielsen Norman Group
- **State Machines in UI Design** - David Khourshid (XState creator)
- **Zero Results in Faceted Navigation** - Baymard Institute research
- **Section 5.1 of facets_spec.md** - "Facet Dependencies" (needs reinterpretation)

## Status

- **Current Implementation:** Hierarchical model (INCORRECT)
- **Target Implementation:** State machine model (CORRECT)
- **Migration:** In progress (2025-10-07)

---

**Key Insight:** Faceted navigation isn't about taxonomy and hierarchies. It's about exploration and valid state transitions through actual data.
