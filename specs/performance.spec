# Olsen Indexer Performance Specification

**Version**: 1.0
**Date**: 2025-10-06
**Status**: Implemented

## Overview

This document specifies the performance instrumentation and tracking capabilities of the Olsen photo indexer. The performance tracking system provides detailed timing information for each stage of the indexing pipeline, enabling developers to identify bottlenecks and optimize processing.

## Goals

1. **Visibility**: Provide detailed timing information for every stage of photo processing
2. **Actionability**: Identify performance bottlenecks to guide optimization efforts
3. **Measurability**: Track performance improvements over time
4. **Debuggability**: Diagnose performance issues in production environments
5. **Zero Overhead**: Performance tracking should have minimal overhead when disabled

## Architecture

### Pipeline Stages

The indexing pipeline consists of 8 instrumented stages:

| Stage | Description | Typical % of Total Time |
|-------|-------------|-------------------------|
| **Hash** | SHA-256 file hashing for change detection | ~1% |
| **Metadata** | EXIF/RAW metadata extraction | ~1% |
| **Image Decode** | RAW decode or standard image decode | ~40-45% |
| **Thumbnails** | Generate 4 thumbnail sizes (64, 256, 512, 1024px) | ~50-55% |
| **Color** | Extract 5 dominant colors via k-means | ~4-5% |
| **Perceptual Hash** | Compute pHash for similarity detection | <1% |
| **Inference** | Infer time of day, season, focal category, shooting condition | <1% |
| **Database** | Insert metadata and thumbnails into SQLite | <1% |

### Data Structures

#### PerfStats (Per-Photo Metrics)
```go
type PerfStats struct {
    FilePath          string        // Full path to photo file
    TotalTime         time.Duration // End-to-end processing time
    HashTime          time.Duration // File hash calculation
    MetadataTime      time.Duration // EXIF/RAW extraction
    ImageDecodeTime   time.Duration // Image decoding
    ThumbnailTime     time.Duration // Thumbnail generation
    ColorTime         time.Duration // Color extraction
    PerceptualHashTime time.Duration // pHash computation
    InferenceTime     time.Duration // Metadata inference
    DatabaseTime      time.Duration // Database insertion
    FileSize          int64         // File size in bytes
    WasSkipped        bool          // File unchanged, skipped
    WasUpdated        bool          // File modified, re-indexed
    Error             string        // Error message if failed
}
```

#### PerfSummary (Aggregate Metrics)
```go
type PerfSummary struct {
    // Counts
    TotalPhotos       int
    ProcessedPhotos   int
    SkippedPhotos     int
    UpdatedPhotos     int
    FailedPhotos      int

    // Cumulative times
    TotalTime         time.Duration
    HashTime          time.Duration
    MetadataTime      time.Duration
    ImageDecodeTime   time.Duration
    ThumbnailTime     time.Duration
    ColorTime         time.Duration
    PerceptualHashTime time.Duration
    InferenceTime     time.Duration
    DatabaseTime      time.Duration

    TotalBytes        int64

    // Running averages (milliseconds)
    AvgTotalMs         float64
    AvgHashMs          float64
    AvgMetadataMs      float64
    AvgImageDecodeMs   float64
    AvgThumbnailMs     float64
    AvgColorMs         float64
    AvgPerceptualHashMs float64
    AvgInferenceMs     float64
    AvgDatabaseMs      float64

    AvgThroughputMBps float64  // Overall throughput
}
```

## Usage

### Enabling Performance Tracking

```bash
# Enable with --perfstats flag
olsen index --db photos.db --w 4 --perfstats ~/Pictures

# With explicit parameters
olsen index --db mydb.db --w 8 --perfstats /path/to/photos
```

### Output Format

#### Console Output

The console output is human-friendly but machine-parseable:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 PERFORMANCE STATISTICS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SUMMARY:
  Total Photos:      13
  Processed:         13
  Skipped:           0 (unchanged)
  Updated:           0 (re-indexed)
  Failed:            0
  Total Data:        253.85 MB
  Throughput:        16.73 MB/s

AVERAGE TIMINGS PER PHOTO (processed only):
  Total:              1166.92 ms  (100.00%)
  Hash:                  8.69 ms  (  0.74%)
  Metadata:              9.54 ms  (  0.82%)
  Image Decode:        490.08 ms  ( 42.00%)
  Thumbnails:          604.92 ms  ( 51.84%)
  Color Extraction:     51.69 ms  (  4.43%)
  Perceptual Hash:       0.85 ms  (  0.07%)
  Inference:             0.00 ms  (  0.00%)
  Database:              0.77 ms  (  0.07%)

PIPELINE BREAKDOWN (by time %):
  Hash         [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0.74%
  Metadata     [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0.82%
  Decode       [████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  42.00%
  Thumbnails   [█████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░]  51.84%
  Color        [██░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   4.43%
  PHash        [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0.07%
  Inference    [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0.00%
  Database     [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0.07%

TOP 10 SLOWEST PHOTOS:
   1. 03_nikon_z9_85mm_summer_midday_iso3200_yellow_g...   1237.00 ms
   2. 02_canon_r5_50mm_summer_morning_iso800_orange_n...   1230.00 ms
   3. 06_nikon_z9_50mm_winter_blue_hour_iso1600_blue_...   1211.00 ms
   ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### JSON Export

Performance data is automatically exported to a timestamped JSON file:

**Filename**: `perfstats_YYYYMMDD_HHMMSS.json`

**Structure**:
```json
{
  "summary": {
    "TotalPhotos": 13,
    "ProcessedPhotos": 13,
    "SkippedPhotos": 0,
    "UpdatedPhotos": 0,
    "FailedPhotos": 0,
    "TotalTime": 15170196665,
    "HashTime": 113966416,
    "MetadataTime": 124478458,
    "ImageDecodeTime": 6371674252,
    "ThumbnailTime": 7864875459,
    "ColorTime": 672205334,
    "PerceptualHashTime": 11428624,
    "InferenceTime": 8125,
    "DatabaseTime": 10255998,
    "TotalBytes": 266184424,
    "AvgTotalMs": 1166.92,
    "AvgHashMs": 8.69,
    "AvgMetadataMs": 9.54,
    "AvgImageDecodeMs": 490.08,
    "AvgThumbnailMs": 604.92,
    "AvgColorMs": 51.69,
    "AvgPerceptualHashMs": 0.85,
    "AvgInferenceMs": 0.00,
    "AvgDatabaseMs": 0.77,
    "AvgThroughputMBps": 16.73
  },
  "detailed": [
    {
      "FilePath": "testdata/dng/01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng",
      "TotalTime": 1190742250,
      "HashTime": 8526542,
      "MetadataTime": 14826625,
      "ImageDecodeTime": 481384334,
      "ThumbnailTime": 625326541,
      "ColorTime": 59172959,
      "PerceptualHashTime": 294959,
      "InferenceTime": 583,
      "DatabaseTime": 847208,
      "FileSize": 20490959,
      "WasSkipped": false,
      "WasUpdated": false,
      "Error": ""
    },
    ...
  ]
}
```

## Performance Characteristics

### Observed Bottlenecks (From Test Data)

Based on actual measurements from the test suite:

1. **Thumbnail Generation (51.84%)** - PRIMARY BOTTLENECK
   - Generating 4 sizes per image
   - Lanczos3 resampling is high quality but slow
   - JPEG encoding at quality 85

2. **Image Decoding (42.00%)** - SECONDARY BOTTLENECK
   - RAW file decoding via LibRaw
   - Embedded JPEG extraction fallback
   - Standard image decode for non-RAW files

3. **Color Extraction (4.43%)**
   - K-means clustering with 100 iterations
   - Processing 256px thumbnail
   - 5 dominant colors extracted

4. **Everything Else (<2% combined)**
   - Hash, metadata, pHash, inference, database operations are well-optimized

### Throughput Expectations

| Scenario | Expected Throughput | Notes |
|----------|---------------------|-------|
| **Standard JPEG** | 3-5 photos/sec | No RAW decoding overhead |
| **DNG with embedded preview** | 1-2 photos/sec | Current test measurement |
| **Full RAW decode** | 0.5-1 photo/sec | Full LibRaw pipeline |
| **Cached/unchanged** | 100-500 photos/sec | Hash check only |

### Worker Scaling

| Workers | Expected Speedup | Notes |
|---------|------------------|-------|
| 1 | 1.0x (baseline) | Single-threaded |
| 2 | ~1.8x | Good scaling |
| 4 | ~3.2x | Diminishing returns |
| 8 | ~4.5x | I/O bound limits |
| 16+ | ~5x | Likely I/O bound |

## Optimization Opportunities

### High Impact (>10% potential improvement)

1. **Parallelize thumbnail generation**
   ```go
   // Current: Sequential generation of 4 sizes
   // Potential: Parallel generation with goroutines
   // Expected gain: 30-40% reduction in thumbnail time
   ```

2. **Use faster resampling algorithm**
   ```go
   // Current: Lanczos3 (highest quality)
   // Alternative: Bilinear or Box (faster)
   // Expected gain: 20-30% reduction in thumbnail time
   // Trade-off: Slightly lower quality
   ```

3. **Cache decoded images**
   ```go
   // Current: Decode once, then decode thumbnail again for color
   // Potential: Reuse decoded thumbnail
   // Expected gain: Eliminate redundant decode (~5% total time)
   ```

### Medium Impact (3-10% potential improvement)

4. **Optimize color extraction**
   ```go
   // Current: 100 k-means iterations
   // Potential: Reduce to 50 iterations or use faster algorithm
   // Expected gain: 2-3% of total time
   ```

5. **Batch database inserts**
   ```go
   // Current: One transaction per photo
   // Potential: Batch N photos per transaction
   // Expected gain: Minor but reduces lock contention
   ```

### Low Impact (<3% potential improvement)

6. **Use xxHash instead of SHA-256**
   - Faster non-cryptographic hash for change detection
   - Expected gain: <1% of total time

7. **Optimize EXIF parsing**
   - Use streaming parser instead of loading full file
   - Expected gain: <1% of total time

## Monitoring and Analysis

### Key Metrics to Track

1. **Average total time per photo** - Overall system performance
2. **Thumbnail time percentage** - Primary bottleneck indicator
3. **Decode time percentage** - Secondary bottleneck indicator
4. **Throughput (MB/s)** - Hardware utilization indicator
5. **Top slowest photos** - Outlier detection

### Performance Regression Detection

Compare JSON exports over time:

```bash
# Example: Compare two performance runs
jq '.summary.AvgTotalMs' perfstats_20251006_131024.json
jq '.summary.AvgTotalMs' perfstats_20251007_095500.json

# Check for regressions (>10% slower)
```

### Production Monitoring

When running in production:

```bash
# Index with performance tracking
olsen index --db prod.db --w 8 --perfstats /mnt/photos

# Analyze results
cat perfstats_*.json | jq '.summary | {
  photos: .ProcessedPhotos,
  avg_ms: .AvgTotalMs,
  throughput: .AvgThroughputMBps,
  bottleneck: (
    if .AvgThumbnailMs > .AvgImageDecodeMs
    then "thumbnails"
    else "decode"
    end
  )
}'
```

## Implementation Details

### Code Locations

- **Data structures**: `pkg/models/types.go:139-189`
- **Engine instrumentation**: `internal/indexer/indexer.go:27-36, 170-375`
- **Performance output**: `internal/indexer/perfoutput.go`
- **CLI integration**: `cmd/olsen/main.go:188, 232-234, 288-298`

### Zero-Overhead Design

When `--perfstats` is NOT enabled:
- No memory allocation for PerfStats slices
- No timing calls (`time.Now()` not called)
- No lock contention for updating summary
- Zero runtime overhead

When `--perfstats` IS enabled:
- Minimal overhead (<1% impact on total time)
- One `time.Now()` call per pipeline stage
- Lock contention only for summary updates (infrequent)

## Testing

### Verification

Run performance tracking on test data:

```bash
# Index test photos with performance tracking
make build-raw
./bin/olsen index --db test_perf.db --w 2 --perfstats testdata/dng

# Verify output
# - Console shows summary, averages, breakdown, top 10
# - JSON file created with timestamp
# - All timing values are positive
# - Percentages sum to ~100%
```

### Continuous Monitoring

Add to CI/CD:

```bash
# Baseline performance measurement
./bin/olsen index --db ci.db --w 4 --perfstats testdata/dng
BASELINE=$(jq '.summary.AvgTotalMs' perfstats_*.json)

# Fail if regression > 20%
if [ "$CURRENT" -gt $(echo "$BASELINE * 1.2" | bc) ]; then
  echo "Performance regression detected!"
  exit 1
fi
```

## Future Enhancements

### Planned

1. **Per-worker statistics** - Identify worker imbalance
2. **Memory usage tracking** - Monitor peak memory consumption
3. **Disk I/O metrics** - Measure read/write bandwidth
4. **Network timing** (if applicable) - For remote storage
5. **GPU acceleration metrics** - If GPU decode is added

### Under Consideration

1. **Real-time dashboard** - Live performance visualization
2. **Historical trending** - Track performance over weeks/months
3. **Automatic optimization hints** - Suggest configuration changes
4. **Profiling integration** - Generate pprof profiles on demand

## References

- Go standard library: `time` package for high-resolution timing
- Lanczos3 resampling: `github.com/nfnt/resize`
- K-means clustering: `github.com/mccutchen/palettor`
- LibRaw documentation: https://www.libraw.org/docs

## Changelog

### Version 1.0 (2025-10-06)
- Initial implementation
- 8 pipeline stages instrumented
- Console and JSON output
- Zero-overhead design when disabled
- Top 10 slowest photos report
- Identified thumbnail generation as primary bottleneck (51.84%)
- Identified image decode as secondary bottleneck (42.00%)
