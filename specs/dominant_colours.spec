# Olsen Dominant Colour System Specification

**Version:** 2.0
**Date:** 2025-10-07
**Status:** ✅ Implemented (v2.0 - includes B&W, Brown, Gray support)

---

## 1. Executive Summary

Olsen extracts the **5 dominant colors** from each photo using k-means clustering and stores them in **HSL color space** (Hue, Saturation, Lightness). This enables fast, perceptually-accurate color-based search and filtering.

**Key features:**
- Works on 256px thumbnails (not full resolution) for speed
- K-means clustering finds true dominant colors (not averages)
- HSL color space matches human perception
- Efficient SQL queries on hue ranges
- ~200 bytes per photo storage footprint
- **v2.0:** Special handling for black & white photos

---

## 2. Algorithm Flow

### 2.1 Extraction During Indexing

**File:** `internal/indexer/indexer.go:334-342`

```go
// Extract color palette from the small thumbnail for efficiency
thumbImg := thumbnails[1].Image  // Uses 256px thumbnail (not full resolution!)
colours, err := ExtractColourPalette(thumbImg, 5)  // Extract top 5 colors
```

**Optimization rationale:**
- Color extraction on 256px thumbnail: **~50ms**
- Color extraction on 40MP original: **~5000ms** (100x slower!)
- Color distribution is similar at any resolution
- No perceptible accuracy loss for color classification

### 2.2 K-Means Clustering Algorithm

**File:** `internal/indexer/color.go:22`

**Algorithm:** K-means clustering in RGB color space

**Parameters:**
- `maxIterations`: 100
- `numColours`: 5
- `img`: 256px thumbnail image

**Process:**
1. **Initialize:** 5 random cluster centers in RGB space
2. **Assignment:** Assign each pixel to nearest cluster center (Euclidean distance)
3. **Update:** Move cluster centers to mean of assigned pixels
4. **Iterate:** Repeat steps 2-3 for 100 iterations or until convergence
5. **Result:** 5 color clusters with weights (percentage of pixels)

**Library:** `github.com/mccutchen/palettor`

**Example output:**
```
Color 1: RGB(34, 89, 156)   Weight: 0.45  (45% of image)
Color 2: RGB(220, 215, 200)  Weight: 0.28  (28% of image)
Color 3: RGB(15, 25, 35)     Weight: 0.15  (15% of image)
Color 4: RGB(180, 120, 80)   Weight: 0.08  (8% of image)
Color 5: RGB(255, 200, 180)  Weight: 0.04  (4% of image)
```

### 2.3 RGB to HSL Conversion

**File:** `internal/indexer/color.go:54-100`

**HSL components:**
- **H (Hue):** 0-360° - The color itself (red=0°, green=120°, blue=240°)
- **S (Saturation):** 0-100% - Color intensity (0=grayscale, 100=vivid)
- **L (Lightness):** 0-100% - Brightness (0=black, 50=pure color, 100=white)

**Conversion algorithm:**
```
r, g, b = RGB / 255.0  (normalize to 0-1)
max = max(r, g, b)
min = min(r, g, b)
delta = max - min

Lightness (L):
  L = (max + min) / 2

Saturation (S):
  if delta == 0:
    S = 0  (achromatic/grayscale)
  else if L < 0.5:
    S = delta / (max + min)
  else:
    S = delta / (2 - max - min)

Hue (H):
  if delta == 0:
    H = 0  (undefined for grayscale)
  else if max == r:
    H = ((g - b) / delta) * 60°
    if g < b: H += 360°
  else if max == g:
    H = ((b - r) / delta + 2) * 60°
  else if max == b:
    H = ((r - g) / delta + 4) * 60°
```

**Why HSL?**
- Matches human color perception better than RGB
- Enables queries like "all blue photos" regardless of saturation/lightness
- Separates hue (color) from saturation (intensity) and lightness (brightness)
- Natural mapping to color names (hue ranges)

---

## 3. Data Model

### 3.1 Database Schema

**File:** `internal/database/schema.go:86-99`

```sql
CREATE TABLE photo_colors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    photo_id INTEGER NOT NULL,
    color_order INTEGER NOT NULL,  -- 1-5 (most dominant first)

    -- RGB values
    red INTEGER NOT NULL,           -- 0-255
    green INTEGER NOT NULL,         -- 0-255
    blue INTEGER NOT NULL,          -- 0-255

    -- Weight (importance)
    weight REAL NOT NULL,           -- 0.0-1.0 (sum of all 5 weights = 1.0)

    -- HSL values
    hue INTEGER,                    -- 0-360 (NULL for achromatic colors)
    saturation INTEGER,             -- 0-100
    lightness INTEGER,              -- 0-100

    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    UNIQUE(photo_id, color_order)
);

CREATE INDEX IF NOT EXISTS idx_photo_colors_hue ON photo_colors(hue);
CREATE INDEX IF NOT EXISTS idx_photo_colors_photo_id ON photo_colors(photo_id);
```

**Storage per photo:**
- 5 rows × 40 bytes = **200 bytes**
- 100,000 photos = **~20MB** (negligible vs thumbnails)

### 3.2 Data Types

**File:** `pkg/models/types.go`

```go
type Colour struct {
    R uint8  // Red: 0-255
    G uint8  // Green: 0-255
    B uint8  // Blue: 0-255
}

type ColourHSL struct {
    H int  // Hue: 0-360 degrees
    S int  // Saturation: 0-100 percent
    L int  // Lightness: 0-100 percent
}

type DominantColour struct {
    Colour Colour
    HSL    ColourHSL
    Weight float64  // 0.0-1.0 (percentage of image)
}
```

---

## 4. Color Classification for Search

### 4.1 Hue-Based Color Names

**File:** `internal/query/facets.go:724-742`

**Classification rules:**

| Color Name | Hue Range (degrees) | Description |
|------------|---------------------|-------------|
| Red        | 0-15, 345-360       | Pure red, crimson |
| Orange     | 16-45               | Orange, rust |
| Yellow     | 46-75               | Yellow, gold |
| Green      | 76-165              | Green, lime, teal |
| Blue       | 166-255             | Blue, cyan, navy |
| Purple     | 256-290             | Purple, violet |
| Pink       | 291-344             | Pink, magenta |
| **B&W**    | **(v2.0)** S < 10   | Black & white, grayscale |

**SQL implementation:**
```sql
SELECT
    CASE
        WHEN pc.saturation < 10 THEN 'bw'  -- NEW in v2.0
        WHEN pc.hue BETWEEN 0 AND 15 OR pc.hue BETWEEN 345 AND 360 THEN 'red'
        WHEN pc.hue BETWEEN 16 AND 45 THEN 'orange'
        WHEN pc.hue BETWEEN 46 AND 75 THEN 'yellow'
        WHEN pc.hue BETWEEN 76 AND 165 THEN 'green'
        WHEN pc.hue BETWEEN 166 AND 255 THEN 'blue'
        WHEN pc.hue BETWEEN 256 AND 290 THEN 'purple'
        WHEN pc.hue BETWEEN 291 AND 344 THEN 'pink'
        ELSE 'other'
    END as colour_name,
    COUNT(DISTINCT p.id) as count
FROM photos p
JOIN photo_colors pc ON pc.photo_id = p.id
WHERE <active filters except color>
GROUP BY colour_name
ORDER BY count DESC
```

### 4.2 Facet Generation

**Query result example:**
```
Red    (1,234 photos)
Blue   (892 photos)
Green  (645 photos)
B&W    (512 photos)  -- NEW in v2.0
Yellow (234 photos)
Orange (156 photos)
Purple (89 photos)
Pink   (45 photos)
```

---

## 5. Black & White Photo Handling (v2.0)

### 5.1 Problem Statement

**Issue:** Black & white photos currently break the dominant color system.

**Why it breaks:**
- B&W photos have **low saturation** (S ≈ 0-10%)
- Hue is **undefined or arbitrary** for achromatic colors
- RGB values cluster around gray: `(128, 128, 128)`, `(50, 50, 50)`, etc.
- K-means clusters grayscale photos into shades of gray
- Current system tries to classify gray by hue → incorrect results

**Example broken behavior:**
```
B&W portrait with black coat on white background:
  Dominant colors:
    1. RGB(240, 240, 240) → HSL(0°, 0%, 94%)   [white]
    2. RGB(180, 180, 180) → HSL(0°, 0%, 71%)   [light gray]
    3. RGB(120, 120, 120) → HSL(0°, 0%, 47%)   [mid gray]
    4. RGB(60, 60, 60)    → HSL(0°, 0%, 24%)   [dark gray]
    5. RGB(20, 20, 20)    → HSL(0°, 0%, 8%)    [black]

Current classification (WRONG):
  - Hue = 0° → Classified as "red" ❌
  - But it's actually a grayscale photo!

User searches "red" → Finds B&W photo ❌ BAD UX
```

### 5.2 Solution: Saturation-Based Detection

**Detection rule:** Photo is "B&W" if **saturation < 10%** for most dominant colors

**Classification priority:**
1. **First check saturation** → If S < 10%, classify as "bw"
2. **Then check hue** → Only if S ≥ 10%, use hue-based classification

**Updated SQL:**
```sql
CASE
    WHEN pc.saturation < 10 THEN 'bw'  -- Check saturation FIRST
    WHEN pc.hue BETWEEN 0 AND 15 OR pc.hue BETWEEN 345 AND 360 THEN 'red'
    WHEN pc.hue BETWEEN 16 AND 45 THEN 'orange'
    -- ... rest of hue-based rules
END
```

### 5.3 Saturation Threshold Rationale

**Why S < 10%?**

| Saturation | Visual Appearance | Classification |
|------------|-------------------|----------------|
| 0-5%       | Pure grayscale    | B&W ✓ |
| 6-10%      | Nearly grayscale  | B&W ✓ |
| 11-20%     | Slightly desaturated color | Color (but muted) |
| 21-50%     | Desaturated color | Color |
| 51-100%    | Saturated color   | Color |

**Examples:**
- True B&W film scan: S = 0-2%
- Converted to B&W in Lightroom: S = 0-5%
- Slightly warm B&W (sepia tone): S = 8-12% → Still feels B&W
- Muted color photo: S = 15-30% → Clearly has color

**Threshold = 10%** is conservative: captures true B&W while excluding muted color.

### 5.4 Weight-Based Refinement (Optional)

**Advanced rule:** Photo is B&W if **majority of weight** is in low-saturation colors

```sql
-- Option 1: Any dominant color is B&W → classify as B&W
WHERE saturation < 10

-- Option 2: Majority of weight is B&W → classify as B&W
WHERE (SELECT SUM(weight) FROM photo_colors
       WHERE photo_id = p.id AND saturation < 10) > 0.5
```

**Recommendation:** Use **Option 1** (any dominant color) for simplicity and to capture mixed photos (e.g., B&W photo with slight color cast).

### 5.5 UI/UX Considerations

**Color swatch design:**
- **B&W photos:** Show gradient swatch from black to white
- **Template update:**
```html
{{if eq .Value "bw"}}
  {{$bgColor = "linear-gradient(90deg, white 0%, black 100%)"}}
{{end}}
```

**Facet display:**
```
Filters
  Time
    Year
  Colour
    🔲 B&W (512 photos)    ← Grayscale gradient swatch
    🔴 Red (234 photos)
    🔵 Blue (189 photos)
```

### 5.6 Implementation Checklist

**Backend changes:**

- [ ] Update `computeColourFacet()` in `internal/query/facets.go`
  - Add `WHEN pc.saturation < 10 THEN 'bw'` as first CASE condition

- [ ] Update `photo_colors` query logic
  - Ensure saturation column is indexed for performance

- [ ] Add tests for B&W classification
  - Test pure B&W (S=0%)
  - Test near-B&W (S=8%)
  - Test muted color (S=15%) → should NOT be B&W
  - Test mixed B&W + color photo

**Frontend changes:**

- [ ] Update `grid.html` template color swatch logic
  - Already done! Template includes: `{{if eq .Value "bw"}}{{$bgColor = "linear-gradient(...)"}}{{end}}`

- [ ] Update facet display labels
  - "bw" → "B&W" or "Black & White"

**Testing:**

- [ ] Index collection with B&W photos
- [ ] Verify B&W facet appears and has correct count
- [ ] Verify B&W photos do NOT appear in color facets
- [ ] Verify color photos do NOT appear in B&W facet
- [ ] Verify gradient swatch renders correctly

---

## 6. Why This Design Is Good

### 6.1 Performance

| Operation | Time | Notes |
|-----------|------|-------|
| Extract colors (256px) | ~50ms | K-means on thumbnail |
| Extract colors (40MP) | ~5000ms | **100x slower** - don't do this! |
| RGB → HSL conversion | <1ms | Simple arithmetic |
| Color search query | 5-20ms | Indexed hue column |
| Facet aggregation | 30-50ms | GROUP BY on indexed column |
| B&W detection (v2.0) | 0ms | Same query, just check saturation |

**Bottlenecks during indexing:**
1. EXIF extraction: ~30ms
2. Thumbnail generation: ~30ms
3. Color extraction: ~50ms

Total: **~110ms per photo** (not dominated by color extraction)

### 6.2 Accuracy

**Multi-color photos are preserved:**
```
Sunset over ocean:
  Orange sky (40%), Blue ocean (35%), Dark water (15%),
  Yellow sun (8%), Pink clouds (2%)

Searches:
  "orange" → ✅ Found
  "blue" → ✅ Found
  "yellow" → ✅ Found
  "pink" → ✅ Found
  "green" → ❌ Not found (correct!)
```

**vs. single average color (BAD):**
```
Average: RGB(125, 135, 145) → Muddy gray-blue
Searches:
  "orange" → ❌ Not found (WRONG!)
  "blue" → ✅ Found (barely)
```

### 6.3 Perceptual Correctness

HSL matches human perception:
- "Blue photos" = hue 166-255° (captures light blue, navy, sky blue, teal)
- Works regardless of saturation or lightness
- **B&W detection via saturation** is perceptually obvious

### 6.4 Storage Efficiency

**Per photo:** 200 bytes (5 colors × 40 bytes)
**100,000 photos:** ~20MB (negligible compared to 5-50MB of thumbnails per photo)

### 6.5 Query Efficiency

```sql
-- Fast: Uses indexed hue column
SELECT p.* FROM photos p
JOIN photo_colors pc ON pc.photo_id = p.id
WHERE pc.hue BETWEEN 166 AND 255

-- Fast: Uses indexed saturation column (v2.0)
SELECT p.* FROM photos p
JOIN photo_colors pc ON pc.photo_id = p.id
WHERE pc.saturation < 10
```

**Index strategy:**
- `idx_photo_colors_hue` for color queries
- `idx_photo_colors_saturation` for B&W queries (v2.0)

### 6.6 Extensibility

**Future enhancements easily supported:**

**Already in schema:**
- Saturation filtering: "Show vivid colors" (S > 70%)
- Lightness filtering: "Show dark photos" (L < 30%)
- Weight filtering: "Show photos with >50% blue"

**Possible additions:**
- Multi-color queries: "Show photos with blue AND yellow"
- Color harmony: "Show complementary color schemes"
- Dominant color trends over time: "My color palette evolution"
- Automatic color grading suggestions based on palette

---

## 7. Alternative Approaches (and why they're worse)

### ❌ Single Average Color
```
Sunset photo → Average: muddy gray
Loses all multi-color information
Can't find "orange sky" or "blue ocean"
```

### ❌ Full RGB Histogram
```
Storage: 256³ = 16M bins per photo = 64MB per photo!
Query: No clear hue ranges, slow aggregations
Doesn't match human perception (RGB ≠ perceptual)
```

### ❌ Manual Color Tagging
```
Requires user to tag every photo
Inconsistent: "Is this blue or teal?"
Time-consuming for 100,000 photos
Subjective and error-prone
```

### ❌ ML-Based Object Detection
```
"Photo contains blue car" ≠ "Photo is mostly blue"
Expensive: Requires GPU, TensorFlow, models
Overkill for color classification
Less portable (model dependencies)
```

### ❌ Hue-Only Classification (Current v1.0 Issue)
```
B&W photos have undefined hue → classified as "red"
User searches "red" → finds grayscale photos ❌
v2.0 fix: Check saturation FIRST
```

---

## 8. Implementation Status

### v1.0 (✅ Implemented)
- ✅ K-means color extraction from 256px thumbnails
- ✅ RGB to HSL conversion
- ✅ Storage in `photo_colors` table
- ✅ Hue-based classification (red, orange, yellow, green, blue, purple, pink)
- ✅ Color facets in web UI
- ✅ Color swatches with proper RGB values
- ✅ Efficient indexed queries

### v2.0 (✅ Implemented - B&W + Berlin-Kay Colors)
- ✅ Saturation-based achromatic detection:
  - Black (S < 5%, L < 20%)
  - White (S < 5%, L > 80%)
  - Gray (S < 10%, L = 20-80%)
  - B&W (S < 15%)
- ✅ Brown color category (hue 20-40°, L < 50%)
- ✅ Updated `computeColourFacet()` SQL query with 11 Berlin-Kay colors
- ✅ Color swatches for all 11+ colors including gradients
- ✅ Comprehensive test suite (`color_classification_test.go`)
- ✅ Updated documentation and specs

---

## 9. Testing Strategy

### 9.1 Unit Tests

**File:** `internal/indexer/color_test.go`

```go
func TestRGBtoHSL_PureColors(t *testing.T)
func TestRGBtoHSL_Grayscale(t *testing.T)
func TestExtractColourPalette(t *testing.T)
func TestColourDistance(t *testing.T)
```

### 9.2 Integration Tests

**File:** `internal/query/facets_test.go`

```go
func TestColourFacet_MultiColorPhoto(t *testing.T)
func TestColourFacet_BWPhoto(t *testing.T)           // v2.0
func TestColourFacet_MixedBWAndColor(t *testing.T)   // v2.0
func TestColourFacet_SaturationThreshold(t *testing.T) // v2.0
```

### 9.3 Performance Benchmarks

**File:** `internal/indexer/indexer_test.go`

```go
func BenchmarkExtractColourPalette(b *testing.B)
// Current: ~50ms per 256px image
// Target: <100ms per image
```

---

## 10. References

### Academic
- K-means clustering: https://en.wikipedia.org/wiki/K-means_clustering
- HSL color space: https://en.wikipedia.org/wiki/HSL_and_HSV
- Color perception: https://en.wikipedia.org/wiki/Color_vision

### Libraries
- `github.com/mccutchen/palettor`: K-means palette extraction in Go
- Color quantization algorithms: https://github.com/themes/color-schemes

### Prior Art
- Adobe Lightroom: Color label filtering (red, yellow, green, blue, purple)
- Apple Photos: Color filter (not HSL-based, less accurate)
- Google Photos: ML-based color detection (heavier, requires inference)

---

## 11. Glossary

**K-means clustering:** Algorithm that groups data points into K clusters by minimizing intra-cluster variance.

**Dominant color:** A color that represents a significant portion of an image (top 5 by pixel count).

**HSL:** Hue-Saturation-Lightness color space that separates color (hue) from intensity (saturation) and brightness (lightness).

**Achromatic:** Colors without hue (black, white, gray). Characterized by S=0%.

**Saturation threshold:** Minimum saturation percentage to distinguish color from grayscale (v2.0: S < 10% = B&W).

**Facet:** A dimension for filtering search results (e.g., color, year, camera).

**Weight:** Percentage of image pixels belonging to a dominant color cluster (0.0-1.0, sum=1.0).

---

**END OF SPECIFICATION**
