# Faceted Navigation Implementation Plan

**Version:** 1.0
**Date:** October 2025
**Status:** Planning

---

## Executive Summary

This document outlines improvements to Olsen's faceted navigation UI based on industry best practices and faceted classification principles. The current implementation treats facets as simple links, but proper faceted navigation requires understanding the **temporary hierarchy** created by user choices.

---

## Key Principles from Research

### 1. Faceted Classification vs. Hierarchical Classification

**Hierarchical systems** impose a fixed, single-path taxonomy:
```
Photos → 2024 → October → Canon → EOS R5
```

**Faceted systems** allow multiple orthogonal dimensions navigated in any order:
```
All Photos (12,453)
  ├─ [Color: Blue] → Blue Photos (1,245)
  ├─ [Year: 2024] → 2024 Photos (3,456)
  ├─ [Camera: Canon] → Canon Photos (5,678)
  └─ [Any combination of above]
```

**Key difference**: In faceted systems, each selection creates a **temporary hierarchy** based on user's navigation path, not a predefined structure.

### 2. The Temporary Hierarchy Concept

When users navigate facets, they create a journey:

```
Start: All Photos (12,453)
  ↓ [Select: Blue]
Blue Photos (1,245)
  ↓ [Select: 2024]
Blue Photos from 2024 (89)
  ↓ [Select: Canon EOS R5]
Blue Photos from 2024 with Canon EOS R5 (12)
```

Each step:
- **Narrows** the result set (AND conjunction)
- **Updates** facet counts for remaining options
- **Creates** a breadcrumb trail
- **Is reversible** at any level

### 3. Critical UX Patterns

#### Active Filters Display
- Show currently applied filters prominently
- Allow removal of individual filters (chips/pills)
- "Clear All" option when multiple filters active
- Each chip is clickable to remove that filter

#### Dynamic Facet Updates
- Facet counts must reflect current query context
- Show zero-count facets but disable/dim them
- Never hide facets completely (causes confusion)
- Update counts on every selection

#### Breadcrumb Navigation
- Show the path taken, not just "← Back"
- Each breadcrumb is clickable to undo filters after it
- Format: Home > Blue > 2024 > Canon EOS R5

#### Facet Refinement (NOT Replacement)
- Clicking a facet should **ADD** to current filters
- Should NOT replace existing filters
- Example: On `/color/blue`, clicking "2024" goes to `/color/blue/year/2024`

---

## Current Implementation Issues

### Issue 1: Facet URLs Replace Instead of Refine

**Current Behavior:**
```go
// In addFacetURLs(), each facet gets a fresh params
p := baseParams  // Starts fresh
p.ColorName = []string{facet.Value}  // Only sets one filter
```

**Problem:**
- On `/color/blue`, clicking "2024" goes to `/2024`
- Loses the color filter
- Breaks faceted navigation model

**Solution:**
Facets should append to existing filters:
```go
// When on /color/blue, clicking "2024" should:
// - Keep ColorName: ["blue"]
// - Add Year: 2024
// - Result: /color/blue/year/2024
```

### Issue 2: No Active Filter Display

**Current:**
- Users don't see what filters are applied
- No way to remove individual filters
- Must use "← Back" which removes everything

**Solution:**
Add active filter chips above results:
```
[× Blue] [× 2024] [× Canon] [Clear All]
```

### Issue 3: No Breadcrumb Trail

**Current:**
- Simple "← Back" link
- Doesn't show navigation path
- Unclear how to get back to specific point

**Solution:**
Replace with proper breadcrumbs:
```
Home > Blue Photos > 2024 > Canon EOS R5
```

### Issue 4: Facet Organization

**Current:**
- Facets shown in arbitrary order
- No grouping by type
- Equal visual weight for all facets

**Solution:**
Group and order by usage patterns:
1. **Temporal** (Year, Season, Time of Day)
2. **Visual** (Color, Shooting Condition)
3. **Equipment** (Camera, Lens, Focal Length)
4. **Special** (Bursts, Duplicates)

---

## Implementation Plan

### Phase 1: Fix Core Navigation (Priority: HIGH)

#### Task 1.1: Fix Facet URL Generation
**File:** `internal/query/facets.go:addFacetURLs()`

**Current Code:**
```go
if facets.ColorName != nil {
    for i := range facets.ColorName.Values {
        p := baseParams  // WRONG: Starts fresh
        p.ColorName = []string{facets.ColorName.Values[i].Value}
        facets.ColorName.Values[i].URL = mapper.BuildFullURL(p)
    }
}
```

**Fix:**
```go
if facets.ColorName != nil {
    for i := range facets.ColorName.Values {
        p := baseParams  // Start with current filters

        // ADD or REPLACE color in current context
        if facets.ColorName.Values[i].Selected {
            // Already selected - URL should REMOVE it
            p.ColorName = nil  // Remove color filter
        } else {
            // Not selected - URL should ADD it
            p.ColorName = []string{facets.ColorName.Values[i].Value}
        }

        facets.ColorName.Values[i].URL = mapper.BuildFullURL(p)
    }
}
```

**Apply to all facet types:** Year, Camera, Lens, TimeOfDay, Season, etc.

#### Task 1.2: Add Active Filters Display
**Files:**
- `internal/explorer/templates/grid.html`
- `internal/explorer/server.go:handleQuery()`

**Add to grid.html before results:**
```html
{{if .ActiveFilters}}
<div class="active-filters">
    <span>Active filters:</span>
    {{range .ActiveFilters}}
    <a href="{{.RemoveURL}}" class="filter-chip">
        × {{.Label}}
    </a>
    {{end}}
    {{if gt (len .ActiveFilters) 1}}
    <a href="/" class="clear-all">Clear all</a>
    {{end}}
</div>
{{end}}
```

**Add to server.go:**
```go
type ActiveFilter struct {
    Type      string  // "color", "year", "camera", etc.
    Label     string  // "Blue", "2024", "Canon EOS R5"
    RemoveURL string  // URL to remove this filter
}

func buildActiveFilters(params QueryParams, mapper *URLMapper) []ActiveFilter {
    filters := []ActiveFilter{}

    if len(params.ColorName) > 0 {
        for _, color := range params.ColorName {
            p := params
            p.ColorName = nil
            filters = append(filters, ActiveFilter{
                Type:      "color",
                Label:     strings.Title(color),
                RemoveURL: mapper.BuildFullURL(p),
            })
        }
    }

    if params.Year != nil {
        p := params
        p.Year = nil
        filters = append(filters, ActiveFilter{
            Type:      "year",
            Label:     fmt.Sprintf("%d", *params.Year),
            RemoveURL: mapper.BuildFullURL(p),
        })
    }

    // ... similar for all filter types

    return filters
}
```

#### Task 1.3: Add Proper Breadcrumbs
**File:** `internal/query/url_mapper.go`

**Current:**
```go
type Breadcrumb struct {
    Label string
    URL   string
}

func (m *URLMapper) BuildBreadcrumbs(params QueryParams) []Breadcrumb {
    // Partially implemented
}
```

**Enhance:**
```go
func (m *URLMapper) BuildBreadcrumbs(params QueryParams) []Breadcrumb {
    crumbs := []Breadcrumb{{Label: "Home", URL: "/"}}

    // Build breadcrumb trail based on ACTIVE filters
    // Order: Temporal → Visual → Equipment

    if params.Year != nil {
        crumbs = append(crumbs, Breadcrumb{
            Label: fmt.Sprintf("%d", *params.Year),
            URL:   fmt.Sprintf("/%d", *params.Year),
        })
    }

    if len(params.ColorName) > 0 {
        // Show first color in breadcrumb
        color := params.ColorName[0]
        // URL should include previous filters
        p := QueryParams{Year: params.Year}
        p.ColorName = []string{color}
        crumbs = append(crumbs, Breadcrumb{
            Label: strings.Title(color),
            URL:   m.BuildFullURL(p),
        })
    }

    if len(params.CameraMake) > 0 {
        // Include all previous filters
        p := QueryParams{
            Year:      params.Year,
            ColorName: params.ColorName,
            CameraMake: params.CameraMake,
        }
        crumbs = append(crumbs, Breadcrumb{
            Label: params.CameraMake[0],
            URL:   m.BuildFullURL(p),
        })
    }

    // Continue for all active filters...

    return crumbs
}
```

**Update template:**
```html
<div class="breadcrumbs">
    {{range $i, $crumb := .Breadcrumbs}}
        {{if $i}} › {{end}}
        <a href="{{$crumb.URL}}">{{$crumb.Label}}</a>
    {{end}}
</div>
```

---

### Phase 2: Improve Facet Presentation (Priority: MEDIUM)

#### Task 2.1: Facet State Indicators

**Add to grid.html:**
```html
<li class="facet-item {{if .Selected}}selected{{end}} {{if eq .Count 0}}disabled{{end}}">
    <a href="{{.URL}}">
        <span>
            {{if .Selected}}✓ {{end}}
            {{.Label}}
        </span>
        <span class="facet-count">({{.Count}})</span>
    </a>
</li>
```

**Add CSS:**
```css
.facet-item.selected a {
    color: #4a9eff;
    font-weight: bold;
}

.facet-item.disabled a {
    color: #444;
    cursor: not-allowed;
    pointer-events: none;
}

.facet-item.disabled .facet-count {
    color: #333;
}
```

#### Task 2.2: Reorder Facets by Logical Groups

**Modify templates/grid.html facet order:**
```html
<aside class="sidebar">
    <h3>Refine Results</h3>

    <!-- Temporal Facets -->
    {{if or .Facets.Year .Facets.Season .Facets.TimeOfDay}}
    <div class="facet-group">
        <h4>When</h4>
        {{template "yearFacet" .}}
        {{template "seasonFacet" .}}
        {{template "timeOfDayFacet" .}}
    </div>
    {{end}}

    <!-- Visual Facets -->
    {{if or .Facets.ColorName .Facets.ShootingCondition}}
    <div class="facet-group">
        <h4>Appearance</h4>
        {{template "colorFacet" .}}
        {{template "shootingConditionFacet" .}}
    </div>
    {{end}}

    <!-- Equipment Facets -->
    {{if or .Facets.Camera .Facets.Lens .Facets.FocalCategory}}
    <div class="facet-group">
        <h4>Equipment</h4>
        {{template "cameraFacet" .}}
        {{template "lensFacet" .}}
        {{template "focalCategoryFacet" .}}
    </div>
    {{end}}

    <!-- Special Facets -->
    {{if or .Facets.InBurst}}
    <div class="facet-group">
        <h4>Collections</h4>
        {{template "burstFacet" .}}
    </div>
    {{end}}
</aside>
```

#### Task 2.3: Collapsible Facet Sections

**Add expand/collapse without JavaScript:**
```html
<details class="facet-section" open>
    <summary class="facet-title">{{.Facets.ColorName.Label}}</summary>
    <ul class="facet-list">
        {{range .Facets.ColorName.Values}}
        <li class="facet-item">
            <a href="{{.URL}}">
                <span>{{.Label}}</span>
                <span class="facet-count">({{.Count}})</span>
            </a>
        </li>
        {{end}}
    </ul>
</details>
```

**CSS for details/summary:**
```css
details.facet-section {
    margin-bottom: 1rem;
}

summary.facet-title {
    cursor: pointer;
    font-weight: bold;
    padding: 0.5rem 0;
    list-style: none;  /* Hide default arrow */
}

summary.facet-title::-webkit-details-marker {
    display: none;  /* Hide in webkit */
}

summary.facet-title::before {
    content: '▸ ';
    display: inline-block;
    transition: transform 0.2s;
}

details[open] summary.facet-title::before {
    transform: rotate(90deg);
}
```

---

### Phase 3: Advanced Features (Priority: LOW)

#### Task 3.1: Multi-Select Within Facets

**Allow OR logic within a facet:**
- Select "morning" AND "evening" → photos from either time
- Requires query engine update to support multi-value facets
- More complex URL structure

**Implementation:**
```go
// QueryParams already supports this:
TimeOfDay []string  // Can have multiple values

// URL format: /time/morning,evening
// Or: /time/morning/time/evening
```

**UI Change:**
- Checkboxes instead of links
- "Apply" button to execute multi-select
- Requires minimal JavaScript OR form submission

#### Task 3.2: Facet Search/Filter

**For facets with many values (cameras, lenses):**
```html
<div class="facet-section">
    <div class="facet-title">Cameras</div>
    <input type="text" class="facet-search" placeholder="Search cameras...">
    <ul class="facet-list">
        {{range .Facets.Camera.Values}}
        <li class="facet-item" data-label="{{.Label}}">
            <a href="{{.URL}}">
                <span>{{.Label}}</span>
                <span class="facet-count">({{.Count}})</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
```

**Requires JavaScript:**
```javascript
document.querySelector('.facet-search').addEventListener('input', (e) => {
    const query = e.target.value.toLowerCase();
    document.querySelectorAll('.facet-item').forEach(item => {
        const label = item.dataset.label.toLowerCase();
        item.style.display = label.includes(query) ? 'block' : 'none';
    });
});
```

#### Task 3.3: Range Sliders for Continuous Facets

**For ISO, Aperture, Focal Length:**
- Visual slider showing distribution
- Select range of values
- Better than discrete options

**Requires:**
- Query engine support for range queries (already exists)
- JavaScript for interactive slider
- Or: Simple form with min/max inputs

---

### Phase 4: Polish & Performance (Priority: MEDIUM)

#### Task 4.1: URL Strategy Consistency

**Current URL patterns:**
```
/color/blue          → Color filter
/2024                → Year filter
/camera/Canon/EOS-R5 → Camera filter
```

**Issue**: Mixing path segments and query strings

**Decision Options:**

**Option A: Path-based (clean URLs)**
```
/color/blue/year/2024/camera/Canon-EOS-R5
```
Pros: Clean, bookmarkable, SEO-friendly
Cons: Complex parsing, order matters

**Option B: Query-string (flexible)**
```
/?color=blue&year=2024&camera=Canon-EOS-R5
```
Pros: Order-independent, easy to parse
Cons: Less clean, harder to read

**Option C: Hybrid (recommended)**
```
/color/blue?year=2024&camera=Canon-EOS-R5
```
Pros: Clean primary path, flexible additional filters
Cons: Some parsing complexity

**Recommendation**: Stick with current path-based approach but ensure consistency in URL generation.

#### Task 4.2: Performance - Dynamic Updates

**Current**: Full page reload on facet click

**Improvement**: AJAX updates (optional)
- Fetch results via JSON API
- Update grid without page reload
- Update URL with history.pushState()
- Show loading states

**Trade-off**: Adds JavaScript complexity vs. better UX

#### Task 4.3: Mobile Optimization

**Current**: Sidebar always visible

**Improvement**: Tray overlay on mobile
- "Filters" button shows count of active filters
- Clicking opens tray overlay with facets
- Apply filters updates results
- Close tray to see results

**Implementation:**
```html
<button class="filters-toggle" aria-expanded="false">
    Filters
    {{if gt (len .ActiveFilters) 0}}
        <span class="filter-count">{{len .ActiveFilters}}</span>
    {{end}}
</button>

<div class="filters-tray" hidden>
    <div class="tray-header">
        <h3>Filters</h3>
        <button class="tray-close">×</button>
    </div>
    <div class="tray-content">
        {{template "facets" .}}
    </div>
</div>
```

---

## Updated Faceted Navigation Spec

### Home Page Facets

**Purpose**: Starting point for exploration

**Display**: Grid layout showing all available facets

**Behavior**: Clicking a facet takes you to initial filter view

**Example:**
```
Explore by:

[Colors]          [Years]           [Cameras]
Red (234)         2024 (456)        Canon (789)
Blue (123)        2023 (321)        Nikon (234)
...               ...               ...

[Time of Day]     [Season]          [Focal Length]
Morning (345)     Summer (234)      Wide (345)
...               ...               ...
```

### Grid View Facets

**Purpose**: Refine current results

**Display**: Sidebar with grouped facets

**Behavior**:
- Facets show counts based on current filters
- Clicking adds/removes filter from current query
- Selected facets are highlighted
- Zero-count facets are disabled but visible

**Example:**
```
Active Filters:
[× Blue] [× 2024] [Clear All]

Breadcrumbs:
Home › Blue › 2024

Refine Results:

▾ When
  □ 2023 (45)
  ☑ 2024 (89)  ← selected
  □ 2025 (12)

▾ Appearance
  ☑ Blue (89)  ← selected
  □ Red (12)
  □ Green (0)  ← disabled

▾ Equipment
  □ Canon EOS R5 (34)
  □ Canon EOS R6 (28)
  □ Nikon Z9 (15)
```

---

## Success Metrics

### User Experience
- Users can navigate facets in any order
- Current filters are always visible
- Easy to remove individual filters
- Clear breadcrumb trail shows journey
- Zero-result facets don't trap users

### Technical Performance
- Facet counts update correctly
- URLs are clean and bookmarkable
- Page loads < 500ms
- No broken navigation paths

### Code Quality
- Facet URL generation is consistent
- Active filter extraction is DRY
- Templates are maintainable
- Logic is testable

---

## Testing Plan

### Manual Testing Scenarios

1. **Basic Navigation**
   - Start at home
   - Click "Blue" → verify URL and results
   - Click "2024" → verify both filters applied
   - Remove "Blue" filter → verify only 2024 remains

2. **Breadcrumb Navigation**
   - Apply 3 filters: Blue → 2024 → Canon
   - Click "2024" breadcrumb → verify Canon filter removed
   - Verify results and facet counts update

3. **Zero-Count Facets**
   - Apply filters until some facets show (0)
   - Verify disabled state
   - Verify cannot click disabled facets

4. **Mobile View**
   - Test on narrow viewport
   - Verify filters tray works
   - Verify active filters visible

### Automated Tests

```go
func TestFacetURLGeneration(t *testing.T) {
    // Test that facet URLs preserve existing filters
}

func TestActiveFilters(t *testing.T) {
    // Test extraction of active filters from params
}

func TestBreadcrumbs(t *testing.T) {
    // Test breadcrumb generation for various filter combinations
}

func TestFacetRefinement(t *testing.T) {
    // Test that clicking facet adds to (not replaces) filters
}
```

---

## Implementation Timeline

### Week 1: Core Fixes
- ✅ Fix facet URL generation (Task 1.1)
- ✅ Add active filter display (Task 1.2)
- ✅ Add proper breadcrumbs (Task 1.3)
- ✅ Test core navigation flows

### Week 2: UI Improvements
- ⏳ Add facet state indicators (Task 2.1)
- ⏳ Reorder and group facets (Task 2.2)
- ⏳ Add collapsible sections (Task 2.3)
- ⏳ Mobile optimization (Task 4.3)

### Week 3: Polish
- ⏳ URL strategy consistency (Task 4.1)
- ⏳ Performance optimization
- ⏳ Comprehensive testing
- ⏳ Documentation updates

### Future
- ⭕ Multi-select facets (Task 3.1)
- ⭕ Facet search (Task 3.2)
- ⭕ Range sliders (Task 3.3)
- ⭕ AJAX updates (Task 4.2)

---

## References

- [Faceted Navigation: Definition, Examples & Tips - OptiMonk](https://www.optimonk.com/16-tips-effective-user-friendly-faceted-navigation/)
- [Faceted Classification - Wikipedia](https://en.wikipedia.org/wiki/Faceted_classification)
- [Best Practices for Faceted Search Filters - UXmatters](https://www.uxmatters.com/mt/archives/2009/09/best-practices-for-designing-faceted-search-filters.php)
- [Mobile Faceted Search - Nielsen Norman Group](https://www.nngroup.com/articles/mobile-faceted-search/)

---

**END OF PLAN**
