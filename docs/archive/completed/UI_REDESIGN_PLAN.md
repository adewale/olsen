# Olsen UI Redesign Implementation Plan

**Version**: 1.0
**Date**: 2025-10-06
**Based on**: `faceted_navigation.spec` v2.1 + `olsen_faceted_ui_mock.png`

---

## Executive Summary

This plan outlines the step-by-step implementation of the new faceted navigation UI based on the refined spec. The redesign moves from a left sidebar to a **right-rail layout** with improved facet organization, active filter chips, and mobile-friendly tray interface.

**Key Changes**:
- Right-rail facets (instead of left sidebar)
- Prominent active filter chips with removal
- Grouped and collapsible facet sections
- Search bar + sort dropdown in top sticky bar
- Mobile filter tray with badge count
- Map/GPS explicitly out of scope

---

## Current State Analysis

### What We Have (from `grid.html`)

**Layout**:
- Left sidebar (250px) with facets
- Main content area with photo grid
- Basic breadcrumbs
- Active filter chips ✅

**Facets Currently Shown**:
- Colour (4 shown in UI)
- Year (4 shown)
- Camera (3 shown)
- Time of Day (4 shown)

**What's Missing**:
- Search bar
- Sort dropdown (exists in backend but not prominent in UI)
- Facet grouping/categorization
- Collapsible sections
- Mobile tray
- Many facet types from spec (Lens, Season, Orientation, White Balance, etc.)

---

## Target State (from Mock)

### Layout Structure

```
┌────────────────────────────────────────────────────────────────┐
│ [Olsen]  [Search photos, cameras, lenses...]    [Sort: Date ▾] │ ← Sticky top bar
├────────────────────────────────────────────────────────────────┤
│ [Colour: Red ×] [Year: 2025 ×] [Orientation: Portrait ×]      │ ← Active filter chips
├─────────────────────────────────────────┬──────────────────────┤
│                                         │ Filters              │
│  [Photo Grid - 3-4 columns]            │                      │
│                                         │ ▾ Time               │
│  ┌────────┐ ┌────────┐ ┌────────┐     │   Year               │
│  │        │ │        │ │        │     │   • 2025      1240   │
│  │ Photo  │ │ Photo  │ │ Photo  │     │   • 2024      3812   │
│  │        │ │        │ │        │     │   • Unknown     98   │
│  └────────┘ └────────┘ └────────┘     │   Month...           │
│                                         │   Time of day        │
│  ┌────────┐ ┌────────┐ ┌────────┐     │   [dawn][golden AM]  │
│  │ Photo  │ │ Photo  │ │ Photo  │     │                      │
│  └────────┘ └────────┘ └────────┘     │ ▾ Equipment          │
│                                         │   Camera             │
│  [Load more / Pagination]              │   • Canon EOS R5 420 │
│                                         │   • Fuji X-T5    288 │
│                                         │   Lens               │
│                                         │   • RF 24-70mm   266 │
│                                         │                      │
│                                         │ ▾ Colour             │
│                                         │   ○ Red              │
│                                         │   ○ Orange           │
│                                         │   ○ Yellow           │
│                                         │                      │
│                                         │ ▾ Composition        │
│                                         │   ☑ Portrait         │
│                                         │   ☐ Landscape        │
│                                         │                      │
│                                         │ ▾ Capture conditions │
│                                         │   ☑ Auto             │
│                                         │   ☐ Daylight         │
│                                         │   [Flash fired]      │
│                                         │                      │
│                                         │ ▾ Bursts             │
│                                         │   ☑ In burst         │
└─────────────────────────────────────────┴──────────────────────┘
```

---

## Implementation Phases

### Phase 1: Core Layout Restructure (Priority: HIGH)

**Goal**: Flip layout from left-sidebar to right-rail

#### Task 1.1: Update Grid Template Structure

**File**: `internal/explorer/templates/grid.html`

**Changes**:
```html
<!-- NEW STRUCTURE -->
<div class="page-container">
    <!-- Sticky top bar -->
    <header class="top-bar">
        <div class="top-bar-left">
            <form class="search-form" action="/photos" method="GET">
                <input type="search" name="q" placeholder="Search photos, cameras, lenses..."
                       value="{{.SearchQuery}}" class="search-input">
            </form>
        </div>
        <div class="top-bar-right">
            <select name="sort" class="sort-dropdown" onchange="handleSortChange(this)">
                <option value="date_taken:desc" {{if eq .SortBy "date_taken"}}selected{{end}}>Date ▾</option>
                <option value="date_taken:asc">Date ▴</option>
                <option value="camera:asc">Camera A-Z</option>
                <option value="focal_length:asc">Focal Length</option>
                <option value="iso:desc">ISO High-Low</option>
                <option value="aperture:asc">Aperture f/1.4→f/22</option>
            </select>
            <!-- Future: Save view, Share buttons -->
        </div>
    </header>

    <!-- Active filter chips -->
    {{if .ActiveFilters}}
    <div class="active-filters-bar">
        {{range .ActiveFilters}}
        <a href="{{.RemoveURL}}" class="filter-chip">
            {{.Type}}: {{.Label}} <span class="remove-icon">×</span>
        </a>
        {{end}}
        {{if gt (len .ActiveFilters) 1}}
        <a href="/photos" class="clear-all-btn">Clear all</a>
        {{end}}
    </div>
    {{end}}

    <!-- Main content area -->
    <div class="content-wrapper">
        <!-- Photo grid (left/center) -->
        <main class="results-grid">
            <div class="results-header">
                <h1 class="results-title">{{.Title}}</h1>
                <span class="results-count">{{.TotalCount}} photos</span>
            </div>

            <div class="photo-grid">
                {{range .Photos}}
                <a href="/photo/{{.ID}}" class="photo-card">
                    <img src="/api/thumbnail/{{.ID}}/256" alt="Photo" loading="lazy">
                    <div class="card-meta">
                        <div class="card-camera">{{.CameraMake}} {{.CameraModel}}</div>
                        <div class="card-specs">{{.FocalLength}}mm · f/{{.Aperture}} · ISO {{.ISO}}</div>
                    </div>
                </a>
                {{end}}
            </div>

            <!-- Pagination -->
            {{if or .PrevPage .NextPage}}
            <nav class="pagination">
                {{if .PrevPage}}<a href="{{.PrevPage}}" class="page-prev">← Previous</a>{{end}}
                <span class="page-current">Page {{.Page}}</span>
                {{if .NextPage}}<a href="{{.NextPage}}" class="page-next">Next →</a>{{end}}
            </nav>
            {{end}}
        </main>

        <!-- Facet rail (right) -->
        <aside class="facet-rail">
            <h2 class="rail-title">Filters</h2>

            <!-- Facet sections will go here -->
            {{template "facetSections" .}}
        </aside>
    </div>

    <!-- Mobile filter button (sticky bottom) -->
    <button class="mobile-filter-btn" onclick="toggleFilterTray()">
        Filters
        {{if .ActiveFilters}}
        <span class="filter-badge">{{len .ActiveFilters}}</span>
        {{end}}
    </button>

    <!-- Mobile filter tray -->
    <div class="filter-tray" id="filterTray" hidden>
        <div class="tray-header">
            <h2>Filters</h2>
            <button class="tray-close" onclick="closeFilterTray()">×</button>
        </div>
        <div class="tray-content">
            {{template "facetSections" .}}
        </div>
        <div class="tray-footer">
            <a href="/photos" class="tray-reset">Reset all</a>
            <button class="tray-apply" onclick="closeFilterTray()">Apply</button>
        </div>
    </div>
</div>
```

#### Task 1.2: Create Facet Sections Template

**New template**: `facetSections` (defined within `grid.html` or separate file)

```html
{{define "facetSections"}}

<!-- TIME -->
<details class="facet-section" open>
    <summary class="facet-header">Time</summary>
    <div class="facet-body">
        {{template "yearFacet" .}}
        {{template "monthFacet" .}}
        {{template "dayFacet" .}}
        {{template "timeOfDayFacet" .}}
        {{template "seasonFacet" .}}
    </div>
</details>

<!-- EQUIPMENT -->
<details class="facet-section" open>
    <summary class="facet-header">Equipment</summary>
    <div class="facet-body">
        {{template "cameraFacet" .}}
        {{template "lensFacet" .}}
        {{template "lensMakeFacet" .}}
    </div>
</details>

<!-- COLOUR -->
<details class="facet-section" open>
    <summary class="facet-header">Colour</summary>
    <div class="facet-body">
        {{template "colourSwatchesFacet" .}}
        <!-- Future: Advanced HSL sliders -->
    </div>
</details>

<!-- COMPOSITION & ORIENTATION -->
<details class="facet-section">
    <summary class="facet-header">Composition & Orientation</summary>
    <div class="facet-body">
        {{template "orientationFacet" .}}
        <!-- Future: Width/Height ranges -->
    </div>
</details>

<!-- CAPTURE CONDITIONS -->
<details class="facet-section">
    <summary class="facet-header">Capture Conditions</summary>
    <div class="facet-body">
        {{template "whiteBalanceFacet" .}}
        {{template "flashFiredFacet" .}}
    </div>
</details>

<!-- BURSTS -->
<details class="facet-section">
    <summary class="facet-header">Bursts</summary>
    <div class="facet-body">
        {{template "inBurstFacet" .}}
        {{template "burstGroupFacet" .}}
        {{template "isRepresentativeFacet" .}}
    </div>
</details>

<!-- FILE / SPACE (collapsed by default) -->
<details class="facet-section">
    <summary class="facet-header">File / Space</summary>
    <div class="facet-body">
        {{template "colourSpaceFacet" .}}
        {{template "isoRangeFacet" .}}
        {{template "apertureRangeFacet" .}}
        {{template "focalLengthRangeFacet" .}}
    </div>
</details>

{{end}}
```

#### Task 1.3: Update CSS for New Layout

**File**: `internal/explorer/templates/grid.html` `<style>` section (or extract to separate CSS file)

```css
/* Top bar (sticky) */
.top-bar {
    position: sticky;
    top: 0;
    z-index: 100;
    background: #000;
    border-bottom: 1px solid #333;
    padding: 1rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
}

.search-form {
    flex: 1;
    max-width: 600px;
}

.search-input {
    width: 100%;
    padding: 0.5rem 1rem;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 4px;
    color: #fff;
    font-size: 1rem;
}

.sort-dropdown {
    padding: 0.5rem 1rem;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 4px;
    color: #fff;
    cursor: pointer;
}

/* Active filter chips bar */
.active-filters-bar {
    background: #0a0a0a;
    border-bottom: 1px solid #333;
    padding: 1rem 2rem;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
}

.filter-chip {
    background: #1a1a1a;
    border: 1px solid #4a9eff;
    border-radius: 16px;
    padding: 0.25rem 0.75rem;
    color: #4a9eff;
    text-decoration: none;
    font-size: 0.9rem;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    transition: all 0.2s;
}

.filter-chip:hover {
    background: #4a9eff;
    color: #000;
}

.remove-icon {
    font-weight: bold;
    font-size: 1.2rem;
}

.clear-all-btn {
    background: #333;
    border: 1px solid #666;
    border-radius: 16px;
    padding: 0.25rem 0.75rem;
    color: #fff;
    text-decoration: none;
    font-size: 0.9rem;
    transition: all 0.2s;
}

.clear-all-btn:hover {
    background: #666;
}

/* Main content wrapper */
.content-wrapper {
    display: flex;
    gap: 2rem;
    padding: 2rem;
    max-width: 1800px;
    margin: 0 auto;
}

/* Results grid (left/center) */
.results-grid {
    flex: 1;
    min-width: 0; /* Important for flex shrinking */
}

.results-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.results-title {
    font-size: 1.5rem;
    margin: 0;
}

.results-count {
    color: #666;
    font-size: 1rem;
}

.photo-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1rem;
}

.photo-card {
    position: relative;
    display: block;
    border-radius: 4px;
    overflow: hidden;
    background: #1a1a1a;
    transition: transform 0.2s;
}

.photo-card:hover {
    transform: scale(1.02);
}

.photo-card img {
    width: 100%;
    height: auto;
    display: block;
}

.card-meta {
    padding: 0.5rem;
    font-size: 0.8rem;
}

.card-camera {
    color: #fff;
    margin-bottom: 0.25rem;
}

.card-specs {
    color: #666;
}

/* Facet rail (right) */
.facet-rail {
    width: 280px;
    flex-shrink: 0;
    position: sticky;
    top: 80px; /* Below sticky top bar */
    max-height: calc(100vh - 100px);
    overflow-y: auto;
}

.rail-title {
    font-size: 1.2rem;
    margin-bottom: 1rem;
}

/* Facet sections */
.facet-section {
    border-bottom: 1px solid #333;
    margin-bottom: 1rem;
}

.facet-header {
    cursor: pointer;
    font-weight: bold;
    padding: 0.75rem 0;
    list-style: none;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.facet-header::after {
    content: '▾';
    transition: transform 0.2s;
}

.facet-section:not([open]) .facet-header::after {
    transform: rotate(-90deg);
}

.facet-body {
    padding-bottom: 1rem;
}

/* Mobile filter button (hidden on desktop) */
.mobile-filter-btn {
    display: none;
    position: fixed;
    bottom: 1rem;
    right: 1rem;
    z-index: 1000;
    background: #4a9eff;
    color: #000;
    border: none;
    border-radius: 24px;
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    font-weight: bold;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(74, 158, 255, 0.5);
}

.filter-badge {
    background: #000;
    color: #4a9eff;
    border-radius: 50%;
    width: 20px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    margin-left: 0.5rem;
}

/* Mobile filter tray */
.filter-tray {
    position: fixed;
    inset: 0;
    z-index: 2000;
    background: #000;
    display: flex;
    flex-direction: column;
}

.tray-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid #333;
}

.tray-close {
    background: none;
    border: none;
    color: #fff;
    font-size: 2rem;
    cursor: pointer;
    padding: 0;
    width: 40px;
    height: 40px;
}

.tray-content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
}

.tray-footer {
    display: flex;
    gap: 1rem;
    padding: 1rem;
    border-top: 1px solid #333;
}

.tray-reset {
    flex: 1;
    padding: 0.75rem;
    text-align: center;
    border: 1px solid #666;
    border-radius: 4px;
    color: #fff;
    text-decoration: none;
}

.tray-apply {
    flex: 1;
    padding: 0.75rem;
    background: #4a9eff;
    border: none;
    border-radius: 4px;
    color: #000;
    font-weight: bold;
    cursor: pointer;
}

/* Mobile responsive */
@media (max-width: 768px) {
    .content-wrapper {
        flex-direction: column;
        padding: 1rem;
    }

    .facet-rail {
        display: none; /* Hidden on mobile, use tray instead */
    }

    .mobile-filter-btn {
        display: block;
    }

    .photo-grid {
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
        gap: 0.5rem;
    }

    .top-bar {
        flex-direction: column;
        gap: 0.5rem;
    }

    .search-form {
        max-width: none;
        width: 100%;
    }
}
```

---

### Phase 2: Individual Facet Components (Priority: HIGH)

#### Task 2.1: Year Facet (List with counts)

```html
{{define "yearFacet"}}
{{if .Facets.Year}}
<div class="facet-group">
    <div class="facet-label">Year</div>
    <ul class="facet-list">
        {{range .Facets.Year.Values}}
        <li class="facet-item {{if .Selected}}selected{{end}}">
            <a href="{{.URL}}">
                <span class="facet-name">
                    {{if .Selected}}✓ {{end}}{{.Label}}
                </span>
                <span class="facet-count">{{.Count}}</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
{{end}}
{{end}}
```

CSS:
```css
.facet-group {
    margin-bottom: 1rem;
}

.facet-label {
    font-size: 0.85rem;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.5rem;
}

.facet-list {
    list-style: none;
    padding: 0;
    margin: 0;
}

.facet-item {
    margin-bottom: 0.25rem;
}

.facet-item a {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    color: #888;
    text-decoration: none;
    transition: all 0.2s;
}

.facet-item a:hover {
    background: #1a1a1a;
    color: #fff;
}

.facet-item.selected a {
    color: #4a9eff;
    font-weight: bold;
}

.facet-count {
    font-size: 0.85rem;
    color: #666;
}
```

#### Task 2.2: Month/Day Facets (Progressive Disclosure)

```html
{{define "monthFacet"}}
{{if .Facets.Month}}
{{if .Params.Year}}
<div class="facet-group facet-sub">
    <div class="facet-label">Month</div>
    <ul class="facet-list">
        {{range .Facets.Month.Values}}
        <li class="facet-item {{if .Selected}}selected{{end}}">
            <a href="{{.URL}}">
                <span class="facet-name">{{if .Selected}}✓ {{end}}{{.Label}}</span>
                <span class="facet-count">{{.Count}}</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
{{end}}
{{end}}
{{end}}

{{define "dayFacet"}}
{{if .Facets.Day}}
{{if .Params.Month}}
<div class="facet-group facet-sub">
    <div class="facet-label">Day</div>
    <ul class="facet-list">
        {{range .Facets.Day.Values}}
        <li class="facet-item {{if .Selected}}selected{{end}}">
            <a href="{{.URL}}">
                <span class="facet-name">{{if .Selected}}✓ {{end}}{{.Label}}</span>
                <span class="facet-count">{{.Count}}</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
{{end}}
{{end}}
{{end}}
```

CSS:
```css
.facet-sub {
    margin-left: 1rem;
    padding-left: 0.5rem;
    border-left: 2px solid #333;
}
```

#### Task 2.3: Time of Day Facet (Chip-style, multi-select)

```html
{{define "timeOfDayFacet"}}
{{if .Facets.TimeOfDay}}
<div class="facet-group">
    <div class="facet-label">Time of day</div>
    <div class="facet-chips">
        {{range .Facets.TimeOfDay.Values}}
        <a href="{{.URL}}" class="facet-chip-btn {{if .Selected}}selected{{end}}">
            {{.Label}}
        </a>
        {{end}}
    </div>
</div>
{{end}}
{{end}}
```

CSS:
```css
.facet-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.facet-chip-btn {
    padding: 0.25rem 0.75rem;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 16px;
    color: #888;
    text-decoration: none;
    font-size: 0.85rem;
    transition: all 0.2s;
}

.facet-chip-btn:hover {
    border-color: #4a9eff;
    color: #4a9eff;
}

.facet-chip-btn.selected {
    background: #4a9eff;
    border-color: #4a9eff;
    color: #000;
    font-weight: bold;
}
```

#### Task 2.4: Season Facet (Similar to Time of Day)

```html
{{define "seasonFacet"}}
{{if .Facets.Season}}
<div class="facet-group">
    <div class="facet-label">Season</div>
    <div class="facet-chips">
        {{range .Facets.Season.Values}}
        <a href="{{.URL}}" class="facet-chip-btn {{if .Selected}}selected{{end}}">
            {{.Label}}
        </a>
        {{end}}
    </div>
</div>
{{end}}
{{end}}
```

#### Task 2.5: Camera Facet (List with search)

```html
{{define "cameraFacet"}}
{{if .Facets.Camera}}
<div class="facet-group">
    <div class="facet-label">Camera</div>
    {{if gt (len .Facets.Camera.Values) 5}}
    <input type="text" class="facet-search" placeholder="Search cameras..."
           onkeyup="filterFacetList(this, 'camera-list')">
    {{end}}
    <ul class="facet-list" id="camera-list">
        {{range .Facets.Camera.Values}}
        <li class="facet-item {{if .Selected}}selected{{end}}" data-label="{{.Label}}">
            <a href="{{.URL}}">
                <span class="facet-name">{{if .Selected}}✓ {{end}}{{.Label}}</span>
                <span class="facet-count">{{.Count}}</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
{{end}}
{{end}}
```

CSS:
```css
.facet-search {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 0.5rem;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 4px;
    color: #fff;
    font-size: 0.85rem;
}

.facet-search:focus {
    border-color: #4a9eff;
    outline: none;
}
```

JavaScript (minimal, for search):
```javascript
function filterFacetList(input, listId) {
    const query = input.value.toLowerCase();
    const list = document.getElementById(listId);
    const items = list.querySelectorAll('.facet-item');

    items.forEach(item => {
        const label = item.dataset.label.toLowerCase();
        item.style.display = label.includes(query) ? '' : 'none';
    });
}
```

#### Task 2.6: Lens Facet (Multi-select list with search)

```html
{{define "lensFacet"}}
{{if .Facets.Lens}}
<div class="facet-group">
    <div class="facet-label">Lens</div>
    {{if gt (len .Facets.Lens.Values) 5}}
    <input type="text" class="facet-search" placeholder="Search lenses..."
           onkeyup="filterFacetList(this, 'lens-list')">
    {{end}}
    <ul class="facet-list" id="lens-list">
        {{range .Facets.Lens.Values}}
        <li class="facet-item {{if .Selected}}selected{{end}}" data-label="{{.Label}}">
            <a href="{{.URL}}">
                <span class="facet-name">{{if .Selected}}✓ {{end}}{{.Label}}</span>
                <span class="facet-count">{{.Count}}</span>
            </a>
        </li>
        {{end}}
    </ul>
</div>
{{end}}
{{end}}
```

#### Task 2.7: Colour Facet (Radio buttons with swatches)

```html
{{define "colourSwatchesFacet"}}
{{if .Facets.ColourName}}
<div class="facet-group">
    <div class="facet-label">Colour</div>
    <div class="colour-swatches">
        {{range .Facets.ColourName.Values}}
        <a href="{{.URL}}" class="colour-swatch {{if .Selected}}selected{{end}}"
           title="{{.Label}} ({{.Count}})">
            <span class="swatch-circle" style="background-color: {{colourToHex .Value}}"></span>
            <span class="swatch-label">{{.Label}}</span>
        </a>
        {{end}}
    </div>
</div>
{{end}}
{{end}}
```

CSS:
```css
.colour-swatches {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
}

.colour-swatch {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem;
    border: 1px solid #333;
    border-radius: 4px;
    text-decoration: none;
    color: #888;
    transition: all 0.2s;
}

.colour-swatch:hover {
    border-color: #4a9eff;
}

.colour-swatch.selected {
    border-color: #4a9eff;
    background: #1a1a1a;
}

.swatch-circle {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 1px solid #666;
}

.swatch-label {
    font-size: 0.85rem;
}
```

Helper function needed in Go:
```go
// Add to template functions
"colourToHex": func(colorName string) string {
    colors := map[string]string{
        "red":    "#ff0000",
        "orange": "#ff8800",
        "yellow": "#ffff00",
        "green":  "#00ff00",
        "blue":   "#0088ff",
        "purple": "#8800ff",
        "pink":   "#ff00ff",
        "grey":   "#808080",
        "black":  "#000000",
        "white":  "#ffffff",
    }
    if hex, ok := colors[colorName]; ok {
        return hex
    }
    return "#888888"
},
```

#### Task 2.8: Orientation Facet (Checkboxes)

```html
{{define "orientationFacet"}}
{{if .Facets.Orientation}}
<div class="facet-group">
    <div class="facet-label">Orientation</div>
    <div class="facet-checkboxes">
        {{range .Facets.Orientation.Values}}
        <label class="facet-checkbox">
            <input type="checkbox" {{if .Selected}}checked{{end}}
                   onclick="location.href='{{.URL}}'">
            <span class="checkbox-label">{{.Label}}</span>
            <span class="facet-count">{{.Count}}</span>
        </label>
        {{end}}
    </div>
</div>
{{end}}
{{end}}
```

CSS:
```css
.facet-checkboxes {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.facet-checkbox {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    padding: 0.25rem;
}

.facet-checkbox input[type="checkbox"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
}

.checkbox-label {
    flex: 1;
    color: #888;
}

.facet-checkbox:has(input:checked) .checkbox-label {
    color: #4a9eff;
    font-weight: bold;
}
```

#### Task 2.9: Boolean Facets (Toggle switches)

```html
{{define "flashFiredFacet"}}
{{if .Facets.FlashFired}}
<div class="facet-group">
    <label class="facet-toggle">
        <span class="toggle-label">Flash fired</span>
        <input type="checkbox" {{if .Selected}}checked{{end}}
               onclick="location.href='{{.URL}}'">
        <span class="toggle-switch"></span>
    </label>
</div>
{{end}}
{{end}}

{{define "inBurstFacet"}}
{{if .Facets.InBurst}}
<div class="facet-group">
    {{range .Facets.InBurst.Values}}
    {{if eq .Value "yes"}}
    <label class="facet-toggle">
        <span class="toggle-label">In burst</span>
        <input type="checkbox" {{if .Selected}}checked{{end}}
               onclick="location.href='{{.URL}}'">
        <span class="toggle-switch"></span>
    </label>
    {{end}}
    {{end}}
</div>
{{end}}
{{end}}
```

CSS (toggle switch):
```css
.facet-toggle {
    display: flex;
    justify-content: space-between;
    align-items: center;
    cursor: pointer;
    padding: 0.5rem 0;
}

.toggle-label {
    color: #888;
}

.facet-toggle input[type="checkbox"] {
    display: none;
}

.toggle-switch {
    position: relative;
    width: 40px;
    height: 20px;
    background: #333;
    border-radius: 10px;
    transition: background 0.2s;
}

.toggle-switch::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    background: #666;
    border-radius: 50%;
    transition: all 0.2s;
}

.facet-toggle input[type="checkbox"]:checked + .toggle-switch {
    background: #4a9eff;
}

.facet-toggle input[type="checkbox"]:checked + .toggle-switch::after {
    left: 22px;
    background: #fff;
}
```

#### Task 2.10: Range Facets (Sliders - Future Enhancement)

For now, show as text with values. Full slider implementation can be Phase 3.

```html
{{define "isoRangeFacet"}}
{{if .Params.ISOMin}}
<div class="facet-group">
    <div class="facet-label">ISO Range</div>
    <div class="range-display">
        {{.Params.ISOMin}} – {{.Params.ISOMax}}
        <a href="{{removeParam .CurrentURL "iso_min" "iso_max"}}" class="range-clear">×</a>
    </div>
</div>
{{end}}
{{end}}
```

---

### Phase 3: Backend Updates (Priority: MEDIUM)

#### Task 3.1: Add Missing Facet Computations

**File**: `internal/query/facets.go`

Need to add computation for:
- Lens Make
- Season
- Orientation
- White Balance
- Flash Fired
- Colour Space
- Is Burst Representative

```go
// Add to ComputeFacets()
facets.LensMake, err = e.computeLensMakeFacet(params)
if err != nil {
    return nil, fmt.Errorf("failed to compute lens make facet: %w", err)
}

facets.Orientation, err = e.computeOrientationFacet(params)
if err != nil {
    return nil, fmt.Errorf("failed to compute orientation facet: %w", err)
}

// ... similar for others
```

Implementation examples:
```go
func (e *Engine) computeLensMakeFacet(params QueryParams) (*Facet, error) {
    paramsWithoutLensMake := params
    paramsWithoutLensMake.LensMake = nil

    where, args := e.buildWhereClause(paramsWithoutLensMake)
    whereClause := ""
    if len(where) > 0 {
        whereClause = "WHERE " + strings.Join(where, " AND ")
    }

    additionalWhere := "lens_make IS NOT NULL AND lens_make != ''"
    if whereClause != "" {
        whereClause += " AND " + additionalWhere
    } else {
        whereClause = "WHERE " + additionalWhere
    }

    query := fmt.Sprintf(`
        SELECT lens_make, COUNT(*) as count
        FROM photos p
        %s
        GROUP BY lens_make
        ORDER BY count DESC
        LIMIT 20
    `, whereClause)

    rows, err := e.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    values := []FacetValue{}
    for rows.Next() {
        var make string
        var count int
        if err := rows.Scan(&make, &count); err != nil {
            return nil, err
        }

        selected := false
        for _, m := range params.LensMake {
            if make == m {
                selected = true
                break
            }
        }

        values = append(values, FacetValue{
            Value:    make,
            Label:    make,
            Count:    count,
            Selected: selected,
        })
    }

    return &Facet{
        Name:   "lens_make",
        Label:  "Lens Make",
        Values: values,
    }, nil
}

func (e *Engine) computeOrientationFacet(params QueryParams) (*Facet, error) {
    paramsWithoutOrientation := params
    paramsWithoutOrientation.IsLandscape = nil
    paramsWithoutOrientation.IsPortrait = nil

    where, args := e.buildWhereClause(paramsWithoutOrientation)
    whereClause := ""
    if len(where) > 0 {
        whereClause = "WHERE " + strings.Join(where, " AND ")
    }

    query := fmt.Sprintf(`
        SELECT
            CASE
                WHEN width > height THEN 'landscape'
                WHEN height > width THEN 'portrait'
                ELSE 'square'
            END as orientation,
            COUNT(*) as count
        FROM photos p
        %s
        GROUP BY orientation
        ORDER BY count DESC
    `, whereClause)

    rows, err := e.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    values := []FacetValue{}
    for rows.Next() {
        var orientation string
        var count int
        if err := rows.Scan(&orientation, &count); err != nil {
            return nil, err
        }

        selected := false
        if orientation == "landscape" && params.IsLandscape != nil && *params.IsLandscape {
            selected = true
        } else if orientation == "portrait" && params.IsPortrait != nil && *params.IsPortrait {
            selected = true
        }

        values = append(values, FacetValue{
            Value:    orientation,
            Label:    strings.Title(orientation),
            Count:    count,
            Selected: selected,
        })
    }

    return &Facet{
        Name:   "orientation",
        Label:  "Orientation",
        Values: values,
    }, nil
}
```

#### Task 3.2: Update FacetCollection Type

**File**: `internal/query/types.go`

```go
type FacetCollection struct {
    Camera            *Facet
    Lens              *Facet
    LensMake          *Facet  // NEW
    Year              *Facet
    Month             *Facet
    TimeOfDay         *Facet
    Season            *Facet  // NEW
    FocalCategory     *Facet
    ShootingCondition *Facet
    InBurst           *Facet
    ColourName        *Facet
    Orientation       *Facet  // NEW
    WhiteBalance      *Facet  // NEW
    FlashFired        *Facet  // NEW (boolean)
    ColourSpace       *Facet  // NEW
    ISO               *Facet  // Future: range
    Aperture          *Facet  // Future: range
}
```

#### Task 3.3: Add Facet URL Builders

**File**: `internal/query/facet_url_builder.go`

```go
func (b *FacetURLBuilder) BuildURLsForFacets(facets *FacetCollection, baseParams QueryParams) {
    if facets.ColourName != nil {
        b.buildColourURLs(facets.ColourName, baseParams)
    }
    if facets.Year != nil {
        b.buildYearURLs(facets.Year, baseParams)
    }
    if facets.Camera != nil {
        b.buildCameraURLs(facets.Camera, baseParams)
    }
    if facets.Lens != nil {
        b.buildLensURLs(facets.Lens, baseParams)
    }
    if facets.LensMake != nil {
        b.buildLensMakeURLs(facets.LensMake, baseParams)  // NEW
    }
    if facets.TimeOfDay != nil {
        b.buildTimeOfDayURLs(facets.TimeOfDay, baseParams)
    }
    if facets.Season != nil {
        b.buildSeasonURLs(facets.Season, baseParams)
    }
    if facets.FocalCategory != nil {
        b.buildFocalCategoryURLs(facets.FocalCategory, baseParams)
    }
    if facets.ShootingCondition != nil {
        b.buildShootingConditionURLs(facets.ShootingCondition, baseParams)
    }
    if facets.InBurst != nil {
        b.buildBurstURLs(facets.InBurst, baseParams)
    }
    if facets.Orientation != nil {
        b.buildOrientationURLs(facets.Orientation, baseParams)  // NEW
    }
    if facets.WhiteBalance != nil {
        b.buildWhiteBalanceURLs(facets.WhiteBalance, baseParams)  // NEW
    }
    if facets.ColourSpace != nil {
        b.buildColourSpaceURLs(facets.ColourSpace, baseParams)  // NEW
    }
}

func (b *FacetURLBuilder) buildLensMakeURLs(facet *Facet, baseParams QueryParams) {
    for i := range facet.Values {
        p := baseParams
        if facet.Values[i].Selected {
            p.LensMake = removeFromSlice(p.LensMake, facet.Values[i].Value)
        } else {
            p.LensMake = append(p.LensMake, facet.Values[i].Value)
        }
        facet.Values[i].URL = b.mapper.BuildFullURL(p)
    }
}

func (b *FacetURLBuilder) buildOrientationURLs(facet *Facet, baseParams QueryParams) {
    for i := range facet.Values {
        p := baseParams

        // Reset orientation filters
        p.IsLandscape = nil
        p.IsPortrait = nil

        if !facet.Values[i].Selected {
            // Add selected orientation
            if facet.Values[i].Value == "landscape" {
                landscape := true
                p.IsLandscape = &landscape
            } else if facet.Values[i].Value == "portrait" {
                portrait := true
                p.IsPortrait = &portrait
            }
        }

        facet.Values[i].URL = b.mapper.BuildFullURL(p)
    }
}
```

#### Task 3.4: Extend URL Mapper for New Parameters

**File**: `internal/query/url_mapper.go`

Add parsing for new query parameters:

```go
func (m *URLMapper) parseQueryString(values url.Values, params *QueryParams) {
    // ... existing code ...

    // Lens make
    if lensMake := values["lens_make"]; len(lensMake) > 0 {
        params.LensMake = append(params.LensMake, lensMake...)
    }

    // Orientation (convert to boolean filters)
    if orientation := values["orientation"]; len(orientation) > 0 {
        for _, o := range orientation {
            if o == "landscape" {
                landscape := true
                params.IsLandscape = &landscape
            } else if o == "portrait" {
                portrait := true
                params.IsPortrait = &portrait
            }
        }
    }

    // White balance
    if wb := values["white_balance"]; len(wb) > 0 {
        params.WhiteBalance = append(params.WhiteBalance, wb...)
    }

    // Flash fired
    if flash := values.Get("flash_fired"); flash != "" {
        if flash == "true" || flash == "1" {
            flashFired := true
            params.FlashFired = &flashFired
        } else if flash == "false" || flash == "0" {
            flashFired := false
            params.FlashFired = &flashFired
        }
    }

    // Colour space
    if cs := values["color_space"]; len(cs) > 0 {
        params.ColourSpace = append(params.ColourSpace, cs...)
    }
}
```

---

### Phase 4: Mobile Enhancements (Priority: MEDIUM)

#### Task 4.1: Filter Tray JavaScript

**File**: Add `<script>` section to `grid.html` or external JS file

```javascript
function toggleFilterTray() {
    const tray = document.getElementById('filterTray');
    tray.hidden = !tray.hidden;
    if (!tray.hidden) {
        document.body.style.overflow = 'hidden';
        // Focus trap
        tray.querySelector('.tray-close').focus();
    } else {
        document.body.style.overflow = '';
    }
}

function closeFilterTray() {
    const tray = document.getElementById('filterTray');
    tray.hidden = true;
    document.body.style.overflow = '';
}

// Close on escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        closeFilterTray();
    }
});

// Close on backdrop click
document.getElementById('filterTray')?.addEventListener('click', (e) => {
    if (e.target.id === 'filterTray') {
        closeFilterTray();
    }
});
```

#### Task 4.2: Handle Sort Change

```javascript
function handleSortChange(select) {
    const [sortBy, order] = select.value.split(':');
    const url = new URL(window.location);
    url.searchParams.set('sort', sortBy);
    url.searchParams.set('order', order);
    window.location = url.toString();
}
```

---

### Phase 5: Testing & Polish (Priority: HIGH)

#### Task 5.1: Test Checklist

- [ ] Desktop layout with right rail
- [ ] Mobile layout with filter tray
- [ ] All facet types render correctly
- [ ] Active filter chips work
- [ ] Breadcrumbs (if kept)
- [ ] Pagination
- [ ] Sort dropdown
- [ ] Deep links (URL sharing)
- [ ] Progressive disclosure (Month after Year, Day after Month)
- [ ] Multi-select facets (Time of Day, Season, Lens)
- [ ] Search within facets (Camera, Lens)
- [ ] Keyboard navigation
- [ ] Screen reader compatibility
- [ ] Loading states
- [ ] Empty states
- [ ] Error states

#### Task 5.2: Performance Testing

- [ ] Facet computation time < 200ms
- [ ] Page render time < 1s
- [ ] Smooth scrolling with large result sets
- [ ] Mobile tray animation smooth

#### Task 5.3: Cross-Browser Testing

- [ ] Chrome/Edge (desktop + mobile)
- [ ] Firefox (desktop + mobile)
- [ ] Safari (desktop + mobile)

---

## Implementation Order (Recommended)

### Week 1: Foundation
1. ✅ Review spec and mock
2. Create new layout structure (Task 1.1)
3. Add CSS for new layout (Task 1.3)
4. Create facet sections template structure (Task 1.2)
5. Verify basic layout works

### Week 2: Core Facets
1. Implement Year facet (Task 2.1)
2. Implement Month/Day progressive disclosure (Task 2.2)
3. Implement Time of Day chips (Task 2.3)
4. Implement Camera list (Task 2.5)
5. Implement Colour swatches (Task 2.7)
6. Test facet toggling and URL updates

### Week 3: Additional Facets
1. Add Season facet (Task 2.4)
2. Add Lens facet (Task 2.6)
3. Add Orientation checkboxes (Task 2.8)
4. Add boolean toggles (Task 2.9)
5. Add backend facet computations (Task 3.1-3.3)
6. Test all facets work with backend

### Week 4: Mobile & Polish
1. Implement mobile filter tray (Task 4.1)
2. Add JavaScript interactions (Task 4.2)
3. Polish styling and animations
4. Accessibility audit
5. Performance testing (Task 5.2)
6. Cross-browser testing (Task 5.3)

---

## Out of Scope (Explicitly)

As per spec v2.1:
- ❌ Map facet with pan/zoom
- ❌ GPS bounding box filters
- ❌ "Use current view" map feature
- ❌ Saved places / geo presets
- ❌ Any `has_gps`, `lat_*`, `lon_*` parameters

These can be added in a future release after core facets are stable.

---

## Success Criteria

✅ **Functional**:
- All facet groups render correctly
- Facet counts update based on active filters (self-exclusion works)
- URLs reflect all active filters
- Deep links work (shareable URLs)
- Mobile tray functions smoothly

✅ **Performance**:
- Facet computation < 200ms for 10k photos
- Page load < 1s
- Smooth scrolling and interactions

✅ **Accessibility**:
- Keyboard navigable
- Screen reader friendly
- WCAG AA contrast
- Focus indicators visible

✅ **UX**:
- Active filters clearly visible
- Easy to remove individual filters
- No dead ends (zero-count facets handled gracefully)
- Progressive disclosure for hierarchical facets

---

## References

- **Spec**: `docs/faceted_navigation.spec` v2.1
- **Mock**: `docs/olsen_faceted_ui_mock.png`
- **Current Templates**: `internal/explorer/templates/grid.html`
- **Query Engine**: `internal/query/engine.go`, `facets.go`
- **URL Mapper**: `internal/query/url_mapper.go`
- **Facet URL Builder**: `internal/query/facet_url_builder.go`

---

**END OF PLAN**
