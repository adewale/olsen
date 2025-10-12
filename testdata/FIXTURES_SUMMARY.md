# Test Fixtures Summary

## Overview

Olsen includes two sets of test fixtures to support different testing scenarios:

1. **Basic Test Fixtures** (`testdata/photos/`) - 4 files, ~300KB
2. **Complete DNG Fixtures** (`testdata/dng/`) - 13 files, ~254MB

## Basic Test Fixtures (`testdata/photos/`)

### Purpose
- Quick unit and integration tests
- Continuous integration (CI) testing
- File format support verification (JPEG, BMP)
- Basic functionality validation

### Contents
- `test1.jpg` - 800×600 landscape JPEG
- `test2.jpg` - 1200×800 landscape JPEG
- `scan1.bmp` - 600×800 portrait BMP
- `subfolder/test3.jpg` - 640×480 gradient JPEG

### Size
**Total**: ~300KB

### Generation
```bash
go run testdata/generate_fixtures.go
```

### Test Coverage
- ✅ Multi-format support (JPEG, BMP)
- ✅ Recursive directory scanning
- ✅ Thumbnail generation (aspect ratio preservation)
- ✅ Color extraction
- ✅ Perceptual hashing
- ✅ Re-indexing skip logic
- ❌ Limited EXIF metadata (no camera, lens, GPS data)
- ❌ Incomplete facet coverage

### Used By
- `internal/indexer/integration_test.go` (5 tests)
- All unit tests that need sample images

---

## Complete DNG Fixtures (`testdata/dng/`)

### Purpose
- **Complete facet coverage testing**
- Validation of all metadata inference logic
- Burst detection testing
- Duplicate clustering testing
- Stress testing with large files

### Contents
13 synthetic DNG files (actually JPEG with .dng extension):
- 8 unique photos covering all metadata dimensions
- 3 photos forming a burst sequence (1-second intervals)
- 2 photos forming a duplicate pair

### Size
**Total**: ~254MB (~20MB per file)

### Generation
```bash
# Requires exiftool
brew install exiftool  # macOS
# or
apt-get install libimage-exiftool-perl  # Linux

go run testdata/generate_dng_fixtures.go
```

### Complete Facet Coverage

#### ✅ Temporal Facets
- **Time of Day** (7 periods): Golden morning, Morning, Midday, Afternoon, Golden evening, Blue hour, Night
- **Seasons** (4): Spring, Summer, Autumn, Winter
- **Date Range**: 2025-01-10 to 2025-12-05

#### ✅ Equipment Facets
- **Camera Makes** (2): Canon, Nikon
- **Camera Models** (2): Canon EOS R5, Nikon Z9
- **Focal Lengths** (4 categories):
  - Wide: 24mm (5 photos)
  - Normal: 50mm (4 photos)
  - Telephoto: 85mm (2 photos)
  - Super Telephoto: 300mm (2 photos)

#### ✅ Shooting Conditions
- **Bright** (ISO ≤ 400): 8 photos
- **Moderate** (ISO 401-1599): 1 photo
- **Low Light** (ISO ≥ 1600): 3 photos
- **Flash**: 1 photo ✅ Working with exif-go library

#### ✅ Color Facets
- **All 8 Hue Categories**: Red, Orange, Yellow, Green, Cyan, Blue, Purple, Pink
- **~450+ Colors Total**: Average 35 colors per photo
- **Synthetic Gradients**: Each image has dominant color + variations

#### ✅ Location Facets
- **With GPS**: 7 photos (San Francisco, New York, London, Paris, Los Angeles)
- **Without GPS**: 6 photos

#### ✅ Burst Detection
- **1 Burst Group**: Photos 9-11
- **Timing**: 1-second intervals (12:00:00, 12:00:01, 12:00:02)
- **Consistency**: Same camera, lens, location, settings

#### ✅ Duplicate Detection
- **1 Duplicate Pair**: Photos 12-13
- **Timing**: 5-second interval
- **Similarity**: Near-identical composition and color

### Verification

After indexing, verify complete coverage:

```bash
# Index fixtures
go run testdata/test_fixtures.go

# Verify coverage
go run testdata/verify_coverage.go /tmp/olsen_fixtures_test.db
```

**Expected Results:**
```
✓ Total photos: 13
✓ Time of Day: 7 periods covered
✓ Seasons: 4 seasons covered
✓ Camera Coverage: 2 makes, 2 models
✓ Focal Length: 4 categories covered
✓ Shooting Conditions: 4 of 4 (including flash)
✓ GPS: Both states covered
✓ Colors: ~450+ extracted
✓ Thumbnails: 52 (4 sizes × 13 photos)
✓ Perceptual Hashes: 13/13 computed
```

### Used By
- `internal/indexer/indexer_test.go::TestIndexDirectoryIntegration`
- Future: Burst detection tests
- Future: Duplicate clustering tests
- Future: Performance benchmarks

---

## Comparison

| Aspect | Basic Fixtures | Complete DNG Fixtures |
|--------|---------------|----------------------|
| **File Count** | 4 files | 13 files |
| **Total Size** | ~300KB | ~254MB |
| **Formats** | JPEG, BMP | JPEG (named .dng) |
| **EXIF Metadata** | Minimal | Complete |
| **Facet Coverage** | Partial | Complete |
| **Camera Data** | ❌ | ✅ (Canon, Nikon) |
| **GPS Data** | ❌ | ✅ (7 of 13) |
| **Temporal Data** | ❌ | ✅ (All seasons, times) |
| **Color Coverage** | Partial | ✅ (All 8 hues) |
| **Burst Testing** | ❌ | ✅ (1 group) |
| **Duplicate Testing** | ❌ | ✅ (1 pair) |
| **CI Friendly** | ✅ (fast, small) | ⚠️  (slow, large) |
| **Generation Time** | <1 second | ~15 seconds |

---

## Usage Guidelines

### For Unit Tests
**Use**: `testdata/photos/`
- Fast, lightweight
- No external dependencies
- Suitable for CI/CD pipelines

### For Integration Tests
**Use**: `testdata/photos/` for basic integration
- Quick feedback loop
- Tests core functionality

### For Complete Validation
**Use**: `testdata/dng/`
- Full facet coverage
- Metadata inference validation
- Burst/duplicate detection
- Performance testing

### For CI/CD
**Recommended**: `testdata/photos/` only
- Faster test execution
- Smaller repository size
- Optional: Add `testdata/dng/` download step for nightly builds

---

## Known Limitations

### Basic Fixtures
1. **No EXIF**: BMP and minimal JPEG EXIF
2. **Synthetic Content**: Simple gradients, not real photos
3. **Limited Dimensions**: Small file sizes

### Complete DNG Fixtures
1. **Not True DNG**: JPEG files with .dng extension
   - Works because indexer supports JPEG
   - All functionality validated

2. **Flash Detection**: ✅ Working with exif-go
   - Migrated from goexif to exif-go (dsoprea/go-exif/v3)
   - File #4 flash metadata now detected correctly
   - Complete coverage of all shooting conditions

3. **Synthetic Content**: Gradient images, not photographs
   - Sufficient for testing
   - Perceptual hashing and color extraction work correctly

4. **File Size**: 20MB vs 50-80MB for real DNGs
   - Large enough to stress-test
   - Smaller than actual camera RAW files

---

## Future Improvements

### Short Term
- [ ] Add tests for burst detection algorithm using fixtures 9-11
- [ ] Add tests for duplicate clustering using fixtures 12-13
- [ ] Document performance benchmarks with complete fixtures

### Medium Term
- [ ] Switch to EXIF library that supports Flash tag
- [ ] Add more burst groups (different time intervals)
- [ ] Add more duplicate clusters (varying similarity levels)

### Long Term
- [ ] Source real DNG files from camera manufacturers
- [ ] Add fixtures for edge cases (corrupted EXIF, extreme ISOs)
- [ ] Create fixtures for video file support (if added)

---

## Maintenance

### Regenerating Fixtures

**Basic fixtures** (when image content changes):
```bash
go run testdata/generate_fixtures.go
```

**Complete DNG fixtures** (when EXIF specs change):
```bash
rm -rf testdata/dng
go run testdata/generate_dng_fixtures.go
```

### Validation After Regeneration
```bash
# Run all tests
go test -v ./internal/indexer/

# Verify coverage
go run testdata/test_fixtures.go
go run testdata/verify_coverage.go /tmp/olsen_fixtures_test.db
```

---

## Summary

Olsen's test fixtures provide:

✅ **Complete facet coverage** with 13 DNG files covering all metadata dimensions
✅ **Fast iteration** with 4 basic files for quick testing
✅ **Format diversity** with JPEG and BMP support
✅ **Burst detection** testing with sequential photos
✅ **Duplicate detection** testing with similar photos
✅ **Real-world metadata** with Canon and Nikon EXIF
✅ **Geographic diversity** with GPS coordinates from 5 cities
✅ **Temporal diversity** with photos across all seasons and times
✅ **Color diversity** with all 8 hue categories represented

**Storage**: 254.3 MB total (254 MB DNG + 300 KB basic)
**Coverage**: 100% of faceted URL patterns (including flash detection)
