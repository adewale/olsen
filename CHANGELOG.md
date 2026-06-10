# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- `DeletePhoto` referenced a non-existent `color_palette` table (schema uses `photo_colors`), which made every re-index of a modified file fail; it now deletes related rows explicitly and is covered by regression tests
- SQLite PRAGMAs (`foreign_keys`, `busy_timeout`, `journal_mode`, `synchronous`) are now passed as DSN parameters so they apply to every pooled connection, not just the first; `busy_timeout=5000` prevents `database is locked` errors when browsing during indexing
- `?limit=0` on `/photos` caused a division-by-zero panic; pagination limits are now clamped to 1–500 and offsets to 0–1,000,000
- The 60s per-file indexing timeout now actually cancels the work via context instead of abandoning a goroutine that kept decoding and inserted into the database after being reported as failed
- One unreadable directory no longer aborts indexing of the entire library (walk errors are logged and skipped)
- Images without EXIF dimensions are now guarded against decompression bombs via a header-only `image.DecodeConfig` check before full decode
- Unknown URLs in the explorer now return 404 instead of a blank 200 page
- Skipped files are no longer double-counted as processed in indexing stats
- Flags can now be passed before or after positional CLI arguments (`olsen index <dir> --db x.db` works as documented in the README)
- CLI integration tests build the binary themselves and the full `go test ./...` suite passes without pre-existing fixtures
- Intentionally-failing diagnostic tests converted to skipped documentation tests

### Added
- Ctrl+C / SIGTERM handling: `index` stops cleanly at the next stage boundary (committed photos are kept; re-run resumes), `explore` shuts down gracefully
- Schema versioning via `PRAGMA user_version` with a migration hook for future schema changes
- HTTP server timeouts (read/write/idle) on the explorer
- `internal/quality` unit tests (orientation, SSIM/MSE metrics, thumbnail pipeline, cancellation)
- `.golangci.yml` lint configuration; codebase is lint-clean
- CI: `go mod verify`, full-package `go vet`, golangci-lint, race-detector test run, govulncheck, and a LibRaw build job for both RAW libraries
- Build-time version injection via `-ldflags` (`make build` embeds `git describe` output)
- `benchmark-libraw` / `benchmark-thumbnails` commands are now wired into the CLI (RAW builds)

### Changed
- `golang.org/x/image` upgraded v0.31.0 → v0.41.0, fixing three reachable vulnerabilities (GO-2026-5032, GO-2026-5031, GO-2026-4815); `golang.org/x/net` upgraded from a 2022 snapshot to v0.55.0; Go toolchain raised to 1.25
- Explorer error responses no longer leak internal error details to clients (logged server-side instead)
- `make test` / `make test-ci` run the real test suite (CGO on, race detector in CI) instead of a narrow always-green subset
- Windows builds no longer fail on the unix-only disk-space check (now build-tagged with a graceful fallback)

### Removed
- ~165 lines of dead, unregistered explorer handlers; unused `compare-raw` Makefile target referencing a non-existent file; committed test log files

## [0.1.0] - 2025-10-12

### Added

#### Core Indexing Engine
- Photo indexing for DNG, JPEG, and BMP files with concurrent processing
- EXIF metadata extraction using go-exif library
- Aspect-ratio-preserving thumbnail generation (4 sizes: 64px, 256px, 512px, 1024px longest edge)
- Color palette analysis with k-means clustering (5 dominant colors per photo)
- RGB to HSL color space conversion for perceptual color classification
- Perceptual hash (pHash) computation for near-duplicate detection
- Metadata inference: time of day, season, focal length category, shooting conditions
- Burst detection using 2-second temporal window
- Worker pool architecture for parallel file processing (configurable worker count)
- SHA-256 file hashing with modification detection
- Read-only guarantee: never modifies source photo files

#### Database & Storage
- SQLite database with comprehensive schema
- Photos table with 50+ metadata fields
- Thumbnails table with 4 sizes per photo stored as BLOBs
- Photo_colors table for dominant color data with HSL values
- Burst_groups, duplicate_clusters, tags, and collections tables
- WAL (Write-Ahead Logging) mode for concurrent read access
- Transaction-based inserts for data consistency
- Foreign key constraints with CASCADE delete
- Portable single-file database design

#### Web Explorer UI
- HTTP server with embedded HTML templates
- State machine-based faceted navigation
- Dynamic thumbnail serving with ETag caching
- Breadcrumb navigation
- Active filter display with removal links
- Responsive photo grid layout
- Photo detail view
- Statistics dashboard

#### Faceted Search Query Engine
- State machine navigation model (no hierarchical assumptions)
- Independent filter dimensions: Year, Month, Day, Color, Camera, Lens, etc.
- SQL-based query building with WHERE clauses
- Facet computation with result counts
- URL-based filter state management
- Zero-result prevention: disabled facet values shown but not clickable
- Filter preservation across transitions
- Support for multiple filter values (multi-select)

#### Color Classification System
- 11 Berlin-Kay universal basic colors
- Achromatic colors: black, white, gray, b&w (near-grayscale)
- Chromatic colors: red, orange, yellow, green, blue, purple, pink
- Special colors: brown (dark orange with low lightness)
- Saturation-first logic prevents B&W photos from being misclassified
- See `specs/dominant_colours.spec` for complete algorithm

#### CLI Commands
- `index` - Index photos from a directory
- `analyze` - Run burst detection
- `stats` - Display database statistics
- `show` - Show photo metadata
- `thumbnail` - Extract thumbnails
- `verify` - Verify database integrity
- `explore` - Start web explorer server
- `version` - Display version and RAW support status

#### Documentation
- Comprehensive README with architecture overview
- CLAUDE.md for AI-assisted development guidance
- Technical specifications in `specs/` directory
- Lessons learned documentation
- State machine faceted navigation specification
- DNG format deep dive
- Testing guide
- Performance benchmarks

### Technical Details

#### Dependencies
- github.com/dsoprea/go-exif/v3 - EXIF metadata extraction
- github.com/mattn/go-sqlite3 - SQLite database driver
- github.com/nfnt/resize - Image resizing with Lanczos3
- github.com/mccutchen/palettor - K-means color extraction
- github.com/corona10/goimagehash - Perceptual hashing

#### Performance
- ~15-25 photos/second indexing throughput on Apple M3 Max
- ~62ms per photo average processing time
- File hash: 0.4ms
- Thumbnail generation: 34ms
- Color extraction: 28ms
- Perceptual hash: 0.2ms

#### Architecture Decisions
- Aspect-ratio preservation over forced square crops
- Database portability (single SQLite file)
- Concurrent processing with worker pool pattern
- pHash algorithm for duplicate detection
- K-means clustering for color extraction
- State machine model for faceted navigation
- Read-only file access guarantee

### Known Limitations

- LibRaw support requires CGO and libraw library installation
- Limited RAW format support (DNG primarily, others via LibRaw)
- No GUI (CLI + web UI only)
- Single database (no distributed mode)
- Read-only design (cannot write EXIF back to files)
- Test fixtures may need to be generated by users

### Breaking Changes

None (initial release)

### Deprecated

None (initial release)

### Security

- Read-only file access prevents accidental file modifications
- No network operations during indexing
- Local SQLite database with file system permissions

---

## Release Notes

### v0.1.0 - Initial Public Release

This is the first public release of Olsen, a portable photo corpus explorer with state machine-based faceted search.

**Highlights:**
- Complete photo indexing pipeline for DNG/JPEG/BMP files
- Web-based explorer with intuitive faceted navigation
- 11-color classification system based on Berlin-Kay universal colors
- High-performance concurrent indexing (15-25 photos/second)
- Read-only guarantee: your photos are never modified
- Portable catalog: single SQLite file contains all metadata and thumbnails

**Getting Started:**
```bash
# Index your photos
./bin/olsen index ~/Pictures/Photos --db my-photos.db --w 4

# Start the web explorer
./bin/olsen explore --db my-photos.db --addr localhost:8080

# Open http://localhost:8080 in your browser
```

**Documentation:**
- README.md - Quick start and architecture overview
- CONTRIBUTING.md - Development guidelines
- specs/ - Technical specifications
- docs/ - Detailed documentation and lessons learned

**Repository:** https://github.com/adewale/olsen

---

[Unreleased]: https://github.com/adewale/olsen/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/adewale/olsen/releases/tag/v0.1.0
