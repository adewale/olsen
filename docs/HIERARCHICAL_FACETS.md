# Faceted Navigation: From Hierarchical to State Machine Model

**Status:** ✅ Corrected (2025-10-07)

## The Problem We Discovered

**Original Implementation (INCORRECT):** We implemented faceted navigation with "hierarchical" relationships, where changing Year would automatically clear Month and Day filters. This was based on the assumption that temporal facets form a containment hierarchy.

**What We Learned:** This hierarchical model was fundamentally wrong. The issue wasn't about hierarchy—it was about **valid state transitions in a state machine**.

## The Core Insight

> **Faceted navigation is a state machine where users explore a dataset through valid state transitions. Users should never be able to transition from a state with results to a state with zero results.**

### What Actually Matters

- **NOT**: "Year contains Month, so changing Year should clear Month"
- **YES**: "Users should only see facet values that lead to states with results > 0"

## The Broken Behavior (Old Hierarchical Model)

### Example: The Bug That Revealed the Truth

```
Current State: year=2024&month=11 (50 photos from November 2024)

User clicks: Year 2025

Old Behavior (BROKEN):
→ Result: year=2025 (month=11 was cleared)
→ Problem: If user had photos from November 2025, they can't get to them
→ Problem: System made assumptions about what user wanted
→ Problem: Surprising filter removal

The Bug: If NO photos exist from November 2025, the Year facet should have
shown "2025" as DISABLED or with count=0. The problem wasn't that we
needed to clear Month—it was that we allowed an invalid transition!
```

## The Correct Behavior (State Machine Model)

### Same Example, Fixed

```
Current State: year=2024&month=11 (50 photos from November 2024)

Facet Display:
  Year Facet:
    □ 2023 (120) ← Enabled (120 photos from Nov 2023 exist)
    ☑ 2024 (50)  ← Selected
    □ 2025 (0)   ← DISABLED (no photos from Nov 2025)

User clicks: Year 2023
→ Result: year=2023&month=11 (120 photos from November 2023)
→ Why: Month filter was PRESERVED because Nov 2023 is a valid state

User tries to click: Year 2025 (disabled)
→ Result: Nothing happens (or shows tooltip: "No results with current filters")
→ Why: Invalid transition prevented at UI level
```

### Key Principles

1. **Preserve All Filters**: Changing any facet value preserves ALL other active filters
2. **Compute Validity**: The facet computation layer determines which transitions are valid
3. **Disable Invalid**: Facet values with count=0 are shown but disabled in the UI
4. **No Assumptions**: System doesn't assume relationships—it uses actual data

## Implementation Details

### What Changed

**File: `internal/query/facet_url_builder.go`**

**Before (WRONG - Hierarchical):**
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

**After (CORRECT - State Machine):**
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

### How Facet Computation Works (Already Correct!)

**File: `internal/query/facets.go`**

The facet computation was already doing the right thing:

```go
// Computing Year facet with month=11 already selected:
SELECT
    CAST(strftime('%Y', date_taken) AS INTEGER) as year,
    COUNT(DISTINCT p.id) as count
FROM photos p
WHERE strftime('%m', date_taken) = '11'  -- ← Month filter PRESERVED
GROUP BY year
```

This query returns:
- `2023: 120` (if 120 photos from Nov 2023 exist)
- `2024: 50` (if 50 photos from Nov 2024 exist)
- (2025 not returned if count = 0)

**This is perfect!** The SQL naturally computes which Year values are valid transitions given the current Month filter.

## Real-World Examples

### Example 1: Temporal Navigation

```
State: year=2024&month=11 (50 photos)

Year Facet Shows:
- 2023 (120) ← Valid: 120 photos from Nov 2023
- 2024 (50)  ← Selected
- 2025 (0)   ← Disabled: No Nov 2025 photos

Click 2023 → year=2023&month=11 (120 photos)
✅ Month filter preserved, valid transition
```

### Example 2: Equipment + Time

```
State: camera=Canon&lens=RF 50mm (30 photos)

Camera Facet Shows:
- Canon (30)  ← Selected
- Nikon (15)  ← Valid: 15 photos with Nikon + RF 50mm... wait, that's impossible!
- Sony (0)    ← Disabled: Sony can't use RF mount

Actually: Nikon won't appear either (count=0) because no Nikon cameras
use Canon RF mount lenses. The data determines valid combinations, not
hardcoded rules!
```

### Example 3: Color + Season

```
State: season=winter (200 photos)

Color Facet Shows:
- white (80)  ← Snow scenes
- blue (60)   ← Blue hour
- brown (40)  ← Trees
- red (10)    ← Winter sunsets
- orange (3)  ← Rare
- green (0)   ← Disabled: No green in winter photos (in this dataset)

If photographer shoots evergreen trees next winter and re-indexes,
green will automatically become enabled. No code changes needed!
```

## Why This Matters

### User Experience

| Old Model (Hierarchical) | New Model (State Machine) |
|---|---|
| System makes assumptions | System guides based on data |
| Filters disappear unexpectedly | All filters visible, some disabled |
| Surprising behavior | Transparent behavior |
| "Smart" but unpredictable | Simple and predictable |
| Users lose context | Users maintain full context |

### Engineering

| Old Model (Hierarchical) | New Model (State Machine) |
|---|---|
| Special cases for each relationship | One rule for all facets |
| Hardcoded clearing logic | Emergent behavior from data |
| Breaks with new facet types | Scales automatically |
| Tightly coupled | Loosely coupled |

## Migration Completed

### Phase 1: Remove Hierarchical Clearing ✅

**Changes:**
- Updated `facet_url_builder.go` to preserve all filters
- Removed `p.Month = nil` and `p.Day = nil` from year facet builder
- Removed `p.Day = nil` from month facet builder
- Updated all tests to expect filter preservation

**Result:**
- Filters are now preserved during transitions
- Facet computation already calculates correct counts
- Some facet values may have count=0 (temporarily clickable)

### Phase 2: Disable Zero-Result Facet Values (TODO)

**Goal:** Prevent users from making invalid transitions

**Changes Needed:**
1. Update `internal/explorer/templates/grid.html` template
2. Check `{{if gt .Count 0}}` for each facet value
3. Render disabled state for zero-count values
4. Add CSS for disabled styling
5. Add tooltip: "No results with current filters"

**Expected Behavior:**
- Facet values with count=0 visible but not clickable
- Users understand why options are unavailable
- No way to reach zero-result states via UI

### Phase 3: Progressive Disclosure (Future)

**Goal:** Cleaner UI with guided exploration

**Possible Enhancements:**
- Hide Month facet until Year is selected (contextual relevance)
- Hide Day facet until Month is selected
- "Show more filters" for less relevant facets
- Still based on data, not assumed hierarchy

## Testing

### Updated Test Files

1. **`facet_hierarchy_test.go`** - Renamed functions, now tests filter preservation
2. **`facet_state_transitions_test.go`** - Updated all assertions to expect preservation
3. **`facet_state_machine_test.go`** - Already had correct state machine tests!

### Test Results

```bash
$ make test-facets
PASS: TestYearFacetPreservesMonthAndDay
PASS: TestMonthFacetPreservesDay
PASS: TestRemovingYearPreservesMonth
PASS: TestTransition_YearMonth_To_DifferentYear
PASS: TestTransition_YearColorCamera_To_DifferentYear
PASS: TestTransition_RemoveYearWithColorAndCamera
PASS: TestTransition_MonthDay_To_DifferentMonth
```

All tests pass! ✅

## Key Takeaways

1. **Facets aren't hierarchical** - They're independent dimensions that can be combined
2. **Valid transitions matter** - Prevent zero-result states, not based on hierarchy
3. **Data drives behavior** - Don't hardcode relationships, compute them
4. **SQL does the work** - WHERE clauses with GROUP BY naturally compute valid transitions
5. **UI provides guidance** - Show what's possible, disable what's not

## References

- **New Spec**: `specs/facet_state_machine.spec` - Complete state machine model
- **Faceted Search**: Ben Shneiderman's "Dynamic queries for visual information seeking"
- **State Machines in UX**: David Khourshid - XState and UI state modeling
- **Zero Results**: Baymard Institute - "Show count=0 options as disabled, don't hide them"

---

**Date Corrected:** 2025-10-07
**Previous Status:** Hierarchical model (INCORRECT)
**Current Status:** State machine model (CORRECT)
**Breaking Change:** Yes - URL behavior changed, but for the better!
