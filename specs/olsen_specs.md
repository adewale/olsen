# Complete System Specification
## DNG Photo Indexer v2.0

**Version:** 2.0  
**Target Audience:** AI Coding Agents, Software Developers  
**Last Updated:** October 2025

---

## 1. System Architecture Overview

The system is composed of four distinct architectural layers:

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Interface                           │
│           (User-facing commands & output)                   │
└────────────────────────┬────────────────────────────────────┘
                         │
           ┌─────────────┴─────────────┐
           │                           │
    ┌──────▼──────┐            ┌──────▼──────┐
    │ Repository  │            │   Indexer   │
    │  (URL Map)  │            │   Engine    │
    └──────┬──────┘            └──────┬──────┘
           │                           │
           │      ┌────────────────────┘
           │      │
    ┌──────▼──────▼──────┐
    │   Query Engine      │
    │ (Faceted Search)    │
    └──────┬──────────────┘
           │
    ┌──────▼──────────────────────────────────┐
    │         SQLite Database                  │
    │  ├─ Metadata Catalog                    │
    │  ├─ Multiple Thumbnail Sizes            │
    │  ├─ Color Palettes                      │
    │  ├─ Perceptual Hashes                   │
    │  └─ Burst/Cluster Metadata              │
    └─────────────────────────────────────────┘
```

### 1.1 Component Responsibilities

**Indexer Engine:**
- Recursively scans directories for DNG files
- Extracts comprehensive EXIF metadata
- Generates multiple thumbnail sizes (64×64, 256×256, 512×512, 1024×1024)
- Extracts color palettes (5 dominant colors)
- Computes perceptual hashes for similarity detection
- Stores everything in SQLite database
- Runs post-indexing analysis (bursts, duplicates)

**Query Engine:**
- Understands faceted search semantics
- Executes multi-dimensional filters
- Computes facet counts with active filters
- Handles pagination efficiently
- Returns structured result sets

**Repository (URL Mapper):**
- Maps URL patterns to query parameters
- Implements RESTful routing
- Converts hierarchical paths to queries
- Generates breadcrumb trails
- Serializes queries back to URLs

**CLI Interface:**
- Exposes all functionality via command-line
- Provides indexing commands
- Executes queries and displays results
- Outputs thumbnails and metadata
- Formats results for human consumption

### 1.2 Key Design Principle: Database Portability

**Critical:** The SQLite database IS the catalog. Moving the database file moves the entire corpus.

**What's stored in the database:**
- ✅ All metadata extracted from images
- ✅ Multiple thumbnail sizes (64×64 through 1024×1024)
- ✅ Color palettes
- ✅ Perceptual hashes
- ✅ Burst and cluster information
- ✅ All indexes and facet metadata

**What's NOT in the database:**
- ❌ Original DNG files (too large)
- ❌ Full-resolution images

**Implications:**
- Database contains everything needed for browsing/searching
- Original files only needed when user wants full-resolution export
- Database can be copied/backed up as single file
- User can browse catalog even if originals are offline

---

## 2. Data Model

### 2.1 Core Data Structures

#### PhotoMetadata
```go
type PhotoMetadata struct {
    ID       int
    FilePath string    // Path to original (for full-res access)
    FileHash string    // SHA-256 of original file
    FileSize int64
    LastModified time.Time
    IndexedAt time.Time

    // Camera & Lens
    CameraMake  string
    CameraModel string
    LensMake    string
    LensModel   string

    // Exposure Settings
    ISO                  int
    Aperture             float64
    ShutterSpeed         string
    ExposureCompensation float64
    FocalLength          float64
    FocalLength35mm      int

    // Temporal
    DateTaken     time.Time
    DateDigitized time.Time

    // Image Properties
    Width       int
    Height      int
    Orientation int
    ColorSpace  string

    // Location
    Latitude  float64
    Longitude float64
    Altitude  float64

    // DNG-Specific
    DNGVersion          string
    OriginalRawFilename string

    // Lighting
    FlashFired    bool
    WhiteBalance  string
    FocusDistance float64

    // Inferred Metadata
    TimeOfDay         string
    Season            string
    FocalCategory     string
    ShootingCondition string

    // Visual Analysis
    Thumbnails     map[ThumbnailSize][]byte  // Multiple sizes
    DominantColors []Color
    ColorWeights   []float64
    
    // Perceptual Hash & Similarity
    PerceptualHash string  // 16-char hex (64-bit pHash)
    
    // Burst Detection
    BurstGroupID          string
    BurstSequence         int
    BurstCount            int
    IsBurstRepresentative bool
    
    // Duplicate Clustering
    DuplicateClusterID      string
    ClusterSize             int
    IsClusterRepresentative bool
    SimilarityScore         float64
}

type ThumbnailSize string
const (
    ThumbnailTiny   ThumbnailSize = "64"    // Grid view (longest edge)
    ThumbnailSmall  ThumbnailSize = "256"  // List view (longest edge)
    ThumbnailMedium ThumbnailSize = "512"  // Preview (longest edge)
    ThumbnailLarge  ThumbnailSize = "1024" // Large preview (longest edge)
)

type Color struct {
    R, G, B uint8
}
```

### 2.2 Complete Database Schema

```sql
-- ============================================================
-- PHOTOS TABLE (Core metadata)
-- ============================================================
CREATE TABLE photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT UNIQUE NOT NULL,
    file_hash TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_modified DATETIME NOT NULL,
    
    -- Camera metadata
    camera_make TEXT,
    camera_model TEXT,
    lens_make TEXT,
    lens_model TEXT,
    
    -- Exposure metadata
    iso INTEGER,
    aperture REAL,
    shutter_speed TEXT,
    exposure_compensation REAL,
    focal_length REAL,
    focal_length_35mm INTEGER,
    
    -- Temporal metadata
    date_taken DATETIME,
    date_digitized DATETIME,
    
    -- Image properties
    width INTEGER,
    height INTEGER,
    orientation INTEGER,
    color_space TEXT,
    
    -- Location metadata
    latitude REAL,
    longitude REAL,
    altitude REAL,
    
    -- DNG-specific
    dng_version TEXT,
    original_raw_filename TEXT,
    
    -- Lighting metadata
    flash_fired BOOLEAN,
    white_balance TEXT,
    focus_distance REAL,
    
    -- Inferred metadata
    time_of_day TEXT,
    season TEXT,
    focal_category TEXT,
    shooting_condition TEXT,
    
    -- Perceptual hash
    perceptual_hash TEXT,
    
    -- Burst metadata
    burst_group_id TEXT,
    burst_sequence INTEGER,
    burst_count INTEGER,
    is_burst_representative BOOLEAN DEFAULT FALSE,
    
    -- Duplicate cluster metadata
    duplicate_cluster_id TEXT,
    cluster_size INTEGER,
    is_cluster_representative BOOLEAN DEFAULT FALSE,
    similarity_score REAL
);

-- ============================================================
-- THUMBNAILS TABLE (Multiple sizes stored)
-- ============================================================
CREATE TABLE thumbnails (
    photo_id INTEGER NOT NULL,
    size TEXT NOT NULL,  -- "64", "256", "512", "1024" (longest edge)
    data BLOB NOT NULL,
    format TEXT DEFAULT 'jpeg',  -- "jpeg" or "webp"
    quality INTEGER DEFAULT 85,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (photo_id, size),
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE
);

-- ============================================================
-- PHOTO COLORS TABLE (Dominant color palette)
-- ============================================================
CREATE TABLE photo_colors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    photo_id INTEGER NOT NULL,
    color_order INTEGER NOT NULL,
    red INTEGER NOT NULL,
    green INTEGER NOT NULL,
    blue INTEGER NOT NULL,
    weight REAL NOT NULL,
    hue INTEGER,
    saturation INTEGER,
    lightness INTEGER,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    UNIQUE(photo_id, color_order)
);

-- ============================================================
-- BURST GROUPS TABLE
-- ============================================================
CREATE TABLE burst_groups (
    id TEXT PRIMARY KEY,
    photo_count INTEGER NOT NULL,
    date_taken DATETIME,
    camera_make TEXT,
    camera_model TEXT,
    representative_photo_id INTEGER,
    time_span_seconds REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (representative_photo_id) REFERENCES photos(id)
);

-- ============================================================
-- DUPLICATE CLUSTERS TABLE
-- ============================================================
CREATE TABLE duplicate_clusters (
    id TEXT PRIMARY KEY,
    photo_count INTEGER NOT NULL,
    max_hamming_distance INTEGER,
    representative_photo_id INTEGER,
    cluster_type TEXT CHECK(cluster_type IN ('exact', 'near', 'similar')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (representative_photo_id) REFERENCES photos(id)
);

-- ============================================================
-- TAGS TABLE (User-defined)
-- ============================================================
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE photo_tags (
    photo_id INTEGER,
    tag_id INTEGER,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (photo_id, tag_id)
);

-- ============================================================
-- COLLECTIONS TABLE (Virtual collections)
-- ============================================================
CREATE TABLE collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT CHECK(type IN ('manual', 'smart')) DEFAULT 'manual',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE collection_photos (
    collection_id INTEGER,
    photo_id INTEGER,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, photo_id)
);

-- ============================================================
-- FACET METADATA (For display configuration)
-- ============================================================
CREATE TABLE facet_metadata (
    facet_type TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    facet_order INTEGER,
    allow_multiple BOOLEAN DEFAULT FALSE,
    hierarchical BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE
);

INSERT INTO facet_metadata VALUES
('camera_make', 'Camera', 1, FALSE, TRUE, TRUE),
('lens_model', 'Lens', 2, FALSE, FALSE, TRUE),
('time_of_day', 'Time of Day', 3, TRUE, FALSE, TRUE),
('season', 'Season', 4, TRUE, FALSE, TRUE),
('color', 'Dominant Color', 5, TRUE, FALSE, TRUE),
('iso', 'ISO', 6, FALSE, FALSE, TRUE),
('aperture', 'Aperture', 7, FALSE, FALSE, TRUE),
('focal_category', 'Focal Length', 8, TRUE, FALSE, TRUE),
('burst_group', 'Bursts', 9, FALSE, FALSE, TRUE),
('duplicate_cluster', 'Duplicates', 10, FALSE, FALSE, TRUE);

-- ============================================================
-- PERFORMANCE INDEXES
-- ============================================================
-- Core queries
CREATE INDEX idx_photos_date_taken ON photos(date_taken);
CREATE INDEX idx_photos_camera ON photos(camera_make, camera_model);
CREATE INDEX idx_photos_lens ON photos(lens_model);
CREATE INDEX idx_photos_gps ON photos(latitude, longitude);
CREATE INDEX idx_photos_hash ON photos(file_hash);
CREATE INDEX idx_photos_phash ON photos(perceptual_hash);

-- Faceted browsing
CREATE INDEX idx_photos_iso ON photos(iso);
CREATE INDEX idx_photos_aperture ON photos(aperture);
CREATE INDEX idx_photos_focal_length ON photos(focal_length);
CREATE INDEX idx_photos_time_of_day ON photos(time_of_day);
CREATE INDEX idx_photos_season ON photos(season);
CREATE INDEX idx_photos_focal_category ON photos(focal_category);
CREATE INDEX idx_photos_shooting_condition ON photos(shooting_condition);

-- Burst and cluster queries
CREATE INDEX idx_photos_burst ON photos(burst_group_id);
CREATE INDEX idx_photos_cluster ON photos(duplicate_cluster_id);
CREATE INDEX idx_burst_groups_date ON burst_groups(date_taken);
CREATE INDEX idx_duplicate_clusters_type ON duplicate_clusters(cluster_type);

-- Color search
CREATE INDEX idx_colors_hue ON photo_colors(hue);
CREATE INDEX idx_colors_saturation ON photo_colors(saturation);
CREATE INDEX idx_colors_lightness ON photo_colors(lightness);
CREATE INDEX idx_colors_rgb ON photo_colors(red, green, blue);
CREATE INDEX idx_colors_photo ON photo_colors(photo_id, color_order);
```

---

## 3. Module Specifications

### 3.1 Indexer Engine

**Responsibility:** Crawl filesystem, extract metadata, populate database

#### Core Components

```go
type IndexerEngine struct {
    db          *sql.DB
    workerCount int
    stats       IndexStats
}

type IndexStats struct {
    FilesFound     int
    FilesProcessed int
    FilesFailed    int
    ThumbnailsGenerated int
    HashesComputed int
    StartTime      time.Time
}

// Main indexing flow
func (ie *IndexerEngine) IndexDirectory(rootPath string) error {
    // 1. Scan filesystem for DNG files
    // 2. Process each file concurrently
    // 3. Extract metadata + generate thumbnails + compute hash
    // 4. Store in database
    // 5. Report progress
}

// Process single file
func (ie *IndexerEngine) processFile(filePath string) error {
    // 1. Extract EXIF metadata
    // 2. Generate thumbnails (all 4 sizes)
    // 3. Extract color palette from 256×256 thumbnail
    // 4. Compute perceptual hash from 256×256 thumbnail
    // 5. Calculate SHA-256 file hash
    // 6. Infer metadata (time of day, season, etc.)
    // 7. Store in database (transactional)
}
```

#### Thumbnail Generation Strategy

```go
func generateThumbnails(img image.Image) map[ThumbnailSize][]byte {
    thumbnails := make(map[ThumbnailSize][]byte)

    sizes := []struct {
        name ThumbnailSize
        maxDimension uint
    }{
        {ThumbnailTiny, 64},
        {ThumbnailSmall, 256},
        {ThumbnailMedium, 512},
        {ThumbnailLarge, 1024},
    }

    for _, size := range sizes {
        // Preserve aspect ratio by constraining longest edge
        bounds := img.Bounds()
        width := uint(bounds.Dx())
        height := uint(bounds.Dy())

        var newWidth, newHeight uint
        if width > height {
            newWidth = size.maxDimension
            newHeight = 0 // resize library will calculate
        } else {
            newWidth = 0 // resize library will calculate
            newHeight = size.maxDimension
        }

        thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
        var buf bytes.Buffer
        jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85})
        thumbnails[size.name] = buf.Bytes()
    }

    return thumbnails
}
```

**Rationale for multiple sizes:**
- 64px: Ultra-fast grid views, minimal bandwidth
- 256px: Standard thumbnails, good for lists
- 512px: Preview pane, quick inspection
- 1024px: Large preview, almost full-screen quality

**Note:** All sizes represent the longest edge dimension, preserving aspect ratio.

#### Post-Indexing Analysis

```go
func (ie *IndexerEngine) RunAnalysis() error {
    // 1. Detect burst sequences
    bd := &BurstDetector{db: ie.db}
    bd.DetectBursts()
    
    // 2. Cluster near-duplicates
    dc := &DuplicateClusterer{db: ie.db, threshold: 15}
    dc.ClusterDuplicates()
    
    // 3. Update statistics
    ie.updateDatabaseStats()
}
```

---

### 3.2 Query Engine

**Responsibility:** Execute faceted searches, return result sets

#### Query Parameters Structure

```go
type QueryParams struct {
    // Temporal filters
    Year      int
    Month     int
    Day       int
    DateFrom  time.Time
    DateTo    time.Time
    
    // Equipment filters
    CameraMake  string
    CameraModel string
    LensModel   string
    
    // Technical filters
    ISOMin         int
    ISOMax         int
    ApertureMin    float64
    ApertureMax    float64
    FocalLengthMin float64
    FocalLengthMax float64
    
    // Categorical filters (multi-select)
    TimeOfDay         []string
    Season            []string
    FocalCategory     []string
    ShootingCondition []string
    
    // Color filters
    ColorName         string  // "red", "blue", etc.
    HueMin            int     // 0-360
    HueMax            int     // 0-360
    
    // Location filters
    LatMin, LatMax    float64
    LonMin, LonMax    float64
    
    // Burst/Cluster filters
    BurstGroupID      string
    DuplicateClusterID string
    ClusterType       string  // "exact", "near", "similar"
    OnlyRepresentatives bool  // Show only burst/cluster reps
    
    // Pagination
    Offset int
    Limit  int
    
    // Sort
    OrderBy   string  // "date_taken", "camera_make", etc.
    Direction string  // "ASC", "DESC"
    
    // Output control
    ThumbnailSize ThumbnailSize
    IncludeFacets bool
}

type QueryResult struct {
    Photos      []PhotoSummary
    TotalCount  int
    Facets      FacetCollection
    Breadcrumbs []Breadcrumb
    Query       QueryParams
}

type PhotoSummary struct {
    ID            int
    ThumbnailData []byte
    DateTaken     time.Time
    CameraMake    string
    CameraModel   string
    DominantColor Color
    BurstInfo     *BurstInfo      // nil if not part of burst
    ClusterInfo   *ClusterInfo    // nil if not part of cluster
}

type BurstInfo struct {
    GroupID       string
    Sequence      int
    TotalCount    int
    IsRepresentative bool
}

type ClusterInfo struct {
    ClusterID     string
    Size          int
    Type          string
    IsRepresentative bool
    SimilarityScore float64
}
```

#### Query Execution

```go
type QueryEngine struct {
    db *sql.DB
}

func (qe *QueryEngine) Execute(params QueryParams) (*QueryResult, error) {
    // 1. Build SQL query from parameters
    query, args := qe.buildQuery(params)
    
    // 2. Execute query
    rows, err := qe.db.Query(query, args...)
    
    // 3. Load photos with thumbnails
    photos := qe.loadPhotos(rows, params.ThumbnailSize)
    
    // 4. Get total count
    totalCount := qe.getCount(params)
    
    // 5. Compute facets (if requested)
    var facets FacetCollection
    if params.IncludeFacets {
        facets = qe.computeFacets(params)
    }
    
    // 6. Build breadcrumbs
    breadcrumbs := qe.buildBreadcrumbs(params)
    
    return &QueryResult{
        Photos:      photos,
        TotalCount:  totalCount,
        Facets:      facets,
        Breadcrumbs: breadcrumbs,
        Query:       params,
    }, nil
}
```

#### Facet Computation

```go
type FacetCollection struct {
    Camera      []Facet
    Lens        []Facet
    TimeOfDay   []Facet
    Season      []Facet
    Color       []ColorFacet
    Year        []Facet
    Month       []Facet
    Burst       []BurstFacet
    Duplicates  []DuplicateFacet
}

type Facet struct {
    Value    string
    Count    int
    Selected bool
}

type ColorFacet struct {
    Name  string
    RGB   Color
    Count int
}

type BurstFacet struct {
    GroupID    string
    Count      int
    DateTaken  time.Time
    TimeSpan   float64
}

type DuplicateFacet struct {
    ClusterID   string
    Count       int
    Type        string
    MaxDistance int
}

func (qe *QueryEngine) computeFacets(params QueryParams) FacetCollection {
    // For each facet type, count photos matching each value
    // WHILE respecting currently active filters
    
    // Example: Camera facets
    cameraFacets := qe.computeCameraFacets(params)
    
    // Example: Color facets
    colorFacets := qe.computeColorFacets(params)
    
    // Example: Burst facets
    burstFacets := qe.computeBurstFacets(params)
    
    // Return all facets
}
```

---

### 3.3 Repository (URL Mapper)

**Responsibility:** Map URLs to queries, generate URLs from queries

#### URL Patterns

```
# Temporal browsing
/YYYY                → Year view
/YYYY/MM             → Month view
/YYYY/MM/DD          → Day view

# Equipment browsing
/camera/:make        → By camera make
/camera/:make/:model → By specific camera
/lens/:model         → By lens

# Color browsing
/color/:name         → By color name (red, blue, etc.)
/color/hue/:degrees  → By hue range

# Burst browsing
/bursts              → All burst groups
/bursts/:id          → Specific burst

# Duplicate browsing
/duplicates          → All clusters
/duplicates/exact    → Only exact duplicates
/duplicates/near     → Only near duplicates
/duplicates/:id      → Specific cluster

# Faceted queries (query string parameters)
/?camera=Canon&lens=24-70mm&iso=100-400&tod=golden_hour_morning
```

#### Repository Implementation

```go
type Repository struct {
    queryEngine *QueryEngine
    urlPatterns map[string]URLHandler
}

type URLHandler func(path string, query url.Values) (*QueryResult, error)

func NewRepository(qe *QueryEngine) *Repository {
    repo := &Repository{
        queryEngine: qe,
        urlPatterns: make(map[string]URLHandler),
    }
    
    // Register URL patterns
    repo.register(`^/(\d{4})$`, repo.handleYear)
    repo.register(`^/(\d{4})/(\d{2})$`, repo.handleMonth)
    repo.register(`^/(\d{4})/(\d{2})/(\d{2})$`, repo.handleDay)
    repo.register(`^/camera/([^/]+)$`, repo.handleCameraMake)
    repo.register(`^/camera/([^/]+)/([^/]+)$`, repo.handleCameraModel)
    repo.register(`^/color/([^/]+)$`, repo.handleColorName)
    repo.register(`^/bursts/([^/]+)$`, repo.handleBurst)
    repo.register(`^/duplicates/([^/]+)$`, repo.handleDuplicateCluster)
    
    return repo
}

func (r *Repository) HandleURL(urlPath string, queryString url.Values) (*QueryResult, error) {
    // Match URL pattern and dispatch to handler
    for pattern, handler := range r.urlPatterns {
        if matches(pattern, urlPath) {
            return handler(urlPath, queryString)
        }
    }
    
    return nil, fmt.Errorf("unknown URL pattern: %s", urlPath)
}

// Example handler
func (r *Repository) handleMonth(path string, query url.Values) (*QueryResult, error) {
    // Extract year and month from path: /2025/10
    year, month := parseYearMonth(path)
    
    // Build query params
    params := QueryParams{
        Year:          year,
        Month:         month,
        IncludeFacets: true,
        ThumbnailSize: ThumbnailSmall,
        Limit:         100,
    }
    
    // Add query string filters
    params = r.applyQueryString(params, query)
    
    // Execute query
    return r.queryEngine.Execute(params)
}

// Generate URL from query params
func (r *Repository) GenerateURL(params QueryParams) string {
    // Build URL path from primary filter
    path := r.buildPath(params)
    
    // Add query string for additional filters
    queryString := r.buildQueryString(params)
    
    if queryString != "" {
        return path + "?" + queryString
    }
    return path
}
```

---

### 3.4 CLI Interface

**Responsibility:** User-facing commands, format output

#### Command Structure

```bash
# Indexing commands
indexer index <path>              # Index directory
indexer index <path> -w 8         # Use 8 workers
indexer reindex <path>            # Re-index changed files
indexer analyze                   # Run burst/duplicate analysis

# Query commands
indexer query "/2025/10"          # Query by URL
indexer query -y 2025 -m 10       # Query by parameters
indexer query --camera Canon      # Filter by camera
indexer query --bursts            # Show all bursts
indexer query --duplicates exact  # Show exact duplicates

# Output commands
indexer show <photo-id>           # Show photo details
indexer thumbnail <photo-id> -s large # Export thumbnail
indexer export <query> -o /path   # Export thumbnails

# Statistics
indexer stats                     # Database statistics
indexer stats --bursts            # Burst statistics
indexer stats --duplicates        # Duplicate statistics

# Maintenance
indexer compact                   # Vacuum database
indexer verify                    # Verify integrity
```

#### CLI Implementation

```go
type CLI struct {
    repo   *Repository
    engine *IndexerEngine
}

func (cli *CLI) Run(args []string) error {
    cmd := args[0]
    
    switch cmd {
    case "index":
        return cli.runIndex(args[1:])
    case "query":
        return cli.runQuery(args[1:])
    case "show":
        return cli.runShow(args[1:])
    case "thumbnail":
        return cli.runThumbnail(args[1:])
    case "stats":
        return cli.runStats(args[1:])
    default:
        return fmt.Errorf("unknown command: %s", cmd)
    }
}

func (cli *CLI) runQuery(args []string) error {
    // Parse arguments into QueryParams or URL
    params := cli.parseQueryArgs(args)
    
    // Execute query
    result, err := cli.repo.queryEngine.Execute(params)
    if err != nil {
        return err
    }
    
    // Format output
    cli.printResults(result)
    
    return nil
}

func (cli *CLI) printResults(result *QueryResult) {
    // Print summary
    fmt.Printf("Found %d photos\n\n", result.TotalCount)
    
    // Print photos
    for i, photo := range result.Photos {
        fmt.Printf("%d. %s\n", i+1, photo.DateTaken.Format("2006-01-02 15:04"))
        fmt.Printf("   Camera: %s %s\n", photo.CameraMake, photo.CameraModel)
        
        if photo.BurstInfo != nil {
            fmt.Printf("   Burst: %d/%d\n", photo.BurstInfo.Sequence, photo.BurstInfo.TotalCount)
        }
        
        if photo.ClusterInfo != nil {
            fmt.Printf("   Cluster: %s (%d photos, %.1f%% similar)\n",
                photo.ClusterInfo.Type,
                photo.ClusterInfo.Size,
                photo.ClusterInfo.SimilarityScore*100)
        }
        
        fmt.Println()
    }
    
    // Print facets
    if result.Facets != nil {
        cli.printFacets(result.Facets)
    }
}

func (cli *CLI) printFacets(facets FacetCollection) {
    fmt.Println("Available Facets:")
    fmt.Println()
    
    if len(facets.Camera) > 0 {
        fmt.Println("Cameras:")
        for _, f := range facets.Camera {
            fmt.Printf("  %s (%d)\n", f.Value, f.Count)
        }
        fmt.Println()
    }
    
    if len(facets.Burst) > 0 {
        fmt.Println("Bursts:")
        for _, f := range facets.Burst {
            fmt.Printf("  %s: %d photos (%.1fs span)\n",
                f.DateTaken.Format("2006-01-02 15:04"),
                f.Count,
                f.TimeSpan)
        }
        fmt.Println()
    }
    
    // ... other facets
}
```

---

## 4. Implementation Details

### 4.1 Thumbnail Storage Strategy

**Decision: Store all thumbnails in database**

**Rationale:**
- Database IS the catalog - must be self-contained
- Thumbnails are small enough (even 1024×1024 ≈ 150KB)
- Simplifies backup/restore
- Enables offline browsing
- Transaction integrity guaranteed

**Storage Calculation:**
```
Per photo (assuming 3:2 aspect ratio):
- 64px:    ~2 KB
- 256px:   ~15 KB
- 512px:   ~50 KB
- 1024px:  ~120 KB
Total:     ~187 KB per photo

For 100K photos: 18.7 GB thumbnail storage

Note: Actual sizes vary by aspect ratio and content complexity.
Aspect-ratio-preserving thumbnails save ~20% storage vs forced square crops.
```

**Mitigation for size:**
- Use JPEG quality 85 (good balance)
- Consider WebP for 30% smaller files (future optimization)
- Vacuum database periodically

### 4.2 Perceptual Hash Algorithm

**Algorithm: pHash (Discrete Cosine Transform)**

**Process:**
1. Resize to 32×32 grayscale
2. Apply DCT (Discrete Cosine Transform)
3. Extract low-frequency components (8×8)
4. Compute median
5. Set bits above median to 1, below to 0
6. Result: 64-bit hash (stored as 16-char hex string)

**Implementation:**
```go
import "github.com/corona10/goimagehash"

func computePerceptualHash(img image.Image) (string, error) {
    hash, err := goimagehash.PerceptionHash(img)
    if err != nil {
        return "", err
    }
    return hash.ToString(), nil
}

func hammingDistance(hash1, hash2 string) (int, error) {
    h1, err := goimagehash.ImageHashFromString(hash1)
    if err != nil {
        return 0, err
    }
    h2, err := goimagehash.ImageHashFromString(hash2)
    if err != nil {
        return 0, err
    }
    distance, err := h1.Distance(h2)
    return distance, err
}
```

**Distance Thresholds:**
- 0-5: Identical/near-identical (same photo, minor compression differences)
- 6-10: Very similar (minor edits, crops)
- 11-15: Similar (burst variations, different exposures)
- 16-20: Somewhat similar (same scene, different angles)
- 21+: Different images

### 4.3 Burst Detection Algorithm

**Strategy:** Temporal proximity + camera matching

**Algorithm:**
```go
func detectBursts(photos []*PhotoMetadata) []BurstGroup {
    // Sort by camera + date
    sort.Slice(photos, func(i, j int) bool {
        if photos[i].CameraMake != photos[j].CameraMake {
            return photos[i].CameraMake < photos[j].CameraMake
        }
        if photos[i].CameraModel != photos[j].CameraModel {
            return photos[i].CameraModel < photos[j].CameraModel
        }
        return photos[i].DateTaken.Before(photos[j].DateTaken)
    })
    
    var bursts []BurstGroup
    var currentBurst []*PhotoMetadata
    
    for i, photo := range photos {
        if i == 0 {
            currentBurst = []*PhotoMetadata{photo}
            continue
        }
        
        prev := photos[i-1]
        
        // Check burst criteria
        sameCamera := prev.CameraMake == photo.CameraMake && 
                     prev.CameraModel == photo.CameraModel
        
        timeDiff := photo.DateTaken.Sub(prev.DateTaken)
        within2Seconds := timeDiff >= 0 && timeDiff <= 2*time.Second
        
        similarFocal := math.Abs(prev.FocalLength - photo.FocalLength) <= 5
        
        if sameCamera && within2Seconds && similarFocal {
            currentBurst = append(currentBurst, photo)
        } else {
            // End current burst
            if len(currentBurst) >= 3 {
                bursts = append(bursts, createBurstGroup(currentBurst))
            }
            currentBurst = []*PhotoMetadata{photo}
        }
    }
    
    // Handle final burst
    if len(currentBurst) >= 3 {
        bursts = append(bursts, createBurstGroup(currentBurst))
    }
    
    return bursts
}

func createBurstGroup(photos []*PhotoMetadata) BurstGroup {
    groupID := uuid.New().String()
    representative := selectRepresentative(photos)
    
    timeSpan := photos[len(photos)-1].DateTaken.Sub(photos[0].DateTaken).Seconds()
    
    return BurstGroup{
        ID:                    groupID,
        Photos:                photos,
        RepresentativePhotoID: representative.ID,
        DateTaken:            photos[0].DateTaken,
        TimeSpanSeconds:      timeSpan,
    }
}

func selectRepresentative(photos []*PhotoMetadata) *PhotoMetadata {
    // Strategy: Select middle photo (most stable)
    // Alternative: Select sharpest (requires sharpness scoring)
    return photos[len(photos)/2]
}
```

### 4.4 Duplicate Clustering Algorithm

**Strategy:** DBSCAN-like clustering with Hamming distance

**Algorithm:**
```go
func clusterDuplicates(photos []*PhotoMetadata, threshold int) []DuplicateCluster {
    visited := make(map[int]bool)
    clusters := []DuplicateCluster{}
    
    // Build spatial index for fast similarity search
    index := buildSimilarityIndex(photos)
    
    for _, photo := range photos {
        if visited[photo.ID] {
            continue
        }
        
        // Find all similar photos (density-based clustering)
        cluster := expandCluster(photo, photos, index, threshold, visited)
        
        if len(cluster) >= 2 {
            clusters = append(clusters, createDuplicateCluster(cluster, threshold))
        }
    }
    
    return clusters
}

func expandCluster(
    seed *PhotoMetadata,
    allPhotos []*PhotoMetadata,
    index *SimilarityIndex,
    threshold int,
    visited map[int]bool,
) []*PhotoMetadata {
    cluster := []*PhotoMetadata{seed}
    visited[seed.ID] = true
    queue := []*PhotoMetadata{seed}
    
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        
        // Find neighbors within threshold
        neighbors := index.FindSimilar(current.PerceptualHash, threshold)
        
        for _, neighbor := range neighbors {
            if !visited[neighbor.ID] {
                visited[neighbor.ID] = true
                cluster = append(cluster, neighbor)
                queue = append(queue, neighbor)
            }
        }
    }
    
    return cluster
}

func createDuplicateCluster(photos []*PhotoMetadata, threshold int) DuplicateCluster {
    clusterID := uuid.New().String()
    representative := selectClusterRepresentative(photos)
    maxDistance := computeMaxDistance(photos)
    
    clusterType := "similar"
    if maxDistance <= 5 {
        clusterType = "exact"
    } else if maxDistance <= 10 {
        clusterType = "near"
    }
    
    return DuplicateCluster{
        ID:                    clusterID,
        Photos:                photos,
        RepresentativePhotoID: representative.ID,
        MaxHammingDistance:    maxDistance,
        ClusterType:           clusterType,
    }
}

func selectClusterRepresentative(photos []*PhotoMetadata) *PhotoMetadata {
    // Find photo with minimum average distance to all others
    // This is the "most central" photo in the cluster
    
    bestPhoto := photos[0]
    bestAvgDist := math.MaxFloat64
    
    for _, candidate := range photos {
        totalDist := 0
        for _, other := range photos {
            if candidate.ID != other.ID {
                dist, _ := hammingDistance(candidate.PerceptualHash, other.PerceptualHash)
                totalDist += dist
            }
        }
        avgDist := float64(totalDist) / float64(len(photos)-1)
        
        if avgDist < bestAvgDist {
            bestAvgDist = avgDist
            bestPhoto = candidate
        }
    }
    
    return bestPhoto
}
```

### 4.5 Similarity Index (BK-Tree Optimization)

**Purpose:** Fast similarity search in O(log n) instead of O(n)

**Implementation:**
```go
type BKNode struct {
    Hash     string
    PhotoID  int
    Children map[int]*BKNode  // key = Hamming distance
}

type BKTree struct {
    root *BKNode
}

func (tree *BKTree) Insert(hash string, photoID int) {
    if tree.root == nil {
        tree.root = &BKNode{
            Hash:     hash,
            PhotoID:  photoID,
            Children: make(map[int]*BKNode),
        }
        return
    }
    
    node := tree.root
    for {
        dist, _ := hammingDistance(hash, node.Hash)
        
        if child, ok := node.Children[dist]; ok {
            node = child
        } else {
            node.Children[dist] = &BKNode{
                Hash:     hash,
                PhotoID:  photoID,
                Children: make(map[int]*BKNode),
            }
            return
        }
    }
}

func (tree *BKTree) Search(hash string, threshold int) []int {
    if tree.root == nil {
        return nil
    }
    
    results := []int{}
    tree.searchRecursive(tree.root, hash, threshold, &results)
    return results
}

func (tree *BKTree) searchRecursive(node *BKNode, hash string, threshold int, results *[]int) {
    dist, _ := hammingDistance(hash, node.Hash)
    
    if dist <= threshold {
        *results = append(*results, node.PhotoID)
    }
    
    // Only explore children within threshold range
    minDist := dist - threshold
    maxDist := dist + threshold
    
    for childDist, child := range node.Children {
        if childDist >= minDist && childDist <= maxDist {
            tree.searchRecursive(child, hash, threshold, results)
        }
    }
}
```

---

## 5. Data Flow Examples

### 5.1 Indexing Flow

```
User: indexer index /photos -w 8

┌─────────────────────────────────────────┐
│ 1. Scan filesystem                      │
│    - Walk directory tree                │
│    - Filter for .dng files              │
│    - Return 1,247 files                 │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 2. Spawn 8 worker goroutines            │
│    - Create file channel (buffered)     │
│    - Distribute work                    │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 3. Process each file (parallel)         │
│    - Extract EXIF metadata              │
│    - Generate 4 thumbnail sizes         │
│    - Extract color palette              │
│    - Compute perceptual hash            │
│    - Calculate file hash                │
│    - Infer metadata                     │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 4. Store in database (transactional)    │
│    - INSERT photo record                │
│    - INSERT 4 thumbnail records         │
│    - INSERT 5 color records             │
│    - COMMIT transaction                 │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 5. Report progress                      │
│    - Every 100 photos: log progress     │
│    - Track success/failure stats        │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 6. Post-indexing analysis                │
│    - Detect bursts (temporal)           │
│    - Cluster duplicates (visual)        │
│    - Update statistics                  │
└────────────┬────────────────────────────┘
             │
             ▼
       Complete: 1,247 photos indexed
       - 34 burst groups detected
       - 18 duplicate clusters found
```

### 5.2 Query Flow

```
User: indexer query "/2025/10?camera=Canon"

┌─────────────────────────────────────────┐
│ 1. Repository: Parse URL                │
│    - Path: /2025/10                     │
│    - Query: camera=Canon                │
│    - Map to QueryParams                 │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 2. QueryParams created                   │
│    - Year: 2025                         │
│    - Month: 10                          │
│    - CameraMake: Canon                  │
│    - IncludeFacets: true                │
│    - ThumbnailSize: 256x256             │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 3. Query Engine: Build SQL              │
│    SELECT * FROM photos                 │
│    WHERE date_taken >= '2025-10-01'     │
│      AND date_taken < '2025-11-01'      │
│      AND camera_make = 'Canon'          │
│    ORDER BY date_taken DESC             │
│    LIMIT 100                            │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 4. Execute query                         │
│    - Fetch 73 matching photos           │
│    - Join with thumbnails table         │
│    - Join with photo_colors table       │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 5. Compute facets                        │
│    - Count by camera model (Canon only) │
│    - Count by lens                      │
│    - Count by time of day               │
│    - Count by color                     │
│    - Find burst groups                  │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 6. Build result                          │
│    - Photos: 73 PhotoSummary objects    │
│    - Facets: All computed facets        │
│    - Breadcrumbs: 2025 > October > Canon│
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 7. CLI: Format output                    │
│    - Print summary                      │
│    - Print photo list with thumbnails   │
│    - Print available facets             │
└─────────────────────────────────────────┘
```

### 5.3 Thumbnail Export Flow

```
User: indexer thumbnail 42 -s large -o /tmp/photo.jpg

┌─────────────────────────────────────────┐
│ 1. Parse command                         │
│    - Photo ID: 42                       │
│    - Size: large (1024x1024)            │
│    - Output: /tmp/photo.jpg             │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 2. Query database                        │
│    SELECT data FROM thumbnails          │
│    WHERE photo_id = 42                  │
│      AND size = '1024x1024'             │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 3. Retrieve BLOB                         │
│    - Read ~150KB JPEG data              │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│ 4. Write to file                         │
│    - os.WriteFile(/tmp/photo.jpg, data) │
└────────────┬────────────────────────────┘
             │
             ▼
       Thumbnail exported successfully
```

---

## 6. Performance Specifications

### 6.1 Indexing Performance

**Target:** 10-30 photos/second on modern hardware

**Breakdown per photo:**
- File I/O (read DNG): ~100ms
- EXIF extraction: ~50ms
- Image decode: ~200ms
- Thumbnail generation (4 sizes): ~500ms
- Color palette extraction: ~50ms
- Perceptual hash: ~50ms
- File hash (SHA-256): ~100ms
- Database insert: ~50ms
**Total: ~1,100ms per photo (sequential)**

**With 8 workers:** ~88 photos/second theoretical max
**Realistic with I/O contention:** 15-25 photos/second

### 6.2 Query Performance

**Targets:**
- Simple queries (single filter): < 100ms
- Complex queries (multiple filters): < 500ms
- Facet computation: < 500ms
- Thumbnail retrieval: < 50ms per thumbnail

**Optimizations:**
- All facetable fields indexed
- Prepared statements
- Connection pooling
- Query result caching (optional)

### 6.3 Analysis Performance

**Burst Detection:** O(n log n)
- 100K photos: ~30 seconds
- Scales well

**Duplicate Clustering:** O(n²) worst case, O(n log n) with BK-tree
- Without BK-tree: 100K photos = ~30 minutes
- With BK-tree: 100K photos = ~5 minutes
- Run as background batch job

---

## 7. Testing Strategy

### 7.1 Unit Tests

```go
// Indexer tests
func TestEXIFExtraction(t *testing.T)
func TestThumbnailGeneration(t *testing.T)
func TestColorPaletteExtraction(t *testing.T)
func TestPerceptualHashing(t *testing.T)
func TestInference(t *testing.T)

// Query engine tests
func TestQueryBuilder(t *testing.T)
func TestFacetComputation(t *testing.T)
func TestPagination(t *testing.T)
func TestMultipleFilters(t *testing.T)

// Repository tests
func TestURLParsing(t *testing.T)
func TestURLGeneration(t *testing.T)
func TestBreadcrumbs(t *testing.T)

// Burst detection tests
func TestBurstDetection(t *testing.T)
func TestBurstRepresentative(t *testing.T)

// Duplicate clustering tests
func TestClusterDetection(t *testing.T)
func TestHammingDistance(t *testing.T)
func TestBKTree(t *testing.T)
```

### 7.2 Integration Tests

```go
func TestEndToEndIndexing(t *testing.T) {
    // 1. Create temp database
    // 2. Index test directory (50 photos)
    // 3. Verify all photos indexed
    // 4. Verify thumbnails generated
    // 5. Verify color palettes extracted
    // 6. Run burst detection
    // 7. Run duplicate clustering
    // 8. Verify database integrity
}

func TestEndToEndQuery(t *testing.T) {
    // 1. Load test database
    // 2. Execute various queries
    // 3. Verify result counts
    // 4. Verify facet counts
    // 5. Verify thumbnail retrieval
}
```

### 7.3 Performance Tests

```go
func BenchmarkIndexing(b *testing.B)
func BenchmarkQuery(b *testing.B)
func BenchmarkFacetComputation(b *testing.B)
func BenchmarkBurstDetection(b *testing.B)
func BenchmarkDuplicateClustering(b *testing.B)
```

---

## 8. Deployment & Usage

### 8.1 Installation

```bash
# Install dependencies
go get github.com/rwcarlsen/goexif/exif
go get github.com/mattn/go-sqlite3
go get github.com/nfnt/resize
go get github.com/mccutchen/palettor
go get github.com/corona10/goimagehash
go get github.com/google/uuid

# Build
go build -o indexer

# Verify
./indexer --version
```

### 8.2 Typical Workflow

```bash
# 1. Initial indexing
./indexer index /path/to/photos -w 8
# Output: Indexed 10,247 photos in 8m 32s

# 2. Run analysis
./indexer analyze
# Output: 
#   - 234 burst groups detected
#   - 156 duplicate clusters found

# 3. Query photos
./indexer query "/2025/10"
# Shows October 2025 photos with facets

# 4. Query with filters
./indexer query "/2025/10?camera=Canon&tod=golden_hour_morning"
# Shows Canon photos taken during golden hour in October

# 5. View bursts
./indexer query --bursts
# Lists all burst groups

# 6. View specific burst
./indexer query "/bursts/a3f2b91d-..."
# Shows all photos in burst

# 7. View duplicates
./indexer query --duplicates exact
# Shows exact duplicate clusters

# 8. Export thumbnail
./indexer thumbnail 42 -s large -o /tmp/photo.jpg

# 9. Statistics
./indexer stats
./indexer stats --bursts
./indexer stats --duplicates

# 10. Maintenance
./indexer compact     # Vacuum database
./indexer verify      # Check integrity
```

---

## 9. Future Enhancements

### Phase 2 Features

1. **HTTP API Server**
   - Serve JSON responses
   - CORS support for web apps
   - Thumbnail serving endpoint

2. **Additional RAW Formats**
   - CR2, NEF, ARW, ORF support
   - Use LibRaw for decoding

3. **Video Support**
   - Extract video metadata
   - Generate frame thumbnails
   - FFmpeg integration

4. **Sharpness Scoring**
   - Laplacian variance method
   - Help select best burst shot

5. **Face Detection**
   - Local processing (privacy)
   - Face clustering
   - People facet

6. **XMP Sidecar Support**
   - Read XMP files
   - Write metadata changes

7. **Full-Text Search**
   - SQLite FTS5
   - Search across all text fields

8. **Smart Collections**
   - Rule-based collections
   - SQL query builder

9. **Export Features**
   - HTML gallery generation
   - CSV export
   - JSON API

10. **Web UI**
    - Browse via browser
    - Interactive filtering
    - Visual query builder

---

## 10. Success Criteria

### Functional Requirements
- ✅ Index 100K DNG files without errors
- ✅ Extract metadata from 95%+ of files
- ✅ Generate all 4 thumbnail sizes
- ✅ Detect bursts with 90%+ accuracy
- ✅ Cluster duplicates with 85%+ accuracy
- ✅ Support all documented URL patterns
- ✅ Return correct facet counts

### Performance Requirements
- ✅ Index at 15+ photos/second
- ✅ Query response < 500ms
- ✅ Thumbnail retrieval < 50ms
- ✅ Burst detection < 1 minute for 100K photos
- ✅ Database size < 250KB per photo (including thumbnails)

### Quality Requirements
- ✅ Zero data corruption
- ✅ Graceful error handling
- ✅ Clear error messages
- ✅ Cross-platform operation (Windows, macOS, Linux)
- ✅ Database portability

---

**END OF SPECIFICATION**

This specification provides complete architectural guidance for implementing a four-layer photo indexing and query system with faceted browsing capabilities, perceptual hashing, burst detection, and near-duplicate clustering.