# Simplification & Subtraction Opportunities for Olsen

**Date:** October 12, 2025
**Purpose:** Identify what can be removed, consolidated, or simplified
**Philosophy:** "Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away." — Antoine de Saint-Exupéry

---

## Executive Summary

**Current State:**
- 85 Go files
- 55 markdown documentation files (624KB)
- 6 spec files
- ~9,000 lines of test code
- 21 performance stat JSON files (~12MB of old benchmarks)
- `internal/quality/` package (10 files) with only 21 usages

**Opportunities Identified:**
1. **Remove unused `internal/quality/` package** → Save 10 files
2. **Consolidate 36 doc files into ~15** → Reduce by 60%
3. **Delete 21 old perf stat JSON files** → Save 12MB
4. **Merge 13 similar test files** → Reduce by 50%
5. **Simplify state machine implementation** → Remove logging overhead
6. **Archive completed planning docs** → Clean up active workspace

**Potential Impact:**
- **20-30% fewer files** to maintain
- **60% less documentation** to search through
- **12MB disk space** recovered
- **Faster searches** (fewer files to grep)
- **Clearer mental model** (focused, not sprawling)

---

## Category 1: Unused/Orphaned Code (HIGH PRIORITY)

### 1.1 `internal/quality/` Package — CANDIDATE FOR REMOVAL

**Status:** Nearly unused (only 21 references, all from indexer)

**Files (10 total):**
```
internal/quality/compare.go              (6719 bytes)
internal/quality/diagnostics.go          (5058 bytes)
internal/quality/logging.go              (5032 bytes)
internal/quality/metrics.go              (8752 bytes)
internal/quality/orientation.go          (3926 bytes)
internal/quality/pipeline.go             (8351 bytes)
internal/quality/raw_diag_golibraw.go    (1006 bytes)
internal/quality/raw_diag_seppedelanghe.go (2063 bytes)
internal/quality/report.go               (11186 bytes)
internal/quality/raw_diag.go.bak         (2014 bytes) ← ORPHANED BACKUP
```

**Usage Analysis:**
```bash
$ grep -r "quality\." cmd/ internal/explorer/ internal/indexer/ | wc -l
21
```

**All 21 usages are in `internal/indexer/indexer.go`** for thumbnail quality pipeline.

**Recommendation:**
- **Option A (Aggressive):** Move quality pipeline logic directly into `internal/indexer/thumbnail.go` (~200 lines), delete entire `internal/quality/` package
- **Option B (Conservative):** Keep only `pipeline.go`, delete the other 9 files (diagnostics, logging, comparison, reporting are over-engineered for current needs)
- **Option C (Minimal):** Delete `raw_diag.go.bak` backup file immediately

**Impact:** Save 10 files (52KB), simplify import graph

---

### 1.2 Performance Stat JSON Files — DELETE

**Location:** Root directory
**Count:** 21 files
**Total Size:** ~12MB

```
perfstats_20251006_131024.json    7.0K
perfstats_20251006_131939.json    612K
perfstats_20251007_012518.json    611K
... (18 more files)
```

**Recommendation:** **DELETE ALL** or move to `perftools/archive/` if historical data matters

**Why:**
- These are snapshots from Oct 6-12 benchmarking sessions
- No code references them
- Can regenerate if needed
- 12MB is significant for a repo

**Action:**
```bash
mkdir -p perftools/archive
mv perfstats_*.json perftools/archive/
# Or just delete:
rm perfstats_*.json
```

---

### 1.3 Unused Shell Scripts — EVALUATE

**Files:**
- `test_filter_removal.sh` (2.7K) - Used during state machine debugging
- `explorer.sh` (1.8K) - Wrapper for `olsen explore`
- `indexphotos.sh` (2.6K) - Wrapper for `olsen index`

**Recommendation:**
- **Keep:** `explorer.sh` (useful convenience wrapper)
- **Archive or Delete:** `test_filter_removal.sh` (debugging artifact, tests cover this now)
- **Evaluate:** `indexphotos.sh` (does it add value over `olsen index`?)

---

## Category 2: Documentation Consolidation (HIGH PRIORITY)

### 2.1 Current Documentation Sprawl

**55 markdown files, 624KB total**

**Largest files:**
```
41K  UI_REDESIGN_PLAN.md
31K  flow.md
27K  LESSONS_LEARNED.md
27K  DNG_FORMAT_DEEP_DIVE.md
24K  QUERIES.md
23K  STATE_MACHINE_MIGRATION.md
19K  FACETED_NAVIGATION_PLAN.md
19K  architecture.md
18K  EXPLORER_PLAN.md
```

**Problem:** Too many overlapping planning/research docs make it hard to find current information.

---

### 2.2 Consolidation Opportunities

#### Group 1: Completed Migration/Planning Docs → ARCHIVE

**These docs served their purpose during development:**
```
STATE_MACHINE_MIGRATION.md        23K  ✅ Migration complete
FACETED_NAVIGATION_PLAN.md        19K  ✅ Implemented
EXPLORER_PLAN.md                  18K  ✅ Built
UI_REDESIGN_PLAN.md               41K  ⚠️  Partially done
THUMBNAIL_QUALITY_IMPLEMENTATION_STATUS.md  14K  ✅ Complete
THUMBNAIL_FIDELITY_FIX.md         15K  ✅ Fixed
TEST_COVERAGE_PLAN.md             9.3K ⚠️  Ongoing
```

**Recommendation:** Move to `docs/archive/completed/`

**Why:** These are historical artifacts. Lessons learned captured in `LESSONS_LEARNED.md`. Current state documented elsewhere.

---

#### Group 2: LibRaw Research Docs → CONSOLIDATE

**Multiple docs about same topic:**
```
LIBRAW_IMPLEMENTATION_SUMMARY.md
LIBRAW_API_INVESTIGATION.md
LIBRAW_FORK_PLAN.md
LIBRAW_BUFFER_OVERFLOW_RESEARCH.md
LIBRAW_DUAL_LIBRARY_SUPPORT.md
LIBRAW_FIX_COMPLETE.md
```

**Recommendation:** Consolidate into single `LIBRAW_GUIDE.md` with sections:
- Quick Start (which build target to use)
- Known Issues (monochrome DNGs, buffer overflow)
- Dual Library Support (why and how)
- Historical Research (archived sections for deep dives)

**Impact:** 6 files → 1 file

---

#### Group 3: Thumbnail Research Docs → CONSOLIDATE

**Multiple docs on thumbnails:**
```
THUMBNAIL_QUALITY_RESEARCH.md
thumbnail_quality_research_results.md
THUMBNAIL_VALIDATION_FINDINGS.md
THUMBNAIL_QUALITY_IMPLEMENTATION_STATUS.md
THUMBNAIL_FIDELITY_FIX.md
```

**Recommendation:** Merge into `THUMBNAIL_IMPLEMENTATION.md` with:
- Current implementation (what's in master)
- Known issues and solutions
- Research notes (archived section)

**Impact:** 5 files → 1 file

---

#### Group 4: Faceted Navigation Docs → CONSOLIDATE

**Multiple overlapping docs:**
```
FACETED_NAVIGATION_PLAN.md        19K
HIERARCHICAL_FACETS.md
HIERARCHICAL_AUDIT.md
WHERE_CLAUSE_BUG.md
ZERO_RESULTS_HANDLING.md
FIX_VALIDATION_CHECKLIST.md
```

**Current Status:** All implemented, bugs fixed

**Recommendation:**
- Keep: `HIERARCHICAL_FACETS.md` (explains core architecture decision)
- Merge others into `FACETED_NAVIGATION_POSTMORTEM.md` (shorter, focused on outcomes)
- OR: Move all to `docs/archive/faceted-navigation/`

**Impact:** 6 files → 1-2 files

---

#### Group 5: Datasette Inspiration Docs → QUESTION NECESSITY

```
DATASETTE_LESSONS.md              11K
WHAT_DATASETTE_COULD_LEARN.md     16K
```

**Recommendation:**
- **Option A:** Merge into single `DATASETTE_INSPIRATION.md` (brief, focused)
- **Option B:** Delete if lessons already integrated into design

**Question:** Do these docs inform future decisions, or were they one-time research?

---

### 2.3 Proposed Documentation Structure

**After consolidation (15-20 files instead of 55):**

```
docs/
  # USER DOCS (what users need)
  README.md                       ← Main entry point
  INSTALLATION.md                 ← How to build/install
  GETTING_STARTED.md              ← First steps
  CLI_REFERENCE.md                ← Command documentation

  # DEVELOPER DOCS (what contributors need)
  ARCHITECTURE.md                 ← System design (keep current)
  LESSONS_LEARNED.md              ← Unified lessons (already done!)
  CLAUDE.md                       ← AI assistant guidance (keep as-is)
  TESTING.md                      ← How to test
  CONTRIBUTING.md                 ← How to contribute

  # TECHNICAL SPECS (implementation details)
  DNG_FORMAT_GUIDE.md             ← Consolidate DNG_FORMAT_* files
  LIBRAW_GUIDE.md                 ← Consolidate 6 LibRaw docs
  THUMBNAIL_IMPLEMENTATION.md     ← Consolidate 5 thumbnail docs
  FACETED_NAVIGATION.md           ← Keep core explanation
  COLOR_SYSTEM.md                 ← Dominant colors (from spec)

  # RESEARCH & PLANNING (optional, can archive)
  archive/
    completed/                    ← Migration plans, old research
    historical/                   ← Initial planning docs
```

**Impact:** 55 files → 15 core files + archived folder

---

## Category 3: Test File Consolidation (MEDIUM PRIORITY)

### 3.1 Current Test Organization

**Query package: 13 test files, 4610 total lines**

**Opportunity:** Merge similar test files

```
facet_hierarchy_test.go              144 lines  ← Old hierarchical model tests
facet_state_transitions_test.go      378 lines  ← State machine tests
facet_state_machine_test.go          584 lines  ← More state machine tests
facet_lifecycle_test.go              527 lines  ← Lifecycle tests
```

**Recommendation:** Merge these 4 into `facet_behavior_test.go` (comprehensive state machine tests)

```
facet_count_validation_test.go       364 lines  ← Count validation
facet_counts_simple_test.go          293 lines  ← Simple count tests
```

**Recommendation:** Merge these 2 into `facet_counts_test.go`

```
color_integration_test.go            228 lines  ← Color integration
color_classification_test.go         460 lines  ← Color classification
```

**Recommendation:** Keep separate (different purposes)

```
facet_url_bug_test.go               148 lines  ← Bug regression test
where_clause_test.go                140 lines  ← WHERE clause tests
```

**Recommendation:** These are fine, document specific bugs/features

**Impact:** 13 test files → 9 test files (still comprehensive, less fragmentation)

---

### 3.2 Indexer Test Consolidation

**Indexer package: 19 test files, 4527 total lines**

**Opportunity:** Merge RAW-related tests

```
raw_decode_validation_test.go        291 lines  ← Validation
raw_brightness_test.go               ~100 lines ← Brightness
raw_brightness_golibraw_test.go      ~100 lines ← Brightness (other lib)
raw_comparison_test.go               ~150 lines ← Library comparison
raw_buffer_overflow_test.go          181 lines  ← Buffer overflow
raw_buffer_overflow_golibraw_test.go ~100 lines ← Buffer overflow (other lib)
```

**Recommendation:** Consolidate into:
- `raw_validation_test.go` (validation + brightness + quality)
- `raw_library_comparison_test.go` (both libraries, buffer overflow)

**Impact:** 6 test files → 2 test files

---

## Category 4: Code Simplification (MEDIUM PRIORITY)

### 4.1 Facet Logging Overhead

**File:** `internal/query/facet_logger.go`

**Purpose:** Structured logging for facet state (Phase 2d of state machine migration)

**Current Usage:** Every page render logs facet state

**Question:** Is this still needed, or was it debugging aid during migration?

**Recommendation:**
- **Option A:** Remove entirely if migration is stable
- **Option B:** Add `--debug` flag, only log when enabled
- **Option C:** Keep but reduce verbosity (only log on zero results)

**Impact:** Reduce log noise, slight performance improvement

---

### 4.2 Duplicate DNG Format Docs

**Files:**
```
docs/DNG_FORMAT_DEEP_DIVE.md       27K  ← Comprehensive research
docs/DNG_FORMAT_QUICK_REFERENCE.md ~10K ← Quick lookup
```

**Recommendation:** Keep both (serve different purposes), but cross-reference clearly

---

### 4.3 Build Tag Complexity

**Current:** 3 build configurations for LibRaw
- No CGO (fallback)
- CGO + golibraw
- CGO + seppedelanghe

**Question:** Is this complexity worth it? Do we actually switch between libraries?

**Recommendation:**
- If seppedelanghe works well, consider making it the ONLY CGO option
- Keep no-CGO fallback for pure Go builds
- **Impact:** Remove 2 build tag combinations, simpler builds

---

## Category 5: Database/Schema Simplification (LOW PRIORITY)

### 5.1 Unused Tables?

**Current schema includes:**
- `tags` table (not implemented yet)
- `collections` table (not implemented yet)
- `burst_groups` table (implemented but not exposed in UI)

**Recommendation:** Keep for now (planned features), but consider in future if not used

---

### 5.2 Query Complexity

**Current:** Facet computation runs multiple SQL queries per page render

**Future Optimization Opportunity:**
- Materialize facet counts in dedicated table
- Update on write, not compute on read
- **Trade-off:** Faster reads, slower writes (probably worth it)

---

## Category 6: Specification Files (LOW PRIORITY)

**Current: 6 spec files**

```
specs/flags.spec
specs/performance.spec
specs/perftools.spec
specs/faceted_navigation.spec
specs/dominant_colours.spec
specs/facet_state_machine.spec
```

**Analysis:** These are reference specs, not code. Keep all.

**Recommendation:** Add `specs/README.md` that explains what each spec covers and links to implementation.

---

## Action Plan

### Phase 1: Quick Wins (1-2 hours)

1. **Delete perf stat JSON files** → Save 12MB
   ```bash
   mkdir -p perftools/archive
   mv perfstats_*.json perftools/archive/
   ```

2. **Delete backup file**
   ```bash
   rm internal/quality/raw_diag.go.bak
   ```

3. **Delete or archive test script**
   ```bash
   mv test_filter_removal.sh docs/archive/scripts/
   ```

4. **Create archive directories**
   ```bash
   mkdir -p docs/archive/{completed,historical,research}
   ```

**Impact:** Immediate cleanup, 12MB saved

---

### Phase 2: Documentation Consolidation (4-6 hours)

1. **Archive completed planning docs**
   ```bash
   mv docs/STATE_MACHINE_MIGRATION.md docs/archive/completed/
   mv docs/FACETED_NAVIGATION_PLAN.md docs/archive/completed/
   mv docs/EXPLORER_PLAN.md docs/archive/completed/
   mv docs/THUMBNAIL_FIDELITY_FIX.md docs/archive/completed/
   mv docs/THUMBNAIL_QUALITY_IMPLEMENTATION_STATUS.md docs/archive/completed/
   # ... (10-15 files total)
   ```

2. **Consolidate LibRaw docs**
   - Create `docs/LIBRAW_GUIDE.md` (merge 6 files)
   - Move originals to `docs/archive/research/libraw/`

3. **Consolidate Thumbnail docs**
   - Create `docs/THUMBNAIL_IMPLEMENTATION.md` (merge 5 files)
   - Move originals to `docs/archive/research/thumbnails/`

4. **Consolidate Faceted Navigation docs**
   - Keep `HIERARCHICAL_FACETS.md` (core explanation)
   - Move others to `docs/archive/faceted-navigation/`

5. **Update README with new doc structure**

**Impact:** 55 docs → 15 active docs + archived folder

---

### Phase 3: Code Cleanup (6-8 hours)

1. **Evaluate `internal/quality/` package**
   - Audit all 21 usages in indexer
   - Decision: Inline or simplify?
   - **Recommended:** Move pipeline logic to `internal/indexer/thumbnail.go`, delete `internal/quality/`

2. **Consolidate test files**
   - Merge facet test files (13 → 9)
   - Merge RAW test files (6 → 2)
   - Run full test suite to verify

3. **Simplify logging**
   - Make facet logging conditional (debug flag)
   - Reduce verbosity in production

**Impact:** ~15 fewer files, simpler dependency graph

---

### Phase 4: Future Simplifications (Deferred)

1. **Single LibRaw implementation** (if one proves sufficient)
2. **Materialized facet counts** (performance optimization)
3. **Remove unimplemented features** (tags, collections) if not building them

---

## Subtraction Philosophy

### Questions to Ask for Each File/Feature

1. **Is this used?** (grep, import analysis)
2. **Does this inform current decisions?** (vs historical artifact)
3. **Can this be merged with something else?** (consolidation)
4. **Does this add complexity without proportional value?** (cost/benefit)
5. **Would we re-create this if starting fresh?** (necessity test)

### What NOT to Simplify

**Don't remove:**
- ✅ Core functionality (indexer, query engine, explorer)
- ✅ Working tests (even if numerous, they catch regressions)
- ✅ `LESSONS_LEARNED.md` (unified, valuable)
- ✅ `CLAUDE.md` (working well as AI guidance)
- ✅ Active specs (reference material)
- ✅ `HIERARCHICAL_FACETS.md` (explains key architectural decision)

**Keep complexity when:**
- It prevents bugs (comprehensive tests)
- It enables future features (database schema with unused tables)
- It improves debugging (diagnostic logging, if used)
- It serves different audiences (deep dive vs quick reference docs)

---

## Summary Table

| Category | Current | After Cleanup | Savings |
|----------|---------|---------------|---------|
| **Go Files** | 85 | 75 (-12%) | 10 files |
| **Documentation** | 55 files | 15 active + archive | 73% reduction in active docs |
| **Test Files** | 32 | 25 (-22%) | 7 files |
| **Perf JSON Files** | 21 (12MB) | 0 | 12MB |
| **Total Impact** | 193 files | 115 files + archive | 40% reduction |

---

## Expected Benefits

### Immediate
- ✅ 12MB disk space recovered
- ✅ Faster grep/search (fewer files)
- ✅ Clearer project structure
- ✅ Easier onboarding (find relevant docs faster)

### Medium-Term
- ✅ Simpler mental model (less to remember)
- ✅ Faster builds (fewer files to compile)
- ✅ Reduced maintenance burden
- ✅ Clear separation: active vs historical docs

### Long-Term
- ✅ Sets precedent for continuous simplification
- ✅ Easier to spot actual complexity growth
- ✅ More confidence to delete (knowing we have good history)

---

## Risk Assessment

**Low Risk:**
- Deleting perf JSON files (can regenerate)
- Archiving completed planning docs (git history preserves)
- Deleting backup files (obvious candidates)

**Medium Risk:**
- Consolidating test files (must verify coverage maintained)
- Removing `internal/quality/` package (audit usages first)
- Simplifying logging (might lose debugging capability)

**Mitigation:**
- Run full test suite after each change
- Keep git history (can always revert)
- Move to archive first, delete later if confident
- Create git tags before major cleanup: `git tag before-simplification-2025-10-12`

---

## Next Steps

**Recommended Order:**

1. **Quick wins** (Phase 1) - Do immediately, no risk
2. **Documentation** (Phase 2) - High value, low risk
3. **Code cleanup** (Phase 3) - Test thoroughly, higher risk but high reward
4. **Future optimizations** (Phase 4) - Defer until needed

**Get Approval For:**
- Deleting/archiving specific documentation (does user want to keep anything?)
- Removing `internal/quality/` package (is performance instrumentation needed?)
- Consolidating tests (is current granularity important?)

---

**Philosophy:** Every file, every line of code, every document is a commitment to maintain. Only keep what earns its place.

**Measurement:** Track complexity metrics monthly:
```bash
# File count
find . -name "*.go" -o -name "*.md" | wc -l

# Active documentation
ls docs/*.md | wc -l

# Test coverage
go test -cover ./...
```

**Goal:** Maintain or improve functionality while steadily reducing file count and complexity.
