# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build the CLI
make build

# Index photos
./bin/olsen index <path-to-photos> --db photos.db --w 4

# Run burst detection
./bin/olsen analyze --db photos.db

# View statistics
./bin/olsen stats --db photos.db

# Show photo metadata
./bin/olsen show <photo-id> --db photos.db

# Extract thumbnail
./bin/olsen thumbnail -o output.jpg -s 512 <photo-id> --db photos.db

# Verify database integrity
./bin/olsen verify --db photos.db

# Start web explorer
./bin/olsen explore --db photos.db --addr localhost:8080
# Or use the helper script:
./explorer.sh --db photos.db --open
```

## Building with RAW Support

Olsen supports two LibRaw Go bindings with seamless switching:

```bash
# Build with seppedelanghe/go-libraw (default - more capable)
make build-seppedelanghe
# or simply:
make build-raw

# Build with inokone/golibraw (simpler, stable fallback)
make build-golibraw

# Check which library is active
./bin/olsen version
```

**Why two libraries?**
- `seppedelanghe/go-libraw`: Full configuration control, complete diagnostics, required for quality instrumentation
- `inokone/golibraw`: Simple, stable, no configuration options

See `docs/LIBRAW_DUAL_LIBRARY_SUPPORT.md` for detailed comparison.

## Testing

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/indexer/

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchtime=3s ./internal/indexer/

# Benchmark LibRaw libraries
make benchmark-libraw  # Compares both libraries
```

## Architecture Overview

Olsen is a read-only photo indexing system that extracts metadata from DNG, JPEG, and BMP files into a portable SQLite database. The system **never modifies source photo files**.

### Core Components

**1. Indexer Engine** (`internal/indexer/`)
- `indexer.go` - Main concurrent processing engine with worker pool
- `metadata.go` - EXIF extraction using go-exif (handles DNG/JPEG/BMP)
- `thumbnail.go` - Aspect-ratio-preserving thumbnails (4 sizes: 64, 256, 512, 1024px longest edge)
- `color.go` - K-means color palette extraction (5 dominant colors) + RGB-to-HSL conversion
- `phash.go` - Perceptual hash computation for near-duplicate detection
- `inference.go` - Metadata inference (time of day, season, focal length category, shooting conditions)
- `burst.go` - Temporal burst detection (2-second window, min 3 photos)

**2. Database Layer** (`internal/database/`)
- `schema.go` - Complete SQLite schema (photos, thumbnails, photo_colors, burst_groups, tags, collections)
- `database.go` - Database operations with WAL mode enabled for concurrent read access
- Uses transaction-based inserts for consistency
- All thumbnails stored as BLOBs in database (portable catalog design)

**3. Explorer Web UI** (`internal/explorer/`)
- `server.go` - HTTP server with embedded HTML templates
- `repository.go` - Query methods for photo retrieval (by year/month/day, by camera/lens)
- `templates/grid.html` - Main faceted search UI with right-rail layout
- Serves thumbnails from database with cache-busting (indexed_at timestamps)
- **State machine-based faceted navigation** (see Query Engine below)

**4. Query Engine** (`internal/query/`)
- `engine.go` - Main query execution with SQL building
- `facets.go` - Facet computation (counts for Year, Month, Camera, Color, etc.)
- `facet_url_builder.go` - URL generation for facet values
- `url_mapper.go` - URL parsing and generation
- `types.go` - Query parameters and facet data structures
- **CRITICAL**: Implements state machine model for faceted navigation (see below)

**5. Data Models** (`pkg/models/`)
- `types.go` - Core data structures (PhotoMetadata, Color, DominantColor, ThumbnailSize, IndexStats)

### Key Design Decisions

**State Machine-Based Faceted Navigation:**
- Facets are **independent dimensions**, not hierarchical
- Fundamental rule: **Users cannot transition from a state with results to a state with zero results**
- ALL filters are preserved during transitions (Year, Month, Color, Camera, etc.)
- SQL queries with WHERE clauses + GROUP BY naturally compute valid transitions
- Facet values with count=0 are shown but disabled in UI
- **NO hardcoded clearing logic** based on assumed relationships

**Example:**
- State: `year=2024&month=11` (50 photos from November 2024)
- Year facet shows: 2023 (120) ✓, 2024 (50) ✓ selected, 2025 (0) ✗ disabled
- Click 2023 → `year=2023&month=11` (120 photos from November 2023)
- Month filter PRESERVED because combination exists in data

See `specs/facet_state_machine.spec` and `docs/HIERARCHICAL_FACETS.md` for detailed explanation.

**Database Portability:** The SQLite database IS the catalog. It contains all metadata, thumbnails (4 sizes), color palettes, and perceptual hashes. Original photo files are only needed for full-resolution access. This enables:
- Single-file backup/restore
- Offline browsing (originals can be disconnected)
- Easy database migration

**Concurrent Processing:** Worker pool architecture (default 4 workers) processes photos in parallel. Uses sync.Mutex to protect shared statistics. Progress callbacks report every file processed.

**Aspect-Ratio Preservation:** Thumbnails constrain the longest edge (not forced square crops), maintaining photo composition.

**Read-Only Guarantee:** All file operations use read-only mode. Processing happens entirely in memory. Only the database is modified.

**WAL Mode:** Database uses Write-Ahead Logging, enabling concurrent reads during indexing. The explorer can be used while indexing is running.

## Code Organization Patterns

### Indexer Processing Flow
```
IndexDirectory()
  → findDNGFiles() (recursive scan)
  → worker pool processes files concurrently
  → processFile() for each photo:
      1. Check if already indexed (by file_path)
      2. ExtractMetadata() - EXIF extraction
      3. calculateFileHash() - SHA-256
      4. image.Decode() - open image
      5. GenerateThumbnailsFromImage() - 4 sizes
      6. ExtractColorPalette() - k-means on 256px thumbnail
      7. ComputePerceptualHash() - pHash from thumbnail
      8. InferMetadata() - classify time/season/etc
      9. db.InsertPhoto() - single transaction
```

### Database Insertion Pattern
All database inserts use transactions:
```go
tx, err := db.Begin()
// ... insert operations
tx.Commit()
```

Photos are inserted into `photos` table, then thumbnails into `thumbnails` table, then colors into `photo_colors` table. Foreign key constraints maintain referential integrity with CASCADE delete.

### CLI Command Pattern
Each command in `cmd/olsen/main.go` follows this structure:
```go
func commandNameCommand(args []string) error {
    flags := flag.NewFlagSet("commandName", flag.ExitOnError)
    // Define flags
    flags.Parse(args)

    // Open database
    db, err := database.Open(dbPath)
    defer db.Close()

    // Execute operation
    // Display results
    return nil
}
```

Flag order matters: flags must come BEFORE positional arguments in Go's flag package.

## Critical Implementation Details

**EXIF Extraction:** Uses `github.com/dsoprea/go-exif/v3` which supports all standard EXIF tags including Flash. For BMP files without EXIF, falls back to basic file metadata (size, mod time).

**Thumbnail Generation:** Uses Lanczos3 resampling (`github.com/nfnt/resize`) for high-quality downsampling. Thumbnails are stored as JPEG with 85% quality.

**Color Extraction & Classification:**
- Uses `github.com/mccutchen/palettor` for k-means clustering on 256px thumbnail
- Extracts 5 dominant colors with weights
- Converts RGB to HSL for perceptual classification
- Classifies into 11 Berlin-Kay universal basic colors:
  - **Achromatic**: black, white, gray, b&w (near-grayscale with S < 15%)
  - **Chromatic**: red, orange, yellow, green, blue, purple, pink
  - **Special**: brown (orange hue 20-40° with lightness < 50%)
- **Saturation-first logic** prevents B&W photos from being misclassified as colored
- See `specs/dominant_colours.spec` for complete algorithm

**Perceptual Hash:** Uses `github.com/corona10/goimagehash` to compute 64-bit pHash. Hamming distance calculates similarity (threshold: 10 bits = near-duplicate).

**Burst Detection:** Groups photos taken within 2 seconds with same camera and similar focal length. Minimum burst size is 3 photos. Representative is middle photo (TODO: use sharpest).

## Current Status & TODO

See `TODO.md` for comprehensive task tracking.

**Completed (~80% overall):**
- ✅ Core indexing engine (100%)
- ✅ Database schema (100%)
- ✅ CLI commands (90% - index, analyze, stats, show, thumbnail, verify, explore)
- ✅ Burst detection (100%)
- ✅ Web explorer UI (90% - state machine faceted navigation)
- ✅ **Query engine (95% - state machine model implemented)**
- ✅ **Faceted search system (90% - implemented with state machine model)**
- ✅ **Color classification (100% - 11 Berlin-Kay colors with B&W support)**
- ✅ **URL mapper (95% - full parsing and generation)**

**Recent Major Achievements:**
- ✅ Implemented state machine-based faceted navigation (replaces incorrect hierarchical model)
- ✅ Added 11-color classification system with proper B&W handling
- ✅ Fixed brown/orange confusion with lightness-based distinction
- ✅ Comprehensive test suite (90+ color tests, state transition tests, facet tests)

**Remaining Work:**
- ⚠️ UI: Disable zero-count facet values (prevent invalid transitions)
- ⚠️ Progressive disclosure for temporal facets (show Month after Year selected)
- ⚠️ Duplicate detection UI and clustering visualization

**Performance Target:**
- 10-30 photos/second indexing throughput
- <500ms query response time
- Scale to 100K+ photos

Current benchmarks (M3 Max): ~62ms per photo for metadata extraction + thumbnails + color + hash. With 4 workers: ~431ms per photo observed (includes I/O).

## Development Workflow

**When modifying the indexer:**
1. Update the relevant component file (metadata.go, thumbnail.go, etc.)
2. Add/update unit tests in corresponding `*_test.go` file
3. Run tests: `go test -v ./internal/indexer/`
4. Test with real photos: `./bin/olsen index testdata/dng --db test.db --w 2`

**When modifying the database schema:**
1. Update `internal/database/schema.go`
2. Update `pkg/models/types.go` if needed
3. Update insert/query methods in `internal/database/database.go`
4. Delete old test databases and regenerate: `rm test.db && ./bin/olsen index testdata/dng --db test.db`

**When adding CLI commands:**
1. Add case to main switch in `cmd/olsen/main.go`
2. Implement `commandNameCommand(args []string) error` function
3. Use `flag.NewFlagSet` for argument parsing
4. Remember: flags before positional args

**When modifying the web UI:**
1. Update `internal/explorer/server.go` (HTTP handlers)
2. Update `internal/explorer/templates/grid.html` (faceted search UI)
3. Update `internal/explorer/repository.go` for new query methods
4. Rebuild and restart: `make build && ./bin/olsen explore --db test.db`

**When modifying faceted navigation:**
1. **CRITICAL**: Follow state machine model, NOT hierarchical model
2. Update `internal/query/facets.go` for facet computation (SQL with WHERE clauses)
3. Update `internal/query/facet_url_builder.go` for URL generation (preserve ALL filters)
4. **NEVER** add clearing logic based on assumed hierarchies
5. Add tests in `internal/query/facet_state_machine_test.go`
6. See `specs/facet_state_machine.spec` for principles

## Dependencies

Core dependencies (all specified in `go.mod`):
- `github.com/dsoprea/go-exif/v3` - EXIF metadata extraction
- `github.com/mattn/go-sqlite3` - SQLite driver (requires CGO)
- `github.com/nfnt/resize` - Image resizing
- `github.com/mccutchen/palettor` - K-means color extraction
- `github.com/corona10/goimagehash` - Perceptual hashing
- `golang.org/x/image/bmp` - BMP format support

## Test Data

**Location:** `testdata/dng/` contains 13 sample DNG files for testing

**Generating fixtures:**
```bash
go run testdata/generate_fixtures.go
```

**Creating test database:**
```bash
./bin/olsen index testdata/dng --db test.db --w 2
```

## Common Pitfalls

1. **Flag parsing order:** Go's flag package requires flags BEFORE positional args
   - ✅ `./bin/olsen thumbnail -o out.jpg -s 512 2`
   - ❌ `./bin/olsen thumbnail 2 -s 512 -o out.jpg`

2. **Database locking:** Don't run multiple indexers on same database simultaneously (single writer in WAL mode). Multiple readers are fine.

3. **EXIF extraction failures:** Some images lack EXIF data (especially BMP). Code falls back to basic file metadata - this is expected behavior, not an error.

4. **Memory usage:** Large photos are decoded into memory for processing. With 4 workers, expect ~500MB peak usage.

5. **Progress callback timing:** Progress updates happen after successful processing, not during. Failed files don't increment processed count.

## Specification Documents

For detailed requirements and architecture:
- `specs/olsen_specs.md` - Complete v2.0 system specification
- `README.md` - User-facing documentation
- `TODO.md` - Detailed implementation status and next steps
