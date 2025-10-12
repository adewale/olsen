# Facet 404 Logging

The explorer server logs detailed information whenever a facet navigation results in a 404 or empty result set. This information can be used to write tests that ensure the behavior is correct.

## Log Format

All facet-related 404 logs start with `FACET_404:` or `FACET_ERROR:` prefix.

### Types of Logs

#### 1. URL Parse Failed
```
FACET_404: URL parse failed - path=/color/red/year/9999 query= error=invalid year format
```

**Fields:**
- `path`: The URL path that was requested
- `query`: The query string (may be empty)
- `error`: The parsing error message

**Meaning:** The URL format is invalid and couldn't be parsed into QueryParams.

**Test Action:**
- Verify the URL is properly rejected
- If it should be valid, update URLMapper to support it

#### 2. No Results Found
```
FACET_404: No results found - path=/color/purple query= params={ColorName:[purple] Year:<nil> Month:<nil> Day:<nil> CameraMake:[] CameraModel:[] LensModel:[] TimeOfDay:[] Season:[] FocalCategory:[] ShootingCondition:[] InBurst:<nil> Limit:100 Offset:0}
```

**Fields:**
- `path`: The URL path that was requested
- `query`: The query string (may be empty)
- `params`: The full parsed QueryParams struct

**Meaning:** The URL parsed successfully, the query executed successfully, but returned 0 photos.

**Test Action:**
- Verify the query is semantically valid
- Create a test database with known data
- Execute the query and verify it correctly returns 0 results
- Verify facets are still computed correctly
- Consider if this combination should exist in your dataset

#### 3. No Route Matched
```
FACET_404: No route matched - path=/invalid_path query=
```

**Fields:**
- `path`: The URL path that was requested
- `query`: The query string (may be empty)

**Meaning:** The URL didn't match any registered route handler.

**Test Action:**
- Verify 404 is returned correctly
- If this path should be supported, add a route in setupRoutes()

#### 4. Query Execution Failed
```
FACET_ERROR: Query execution failed - path=/color/red params={...} error=database connection lost
```

**Fields:**
- `path`: The URL path that was requested
- `params`: The parsed QueryParams
- `error`: The query execution error

**Meaning:** The URL parsed successfully but query execution failed. This usually indicates a bug.

**Test Action:**
- Investigate the query construction
- Verify database constraints
- Check for SQL errors
- This is usually a bug that needs fixing

## Writing Tests from Logs

### Example 1: Testing URL Parsing

**Log Entry:**
```
FACET_404: URL parse failed - path=/color/red/year/2025 query= error=unexpected segment 'year'
```

**Test:**
```go
func TestColorYearURLParsing(t *testing.T) {
    mapper := query.NewURLMapper()

    // This URL was failing in production
    params, err := mapper.ParsePath("/color/red/year/2025", "")

    // Verify it now works (after fix) or properly errors
    if err != nil {
        t.Errorf("Should parse correctly: %v", err)
    }

    // Verify parsed params
    if len(params.ColorName) != 1 || params.ColorName[0] != "red" {
        t.Errorf("Color not parsed correctly")
    }
    if params.Year == nil || *params.Year != 2025 {
        t.Errorf("Year not parsed correctly")
    }
}
```

### Example 2: Testing Empty Results

**Log Entry:**
```
FACET_404: No results found - path=/color/blue query= params={ColorName:[blue] ...}
```

**Test:**
```go
func TestColorBlueNoResults(t *testing.T) {
    // Create test DB with photos but none blue
    db := createTestDB(t)
    defer db.Close()

    engine := query.NewEngine(db.DB)
    params := query.QueryParams{
        ColorName: []string{"blue"},
        Limit:     100,
    }

    result, err := engine.Query(params)

    // Should not error
    if err != nil {
        t.Fatalf("Query should not error: %v", err)
    }

    // Should return 0 results
    if result.Total != 0 {
        t.Errorf("Expected 0 results, got %d", result.Total)
    }

    // Facets should still compute
    facets, err := engine.ComputeFacets(params)
    if err != nil {
        t.Errorf("Facets should still compute: %v", err)
    }
    if facets == nil {
        t.Error("Facets should not be nil")
    }
}
```

### Example 3: Testing Facet Combinations

**Log Entry:**
```
FACET_404: No results found - path=/2025/color/red query= params={ColorName:[red] Year:2025 Limit:100 ...}
```

**Test:**
```go
func TestYearColorCombinationNoResults(t *testing.T) {
    // Create test DB with:
    // - Photos from 2025 (not red)
    // - Red photos (not from 2025)
    db := createTestDBWithData(t, []TestPhoto{
        {Year: 2025, Color: "blue"},
        {Year: 2024, Color: "red"},
    })
    defer db.Close()

    engine := query.NewEngine(db.DB)
    year := 2025
    params := query.QueryParams{
        Year:      &year,
        ColorName: []string{"red"},
        Limit:     100,
    }

    result, err := engine.Query(params)

    // Should not error
    if err != nil {
        t.Fatalf("Query should not error: %v", err)
    }

    // Should return 0 results (no red photos from 2025)
    if result.Total != 0 {
        t.Errorf("Expected 0 results, got %d", result.Total)
    }
}
```

## Monitoring Logs

To monitor facet 404s in production:

```bash
# Watch for all facet-related issues
tail -f server.log | grep "FACET_404\|FACET_ERROR"

# Count 404s by type
grep "FACET_404" server.log | cut -d: -f2 | sort | uniq -c

# Find most common failing URLs
grep "path=" server.log | sed 's/.*path=\([^ ]*\).*/\1/' | sort | uniq -c | sort -rn | head -20
```

## Common Scenarios

### Scenario: User clicks facet link, gets 0 results

**Log:**
```
FACET_404: No results found - path=/color/red/2025 query= params={ColorName:[red] Year:2025 ...}
```

**Diagnosis:** The facet UI showed "Red (3)" but combining with year gives 0 results.

**Action:** This is usually correct behavior - the facet counts are for the base query, not the intersection. Consider adding intersection counts to facet UI.

### Scenario: Malformed facet URL

**Log:**
```
FACET_404: URL parse failed - path=/color/red/invalid query= error=unrecognized segment 'invalid'
```

**Diagnosis:** A bug in facet URL generation created an invalid URL.

**Action:** Find and fix the FacetURLBuilder code that generated this URL. Add a test.

### Scenario: Valid URL but no handler

**Log:**
```
FACET_404: No route matched - path=/colors query=
```

**Diagnosis:** User tried to access a non-existent route (maybe typed URL).

**Action:** If this is a common request, consider adding the route. Otherwise, this is expected behavior.
