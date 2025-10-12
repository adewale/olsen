# Faceted UI Implementation Specification

## Core Principle: State Machine Model

**Faceted navigation is a state machine where users explore a dataset through valid state transitions.**

### The Fundamental Rule

> **Users must never be able to transition from a state with results (count > 0) to a state with zero results (count = 0).**

This is NOT about hierarchical relationships between facets (e.g., "Year contains Month"). It's about preventing invalid transitions through the data space.

**Key Insights:**
- Facets are **independent dimensions** that can be combined
- The **data determines** which combinations are valid, not hardcoded rules
- SQL queries with WHERE clauses + GROUP BY **naturally compute** valid transitions
- UI **disables** invalid transitions (count = 0), doesn't hide them

See `specs/facet_state_machine.spec` for detailed explanation.

## 1. Data Model Requirements

### 1.1 Facet Structure
- Each facet MUST have: `id`, `label`, `type` (single-select | multi-select), `values[]`
- Each facet value MUST have: `id`, `label`, `count`, `enabled` state
- Facets MUST reflect user mental models, validated through user research
- Facet values MUST include result counts showing items available when that filter is applied
- **Facet values with count = 0 MUST be shown as disabled**, not hidden

### 1.2 State Management - State Machine Model
- Filter state MUST be fully represented in URL parameters
- URL structure: `?category=electronics&brand=apple,samsung&price=100-500`
- Browser back/forward navigation MUST work correctly with filter states
- Bookmarked/shared URLs MUST restore exact filter state
- **ALL filters are preserved during transitions** - no automatic clearing
- **Invalid transitions (count = 0) are prevented at UI level**, not by clearing filters

## 2. Performance Requirements

### 2.1 Count Calculation
- Facet counts MUST update dynamically as filters are applied
- Implement caching strategy with TTL based on data update frequency
- Cache key format: `facets:{category}:{filter_hash}:{page}`

### 2.2 Request Management
- MUST implement debouncing with 300ms delay for filter changes
- MUST cancel in-flight requests when new filters are applied
- Batch multiple filter changes into single API request

### 2.3 Lazy Loading
- Initial load: Display first 10-20 values for facets with >20 options
- Implement "Show more" / "Show less" controls
- Load additional values in chunks of 20
- For facets with >100 values, MUST provide search-within-facet functionality

## 3. UI/UX Requirements

### 3.1 Layout Structure
```
Desktop:
- Left sidebar: 200-250px fixed width
- Show 5-7 primary facets initially
- "Show more facets" link for additional facets
- Sticky positioning when scrolling (optional)

Mobile:
- Hidden by default
- Single "Filters" button in header/toolbar
- Full-screen overlay when activated
```

### 3.2 Active Filter Display
- MUST show active filters in THREE places:
  1. Within facet panel (checked/selected state)
  2. As removable chips above result list
  3. In breadcrumb trail (optional for simple implementations)
- Each active filter chip MUST have clear remove/close action

### 3.3 Filter Controls
- Checkboxes for multi-select facets
- Radio buttons OR single-select dropdown for single-select facets
- "Clear all" action MUST be present when any filters are active
- Individual filter removal via clicking chips or unchecking boxes

### 3.4 Zero Results Handling
```javascript
// When results.length === 0
{
  showMessage: "No results found with current filters",
  showActiveFilters: true,
  suggestActions: [
    "Remove most restrictive filter: {filter_name}",
    "Clear all filters",
    "Try broader category"
  ]
}
```

## 4. Interaction Behaviors

### 4.1 Desktop Behavior
- Filters apply immediately on selection (live updating)
- Optional: 100ms transition/fade for result updates
- Maintain scroll position when possible
- Show loading indicator during updates

### 4.2 Mobile Behavior
- Two-step process required:
  1. Select filters in overlay
  2. Tap "Apply Filters" button to execute
- Show filter count badge: "Apply Filters (3)"
- "Clear" button in overlay header
- Swipe down or X button to cancel without applying

### 4.3 Facet Sorting
```javascript
// Priority order for facet value sorting
sortingStrategy: {
  price: 'dynamic-buckets',      // Calculate from actual distribution
  category: 'result-count-desc',  // Most results first
  brand: 'popularity',            // Based on click/purchase data
  color: 'custom-order',          // Define specific order
  default: 'alphabetical'
}
```

## 5. Advanced Features

### 5.1 State Machine Transitions (NOT Hierarchical Dependencies)
**IMPORTANT:** Facets do NOT have hierarchical dependencies. They are independent dimensions.

**Correct Model:**
- When filter A is selected, facet B values are recomputed with A in the WHERE clause
- SQL naturally returns only valid combinations (those with count > 0)
- Invalid combinations (count = 0) are shown as **disabled**, not hidden
- Show tooltip: "No results with current filters"

**Incorrect Model (NEVER DO THIS):**
- ❌ "Changing Year clears Month because Year contains Month hierarchically"
- ❌ "Changing Camera clears Lens because Camera contains Lens"
- ❌ Hardcoded clearing logic based on assumed relationships

**Why This Matters:**
- Data determines valid combinations, not assumptions
- Same rule applies to ALL facets (temporal, equipment, visual)
- System scales automatically when new facets are added
- No special cases needed

See `specs/facet_state_machine.spec` and `docs/HIERARCHICAL_FACETS.md` for detailed explanation.

### 5.2 Negative Filtering
- Implement exclude option: "NOT" or "-" prefix
- UI: Long-press or right-click to exclude (desktop)
- Mobile: Toggle between include/exclude mode

### 5.3 Smart Defaults
- Track filter combinations with analytics
- Surface popular combinations as "Quick filters" or "Suggested filters"
- Remove facets with <1% usage after 30 days of data

## 6. State Persistence Rules

### 6.1 When to Maintain Filters
- Pagination
- Sorting changes
- View mode changes (grid/list)
- Page refresh

### 6.2 When to Clear Filters
- Explicit "Clear all" action
- Navigation to different category/section
- New search query entered

## 7. Analytics Requirements

### 7.1 Required Tracking Events
```javascript
// Track these events
{
  'facet_opened': { facet_id, position },
  'facet_applied': { facet_id, value, result_count },
  'facet_removed': { facet_id, value, method },
  'facet_cleared_all': { filter_count },
  'facet_search': { facet_id, query },
  'zero_results': { active_filters },
  'filter_combination': { filters[], result_count }
}
```

### 7.2 Metrics to Monitor
- Facet usage rate by position
- Filter abandonment rate
- Most common filter combinations
- Zero result frequency by filter
- Time to first filter interaction

## 8. Accessibility Requirements

- All facets keyboard navigable
- ARIA labels for screen readers
- Focus management when opening/closing facets
- Announce result count changes
- High contrast mode support

## 9. Error Handling

### 9.1 Failed Facet Load
- Show facets panel with message: "Filters temporarily unavailable"
- Allow browsing without filters
- Retry loading every 5 seconds (max 3 attempts)

### 9.2 Failed Filter Apply
- Maintain previous state
- Show error message
- Provide retry action
- Log error for monitoring

## 10. Implementation Checklist

### Core Features (MVP)
- [ ] URL state synchronization
- [ ] Basic facet rendering (checkbox/radio)
- [ ] Active filter chips
- [ ] Clear all functionality
- [ ] Mobile filter overlay
- [ ] Facet counts
- [ ] Zero results message

### Enhanced Features (Phase 2)
- [ ] Lazy loading for long lists
- [ ] Search within facet
- [ ] Facet dependencies
- [ ] Smart sorting
- [ ] Analytics integration
- [ ] Negative filtering
- [ ] Quick filter suggestions

### Performance Optimizations (Phase 3)
- [ ] Request debouncing
- [ ] Request cancellation
- [ ] Caching layer
- [ ] Dynamic bucketing
- [ ] Virtual scrolling for long lists

## 11. Testing Requirements

### 11.1 Unit Tests
- Filter state management
- URL serialization/deserialization
- Facet count calculations
- Sort algorithms

### 11.2 Integration Tests
- Filter application -> API call -> Result update
- Multi-filter combinations
- Clear filters functionality
- Mobile overlay interactions

### 11.3 E2E Tests
- Complete filter journey
- Deep linking with filters
- Browser back/forward
- Zero results scenarios
- Mobile filter application

### 11.4 Performance Tests
- Load time with 100+ facet values
- Response time with 10+ active filters
- Memory usage with extended filtering sessions