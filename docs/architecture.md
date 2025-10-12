# Olsen Architecture

## System Overview

Olsen is designed as a modular, high-performance photo indexing system with four distinct architectural layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│              (Command-line interface & output)                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
       ┌──────▼──────────┐         ┌───────▼────────┐
       │   Repository     │         │    Indexer     │
       │   (URL Mapper)   │         │    Engine      │
       │   [Future]       │         │  [Implemented] │
       └──────┬───────────┘         └───────┬────────┘
              │                             │
              │         ┌───────────────────┘
              │         │
       ┌──────▼─────────▼──────┐
       │    Query Engine        │
       │  (Faceted Search)      │
       │      [Future]          │
       └──────┬─────────────────┘
              │
┌─────────────▼──────────────────────────────────────────────────┐
│                     SQLite Database                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ photos: Core metadata (camera, exposure, location)      │   │
│  │ thumbnails: 4 sizes (64, 256, 512, 1024px)             │   │
│  │ photo_colors: Dominant colors (RGB + HSL)              │   │
│  │ burst_groups: Temporal burst detection                 │   │
│  │ tags: User-defined tags                                │   │
│  │ collections: Virtual photo collections                 │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. CLI Layer (Future)
**Purpose**: User-facing commands and output formatting

**Responsibilities**:
- Parse command-line arguments
- Execute indexing, query, and management commands
- Format and display results to user
- Handle progress reporting

**Commands**:
```bash
olsen index <path>              # Index photos
olsen query <filters>           # Search photos
olsen stats                     # Show statistics
olsen analyze                   # Run burst analysis
```

### 2. Indexer Engine (Implemented)
**Purpose**: Extract metadata and visual features from photos

**⚠️ CRITICAL GUARANTEE: READ-ONLY OPERATION**

The indexer is **strictly read-only** and will NEVER modify, move, delete, or otherwise alter source photo files.

**Enforcement mechanisms:**
- All file access uses `os.Open()` with `O_RDONLY` flag (read-only mode)
- No use of write operations (`os.Create`, `os.OpenFile`, `os.WriteFile`, `os.Remove`, `os.Rename`)
- Image processing is entirely in-memory (no writes back to disk)
- EXIF parsing operates on byte buffers only
- Code reviews must verify no new write operations are introduced

**What the indexer does:**
- ✅ Reads photo files to extract EXIF metadata
- ✅ Decodes images in memory to generate thumbnails
- ✅ Computes perceptual hashes and dominant colors
- ✅ Writes extracted data to SQLite database

**What the indexer NEVER does:**
- ❌ Modify photo files
- ❌ Write EXIF data back to files
- ❌ Move or rename files
- ❌ Delete files
- ❌ Create temporary files in photo directories

**Components**:
```
┌──────────────────────────────────────────────────────────┐
│                    Indexer Engine                         │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   Worker 1   │  │   Worker 2   │  │   Worker N   │   │
│  │              │  │              │  │              │   │
│  │ processFile()│  │ processFile()│  │ processFile()│   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │
│         │                 │                 │           │
│         └─────────────────┴─────────────────┘           │
│                          │                              │
│               ┌──────────▼──────────┐                    │
│               │   Work Channel      │                    │
│               │  (buffered queue)   │                    │
│               └──────────┬──────────┘                    │
│                          │                              │
│               ┌──────────▼──────────┐                    │
│               │  File Scanner       │                    │
│               │ (recursive walk)    │                    │
│               └─────────────────────┘                    │
└──────────────────────────────────────────────────────────┘
```

**Processing Pipeline** (per file):
```
File Path
    │
    ▼
┌─────────────────────┐
│ 1. EXIF Extraction  │  Extract camera, lens, exposure metadata
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 2. Image Decode     │  Load image into memory
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 3. Thumbnail Gen    │  Create 4 sizes (64, 256, 512, 1024px)
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 4. Color Palette    │  K-means clustering (5 colors)
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 5. Perceptual Hash  │  pHash for similarity
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 6. File Hash        │  SHA-256 for integrity
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 7. Infer Metadata   │  Time of day, season, conditions
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 8. Database Insert  │  Transactional write
└─────────────────────┘
```

### 3. Query Engine (Future)
**Purpose**: Execute faceted searches and return result sets

**Capabilities**:
- Multi-dimensional filtering (camera, date, color, location)
- Facet computation (counts per attribute value)
- Pagination and sorting
- Burst filtering
- Full-text search (future: FTS5)

### 4. Repository / URL Mapper (Future)
**Purpose**: Map URL patterns to query parameters

**URL Patterns**:
```
/2025/10           → Photos from October 2025
/camera/Canon      → Photos by camera make
/color/blue        → Photos with dominant blue
/bursts            → All burst groups
```

### 5. Database Layer (Implemented)
**Purpose**: Persistent storage with efficient querying

**Schema Design**:
- **Normalized tables**: Separate tables for photos, thumbnails, colors
- **Comprehensive indexes**: All searchable fields indexed
- **Foreign key constraints**: Data integrity enforcement
- **Transaction support**: Atomic operations

## Data Flow

### Indexing Data Flow

```
┌──────────────┐
│  Photo File  │
│ (.dng/.jpg/  │
│   .bmp)      │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────────┐
│            Indexer Engine                        │
│                                                  │
│  ┌─────────────┐                                │
│  │ File exists?├─ Yes → Skip                    │
│  └─────┬───────┘                                │
│        │ No                                      │
│        ▼                                         │
│  ┌─────────────┐    ┌──────────────┐           │
│  │Extract EXIF │───→│ PhotoMetadata│           │
│  └─────────────┘    │  (50+ fields)│           │
│                      └──────┬───────┘           │
│                             │                    │
│  ┌─────────────┐           │                    │
│  │ Decode Image│←──────────┘                    │
│  └─────┬───────┘                                │
│        │                                         │
│        ├──→ Generate Thumbnails (4 sizes)       │
│        ├──→ Extract Colors (k-means)            │
│        ├──→ Compute pHash                       │
│        └──→ Calculate SHA-256                   │
│                                                  │
│  ┌─────────────┐                                │
│  │   Infer     │                                │
│  │  Metadata   │                                │
│  └─────┬───────┘                                │
│        │                                         │
│        ▼                                         │
│  ┌─────────────────────────────┐               │
│  │ Complete PhotoMetadata      │               │
│  │ + Thumbnails                │               │
│  │ + Colors                    │               │
│  │ + Hashes                    │               │
│  └─────┬───────────────────────┘               │
└────────┼─────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│     SQLite Transaction              │
│                                     │
│  INSERT INTO photos (...)           │
│  INSERT INTO thumbnails (4 rows)    │
│  INSERT INTO photo_colors (5 rows)  │
│                                     │
│  COMMIT                             │
└─────────────────────────────────────┘
```

### Query Data Flow (Future)

```
User Query
    │
    ▼
┌──────────────┐
│  Repository  │  Parse URL → QueryParams
└──────┬───────┘
       │
       ▼
┌──────────────┐
│Query Engine  │  Build SQL → Execute
└──────┬───────┘
       │
       ├──→ Fetch photos with filters
       ├──→ Compute facets (counts)
       └──→ Load thumbnails
       │
       ▼
┌──────────────┐
│Query Result  │  Photos + Facets + Breadcrumbs
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ CLI Display  │  Format and print
└──────────────┘
```

## Concurrency Model

### Worker Pool Pattern

```
                Main Thread
                     │
                     ▼
              ┌──────────────┐
              │ Find DNG/JPG │
              │ /BMP Files   │
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────┐
              │ Create Work  │
              │   Channel    │
              └──────┬───────┘
                     │
         ┌───────────┼───────────┬──────────┐
         │           │           │          │
         ▼           ▼           ▼          ▼
    ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐
    │Worker 1│  │Worker 2│  │Worker 3│  │Worker N│
    └───┬────┘  └───┬────┘  └───┬────┘  └───┬────┘
        │           │           │           │
        └───────────┴───────────┴───────────┘
                     │
              Shared Database
              (with locks)
```

**Concurrency Features**:
- Configurable worker count (default: 4)
- Buffered work channel (size: 100)
- Mutex-protected statistics
- Thread-safe database operations
- Graceful shutdown with WaitGroup

## Storage Architecture

### Database Portability

**Design Principle**: The SQLite database IS the catalog.

```
┌─────────────────────────────────────────┐
│        photos.db (SQLite)               │
├─────────────────────────────────────────┤
│                                         │
│  Metadata         ✓ Included           │
│  Thumbnails       ✓ Included (BLOBs)   │
│  Color Palettes   ✓ Included           │
│  Perceptual Hash  ✓ Included           │
│  Burst Info       ✓ Included           │
│                                         │
│  Original Files   ✗ NOT Included       │
│  (Too large)                            │
└─────────────────────────────────────────┘
```

**Benefits**:
- Single file backup/restore
- Easy database migration
- Offline browsing capability
- No external dependencies

**Storage Estimates**:
- Per photo: ~187 KB (thumbnails + metadata)
- 100K photos: ~18.7 GB database size
- Original DNG files: Stored separately

## Error Handling Strategy

### Graceful Degradation

```
┌─────────────────┐
│  Process File   │
└────────┬────────┘
         │
    ┌────▼────┐
    │ Success?│
    └────┬────┘
         │
    ┌────┴────┐
    │   Yes   │   No
    ▼         ▼
┌────────┐  ┌──────────────┐
│Counter │  │ Log Error    │
│  ++    │  │ Counter++    │
└────────┘  │ Continue     │
            └──────────────┘
```

**Error Handling Rules**:
1. Log errors, don't stop indexing
2. Track failed file count
3. Continue processing remaining files
4. Report failures in summary
5. No partial database writes (transactions)

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Find files | O(n) | Filesystem walk |
| Process file | O(1) | Constant per file |
| EXIF extraction | O(1) | Fixed metadata size |
| Thumbnail generation | O(w×h) | Image pixels |
| Color extraction | O(k×i×p) | k colors, i iterations, p pixels |
| pHash computation | O(1) | Fixed 32×32 |
| Database insert | O(1) | Indexed writes |

### Space Complexity

| Component | Size per Photo |
|-----------|---------------|
| Metadata row | ~2 KB |
| Thumbnails (4) | ~187 KB |
| Colors (5) | ~0.5 KB |
| Total | ~190 KB |

## Future Enhancements

### Planned Components

1. **Burst Detection Engine**
   - Temporal proximity analysis
   - Camera matching
   - Representative selection

2. **Query Engine**
   - Faceted search implementation
   - Pagination support
   - Sort/filter combinations

3. **CLI Interface**
   - Command parsing
   - Progress bars
   - Formatted output

4. **Web API** (Phase 2)
   - REST endpoints
   - JSON responses
   - Thumbnail serving

## Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.25+ | Performance, concurrency |
| Database | SQLite 3 | Embedded, portable |
| Image Processing | Go stdlib + nfnt/resize | Decoding, resizing |
| Color Extraction | palettor | K-means clustering |
| Perceptual Hash | goimagehash | pHash algorithm |
| EXIF Parsing | exif-go (dsoprea/go-exif/v3) | Metadata extraction |

## Design Patterns Used

1. **Worker Pool**: Concurrent file processing
2. **Repository**: Data access abstraction
3. **Builder**: PhotoMetadata construction
4. **Transaction Script**: Database operations
5. **Factory**: Thumbnail generation for multiple sizes
