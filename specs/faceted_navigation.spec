# Olsen Faceted Navigation — Refined UI Spec (Right-Rail, Map/GPS Out of Scope)

**Version:** 2.1  
**Date:** 2025-10-06  
**Status:** Ready for implementation

**Mock screenshot:** [olsen_faceted_ui_mock.png](olsen_faceted_ui_mock.png)

---

## 1) Executive summary
Facets live on the **right rail**. Results in the center. A sticky top bar controls search, sort, and global actions. Active filters render as removable chips. Facet **counts exclude the current facet**. URLs are the single source of truth for state. **AND across facets, OR within a facet.** Map and GPS are **explicitly out of scope** for this release.

---

## 2) Scope
### In scope (core features — backend ready or minimal changes)
- Right-rail faceted navigation (desktop) + full-height filter tray (mobile).
- **Layout & UI components:** Top bar, chip row, results grid, right rail, mobile tray.
- **Facet types:** discrete lists (with counts), categorical chips, numeric ranges (sliders + inputs), booleans.
- **Facets (existing backend):** Year, Camera, Lens, Colour name, Time of day, Focal category, Shooting condition, In burst.
- **Facets (minor backend additions):** Month (progressive from Year), Orientation (landscape/portrait/square), Range facets (ISO, Aperture, Focal length).
- **Core features:** Sorting (date, camera, focal length, ISO, aperture), pagination, shareable URLs (deep links).
- **Quality:** Accessibility (WCAG AA), loading/empty/error states, crawl hygiene.

### Nice to have (backend work required — defer to later iterations)
- **Day facet** (progressive from Month).
- **Season facet** (spring/summer/fall/winter chips).
- **Lens make facet** (Canon/Fujifilm/Sony).
- **White balance facet** (auto/daylight/cloudy/tungsten/fluorescent).
- **Flash fired facet** (boolean toggle).
- **Colour space facet** (sRGB/Adobe RGB).
- **Burst group facet** (single select within burst).
- **Is burst representative facet** (boolean toggle).
- **Width/Height range facets**.
- **35mm equiv. focal length range**.
- **Advanced HSL sliders** (hue/sat/light ranges).
- **Facet search** (search within Camera/Lens lists).
- **Analytics instrumentation**.

### Out of scope (explicit)
- **Map facet** (pan/zoom map, "Use current view", bbox filtering).
- **GPS-driven filters** (`has_gps`, `lat_*`, `lon_*`) and any geospatial aggregations.
- Saved places / geo presets.

Rationale: Ship UI with working facets first; backend already supports most core filters. Add Month/Orientation/Ranges quickly; defer others to v2.2+.

---

## 3) Principles
- **URLs = state.** Every facet change updates the query string; deep links reproduce the view exactly.
- **Counts that guide, not trap.** For facet *F*, compute counts with all active filters **except** *F* (a.k.a. “self-exclusion”). Prevents dead-ends and encourages exploration.
- **Boolean algebra:** **AND** across facets; **OR** within a facet (multi-select via repeated params).
- **Progressive disclosure:** Show Month only after Year; Day only after Month. Keep the rail tidy.
- **Mobile clarity:** Use a full-height **tray** for filters; show an applied-count badge on the button.
- **A11Y:** Real fieldsets, keyboardable controls, visible focus, counts as text (not just color).

---

## 4) Information architecture

### Layout (desktop)
- **Top bar (sticky):** search field, result count, **Sort** dropdown, **Save view**, **Share**, **Clear all**.
- **Chip row:** `Label: Value` chips with × to remove; multi-select compresses to `Colour: 3 selected` with a popover for quick deselects.
- **Results grid:** responsive cards; infinite scroll or “Load more”.
- **Right rail:** facet sections (collapsible): Time, Equipment, Colour, Composition/Orientation, Capture conditions, Bursts, File/Space.

### Layout (mobile)
- Top bar with **Filters** button (badge shows active count).
- Full-height filter **tray** with the same sections as desktop; **Apply** and **Reset** at bottom. Applied chips sit below the header when the tray is closed.

---

## 5) Facet groups & UI ⇄ query contract

> Within a facet: **multi-select = OR** (repeat the param). Across facets: **AND**.

### A) Time
- **Year** (single): `?year=YYYY` (`-1` = Unknown) — ✅ **Core (backend ready)**
- **Month** (single; visible when `year` is set): `?month=MM` — ✅ **Core (minor backend addition)**
- **Day** (single; visible when `month` is set): `?day=DD` — 🔮 **Nice to have**
- **Time of day** (chips, multi): `?time_of_day=dawn|golden_am|midday|golden_pm|blue_hour` — ✅ **Core (backend ready)**
- **Season** (chips, multi): `?season=spring|summer|fall|winter` — 🔮 **Nice to have**

### B) Equipment
- **Camera** (list, single): `?camera=Canon%20EOS%20R5` — ✅ **Core (backend ready)**
- **Lens** (list, multi): `?lens=RF%2024-70mm` (repeatable) — ✅ **Core (backend ready)**
- **Lens make** (list, multi): `?lens_make=Canon|Fujifilm|Sony` — 🔮 **Nice to have**
- **Facet search** (search within Camera/Lens lists) — 🔮 **Nice to have**

### C) Colour
- **Colour name** (swatches, multi): `?color=red` (repeatable) — ✅ **Core (backend ready)**
- **Advanced HSL** (ranges): `?hue_min=&hue_max=&sat_min=&sat_max=&light_min=&light_max=` — 🔮 **Nice to have**

### D) Composition & Orientation
- **Orientation** (multi): `?orientation=landscape|portrait|square` — ✅ **Core (minor backend addition)**
- **Width/Height** (ranges): `?width_min=&width_max=&height_min=&height_max=` — 🔮 **Nice to have**

### E) Capture conditions
- **White balance** (multi): `?white_balance=auto|daylight|cloudy|tungsten|fluorescent` — 🔮 **Nice to have**
- **Flash fired** (toggle): `?flash_fired=true|false` — 🔮 **Nice to have**

### F) Bursts
- **In burst** (toggle): `?in_burst=true` — ✅ **Core (backend ready)**
- **Burst group** (single; visible only when `in_burst=true`): `?burst_group_id=<id>` — 🔮 **Nice to have**
- **Is representative** (toggle): `?is_burst_representative=true` — 🔮 **Nice to have**

### G) File / Space
- **Colour space** (multi): `?color_space=sRGB|Adobe%20RGB` — 🔮 **Nice to have**
- **ISO** (range): `?iso_min=&iso_max=` — ✅ **Core (minor backend addition)**
- **Aperture** (range): `?aperture_min=&aperture_max=` — ✅ **Core (minor backend addition)**
- **Focal length** (range): `?focal_length_min=&focal_length_max=` — ✅ **Core (minor backend addition)**
- **35mm equiv.** (optional range): `?focal_length_eq_min=&focal_length_eq_max=` — 🔮 **Nice to have**

---

## 6) Interaction rules
- **Selecting a value** appends its param (or replaces it for single-select facets).
- **Counts exclude self:** for facet *F*, compute counts with all active filters **except** *F* (prevents “0” traps).
- **Unknowns:** Year shows **Unknown** when `date_taken` is NULL; swatches include **B&W**.
- **Range facets:** sliders + numeric inputs; chips display as `ISO: 100–800`.
- **List sorting:** default **Count**; toggle to **A–Z** per facet.
- **Empty results:** show recovery UI (remove last filter, Clear all).

---

## 7) URLs, sorting, pagination
- Paths stay flat; **query string is the state**:  
  - `/photos?year=2025&color=red&camera=Canon%20EOS%20R5`  
  - `/photos?hue_min=0&hue_max=30&iso_min=100&iso_max=800&orientation=portrait`
- **Sorting:** `?sort=date_taken|camera|focal_length|iso|aperture&order=asc|desc` (default `date_taken desc`)
- **Pagination:** `?limit=50&offset=0` (default limit 50)

---

## 8) Facet API (server → UI)
Each facet returns values like:
```json
[
  {
    "label": "Canon EOS R5",
    "value": "Canon EOS R5",
    "count": 420,
    "selected": false,
    "url": "/photos?year=2025&camera=Canon%20EOS%20R5"
  }
]
```
Rules:
- `url` is the **next state** if the user toggles this value.
- Selected values pin to the top; others order by **count** unless the facet is in **A–Z** mode.
- Numeric facets also return `{ "min", "max", "selectedMin", "selectedMax" }`.

---

## 9) Performance budgets
- Single facet recompute: **< 50 ms**
- All facets (full set): **100–200 ms** on ~10k photos
- Indexes: date (year, month, day), camera, lens, HSL buckets, burst flags, width/height, ISO, aperture.

---

## 10) Accessibility
- `<fieldset><legend>` per facet group; all controls keyboardable.
- Visible focus states; WCAG AA color contrast; counts always textual (not just color).
- Mobile tray uses dialog semantics (focus trap, aria labels).

---

## 11) Visual states
- **Loading:** grid skeletons; dimmed facet counts with shimmer.
- **No results:** message + last 3 chips with inline remove + “Clear all”.
- **Errors:** preserve state; inline retry; never auto-clear chips.
- **Right rail:** sections are collapsible; default open: **Time**, **Equipment**, **Colour**.

---

## 12) SEO & crawl hygiene
- Add `noindex, follow` on highly faceted states (≥ 2 facet keys or ≥ 4 total params).
- Provide a stable `rel=canonical` to the base results (or the minimum informative state).
- Consider disallowing obvious crawler traps for heavy facet combos.

---

## 13) Analytics
Log on each view:
- Active facet keys + counts
- Time‑to‑first‑facet and time‑to‑first‑result
- Drop-offs after applying facets
- Facet searches (Camera/Lens) query strings and success rate

---

## 14) Testing
- **Unit:** self-exclusion per facet type; URL toggle/add/remove; range clamping.
- **Integration:** deep-link hydration; mobile tray a11y; empty states.
- **Perf:** ensure budgets under realistic DB sizes.

---

## 15) Deliverables
- **Right-rail UI** per mock: `olsen_faceted_ui_mock.png` (in this folder).
- Component checklist: FacetList, RangeFacet, ChipBar, SortDropdown, FilterTray (mobile).

---

## 16) References (external)
- **Datasette — Facets (docs):** https://docs.datasette.io/en/stable/facets.html  
- **Datasette — Facets (older version docs):** https://docs.datasette.io/en/0.56/facets.html  
- **Simon Willison — “Datasette Facets” (blog):** https://simonwillison.net/2018/May/20/datasette-facets/  
- **Simon Willison — “Bad bots” (facet crawl concerns):** https://simonwillison.net/2025/Oct/6/bad-bots/  
- **NN/g — Filters vs. Facets:** https://www.nngroup.com/articles/filters-vs-facets/  
- **NN/g — Mobile Faceted Search with a Tray:** https://www.nngroup.com/articles/mobile-faceted-search/  
- **NN/g — Filter categories & values:** https://www.nngroup.com/articles/filter-categories-values/  
- **Baymard — Applied filters overview:** https://baymard.com/blog/how-to-design-applied-filters  
- **Baymard — Multiple values per filter:** https://baymard.com/blog/allow-applying-of-multiple-filter-values  
- **Algolia — Faceted search (overview):** https://www.algolia.com/blog/ux/faceted-search-an-overview  
- **Algolia — `facets` API parameter:** https://www.algolia.com/doc/api-reference/api-parameters/facets  
- **Algolia — Search for facet values:** https://www.algolia.com/doc/libraries/sdk/v1/methods/search-for-facet-values  
- **Elasticsearch — `post_filter` (aggregations unaffected by filters):** https://www.elastic.co/docs/reference/elasticsearch/rest-apis/filter-search-results

---

## 17) Notes on the uploaded spec
- Your original spec (v2.0) references internal research in `docs/FACETED_NAVIGATION_PLAN.md` but lists no external URLs. The “References (external)” above cover the external sources requested and align with the engine and UI patterns in this refinement.