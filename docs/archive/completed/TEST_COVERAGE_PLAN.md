# Test Coverage Plan: Minimal DNG Set for Complete Facet Coverage

## Objective

Calculate the smallest set of DNG images needed to fully exercise all faceted URL patterns and query dimensions supported by the Olsen indexer.

## Facet Dimensions Analysis

### 1. Temporal Facets
**URL Patterns**: `/YYYY`, `/YYYY/MM`, `/YYYY/MM/DD`

**Time of Day** (7 categories):
- Golden hour morning (5:00-7:00)
- Morning (7:00-11:00)
- Midday (11:00-15:00)
- Afternoon (15:00-18:00)
- Golden hour evening (18:00-20:00)
- Blue hour (20:00-22:00)
- Night (22:00-5:00)

**Seasons** (4 categories):
- Spring (March-May)
- Summer (June-August)
- Autumn (September-November)
- Winter (December-February)

**Minimum Images**: 7 (one per time of day, distributed across seasons)

### 2. Equipment Facets
**URL Patterns**: `/camera/:make`, `/camera/:make/:model`, `/lens/:model`

**Camera Makes** (2 minimum for diversity):
- Canon
- Nikon

**Camera Models** (2 minimum):
- Canon EOS R5
- Nikon Z9

**Lenses** (4 focal categories):
- Wide: < 35mm (e.g., 24mm)
- Normal: 35-70mm (e.g., 50mm)
- Telephoto: 71-200mm (e.g., 85mm)
- Super telephoto: > 200mm (e.g., 300mm)

**Minimum Images**: 4 (one per focal category)

### 3. Shooting Conditions Facets
**Based on ISO and flash**:

- Bright (ISO ≤ 400, no flash)
- Moderate (ISO 401-1599, no flash)
- Low light (ISO ≥ 1600, no flash)
- Flash (any ISO, flash fired)

**Minimum Images**: 4 (one per condition)

### 4. Color Facets
**URL Patterns**: `/color/:name`, `/color/hue/:degrees`

**Hue Names** (8 categories):
- Red (0-15°, 345-360°)
- Orange (16-45°)
- Yellow (46-75°)
- Green (76-165°)
- Cyan (166-195°)
- Blue (196-255°)
- Purple (256-285°)
- Pink (286-344°)

**Minimum Images**: 8 (one dominant color per hue category)

### 5. Burst/Duplicate Facets
**URL Patterns**: `/bursts/:id`, `/duplicates/:id`

**Burst Detection**:
- Requires 3+ photos within 2 seconds, same camera, similar focal length
- **Minimum**: 3 images for one burst group

**Duplicate Detection**:
- Requires perceptual hash similarity
- **Minimum**: 2 images (near-duplicates)

**Minimum Images**: 5 (3 burst + 2 duplicates)

### 6. Location Facets
**GPS Coordinates**:
- At least 1 image with GPS data
- At least 1 image without GPS data

**Minimum Images**: 2

## Optimization Strategy

Many facets can be combined into single images:
- Time of day + Season + Equipment + Shooting condition + Color
- Example: One image can satisfy multiple facets simultaneously

### Combined Requirement Matrix

| # | Time of Day | Season | Camera | Lens/Focal | ISO/Condition | Dominant Color | GPS | Special |
|---|------------|--------|--------|------------|---------------|----------------|-----|---------|
| 1 | Golden morning | Spring | Canon R5 | 24mm/Wide | 100/Bright | Red | Yes | - |
| 2 | Morning | Summer | Canon R5 | 50mm/Normal | 800/Moderate | Orange | No | - |
| 3 | Midday | Summer | Nikon Z9 | 85mm/Telephoto | 3200/Low light | Yellow | Yes | - |
| 4 | Afternoon | Autumn | Nikon Z9 | 300mm/Super tele | 400/Flash | Green | No | - |
| 5 | Golden evening | Autumn | Canon R5 | 24mm/Wide | 200/Bright | Cyan | Yes | - |
| 6 | Blue hour | Winter | Nikon Z9 | 50mm/Normal | 1600/Low light | Blue | No | - |
| 7 | Night | Winter | Canon R5 | 85mm/Telephoto | 6400/Low light | Purple | Yes | - |
| 8 | Morning | Spring | Nikon Z9 | 300mm/Super tele | 100/Bright | Pink | No | - |
| 9 | Midday | Spring | Canon R5 | 24mm/Wide | 100/Bright | Red | Yes | Burst 1/3 |
| 10 | Midday | Spring | Canon R5 | 24mm/Wide | 100/Bright | Red | Yes | Burst 2/3 |
| 11 | Midday | Spring | Canon R5 | 24mm/Wide | 100/Bright | Red | Yes | Burst 3/3 |
| 12 | Afternoon | Summer | Canon R5 | 50mm/Normal | 400/Bright | Green | No | Duplicate 1 |
| 13 | Afternoon | Summer | Canon R5 | 50mm/Normal | 400/Bright | Green | No | Duplicate 2 |

## Final Calculation

### Minimum Required Images: **13 DNG files**

### Coverage Verification

✅ **Time of Day**: All 7 categories covered
- Golden morning (1), Morning (2, 8), Midday (3, 9-11), Afternoon (4, 12-13), Golden evening (5), Blue hour (6), Night (7)

✅ **Seasons**: All 4 categories covered
- Spring (1, 8, 9-11), Summer (2, 3, 12-13), Autumn (4, 5), Winter (6, 7)

✅ **Camera Makes**: 2 makes covered
- Canon (1, 2, 5, 7, 9-13), Nikon (3, 4, 6, 8)

✅ **Camera Models**: 2 models covered
- Canon EOS R5 (1, 2, 5, 7, 9-13), Nikon Z9 (3, 4, 6, 8)

✅ **Focal Categories**: All 4 covered
- Wide (1, 5, 9-11), Normal (2, 6, 12-13), Telephoto (3, 7), Super telephoto (4, 8)

✅ **Shooting Conditions**: All 4 covered
- Bright (1, 5, 8, 9-12), Moderate (2), Low light (3, 6, 7), Flash (4)

✅ **Dominant Colors**: All 8 hue categories covered
- Red (1, 9-11), Orange (2), Yellow (3), Green (4, 12-13), Cyan (5), Blue (6), Purple (7), Pink (8)

✅ **GPS**: Both states covered
- With GPS (1, 3, 5, 7, 9-11), Without GPS (2, 4, 6, 8, 12-13)

✅ **Burst Groups**: 1 group with 3 photos (9-11)

✅ **Duplicate Clusters**: 1 cluster with 2 photos (12-13)

## Storage Requirements

### Per DNG File Estimates

**Typical DNG file from modern cameras**:
- Uncompressed: 50-80 MB
- Compressed: 25-40 MB
- Average: ~35 MB per file

### Total Storage

**13 files × 35 MB = 455 MB**

Range: 325 MB (compressed) to 1.04 GB (uncompressed)

### Database Storage

After indexing:
- Metadata per photo: ~2 KB
- Thumbnails per photo (4 sizes): ~187 KB
- Colors per photo: ~0.5 KB
- **Total per photo**: ~190 KB

**13 photos × 190 KB = 2.47 MB database**

### Grand Total

**Original DNG files**: 455 MB
**SQLite database**: 2.47 MB
**Combined**: ~457 MB

## Implementation Notes

### File Naming Convention

```
01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng
02_canon_r5_50mm_summer_morning_iso800_orange_nogps.dng
03_nikon_z9_85mm_summer_midday_iso3200_yellow_gps.dng
04_nikon_z9_300mm_autumn_afternoon_iso400_flash_green_nogps.dng
05_canon_r5_24mm_autumn_golden_evening_iso200_cyan_gps.dng
06_nikon_z9_50mm_winter_blue_hour_iso1600_blue_nogps.dng
07_canon_r5_85mm_winter_night_iso6400_purple_gps.dng
08_nikon_z9_300mm_spring_morning_iso100_pink_nogps.dng
09_burst_1_canon_r5_24mm_spring_midday_iso100_red_gps.dng
10_burst_2_canon_r5_24mm_spring_midday_iso100_red_gps.dng
11_burst_3_canon_r5_24mm_spring_midday_iso100_red_gps.dng
12_duplicate_1_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng
13_duplicate_2_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng
```

### EXIF Requirements for Each File

Each DNG must have:
1. **Camera Make/Model** (e.g., Canon, EOS R5)
2. **Lens Model** implied by focal length
3. **Focal Length** (24mm, 50mm, 85mm, 300mm)
4. **Focal Length 35mm** (for category classification)
5. **ISO** (100-6400 range)
6. **Aperture** (f/1.4-f/5.6)
7. **Shutter Speed** (1/1000-30s)
8. **Flash Fired** (boolean)
9. **Date Taken** (proper month/day/hour for temporal facets)
10. **GPS Coordinates** (for ~half the images)
11. **Image Dimensions** (various aspect ratios)

### Burst Group Requirements (Images 9-11)

- Same camera make/model
- Within 2 seconds of each other
- Same or similar focal length (±5mm)
- Sequential date_taken timestamps

### Duplicate Requirements (Images 12-13)

- Perceptual hash Hamming distance ≤ 15
- Can be same scene with minor variations
- Similar composition but different exposure

## Acquisition Options

### Option 1: Generate Synthetic DNGs
- Use `exiftool` or Python libraries to create DNGs with required EXIF
- Embed synthetic image data
- **Pros**: Complete control, free, reproducible
- **Cons**: Not "real" photos, may not test edge cases

### Option 2: Source from Sample DNG Repositories
- Adobe DNG samples
- Camera manufacturer sample files
- Creative Commons photo databases
- **Pros**: Real-world data
- **Cons**: May not have exact combinations needed

### Option 3: Hybrid Approach (Recommended)
- Source 2-3 real DNGs from different cameras
- Modify EXIF data to create variants
- Adjust image colors programmatically for dominant color testing
- **Pros**: Real image quality, controlled metadata
- **Cons**: Moderate complexity

## Testing Strategy

### Phase 1: Basic Coverage (8 files)
Start with 8 core images covering all primary facets (exclude burst/duplicate)
- **Storage**: ~280 MB
- **Validates**: All URL patterns except burst/duplicate

### Phase 2: Complete Coverage (13 files)
Add burst and duplicate images
- **Storage**: ~455 MB
- **Validates**: All URL patterns including burst/duplicate detection

### Phase 3: Extended Testing (Optional, 25+ files)
Add variations for:
- Multiple burst groups
- Multiple duplicate clusters
- More camera makes/models
- Edge cases (extreme ISOs, fisheye lenses, etc.)

## Success Criteria

After indexing the 13 DNG files, the system should support queries for:

- ✅ All 7 time-of-day periods
- ✅ All 4 seasons
- ✅ 2 camera makes, 2 models
- ✅ All 4 focal length categories
- ✅ All 4 shooting conditions
- ✅ All 8 dominant color hues
- ✅ GPS and non-GPS photos
- ✅ At least 1 burst group
- ✅ At least 1 duplicate cluster

## Conclusion

**Minimum viable test set**: 13 DNG files totaling ~455 MB

This provides complete coverage of all faceted URL patterns and query dimensions specified in the Olsen system architecture, enabling full validation of:
- Temporal browsing
- Equipment browsing
- Color browsing
- Burst detection
- Duplicate clustering
- All inferred metadata categories
- GPS/location filtering

The set is optimized to maximize facet coverage while minimizing redundancy and storage requirements.
