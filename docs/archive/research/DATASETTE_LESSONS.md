# Lessons from Datasette's Faceted Search

**Date:** 2025-10-07
**Sources:**
- https://github.com/simonw/datasette/issues/255
- https://simonwillison.net/2018/May/20/datasette-facets/

## What Datasette Got Right (That We Initially Got Wrong)

### 1. ✅ Facets Are Independent Dimensions

**Datasette's approach:**
```
?_facet=year&_facet=month&_facet=camera
```

Every facet is a **URL parameter**. Adding/removing facets is just adding/removing URL params. No special relationships, no hierarchies.

**What we were doing wrong:**
- Nested breadcrumbs (Month inside Year check)
- Removing Month when Year removed
- Conditional rendering based on hierarchy

**What we fixed:**
- All facets now independent
- URL structure: `?year=2024&month=10&camera=Canon`
- Can have `month=10` without `year` (all Octobers)
- Can have `day=15` without `month` (all 15ths)

**Status:** ✅ FIXED in Phase 2f (server.go active filters) and Phase 2g (url_mapper.go breadcrumbs)

---

### 2. ⚠️ Performance Limits Are First-Class Features

**Datasette's approach:**
- **Primary facet queries: 200ms timeout**
- **Suggested facet discovery: 50ms timeout**
- Queries limited to 31 results to detect truncation
- Graceful degradation if queries timeout

**Our current implementation:**
```go
func (e *Engine) ComputeFacets(params QueryParams) (*FacetCollection, error) {
    facets := &FacetCollection{}

    // Compute each facet - NO TIMEOUTS! ❌
    facets.Camera, err = e.computeCameraFacet(params)
    facets.Lens, err = e.computeLensFacet(params)
    facets.Year, err = e.computeYearFacet(params)
    // ... 9 more facets
}
```

**Problems:**
- With 100K photos, facet computation could take seconds
- No timeout mechanism
- Could block UI rendering
- No graceful degradation

**What we should add:**

```go
func (e *Engine) ComputeFacetsWithTimeout(params QueryParams, timeout time.Duration) (*FacetCollection, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    facets := &FacetCollection{}

    // Compute facets concurrently with timeout
    // If timeout, return partial results with flag

    return facets, nil
}
```

**Datasette's insight:** "Strict time limits force you to think about performance from day one."

**Status:** ⚠️ NOT IMPLEMENTED - Should add in Phase 3

---

### 3. ⚠️ Facet Discovery/Suggestion

**Datasette's approach:**
- Automatically suggests relevant facets based on data characteristics
- Checks columns for:
  - Less than 30 unique values
  - More than 1 unique option
  - Fewer unique options than total filtered rows
- Discovery queries have **50ms timeout** (even stricter!)

**Our current implementation:**
- We show ALL facets always
- No intelligence about which facets are useful
- Year facet shown even for photos from single year

**Example problem:**
```
Database has 10,000 photos, all from 2024
Year facet shows: 2024 (10,000)
Month facet shows: Jan (800), Feb (850), ... Dec (900)
```

Year facet is **useless** here but still displayed. Month facet is **highly useful**.

**What Datasette does:**
- Hides Year facet (only 1 unique value)
- Shows Month facet (12 useful options)
- Shows Camera facet if < 30 unique cameras

**What we should consider:**

```go
func (e *Engine) SuggestFacets(params QueryParams, timeout time.Duration) ([]string, error) {
    // For each potential facet:
    // 1. Quick query: SELECT COUNT(DISTINCT column) WHERE <filters>
    // 2. If count between 2 and 30: suggest it
    // 3. If timeout exceeded: return partial list

    // Prioritize:
    // - Foreign keys (camera, lens)
    // - Enums with low cardinality (time_of_day, season)
    // - Dates if range > 1 year
}
```

**Status:** ⚠️ NOT IMPLEMENTED - Could add in Phase 3

---

### 4. ✅ URL-Based State (Bookmarkable, Shareable)

**Datasette's principle:** "The URL IS the state."

Every filter combination is a unique URL that can be:
- Bookmarked
- Shared
- Crawled by search engines
- Linked from external sites

**Our implementation:** ✅ We got this right from the start!
```
/photos?year=2024&month=10&camera=Canon&colour=red
```

All state in URL, no server-side sessions, fully bookmarkable.

**Status:** ✅ CORRECT from day 1

---

### 5. ✅ Progressive Disclosure

**Datasette's UX:**
1. Start with data table
2. Click "Add facet" → see options
3. Select facet → results filter + counts update
4. Click facet value → further filtering
5. Repeat

Each step is **reversible** and **transparent**.

**Our implementation:** ✅ Similar approach
- Start at `/photos` (all photos)
- Facets show in right rail with counts
- Click facet value → URL updates, results filter
- Click × on chip → facet removed
- All reversible via browser back button

**Status:** ✅ CORRECT

---

### 6. ⚠️ Query Complexity Limits

**Datasette's approach:**
- Limit facet value results to **31 items** (shows ">30" if truncated)
- Prevents UI from becoming overwhelming
- Forces user to filter more to see granular facets

**Our current implementation:**
```go
func (e *Engine) computeCameraFacet(params QueryParams) (*Facet, error) {
    query := `SELECT camera_make, COUNT(*) ... GROUP BY camera_make`
    // NO LIMIT! ❌
}
```

**Example problem:**
```
Database has 500 different cameras
Camera facet returns ALL 500 values
UI becomes unusable scrollbar nightmare
```

**What Datasette does:**
```sql
SELECT column, COUNT(*)
FROM table
WHERE <filters>
GROUP BY column
ORDER BY COUNT(*) DESC
LIMIT 31  -- Only show top 30, +1 to detect truncation
```

If 31 rows returned, show "Top 30 cameras (500 total)"

**What we should add:**

```go
const MaxFacetValues = 30

func (e *Engine) computeCameraFacet(params QueryParams) (*Facet, error) {
    query := `
        SELECT camera_make, COUNT(*) as count
        FROM photos
        WHERE <filters>
        GROUP BY camera_make
        ORDER BY count DESC
        LIMIT ?
    `

    values := []FacetValue{}
    rows, err := e.db.Query(query, MaxFacetValues+1)
    // ... scan rows

    if len(values) > MaxFacetValues {
        facet.Truncated = true
        facet.TotalValues = getTotalCount() // Separate query
        values = values[:MaxFacetValues]
    }

    return &Facet{Values: values, Truncated: facet.Truncated}, nil
}
```

**Status:** ⚠️ NOT IMPLEMENTED - Should add in Phase 3

---

### 7. ✅ Transparent Count Computation

**Datasette's principle:** "Show the user WHY they're seeing what they're seeing"

Every facet value shows:
- Label (e.g., "Canon")
- **Count** (e.g., "1,234 photos")
- Whether it's currently selected

**Our implementation:** ✅ We do this!
```
Canon (450)  ← Count shown
Nikon (320)
Sony (180)
```

User knows exactly what they'll get before clicking.

**Status:** ✅ CORRECT

---

### 8. ⚠️ JSON API for Programmatic Access

**Datasette's approach:**
- Every faceted view has a `.json` version
- `/photos.json?_facet=camera&_facet=year`
- Returns both data AND facet counts
- Enables building custom UIs

**Our current implementation:**
- HTML only
- No API endpoint for facet data

**What we could add:**

```go
func (s *Server) handleQueryJSON(w http.ResponseWriter, r *http.Request) {
    // Parse params same as handleQuery
    result, _ := s.engine.Query(params)
    facets, _ := s.engine.ComputeFacets(params)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "photos": result.Photos,
        "total": result.Total,
        "facets": facets,
        "query_time_ms": result.QueryTimeMs,
    })
}

// Register route:
s.router.HandleFunc("/photos.json", s.handleQueryJSON)
```

**Status:** ⚠️ NOT IMPLEMENTED - Could add later

---

## Key Quotes from Simon Willison

### On Faceted Search Philosophy

> "I love faceted search engines. One of my first approaches to understanding any new large dataset has long been to throw it into a faceted search engine and see what comes out."

This resonates with our use case - Olsen is for **exploring** a photo corpus, not just searching it.

### On Performance

> "I decided to put strict time limits on facet execution. Facet generation queries have a 200 millisecond limit, and facet suggestion queries have a 50 millisecond limit."

Performance limits = **better design** because they force you to think about scale from day one.

### On Progressive Complexity

> "Facets work best when they progressively narrow down results. Start broad, get specific."

This validates our state machine model - each facet click should ALWAYS reduce or maintain result count, never surprise user with zero results.

---

## Summary: What We Should Implement

### Phase 3 Enhancements (Based on Datasette)

**Priority 1 - Performance & Scale:**
1. ⚠️ Add timeout mechanism to facet computation (200ms default)
2. ⚠️ Limit facet values to top 30, show "truncated" indicator
3. ⚠️ Add total unique count query for truncated facets

**Priority 2 - Intelligence:**
4. ⚠️ Smart facet suggestion (hide facets with 1 unique value)
5. ⚠️ Dynamic facet ordering (most useful first)
6. ⚠️ Facet value search for high-cardinality facets (>30 values)

**Priority 3 - API:**
7. ⚠️ JSON endpoint for programmatic access
8. ⚠️ Facet metadata endpoint (available facets + characteristics)

### What We Already Got Right ✅

1. ✅ Independent facet dimensions (no hierarchies)
2. ✅ URL-based state (bookmarkable, shareable)
3. ✅ Transparent counts on every facet value
4. ✅ Progressive disclosure UX
5. ✅ Disabled facets for zero-result states
6. ✅ Reversible navigation (browser back works)

---

## Architectural Validation

Datasette's implementation validates our state machine model:

**Datasette's implicit rule:**
> "Clicking a facet value should ALWAYS show results (never zero)"

This is exactly what we implemented! Facets with count=0 are disabled.

**Datasette's facet independence:**
> "Facets combine via AND logic, but each is independent"

This is our state machine model - Month can exist without Year.

**Datasette's performance-first design:**
> "If a query takes too long, show partial results or skip that facet"

This is what we SHOULD add - graceful degradation.

---

## Conclusion

**What Datasette taught us:**

1. **We were on the right track** with independent facets and URL-based state
2. **We need to add** performance limits and graceful degradation
3. **We could enhance** with smart facet suggestion and truncation
4. **The hierarchical bugs we fixed** were exactly the anti-pattern Datasette avoids

**The big takeaway:**
> "Faceted search is about **progressive discovery**, not hierarchical navigation."

We learned this the hard way by finding 6 locations with hierarchical logic. Datasette got it right from the start by treating facets as independent dimensions.

**Next steps:**
1. Add performance timeouts (Phase 3)
2. Implement facet value limits (Phase 3)
3. Consider smart facet suggestion (Phase 4)
4. Add JSON API (Phase 4)

But first: **finish manual testing** to ensure our hierarchical fixes actually work!

---

**References:**
- Datasette Issue #255: https://github.com/simonw/datasette/issues/255
- Simon's Blog Post: https://simonwillison.net/2018/May/20/datasette-facets/
- Datasette Docs: https://docs.datasette.io/en/stable/facets.html
