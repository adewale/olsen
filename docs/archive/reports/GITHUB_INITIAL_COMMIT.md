# Files for Initial GitHub Commit

**Date:** October 12, 2025
**Purpose:** Checklist for initial public release on GitHub
**Repository:** github.com/adewale/olsen

---

## Commit Strategy

**Three-phase approach:**
1. **Essential Core** - Minimum viable public repo
2. **Documentation** - User and developer docs
3. **Extended** - Research, specs, and additional context

---

## Phase 1: Essential Core (MUST COMMIT)

### Root Configuration Files
- ✅ `.gitignore` - Already exists, excludes DB files, binaries
- ✅ `go.mod` - Module definition
- ✅ `go.sum` - Dependency checksums
- ✅ `Makefile` - Build targets
- ✅ `README.md` - **CRITICAL: User-facing documentation**
- ⚠️ `LICENSE` - **MISSING: Need to add (MIT, Apache 2.0, or GPL?)**

### Core Documentation (Root)
- ✅ `CLAUDE.md` - AI assistant guidance (helpful for contributors)
- ✅ `TODO.md` - Project status and roadmap
- ⚠️ `CONTRIBUTING.md` - **MISSING: Guidelines for contributors**
- ⚠️ `CHANGELOG.md` - **MISSING: Version history**

### Source Code
```
cmd/
  olsen/
    ✅ main.go
    ✅ version.go
    ✅ benchmark_libraw.go
    ✅ benchmark_thumbnails.go
    ✅ integration_test.go
    ✅ main_test.go

internal/
  database/
    ✅ database.go
    ✅ database_test.go
    ✅ schema.go

  explorer/
    ✅ server.go
    ✅ repository.go
    ✅ repository_test.go
    ✅ facet_404_test.go
    ✅ facet_disabled_test.go
    ✅ templates/
        ✅ grid.html
        ✅ home.html
        ✅ photo.html

  indexer/
    ✅ indexer.go
    ✅ metadata.go
    ✅ thumbnail.go
    ✅ color.go
    ✅ phash.go
    ✅ inference.go
    ✅ burst.go
    ✅ perfoutput.go
    ✅ raw_seppedelanghe.go
    ✅ raw_golibraw.go
    ✅ raw_nocgo.go
    ✅ ALL *_test.go files (19 test files)

  query/
    ✅ engine.go
    ✅ facets.go
    ✅ facet_url_builder.go
    ✅ facet_logger.go
    ✅ url_mapper.go
    ✅ types.go
    ✅ utils.go
    ✅ ALL *_test.go files (13 test files)

  quality/
    ⚠️  EVALUATE: Do we commit this package?
    ✅ pipeline.go (if keeping package)
    ✅ metrics.go
    ✅ diagnostics.go
    ✅ logging.go
    ✅ compare.go
    ✅ orientation.go
    ✅ report.go
    ✅ raw_diag_golibraw.go
    ✅ raw_diag_seppedelanghe.go
    ❌ raw_diag.go.bak (EXCLUDE - backup file)

pkg/
  models/
    ✅ types.go
```

### Test Data & Fixtures
```
testdata/
  ✅ FIXTURES_SUMMARY.md
  ✅ generate_fixtures.go
  ✅ generate_dng_fixtures.go
  ✅ generate_color_images.go
  ✅ create_color_images.go
  ✅ create_color_test_images.py
  ✅ test_fixtures.go
  ✅ verify_coverage.go
  ✅ add_exif_to_test_images.sh
  ⚠️  dng/ - **QUESTION: Are these test DNGs redistributable?**
  ⚠️  photos/ - **QUESTION: Are these redistributable?**
  ⚠️  color_test/ - **QUESTION: Are these redistributable?**
```

### Specification Files
```
specs/
  ✅ olsen_requirements.md
  ✅ olsen_specs.md
  ✅ dominant_colours.spec
  ✅ facet_state_machine.spec
  ✅ faceted_navigation.spec
  ✅ faceted_ui_implementation.md
  ✅ facets_spec.md
  ✅ flags.spec
  ✅ performance.spec
  ✅ perftools.spec
  ✅ olsen_faceted_ui_mock.png
```

### Performance Tools
```
perftools/
  ✅ README.md
  ✅ analyze_filetype.py
  ✅ compare_datasets.py
  ❌ archive/ - EXCLUDE (old perf JSON files)
```

### Utility Scripts
```
Root scripts:
  ✅ explorer.sh - Convenience wrapper
  ✅ indexphotos.sh - Convenience wrapper
  ✅ create_test_db.go - Test database creation
  ❌ test_filter_removal.sh - EXCLUDE (debugging artifact)
```

---

## Phase 2: Documentation (COMMIT SELECTIVELY)

### Essential User/Developer Docs
```
docs/
  ✅ LESSONS_LEARNED.md - Unified lessons (NEW, comprehensive)
  ✅ TESTING.md - Testing guide
  ✅ SIMPLIFICATION_OPPORTUNITIES.md - Maintenance guide (NEW)

  # Architecture & Design
  ⚠️  architecture.md - KEEP but may need update
  ⚠️  flow.md - KEEP if referenced, otherwise archive

  # DNG/RAW Implementation
  ⚠️  DNG_FORMAT_DEEP_DIVE.md - KEEP (valuable reference)
  ⚠️  DNG_FORMAT_QUICK_REFERENCE.md - KEEP (quick lookup)
  ⚠️  LESSONS_LEARNED_MONOCHROM_DNG.md - KEEP (specific case study)

  # State Machine & Faceted Navigation
  ⚠️  HIERARCHICAL_FACETS.md - KEEP (explains key decision)
  ❌ STATE_MACHINE_MIGRATION.md - ARCHIVE (historical)
  ❌ FACETED_NAVIGATION_PLAN.md - ARCHIVE (historical)
  ❌ HIERARCHICAL_AUDIT.md - ARCHIVE (historical)
  ❌ WHERE_CLAUSE_BUG.md - ARCHIVE (in LESSONS_LEARNED.md now)
  ❌ ZERO_RESULTS_HANDLING.md - ARCHIVE (in LESSONS_LEARNED.md now)
  ❌ FIX_VALIDATION_CHECKLIST.md - ARCHIVE (historical)
  ❌ QUERY_STRING_NAVIGATION.md - ARCHIVE (implemented)

  # LibRaw Documentation
  ⚠️  Consolidate into one file before commit
  ❌ LIBRAW_API_INVESTIGATION.md - ARCHIVE/CONSOLIDATE
  ❌ LIBRAW_BUFFER_OVERFLOW_RESEARCH.md - ARCHIVE/CONSOLIDATE
  ✅ LIBRAW_DUAL_LIBRARY_SUPPORT.md - KEEP (explains current design)
  ❌ LIBRAW_FIX_COMPLETE.md - ARCHIVE
  ❌ LIBRAW_FORK_PLAN.md - ARCHIVE
  ❌ LIBRAW_IMPLEMENTATION_SUMMARY.md - ARCHIVE/CONSOLIDATE

  # Thumbnail Documentation
  ⚠️  Consolidate into one file before commit
  ❌ THUMBNAIL_QUALITY_RESEARCH.md - ARCHIVE/CONSOLIDATE
  ❌ thumbnail_quality_research_results.md - ARCHIVE/CONSOLIDATE
  ❌ THUMBNAIL_VALIDATION_FINDINGS.md - ARCHIVE/CONSOLIDATE
  ❌ THUMBNAIL_QUALITY_IMPLEMENTATION_STATUS.md - ARCHIVE
  ❌ THUMBNAIL_FIDELITY_FIX.md - ARCHIVE

  # Explorer/UI Documentation
  ❌ EXPLORER_PLAN.md - ARCHIVE (implemented)
  ❌ EXPLORER_SPEC.md - ARCHIVE (implemented)
  ❌ UI_REDESIGN_PLAN.md - ARCHIVE (work in progress doc)

  # Query/Database
  ⚠️  QUERIES.md - EVALUATE: Still relevant?

  # Testing
  ❌ MISSING_INTEGRATION_TESTS.md - ARCHIVE (now have tests)
  ❌ TEST_COVERAGE_PLAN.md - ARCHIVE (ongoing work)

  # Inspiration/Research
  ❌ DATASETTE_LESSONS.md - ARCHIVE (research artifact)
  ❌ WHAT_DATASETTE_COULD_LEARN.md - ARCHIVE (research artifact)

  # Compliance/Flags
  ❌ flags-compliance-analysis.md - ARCHIVE (internal)
  ❌ flags-compliance-report.md - ARCHIVE (internal)
```

---

## Phase 3: Root Supplementary Files

### Files Currently in Root (Evaluate)
```
✅ CLAUDE.md - KEEP (useful for AI-assisted development)
✅ README.md - KEEP (essential)
✅ TODO.md - KEEP (roadmap)
⚠️  EXIF_LIBRARY_MIGRATION.md - ARCHIVE? (Historical migration notes)
⚠️  IMPLEMENTATION_COMPARISON.md - ARCHIVE? (Internal comparison)
⚠️  TEST_COVERAGE_REPORT.md - ARCHIVE? (Point-in-time report)
```

---

## Files to EXCLUDE from Git

### Build Artifacts
```
❌ bin/ - Binary outputs
❌ *.test - Test binaries
❌ *.exe, *.dll, *.so, *.dylib - Platform binaries
```

### Database Files
```
❌ *.db - Database files
❌ *.db-shm - SQLite shared memory
❌ *.db-wal - SQLite write-ahead log
❌ experimental_photos.db
❌ lightroom.db
❌ other_photos.db
❌ photos.db
❌ perf.db
❌ test*.db
```

### Benchmark/Performance Data
```
❌ perfstats_*.json - Old performance data (21 files, 12MB)
❌ benchmark_baseline.html
❌ benchmark_improved.html
❌ libraw_benchmark.html
```

### IDE/OS Files (already in .gitignore)
```
❌ .DS_Store
❌ .vscode/
❌ .idea/
```

### Private Test Data
```
❌ private-testdata/ - Not in repo, personal photos
```

### Backup/Temp Files
```
❌ internal/quality/raw_diag.go.bak
❌ test_filter_removal.sh (debugging script)
```

---

## Critical Missing Files

### 1. LICENSE (REQUIRED)
**Options:**
- MIT License (permissive, simple)
- Apache 2.0 (permissive, patent grant)
- GPL v3 (copyleft)
- BSD 3-Clause (permissive)

**Recommendation:** MIT or Apache 2.0 for maximum adoption

### 2. CONTRIBUTING.md (RECOMMENDED)
**Should include:**
- How to set up development environment
- How to run tests
- Code style guidelines
- Pull request process
- Issue reporting guidelines

### 3. CHANGELOG.md (RECOMMENDED)
**Format:** Keep a Changelog (keepachangelog.com)
```
## [Unreleased]
- Initial public release

## [0.1.0] - 2025-10-12
### Added
- Core indexing engine
- Web explorer UI
- Faceted search
- State machine navigation
- ...
```

### 4. Enhanced README.md
**Current README should include:**
- Project description
- Screenshots/demo
- Quick start guide
- Installation instructions
- Basic usage examples
- Links to documentation
- License badge
- Build status badge (once CI/CD set up)

---

## Recommended Commit Order

### Commit 1: Core Functionality
```bash
git add go.mod go.sum Makefile .gitignore
git add cmd/ internal/ pkg/
git add testdata/*.go testdata/*.md testdata/*.sh
git add LICENSE  # After creating
git commit -m "Initial commit: Core indexing and query engine"
```

### Commit 2: Specifications & Documentation
```bash
git add specs/
git add README.md CLAUDE.md TODO.md CONTRIBUTING.md CHANGELOG.md
git commit -m "Add specifications and project documentation"
```

### Commit 3: Essential Technical Docs
```bash
git add docs/LESSONS_LEARNED.md
git add docs/HIERARCHICAL_FACETS.md
git add docs/DNG_FORMAT_DEEP_DIVE.md
git add docs/DNG_FORMAT_QUICK_REFERENCE.md
git add docs/LESSONS_LEARNED_MONOCHROM_DNG.md
git add docs/LIBRAW_DUAL_LIBRARY_SUPPORT.md
git add docs/TESTING.md
git commit -m "Add technical documentation and lessons learned"
```

### Commit 4: Utility Scripts & Tools
```bash
git add explorer.sh indexphotos.sh create_test_db.go
git add perftools/
git commit -m "Add utility scripts and performance tools"
```

### Commit 5: Test Fixtures (If Redistributable)
```bash
git add testdata/dng/
git add testdata/photos/
git add testdata/color_test/
git commit -m "Add test fixtures for validation"
```

---

## Pre-Commit Checklist

### Code Quality
- [ ] All tests pass: `make test`
- [ ] Code builds successfully: `make build`
- [ ] No sensitive data in files (API keys, personal photos, etc.)
- [ ] No absolute paths in code (use relative paths)
- [ ] Remove debug logging or make it conditional

### Documentation
- [ ] README.md is comprehensive and accurate
- [ ] LICENSE file added
- [ ] CONTRIBUTING.md added
- [ ] All code examples in docs work
- [ ] Links in documentation are correct

### File Hygiene
- [ ] No backup files (*.bak, *~)
- [ ] No binary files (*.db, *.exe, *.test)
- [ ] No IDE config committed (.vscode/, .idea/)
- [ ] No OS files (.DS_Store, Thumbs.db)
- [ ] .gitignore covers all generated files

### Legal/Copyright
- [ ] LICENSE file present
- [ ] Copyright notices in source files (optional but recommended)
- [ ] Third-party licenses acknowledged (if dependencies have special licenses)
- [ ] Test fixtures are redistributable (check copyright on test DNGs)

---

## Test Fixtures: Copyright Consideration

**CRITICAL QUESTION:** Where did the test DNG files come from?

**If they're your personal photos:** You can include them with a note:
```
testdata/README.md:
"Test fixtures in this directory are provided under CC0 (public domain)
for testing purposes. These are sample photos created by the project author."
```

**If they're from other sources:**
- Cannot redistribute without permission
- Either:
  1. Remove from repo, document how to obtain test files
  2. Use synthetic test data (generated programmatically)
  3. Use public domain images

**Recommendation:** Create `testdata/README.md` explaining fixture sources and licensing.

---

## GitHub Repository Setup

### 1. Repository Description
```
Olsen - A portable photo corpus explorer with faceted search and advanced metadata extraction
```

### 2. Topics/Tags
```
photography, photo-management, dng, raw, metadata, exif, faceted-search,
sqlite, golang, web-app, photo-library, image-processing
```

### 3. Repository Settings
- [ ] Enable Issues
- [ ] Enable Wiki (optional, docs are in-repo)
- [ ] Enable Discussions (optional, for Q&A)
- [ ] Add description
- [ ] Add website URL (if hosted demo exists)
- [ ] Add topics

### 4. GitHub Actions (Future)
```
.github/workflows/
  test.yml - Run tests on push
  build.yml - Build binaries for releases
  lint.yml - Go linting
```

---

## Final File Count Estimate

**Essential Core:**
- ~85 Go files
- ~10 spec files
- ~15 essential docs
- Test fixtures (size TBD based on copyright)
- Utility scripts (3-4 files)

**Total:** ~115 files (vs current 193 if we archived historical docs)

---

## Post-Initial Commit Tasks

### 1. GitHub Actions CI/CD
- Set up automated testing
- Build releases for multiple platforms
- Linting and code quality checks

### 2. Release Process
- Create v0.1.0 tag
- Generate release notes from CHANGELOG.md
- Build binaries for Linux, macOS, Windows
- Create GitHub Release

### 3. Community Setup
- Add issue templates
- Add pull request template
- Create CODE_OF_CONDUCT.md (optional but recommended)
- Set up project board for tracking work

### 4. Documentation Hosting
- Consider GitHub Pages for rendered docs
- Or link to docs/ folder in README

---

## Summary Checklist

**Must Do Before Initial Commit:**
- [ ] Add LICENSE file
- [ ] Add/enhance README.md
- [ ] Add CONTRIBUTING.md
- [ ] Add CHANGELOG.md
- [ ] Verify no sensitive data in any files
- [ ] Consolidate or archive redundant docs
- [ ] Verify test fixtures are redistributable
- [ ] Run full test suite
- [ ] Build successfully on clean checkout

**Recommended:**
- [ ] Add CI/CD workflows
- [ ] Create release notes
- [ ] Add screenshot/demo to README
- [ ] Set up GitHub repo settings
- [ ] Archive historical documentation

**Can Defer:**
- [ ] Wiki setup
- [ ] Discussions forum
- [ ] Project board
- [ ] GitHub Pages docs site
