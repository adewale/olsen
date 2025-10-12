# Olsen TODO

**Last Updated:** October 7, 2025 (updated after query engine and web integration)
**Current Status:** ~80% complete (core CLI functional, query system complete, web UI integrated)

---

## Phase 1: Core Functionality (High Priority)

### 1.1 CLI Interface
**Status:** 70% complete (7/10 core commands implemented)
**Priority:** Critical

- [x] Implement `olsen index <path>` command
  - [x] Add `-w N` flag for worker count
  - [x] Add `--db <path>` flag for database path
  - [x] Add progress bar display
  - [x] Add completion summary
- [ ] Implement `olsen reindex <path>` command
  - [ ] Only process modified files
  - [ ] Update existing records
- [x] Implement `olsen analyze` command
  - [x] Run burst detection
  - [x] Display analysis results
- [x] Implement `olsen stats` command
  - [x] Overall statistics
  - [ ] `--bursts` flag for burst statistics (shows count already)
  - [x] Top cameras/lenses display
  - [x] Date range display
- [x] Implement `olsen show <photo-id>` command
  - [x] Display full metadata
  - [ ] Show burst/collection membership
- [x] Implement `olsen thumbnail <photo-id>` command
  - [x] `-s <size>` flag (64, 256, 512, 1024)
  - [x] `-o <output>` flag for output path
- [x] Implement `olsen verify` command
  - [x] Check database integrity
  - [ ] Verify file hashes (checks file existence)
  - [x] Report missing files
- [ ] Implement `olsen compact` command
  - [ ] Vacuum database
  - [ ] Show size savings

### 1.2 Indexer Improvements
**Status:** 70% complete (core works, needs polish)
**Priority:** Medium

- [ ] Complete EXIF extraction
  - [ ] Extract focus distance field
  - [ ] Extract DNG version
  - [ ] Extract original RAW filename
  - [ ] Verify 95%+ extraction success rate
- [ ] Add incremental indexing
  - [ ] Skip already-indexed files (by hash)
  - [ ] Update modified files only
  - [ ] Remove deleted files from database
- [ ] Improve error reporting
  - [ ] Structured error log
  - [ ] Error categories (EXIF missing, decode failed, etc.)
  - [ ] Export error report
- [ ] Add indexing options
  - [ ] Skip thumbnails flag (faster indexing)
  - [ ] Skip color extraction flag
  - [ ] Custom thumbnail sizes

### 1.3 Burst Detection
**Status:** 90% complete (functional, needs refinement)
**Priority:** Low

- [ ] Add sharpness scoring
  - [ ] Laplacian variance method
  - [ ] Use sharpest photo as representative (not middle)
- [ ] Add CLI commands
  - [ ] `olsen bursts` - list all burst groups
  - [ ] `olsen burst <id>` - show specific burst
- [ ] Improve detection accuracy
  - [ ] Test with various camera models
  - [ ] Tune time window (currently 2s)
  - [ ] Validate 90%+ accuracy

---

## Phase 2: Query and Search (High Priority)

### 2.1 Query Engine Foundation
**Status:** 100% complete ✅
**Priority:** Critical

- [x] Implement QueryParams structure
  - [x] All temporal filters
  - [x] All equipment filters
  - [x] All technical filters (ISO, aperture, focal length ranges)
  - [x] Categorical filters (time_of_day, season, etc.)
  - [x] Location filters (lat/lon ranges)
  - [x] Burst filters
- [x] Implement QueryResult structure
  - [x] PhotoSummary with all fields
  - [x] Pagination metadata
- [x] Build SQL query builder
  - [x] WHERE clause construction
  - [x] Multi-dimensional filtering
  - [x] Proper parameter binding
  - [x] Index utilization

### 2.2 Faceted Search System
**Status:** 100% complete ✅
**Priority:** Critical (core feature of v2.0 spec)

- [x] Implement facet computation
  - [x] Camera/lens facets
  - [x] Time of day facets
  - [x] Season facets
  - [x] Year/month facets
  - [x] Focal category facets
  - [x] Shooting condition facets
  - [x] Burst facets
  - [x] Color facets
- [x] Implement facet count logic
  - [x] Respect active filters
  - [x] Exclude own dimension
  - [x] Sort by count or alphabetically
- [x] Create FacetCollection structure
  - [x] All facet types
  - [x] Selected state tracking
- [x] Add breadcrumb generation
  - [x] Hierarchical path tracking
  - [x] URL generation for each level

### 2.3 Color Search
**Status:** 100% complete ✅
**Priority:** High (specified in v1.0)

- [x] Implement color name search
  - [x] Hue range mapping (red, orange, yellow, etc.)
  - [x] Query by color name
- [x] Implement HSL filtering
  - [x] Hue range queries
  - [x] Saturation filtering (infrastructure ready)
  - [x] Lightness filtering (infrastructure ready)
- [x] Add color facets
  - [x] Dominant color distribution
- [x] Add CLI commands
  - [x] `olsen query --color red`
- [ ] Implement hex code search (future enhancement)
  - [ ] RGB tolerance (±30 per channel)
  - [ ] Similarity ranking

### 2.4 Repository/URL Mapper
**Status:** 100% complete ✅
**Priority:** High

- [x] Implement URL pattern matching
  - [x] `/YYYY/MM/DD` temporal patterns
  - [x] `/camera/:make/:model` equipment patterns
  - [x] `/color/:name` color patterns
  - [x] `/bursts` and `/bursts/:id` patterns
  - [x] Time of day patterns (`/morning`, `/afternoon`, etc.)
  - [x] Season patterns (`/spring`, `/summer`, etc.)
  - [x] Focal category patterns (`/wide`, `/normal`, `/telephoto`)
- [x] Implement URL-to-QueryParams conversion
  - [x] Path parameter extraction
  - [x] Query string parsing
  - [x] Validation
- [x] Implement QueryParams-to-URL generation
  - [x] Primary dimension selection
  - [x] Query string building
  - [x] URL encoding
- [x] Add CLI support
  - [x] `olsen query` command with 20+ flags
  - [x] Support for all filter types

### 2.5 Advanced Queries
**Status:** 100% complete ✅
**Priority:** Medium

- [x] Multi-value filters
  - [x] Comma-separated values
  - [x] OR logic within dimension
- [x] Range queries
  - [x] Date ranges
  - [x] ISO ranges
  - [x] Aperture ranges
  - [x] Focal length ranges
- [x] Location queries
  - [x] GPS presence filtering
  - [ ] Bounding box search (infrastructure ready)
- [x] Combined filters
  - [x] AND logic across dimensions
  - [x] Complex WHERE clauses

---

## Phase 3: Web Explorer Enhancements (Medium Priority)

### 3.1 Core UI Improvements
**Status:** 85% complete ✅
**Priority:** Medium

- [x] Add faceted filtering UI
  - [x] Sidebar with facet types (color, year, camera, time of day)
  - [x] Clickable links for each facet value
  - [x] Count display
  - [x] Selected state indicators (checkmarks)
- [x] Integrated query engine with web routes
  - [x] `/color/red`, `/2025/10`, `/camera/Canon/EOS-R5` patterns
  - [x] Query string parameter support
  - [x] Breadcrumb navigation
- [ ] Improve photo grid
  - [ ] Lazy loading
  - [ ] Infinite scroll option
  - [ ] Grid/list view toggle
  - [ ] Thumbnail size selector
- [ ] Add advanced search form
  - [ ] All query parameters
  - [ ] Date pickers
  - [ ] Range sliders (ISO, aperture)
  - [ ] Color picker

### 3.2 Burst Visualization
**Status:** 0% complete
**Priority:** Low

- [ ] Burst group view
  - [ ] Thumbnail strip
  - [ ] Sequence navigation
  - [ ] Representative indicator
  - [ ] Burst metadata display
- [ ] Burst detection in photo grid
  - [ ] Stack indicator
  - [ ] Expand/collapse
  - [ ] Count badge

### 3.3 Collections and Tags
**Status:** 0% complete
**Priority:** Low

- [ ] Tag management UI
  - [ ] Add/remove tags
  - [ ] Tag autocomplete
  - [ ] Tag cloud
- [ ] Collection management
  - [ ] Create collections
  - [ ] Add/remove photos
  - [ ] Smart collections (future)
- [ ] Batch operations
  - [ ] Select multiple photos
  - [ ] Bulk tag assignment
  - [ ] Bulk collection assignment

### 3.4 Enhanced Views
**Status:** 0% complete
**Priority:** Low

- [ ] Timeline view
  - [ ] Chronological scroll
  - [ ] Month/year headers
  - [ ] Date histogram
- [ ] Map view (for GPS-tagged photos)
  - [ ] Clustered markers
  - [ ] Photo preview on hover
  - [ ] Filter by map bounds
- [ ] Color palette view
  - [ ] Group by dominant color
  - [ ] Color histogram
  - [ ] Color harmony analysis

---

## Phase 4: Performance & Scale (Medium Priority)

### 4.1 Performance Testing
**Status:** 0% complete
**Priority:** High

- [ ] Create benchmark suite
  - [ ] Indexing throughput (target: 10-30 photos/sec)
  - [ ] Query performance (target: <500ms)
  - [ ] Facet computation (target: <500ms)
  - [ ] Burst detection performance
- [ ] Test at scale
  - [ ] 1,000 photo collection
  - [ ] 10,000 photo collection
  - [ ] 100,000 photo collection
  - [ ] Memory profiling
  - [ ] Database size tracking
- [ ] Optimize bottlenecks
  - [ ] Profile CPU usage
  - [ ] Optimize hot paths
  - [ ] Reduce allocations
  - [ ] Improve concurrency

### 4.2 Database Optimization
**Status:** Schema complete, queries need work
**Priority:** Medium

- [ ] Query optimization
  - [ ] Analyze query plans
  - [ ] Add covering indexes if needed
  - [ ] Optimize JOIN operations
- [ ] Database maintenance
  - [ ] Auto-vacuum setup
  - [ ] Index statistics updates
  - [ ] Corruption detection
- [ ] Storage optimization
  - [ ] Thumbnail compression (WebP option)
  - [ ] BLOB compression experiments
  - [ ] Incremental vacuum

### 4.3 Large-Scale Testing
**Status:** 0% complete
**Priority:** Medium

- [ ] Generate large test datasets
  - [ ] 1,000 photos with varied metadata
  - [ ] 10,000 photos
  - [ ] 100,000 photos (synthetic)
- [ ] Stress testing
  - [ ] Concurrent indexing
  - [ ] Concurrent queries
  - [ ] Database locking behavior
- [ ] Validate success criteria
  - [ ] 95%+ EXIF extraction rate
  - [ ] Query response times
  - [ ] Memory usage limits

---

## Phase 5: Documentation & Testing (Medium Priority)

### 5.1 Documentation
**Status:** Good specification docs, missing user docs
**Priority:** Medium

- [ ] User guide
  - [ ] Installation instructions
  - [ ] Getting started tutorial
  - [ ] Common workflows
  - [ ] Troubleshooting
- [ ] API documentation
  - [ ] CLI command reference
  - [ ] Query syntax guide
  - [ ] Web API endpoints (future)
- [ ] Developer documentation
  - [ ] Architecture overview (exists, needs update)
  - [ ] Code organization
  - [ ] Testing guide
  - [ ] Contributing guide

### 5.2 Test Coverage
**Status:** 65% (good unit tests, missing integration tests)
**Priority:** Medium

- [ ] Complete unit tests
  - [ ] Query engine tests
  - [ ] Repository/URL mapper tests
  - [ ] Facet computation tests
  - [ ] Color search tests
- [ ] Integration tests
  - [ ] End-to-end indexing (large dataset)
  - [ ] Complex query workflows
  - [ ] CLI command tests
  - [ ] Web UI tests (Selenium/Playwright)
- [ ] Validation tests
  - [ ] EXIF extraction accuracy
  - [ ] Burst detection accuracy
  - [ ] Performance benchmarks

---

## Phase 6: Future Enhancements (Low Priority / Optional)

### 6.1 Additional Format Support
**Priority:** Low

- [ ] RAW format support
  - [ ] CR2 (Canon)
  - [ ] NEF (Nikon)
  - [ ] ARW (Sony)
  - [ ] ORF (Olympus)
- [ ] Video support
  - [ ] Frame extraction
  - [ ] Video metadata
  - [ ] Duration tracking

### 6.2 Advanced Features
**Priority:** Low

- [ ] Face detection
  - [ ] Local processing (privacy)
  - [ ] Face clustering
  - [ ] People facet
- [ ] Object/scene detection
  - [ ] ML-based tagging
  - [ ] Automatic categorization
- [ ] XMP sidecar support
  - [ ] Read XMP files
  - [ ] Write metadata changes
- [ ] Full-text search
  - [ ] SQLite FTS5
  - [ ] Search across all fields

### 6.3 Smart Features
**Priority:** Low

- [ ] Smart collections
  - [ ] Rule-based collections
  - [ ] SQL query builder
  - [ ] Auto-update
- [ ] Recommendations
  - [ ] Similar photos
  - [ ] "More like this"
  - [ ] Best shots (sharpness + composition)
- [ ] Export features
  - [ ] HTML gallery generation
  - [ ] CSV export
  - [ ] JSON API

### 6.4 Advanced Performance
**Priority:** Low

- [ ] Perceptual hash optimization
  - [ ] Pre-compute distances
  - [ ] Similarity matrix caching
  - [ ] Approximate nearest neighbor search
- [ ] Distributed processing
  - [ ] Indexing job queue
  - [ ] Worker pool across machines
  - [ ] Shared database

---

## Completed Features ✓

### Core Indexing
- ✅ Recursive directory scanning (DNG, JPEG, BMP)
- ✅ EXIF metadata extraction (95% complete)
- ✅ Thumbnail generation (4 sizes: 64, 256, 512, 1024)
- ✅ Color palette extraction (k-means, 5 colors)
- ✅ RGB to HSL conversion
- ✅ Perceptual hash computation (pHash)
- ✅ File hash calculation (SHA-256)
- ✅ Metadata inference (time of day, season, focal category, shooting conditions)
- ✅ Concurrent processing with worker pool
- ✅ Progress reporting
- ✅ Error handling and logging
- ✅ Transaction-based database updates

### Database
- ✅ Complete schema implementation (100% matches spec)
- ✅ All indexes defined
- ✅ Foreign key constraints
- ✅ Facet metadata table
- ✅ Burst groups table
- ✅ Collections and tags tables

### Burst Detection
- ✅ Temporal proximity detection (2s window)
- ✅ Camera matching
- ✅ Focal length similarity
- ✅ Minimum burst size (3 photos)
- ✅ Representative selection (middle photo)
- ✅ Database storage

### Web Explorer
- ✅ HTTP server with embedded templates
- ✅ Home page with statistics
- ✅ Browse by year/month/day
- ✅ Browse by camera/lens
- ✅ Photo detail page
- ✅ Thumbnail serving API
- ✅ Pagination
- ✅ Responsive layout
- ✅ Photo navigation (prev/next)

### Testing
- ✅ Comprehensive unit tests for indexer
- ✅ EXIF extraction tests
- ✅ Thumbnail generation tests
- ✅ Color extraction tests
- ✅ Perceptual hash tests
- ✅ Burst detection tests
- ✅ Inference tests
- ✅ Test fixture generation

---

## Known Issues & Technical Debt

### High Priority
1. **CLI is almost non-existent** - Only `explore` command implemented
2. **No faceted search** - Core feature of v2.0 spec missing
3. **No color search** - Specified in v1.0, not implemented
4. **Query engine is minimal** - Only basic queries work

### Medium Priority
5. **No performance validation** - Haven't tested at scale (100K+ photos)
6. **Missing focus distance extraction** - EXIF field exists, extraction incomplete
7. **No incremental indexing** - Re-processes all files every time
8. **Web UI lacks advanced features** - No faceted filtering, no burst visualization

### Low Priority
9. **Burst representative selection is simplistic** - Uses middle photo, not sharpest
10. **No database maintenance tools** - No vacuum, no integrity checks
11. **Limited error reporting** - Errors logged but not structured
12. **Test coverage gaps** - No integration tests at scale

---

## Quick Wins (Easy Improvements)

1. **Add basic CLI commands** (1-2 days)
   - `olsen stats` - easy, repo methods exist
   - `olsen show <id>` - straightforward
   - `olsen thumbnail <id>` - simple file I/O

2. **Add color name search** (1 day)
   - Hue ranges already defined in inference code
   - Just needs query building

3. **Improve web UI navigation** (1 day)
   - Add "All Photos" link
   - Add search box
   - Add filter clear button

4. **Add error log export** (0.5 day)
   - Export failed files to CSV
   - Include error messages

5. **Add incremental indexing** (2-3 days)
   - Check file hash before processing
   - Skip unchanged files

---

## Dependencies & Requirements

### Current Dependencies (all installed)
- Go 1.25+
- SQLite 3
- github.com/corona10/goimagehash (perceptual hashing)
- github.com/dsoprea/go-exif/v3 (EXIF extraction)
- github.com/mattn/go-sqlite3 (SQLite driver)
- github.com/mccutchen/palettor (color extraction)
- github.com/nfnt/resize (image resizing)

### Future Dependencies
- github.com/google/uuid (for IDs) - may already be indirect
- Testing frameworks (Testify for assertions, optional)
- WebP encoder (for thumbnail optimization, future)

---

## Development Priorities Summary

### Sprint 1 (Weeks 1-2): CLI Foundation
Focus: Make Olsen usable from command line
1. Implement core CLI commands (index, stats, show)
2. Add color search functionality
3. Improve error reporting

### Sprint 2 (Weeks 3-4): Query System
Focus: Build out query engine
1. Implement QueryParams structure
2. Build SQL query builder
3. Add multi-dimensional filtering

### Sprint 3 (Weeks 5-6): Faceted Search
Focus: Core v2.0 feature
1. Implement facet computation
2. Add facet count logic
3. Integrate with web UI

### Sprint 4 (Weeks 7-8): Performance & Polish
Focus: Validation and optimization
1. Performance benchmarking
2. Large-scale testing
3. Bug fixes and refinements

---

## Success Metrics (from v1.0 spec)

### Must Meet
- [ ] Index 10,000 DNG files without errors
- [ ] Extract metadata from 95%+ of files
- [ ] Process at 10-30 photos/second
- [ ] Query response time < 1 second
- [ ] Database size reasonable (~200KB per photo including thumbnails)
- [ ] Memory usage < 500MB during indexing

### Nice to Have
- [ ] Color search returns relevant results
- [ ] Faceted search works smoothly
- [ ] Burst detection 90%+ accuracy
- [ ] Works on Windows, macOS, Linux

---

**Next Steps:** Start with Sprint 1 - implement core CLI commands to make Olsen practically usable for basic workflows.
