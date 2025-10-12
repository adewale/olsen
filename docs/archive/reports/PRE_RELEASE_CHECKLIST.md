# Pre-Release Checklist for v0.1.0

**Date:** October 12, 2025
**Target:** GitHub public release
**Status:** ❌ NOT READY - Critical blockers identified

---

## Critical Blockers (MUST FIX)

### 1. ❌ Build is Broken
**Issue:** `make build` fails with undefined functions

**Error:**
```
cmd/olsen/main.go:75:13: undefined: benchmarkThumbnailsCommand
cmd/olsen/main.go:79:13: undefined: benchmarkLibrawCommand
cmd/olsen/main.go:83:13: undefined: versionCommand
```

**Root Cause:** These functions are in files with build tags:
- `benchmark_thumbnails.go` - Has `//go:build` tag
- `benchmark_libraw.go` - Has `//go:build` tag
- `version.go` - May have build constraints

**Fix Required:**
- Option A: Remove benchmark commands from default build (comment out in main.go)
- Option B: Add build tags to include these files in default build
- Option C: Create stub implementations when build tags exclude the real ones

**Recommended:** Option A - These are developer tools, not needed for end users

**Action:**
```go
// In cmd/olsen/main.go, comment out or conditionally compile:
// case "benchmark-thumbnails":
// case "benchmark-libraw":
```

---

### 2. ❌ LICENSE File Missing
**Issue:** No LICENSE file in repository

**Why Critical:**
- GitHub requires license for public repos
- Users need to know usage terms
- Contributors need to know their rights

**Options:**
1. **MIT License** (Most permissive, simple)
   - ✅ Allows commercial use
   - ✅ Allows modification
   - ✅ Simple attribution requirement
   - ✅ Most popular for Go projects

2. **Apache 2.0** (Permissive with patent grant)
   - ✅ Explicit patent grant
   - ✅ Allows commercial use
   - ⚠️ Slightly more complex

3. **GPL v3** (Copyleft)
   - ❌ Requires derivatives to be GPL
   - ❌ Less suitable for library code

**Recommendation:** **MIT License** (simple, widely adopted, Go community standard)

**Action:** Create `LICENSE` file with MIT license text

---

### 3. ⚠️ CONTRIBUTING.md Missing (Recommended)
**Why Important:**
- Tells contributors how to help
- Sets expectations for PRs
- Reduces maintainer burden

**Should Include:**
- Development setup instructions
- How to run tests
- Code style guidelines
- PR submission process
- Issue reporting guidelines

**Action:** Create `CONTRIBUTING.md` based on Go community standards

---

### 4. ⚠️ CHANGELOG.md Missing (Recommended)
**Why Important:**
- Documents changes between versions
- Required for good release process
- Helps users understand what changed

**Format:** Keep a Changelog (keepachangelog.com)

**Action:** Create `CHANGELOG.md` with v0.1.0 initial release notes

---

## Secondary Issues (Should Fix)

### 5. ⚠️ Test Fixtures Copyright
**Issue:** Unclear if testdata/dng files are redistributable

**Files:**
- `testdata/dng/` - DNG test files
- `testdata/photos/` - Photo test files
- `testdata/color_test/` - Color test images

**Questions:**
- Are these your original photos?
- Are they from other sources?
- Can they be legally distributed?

**Options:**
1. If your photos: Add CC0 (public domain) declaration
2. If from others: Remove from repo, document how to obtain
3. If unsure: Remove from repo, use generated test data only

**Action:** Review test fixture sources, add testdata/README.md documenting licensing

---

### 6. ⚠️ Git Status - Many Untracked Files
**Issue:** 40+ untracked files need to be added or ignored

**Categories:**
- Core documentation (README, TODO, CLAUDE.md) - **COMMIT**
- New simplification docs - **COMMIT**
- .gitignore - **COMMIT**
- Makefile - **COMMIT**
- All docs/ markdown files - **REVIEW & COMMIT**

**Action:** Stage and commit appropriate files

---

### 7. ⚠️ README Lacks Quick Start
**Current README:** Good technical detail

**Missing:**
- Installation instructions (how to build)
- Quick start (indexing first photos)
- Screenshot or demo
- Clear "what is this?" at top
- Badge (build status, license)

**Action:** Enhance README with quick start section

---

## Build & Test Verification

### Build Tests Needed
- [ ] `make build` succeeds (currently fails)
- [ ] `make build-raw` succeeds (with LibRaw)
- [ ] `make test` passes (all tests)
- [ ] Binary runs: `./bin/olsen --help`
- [ ] Can index test photos: `./bin/olsen index testdata/dng`
- [ ] Can start explorer: `./bin/olsen explore --db test.db`

### Clean Checkout Test
- [ ] Clone to new directory
- [ ] `make build` works
- [ ] `make test` works
- [ ] All docs render correctly on GitHub

---

## GitHub Repository Setup

### Before Creating Public Repo
- [ ] Remove any sensitive data from git history
- [ ] Verify no API keys, passwords, or personal data
- [ ] Check all committed files are intentional
- [ ] Review .gitignore is comprehensive

### Repository Settings
- [ ] Repository name: `olsen`
- [ ] Description: "Portable photo corpus explorer with faceted search and advanced metadata extraction"
- [ ] Topics: `photography`, `photo-management`, `dng`, `raw`, `metadata`, `exif`, `faceted-search`, `sqlite`, `golang`
- [ ] Enable Issues
- [ ] Enable Discussions (optional)
- [ ] Set default branch to `main`

### Repository Files
- [ ] LICENSE (MIT)
- [ ] README.md (enhanced with quick start)
- [ ] CONTRIBUTING.md
- [ ] CHANGELOG.md
- [ ] .gitignore (already exists)

---

## Pre-Release Action Plan

### Step 1: Fix Build (30 minutes)
```bash
# Option: Comment out benchmark commands in main.go
# Or add build tags to include them
```

### Step 2: Add LICENSE (5 minutes)
```bash
# Create LICENSE file with MIT license
# Update README.md to reference license
```

### Step 3: Add CONTRIBUTING.md (20 minutes)
```markdown
# Development setup
# Running tests
# Submitting PRs
# Code style
```

### Step 4: Add CHANGELOG.md (15 minutes)
```markdown
## [0.1.0] - 2025-10-12
### Added
- Core indexing engine
- Web explorer with faceted search
- State machine navigation model
- (list all major features)
```

### Step 5: Review Test Fixtures (15 minutes)
```bash
# Document test fixture sources
# Add testdata/README.md
# Remove or relicense as needed
```

### Step 6: Enhance README.md (30 minutes)
```markdown
# Add Quick Start section
# Add Installation section
# Add screenshot/demo
# Add license badge
```

### Step 7: Clean Git Status (20 minutes)
```bash
# Review all untracked files
# Stage and commit appropriate files
# Update .gitignore for excluded files
```

### Step 8: Verification (30 minutes)
```bash
# Full clean build
# All tests pass
# Manual smoke test (index photos, explore)
# Check all docs render on GitHub
```

**Total Estimated Time:** ~3 hours

---

## Release Process (After Fixes)

### 1. Final Commit
```bash
git add LICENSE CONTRIBUTING.md CHANGELOG.md
git add README.md  # If enhanced
git add docs/ specs/ # All documentation
git add cmd/ internal/ pkg/ testdata/
git add .gitignore go.mod go.sum Makefile
git commit -m "Prepare for v0.1.0 release

- Add LICENSE (MIT)
- Add CONTRIBUTING.md
- Add CHANGELOG.md
- Enhanced README with quick start
- Documentation cleanup (archived 26 historical docs)
- All tests passing
"
```

### 2. Create Tag
```bash
git tag -a v0.1.0 -m "Release v0.1.0 - Initial public release

Major Features:
- Photo indexing for DNG/JPEG/BMP
- EXIF metadata extraction
- 4-size thumbnail generation
- Color palette analysis (11 Berlin-Kay colors)
- Perceptual hash for duplicates
- Web explorer with faceted search
- State machine navigation model
- Read-only guarantee (never modifies photos)

Documentation:
- Comprehensive specs and requirements
- Lessons learned from development
- Architectural decision records
- Testing guide

Performance:
- ~15-25 photos/second on M3 Max
- Concurrent processing
- SQLite portable catalog
"
```

### 3. Push to GitHub
```bash
# Set official repository remote
git remote add origin https://github.com/adewale/olsen.git
# Or with SSH:
# git remote add origin git@github.com:adewale/olsen.git

git push -u origin main
git push origin v0.1.0
```

### 4. Create GitHub Release
- Go to: https://github.com/adewale/olsen/releases/new
- Select tag: v0.1.0
- Release title: "v0.1.0 - Initial Public Release"
- Description: Copy from CHANGELOG.md
- Attach binaries (optional for 0.1.0)
- Publish release

**Repository URL:** https://github.com/adewale/olsen

---

## Post-Release Tasks

### Immediate
- [ ] Create GitHub Issues for known TODO items
- [ ] Set up GitHub Actions CI/CD
- [ ] Add build status badge to README
- [ ] Announce on social media / Hacker News (optional)

### Week 1
- [ ] Monitor issues
- [ ] Respond to questions
- [ ] Fix critical bugs if reported
- [ ] Start planning v0.2.0

### Month 1
- [ ] Review usage patterns
- [ ] Prioritize feature requests
- [ ] Improve documentation based on feedback

---

## Known Limitations to Document

### For v0.1.0 Release Notes:
1. **LibRaw support** - Optional, requires CGO and libraw
2. **Test fixtures** - May need to be generated by users
3. **No GUI** - CLI + web UI only
4. **Single database** - No distributed/server mode
5. **Read-only** - Cannot write EXIF back to files (by design)
6. **Limited RAW formats** - DNG primarily, other RAW formats via LibRaw

---

## Success Criteria for v0.1.0

### Must Have
- [x] Core indexing works (DONE)
- [ ] Build succeeds without errors
- [ ] All tests pass
- [ ] LICENSE file present
- [ ] README with quick start
- [x] Documentation organized (DONE - archived 26 files)

### Should Have
- [ ] CONTRIBUTING.md
- [ ] CHANGELOG.md
- [x] Comprehensive specs (DONE)
- [x] Test coverage >60% (DONE ~70%)

### Nice to Have
- [ ] CI/CD workflows
- [ ] Pre-built binaries
- [ ] Demo deployment
- [ ] Video tutorial

---

## Current Blockers Summary

**CRITICAL (Blocking release):**
1. Build failure - benchmark commands
2. LICENSE file missing

**HIGH PRIORITY (Should fix before release):**
3. CONTRIBUTING.md missing
4. CHANGELOG.md missing
5. README lacks quick start
6. Test fixture licensing unclear

**MEDIUM PRIORITY (Can defer):**
7. CI/CD setup
8. Pre-built binaries
9. Demo site

---

**Recommendation:** Fix critical blockers (1-2), add missing docs (3-5), then release.
**Estimated time to release-ready:** 3-4 hours of focused work.
