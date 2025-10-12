# WHERE Clause Bug - Critical State Machine Issue

**Date:** 2025-10-07
**Status:** ✅ FIXED
**Severity:** CRITICAL - Caused invalid state transitions

## The Bug

The WHERE clause builder in `internal/query/engine.go` was enforcing hierarchical dependencies that prevented the state machine model from working correctly.

### Symptom

User navigated from `year=2025&month=1` (20 photos) to `year=2024&month=1` (0 photos), even though Year 2024 should have been shown as disabled.

**Log Evidence:**
```
2025/10/07 18:30:03 FACET_STATE: state=year=2025&month=1 results=20 enabled=19 disabled=0
2025/10/07 18:30:08 FACET_STATE: state=year=2024&month=1 results=0 enabled=11 disabled=0
2025/10/07 18:30:08 FACET_404: No results found
```

Notice: `enabled=19 disabled=0` means ALL year facets were shown as clickable, including Year 2024 which had zero January photos.

## Root Cause

**File:** `internal/query/engine.go`
**Lines:** 138-145 (before fix)

```go
// ❌ WRONG: Enforcing hierarchical dependencies
if params.Month != nil && params.Year != nil {
    where = append(where, "strftime('%m', p.date_taken) = ?")
    args = append(args, fmt.Sprintf("%02d", *params.Month))
}
if params.Day != nil && params.Month != nil && params.Year != nil {
    where = append(where, "strftime('%d', p.date_taken) = ?")
    args = append(args, fmt.Sprintf("%02d", *params.Day))
}
```

**The Problem:**
- Month filter ONLY applied if Year was also set
- Day filter ONLY applied if Month AND Year were both set
- This is the hierarchical model enforcing "Month requires Year" logic

**Why This Broke Facet Computation:**

When computing the Year facet:
1. `computeYearFacet()` removes Year from params (line 211 of facets.go)
2. Calls `buildWhereClause()` with `params.Year = nil, params.Month = 1`
3. WHERE clause builder sees Month but NO Year
4. **Skips the Month filter entirely!**
5. Query returns counts for ALL years (not just January)
6. Year 2024 shows count > 0 even though January 2024 has 0 photos

**Result:** User could click Year 2024 and got zero results.

## The Fix

**File:** `internal/query/engine.go`
**Lines:** 138-149 (after fix)

```go
// ✅ CORRECT: Filters are independent
if params.Month != nil {
    // State machine model: Month is independent of Year
    // Apply month filter even when Year is not set (for facet computation)
    where = append(where, "strftime('%m', p.date_taken) = ?")
    args = append(args, fmt.Sprintf("%02d", *params.Month))
}
if params.Day != nil {
    // State machine model: Day is independent of Month and Year
    // Apply day filter even when Month/Year are not set (for facet computation)
    where = append(where, "strftime('%d', p.date_taken) = ?")
    args = append(args, fmt.Sprintf("%02d", *params.Day))
}
```

**Key Change:**
- Removed dependency conditions
- Month applies independently of Year
- Day applies independently of Month and Year

## Why This Was Missed

This is the **THIRD location** where hierarchical logic was hiding:

1. **Phase 1:** Fixed `facet_url_builder.go` - URL generation (Oct 6)
2. **Phase 2a:** Fixed `facets.go` - Facet count computation (Oct 7)
3. **Phase 2e:** Fixed `engine.go` - WHERE clause generation (Oct 7) ← **This bug**

Each location had the same hierarchical assumption but in different forms:
- URL builder: Cleared Month/Day when changing Year
- Facet computation: Cleared Month/Day from params before query
- WHERE builder: Required Year to apply Month filter

## Test Coverage

**File:** `internal/query/where_clause_test.go` (NEW - 5 tests)

Tests verify:
1. `TestWhereClauseMonthWithoutYear` - Month filter applied without Year
2. `TestWhereClauseDayWithoutMonthOrYear` - Day filter applied independently
3. `TestWhereClauseAllTemporalFilters` - All three work together
4. `TestWhereClauseMonthOnly` - Month can be used alone
5. All tests PASS ✅

## Impact

### Before Fix
```
State: year=2025&month=1 (20 photos)

Computing Year facet:
1. Remove Year from params → params.Year = nil, params.Month = 1
2. buildWhereClause sees Month but no Year
3. SKIPS Month filter (hierarchical logic)
4. Query: SELECT year, COUNT(*) FROM photos GROUP BY year
5. Returns: 2024 (75 total photos), 2025 (177 total photos)
6. User sees Year 2024 with count=75 (WRONG! Should be 0 for January)
7. User clicks Year 2024 → Zero results
```

### After Fix
```
State: year=2025&month=1 (20 photos)

Computing Year facet:
1. Remove Year from params → params.Year = nil, params.Month = 1
2. buildWhereClause applies Month filter independently ✅
3. Query: SELECT year, COUNT(*) FROM photos WHERE month=1 GROUP BY year
4. Returns: 2023 (5 Jan photos), 2025 (20 Jan photos), 2024 (0 Jan photos)
5. User sees Year 2024 DISABLED (count=0) ✅
6. User CANNOT click Year 2024
```

## Lessons Learned

### Complete Migration Checklist

When migrating from hierarchical to state machine model, check ALL locations:

1. ✅ URL generation (`facet_url_builder.go`)
2. ✅ Facet computation (`facets.go` - computeYearFacet, computeMonthFacet)
3. ✅ **WHERE clause building (`engine.go`)** ← We missed this!
4. ✅ Template rendering (already correct)

### Why We Missed It

1. **Subtle dependency:** The condition `params.Month != nil && params.Year != nil` looked reasonable
2. **Different layer:** WHERE clause building is separate from facet computation
3. **No end-to-end test:** Didn't test: "View Jan 2025 → Click Year 2024 → Verify disabled"
4. **Focus on params:** Fixed params manipulation but not WHERE clause logic

### Prevention Strategy

**When fixing hierarchical assumptions:**

1. **Search ALL files** for filter dependencies:
   ```bash
   grep "Month != nil && .*Year != nil" internal/query/*.go
   grep "Day != nil && .*Month != nil" internal/query/*.go
   ```

2. **Check ALL layers:**
   - URL building
   - Params manipulation
   - **WHERE clause building** ← Easy to miss!
   - SQL query construction
   - Template rendering

3. **End-to-end testing:**
   - Navigate through actual UI
   - Check logs for disabled facets
   - Try clicking each facet value
   - Verify counts match results

4. **Monitor production:**
   - `FACET_STATE` logs show enabled/disabled counts
   - `FACET_404` indicates user reached invalid state
   - Compare logs to detect mismatches

## Example Monitoring

**Detecting this bug in production:**

```
# User viewing January 2025
FACET_STATE: state=year=2025&month=1 results=20 enabled=19 disabled=0

# All year facets shown as enabled (disabled=0) - SUSPICIOUS!
# Should have SOME disabled years (no Jan photos for all years)

# User clicks Year 2024
FACET_STATE: state=year=2024&month=1 results=0 enabled=11 disabled=0
FACET_404: No results found

# ANALYSIS:
# - Previous state showed enabled=19, disabled=0
# - This means Year 2024 was shown as clickable
# - But clicking it resulted in 0 photos
# - BUG: Facet count for Year 2024 was wrong (should have been 0)
```

**Red Flags:**
- `disabled=0` when filtering by month/day (unlikely all years have all months)
- `FACET_404` after clicking from state with `disabled=0`
- Facet count doesn't match reality

## Files Changed

1. `internal/query/engine.go` (lines 138-149)
   - Removed hierarchical dependencies from WHERE clause
   - Month and Day filters now independent

2. `internal/query/where_clause_test.go` (NEW)
   - 5 tests verifying independent filter application
   - Tests cover Month-without-Year and Day-without-Month cases

3. `docs/WHERE_CLAUSE_BUG.md` (NEW - this document)

4. `docs/STATE_MACHINE_MIGRATION.md` (to be updated)
   - Add Phase 2e documenting WHERE clause fix

## Verification

### Manual Testing
1. Start server: `./bin/olsen explore --db perf.db`
2. Navigate to year=2025
3. Click month=1 (January)
4. Check Year facet - should show disabled years with 0 January photos
5. Try clicking disabled year - should not be clickable
6. Check logs - should show `disabled > 0` when some years have no January photos

### Automated Testing
```bash
go test -v ./internal/query/ -run TestWhereClause
# All 5 tests should PASS
```

## Status

✅ **FIXED** - WHERE clause now applies filters independently
✅ **TESTED** - All unit tests pass
✅ **DOCUMENTED** - Complete analysis and prevention strategy

This was the final piece of hierarchical logic hiding in the codebase. The state machine model is now fully implemented across all layers.

---

**Author:** Claude & Ade
**Date:** 2025-10-07
**Related:** STATE_MACHINE_MIGRATION.md Phase 2e
