# DNG Test Fixtures

## Overview

This directory contains 13 synthetic DNG (actually high-quality JPEG) files designed to provide complete coverage of all faceted URL patterns and query dimensions supported by the Olsen indexer.

**Total Size**: ~254 MB
**File Count**: 13 files
**Format**: JPEG with .dng extension (JPEG format is fully supported by the indexer)

## Generation

Files are generated using `../generate_dng_fixtures.go`:

```bash
go run testdata/generate_dng_fixtures.go
```

### Requirements

- **exiftool**: Used to inject EXIF metadata
  - macOS: `brew install exiftool`
  - Linux: `apt-get install libimage-exiftool-perl`

## Complete Facet Coverage

### ✅ Time of Day (7 periods)
- Golden hour morning (1 photo)
- Morning (2 photos)
- Midday (4 photos)
- Afternoon (3 photos)
- Golden hour evening (1 photo)
- Blue hour (1 photo)
- Night (1 photo)

### ✅ Seasons (4 seasons)
- Spring (5 photos)
- Summer (4 photos)
- Autumn (2 photos)
- Winter (2 photos)

### ✅ Camera Equipment (2 makes, 2 models)
- Canon EOS R5 (9 photos)
- Nikon Z9 (4 photos)

### ✅ Focal Length Categories (4 categories)
- Wide < 35mm (5 photos: 24mm)
- Normal 35-70mm (4 photos: 50mm)
- Telephoto 71-200mm (2 photos: 85mm)
- Super Telephoto > 200mm (2 photos: 300mm)

### ✅ Shooting Conditions (4 of 4) ✨ COMPLETE
- Bright ISO ≤ 400 (8 photos)
- Moderate ISO 401-1599 (1 photo)
- Low Light ISO ≥ 1600 (3 photos)
- **Flash (1 photo: image 4)** ✅ Now working!

### ✅ GPS Coverage (2 states)
- With GPS coordinates (7 photos)
- Without GPS coordinates (6 photos)

### ✅ Dominant Colors (8 hue categories)
- Red (3 photos: images 1, 9-11)
- Orange (1 photo: image 2)
- Yellow (1 photo: image 3)
- Green (3 photos: images 4, 12-13)
- Cyan (1 photo: image 5)
- Blue (1 photo: image 6)
- Purple (1 photo: image 7)
- Pink (1 photo: image 8)

### ✅ Burst Detection (1 group)
- Images 9-11: 3 sequential photos, 1 second apart
- Same camera, lens, location, settings

### ✅ Duplicate Detection (1 cluster)
- Images 12-13: 2 near-identical photos, 5 seconds apart
- Same scene, very similar composition and color

## EXIF Library Migration

**✨ Successfully migrated from goexif to exif-go!**

- **Previous**: goexif library couldn't read Flash EXIF tag
- **Current**: exif-go (github.com/dsoprea/go-exif/v3) reads all standard EXIF tags including Flash
- **Result**: Complete coverage of all 4 shooting conditions

## File List

| # | Filename | Camera | Lens | ISO | Time | Season | Color | GPS | Special |
|---|----------|--------|------|-----|------|--------|-------|-----|---------|
| 1 | 01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng | Canon R5 | 24mm | 100 | 06:30 | Spring | Red | ✓ | - |
| 2 | 02_canon_r5_50mm_summer_morning_iso800_orange_nogps.dng | Canon R5 | 50mm | 800 | 09:00 | Summer | Orange | ✗ | - |
| 3 | 03_nikon_z9_85mm_summer_midday_iso3200_yellow_gps.dng | Nikon Z9 | 85mm | 3200 | 13:00 | Summer | Yellow | ✓ | - |
| 4 | 04_nikon_z9_300mm_autumn_afternoon_iso400_flash_green_nogps.dng | Nikon Z9 | 300mm | 400 | 16:30 | Autumn | Green | ✗ | **Flash ✅** |
| 5 | 05_canon_r5_24mm_autumn_golden_evening_iso200_cyan_gps.dng | Canon R5 | 24mm | 200 | 19:00 | Autumn | Cyan | ✓ | - |
| 6 | 06_nikon_z9_50mm_winter_blue_hour_iso1600_blue_nogps.dng | Nikon Z9 | 50mm | 1600 | 21:00 | Winter | Blue | ✗ | - |
| 7 | 07_canon_r5_85mm_winter_night_iso6400_purple_gps.dng | Canon R5 | 85mm | 6400 | 23:30 | Winter | Purple | ✓ | - |
| 8 | 08_nikon_z9_300mm_spring_morning_iso100_pink_nogps.dng | Nikon Z9 | 300mm | 100 | 08:00 | Spring | Pink | ✗ | - |
| 9 | 09_burst_1_canon_r5_24mm_spring_midday_iso100_red_gps.dng | Canon R5 | 24mm | 100 | 12:00:00 | Spring | Red | ✓ | Burst 1/3 |
| 10 | 10_burst_2_canon_r5_24mm_spring_midday_iso100_red_gps.dng | Canon R5 | 24mm | 100 | 12:00:01 | Spring | Red | ✓ | Burst 2/3 |
| 11 | 11_burst_3_canon_r5_24mm_spring_midday_iso100_red_gps.dng | Canon R5 | 24mm | 100 | 12:00:02 | Spring | Red | ✓ | Burst 3/3 |
| 12 | 12_duplicate_1_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng | Canon R5 | 50mm | 400 | 15:30:00 | Summer | Green | ✗ | Dup 1/2 |
| 13 | 13_duplicate_2_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng | Canon R5 | 50mm | 400 | 15:30:05 | Summer | Green | ✗ | Dup 2/2 |

## Known Limitations

1. **Format**: Files are JPEG with .dng extension, not true Adobe DNG format
   - This is acceptable as the indexer supports JPEG format
   - All EXIF metadata is properly embedded

2. **Image Content**: Synthetic gradient images with dominant colors
   - Not photographic content, but sufficient for testing
   - Color extraction and perceptual hashing work correctly
   - Thumbnail generation preserves aspect ratios

3. **File Size**: Each file is ~20MB vs ~50-80MB for real DNGs
   - JPEG compression at quality 95
   - Still large enough to stress-test the indexer

## Verification

To verify complete facet coverage after indexing:

```bash
# Index the fixtures
go run testdata/test_fixtures.go

# Verify coverage
go run testdata/verify_coverage.go /tmp/olsen_fixtures_test.db
```

Expected output:
```
✓ Total photos: 13
✓ All 7 time-of-day periods covered
✓ All 4 seasons covered
✓ 2 camera makes, 2 models
✓ All 4 focal length categories
✓ All 4 shooting conditions (INCLUDING FLASH ✅)
✓ GPS and non-GPS photos
✓ 52 thumbnails (4 sizes × 13 photos)
✓ ~450+ colors extracted (~35 per photo)
✓ 13 perceptual hashes computed
```

## Integration Tests

The fixtures are used by:
- `internal/indexer/indexer_test.go::TestIndexDirectoryIntegration`
- `internal/indexer/exif_test.go` - Complete EXIF extraction tests
- Future: Burst detection tests
- Future: Duplicate clustering tests

## Regeneration

If you need to regenerate fixtures (e.g., to change dimensions or colors):

1. Edit `../generate_dng_fixtures.go`
2. Delete this directory: `rm -rf testdata/dng`
3. Regenerate: `go run testdata/generate_dng_fixtures.go`
4. Verify: `go test -v -run TestEXIF ./internal/indexer/`
