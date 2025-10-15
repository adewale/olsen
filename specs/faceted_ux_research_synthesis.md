# Olsen Faceted Browsing: Research-Backed UX Specification

**Version:** 3.0  
**Date:** 2025-10-13  
**Status:** Research-backed specification  
**Purpose:** Make olsen easier for new users through world-class faceted browsing

---

## Executive Summary

This specification synthesizes research from leading experts in information architecture and usability—Nielsen Norman Group, Peter Morville, and Daniel Tunkelang—with modern design patterns to create a faceted browsing experience that makes olsen immediately accessible to photographers.

**Core Principle:** Faceted navigation is a **state machine** where users explore their photo collection through **valid state transitions**. Users should never be able to transition from a state with results to a state with zero results.

**Key Insight:** The existing implementation already has the correct state machine foundation. This specification focuses on making that foundation **visible, understandable, and delightful** for users.

---

## Part 1: Research Foundation

### 1.1 Nielsen Norman Group Principles

#### Critical Success Factors

**Simultaneous Display** (Nielsen's #1 Principle)[NN1]
- Users must see **facet controls AND results together**
- Dynamic updates: results change immediately when filters applied
- Embodies two fundamental heuristics:
  - Rapid system feedback
  - User control and freedom

**Simple Controls for Sophisticated Searches** (Nielsen's #2 Principle)[NN1]
- Use familiar interface elements (checkboxes, dropdowns, radio buttons)
- Natural-language labels, no Boolean logic required
- Empowers ordinary users to construct complex queries

**Appropriate, Predictable, Prioritized Filters** (Nielsen)[NN2]
- Filter categories must cover the most important aspects users care about
- Customize filters for the content type (don't rely on generic filters alone)
- Users notice and complain when critical filters are missing
- Domain expertise matters: specialized sites can provide superior filtering

**Mobile: Push-Out Tray Pattern** (Nielsen)[NN1]
- Overlay facet controls on top of results (not replacing)
- Keep results visible in background while filtering
- Translucent shadow to maintain hierarchy
- Show filter panel from right edge, leaving left edge of results visible
- Fixed header shows total count even when scrolling through facets

#### Implementation Priorities for Olsen

✅ Already Have:
- Simultaneous display (filters and results on same page)
- Simple controls (checkboxes, links, buttons)
- URL-based state (shareable, bookmarkable)

🎯 Need to Improve:
- Mobile tray implementation
- Visual hierarchy (make relationship clearer)
- Persistent result count display
- Filter customization for photography domain

### 1.2 Peter Morville's Ambient Findability

#### Core Concepts

**Information Literacy is Critical** (Morville)[PM1]
- Users need to understand how to navigate faceted systems
- First-time experience must establish mental model quickly
- Progressive disclosure: reveal complexity gradually

**Faceted Navigation as Master Pattern** (Morville)[PM2]
- Deployment impacts entire information architecture
- Not just a feature—a fundamental organizing principle
- Changes how users think about content organization

**Flexibility Through Multiple Access Points** (Morville)[PM2]
- Users should enter from any facet dimension
- No forced ordering of attributes
- Accommodates diverse search strategies

#### Implementation Priorities for Olsen

✅ Already Have:
- Multiple access points (year, camera, color, etc.)
- No forced hierarchies (state machine model)
- Flexible navigation paths

🎯 Need to Improve:
- First-time user experience (onboarding)
- Progressive disclosure of advanced facets
- Clear mental model establishment

### 1.3 Daniel Tunkelang's Faceted Search Theory

#### Fundamental Principles

**Progressive Query Refinement** (Tunkelang)[DT1]
- Users build queries incrementally
- Each choice updates available options in other dimensions
- Eliminates frustrating "dead ends"

**Faceted Navigation vs Parametric Search** (Tunkelang)[DT1]
- Parametric search: set all facets at once (bad)
- Faceted navigation: progressive elaboration with guidance (good)
- Dynamic feedback guides users toward successful queries

**Guided Exploration** (Tunkelang)[DT1]
- System shows what's possible at each step
- Users discover relevant information through navigation
- Balance control with exploration

#### Implementation Priorities for Olsen

✅ Already Have:
- Progressive refinement (add filters one at a time)
- Dynamic facet counts (guidance mechanism)
- State machine prevents dead ends

🎯 Need to Improve:
- Visual feedback on state transitions
- Clear indication of "what's possible next"
- Better guidance for exploration patterns

### 1.4 Modern Design Patterns (2024-2025)

#### Current Best Practices

**Visual Thumbnails** (Modern DAM)[MD1]
- Grid view with large, high-quality previews
- Hover interactions for quick metadata
- Responsive sizing for different viewports

**AI-Powered Metadata** (Modern DAM)[MD1]
- Auto-tagging and semantic search
- Content recognition for findability
- Smart collections based on metadata rules

**Instant Search & Autocomplete** (Algolia/Elasticsearch pattern)[MD1]
- Real-time results as users type
- Typo tolerance and suggestions
- Context-aware filtering

**Active Filter Chips** (Modern pattern)[MD2]
- Show applied filters as removable chips
- Clear visual indication of active state
- Quick removal without returning to filter panel

#### Implementation Priorities for Olsen

✅ Already Have:
- Active filter chips with removal
- Grid view with thumbnails
- URL-based state for shareability

🎯 Need to Improve:
- Thumbnail size and hover interactions
- Search autocomplete functionality
- Visual polish and modern aesthetics

---

## Part 2: User-Centered Design Framework

### 2.1 User Mental Models

#### Primary User: Photographer Managing Personal Collection

**Goals:**
- Find specific photos quickly ("that sunset from Iceland")
- Rediscover forgotten photos ("what did I shoot in 2020?")
- Identify patterns in shooting ("how many wide-angle shots?")
- Create collections for sharing or editing

**Mental Model:**
- Photos organized by **what, when, where, how**
- Filtering as narrowing down, not searching for keywords
- Expectation of seeing results immediately
- Want to explore, not just retrieve

**Pain Points with Current Systems:**
- Too many hierarchical folders (forced organization)
- Keyword search requires exact memory
- Can't combine multiple dimensions easily
- Dead ends with zero results are frustrating

#### Novice User: First-Time Visitor

**Goals:**
- Understand what the system does
- Explore available photos without commitment
- Learn filtering patterns through discovery
- Not break anything or get lost

**Mental Model:**
- Expects filtering to work like e-commerce
- Familiar with basic web conventions
- May not understand photography metadata
- Wants immediate success, not learning curve

**Pain Points:**
- Technical jargon (focal length, aperture)
- Too many options presented at once
- Unclear what filters do
- No guidance on "what to try next"

### 2.2 Interaction Principles

#### Principle 1: Immediate Feedback (Nielsen Heuristic #1)

**Implementation:**
- Results update within 200ms of filter selection
- Show loading states only if necessary
- Animate transitions to indicate change
- Display result count prominently

**Example:**
```
User clicks: Color = Blue
→ Results fade slightly
→ Count updates: "1,240 photos" → "350 photos"
→ Results re-render with new photos
→ Time elapsed: 150ms
```

#### Principle 2: Visible System Status (Nielsen Heuristic #2)

**Implementation:**
- Always show applied filters as chips
- Display total count at all times
- Indicate when facets are computing
- Show progress for long operations

**Example:**
```
[Top bar] 350 photos | [Filters: Blue × | 2024 ×] | Clear all
```

#### Principle 3: User Control & Freedom (Nielsen Heuristic #3)

**Implementation:**
- Every filter can be easily removed
- "Clear all" is always visible
- No irreversible actions
- Back button works correctly (URL-based state)

**Example:**
```
Applied filter by mistake?
→ Click × on chip (immediate removal)
→ Or use browser back button (restores previous state)
→ Or click "Clear all" (returns to full collection)
```

#### Principle 4: Prevent Errors (Nielsen Heuristic #5)

**Implementation:**
- Disable facet values with count=0
- Show why options are disabled
- Prevent impossible filter combinations
- Guide users toward valid states

**Example:**
```
Year facet:
  2025 (1,240) ← Clickable
  2024 (350)   ← Currently selected
  2023 (0)     ← Disabled, tooltip: "No photos from 2023 with current filters"
```

#### Principle 5: Recognition Over Recall (Nielsen Heuristic #6)

**Implementation:**
- Show all available options (don't hide)
- Use visual cues (color swatches, icons)
- Provide examples in labels
- Make context clear through layout

**Example:**
```
Time of Day:
  [Dawn 🌅] [Golden Hour 🌄] [Midday ☀️] [Blue Hour 🌆]
  
Better than:
  Dawn | Golden Hour | Midday | Blue Hour
```

---

## Part 3: Information Architecture

### 3.1 Facet Organization Strategy

#### Primary Facets (Always Visible)

**Time** - Most universal dimension
- Year (single select)
- Month (progressive: shown when year selected)
- Time of Day (multi-select chips)

**Equipment** - Core to photographers
- Camera (single select)
- Lens (multi-select)

**Colour** - Visual and intuitive
- Colour Name (single select with swatches)

**Rationale:** These facets work for every photographer, require no technical knowledge, and provide immediate value.

#### Secondary Facets (Collapsible)

**Composition & Orientation**
- Orientation (multi-select: landscape, portrait, square)
- In Burst (boolean toggle)

**Capture Conditions** (Advanced)
- Focal Category (wide, normal, telephoto)
- Shooting Condition (bright, moderate, low light)
- White Balance (multi-select)
- Flash Fired (boolean toggle)

**Rationale:** These facets are valuable but less universally understood. Collapsible to reduce initial cognitive load.

### 3.2 Progressive Disclosure Strategy

#### First Visit: Simplified View

Show only:
- Time (Year + Time of Day)
- Equipment (Camera)
- Colour
- Result count: ~5-7 facet sections

**Goal:** Immediate success with minimal learning

#### After First Interaction: Expanded View

Reveal:
- Month (when year selected)
- Lens (with camera counts)
- Orientation options

**Goal:** Reward exploration with more capability

#### Power User: Full View

Show:
- All facets expanded
- Advanced ranges (if implemented)
- Saved searches
- Bulk operations

**Goal:** Support sophisticated workflows

### 3.3 Facet Count Display Strategy (Critical!)

#### Self-Exclusion Rule (Prevents Dead Ends)

**For facet F with filters {A, B, C, F} applied:**
- Compute F's counts with filters {A, B, C} only
- Exclude F's current value from the count query
- This shows "how many photos if I change F?"

**Example:**
```
Current state: year=2024 & color=blue (350 photos)

Year facet computation:
  SELECT year, COUNT(*) 
  FROM photos 
  WHERE color='blue'      ← Include color filter
  -- (NO year filter)      ← Exclude year filter
  GROUP BY year

Results:
  2025: 280  ← 280 blue photos from 2025
  2024: 350  ← Currently selected (shown as selected)
  2023: 120  ← 120 blue photos from 2023
  2022: 0    ← No blue photos from 2022 (disabled)
```

#### Why This Matters

**Bad approach (include self in count):**
```
Year 2024: 350  ← Currently selected
Year 2025: 0    ← Disabled (wrong! there ARE blue 2025 photos)
Year 2023: 0    ← Disabled (wrong! there ARE blue 2023 photos)
```
User gets trapped in 2024, can't explore other years with blue photos.

**Good approach (exclude self from count):**
```
Year 2024: 350  ← Currently selected
Year 2025: 280  ← Clickable! Shows user can pivot to 2025
Year 2023: 120  ← Clickable! Shows user can pivot to 2023
```
User can freely explore all years that have blue photos.

#### Implementation Status

✅ Already implemented correctly in `internal/query/facets.go`
- SQL queries already exclude current facet from WHERE clause
- Count computation follows self-exclusion rule
- State machine model ensures valid transitions

🎯 Need to improve:
- Make the guidance more visible in UI
- Add tooltips explaining counts
- Highlight "suggested next filters"

---

## Part 4: Visual Design Specification

### 4.1 Layout Structure (Desktop)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Olsen] [Search...........................] [Sort ▾] [350 photos]│ ← Sticky header
├─────────────────────────────────────────────────────────────────┤
│ [Blue ×] [2024 ×] [Canon ×]                        [Clear all]   │ ← Filter chips
├──────────────────────────────────────┬──────────────────────────┤
│                                      │ Filters                   │
│  [Photo Grid]                        │ ─────────────             │
│  ┌────┬────┬────┐                   │ ▼ Time                    │
│  │    │    │    │                    │   Year                    │
│  │    │    │    │                    │   □ 2025  (1,240)        │
│  │    │    │    │                    │   ☑ 2024  (350)          │
│  ├────┼────┼────┤                   │   ☐ 2023  (0) [disabled] │
│  │    │    │    │                    │                           │
│  │    │    │    │                    │   Month (when year set)   │
│  │    │    │    │                    │                           │
│  └────┴────┴────┘                   │   Time of Day             │
│                                      │   [Dawn][Morning*][Midday]│
│  [Load more...]                      │                           │
│                                      │ ▼ Equipment               │
│                                      │   Camera                  │
│                                      │   □ Canon     (420)       │
│                                      │   □ Nikon     (280)       │
│                                      │                           │
│                                      │ ▼ Colour                  │
│                                      │   ● Red                   │
│                                      │   ● Blue                  │
│                                      │                           │
└──────────────────────────────────────┴──────────────────────────┘
```

**Key Measurements:**
- Sidebar width: 320px (fixed)
- Grid: Flexible (fills remaining space)
- Chip bar height: 60px
- Top bar height: 64px
- Gap between sections: 1.5rem (24px)

### 4.2 Layout Structure (Mobile)

```
┌─────────────────────────────┐
│ [☰] Olsen        350 photos │ ← Sticky header
├─────────────────────────────┤
│ [Filters (3)] [Sort ▾]      │ ← Action bar
├─────────────────────────────┤
│ [Blue ×] [2024 ×] [Canon ×] │ ← Chips
├─────────────────────────────┤
│                             │
│  [Photo Grid]               │
│  ┌──────┬──────┐           │
│  │      │      │            │
│  │      │      │            │
│  ├──────┼──────┤           │
│  │      │      │            │
│  └──────┴──────┘           │
│                             │
└─────────────────────────────┘

When [Filters] clicked:
┌─────────────────────────────┐
│ ╳ Filters     [Apply] [Reset]│ ← Tray header
├─────────────────────────────┤
│ [Translucent overlay]        │ ← Photos visible behind
│                             │
│ ▼ Time                       │
│   Year                       │
│   □ 2025  (1,240)           │
│   ☑ 2024  (350)             │
│                             │
│ ▼ Equipment                  │
│   Camera                     │
│   □ Canon (420)             │
│   □ Nikon (280)             │
│                             │
│ [Apply] [Reset]              │ ← Bottom actions
└─────────────────────────────┘
```

**Key Behaviors:**
- Tray slides in from right
- Photos visible through translucent overlay
- Apply/Reset buttons fixed at bottom
- Badge on "Filters" shows active count

### 4.3 Visual Hierarchy

#### Typography

```css
/* Header */
.page-title: 24px, weight 700
.result-count: 16px, weight 400, color: #888

/* Facets */
.section-header: 14px, weight 600, uppercase, letter-spacing: 0.05em, color: #aaa
.facet-label: 14px, weight 400, color: #ccc
.facet-count: 13px, weight 400, color: #888, tabular-nums

/* Chips */
.chip-label: 14px, weight 500, color: #4a9eff
```

#### Color Palette (Dark Theme)

```css
/* Backgrounds */
--bg-primary: #0a0a0a;      /* Main background */
--bg-secondary: #1a1a1a;    /* Hover states */
--bg-tertiary: #2a2a2a;     /* Active states */

/* Borders */
--border-subtle: #222;      /* Section dividers */
--border-medium: #333;      /* Input borders */
--border-strong: #555;      /* Focused borders */

/* Text */
--text-primary: #ffffff;    /* Headings */
--text-secondary: #ccc;     /* Body text */
--text-tertiary: #888;      /* Counts, metadata */
--text-disabled: #555;      /* Disabled state */

/* Accent */
--accent-primary: #4a9eff;  /* Links, buttons, chips */
--accent-hover: #6bb0ff;    /* Hover state */
--accent-active: #2a8eff;   /* Active state */

/* Status */
--success: #10b981;         /* Success states */
--warning: #f59e0b;         /* Warning states */
--error: #ef4444;           /* Error states */
```

#### Spacing System

```css
--space-xs: 0.25rem;   /* 4px */
--space-sm: 0.5rem;    /* 8px */
--space-md: 1rem;      /* 16px */
--space-lg: 1.5rem;    /* 24px */
--space-xl: 2rem;      /* 32px */
--space-2xl: 3rem;     /* 48px */
```

### 4.4 Component Specifications

#### Filter Chip Component

```css
.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--accent-primary);
  border-radius: 16px;
  color: var(--accent-primary);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.filter-chip:hover {
  background: var(--accent-primary);
  color: var(--bg-primary);
  cursor: pointer;
}

.filter-chip-remove {
  font-size: 16px;
  line-height: 1;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.filter-chip:hover .filter-chip-remove {
  opacity: 1;
}
```

#### Facet List Item

```css
.facet-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) var(--space-sm);
  border-radius: 4px;
  transition: background 0.15s;
}

.facet-item:hover:not(.disabled) {
  background: var(--bg-secondary);
  cursor: pointer;
}

.facet-item.active {
  background: var(--bg-secondary);
  font-weight: 600;
}

.facet-item.active .facet-label {
  color: var(--accent-primary);
}

.facet-item.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.facet-count {
  font-variant-numeric: tabular-nums;
  color: var(--text-tertiary);
}
```

#### Button Group (Time of Day)

```css
.button-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.btn-pill {
  padding: var(--space-sm) var(--space-md);
  border: 1px solid var(--border-strong);
  border-radius: 16px;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 14px;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-pill:hover:not(.disabled) {
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border-color: var(--border-strong);
}

.btn-pill.active {
  background: var(--accent-primary);
  color: var(--bg-primary);
  border-color: var(--accent-primary);
  font-weight: 600;
}

.btn-pill.disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
}
```

### 4.5 Loading & Empty States

#### Loading State

```html
<div class="loading-state">
  <div class="skeleton-grid">
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
  </div>
  <div class="loading-overlay">
    <div class="spinner"></div>
    <p>Loading photos...</p>
  </div>
</div>
```

**Behavior:**
- Show skeleton cards in grid
- Fade in actual photos when loaded
- Max loading time before showing spinner: 200ms

#### Empty State (No Results)

```html
<div class="empty-state">
  <svg class="empty-icon"><!-- Illustration --></svg>
  <h3>No photos found</h3>
  <p>Try removing some filters to see more results</p>
  
  <div class="recent-filters">
    <p>Recently applied:</p>
    <a href="..." class="filter-chip">Blue ×</a>
    <a href="..." class="filter-chip">2024 ×</a>
  </div>
  
  <a href="/photos" class="btn-primary">Clear all filters</a>
</div>
```

**Behavior:**
- Show friendly message (not error)
- Display recently applied filters with remove links
- Prominent "Clear all" button
- Never auto-remove filters (user control)

---

## Part 5: Implementation Roadmap

### 5.1 Phase 1: Foundation (Week 1)

**Goal:** Ensure state machine model is solid and well-tested

**Tasks:**
1. ✅ Verify facet self-exclusion logic (already done)
2. ✅ Confirm URL state management (already done)
3. ✅ Test state transitions (already done)
4. 🎯 Add comprehensive logging for debugging
5. 🎯 Document state machine behavior for developers

**Deliverable:** Rock-solid foundation with excellent test coverage

### 5.2 Phase 2: Visual Polish (Week 2)

**Goal:** Make the state machine visible and beautiful

**Tasks:**
1. 🎯 Implement disabled state for count=0 facets
2. 🎯 Add tooltips explaining why options disabled
3. 🎯 Improve filter chip styling
4. 🎯 Add animations for state transitions
5. 🎯 Implement loading states

**Deliverable:** Visually polished interface that clearly communicates state

### 5.3 Phase 3: Mobile Experience (Week 3)

**Goal:** Perfect mobile faceted navigation

**Tasks:**
1. 🎯 Implement filter tray with overlay
2. 🎯 Add Apply/Reset buttons
3. 🎯 Test on multiple devices
4. 🎯 Optimize touch targets (44×44px minimum)
5. 🎯 Implement swipe gestures

**Deliverable:** Mobile experience as good as desktop

### 5.4 Phase 4: Progressive Disclosure (Week 4)

**Goal:** Reduce cognitive load for new users

**Tasks:**
1. 🎯 Implement collapsible sections
2. 🎯 Add "Show more" for long facet lists
3. 🎯 Progressive Month/Day revelation
4. 🎯 Smart defaults based on usage
5. 🎯 First-run tutorial/tips

**Deliverable:** Friendly first-time experience

### 5.5 Phase 5: Performance & Polish (Week 5)

**Goal:** Fast, delightful interactions

**Tasks:**
1. 🎯 Optimize facet computation (<50ms per facet)
2. 🎯 Add result count animations
3. 🎯 Implement virtualized scrolling for large result sets
4. 🎯 Add keyboard navigation
5. 🎯 Comprehensive accessibility audit

**Deliverable:** Production-ready, performant, accessible

---

## Part 6: Success Metrics

### 6.1 Quantitative Metrics

**Task Success Rate**
- Baseline: 60% (guess based on typical faceted search)
- Target: 85%+
- Measure: Can users find specific photos in <60 seconds?

**Time to First Result**
- Baseline: 15 seconds (typical first interaction)
- Target: <5 seconds
- Measure: Time from landing to first filter applied

**Filter Abandonment Rate**
- Baseline: 40% (users who start filtering but give up)
- Target: <15%
- Measure: % of sessions where filters applied but cleared without viewing photos

**Zero-Result Rate**
- Baseline: Unknown
- Target: <2% (should be nearly impossible with disabled states)
- Measure: % of filter combinations that produce zero results

### 6.2 Qualitative Metrics

**System Usability Scale (SUS)**
- Baseline: Unknown
- Target: 80+ (excellent usability)
- Measure: Post-session SUS questionnaire

**User Sentiment**
- Positive quotes about ease of use
- Reduction in support requests
- Increase in repeat usage

**Mental Model Accuracy**
- Users understand how facets work
- Can explain system to others
- Predict system behavior correctly

### 6.3 Technical Metrics

**Facet Computation Time**
- Target: <50ms per facet
- Target: <200ms for full facet set (10+ facets)
- Measure: Server-side timing logs

**Page Load Time**
- Target: <1 second for initial render
- Target: <200ms for filter interactions
- Measure: Browser performance API

**Error Rate**
- Target: <0.1% of requests result in errors
- Measure: Server logs

---

## Part 7: Testing Strategy

### 7.1 Usability Testing Protocol

**Test Structure:**
5 participants per round, 3 rounds total (15 participants)

**Participant Profile:**
- Mix of photography experience (amateur to professional)
- Mix of technical aptitude
- First-time users of olsen

**Test Scenarios:**

1. **Discovery Task:** "Explore the photo collection. Tell me what you see."
   - Goal: Understand natural exploration patterns
   - Observe: Which facets do they try first?

2. **Specific Retrieval:** "Find all landscape photos from 2024."
   - Goal: Test multi-facet combinations
   - Observe: Do they understand how to combine filters?

3. **Recovery Task:** "You've applied too many filters and have zero results. Fix it."
   - Goal: Test understanding of state and recovery
   - Observe: Do they understand disabled states and removal?

4. **Mobile Task:** "Find photos taken with a specific camera on your phone."
   - Goal: Test mobile tray usability
   - Observe: Can they use tray effectively?

**Success Criteria:**
- 85%+ complete tasks successfully
- <3 errors per participant
- Positive sentiment (SUS score 80+)

### 7.2 A/B Testing Plan

**Test 1: Disabled vs Hidden Zero-Count Facets**
- Variant A: Show disabled (count=0) facets in gray
- Variant B: Hide zero-count facets completely
- Hypothesis: Showing disabled is better (transparency)
- Metric: Filter abandonment rate

**Test 2: Progressive Disclosure vs Always Expanded**
- Variant A: Collapsed sections by default
- Variant B: All sections expanded
- Hypothesis: Progressive disclosure reduces cognitive load
- Metric: Time to first result, task success rate

**Test 3: Chip Removal Patterns**
- Variant A: × on right side of chip
- Variant B: Hover reveals × overlay
- Hypothesis: Always-visible × is clearer
- Metric: Successful removals on first try

### 7.3 Performance Testing

**Load Testing:**
- 10,000 photos: All facets compute in <200ms
- 50,000 photos: All facets compute in <500ms
- 100,000 photos: All facets compute in <1s

**Stress Testing:**
- Rapid filter application (10 filters in 5 seconds)
- Concurrent users (100 simultaneous sessions)
- Large result sets (10,000+ matching photos)

---

## Part 8: Documentation & Training

### 8.1 User Documentation

**Quick Start Guide** (250 words)
- "Your Photos, Your Way"
- 3 simple examples with screenshots
- Focused on discovery, not features

**Interactive Tutorial** (optional)
- First-run only
- 3 steps: Apply filter, see results, remove filter
- Skippable
- Uses actual user's photos

**Help Tooltips**
- Hover on section headers
- Explain facet purpose in plain language
- "Why is this disabled?" explanations

### 8.2 Developer Documentation

**State Machine Documentation**
- How facets compute counts
- Self-exclusion rule explained
- URL structure and parsing
- Adding new facets (cookbook)

**Component Library**
- Storybook for all UI components
- Usage examples
- Accessibility notes
- Mobile considerations

**Testing Guide**
- How to test new facets
- Performance benchmarks
- Usability testing protocol

---

## Part 9: Accessibility (WCAG 2.1 Level AA)

### 9.1 Keyboard Navigation

**Tab Order:**
1. Skip to main content link (hidden until focused)
2. Search field
3. Sort dropdown
4. Filter chips (with removal via Enter/Space)
5. Facet sections (collapsible with Enter/Space)
6. Facet values (selectable with Enter/Space)
7. Photos in grid (viewable with Enter)

**Keyboard Shortcuts:**
- `/`: Focus search field
- `Esc`: Clear focused element or close mobile tray
- `Arrow keys`: Navigate within facet lists
- `Enter/Space`: Toggle selection

### 9.2 Screen Reader Support

**ARIA Labels:**
```html
<nav aria-label="Photo filters">
  <div role="group" aria-labelledby="time-header">
    <h4 id="time-header">Time</h4>
    <ul role="list">
      <li role="listitem">
        <a href="..." aria-label="Filter by year 2024, 350 photos">
          <span>2024</span>
          <span aria-hidden="true">350</span>
        </a>
      </li>
    </ul>
  </div>
</nav>
```

**Live Regions:**
```html
<div aria-live="polite" aria-atomic="true">
  350 photos found
</div>
```

Updates announced when result count changes.

### 9.3 Visual Accessibility

**Color Contrast:**
- All text meets WCAG AA (4.5:1 minimum)
- Interactive elements meet WCAG AA (3:1 minimum)
- Color swatches have text labels as backup

**Focus Indicators:**
- 3px solid outline on focused elements
- Color: `var(--accent-primary)`
- Visible on dark and light backgrounds

**Text Size:**
- Base: 16px (1rem)
- Scalable with browser zoom (up to 200%)
- No fixed heights that break at large sizes

---

## Part 10: References & Citations

### Research Sources

**[NN1]** Nielsen Norman Group - "Mobile Faceted Search with a Tray"
- https://www.nngroup.com/articles/mobile-faceted-search/
- Key insight: Push-out tray with translucent overlay

**[NN2]** Nielsen Norman Group - "Filter Categories and Values"
- https://www.nngroup.com/articles/filter-categories-values/
- Key insight: Appropriate, predictable, prioritized filters

**[PM1]** Peter Morville - "Ambient Findability" (2005)
- Information literacy is critical for modern systems
- Faceted classification removes single taxonomy limitations

**[PM2]** Peter Morville - "Design Patterns: Faceted Navigation"
- A List Apart article
- Faceted navigation as master pattern affecting entire IA

**[DT1]** Daniel Tunkelang - "Faceted Search" (2009)
- Progressive query refinement eliminates dead ends
- Distinction between parametric search and faceted navigation
- Balance control with exploration

**[MD1]** Modern Design Patterns - Frontify DAM, Algolia implementations
- Visual thumbnails, AI-powered metadata
- Instant search with autocomplete
- Active filter chips

**[MD2]** Baymard Institute - Applied Filters Research
- https://baymard.com/blog/how-to-design-applied-filters
- Visual chip patterns for active filters

### Internal Documentation

- `specs/faceted_navigation.spec` - Technical specification
- `specs/facet_state_machine.spec` - State machine model
- `docs/HIERARCHICAL_FACETS.md` - State machine insights
- `specs/faceted_ui_implementation.md` - Implementation details

---

## Conclusion

This specification provides a research-backed roadmap for making olsen's faceted browsing world-class. The foundation is already solid—the state machine model is correct and well-tested. The remaining work focuses on making that foundation **visible, understandable, and delightful** for users.

**Key Principles:**
1. **Transparent State:** Users always know where they are and what's possible
2. **Progressive Discovery:** Reveal complexity gradually, not all at once
3. **Error Prevention:** Disabled states prevent frustration
4. **Immediate Feedback:** Results update within 200ms
5. **User Control:** Every action is reversible

By following the guidance of Nielsen Norman Group, Peter Morville, Daniel Tunkelang, and modern design patterns, olsen can provide a faceted browsing experience that sets the standard for photography asset management.

The photographer's mental model—**what, when, where, how**—maps perfectly to faceted navigation. We just need to make that mapping visible and delightful.

---

**Status:** Ready for review and refinement  
**Next Steps:** User feedback on priorities, usability testing plan approval, implementation timeline agreement  
**Version:** 3.0  
**Last Updated:** 2025-10-13