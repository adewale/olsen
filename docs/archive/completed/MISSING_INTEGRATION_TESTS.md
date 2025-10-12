# Missing Integration Tests Analysis

**Date**: 2025-10-12
**Purpose**: Identify and implement missing end-to-end integration tests

## Existing Integration Tests

✅ **Indexer Tests** (`internal/indexer/*_integration_test.go`):
- `TestIntegrationIndexTestData` - Basic indexing
- `TestIntegrationReIndexing` - Re-indexing behavior
- `TestIntegrationFileTypeSupport` - Different file types
- `TestIntegrationThumbnailGeneration` - Thumbnail creation
- `TestIntegrationColorExtraction` - Color palette extraction
- `TestIntegrationMonochromeDNG` - Monochrome RAW support
- `TestIntegrationMonochromeBatch` - Batch monochrome processing

✅ **Query Tests** (`internal/query/*_integration_test.go`):
- `TestColorQueryIntegration` - Color filtering
- `TestStateMachineIntegration` - Faceted navigation

## Missing Integration Tests

### 1. ❌ Burst Detection End-to-End
**Command**: `olsen analyze --db photos.db`

**What's missing**:
- No test that indexes photos with temporal sequences
- No test that runs burst detection
- No verification that burst_groups table is populated correctly
- No test for burst representative selection

**User journey not tested**:
```bash
./bin/olsen index photos/ --db test.db
./bin/olsen analyze --db test.db
./bin/olsen bursts --db test.db
./bin/olsen burst abc123 --db test.db
```

### 2. ❌ Web Explorer End-to-End
**Command**: `olsen explore --db photos.db --addr localhost:8080`

**What's missing**:
- No test that starts the server
- No test that queries the HTML endpoints
- No test that thumbnail serving works
- No test that faceted navigation UI works
- No test for cache-busting headers

**User journey not tested**:
```bash
./bin/olsen index photos/ --db test.db
./bin/olsen explore --db test.db &
curl http://localhost:8080/
curl http://localhost:8080/photo/1/thumbnail?size=256
```

### 3. ❌ Query Command End-to-End
**Command**: `olsen query --db photos.db --year 2025 --color blue`

**What's missing**:
- No test of CLI query command
- No test of facet display in CLI
- No test of pagination
- No test of different output formats (table, json, ids)

**User journey not tested**:
```bash
./bin/olsen index photos/ --db test.db
./bin/olsen query --db test.db --year 2025 --facets
./bin/olsen query --db test.db --color blue --limit 10 --format ids
```

### 4. ❌ Stats Command End-to-End
**Command**: `olsen stats --db photos.db`

**What's missing**:
- No test that stats are computed correctly
- No test for top cameras display
- No test for date range display

**User journey not tested**:
```bash
./bin/olsen index photos/ --db test.db
./bin/olsen stats --db test.db
```

### 5. ❌ Show Command End-to-End
**Command**: `olsen show 42 --db photos.db`

**What's missing**:
- No test that photo details are displayed
- No test for missing photo ID
- No test for GPS/EXIF display

### 6. ❌ Thumbnail Command End-to-End
**Command**: `olsen thumbnail 42 -s 256 -o thumb.jpg --db photos.db`

**What's missing**:
- No test that thumbnail extraction works
- No test for different sizes
- No test for stdout vs file output
- No test for missing photo/size

### 7. ❌ Verify Command End-to-End
**Command**: `olsen verify --db photos.db`

**What's missing**:
- No test for database integrity checks
- No test for missing files detection
- No test for orphaned records detection
- No test for `--facets` flag

### 8. ❌ Compact Command End-to-End
**Command**: `olsen compact --db photos.db`

**What's missing**:
- No test that VACUUM works
- No test that ANALYZE works
- No test for size reduction reporting

### 9. ❌ RAW Processing with Different Settings
**What's missing**:
- No test for different demosaic algorithms (Linear, VNG, PPG, AHD)
- No test for different bit depths (8-bit vs 16-bit)
- No test for different color spaces (sRGB, AdobeRGB)
- No test for half-size processing

### 10. ❌ Error Conditions
**What's missing**:
- No test for corrupt files
- No test for missing EXIF data
- No test for unsupported file types
- No test for disk full conditions
- No test for database locked conditions
- No test for permission errors

### 11. ❌ Re-indexing Scenarios
**What's missing**:
- No test for modified files (updated timestamp)
- No test for moved files (same hash, different path)
- No test for deleted files cleanup

### 12. ❌ Performance/Stress Testing
**What's missing**:
- No test with 1000+ photos
- No test with very large RAW files (50MB+)
- No test with concurrent web explorer requests
- No test for memory usage limits

## Priority for Implementation

### High Priority (Blocking user workflows)
1. **Burst Detection End-to-End** - Core feature not tested
2. **Web Explorer End-to-End** - Primary UI not tested
3. **Query Command End-to-End** - CLI query interface not tested

### Medium Priority (Important but workflows exist)
4. **Stats Command** - Simple reporting, less critical
5. **Show Command** - Single photo display
6. **Thumbnail Command** - Thumbnail extraction
7. **Error Conditions** - Edge cases

### Low Priority (Advanced features)
8. **Verify Command** - Admin tool
9. **Compact Command** - Maintenance tool
10. **RAW Processing Settings** - Advanced configuration
11. **Performance Testing** - Optimization concern

## Test Implementation Plan

For each missing test, implement:

1. **Setup phase**:
   - Create temp database
   - Index test photos
   - Create any necessary state

2. **Execution phase**:
   - Run the actual CLI command or equivalent API call
   - Capture output/errors

3. **Verification phase**:
   - Check database state
   - Check file outputs
   - Verify expected results

4. **Cleanup phase**:
   - Remove temp files
   - Close connections

## Example Test Structure

```go
func TestIntegrationBurstDetection(t *testing.T) {
    // Setup
    dbPath := filepath.Join(t.TempDir(), "test.db")
    db, _ := database.Open(dbPath)
    defer db.Close()
    db.InitSchema()

    // Index photos with temporal sequences
    indexer := indexer.NewIndexer(db, 1, false)
    indexer.IndexDirectory("testdata/burst-sequence")

    // Execute burst detection
    detector := indexer.NewBurstDetector(db)
    bursts, err := detector.DetectBursts()
    if err != nil {
        t.Fatalf("Burst detection failed: %v", err)
    }

    // Verify results
    if len(bursts) == 0 {
        t.Error("Expected burst groups, got none")
    }

    // Check database state
    var count int
    db.QueryRow("SELECT COUNT(*) FROM burst_groups").Scan(&count)
    if count != len(bursts) {
        t.Errorf("Expected %d burst groups in DB, got %d", len(bursts), count)
    }

    // Verify burst properties
    for _, burst := range bursts {
        if len(burst) < 3 {
            t.Errorf("Burst too small: %d photos", len(burst))
        }
    }
}
```

## Test Data Requirements

To implement these tests, we need:

1. **Burst sequence testdata**: 3+ photos within 2 seconds
2. **Multi-camera testdata**: Photos from different cameras
3. **Multi-year testdata**: Photos spanning multiple years
4. **Corrupt file testdata**: Intentionally broken JPEG/DNG
5. **Large file testdata**: 50MB+ RAW files (or use existing)

## Success Criteria

Tests are complete when:
- [ ] All CLI commands have integration tests
- [ ] All user workflows are tested end-to-end
- [ ] Error conditions are tested
- [ ] Tests run in < 60 seconds total
- [ ] Tests are documented and maintainable
