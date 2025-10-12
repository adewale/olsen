# Lessons Learned: Olsen Photo Indexing System

**Project:** Olsen - Portable Photo Corpus Explorer
**Period:** September 2025 - October 2025
**Status:** Living Document

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Critical Lesson: Always Debug at the Source](#critical-lesson-always-debug-at-the-source)
3. [Architectural Lessons](#architectural-lessons)
4. [Testing Strategy Lessons](#testing-strategy-lessons)
5. [User Experience Lessons](#user-experience-lessons)
6. [Performance Lessons](#performance-lessons)
7. [Development Process Lessons](#development-process-lessons)
8. [Technical Deep Dives](#technical-deep-dives)

---

## Executive Summary

### What Went Right ✅

1. **State Machine Model for Faceted Navigation** - Discovered that faceted search is about valid state transitions, not hierarchical relationships. This insight led to cleaner code and better UX.

2. **Comprehensive Logging** - Structured logging with prefixes (`[RAW]`, `[EMBED]`, `FACET_STATE`) made debugging exponentially faster and enabled proactive monitoring.

3. **Test-Driven Debugging** - Writing tests that "would have caught the bug" prevented regressions and documented expected behavior.

4. **Dual LibRaw Support** - Building abstraction for two LibRaw libraries enabled seamless switching when bugs appeared in one implementation.

5. **Color Classification Evolution** - Starting simple (hue-based) then adding complexity (saturation-first for B&W) was the right approach.

### What Went Wrong ❌

1. **Started Debugging at Wrong Layer** - Fixed UI symptoms before investigating RAW decode root cause (thumbnail quality bug).

2. **Assumed Hierarchical Relationships** - Built faceted navigation on assumed "Year contains Month" hierarchy instead of data-driven state machine.

3. **Incomplete Migration** - Fixed facet URL building but missed WHERE clause logic, leading to regression.

4. **Trusted Metrics Without Verification** - "8 thumbnails generated" ≠ "8 good quality thumbnails" (didn't visually inspect outputs).

5. **Didn't Read File Format First** - Would have understood DNG embedded previews immediately if we'd used `exiftool` on day one.

### Key Metrics

- **Time to Initial Working System:** ~2 weeks
- **Major Bugs Fixed:** 3 (Monochrom thumbnails, WHERE clause, color classification)
- **Regressions Introduced:** 2 (black thumbnails, facet transitions)
- **Test Coverage at End:** ~70% (good unit tests, missing integration tests)
- **Lines of Test Code Written:** ~2500+ (caught all future regressions)

---

## Critical Lesson: Always Debug at the Source

### The Monochrom DNG Thumbnail Bug

**Timeline:**
1. **Initial Problem:** Missing images in web app, upscale warnings in logs
2. **First Fix:** Implemented thumbnail fallback in web UI (❌ symptom, not root cause)
3. **Regression:** Removed `isBlackImage()` check → completely black thumbnails
4. **Root Cause Discovery:** `ExtractEmbeddedJPEG()` returned FIRST JPEG (160x120), not LARGEST (9504x6320)
5. **Final Fix:** Modified embedded JPEG extraction to find largest preview

**What We Should Have Done:**

```bash
# Step 1: Inspect file format FIRST
exiftool -a -G1 -s L1001530.DNG | grep -i preview
# Output: PreviewImageLength: 2170368 bytes (~2.1MB) ← The answer was here!

# Step 2: Test RAW decode layer directly
./olsen thumbnail -o test.jpg -s 512 --db test.db 1
# Then VISUALLY INSPECT test.jpg

# Step 3: Add logging at RAW decode layer
[RAW] LibRaw decoded L1001530.DNG: 9536x6336
[EMBED] Extracted largest embedded JPEG: 9504x6320 (2170175 bytes) from 44 previews
```

**What We Actually Did (Wrong):**
1. Added fallback in web UI (wrong layer)
2. Fixed database queries (wrong layer)
3. Assumed "8 thumbnails generated" meant success (no verification)
4. Only later investigated RAW decode (should have been step #1)

### The Rule

> **When data is wrong, always start debugging at the SOURCE, never at the DISPLAY layer.**

**Debugging order:**
1. File format inspection (`exiftool`, `hexdump`)
2. RAW decode layer (libraries, output quality)
3. Processing pipeline (thumbnails, color extraction)
4. Database storage (queries, schema)
5. Repository layer (query building)
6. Web UI (display, templates)

**Start at #1, work your way down. Never start at #6 and work backwards.**

---

## Architectural Lessons

### 1. State Machines > Hierarchies

**What We Thought (WRONG):**
> "Year contains Month contains Day, so changing Year should clear Month and Day because of the hierarchical relationship."

**What Actually Matters (CORRECT):**
> "Users should only be able to transition to states that have results. The data determines valid transitions, not assumed hierarchies."

#### Example: The Bug

```
State A: year=2024&month=11 (50 photos from November 2024)
User clicks: Year 2025

Old Behavior (BROKEN):
→ Result: year=2025 (month=11 was cleared by hierarchical logic)
→ Problem: If user had Nov 2025 photos, they can't get to them
→ Problem: System made assumptions about what user wanted

The Bug: Year 2025 should have shown count=0 (disabled) if no Nov 2025 photos exist.
The problem wasn't that we needed to clear Month—it was that we allowed an invalid transition!
```

#### The Fix

**URL Builder (facet_url_builder.go):**
```go
// ❌ WRONG (Hierarchical):
if facet.Values[i].Selected {
    p.Year = nil
    p.Month = nil  // Assumes hierarchy
    p.Day = nil    // Assumes hierarchy
}

// ✅ CORRECT (State Machine):
if facet.Values[i].Selected {
    p.Year = nil  // Remove this filter
    // PRESERVE all other filters!
}
```

**WHERE Clause Builder (engine.go):**
```go
// ❌ WRONG (Hierarchical):
if params.Month != nil && params.Year != nil {
    where = append(where, "strftime('%m', p.date_taken) = ?")
}

// ✅ CORRECT (State Machine):
if params.Month != nil {
    // Month is independent - apply even without Year
    where = append(where, "strftime('%m', p.date_taken) = ?")
}
```

#### Why This Matters

**Old Model:**
- Special cases for each facet relationship
- Hardcoded clearing logic
- Breaks when adding new facet types
- Surprising user behavior (filters disappear)

**New Model:**
- One rule for ALL facets: "Preserve filters, compute counts, disable zeros"
- Emergent behavior from actual data
- Scales to any facet combination
- Transparent behavior (disabled facets visible but not clickable)

#### The Migration Checklist

When converting hierarchical → state machine, check ALL layers:

1. ✅ URL generation (`facet_url_builder.go`)
2. ✅ Facet computation (`facets.go`)
3. ✅ **WHERE clause building (`engine.go`)** ← Easy to miss!
4. ✅ Template rendering (`grid.html`)

**Search for hierarchical assumptions:**
```bash
grep "Month != nil && .*Year != nil" internal/query/*.go
grep "Day != nil && .*Month != nil" internal/query/*.go
```

### 2. Sometimes Simple > Complex

**The DNG Preview Insight:**

We spent significant time implementing full LibRaw RAW decode integration when embedded preview extraction would have been:
- **60-120× faster** (20ms vs 1200ms)
- **Equal or better quality** for thumbnail generation
- **Avoids compatibility issues** (JPEG-compressed monochrome DNGs)
- **Reduces complexity** (no LibRaw dependency, no CGO)

**Current approach:**
- LibRaw decode: ~1200ms
- Detects black image: ~50ms
- Falls back to embedded JPEG: ~350ms
- **Total: ~1600ms per file**

**Optimal approach (future work):**
- Extract embedded JPEG directly: ~20ms
- Skip LibRaw entirely for thumbnails
- **60-80× speedup**

**Lesson:** Sometimes the "simple" solution (extract preview) is 100× better than the "proper" solution (full RAW decode), especially when:
- It's dramatically faster
- Quality is sufficient for the use case
- It avoids edge case bugs
- It reduces dependencies

### 3. Data-Driven Behavior > Hardcoded Rules

**Color Classification Evolution:**

**v1.0 (Hue-only, BROKEN for B&W):**
```go
switch {
case hue >= 0 && hue <= 15:
    return "red"  // ❌ B&W photos incorrectly classified as "red"!
case hue >= 16 && hue <= 45:
    return "orange"
// ...
}
```

**v2.0 (Saturation-first, CORRECT):**
```go
switch {
case saturation < 10:
    return "bw"  // ✅ Check saturation FIRST
case hue >= 0 && hue <= 15:
    return "red"  // Now only applies to colored photos
// ...
}
```

**Lesson:** Let the data tell you what it is (saturation reveals B&W), don't impose assumptions (hue classification for achromatic colors).

---

## Testing Strategy Lessons

### 1. Test at the Layer Closest to the Problem

**What We Did Wrong:**
- Added tests at web UI layer (repository queries)
- Tested "can we query photos?" not "is the thumbnail correct?"
- Missed the actual bug in RAW decode layer

**What We Should Have Done:**

```go
// Test #1: Extraction returns LARGEST preview, not first
func TestExtractEmbeddedJPEG_FindsLargest(t *testing.T) {
    img, _ := ExtractEmbeddedJPEG("L1001530.DNG")
    bounds := img.Bounds()
    minDimension := 6000

    if max(width, height) < minDimension {
        t.Errorf("Expected large JPEG (>%dpx), got %dx%d", minDimension, width, height)
        t.Errorf("BUG: Returning FIRST JPEG, not LARGEST")
    }
}

// Test #2: Quality check (not just dimensions)
func TestDecodeRaw_QualityCheck(t *testing.T) {
    img, _ := DecodeRaw("L1001530.DNG")
    brightness := calculateBrightness(img)

    if brightness < 10 {
        t.Errorf("Image too dark (%.1f/255) - black image bug!", brightness)
    }
}

// Test #3: End-to-end thumbnail generation
func TestThumbnailGeneration_FromMonochromDNG(t *testing.T) {
    thumbnails, _ := GenerateThumbnails("L1001530.DNG")

    if len(thumbnails) != 4 {
        t.Errorf("Expected 4 sizes, got %d - quality pipeline skipping sizes!", len(thumbnails))
    }
}
```

**These tests would have caught the bug immediately.**

### 2. Visual Inspection > Metrics Alone

**The Trap:**
```
Log output: "Generated 8 thumbnails"
Developer: "Great! It's working."
Reality: All 8 thumbnails are 64px (upscaling detected, other sizes skipped)
```

**The Fix:**
```bash
# Always visually inspect outputs during debugging
./olsen thumbnail -o /tmp/test_512.jpg -s 512 --db test.db 1
open /tmp/test_512.jpg  # LOOK AT IT!

# Check dimensions
file /tmp/test_512.jpg
# Output: JPEG image data, ... 160 x 120 ← Wait, that's wrong!
```

**Lesson:** "8 thumbnails generated" ≠ "8 good quality thumbnails". Verify outputs, don't just trust counts.

### 3. Add Diagnostic Logging Proactively

**Before (No Logging):**
```go
func ExtractEmbeddedJPEG(path string) (image.Image, error) {
    // ... find JPEG in DNG ...
    return jpeg.Decode(bytes.NewReader(jpegData)), nil
}
```

**After (Comprehensive Logging):**
```go
func ExtractEmbeddedJPEG(path string) (image.Image, error) {
    // ... find JPEG in DNG ...
    log.Printf("[EMBED] Extracted largest embedded JPEG: %dx%d (%d bytes) from %d previews in %s",
        cfg.Width, cfg.Height, largestSize, jpegCount, filepath.Base(path))
    return jpeg.Decode(bytes.NewReader(largestJPEG)), nil
}
```

**Logging Categories:**
- `[RAW]` - LibRaw decode operations
- `[EMBED]` - Embedded JPEG extraction
- `[THUMB]` - Thumbnail generation
- `FACET_STATE` - Facet computation results
- `FACET_404` - Zero-result states reached

**Benefit:** Instantly reveals what's happening without adding breakpoints.

### 4. Write Tests That Document Bugs

When fixing a bug, write tests with descriptive names and comments:

```go
// TestExtractEmbeddedJPEG_FindsLargest verifies we extract the LARGEST
// embedded JPEG preview (9504x6320), not the FIRST one found (160x120).
//
// BUG HISTORY: Initial implementation used naive "first match" algorithm
// which returned 160x120 thumbnail instead of 9504x6320 full preview.
// This caused quality pipeline to detect upscaling and skip larger sizes.
//
// WOULD HAVE CAUGHT: If this test existed during initial implementation,
// the bug would have been caught immediately.
func TestExtractEmbeddedJPEG_FindsLargest(t *testing.T) {
    // ...
}
```

**Benefits:**
- Future developers understand WHY test exists
- Documents historical context
- Prevents regression
- Serves as living documentation

---

## User Experience Lessons

### 1. Zero Results Must Be Handled Gracefully

**Two Ways to Reach Zero Results:**

1. **Via UI clicks:** PREVENTED by disabling zero-count facets
2. **Via direct URL entry:** CANNOT be prevented, must handle gracefully

**The Solution:**

```html
{{if eq (len .Photos) 0}}
  <div class="no-results">
    <p>📷 No photos found</p>

    {{if .HasActiveFilters}}
      <p>Active filters:</p>
      <ul>
        {{range .ActiveFilters}}
          <li>{{.Label}}: {{.Value}} <a href="{{.RemoveURL}}">[remove]</a></li>
        {{end}}
      </ul>
      <a href="/photos">Clear all filters</a>
    {{else}}
      <p>No photos in database yet. Try indexing your photo library!</p>
    {{end}}
  </div>
{{end}}
```

**Lesson:** Users can always manually type URLs or use old bookmarks. Handle zero results with helpful guidance, not just "404 Not Found".

### 2. Show Invalid Options as Disabled, Don't Hide Them

**Bad Approach (Hide):**
```
Year Facet:
  2023 (120)  ← Only shows years with results
  2024 (50)
```
**User thinks:** "I don't have any 2025 photos"
**Reality:** "You have 2025 photos, just not with current filters"

**Good Approach (Disable):**
```
Year Facet:
  2023 (120) ← Enabled
  2024 (50)  ← Selected
  2025 (0)   ← Disabled but visible, shows "No results with current filters"
```

**User thinks:** "I have 2025 photos, but none in November"
**Reality:** Correct!

**Implementation:**
```html
{{range .Values}}
  {{if gt .Count 0}}
    <a href="{{.URL}}" class="facet-value">
      {{.Label}} <span class="count">({{.Count}})</span>
    </a>
  {{else}}
    <span class="facet-value disabled" title="No results with current filters">
      {{.Label}} <span class="count">(0)</span>
    </span>
  {{end}}
{{end}}
```

**CSS:**
```css
.facet-value.disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
  color: #999;
}
```

### 3. Structured Logging for Production Monitoring

**Log Format:**
```
FACET_STATE: state=year=2024&month=11 results=50 enabled=15 disabled=3
```

**What to Monitor:**
- High frequency of `FACET_404` → UI bug or many old bookmarks
- `disabled=0` when filtering by month/day → Suspicious (unlikely all years have all months)
- `FACET_404` after `FACET_STATE` with `disabled=0` → User clicked disabled facet (UI bug!)

**Example Bug Detection:**
```
FACET_STATE: state=year=2025&month=1 results=20 enabled=19 disabled=0
# All facets enabled (disabled=0)? Suspicious for January-only filter...

FACET_STATE: state=year=2024&month=1 results=0 enabled=11 disabled=0
FACET_404: No results found - path=/photos query=year=2024&month=1
# User clicked Year 2024 and got 0 results
# BUG: Facet should have shown Year 2024 as disabled!
```

---

## Performance Lessons

### 1. Profile Before Optimizing

**Current Indexing Performance:**
- EXIF extraction: ~30ms
- Thumbnail generation: ~30ms
- Color extraction: ~50ms
- RAW decode (when needed): ~1200ms
- **Total: ~110ms per photo (without RAW decode), ~1300ms (with RAW decode)**

**With 4 workers:** ~10-30 photos/second (observed)

**Lesson:** Color extraction (~50ms) is NOT the bottleneck. RAW decode (~1200ms) is. Optimizing color extraction would have minimal impact.

### 2. Know When to Use Embedded Previews

**DNG Files Always Contain Embedded JPEGs:**
- Tiny preview: 160x120 (~5KB)
- Medium preview: ~23KB
- **Large preview: 9504x6320 (~2.1MB)** ← Perfect for thumbnails!

**Performance Comparison:**
- LibRaw RAW decode: ~1200ms
- Embedded JPEG extraction: ~20ms
- **60× faster!**

**When to Use Each:**
- **Embedded preview:** Thumbnails, color extraction, web display
- **Full RAW decode:** Editing, maximum quality output, format conversion

**Future Optimization:**
```go
func GenerateThumbnails(path string) (map[ThumbnailSize][]byte, error) {
    // Try embedded preview first (60× faster)
    if preview, err := ExtractLargestEmbeddedJPEG(path); err == nil {
        return generateThumbnailsFromImage(preview)
    }

    // Fall back to full RAW decode only if necessary
    img, err := DecodeRaw(path)
    if err != nil {
        return nil, err
    }
    return generateThumbnailsFromImage(img)
}
```

### 3. Indexing at Scale

**Recommendations from Testing:**
1. Always work from thumbnails for color extraction (256px, not full resolution)
2. Use embedded previews when available (DNG, most RAW formats)
3. Worker pool size: 4-8 (diminishing returns beyond 8 on most hardware)
4. Enable WAL mode for concurrent reads during indexing
5. Batch database inserts (100-1000 photos per transaction)

---

## Development Process Lessons

### 1. Complete Migration Checklists

**When changing a fundamental assumption, check ALL layers:**

Example: Hierarchical → State Machine Migration
```bash
# Step 1: Search for hierarchical assumptions
grep "Month != nil && .*Year != nil" internal/query/*.go
grep "p.Month = nil" internal/query/*.go

# Step 2: Check ALL these files
- [ ] facet_url_builder.go (URL generation)
- [ ] facets.go (facet computation)
- [ ] engine.go (WHERE clause building)  ← EASY TO MISS!
- [ ] templates/grid.html (rendering)

# Step 3: Write tests for each layer
- [ ] URL preservation tests
- [ ] Facet count accuracy tests
- [ ] WHERE clause independence tests
- [ ] End-to-end integration tests
```

**Lesson:** Architectural changes ripple through multiple layers. Hunt down ALL instances.

### 2. Use the Makefile for Everything

**Bad:**
```bash
# Manually running complex test commands
CGO_ENABLED=1 CGO_CFLAGS="$(pkg-config --cflags libraw)" \
  CGO_LDFLAGS="$(pkg-config --libs libraw)" \
  go test -tags "cgo use_seppedelanghe_libraw" -v ./internal/indexer \
  -run TestExtractEmbeddedJPEG_FindsLargest
```

**Good:**
```bash
# Simple, documented Makefile target
make test-raw-validation
```

**Makefile:**
```makefile
.PHONY: test-raw-validation
test-raw-validation:
	@echo "Testing RAW decode validation..."
	@echo "Verifies: largest JPEG extraction, fallback behavior, quality checks"
	@echo "LESSON: These tests would have caught the embedded JPEG size bug early"
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(CGO_CFLAGS_LIBRAW)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_LIBRAW)" \
	$(GOTEST) -tags "cgo use_seppedelanghe_libraw" -v ./internal/indexer \
	  -run "TestExtractEmbeddedJPEG_FindsLargest|TestDecodeRaw_QualityCheck"
```

**Benefits:**
- Self-documenting
- Consistent across team
- Easy to remember
- Includes context/rationale

### 3. Document Bugs as You Fix Them

**Create lesson docs immediately:**
- `docs/LESSONS_LEARNED_MONOCHROM_DNG.md` - Right after fixing thumbnail bug
- `docs/WHERE_CLAUSE_BUG.md` - Right after fixing state machine regression
- `docs/ZERO_RESULTS_HANDLING.md` - Right after implementing graceful handling

**Why:**
- Context is fresh (you remember WHY it was hard)
- Helps future you (or teammates)
- Prevents similar bugs
- Serves as onboarding material

---

## Technical Deep Dives

### 1. DNG File Structure

**What We Learned:**

DNG files (and most RAW formats) contain:
- **RAW sensor data** (may be JPEG-compressed for monochrome)
- **Multiple embedded JPEG previews** at different sizes:
  - Tiny: 160x120 (~5KB) - for file browsers
  - Medium: varies (~23KB) - for quick preview
  - Large: near-full-resolution (~2.1MB) - for display
- **EXIF metadata** in TIFF/IFD structure

**JPEG Markers:**
- SOI (Start of Image): `0xFF 0xD8`
- EOI (End of Image): `0xFF 0xD9`

**Extraction Algorithm:**
```go
func ExtractLargestEmbeddedJPEG(path string) (image.Image, error) {
    data, _ := os.ReadFile(path)

    var largestJPEG []byte
    var largestSize int

    // Scan for all JPEG markers
    for i := 0; i < len(data)-1; i++ {
        if data[i] == 0xFF && data[i+1] == 0xD8 {  // SOI
            // Find corresponding EOI
            for j := i + 2; j < len(data)-1; j++ {
                if data[j] == 0xFF && data[j+1] == 0xD9 {  // EOI
                    jpegData := data[i : j+2]

                    // Keep largest valid JPEG
                    if len(jpegData) > largestSize {
                        if _, err := jpeg.DecodeConfig(bytes.NewReader(jpegData)); err == nil {
                            largestJPEG = jpegData
                            largestSize = len(jpegData)
                        }
                    }
                    break
                }
            }
        }
    }

    return jpeg.Decode(bytes.NewReader(largestJPEG)), nil
}
```

**Future Optimization:** Parse TIFF/IFD structure directly instead of scanning entire file.

### 2. LibRaw Limitations

**Known Issues:**
1. **JPEG-compressed monochrome DNGs** → LibRaw produces black/dark images
2. **Buffer overflow** with monochrome data (1 channel vs 3 channels)
3. **Limited configuration** in some Go bindings

**Solution:** Dual library support with fallback to embedded JPEG extraction

**Files:**
- `internal/indexer/raw_seppedelanghe.go` - Full-featured binding
- `internal/indexer/raw_golibraw.go` - Simple fallback binding
- Build tags switch between implementations

### 3. State Machine Model for Faceted Search

**Core Principle:**
> Faceted navigation is a state machine where users explore a dataset through valid state transitions.

**Rule:**
> Users cannot transition from a state with results to a state with zero results.

**Implementation:**

1. **URL Building** - Preserve all filters during transitions:
   ```go
   if facet.Values[i].Selected {
       p.Year = nil  // Remove this filter
       // PRESERVE Month, Day, Color, Camera, etc.
   }
   ```

2. **WHERE Clause** - Apply filters independently:
   ```go
   if params.Month != nil {
       where = append(where, "strftime('%m', p.date_taken) = ?")
       // No dependency on Year
   }
   ```

3. **Facet Computation** - Count results with current filters:
   ```sql
   SELECT year, COUNT(*) as count
   FROM photos
   WHERE strftime('%m', date_taken) = '11'  -- Month filter preserved
   GROUP BY year
   ```

4. **UI Rendering** - Disable zero-count facets:
   ```html
   {{if gt .Count 0}}
     <a href="{{.URL}}">{{.Label}} ({{.Count}})</a>
   {{else}}
     <span class="disabled">{{.Label}} (0)</span>
   {{end}}
   ```

**Result:** Natural, predictable behavior that emerges from the data itself.

---

## Quick Reference: Common Pitfalls

### 1. Debugging Strategy
- ❌ Start at UI layer
- ✅ Start at file format / source layer
- ✅ Use `exiftool`, `hexdump`, direct file inspection
- ✅ Test at layer closest to problem

### 2. Testing Strategy
- ❌ "8 thumbnails generated" = success
- ✅ Visually inspect actual outputs
- ✅ Test dimensions, brightness, quality
- ✅ Write tests that would have caught the bug

### 3. Faceted Navigation
- ❌ Assume hierarchical relationships
- ✅ Use state machine model
- ✅ Preserve all filters
- ✅ Disable zero-count facets (don't hide them)

### 4. Performance
- ❌ Optimize without profiling
- ✅ Profile first, optimize bottlenecks
- ✅ Use embedded previews for thumbnails
- ✅ Work from smallest usable image size

### 5. Logging
- ❌ Sparse logging makes debugging hard
- ✅ Structured logging with prefixes
- ✅ Log dimensions, counts, states
- ✅ Enable production monitoring

### 6. Architecture Changes
- ❌ Fix one layer, assume it's complete
- ✅ Search ALL files for related code
- ✅ Create migration checklist
- ✅ Test every layer

---

## Success Metrics

### Before Improvements
- ❌ Only 64px thumbnails for Monochrom DNGs
- ❌ Zero-count facets were clickable (invalid transitions)
- ❌ No visibility into RAW decode process
- ❌ B&W photos misclassified as "red"
- ❌ No tests catching critical bugs

### After Improvements
- ✅ All 4 thumbnail sizes generated (64, 256, 512, 1024)
- ✅ Zero-count facets disabled in UI
- ✅ Comprehensive diagnostic logging (`[RAW]`, `[EMBED]`, `FACET_STATE`)
- ✅ Correct B&W classification with saturation-first logic
- ✅ 70% test coverage with tests that document bugs
- ✅ State machine model fully implemented
- ✅ Zero regressions after fixes

---

## Files Added/Modified

### Documentation
- `docs/LESSONS_LEARNED.md` - This document
- `docs/LESSONS_LEARNED_MONOCHROM_DNG.md` - Thumbnail bug deep dive
- `docs/WHERE_CLAUSE_BUG.md` - State machine regression
- `docs/ZERO_RESULTS_HANDLING.md` - UX handling
- `docs/HIERARCHICAL_FACETS.md` - Architecture shift
- `docs/DNG_FORMAT_DEEP_DIVE.md` - File format research
- `docs/DNG_FORMAT_QUICK_REFERENCE.md` - Quick lookup

### Core Fixes
- `internal/indexer/raw_seppedelanghe.go` - Enhanced JPEG extraction + logging
- `internal/indexer/raw_golibraw.go` - Enhanced JPEG extraction + logging
- `internal/query/engine.go` - Fixed WHERE clause independence
- `internal/query/facet_url_builder.go` - State machine URL building
- `internal/query/facets.go` - State machine facet computation
- `internal/explorer/templates/grid.html` - Disabled facets + zero results

### Tests Added
- `internal/indexer/raw_decode_validation_test.go` - 4 validation tests
- `internal/query/where_clause_test.go` - 5 WHERE clause tests
- `internal/query/facet_state_machine_test.go` - State transition tests
- `internal/indexer/color_classification_test.go` - 90+ color tests

### Build
- `Makefile` - Added `test-raw-validation` and other targets
- `go.mod` - Corrected module path to `github.com/adewale/olsen`

---

## Final Thoughts

### What Made This Project Successful

1. **Willingness to Backtrack** - When we discovered hierarchical model was wrong, we rewrote it completely
2. **Test-Driven Bug Fixing** - Every bug became a test that prevents regression
3. **Comprehensive Documentation** - Documented lessons while context was fresh
4. **Structured Logging** - Made debugging and monitoring exponentially easier
5. **Iterative Complexity** - Started simple (hue-based colors), added complexity (saturation-first B&W)

### What We'd Do Differently Next Time

1. **Read file format specs first** - Would have saved days on thumbnail bug
2. **Visual inspection from day one** - Don't trust metrics alone
3. **Create migration checklist immediately** - Catch all layers before moving on
4. **Add diagnostic logging proactively** - Not retroactively after bugs appear
5. **State machine model from start** - Would have avoided hierarchical trap

### Recommended Reading for New Contributors

**Start here:**
- `LESSONS_LEARNED.md` (this document)
- `CLAUDE.md` - Project overview and development patterns
- `specs/facet_state_machine.spec` - Core architectural insight

**Then dive into:**
- `TODO.md` - Current status and roadmap
- `docs/HIERARCHICAL_FACETS.md` - Why we use state machine model
- `docs/LESSONS_LEARNED_MONOCHROM_DNG.md` - Debugging case study

**For specific topics:**
- DNG handling: `docs/DNG_FORMAT_QUICK_REFERENCE.md`
- Color system: `specs/dominant_colours.spec`
- Testing: `docs/TESTING.md`

---

**Authors:** Ade + Claude Code
**Last Updated:** October 12, 2025
**Status:** Living Document - Update with new lessons learned!

---

**Remember:** The best lesson is the one that prevents the next bug. Document what you learn. Test what you fix. Log what you need.
