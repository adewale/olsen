# Query Reference Guide

**Olsen Photo Indexer - Complete Query Specification**
**Version:** 1.0
**Last Updated:** October 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Query Architecture](#query-architecture)
3. [URL-Based Queries](#url-based-queries)
4. [Parameter-Based Queries](#parameter-based-queries)
5. [Query Result Structure](#query-result-structure)
6. [Faceted Search](#faceted-search)
7. [Performance Considerations](#performance-considerations)
8. [Query Examples](#query-examples)

---

## Overview

Olsen supports a powerful faceted search system that allows browsing and filtering photos across multiple dimensions simultaneously. Queries can be expressed either as:

1. **URL Paths** - RESTful URLs for hierarchical browsing
2. **Query Parameters** - Key-value filters for complex multi-dimensional queries

Both approaches can be combined for maximum flexibility.

### Key Detection Parameters

**Burst Detection:**
- Time window: 2 seconds maximum between photos
- Focal length tolerance: ±5mm
- Same camera required (make + model)
- Minimum burst size: 3 photos

**Duplicate Detection:**
- Perceptual hash Hamming distance threshold: ≤15 for similarity
- Cluster types based on distance (exact=0, near=1-5, similar=>5)
- Minimum cluster size: 2 photos

---

## Query Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      User Request                            │
│         (URL path + query string parameters)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
           ┌─────────────────────────────┐
           │   Repository (URL Mapper)    │
           │  • Parse URL pattern         │
           │  • Extract path parameters   │
           │  • Parse query string        │
           └─────────────┬────────────────┘
                         │
                         ▼
           ┌─────────────────────────────┐
           │    Query Engine             │
           │  • Build SQL query          │
           │  • Apply filters            │
           │  • Execute query            │
           │  • Compute facets           │
           │  • Build breadcrumbs        │
           └─────────────┬────────────────┘
                         │
                         ▼
           ┌─────────────────────────────┐
           │      Query Result           │
           │  • Photos                   │
           │  • Total count              │
           │  • Facet counts             │
           │  • Breadcrumbs              │
           └─────────────────────────────┘
```

---

## URL-Based Queries

### Temporal Browsing

Browse photos by date using hierarchical URL paths:

```
/YYYY                → Year view (e.g., /2025)
/YYYY/MM             → Month view (e.g., /2025/10)
/YYYY/MM/DD          → Day view (e.g., /2025/10/15)
```

**Examples:**
```
/2025                # All photos from 2025
/2025/10             # All photos from October 2025
/2025/10/15          # All photos from October 15, 2025
```

**Implementation Notes:**
- Year must be 4 digits (YYYY)
- Month must be 2 digits with leading zero (01-12)
- Day must be 2 digits with leading zero (01-31)

---

### Equipment Browsing

Browse by camera or lens:

```
/camera/:make              → All photos from camera make
/camera/:make/:model       → Specific camera model
/lens/:model               → Specific lens
```

**Examples:**
```
/camera/Canon              # All Canon cameras
/camera/Canon/EOS-R5       # Canon EOS R5 only
/camera/Nikon/Z9           # Nikon Z9
/lens/RF24-70mm-F2.8       # Specific lens
```

**URL Encoding:**
- Spaces replaced with hyphens
- Special characters URL-encoded
- Case-sensitive matching

---

### Color Browsing

Browse by dominant colors:

```
/color/:name         → By color name
/color/hue/:degrees  → By hue range (0-360)
```

**Supported Color Names:**
```
red, orange, yellow, green, cyan, blue, purple, pink
```

**Examples:**
```
/color/red           # Photos with red as dominant color
/color/blue          # Photos with blue as dominant color
/color/hue/0         # Red hues (0° ±15°)
/color/hue/120       # Green hues (120° ±15°)
/color/hue/240       # Blue hues (240° ±15°)
```

**Hue Ranges:**
- Red: 0° (±15°)
- Orange: 30° (±15°)
- Yellow: 60° (±15°)
- Green: 120° (±15°)
- Cyan: 180° (±15°)
- Blue: 240° (±15°)
- Purple: 270° (±15°)
- Pink: 330° (±15°)

---

### Burst Browsing

Browse burst sequences:

```
/bursts              → List all burst groups
/bursts/:id          → Specific burst group
```

**Examples:**
```
/bursts                        # All burst groups
/bursts/20251015120000_0       # Specific burst by ID
```

**Burst Metadata:**
- Each burst has unique ID (timestamp of analysis run + index)
- Photos in burst sorted by sequence
- Representative photo is first in sequence (position 0)

---

### Duplicate Browsing

Browse duplicate clusters:

```
/duplicates              → All duplicate clusters
/duplicates/exact        → Only exact duplicates (distance 0)
/duplicates/near         → Near duplicates (distance 1-5)
/duplicates/similar      → Similar photos (distance 6-15)
/duplicates/:id          → Specific cluster
```

**Examples:**
```
/duplicates                    # All clusters
/duplicates/exact              # Only exact matches
/duplicates/near               # Near duplicates
/duplicates/similar            # Similar photos
/duplicates/20251015_dup_0     # Specific cluster by ID
```

**Cluster Types:**
- **exact**: Hamming distance = 0 (identical)
- **near**: Hamming distance 1-5 (nearly identical)
- **similar**: Hamming distance > 5 (similar composition, detected up to threshold of 15)

---

## Parameter-Based Queries

All queries support query string parameters for additional filtering:

### Temporal Parameters

```
?year=YYYY                # Year filter
?month=MM                 # Month (1-12)
?day=DD                   # Day (1-31)
?date_from=YYYY-MM-DD     # Start date (inclusive)
?date_to=YYYY-MM-DD       # End date (inclusive)
```

**Examples:**
```
?year=2025
?month=10
?date_from=2025-01-01&date_to=2025-12-31
```

---

### Equipment Parameters

```
?camera_make=<make>       # Camera manufacturer
?camera_model=<model>     # Specific camera model
?lens=<model>             # Lens model
```

**Examples:**
```
?camera_make=Canon
?camera_model=EOS-R5
?lens=RF24-70mm
?camera_make=Canon&lens=RF24-70mm
```

---

### Technical Parameters

```
?iso_min=<value>          # Minimum ISO
?iso_max=<value>          # Maximum ISO
?aperture_min=<f-number>  # Minimum aperture (f/1.4 = 1.4)
?aperture_max=<f-number>  # Maximum aperture
?focal_min=<mm>           # Minimum focal length (mm)
?focal_max=<mm>           # Maximum focal length (mm)
```

**Examples:**
```
?iso_min=100&iso_max=400         # ISO 100-400
?aperture_min=1.4&aperture_max=2.8  # f/1.4 to f/2.8
?focal_min=24&focal_max=70       # 24-70mm
```

---

### Categorical Parameters (Multi-Select)

```
?time_of_day=<value>[,<value>]           # Comma-separated list
?season=<value>[,<value>]                # Comma-separated list
?focal_category=<value>[,<value>]        # Comma-separated list
?shooting_condition=<value>[,<value>]    # Comma-separated list
```

**Time of Day Values:**
```
golden_hour_morning    # 1 hour after sunrise
morning                # Sunrise to midday
midday                 # 11am - 2pm
afternoon              # 2pm - sunset
golden_hour_evening    # 1 hour before sunset
blue_hour              # Civil twilight
night                  # After twilight
```

**Season Values:**
```
spring    # March, April, May
summer    # June, July, August
autumn    # September, October, November
winter    # December, January, February
```

**Focal Category Values:**
```
wide              # < 35mm
normal            # 35-70mm
telephoto         # 71-200mm
super_telephoto   # > 200mm
```

**Shooting Condition Values:**
```
bright     # ISO ≤ 400
moderate   # ISO 401-1599
low_light  # ISO ≥ 1600
flash      # Flash fired
```

**Examples:**
```
?time_of_day=golden_hour_morning,golden_hour_evening
?season=spring,summer
?focal_category=wide,normal
?shooting_condition=bright,moderate
```

---

### Color Parameters

```
?color=<name>             # Color name
?hue_min=<degrees>        # Minimum hue (0-360)
?hue_max=<degrees>        # Maximum hue (0-360)
?saturation_min=<percent> # Minimum saturation (0-100)
?saturation_max=<percent> # Maximum saturation (0-100)
?lightness_min=<percent>  # Minimum lightness (0-100)
?lightness_max=<percent>  # Maximum lightness (0-100)
```

**Examples:**
```
?color=red
?hue_min=0&hue_max=30               # Red-orange range
?saturation_min=50                  # Vibrant colors only
?lightness_min=30&lightness_max=70  # Exclude very dark/light
```

---

### Location Parameters

```
?lat_min=<degrees>        # Minimum latitude
?lat_max=<degrees>        # Maximum latitude
?lon_min=<degrees>        # Minimum longitude
?lon_max=<degrees>        # Maximum longitude
?has_gps=<true|false>     # Filter by GPS presence
```

**Examples:**
```
?lat_min=37.0&lat_max=38.0&lon_min=-123.0&lon_max=-122.0  # San Francisco area
?has_gps=true                                              # Only geotagged photos
?has_gps=false                                             # Only non-geotagged
```

---

### Burst/Cluster Parameters

```
?burst_group=<id>              # Specific burst group
?duplicate_cluster=<id>        # Specific cluster
?cluster_type=<exact|near|similar>  # Filter by cluster type
?only_representatives=<true>   # Show only representatives
```

**Examples:**
```
?burst_group=20251015120000_0       # Photos in specific burst
?cluster_type=exact                 # Exact duplicates only
?only_representatives=true          # One photo per burst/cluster
```

---

### Pagination Parameters

```
?offset=<number>          # Skip first N results (default: 0)
?limit=<number>           # Return max N results (default: 100)
```

**Examples:**
```
?limit=50                 # First 50 results
?offset=100&limit=50      # Results 101-150
```

**Pagination Notes:**
- Default limit: 100 photos
- Maximum limit: 1000 photos
- Use offset for pagination

---

### Sorting Parameters

```
?sort=<field>             # Sort field
?order=<asc|desc>         # Sort direction (default: desc)
```

**Sortable Fields:**
```
date_taken          # Photo timestamp (default)
indexed_at          # When indexed
camera_make         # Camera manufacturer
camera_model        # Camera model
focal_length        # Lens focal length
iso                 # ISO value
aperture            # Aperture (f-number)
file_size           # File size in bytes
```

**Examples:**
```
?sort=date_taken&order=asc     # Oldest first
?sort=iso&order=desc           # Highest ISO first
?sort=focal_length&order=asc   # Wide to telephoto
```

---

### Output Parameters

```
?thumbnail_size=<size>    # Thumbnail size to return
?include_facets=<true|false>  # Include facet counts (default: true)
```

**Thumbnail Sizes:**
```
64      # Tiny (grid view)
256     # Small (list view)
512     # Medium (preview)
1024    # Large (full preview)
```

**Examples:**
```
?thumbnail_size=512            # Medium thumbnails
?include_facets=false          # Skip facet computation (faster)
```

---

## Query Result Structure

All queries return a structured result containing:

```go
type QueryResult struct {
    Photos      []PhotoSummary     // Matching photos
    TotalCount  int                // Total matches (for pagination)
    Facets      FacetCollection    // Facet counts
    Breadcrumbs []Breadcrumb       // Navigation trail
    Query       QueryParams        // Normalized query params
}

type PhotoSummary struct {
    ID            int
    ThumbnailData []byte           // JPEG thumbnail
    DateTaken     time.Time
    CameraMake    string
    CameraModel   string
    DominantColor Color            // Primary color
    BurstInfo     *BurstInfo       // nil if not in burst
    ClusterInfo   *ClusterInfo     // nil if not in cluster
}

type BurstInfo struct {
    GroupID          string
    Sequence         int            // Position in burst (0-based)
    TotalCount       int            // Photos in burst
    IsRepresentative bool           // Is this the representative?
}

type ClusterInfo struct {
    ClusterID        string
    Size             int            // Photos in cluster
    Type             string         // "exact", "near", "similar"
    IsRepresentative bool
    SimilarityScore  float64        // 0.0-1.0 (1.0 = identical)
}
```

---

## Faceted Search

Facets provide counts for each possible filter value, respecting currently active filters.

### Facet Types

```go
type FacetCollection struct {
    Camera      []Facet           // Camera makes/models
    Lens        []Facet           // Lens models
    TimeOfDay   []Facet           // Time of day periods
    Season      []Facet           // Seasons
    Color       []ColorFacet      // Dominant colors
    Year        []Facet           // Years
    Month       []Facet           // Months
    FocalCategory []Facet         // Focal length categories
    ShootingCondition []Facet     // ISO-based conditions
    Burst       []BurstFacet      // Burst groups
    Duplicates  []DuplicateFacet  // Duplicate clusters
}

type Facet struct {
    Value    string               // Facet value (e.g., "Canon")
    Count    int                  // Number of photos
    Selected bool                 // Is this facet active?
}

type ColorFacet struct {
    Name  string                  // Color name
    Color Color                   // RGB values (R, G, B uint8)
    HSL   ColorHSL                // HSL values (H: 0-360, S: 0-100, L: 0-100)
    Count int                     // Number of photos
}

type BurstFacet struct {
    GroupID    string
    Count      int                // Photos in burst
    DateTaken  time.Time          // First photo timestamp
    TimeSpan   float64            // Burst duration (seconds)
}

type DuplicateFacet struct {
    ClusterID   string
    Count       int               // Photos in cluster
    Type        string            // "exact", "near", "similar"
    MaxDistance int               // Maximum Hamming distance
}
```

### Facet Computation Rules

1. **Respects Active Filters**: Facet counts reflect only photos matching current filters
2. **Excludes Own Dimension**: When computing a facet, its own filter is temporarily removed
3. **Zero-Count Facets**: Facets with 0 matches are typically omitted
4. **Sort Order**: Facets sorted by count (descending) or alphabetically

**Example:**

Query: `/camera/Canon?time_of_day=golden_hour_morning`

Camera facets:
- ✅ Show all camera makes/models (Canon filter excluded when computing)
- ✅ Counts respect time_of_day filter

TimeOfDay facets:
- ✅ Show all time periods
- ✅ Counts respect Canon filter
- ✅ `golden_hour_morning` marked as Selected

---

## Performance Considerations

### Indexed Fields

All filterable fields have database indexes for fast queries:

```sql
-- Temporal indexes
CREATE INDEX idx_photos_date_taken ON photos(date_taken);

-- Equipment indexes
CREATE INDEX idx_photos_camera ON photos(camera_make, camera_model);
CREATE INDEX idx_photos_lens ON photos(lens_model);

-- Technical indexes
CREATE INDEX idx_photos_iso ON photos(iso);
CREATE INDEX idx_photos_aperture ON photos(aperture);
CREATE INDEX idx_photos_focal_length ON photos(focal_length);

-- Categorical indexes
CREATE INDEX idx_photos_time_of_day ON photos(time_of_day);
CREATE INDEX idx_photos_season ON photos(season);
CREATE INDEX idx_photos_focal_category ON photos(focal_category);
CREATE INDEX idx_photos_shooting_condition ON photos(shooting_condition);

-- Location indexes
CREATE INDEX idx_photos_gps ON photos(latitude, longitude);

-- Perceptual hash index
CREATE INDEX idx_photos_phash ON photos(perceptual_hash);

-- Burst/cluster indexes
CREATE INDEX idx_photos_burst ON photos(burst_group_id);
CREATE INDEX idx_photos_cluster ON photos(duplicate_cluster_id);

-- Color indexes
CREATE INDEX idx_colors_hue ON photo_colors(hue);
CREATE INDEX idx_colors_saturation ON photo_colors(saturation);
CREATE INDEX idx_colors_lightness ON photo_colors(lightness);
```

### Query Optimization Tips

1. **Use Specific Filters**: Narrow queries run faster
2. **Limit Facet Computation**: Set `include_facets=false` when not needed
3. **Reasonable Limits**: Don't request more than 1000 photos at once
4. **Index-Friendly Ranges**: Use indexed fields in WHERE clauses
5. **Avoid Full Scans**: Always include at least one indexed filter

### Typical Query Performance

| Query Type | Expected Time | Notes |
|------------|--------------|-------|
| Single dimension (year/camera) | < 50ms | Uses single index |
| Multi-dimensional (2-3 filters) | < 100ms | Uses index intersection |
| Color queries | < 200ms | Requires join with photo_colors |
| Facet computation | +50-200ms | Depends on active filters |
| Full catalog scan | 1-5s | Avoid; use pagination |

---

## Query Examples

### Example 1: Golden Hour Photos with Canon R5

```
URL: /camera/Canon/EOS-R5?time_of_day=golden_hour_morning,golden_hour_evening

Query Parameters:
- camera_make: Canon
- camera_model: EOS-R5
- time_of_day: [golden_hour_morning, golden_hour_evening]
- limit: 100
- include_facets: true
- thumbnail_size: 256

Expected Result:
- Photos: All Canon R5 shots during golden hours
- Facets: Lens counts, season counts, year/month counts, etc.
- Total Count: Number of matching photos
```

---

### Example 2: October 2025 Burst Sequences

```
URL: /2025/10?only_representatives=true

Query Parameters:
- year: 2025
- month: 10
- only_representatives: true
- limit: 100

Expected Result:
- Photos: One representative from each burst in October 2025
- Facets: Camera counts, time_of_day counts, burst_group facets
- Total Count: Number of burst groups
```

---

### Example 3: Wide Angle Photos in Low Light

```
URL: /?focal_category=wide&shooting_condition=low_light

Query Parameters:
- focal_category: [wide]
- shooting_condition: [low_light]
- sort: iso
- order: desc
- limit: 50

Expected Result:
- Photos: Wide angle (<35mm) shots with ISO ≥ 1600
- Sorted by ISO (highest first)
- Facets: Camera counts, lens counts, time_of_day counts
```

---

### Example 4: Exact Duplicates Only

```
URL: /duplicates/exact?sort=date_taken&order=desc

Query Parameters:
- cluster_type: exact
- sort: date_taken
- order: desc
- limit: 100

Expected Result:
- Photos: All photos in exact duplicate clusters
- Sorted by date (newest first)
- Facets: Camera counts, cluster facets (with max_distance=0)
```

---

### Example 5: Complex Multi-Dimensional Query

```
URL: /2025/10?camera_make=Canon&focal_min=24&focal_max=70&time_of_day=golden_hour_morning&color=red

Query Parameters:
- year: 2025
- month: 10
- camera_make: Canon
- focal_min: 24
- focal_max: 70
- time_of_day: [golden_hour_morning]
- color: red
- limit: 100

Expected Result:
- Photos: Canon photos from Oct 2025, 24-70mm focal range, golden hour morning, red dominant color
- Facets: All dimensions with counts reflecting these filters
- Breadcrumbs: 2025 > October > Canon > 24-70mm > Golden Hour > Red
```

---

### Example 6: Burst Group Details

```
URL: /bursts/20251015120000_0

Query Parameters:
- burst_group_id: 20251015120000_0
- sort: burst_sequence
- order: asc
- limit: 100

Expected Result:
- Photos: All photos in this burst group, in sequence order
- BurstInfo: Populated for each photo with sequence number
- Facets: Limited (since burst_group is highly specific)
- Total Count: Number of photos in burst
```

---

### Example 7: San Francisco Geotagged Photos

```
URL: /?lat_min=37.7&lat_max=37.8&lon_min=-122.5&lon_max=-122.4

Query Parameters:
- lat_min: 37.7
- lat_max: 37.8
- lon_min: -122.5
- lon_max: -122.4
- sort: date_taken
- order: desc
- limit: 100

Expected Result:
- Photos: All photos within San Francisco bounding box
- Sorted by date (newest first)
- Facets: Camera, lens, time_of_day, season counts for this location
```

---

### Example 8: High-Speed Action Shots

```
URL: /?aperture_min=2.8&iso_min=1600&focal_min=70

Query Parameters:
- aperture_min: 2.8 (f/2.8 or wider)
- iso_min: 1600 (high ISO)
- focal_min: 70 (telephoto or longer)
- sort: iso
- order: desc
- limit: 100

Expected Result:
- Photos: Fast-aperture telephoto shots in low light (likely action/sports)
- Sorted by ISO (highest first)
- Facets: Camera bodies used, specific lenses, shooting conditions
```

---

## Advanced Query Patterns

### Combining URL Path with Parameters

URL paths can be combined with query parameters:

```
/2025/10?camera_make=Canon&time_of_day=golden_hour_morning
```

This combines:
- Path: October 2025
- Parameters: Canon cameras, golden hour morning

### Multi-Value Parameters

Use commas for multiple values:

```
?time_of_day=golden_hour_morning,golden_hour_evening,blue_hour
?camera_model=EOS-R5,EOS-R6,EOS-R3
```

### Range Queries

Use min/max parameters for ranges:

```
?iso_min=100&iso_max=400
?date_from=2025-01-01&date_to=2025-12-31
?focal_min=24&focal_max=70
```

### Boolean Filters

Use boolean parameters:

```
?has_gps=true              # Only geotagged
?only_representatives=true # Burst/cluster reps only
?include_facets=false      # Skip facet computation
```

### Sorting Strategies

```
# Chronological (most recent first)
?sort=date_taken&order=desc

# By equipment
?sort=camera_make&order=asc

# By technical settings
?sort=iso&order=desc
?sort=focal_length&order=asc
```

---

## Query String Encoding

All query parameters must be properly URL-encoded:

```
Raw:      camera_model=EOS R5
Encoded:  camera_model=EOS%20R5

Raw:      lens=RF24-70mm F/2.8
Encoded:  lens=RF24-70mm%20F%2F2.8
```

Special characters requiring encoding:
- Space: `%20` or `+`
- Forward slash: `%2F`
- Question mark: `%3F`
- Ampersand: `%26`
- Equals: `%3D`

---

## Error Handling

### Invalid Parameters

```json
{
  "error": "invalid_parameter",
  "message": "Year must be 4 digits",
  "parameter": "year",
  "value": "25"
}
```

### Unknown URL Pattern

```json
{
  "error": "unknown_pattern",
  "message": "No handler for URL pattern",
  "url": "/unknown/path"
}
```

### Out of Range

```json
{
  "error": "out_of_range",
  "message": "Offset exceeds total count",
  "offset": 10000,
  "total": 500
}
```

---

## Best Practices

1. **Start Broad, Narrow Down**: Begin with year/month, then add filters
2. **Use Facets**: Let facet counts guide refinement
3. **Pagination**: Always use reasonable limits (≤1000)
4. **Cache Results**: Query results are deterministic; cache when possible
5. **Monitor Performance**: Track slow queries and optimize indexes
6. **URL Design**: Use RESTful paths for primary dimension, parameters for filters
7. **Breadcrumbs**: Display breadcrumbs for user navigation context

---

## Future Enhancements

Potential future query capabilities:

- [ ] Full-text search in EXIF metadata
- [ ] Saved queries / smart collections
- [ ] Query history and favorites
- [ ] Machine learning tags (face detection, scene classification)
- [ ] Advanced color matching (color palettes, harmonies)
- [ ] Geospatial queries (within radius, near landmark)
- [ ] Weather data correlation (if available in EXIF)
- [ ] Lens correction metadata queries

---

## See Also

- [Architecture Documentation](./architecture.md)
- [Database Schema](../specs/olsen_specs.md#22-complete-database-schema)
- [Test Coverage](./TEST_COVERAGE_PLAN.md)
- [API Reference](./README.md)
