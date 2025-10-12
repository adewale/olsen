# Complete Hierarchical Logic Audit

**Date:** 2025-10-07
**Status:** IN PROGRESS

## Why This Document Exists

After finding **5 separate locations** with hierarchical logic bugs, it's clear that hierarchical thinking permeated the entire codebase design. This document audits EVERY location where filters (Year/Month/Day) are manipulated.

## The Pattern

Hierarchical logic appears wherever code assumes:
- "Month requires Year"
- "Day requires Month"
- "Removing Year should clear Month and Day"
- "Month breadcrumb only shows with Year"

## All Locations Where Filters Are Manipulated

### ✅ FIXED

1. **`internal/query/facet_url_builder.go`** (Phase 1)
   - Lines 57-73: Year facet URL building
   - Lines 89-106: Month facet URL building
   - **Bug:** Cleared Month/Day when setting Year
   - **Fix:** Preserve all filters

2. **`internal/query/facets.go`** (Phase 2a)
   - Lines 208-216: `computeYearFacet()`
   - Lines 296-300: `computeMonthFacet()`
   - **Bug:** Cleared Month/Day from params before computing
   - **Fix:** Only clear the filter being computed

3. **`internal/query/engine.go`** (Phase 2e)
   - Lines 138-149: `buildWhereClause()`
   - **Bug:** Month filter only applied if Year was set
   - **Bug:** Day filter only applied if Month AND Year were set
   - **Fix:** Apply filters independently

4. **`internal/explorer/server.go`** (Phase 2f - just fixed)
   - Lines 567-603: `buildActiveFilters()`
   - **Bug:** Removing Year chip cleared Month and Day
   - **Bug:** Month and Day chips not shown at all
   - **Fix:** Added Month/Day chips, removed hierarchical clearing

### ❌ NOT YET FIXED

5. **`internal/query/url_mapper.go`** (Phase 2g - NEEDS FIX)
   - Lines 374-434: `BuildBreadcrumbs()`
   - **Bug:** Month breadcrumb nested inside Year check (line 386)
   - **Bug:** Day breadcrumb nested inside Month check (line 394)
   - **Impact:** `month=10` shows no breadcrumb, `day=15` shows no breadcrumb
   - **Fix Needed:** Show Month breadcrumb even without Year, Day even without Month

### 🔍 TO AUDIT

6. **`internal/query/url_mapper.go`** - Other functions
   - `BuildPath()` - Does it handle Month-only paths?
   - `BuildFullURL()` - Does it correctly encode Month without Year?
   - Need to check ALL URL building functions

7. **`internal/explorer/server.go`** - Route handlers
   - How are paths like `/photos?month=10` parsed?
   - Are there any route-level restrictions?

8. **Template rendering** - `internal/explorer/templates/grid.html`
   - Are facets rendered correctly when Month but no Year?
   - Are breadcrumbs positioned correctly?
   - Any conditional logic based on hierarchy?

## Search Patterns Used

```bash
# Find all Month/Day clearing
grep -rn "\.Month = nil\|\.Day = nil" internal/

# Find nested conditionals
grep -rn "if params.Month.*{" internal/ -A 10 | grep "if params.Year"

# Find breadcrumb building
grep -rn "Breadcrumb\|breadcrumb" internal/ -i

# Find URL building
grep -rn "BuildPath\|BuildURL\|Build.*URL" internal/query/url_mapper.go
```

## Testing Strategy

### Manual Testing Checklist

For EACH filter combination, verify:
- [ ] Query returns correct photos
- [ ] Facets show correct counts
- [ ] Active filter chips appear
- [ ] Breadcrumbs show correct path
- [ ] Clicking facets preserves other filters
- [ ] Removing filters (via chip ×) works correctly

**Test combinations:**
- [ ] `month=10` (Month only)
- [ ] `day=15` (Day only)
- [ ] `month=10&day=15` (Month + Day, no Year)
- [ ] `year=2024` (Year only)
- [ ] `year=2024&month=10` (Year + Month)
- [ ] `year=2024&month=10&day=15` (All three)

**For each, test:**
1. Navigate to state
2. Check breadcrumbs
3. Check active chips
4. Check facet counts
5. Click different facets
6. Remove filters via chips
7. Verify results at each step

### Automated Testing

Create integration test: `internal/query/state_machine_integration_test.go`

Test EVERY transition:
- From `{}` (no filters) to any single filter
- From any single filter to any combination
- From any combination to removing one filter
- Verify counts, breadcrumbs, chips at each step

## Prevention Strategy

### Code Review Checklist

When adding ANY code that manipulates QueryParams, verify:
- [ ] Does it clear filters unnecessarily?
- [ ] Does it nest conditionals based on hierarchy?
- [ ] Does it assume Month requires Year?
- [ ] Does it assume Day requires Month?
- [ ] Would it work with Month-only queries?
- [ ] Would it work with Day-only queries?

### Architecture Documentation

Document the principle: **"All filter dimensions are independent"**

This means:
- Month can exist without Year (all Octobers)
- Day can exist without Month (all 15ths)
- Camera can combine with any temporal filter
- Colour can combine with anything

No exceptions. No special cases. No hierarchies.

## Current Status

**Locations audited:** 5 of ~8
**Locations fixed:** 4
**Locations remaining:** 1-4 (breadcrumbs + possibly others)

**Next steps:**
1. Fix breadcrumb generation (url_mapper.go)
2. Audit all URL building functions
3. Manual test all combinations
4. Create integration tests
5. Update documentation

---

**The fundamental problem:** Hierarchical thinking was the ORIGINAL DESIGN, not an implementation bug. Every part of the system was designed assuming hierarchies. We're not fixing bugs - we're migrating architectures.
