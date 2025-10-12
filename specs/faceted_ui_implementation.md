# Faceted UI Implementation Specification

**Status:** Draft
**Date:** 2025-10-08
**Reference:** `specs/olsen_faceted_ui_mock.png`
**Priority:** High

---

## Overview

This specification describes the implementation of the comprehensive faceted navigation UI shown in the mock design. The UI provides a rich, multi-dimensional filtering interface for photo exploration with a right-side filter panel and active filter chips.

**Key Principles:**
- State machine model: All filters are independent
- URL-based state (shareable, bookmarkable)
- Zero-result prevention (disabled facets)
- Real-time facet counts

---

## Current State vs. Target State

### Already Implemented ✅
- State machine query engine (independent filters)
- Active filter chips with removal
- Basic facet computation (Year, Month, Camera, Lens, Color)
- Grid layout with thumbnails
- Search bar and sort dropdown

### Needs Implementation ⚠️
- Right sidebar with organized facet sections
- Multiple facet UI patterns (lists, buttons, checkboxes, radio, toggle)
- Collapsible sections
- Additional facets (Orientation, Flash, White Balance, Bursts)
- Visual polish and styling

---

## UI Components Breakdown

### 1. Top Navigation Bar

**Location:** Sticky header at top of page

**Components:**
```
[Olsen Logo] [Search: "Search photos, cameras, lenses..."] [Sort: Date ∨]
```

**Specifications:**
- **Search field:**
  - Full-width with max-width constraint
  - Placeholder: "Search photos, cameras, lenses..."
  - Dark theme: `background: #1a1a1a`, `border: 1px solid #333`
  - Triggers query on Enter or debounced input

- **Sort dropdown:**
  - Options: Date ∨ (newest first, oldest first, random, etc.)
  - Right-aligned
  - Dark theme matching search field

**Status:** Mostly complete, needs styling verification

---

### 2. Active Filter Chips

**Location:** Below top bar, above main content

**Layout:**
```
[Colour: Red ×] [Year: 2025 ×] [Orientation: Portrait ×] [Clear all]
```

**Specifications:**
- Horizontal row with wrapping
- Each chip shows: `[FilterType: Value ×]`
- Click × to remove that specific filter
- "Clear all" button removes all filters
- Blue theme: `border: 1px solid #4a9eff`, `color: #4a9eff`
- Hover: `background: #4a9eff`, `color: #0a0a0a`

**Status:** ✅ Complete (implemented in state machine migration)

---

### 3. Right Sidebar - Filter Panel

**Location:** Right side of viewport, fixed position

**Overall Structure:**
```
┌─────────────────────────┐
│ Filters                 │ ← Header
├─────────────────────────┤
│ ▼ Time                  │ ← Section (collapsible)
│   Year                  │
│     2025         1240   │
│     2024         3812   │
│     Unknown        38   │
│   Month...              │
│   Time of day           │
│   [Dawn][Golden AM*][Midday]
├─────────────────────────┤
│ ▼ Equipment             │
│   Camera                │
│     Canon EOS R5   420  │
│     ...                 │
├─────────────────────────┤
│ ▼ Colour                │
│   ○ Red                 │
│   ○ Orange              │
│   ...                   │
├─────────────────────────┤
│ ... (more sections)     │
└─────────────────────────┘
```

**Specifications:**
- Width: 320px (fixed)
- Background: `#0a0a0a`
- Scrollable if content exceeds viewport
- Sections are collapsible (accordion pattern)
- Default: All sections expanded (can be changed later)

---

## Facet Types & UI Patterns

### Pattern A: List Facet (Clickable Items with Counts)

**Used for:** Year, Camera, Lens

**Layout:**
```
Year
  2025                    1240
  2024                    3812
  Unknown                   38
```

**Specifications:**
- Each item is a clickable link
- Label on left, count right-aligned
- Hover: `background: #1a1a1a`
- Active/selected: Bold text, blue accent
- Disabled (count=0): Gray text, not clickable
- Maximum items shown: 10 (then "Show more..." link)

**HTML Structure:**
```html
<div class="facet-section">
  <h4 class="facet-header">Year</h4>
  <ul class="facet-list">
    <li class="facet-item">
      <a href="/photos?year=2025">
        <span class="facet-label">2025</span>
        <span class="facet-count">1240</span>
      </a>
    </li>
    <li class="facet-item disabled">
      <span class="facet-label">2023</span>
      <span class="facet-count">0</span>
    </li>
  </ul>
</div>
```

---

### Pattern B: Button Group Facet

**Used for:** Time of day

**Layout:**
```
Time of day
[Dawn] [Golden AM*] [Midday]
```

**Specifications:**
- Horizontal button group (can wrap)
- Pill-shaped buttons with rounded corners
- Selected state: Filled background, bold text
- Unselected: Border only, lighter text
- Disabled: Grayed out, not clickable
- Mutually exclusive (clicking one deselects others)

**HTML Structure:**
```html
<div class="facet-section">
  <h4 class="facet-header">Time of day</h4>
  <div class="button-group">
    <a href="/photos?time_of_day=dawn" class="btn-pill">Dawn</a>
    <a href="/photos?time_of_day=golden_am" class="btn-pill active">Golden AM</a>
    <a href="/photos?time_of_day=midday" class="btn-pill">Midday</a>
  </div>
</div>
```

**CSS:**
```css
.btn-pill {
  padding: 0.4rem 1rem;
  border: 1px solid #555;
  border-radius: 16px;
  background: transparent;
  color: #888;
}
.btn-pill.active {
  background: #4a9eff;
  color: #0a0a0a;
  border-color: #4a9eff;
  font-weight: 600;
}
```

---

### Pattern C: Radio Button Facet

**Used for:** Colour

**Layout:**
```
Colour
  ○ Red
  ○ Orange
  ○ Yellow
  ○ Green
  ○ Blue
  ○ Purple
  ○ B&W
```

**Specifications:**
- Standard radio button inputs
- Each option is a label + radio
- Only one can be selected at a time
- Clicking selected item could deselect (clear filter)
- For colors: Show color swatch next to label

**HTML Structure:**
```html
<div class="facet-section">
  <h4 class="facet-header">Colour</h4>
  <div class="radio-group">
    <label class="radio-item">
      <input type="radio" name="colour" value="red">
      <span class="color-swatch" style="background: #ff0000"></span>
      <span class="radio-label">Red</span>
    </label>
    <!-- ... more options -->
  </div>
</div>
```

---

### Pattern D: Checkbox Facet

**Used for:** Orientation, Capture conditions, Bursts

**Layout:**
```
Composition & Orientation
  □ Landscape
  ☑ Portrait
  □ Square

Bursts
  ☑ In burst
  □ Is representative
```

**Specifications:**
- Standard checkbox inputs
- Multiple selections allowed
- Each checkbox represents adding/removing a filter
- Checkboxes can be combined (AND logic within section)

**HTML Structure:**
```html
<div class="facet-section">
  <h4 class="facet-header">Composition & Orientation</h4>
  <div class="checkbox-group">
    <label class="checkbox-item">
      <input type="checkbox" name="orientation" value="landscape">
      <span class="checkbox-label">Landscape</span>
    </label>
    <label class="checkbox-item">
      <input type="checkbox" name="orientation" value="portrait" checked>
      <span class="checkbox-label">Portrait</span>
    </label>
    <!-- ... -->
  </div>
</div>
```

---

### Pattern E: Toggle Switch

**Used for:** Flash fired

**Layout:**
```
Flash fired     [○──]
```

**Specifications:**
- iOS-style toggle switch
- On/Off state (boolean filter)
- Sliding animation on toggle
- Label on left, switch on right

**HTML Structure:**
```html
<div class="facet-section">
  <div class="toggle-row">
    <h4 class="facet-header">Flash fired</h4>
    <label class="toggle-switch">
      <input type="checkbox" name="flash_fired">
      <span class="slider"></span>
    </label>
  </div>
</div>
```

**CSS:**
```css
.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
}
.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}
.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #333;
  border-radius: 24px;
  transition: 0.3s;
}
.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  border-radius: 50%;
  transition: 0.3s;
}
input:checked + .slider {
  background-color: #4a9eff;
}
input:checked + .slider:before {
  transform: translateX(20px);
}
```

---

## Section Organization

### Time Section

**Components:**
- Year (List facet)
- Month (Collapsible list or dropdown)
- Day (Hidden by default, shown when Month selected?)
- Time of day (Button group: Dawn, Morning, Golden AM, Midday, Afternoon, Golden PM, Dusk, Night)
- Season (List or button group: Spring, Summer, Autumn, Winter)

**Collapsibility:** Yes
**Default State:** Expanded

---

### Equipment Section

**Components:**
- Camera (List facet with hierarchy):
  - Camera Make (collapsible)
    - Camera Model (nested, shown when make selected)
- Lens (List facet):
  - Lens Make (collapsible)
    - Lens Model (nested)

**Example:**
```
Equipment
  Camera
    Canon                        1420
      ▶ Canon EOS R5              420
      ▶ Canon EOS R6              320
    Sony                          820
      ▶ Sony A7 IV                512
  Lens
    Canon                         680
      ▶ EF 24-70mm f/2.8          266
    Fujifilm                      432
      ▶ XF 35mm F1.4              144
```

**Collapsibility:** Yes
**Default State:** Expanded

---

### Colour Section

**Components:**
- Color selection (Radio buttons):
  - Red (with #ff0000 swatch)
  - Orange (with #ff8800 swatch)
  - Yellow (with #ffff00 swatch)
  - Green (with #00ff00 swatch)
  - Blue (with #0088ff swatch)
  - Purple (with #8800ff swatch)
  - B&W (with grayscale swatch)

**Behavior:**
- Selecting a color filters to photos with that dominant color
- Only one color can be selected at a time
- Clicking selected color deselects it

**Collapsibility:** Yes
**Default State:** Expanded

---

### Composition & Orientation Section

**Components:**
- Orientation (Checkboxes):
  - □ Landscape (width > height)
  - □ Portrait (height > width)
  - □ Square (width ≈ height)

**Behavior:**
- Multiple orientations can be selected (OR logic)
- Combines with other filters using AND logic

**Collapsibility:** Yes
**Default State:** Expanded

---

### Capture Conditions Section

**Components:**
- White Balance (Radio or checkbox group):
  - □ Auto
  - □ Daylight
  - □ Cloudy
  - □ Shade
  - □ Tungsten
  - □ Fluorescent
  - □ Flash
  - □ Custom

**Behavior:**
- Checkboxes allow OR logic (any matching white balance)

**Collapsibility:** Yes
**Default State:** Expanded

---

### Flash Fired Section

**Components:**
- Flash fired (Toggle switch)

**Behavior:**
- On: Show only photos where flash fired = true
- Off: No filter on flash status

**Collapsibility:** No (single control, always visible)
**Default State:** Off

---

### Bursts Section

**Components:**
- Burst status (Checkboxes):
  - □ In burst
  - □ Is representative

**Behavior:**
- "In burst": Photos that are part of a burst sequence
- "Is representative": Photos marked as representative of their burst
- Can select both (show burst photos that are representative)

**Collapsibility:** Yes
**Default State:** Expanded

---

## Technical Implementation

### Backend Changes

#### 1. Query Engine Extensions

**File:** `internal/query/facets.go`

**Tasks:**
- ✅ Verify Year, Month, Day facets computed correctly
- ✅ Verify Camera, Lens facets computed
- ✅ Verify Color facets computed
- ✅ Verify TimeOfDay facets computed
- ⚠️ Add Orientation facet computation
- ⚠️ Add WhiteBalance facet computation
- ⚠️ Add FlashFired facet computation
- ⚠️ Add Burst facets (InBurst, IsRepresentative)

**New Facet Functions:**
```go
func (e *Engine) computeOrientationFacet(params QueryParams) (*Facet, error) {
    // Query for landscape/portrait/square counts
    // Orientation determined by width/height ratio:
    //   Landscape: width/height > 1.2
    //   Portrait: height/width > 1.2
    //   Square: abs(width - height) / max(width, height) < 0.2
}

func (e *Engine) computeWhiteBalanceFacet(params QueryParams) (*Facet, error) {
    // Query distinct white_balance values with counts
}

func (e *Engine) computeFlashFiredFacet(params QueryParams) (*Facet, error) {
    // Return count of photos with flash_fired = true vs false
}

func (e *Engine) computeBurstFacets(params QueryParams) (*Facet, error) {
    // Query for photos in burst_groups
}
```

**Update FacetCollection:**
```go
type FacetCollection struct {
    Year         *Facet
    Month        *Facet
    Day          *Facet
    TimeOfDay    *Facet
    Season       *Facet
    Camera       *Facet
    Lens         *Facet
    ColourName   *Facet
    Orientation  *Facet      // NEW
    WhiteBalance *Facet      // NEW
    FlashFired   *Facet      // NEW
    InBurst      *Facet      // NEW
    IsRepresentative *Facet  // NEW
}
```

---

#### 2. URL Parameter Extensions

**File:** `internal/query/types.go`

**Update QueryParams:**
```go
type QueryParams struct {
    // Existing fields...
    Year  *int
    Month *int
    Day   *int
    // ... etc

    // NEW fields
    Orientation  *string  // "landscape", "portrait", "square"
    WhiteBalance *string  // "auto", "daylight", etc.
    FlashFired   *bool    // true = show only photos with flash
    InBurst      *bool    // true = show only burst photos
    IsRepresentative *bool // true = show only representative burst photos
}
```

**File:** `internal/query/url_mapper.go`

**Update ParseQueryString:**
```go
// Add parsing for new parameters
if orientation := r.URL.Query().Get("orientation"); orientation != "" {
    params.Orientation = &orientation
}
// ... etc for other new fields
```

---

#### 3. Active Filter Extensions

**File:** `internal/explorer/server.go`

**Update buildActiveFilters:**
```go
func (s *Server) buildActiveFilters(params query.QueryParams) []ActiveFilter {
    filters := []ActiveFilter{}

    // Existing filters (Year, Month, Camera, Lens, Colour)...

    // NEW filters
    if params.Orientation != nil {
        p := params
        p.Orientation = nil
        filters = append(filters, ActiveFilter{
            Type:      "orientation",
            Label:     capitalizeOrientation(*params.Orientation),
            RemoveURL: s.urlMapper.BuildFullURL(p),
        })
    }

    if params.WhiteBalance != nil {
        p := params
        p.WhiteBalance = nil
        filters = append(filters, ActiveFilter{
            Type:      "white_balance",
            Label:     capitalizeWB(*params.WhiteBalance),
            RemoveURL: s.urlMapper.BuildFullURL(p),
        })
    }

    if params.FlashFired != nil && *params.FlashFired {
        p := params
        p.FlashFired = nil
        filters = append(filters, ActiveFilter{
            Type:      "flash",
            Label:     "Flash fired",
            RemoveURL: s.urlMapper.BuildFullURL(p),
        })
    }

    // ... etc for burst filters

    return filters
}
```

---

### Frontend Changes

#### 1. Template Structure

**File:** `internal/explorer/templates/grid.html`

**New Layout Structure:**
```html
{{define "grid"}}
<div class="container">
    <!-- Top Bar -->
    <div class="top-bar">
        <input type="text" class="search-field" placeholder="Search photos, cameras, lenses...">
        <span class="result-count">{{.Total}} photos</span>
        <select class="sort-dropdown">
            <option>Sort: Date ∨</option>
        </select>
    </div>

    <!-- Active Filter Chips -->
    {{if .ActiveFilters}}
    <div class="chip-row">
        {{range .ActiveFilters}}
        <a href="{{.RemoveURL}}" class="filter-chip">
            {{.Label}} <span class="filter-chip-remove">×</span>
        </a>
        {{end}}
        <a href="/photos" class="clear-all-btn">Clear all</a>
    </div>
    {{end}}

    <!-- Main Layout: Grid + Sidebar -->
    <div class="main-layout">
        <!-- Photo Grid (Left/Center) -->
        <div class="photo-grid">
            {{range .Photos}}
            <div class="photo-card">
                <img src="/thumbnail/{{.ID}}" alt="{{.FilePath}}">
                <div class="photo-meta">
                    {{.FileName}} - {{.FocalLength}}mm - f/{{.Aperture}} - ISO {{.ISO}}
                </div>
            </div>
            {{end}}
        </div>

        <!-- Filter Sidebar (Right) -->
        <aside class="filter-sidebar">
            <h3 class="sidebar-header">Filters</h3>

            <!-- Time Section -->
            {{template "facet-section" dict "Title" "Time" "ID" "time"}}
                {{template "list-facet" dict "Header" "Year" "Facet" .Facets.Year "ParamName" "year"}}
                {{template "collapsible" dict "Header" "Month..." "Content" "..."}}
                {{template "button-group-facet" dict "Header" "Time of day" "Facet" .Facets.TimeOfDay "ParamName" "time_of_day"}}
            {{end}}

            <!-- Equipment Section -->
            {{template "facet-section" dict "Title" "Equipment" "ID" "equipment"}}
                {{template "list-facet" dict "Header" "Camera" "Facet" .Facets.Camera "ParamName" "camera_model"}}
                {{template "list-facet" dict "Header" "Lens" "Facet" .Facets.Lens "ParamName" "lens_model"}}
            {{end}}

            <!-- Colour Section -->
            {{template "facet-section" dict "Title" "Colour" "ID" "colour"}}
                {{template "radio-facet" dict "Facet" .Facets.ColourName "ParamName" "colour"}}
            {{end}}

            <!-- Orientation Section -->
            {{template "facet-section" dict "Title" "Composition & Orientation" "ID" "orientation"}}
                {{template "checkbox-facet" dict "Facet" .Facets.Orientation "ParamName" "orientation"}}
            {{end}}

            <!-- Capture Conditions Section -->
            {{template "facet-section" dict "Title" "Capture conditions" "ID" "capture"}}
                {{template "checkbox-facet" dict "Facet" .Facets.WhiteBalance "ParamName" "white_balance"}}
            {{end}}

            <!-- Flash Section -->
            <div class="facet-section">
                {{template "toggle-facet" dict "Header" "Flash fired" "ParamName" "flash_fired" "Checked" .Params.FlashFired}}
            </div>

            <!-- Bursts Section -->
            {{template "facet-section" dict "Title" "Bursts" "ID" "bursts"}}
                {{template "checkbox-facet" dict "Facet" .Facets.InBurst "ParamName" "in_burst"}}
            {{end}}
        </aside>
    </div>
</div>

<style>
    /* Layout */
    .main-layout {
        display: flex;
        gap: 2rem;
    }
    .photo-grid {
        flex: 1;
        min-width: 0; /* Prevent flex overflow */
    }
    .filter-sidebar {
        width: 320px;
        flex-shrink: 0;
        background: #0a0a0a;
        padding: 1rem;
        border-left: 1px solid #333;
        position: sticky;
        top: 80px; /* Below top bar */
        height: fit-content;
        max-height: calc(100vh - 100px);
        overflow-y: auto;
    }

    /* Sidebar sections */
    .sidebar-header {
        font-size: 1.2rem;
        margin-bottom: 1rem;
        color: #fff;
    }
    .facet-section {
        margin-bottom: 1.5rem;
        padding-bottom: 1.5rem;
        border-bottom: 1px solid #222;
    }
    .facet-section:last-child {
        border-bottom: none;
    }
    .facet-header {
        font-size: 0.875rem;
        font-weight: 600;
        color: #aaa;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        margin-bottom: 0.75rem;
    }

    /* Component styles defined below... */
</style>
{{end}}
```

---

#### 2. Reusable Facet Templates

**File:** `internal/explorer/templates/facets.html` (NEW)

**List Facet Template:**
```html
{{define "list-facet"}}
<div class="facet-list-container">
    <h4 class="facet-header">{{.Header}}</h4>
    <ul class="facet-list">
        {{range .Facet.Values}}
        <li class="facet-item {{if eq .Count 0}}disabled{{end}} {{if .Selected}}active{{end}}">
            {{if gt .Count 0}}
            <a href="{{.URL}}" class="facet-link">
                <span class="facet-label">{{.Label}}</span>
                <span class="facet-count">{{.Count}}</span>
            </a>
            {{else}}
            <span class="facet-link disabled-link">
                <span class="facet-label">{{.Label}}</span>
                <span class="facet-count">0</span>
            </span>
            {{end}}
        </li>
        {{end}}
    </ul>
</div>

<style>
    .facet-list {
        list-style: none;
        padding: 0;
        margin: 0;
    }
    .facet-item {
        margin-bottom: 0.25rem;
    }
    .facet-link {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.4rem 0.5rem;
        color: #ccc;
        text-decoration: none;
        border-radius: 4px;
        transition: background 0.15s;
    }
    .facet-link:hover:not(.disabled-link) {
        background: #1a1a1a;
        color: #fff;
    }
    .facet-item.active .facet-link {
        background: #1a1a1a;
        color: #4a9eff;
        font-weight: 600;
    }
    .facet-item.disabled .facet-link {
        color: #555;
        cursor: not-allowed;
    }
    .facet-count {
        font-size: 0.875rem;
        color: #888;
        font-variant-numeric: tabular-nums;
    }
    .facet-item.active .facet-count {
        color: #4a9eff;
    }
</style>
{{end}}
```

**Button Group Facet Template:**
```html
{{define "button-group-facet"}}
<div class="button-group-container">
    <h4 class="facet-header">{{.Header}}</h4>
    <div class="button-group">
        {{range .Facet.Values}}
        <a href="{{.URL}}"
           class="btn-pill {{if .Selected}}active{{end}} {{if eq .Count 0}}disabled{{end}}"
           {{if eq .Count 0}}aria-disabled="true"{{end}}>
            {{.Label}}
        </a>
        {{end}}
    </div>
</div>

<style>
    .button-group {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
    .btn-pill {
        padding: 0.4rem 1rem;
        border: 1px solid #555;
        border-radius: 16px;
        background: transparent;
        color: #888;
        text-decoration: none;
        font-size: 0.875rem;
        transition: all 0.2s;
        white-space: nowrap;
    }
    .btn-pill:hover:not(.disabled) {
        background: #1a1a1a;
        color: #ccc;
        border-color: #777;
    }
    .btn-pill.active {
        background: #4a9eff;
        color: #0a0a0a;
        border-color: #4a9eff;
        font-weight: 600;
    }
    .btn-pill.disabled {
        opacity: 0.4;
        cursor: not-allowed;
        pointer-events: none;
    }
</style>
{{end}}
```

**Radio Facet Template:**
```html
{{define "radio-facet"}}
<div class="radio-group">
    {{range .Facet.Values}}
    <label class="radio-item {{if eq .Count 0}}disabled{{end}}">
        <input type="radio"
               name="{{$.ParamName}}"
               value="{{.Value}}"
               {{if .Selected}}checked{{end}}
               {{if eq .Count 0}}disabled{{end}}
               onchange="window.location.href='{{.URL}}'">
        {{if .ColorHex}}
        <span class="color-swatch" style="background: {{.ColorHex}}"></span>
        {{end}}
        <span class="radio-label">{{.Label}}</span>
        <span class="radio-count">({{.Count}})</span>
    </label>
    {{end}}
</div>

<style>
    .radio-group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .radio-item {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.4rem 0.5rem;
        cursor: pointer;
        border-radius: 4px;
        transition: background 0.15s;
    }
    .radio-item:hover:not(.disabled) {
        background: #1a1a1a;
    }
    .radio-item input[type="radio"] {
        accent-color: #4a9eff;
    }
    .radio-item.disabled {
        opacity: 0.4;
        cursor: not-allowed;
    }
    .color-swatch {
        width: 16px;
        height: 16px;
        border-radius: 3px;
        border: 1px solid #555;
        flex-shrink: 0;
    }
    .radio-label {
        flex: 1;
        color: #ccc;
    }
    .radio-count {
        font-size: 0.875rem;
        color: #888;
    }
</style>
{{end}}
```

**Checkbox Facet Template:**
```html
{{define "checkbox-facet"}}
<div class="checkbox-group">
    {{range .Facet.Values}}
    <label class="checkbox-item {{if eq .Count 0}}disabled{{end}}">
        <input type="checkbox"
               name="{{$.ParamName}}"
               value="{{.Value}}"
               {{if .Selected}}checked{{end}}
               {{if eq .Count 0}}disabled{{end}}
               onchange="window.location.href='{{.URL}}'">
        <span class="checkbox-label">{{.Label}}</span>
        <span class="checkbox-count">({{.Count}})</span>
    </label>
    {{end}}
</div>

<style>
    .checkbox-group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .checkbox-item {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.4rem 0.5rem;
        cursor: pointer;
        border-radius: 4px;
        transition: background 0.15s;
    }
    .checkbox-item:hover:not(.disabled) {
        background: #1a1a1a;
    }
    .checkbox-item input[type="checkbox"] {
        accent-color: #4a9eff;
    }
    .checkbox-item.disabled {
        opacity: 0.4;
        cursor: not-allowed;
    }
    .checkbox-label {
        flex: 1;
        color: #ccc;
    }
    .checkbox-count {
        font-size: 0.875rem;
        color: #888;
    }
</style>
{{end}}
```

**Toggle Facet Template:**
```html
{{define "toggle-facet"}}
<div class="toggle-row">
    <h4 class="facet-header">{{.Header}}</h4>
    <label class="toggle-switch">
        <input type="checkbox"
               name="{{.ParamName}}"
               {{if .Checked}}checked{{end}}
               onchange="toggleFilter(this, '{{.ParamName}}')">
        <span class="slider"></span>
    </label>
</div>

<style>
    .toggle-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .toggle-switch {
        position: relative;
        width: 44px;
        height: 24px;
        display: inline-block;
    }
    .toggle-switch input {
        opacity: 0;
        width: 0;
        height: 0;
    }
    .slider {
        position: absolute;
        cursor: pointer;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: #333;
        border-radius: 24px;
        transition: 0.3s;
    }
    .slider:before {
        position: absolute;
        content: "";
        height: 18px;
        width: 18px;
        left: 3px;
        bottom: 3px;
        background-color: white;
        border-radius: 50%;
        transition: 0.3s;
    }
    input:checked + .slider {
        background-color: #4a9eff;
    }
    input:checked + .slider:before {
        transform: translateX(20px);
    }
</style>

<script>
function toggleFilter(checkbox, paramName) {
    const url = new URL(window.location);
    if (checkbox.checked) {
        url.searchParams.set(paramName, 'true');
    } else {
        url.searchParams.delete(paramName);
    }
    window.location.href = url.toString();
}
</script>
{{end}}
```

---

#### 3. Collapsible Sections

**JavaScript for Accordion:**
```javascript
// Add to grid.html or separate JS file
document.addEventListener('DOMContentLoaded', function() {
    const sectionHeaders = document.querySelectorAll('.collapsible-header');

    sectionHeaders.forEach(header => {
        header.addEventListener('click', function() {
            const section = this.parentElement;
            const content = section.querySelector('.collapsible-content');
            const icon = this.querySelector('.collapse-icon');

            section.classList.toggle('collapsed');

            if (section.classList.contains('collapsed')) {
                content.style.maxHeight = '0';
                icon.textContent = '▶';
            } else {
                content.style.maxHeight = content.scrollHeight + 'px';
                icon.textContent = '▼';
            }
        });
    });
});
```

**HTML for Collapsible Section:**
```html
<div class="facet-section collapsible">
    <div class="collapsible-header">
        <span class="collapse-icon">▼</span>
        <h3 class="section-title">Time</h3>
    </div>
    <div class="collapsible-content">
        <!-- Facets go here -->
    </div>
</div>
```

**CSS:**
```css
.collapsible-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    padding: 0.5rem 0;
    user-select: none;
}
.collapsible-header:hover {
    color: #fff;
}
.collapse-icon {
    font-size: 0.75rem;
    color: #888;
    transition: transform 0.2s;
}
.section-title {
    font-size: 1rem;
    font-weight: 600;
    color: #ddd;
    margin: 0;
}
.collapsible-content {
    overflow: hidden;
    transition: max-height 0.3s ease;
}
.facet-section.collapsed .collapsible-content {
    max-height: 0;
}
```

---

## Database Schema Extensions

### Check Existing Schema

Most facets already have database columns:
- ✅ `orientation` (computed from width/height)
- ✅ `white_balance` (from EXIF)
- ✅ `flash_fired` (from EXIF)
- ✅ Burst tables exist (`burst_groups`)

**File:** `internal/database/schema.go`

**Verify these columns exist in `photos` table:**
```sql
-- Should already exist:
orientation INTEGER,              -- 1=horizontal, 3=rotate 180, 6=rotate 90 CW, 8=rotate 90 CCW
white_balance TEXT,               -- Auto, Daylight, etc.
flash_fired BOOLEAN DEFAULT 0,   -- 1 if flash fired

-- For bursts, separate table exists:
CREATE TABLE IF NOT EXISTS burst_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    photo_count INTEGER NOT NULL,
    representative_photo_id INTEGER,
    time_start TIMESTAMP,
    time_end TIMESTAMP,
    FOREIGN KEY (representative_photo_id) REFERENCES photos(id)
);

CREATE TABLE IF NOT EXISTS burst_members (
    burst_group_id INTEGER,
    photo_id INTEGER,
    sequence_order INTEGER,
    PRIMARY KEY (burst_group_id, photo_id),
    FOREIGN KEY (burst_group_id) REFERENCES burst_groups(id),
    FOREIGN KEY (photo_id) REFERENCES photos(id)
);
```

**If orientation needs to be computed differently for facets:**

Add a computed column or view:
```sql
-- Add computed orientation type (optional)
ALTER TABLE photos ADD COLUMN orientation_type TEXT
    GENERATED ALWAYS AS (
        CASE
            WHEN width > height * 1.2 THEN 'landscape'
            WHEN height > width * 1.2 THEN 'portrait'
            ELSE 'square'
        END
    ) STORED;
```

---

## Testing Requirements

### Unit Tests

**File:** `internal/query/facets_test.go`

**New Tests:**
```go
func TestComputeOrientationFacet(t *testing.T)
func TestComputeWhiteBalanceFacet(t *testing.T)
func TestComputeFlashFiredFacet(t *testing.T)
func TestComputeBurstFacets(t *testing.T)
```

### Integration Tests

**File:** `internal/query/facet_ui_integration_test.go` (NEW)

**Tests:**
```go
func TestAllFacetTypesRendered(t *testing.T) {
    // Verify all facet sections appear in UI
}

func TestFacetInteractions(t *testing.T) {
    // Test clicking each facet type
    // Verify URL updates correctly
    // Verify filter chips appear
}

func TestFacetDisabling(t *testing.T) {
    // Select filters that result in count=0 for other facets
    // Verify those facets are disabled
}

func TestCollapsibleSections(t *testing.T) {
    // Test accordion behavior
}
```

### Manual Testing Checklist

- [ ] All facet sections visible in sidebar
- [ ] Year facet shows correct counts
- [ ] Camera/Lens facets show hierarchy
- [ ] Time of day buttons work correctly
- [ ] Color radio buttons work
- [ ] Orientation checkboxes work
- [ ] Flash toggle works
- [ ] Filter chips appear when facets selected
- [ ] Filter chips remove correctly
- [ ] "Clear all" removes all filters
- [ ] Facets with count=0 are disabled
- [ ] Collapsible sections work
- [ ] Search field works (if implemented)
- [ ] Sort dropdown works (if implemented)
- [ ] Layout responsive on mobile

---

## Implementation Phases

### Phase 1: Backend Foundation (4-6 hours)
**Priority:** P0

**Tasks:**
1. Add new fields to `QueryParams` struct
2. Update `ParseQueryString()` to parse new parameters
3. Implement new facet computation functions:
   - `computeOrientationFacet()`
   - `computeWhiteBalanceFacet()`
   - `computeFlashFiredFacet()`
   - `computeBurstFacets()`
4. Update `ComputeFacets()` to include new facets
5. Update `buildActiveFilters()` to handle new filter types
6. Write unit tests for new facet functions

**Deliverable:** Backend can compute and return all required facets

---

### Phase 2: Sidebar Structure (2-3 hours)
**Priority:** P0

**Tasks:**
1. Redesign `grid.html` layout to include right sidebar
2. Create sidebar structure with section headers
3. Implement collapsible sections (accordion)
4. Add CSS for sidebar layout and responsiveness
5. Add JavaScript for collapse/expand behavior

**Deliverable:** Sidebar structure in place, sections collapsible

---

### Phase 3: Facet UI Components (4-6 hours)
**Priority:** P0

**Tasks:**
1. Create `facets.html` with reusable templates:
   - `list-facet` template
   - `button-group-facet` template
   - `radio-facet` template
   - `checkbox-facet` template
   - `toggle-facet` template
2. Integrate templates into `grid.html`
3. Pass facet data from server to templates
4. Add CSS for each component type
5. Test each facet type individually

**Deliverable:** All facet types rendering correctly

---

### Phase 4: Visual Polish (2-3 hours)
**Priority:** P1

**Tasks:**
1. Match colors from mock design
2. Refine spacing, typography, sizing
3. Add hover states and transitions
4. Test dark theme consistency
5. Add responsive breakpoints for mobile
6. Add loading states

**Deliverable:** UI matches mock design visually

---

### Phase 5: Testing & Bug Fixes (2-4 hours)
**Priority:** P0

**Tasks:**
1. Write integration tests
2. Manual testing of all facet interactions
3. Test with real photo database
4. Test edge cases (empty facets, no results, etc.)
5. Test on different screen sizes
6. Fix any bugs found

**Deliverable:** All tests passing, no critical bugs

---

## Total Effort Estimate

- **Backend:** 4-6 hours
- **Sidebar Structure:** 2-3 hours
- **Facet Components:** 4-6 hours
- **Visual Polish:** 2-3 hours
- **Testing:** 2-4 hours

**Total:** 14-22 hours (approximately 2-3 days of focused work)

---

## Dependencies

**Required:**
- ✅ State machine query engine (complete)
- ✅ Active filter chips (complete)
- ✅ Basic facet computation (complete)

**Nice to Have:**
- Search functionality (can be stubbed)
- Sort functionality (can be stubbed)
- Pagination (should already exist)

---

## Risks & Mitigation

### Risk 1: Database Schema Changes
**Risk:** Orientation or other fields don't exist in current schema
**Mitigation:** Check schema first, add computed columns if needed
**Impact:** Low (schema changes are straightforward)

### Risk 2: Performance with Many Facets
**Risk:** Computing all facets on every page load is slow
**Mitigation:**
- Use existing 200ms timeout pattern from Datasette
- Cache facet results
- Lazy-load collapsed sections
**Impact:** Medium (affects UX)

### Risk 3: Mobile Responsiveness
**Risk:** Right sidebar doesn't work well on mobile
**Mitigation:**
- Make sidebar collapsible drawer on mobile
- Put behind hamburger menu
- Stack below grid instead of beside
**Impact:** Medium (mobile UX critical)

### Risk 4: Browser Compatibility
**Risk:** CSS Grid, Flexbox, or JS not supported in older browsers
**Mitigation:** Use progressive enhancement, test in multiple browsers
**Impact:** Low (most users on modern browsers)

---

## Future Enhancements

**Post-MVP features to consider:**

1. **Smart Facet Suggestions**
   - Show most useful facets first
   - Hide facets with all count=0
   - Dynamically reorder based on usage

2. **Facet Search**
   - Search within long facet lists (e.g., Camera models)
   - Useful when 100+ distinct values

3. **Range Facets**
   - Aperture range slider (f/1.4 - f/22)
   - ISO range slider (100 - 6400)
   - Focal length range (8mm - 600mm)
   - Date range picker

4. **Histogram Facets**
   - Visual distribution of photos by year (bar chart)
   - Photos by time of day (24-hour heatmap)

5. **Save & Share Filters**
   - Save filter combinations as "Smart Collections"
   - Share filter URLs
   - Export filter as JSON

6. **Keyboard Navigation**
   - Keyboard shortcuts for common filters
   - Arrow keys to navigate facets
   - Escape to clear all filters

7. **Facet Presets**
   - "Best of 2024"
   - "Golden hour portraits"
   - "Burst photos only"
   - User-defined presets

---

## References

- **UI Mock:** `specs/olsen_faceted_ui_mock.png`
- **State Machine Docs:** `docs/STATE_MACHINE_MIGRATION.md`
- **Datasette Lessons:** `docs/DATASETTE_LESSONS.md`
- **Current Implementation:** `internal/explorer/templates/grid.html`

---

## Approval & Sign-off

**Specification Status:** Draft (awaiting manual edits)

**Approval Checklist:**
- [ ] Technical approach reviewed
- [ ] UI design approved
- [ ] Effort estimate accepted
- [ ] Priority order confirmed
- [ ] Database changes approved
- [ ] Testing requirements sufficient

**Next Steps:**
1. User reviews and edits this specification
2. Finalize priorities and scope
3. Begin Phase 1 implementation
4. Iterative development and testing
5. Deploy to staging for review
6. Production deployment

---

*End of Specification*
