# Olsen - Photo Indexer

[![CI](https://github.com/adewale/olsen/actions/workflows/ci.yml/badge.svg)](https://github.com/adewale/olsen/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

WARNING: This project is super early and should not be used on valuable data.

A high-performance photo indexing system for DNG (Digital Negative), JPEG, and BMP files that extracts comprehensive metadata, generates aspect-ratio-preserving thumbnails, analyzes color palettes, and computes perceptual hashes for similarity detection.

## Supported Formats

- **DNG (Digital Negative)**: Adobe's RAW format with full EXIF metadata extraction
- **JPEG**: Standard photographs with EXIF metadata support
- **BMP**: Bitmap images (typically scanned photographs) with basic metadata

## ⚠️ Critical Guarantee: Read-Only Operation

**Olsen NEVER modifies your photo files.** The indexer is strictly read-only and will never:
- ❌ Modify, move, rename, or delete photo files
- ❌ Write EXIF data back to files
- ❌ Create temporary files in photo directories

All file access uses read-only mode (`O_RDONLY`). Image processing happens entirely in memory. Only the SQLite database is modified.

## Features

### Indexer Implementation

- **EXIF Metadata Extraction**: Extracts camera, lens, exposure, location, temporal, and lighting metadata
- **Aspect-Ratio-Preserving Thumbnails**: Generates 4 sizes (64px, 256px, 512px, 1024px) with longest edge constraint
- **Color Palette Analysis**: Extracts top 5 dominant colors using k-means clustering with RGB and HSL values
- **Perceptual Hashing**: Computes pHash for near-duplicate detection and similarity matching
- **Metadata Inference**: Automatically classifies time of day, season, focal length category, and shooting conditions
- **Concurrent Processing**: Multi-worker architecture for parallel file processing
- **SQLite Storage**: Portable database with all metadata, thumbnails, and color data
- **🔒 Read-Only**: Guaranteed never to modify source photo files

### Components

1. **Models** (`pkg/models/`)
   - Core data structures (PhotoMetadata, Color, ThumbnailSize)
   - Statistics tracking

2. **Database** (`internal/database/`)
   - SQLite schema with comprehensive indexing
   - Photo, thumbnail, color, burst, duplicate, tag, and collection tables
   - Transaction-based inserts

3. **Indexer** (`internal/indexer/`)
   - `metadata.go`: EXIF extraction from DNG/JPEG files
   - `thumbnail.go`: Multi-size thumbnail generation
   - `color.go`: Color palette extraction and HSL conversion
   - `phash.go`: Perceptual hashing and similarity detection
   - `inference.go`: Metadata inference (time of day, season, etc.)
   - `indexer.go`: Main engine with concurrent processing

## Documentation

- **[Architecture Diagram](docs/architecture.md)**: System component overview
- **[Processing Flow](docs/flow.md)**: Detailed indexing workflow diagrams
- **[Specifications](specs/olsen_specs.md)**: Complete technical specifications
- **[Requirements](specs/olsen_requirements.md)**: Functional and non-functional requirements

## Architecture

```
┌─────────────────────────────────────┐
│        Indexer Engine               │
│  (Concurrent file processing)       │
└────────────┬────────────────────────┘
             │
    ┌────────┼────────┬────────┬──────┐
    │        │        │        │      │
┌───▼──┐ ┌──▼───┐ ┌──▼───┐ ┌──▼──┐ ┌─▼────┐
│ EXIF │ │Thumb │ │Color │ │pHash│ │Infer │
│Extract│ │Gen  │ │Palette│ │Comp │ │Meta │
└───┬──┘ └──┬───┘ └──┬───┘ └──┬──┘ └─┬────┘
    │       │        │        │      │
    └───────┴────────┴────────┴──────┘
                    │
         ┌──────────▼──────────┐
         │   SQLite Database   │
         │  (Portable catalog) │
         └─────────────────────┘
```

## Performance

Benchmarks on Apple M3 Max:

| Operation | Time per photo |
|-----------|---------------|
| File hash calculation | 0.4 ms |
| Thumbnail generation (4 sizes) | 34 ms |
| Color palette extraction | 28 ms |
| Perceptual hash | 0.2 ms |
| **Total** | **~62 ms** |

**Expected throughput**: 15-25 photos/second with 8 workers

## Test Coverage

Comprehensive test suite with 100% passing tests:

- **Color Tests**: RGB/HSL conversion, palette extraction, distance calculations
- **Thumbnail Tests**: Aspect ratio preservation for landscape, portrait, and square images
- **Perceptual Hash Tests**: Hash generation, Hamming distance, similarity detection
- **Inference Tests**: Time of day, season, focal length, shooting conditions
- **Integration Tests**: Full indexer workflow with database

### Running Tests

The project uses a two-tier testing strategy:

```bash
# Quick tests (no CGO, no database)
make test

# This runs unit tests that don't require SQLite (URL parsing, color classification, etc.)
# Database-dependent tests are skipped when CGO_ENABLED=0
```

**Note**: Most integration tests require CGO_ENABLED=1 and SQLite. The default `make test` target uses CGO_ENABLED=0 to avoid requiring LibRaw dependencies in CI. Command-line parsing tests and non-database tests run successfully without CGO.

Run benchmarks:
```bash
go test -bench=. -benchtime=3s ./internal/indexer/
```

## Database Schema

- **photos**: Core metadata (50+ fields)
- **thumbnails**: 4 sizes per photo (64, 256, 512, 1024px)
- **photo_colors**: Dominant colors with weights and HSL values
- **burst_groups**: Temporal burst detection
- **duplicate_clusters**: Perceptual hash-based clustering
- **tags**: User-defined tags
- **collections**: Virtual photo collections

## Key Design Decisions

1. **Aspect-Ratio Preservation**: Thumbnails constrain longest edge instead of forcing square crops
2. **Database Portability**: All metadata and thumbnails stored in single SQLite file
3. **Concurrent Processing**: Worker pool pattern for parallel file processing
4. **Perceptual Hashing**: pHash algorithm for near-duplicate detection
5. **Color Extraction**: K-means clustering on 256px thumbnail for efficiency

## Dependencies

- `github.com/dsoprea/go-exif/v3`: EXIF metadata extraction (supports all standard tags including Flash)
- `github.com/mattn/go-sqlite3`: SQLite database driver
- `github.com/nfnt/resize`: Image resizing with Lanczos3
- `github.com/mccutchen/palettor`: K-means color palette extraction
- `github.com/corona10/goimagehash`: Perceptual hashing

## Web Explorer (Faceted Search)

Olsen includes a web-based photo explorer with **state machine-based faceted navigation**:

### Key Feature: State Machine Model

**Faceted navigation treats photo exploration as a state machine** where users navigate through valid data combinations. The fundamental rule:

> Users can never transition from a state with results to a state with zero results.

**How it works:**
- ALL filters are preserved during transitions (Year, Month, Color, Camera, etc.)
- SQL queries compute which facet values have results given current filters
- Facet values with count=0 are shown but disabled
- No hardcoded "hierarchical" relationships - data determines valid paths

**Example:** Viewing `year=2024&month=11` (50 photos):
- Year facet shows: 2023 (120) ✓, 2024 (50) ✓ selected, 2025 (0) ✗ disabled
- Clicking 2023 → `year=2023&month=11` (November 2023 photos)
- Month filter preserved because combination exists in data

See `specs/facet_state_machine.spec` and `docs/HIERARCHICAL_FACETS.md` for detailed explanation.

### Available Facets
- **Temporal**: Year, Month, Day
- **Visual**: Color (11 Berlin-Kay universal colors), Time of Day, Season
- **Equipment**: Camera (make + model), Lens
- **Technical**: Focal Category, Shooting Condition, In Burst

### Color Classification
Olsen classifies photos into 11 universal color categories using HSL color space:
- **Achromatic**: black, white, gray, b&w (near-grayscale)
- **Chromatic**: red, orange, yellow, green, blue, purple, pink
- **Special**: brown (dark orange with low lightness)

See `specs/dominant_colours.spec` for algorithm details.

## Next Steps

See `specs/olsen_specs.md` for full system specification including:

- Query Engine with faceted search (implemented as state machine)
- URL Repository for RESTful browsing
- Burst detection algorithm
- Duplicate clustering with BK-tree optimization
- CLI interface
- Analysis and statistics

## Installation

### From Source
```bash
# Clone the repository
git clone https://github.com/adewale/olsen.git
cd olsen

# Build without RAW support (faster, no CGO)
make build

# Or build with LibRaw support (requires libraw installed)
make build-raw

# Binary will be in bin/olsen
./bin/olsen --help
```

### Requirements
- Go 1.21 or later
- SQLite 3 (included via go-sqlite3)
- For RAW support: libraw library

## Quick Start

```bash
# Index your photos
./bin/olsen index ~/Pictures/Photos --db my-photos.db --w 4

# Start the web explorer
./bin/olsen explore --db my-photos.db --addr localhost:8080

# Open http://localhost:8080 in your browser
```

## Repository

**Official Repository:** https://github.com/adewale/olsen

## License

See LICENSE file for details.
