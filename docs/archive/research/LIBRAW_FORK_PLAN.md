# go-libraw Fork and Buffer Overflow Fix Plan

**Date**: 2025-10-12
**Issue**: Buffer overflow in `seppedelanghe/go-libraw` when processing JPEG-compressed DNG files
**Status**: Planning phase

## Executive Summary

The `seppedelanghe/go-libraw` library has a buffer overflow bug (line 403 of `libraw.go`) when processing JPEG-compressed RAW files. Despite this bug, we've chosen this library as our default because it provides:
- Full LibRaw configuration control
- Complete diagnostics capture
- Goroutine-safe design
- Better error messages

This document outlines our plan to fork the library, create a minimal reproducible test case, fix the bug, and contribute back upstream.

## Current Status

### Dual Library Support ✅
- **Build System**: Makefile targets for both libraries
  - `make build-golibraw` - Build with inokone/golibraw
  - `make build-seppedelanghe` - Build with seppedelanghe/go-libraw (default)
  - `make benchmark-libraw` - Benchmark both libraries
- **Runtime Detection**: `olsen version` command shows which library is active
- **Build Tags**:
  - No tag = inokone/golibraw
  - `use_seppedelanghe_libraw` = seppedelanghe/go-libraw

### Bug Identification ✅
**Location**: `github.com/seppedelanghe/go-libraw@v0.2.1/libraw.go:403`

**Symptom**:
```
panic: runtime error: index out of range [53248] with length 53248
```

**Trigger**: JPEG-compressed DNG files (Leica M11 Monochrom format)

**Error Message**:
```
convert to image: unexpected data size: got 60420096, want 181260288
```

## Phase 1: Minimal Reproducible Test Case

### Objective
Create the smallest possible test case that reliably triggers the buffer overflow.

### Approach

**1. Create Test File** (`internal/indexer/raw_test.go`)
```go
//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"testing"
	golibraw "github.com/seppedelanghe/go-libraw"
)

// TestBufferOverflow reproduces the buffer overflow with JPEG-compressed DNG
func TestBufferOverflow(t *testing.T) {
	// Use actual Leica M11 Monochrom file from private-testdata
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"

	processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
		UserQual:    3, // AHD
		OutputBps:   8,
		OutputColor: golibraw.SRGB,
		UseCameraWb: true,
	})

	// This should panic with buffer overflow
	_, _, err := processor.ProcessRaw(testFile)

	if err != nil {
		t.Logf("Expected error: %v", err)
	}
}
```

**2. Makefile Target**
```makefile
test-buffer-overflow: build-seppedelanghe
	@echo "Testing buffer overflow bug..."
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(CGO_CFLAGS_LIBRAW)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_LIBRAW)" \
	$(GOTEST) -tags "cgo use_seppedelanghe_libraw" -v ./internal/indexer -run TestBufferOverflow
```

**3. Document File Characteristics**
```bash
exiftool private-testdata/2024-12-23/L1001530.DNG | grep -i "compression\|photometric\|bits"
# Output:
# Compression: JPEG
# Photometric Interpretation: Linear Raw
# Bits Per Sample: 16
```

### Success Criteria
- Test reliably reproduces the panic
- Test runs in < 5 seconds
- Test file is checked into `testdata/` (or referenced from `private-testdata/`)

## Phase 2: Root Cause Analysis

### Investigation Steps

**1. Examine Buffer Calculation** (`libraw.go:370-410`)
```go
// Current code (buggy):
dataSize := height * width * channels * (bits / 8)
dataBytes := C.GoBytes(unsafe.Pointer(&dataPtr.data[0]), C.int(dataSize))
```

**Hypothesis**: The calculation assumes uncompressed RAW data size, but JPEG-compressed DNG has:
- Smaller actual buffer size (compressed)
- Different data layout
- `libraw_dcraw_ppm_tiff_writer` may need different handling

**2. Check LibRaw Documentation**
- Review `libraw_dcraw_ppm_tiff_writer` behavior with compressed DNGs
- Check if `libraw_unpack` handles JPEG compression correctly
- Investigate `imgdata.rawdata.raw_image` vs `imgdata.image` differences

**3. Compare with Working Code**
- Check how `inokone/golibraw` handles this case
- Review LibRaw examples for JPEG-compressed DNG handling
- Look for `LIBRAW_OPTIONS_NO_JPEG_ERRORS` or similar flags

### Expected Findings
The buffer size calculation doesn't account for:
1. JPEG compression (actual data size != uncompressed size)
2. Possible intermediate buffer requirements
3. Different memory layout for compressed vs uncompressed RAW

## Phase 3: Bug Fix Implementation

### Fix Strategy

**Option A: Use Correct Buffer Size from LibRaw** (Preferred)
```go
// Get actual data size from LibRaw structures
actualSize := C.libraw_get_mem_image_format(&proc.Processor)
dataBytes := C.GoBytes(unsafe.Pointer(&dataPtr.data[0]), C.int(actualSize))
```

**Option B: Detect Compression and Use Alternative Path**
```go
if isJPEGCompressed(proc) {
    // Use libraw_dcraw_mem_image instead
    return decodeViaMemImage(proc)
} else {
    // Use current path
    return decodeViaPPMWriter(proc)
}
```

**Option C: Bounds Checking with Fallback**
```go
dataSize := height * width * channels * (bits / 8)
if dataSize > len(dataPtr.data) {
    // Fallback: use actual buffer size
    dataSize = len(dataPtr.data)
    // or return error
}
```

### Implementation Plan

1. **Fork Repository**
   ```bash
   gh repo fork seppedelanghe/go-libraw
   git clone https://github.com/YOUR_USERNAME/go-libraw
   cd go-libraw
   git checkout -b fix/jpeg-compressed-dng-buffer-overflow
   ```

2. **Implement Fix** (`libraw.go:403`)
   - Add bounds checking
   - Use correct LibRaw API for buffer size
   - Add compression detection if needed

3. **Add Test Case**
   ```go
   // libraw_test.go
   func TestJPEGCompressedDNG(t *testing.T) {
       // Test with Leica M11 Monochrom file
       // Should not panic
   }
   ```

4. **Update Documentation**
   - Add comment explaining JPEG-compressed DNG handling
   - Document any limitations
   - Update README with supported formats

## Phase 4: Testing and Validation

### Test Matrix

| File Type | Camera | Compression | Expected Result |
|-----------|--------|-------------|-----------------|
| Uncompressed DNG | Canon R5 | None | ✅ Works |
| Uncompressed DNG | Nikon Z9 | None | ✅ Works |
| JPEG-compressed DNG | Leica M11 Monochrom | JPEG | 🐛 Panics (to fix) |
| Lossless compressed DNG | Various | Lossless JPEG | ❓ Unknown |
| CR2 | Canon | Various | ✅ Should work |
| NEF | Nikon | Various | ✅ Should work |

### Validation Steps

1. **Unit Tests**
   ```bash
   go test -v ./...
   ```

2. **Integration Test**
   ```bash
   make test-buffer-overflow  # Should pass after fix
   ```

3. **Benchmark Comparison**
   ```bash
   make benchmark-libraw  # Should complete without panic
   ```

4. **Full Indexing Test**
   ```bash
   make build-seppedelanghe
   ./bin/olsen index private-testdata --db test.db
   ```

### Success Criteria
- ✅ No panics on JPEG-compressed DNG
- ✅ Decoded images match quality expectations
- ✅ All existing tests still pass
- ✅ Performance impact < 5%

## Phase 5: Upstream Contribution

### Pull Request Plan

**1. PR Description**
```markdown
## Fix buffer overflow with JPEG-compressed DNG files

### Problem
When processing JPEG-compressed DNG files (e.g., Leica M11 Monochrom),
the library panics with:
```
panic: runtime error: index out of range [53248] with length 53248
```

### Root Cause
Buffer size calculation in `ProcessRaw()` (line 403) assumes uncompressed
RAW data size:
```go
dataSize := height * width * channels * (bits / 8)
```

For JPEG-compressed DNGs, the actual buffer size is smaller due to
compression, causing out-of-bounds access.

### Solution
[Describe fix - TBD based on Phase 3]

### Testing
- Added test case with JPEG-compressed DNG file
- Verified fix with Leica M11 Monochrom files
- All existing tests pass
- No performance regression

### Breaking Changes
None - this is a bug fix that makes previously-broken files work.
```

**2. Include Test Data**
- Either include minimal JPEG-compressed DNG in test suite
- Or document how to generate test file
- Reference public sample files if available

**3. Documentation Updates**
- Update README to mention JPEG-compressed DNG support
- Add to supported formats list
- Include example code if API changed

## Phase 6: Olsen Integration

### After Fix is Merged

**1. Update Dependency**
```bash
go get github.com/seppedelanghe/go-libraw@v0.2.2  # Or our fork until merged
go mod tidy
```

**2. Remove Workarounds**
- Keep dual library support (valuable for comparison)
- Document that seppedelanghe is now fully working
- Update CLAUDE.md to remove bug warnings

**3. Final Validation**
```bash
make benchmark-libraw          # Both libraries should work
make test-integration-raw      # Full integration test
```

**4. Update Documentation**
```markdown
# docs/LIBRAW_MIGRATION_COMPLETE.md
- Document the bug fix
- Benchmark results showing equivalence
- Recommendation to use seppedelanghe/go-libraw
```

## Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Minimal Test Case | 1 hour | Real DNG file access |
| Phase 2: Root Cause Analysis | 2-4 hours | LibRaw documentation |
| Phase 3: Fix Implementation | 2-6 hours | Understanding LibRaw internals |
| Phase 4: Testing | 2 hours | Multiple test files |
| Phase 5: PR & Review | 1-2 weeks | Maintainer response time |
| Phase 6: Integration | 1 hour | PR merged |

**Total Estimated Time**: 8-14 hours of work + 1-2 weeks waiting

## Risks and Mitigation

### Risk 1: Fix is Complex
**Impact**: High
**Probability**: Medium
**Mitigation**:
- Start with simpler bound-checking approach
- Fall back to alternative decoding path if needed
- Consult LibRaw documentation and examples

### Risk 2: Upstream Maintainer Unresponsive
**Impact**: Medium
**Probability**: Low
**Mitigation**:
- Maintain our own fork
- Document fork location in go.mod
- Revisit upstream periodically

### Risk 3: Fix Breaks Other Formats
**Impact**: High
**Probability**: Low
**Mitigation**:
- Comprehensive test matrix
- Add format detection to isolate fix
- Keep dual library support as safety net

### Risk 4: Performance Regression
**Impact**: Medium
**Probability**: Low
**Mitigation**:
- Benchmark before/after
- Profile hot paths
- Optimize if needed

## Success Metrics

1. ✅ Minimal test case created and passing
2. ✅ Root cause identified and documented
3. ✅ Fix implemented and tested locally
4. ✅ PR submitted with test case
5. ✅ No panics on any DNG format
6. ✅ Performance within 5% of baseline
7. ✅ All quality metrics maintained

## References

- **Bug Location**: `github.com/seppedelanghe/go-libraw@v0.2.1/libraw.go:403`
- **Test Files**: `private-testdata/2024-12-23/L1001530.DNG` (Leica M11 Monochrom)
- **LibRaw Docs**: https://www.libraw.org/docs/API-overview.html
- **Fork Plan**: This document

## Next Steps

1. Create minimal reproducible test case (Phase 1)
2. Run `make test-buffer-overflow` to confirm panic
3. Analyze LibRaw buffer handling (Phase 2)
4. Implement fix with bounds checking (Phase 3)
5. Submit PR with test case (Phase 5)
6. Update Olsen after merge (Phase 6)

---

**Author**: Claude Code
**Reviewer**: (TBD)
**Approved**: (TBD)
