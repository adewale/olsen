# Query String-Based Faceted Navigation

## Overview

The faceted navigation system has been simplified to use **query string parameters** exclusively, eliminating the need to manually register routes for every facet value.

## Before vs After

### Old Approach (Path-Based)
```
/color/red           → Required route registration
/midday              → Required route registration
/blue_hour           → Required route registration
/2025                → Worked via catch-all regex
/camera/Canon/EOS-R5 → Required route registration
```

**Problems:**
- Had to manually register 20+ routes
- Easy to miss new facet values
- Caused 404 errors when routes were missing
- Complex URL parsing logic

### New Approach (Query String-Based)
```
/photos?color=red
/photos?time_of_day=midday
/photos?time_of_day=blue_hour
/photos?year=2025
/photos?camera_make=Canon&camera_model=EOS R5
/photos?year=2025&color=red&camera_make=Canon
```

**Benefits:**
- **Single route** (`/photos`) handles ALL combinations
- **Zero 404s** - invalid filters return 0 results, not 404
- **Guaranteed support** for all facet permutations
- **Simpler code** - no complex path parsing
- **Spec compliant** - matches industry standard URL patterns

## Implementation Changes

### 1. URL Mapper (`internal/query/url_mapper.go`)

**BuildPath**
```go
// Before: Complex priority-based path building
func (m *URLMapper) BuildPath(params QueryParams) string {
    if params.Year != nil {
        return fmt.Sprintf("/%d", *params.Year)
    }
    if len(params.CameraMake) > 0 {
        return fmt.Sprintf("/camera/%s", params.CameraMake[0])
    }
    // ... 50+ lines of path logic
}

// After: Always return /photos
func (m *URLMapper) BuildPath(params QueryParams) string {
    return "/photos"
}
```

**BuildQueryString**
```go
// All filters now included in query string
func (m *URLMapper) BuildQueryString(params QueryParams) string {
    values := url.Values{}

    if params.Year != nil {
        values.Set("year", strconv.Itoa(*params.Year))
    }
    for _, c := range params.ColorName {
        values.Add("color", c)
    }
    // ... all other filters

    return "?" + values.Encode()
}
```

**ParsePath**
```go
// Parse query string for /photos path
if path == "" || path == "photos" {
    if queryString != "" {
        values, err := url.ParseQuery(queryString)
        if err == nil {
            m.parseQueryString(values, &params)
        }
    }
    return params, nil
}
```

### 2. Server Routes (`internal/explorer/server.go`)

**Before:**
```go
func (s *Server) setupRoutes() {
    s.router.HandleFunc("/color/", s.handleQuery)
    s.router.HandleFunc("/morning", s.handleQuery)
    s.router.HandleFunc("/afternoon", s.handleQuery)
    s.router.HandleFunc("/evening", s.handleQuery)
    s.router.HandleFunc("/night", s.handleQuery)
    s.router.HandleFunc("/blue_hour", s.handleQuery)
    s.router.HandleFunc("/golden_hour_morning", s.handleQuery)
    s.router.HandleFunc("/golden_hour_evening", s.handleQuery)
    s.router.HandleFunc("/midday", s.handleQuery)
    s.router.HandleFunc("/spring", s.handleQuery)
    s.router.HandleFunc("/summer", s.handleQuery)
    s.router.HandleFunc("/fall", s.handleQuery)
    s.router.HandleFunc("/winter", s.handleQuery)
    s.router.HandleFunc("/wide", s.handleQuery)
    s.router.HandleFunc("/normal", s.handleQuery)
    s.router.HandleFunc("/telephoto", s.handleQuery)
    s.router.HandleFunc("/camera/", s.handleQuery)
    s.router.HandleFunc("/lens/", s.handleQuery)
    s.router.HandleFunc("/bursts", s.handleQuery)
    // ... complex catch-all logic
}
```

**After:**
```go
func (s *Server) setupRoutes() {
    s.router.HandleFunc("/photo/", s.handlePhotoDetail)
    s.router.HandleFunc("/api/thumbnail/", s.handleThumbnail)
    s.router.HandleFunc("/photos", s.handleQuery)  // Single route!
    s.router.HandleFunc("/", s.handleHome)
}
```

## URL Examples

### Single Filter
```
/photos?color=blue
/photos?year=2025
/photos?camera_make=Canon
/photos?time_of_day=golden_hour_morning
```

### Multiple Filters
```
/photos?year=2025&color=red
/photos?camera_make=Canon&camera_model=EOS R5
/photos?year=2025&color=red&time_of_day=morning
```

### Multi-Select Facets
```
/photos?color=red&color=blue
/photos?season=summer&season=fall
```

### With Pagination
```
/photos?year=2025&limit=50&offset=100
```

### Complex Combination
```
/photos?year=2025&color=red&camera_make=Canon&camera_model=EOS R5&time_of_day=morning&limit=100
```

## Testing Benefits

### Before
```go
func TestFacetNavigation(t *testing.T) {
    // Had to register routes manually
    // Couldn't test all combinations easily
    // 404s if route missing
}
```

### After
```go
func TestFacetNavigation(t *testing.T) {
    testCases := []struct{
        query string
        expectedCount int
    }{
        {"color=red", 13},
        {"year=2025", 13},
        {"color=red&year=2025", 4},
        {"time_of_day=midday", 4},
        // Can easily generate hundreds of combinations
    }

    for _, tc := range testCases {
        // Just make request - no route setup needed!
        result := query("/photos?" + tc.query)
        assert.Equal(t, tc.expectedCount, result.Total)
    }
}
```

## Verification

All facet state transitions verified:
```bash
$ ./olsen verify --db other_photos.db --facets

✅ Verification complete: 14/14 tests passed
🎉 All facet transitions verified successfully!
```

## Backwards Compatibility

Old path-based URLs still work via legacy parsing in `ParsePath`:
- `/2025` → parsed as year=2025
- `/camera/Canon/EOS-R5` → parsed as camera filters
- `/color/red` → parsed as color=red

But new facet URLs all use query strings for consistency and reliability.

## Migration Checklist

- [x] Simplify `BuildPath` to always return `/photos`
- [x] Move all filters to `BuildQueryString`
- [x] Update `ParsePath` to read query string for `/photos`
- [x] Simplify server routes to single `/photos` handler
- [x] Remove 20+ individual facet route registrations
- [x] Verify all state transitions pass
- [x] Document new URL format

## Future Improvements

1. **Redirect old URLs** - Add redirects from `/color/red` → `/photos?color=red`
2. **Remove legacy parsing** - Eventually remove path-based parsing entirely
3. **URL shortening** - Add optional short codes for common combinations
4. **Analytics** - Track which filter combinations are used most

## Key Insight

> "The simpler the routing, the more reliable the system."

By moving all filtering to query parameters, we eliminated an entire class of bugs (missing routes) and made the system infinitely more scalable.
