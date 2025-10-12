# LibRaw Integration Plan for Olsen

**Status:** Planning Phase
**Start Date:** 2025-10-05
**Target Completion:** TBD
**Owner:** Development Team

---

## Executive Summary

Olsen currently cannot process RAW DNG files (e.g., Leica M11 Monochrom) due to Go's standard TIFF decoder lacking support for "Linear Raw" photometric interpretation. This plan outlines the steps to integrate LibRaw support, enabling Olsen to index, extract metadata, generate thumbnails, and analyze colours from RAW files.

---

## Goals

### Primary Goals
1. ✅ Successfully index all RAW DNG files in private-testdata (30 Leica M11 Monochrom files)
2. ✅ Extract EXIF metadata from RAW files
3. ✅ Generate thumbnails from RAW sensor data
4. ✅ Extract colour palettes from demosaiced RAW images
5. ✅ Maintain backward compatibility with existing JPEG/PNG/BMP support

### Secondary Goals
- Support other RAW formats (CR2, NEF, RAF, ARW, etc.)
- Make CGO/LibRaw dependency optional via build tags
- Document installation and deployment requirements
- Add configuration options for RAW processing quality vs. speed

---

## Technical Approach

### Selected Library: `github.com/inokone/golibraw`

**Rationale:**
- Simple, clean API (`ExtractMetadata`, `ExtractThumbnail`, `ImportRaw`)
- Returns standard Go `image.Image` type
- Active development
- Well-documented installation process

**Alternative considered:** `github.com/seppedelanghe/go-libraw` (more recent, but similar capabilities)

---

## Implementation Plan

### Phase 1: Environment Setup & Proof of Concept
**Duration:** 1-2 days
**Status:** 🟡 Not Started

#### Tasks:
- [ ] Install LibRaw on development machine
  - macOS: `brew install libraw`
  - Linux: `sudo apt-get install libraw-dev`
- [ ] Verify LibRaw installation (`pkg-config --modversion libraw`)
- [ ] Create proof-of-concept Go program to test `golibraw`
- [ ] Test with single Leica M11 DNG file from private-testdata
- [ ] Verify successful image decoding and metadata extraction

**Success Criteria:**
- Can successfully import RAW DNG file to `image.Image`
- Can extract basic metadata (camera, lens, ISO, aperture)
- No segfaults or memory issues

**Deliverables:**
- `docs/raw-support/poc.go` - Working proof of concept
- `docs/raw-support/SETUP.md` - Installation instructions
- Test results documented

---

### Phase 2: Architecture Design
**Duration:** 1 day
**Status:** 🟡 Not Started

#### Tasks:
- [ ] Design RAW decoder abstraction layer
- [ ] Plan build tag strategy (`//go:build cgo`)
- [ ] Design graceful fallback for non-CGO builds
- [ ] Document RAW processing pipeline
- [ ] Define configuration options (processing quality, speed trade-offs)

**Key Decisions to Make:**
1. **Where to integrate LibRaw?**
   - Option A: Extend `internal/indexer/indexer.go` with conditional imports
   - Option B: Create separate `internal/indexer/raw.go` with build tags
   - **Recommended:** Option B for cleaner separation

2. **Fallback Strategy:**
   - Option A: Skip RAW files entirely if LibRaw unavailable
   - Option B: Extract embedded thumbnails only
   - Option C: Extract EXIF but skip thumbnails/colours
   - **Recommended:** Option C (graceful degradation)

3. **Image Processing:**
   - Use LibRaw's demosaiced output (slower, better quality)
   - Use LibRaw's half-size mode (faster, lower quality)
   - **Recommended:** Configurable via flag

**Deliverables:**
- `docs/raw-support/ARCHITECTURE.md` - Technical design document
- Decision log for key architectural choices

---

### Phase 3: Core Implementation
**Duration:** 3-5 days
**Status:** 🟡 Not Started

#### Task 3.1: Create RAW Decoder Module
- [ ] Create `internal/indexer/raw.go` with build tag `//go:build cgo`
- [ ] Implement RAW decoder interface:
  ```go
  type RawDecoder interface {
      DecodeRaw(path string) (image.Image, error)
      ExtractRawMetadata(path string) (*RawMetadata, error)
  }
  ```
- [ ] Implement LibRaw-based decoder
- [ ] Add comprehensive error handling

#### Task 3.2: Integrate with Indexer
- [ ] Modify `processFile()` in `internal/indexer/indexer.go`
- [ ] Add RAW file detection (.dng, .cr2, .nef, .raf, .arw extensions)
- [ ] Route RAW files to RAW decoder
- [ ] Ensure EXIF extraction works with RAW files
- [ ] Generate thumbnails from decoded RAW images
- [ ] Extract colour palettes from demosaiced images

#### Task 3.3: Add Fallback for Non-CGO Builds
- [ ] Create `internal/indexer/raw_nocgo.go` with build tag `//go:build !cgo`
- [ ] Implement stub functions that extract EXIF only
- [ ] Add clear logging when RAW support unavailable
- [ ] Document limitations in log messages

**Deliverables:**
- Working RAW decoder module
- Integration tests passing
- Build working with and without CGO

---

### Phase 4: Testing & Validation
**Duration:** 2-3 days
**Status:** 🟡 Not Started

#### Task 4.1: Unit Tests
- [ ] Test RAW decoder with various formats (DNG, CR2, NEF, RAF)
- [ ] Test error handling (corrupt files, unsupported formats)
- [ ] Test memory management (no leaks)
- [ ] Test CGO vs non-CGO builds

#### Task 4.2: Integration Tests
- [ ] Update `TestIntegrationIndexPrivateTestData` to expect success
- [ ] Verify all 30 Leica M11 files index successfully
- [ ] Verify thumbnails generated correctly
- [ ] Verify colour extraction works
- [ ] Verify metadata completeness

#### Task 4.3: Performance Testing
- [ ] Benchmark RAW indexing speed
- [ ] Compare with JPEG indexing performance
- [ ] Test with large RAW files (>50MB)
- [ ] Identify any performance bottlenecks

**Success Criteria:**
- All 30 private-testdata files index successfully
- No memory leaks detected
- Performance acceptable (>1 photo/second for RAW)
- All existing tests still pass

**Deliverables:**
- Comprehensive test suite
- Performance benchmarks documented
- Updated integration test showing success

---

### Phase 5: Documentation & Deployment
**Duration:** 1-2 days
**Status:** 🟡 Not Started

#### Task 5.1: User Documentation
- [ ] Update README.md with RAW support information
- [ ] Document LibRaw installation requirements
- [ ] Add supported RAW formats list
- [ ] Document CGO requirement
- [ ] Provide troubleshooting guide

#### Task 5.2: Developer Documentation
- [ ] Document RAW processing pipeline
- [ ] Add architecture diagrams
- [ ] Document build tag strategy
- [ ] Add CGO compilation notes

#### Task 5.3: Deployment Guide
- [ ] Document production deployment requirements
- [ ] Provide Docker image with LibRaw pre-installed
- [ ] Add CI/CD pipeline updates for CGO builds
- [ ] Document cross-compilation process

**Deliverables:**
- Updated README.md
- `docs/raw-support/INSTALLATION.md`
- `docs/raw-support/TROUBLESHOOTING.md`
- Dockerfile with LibRaw support

---

### Phase 6: Optional Enhancements
**Duration:** Variable
**Status:** 🟡 Future

#### Potential Enhancements:
- [ ] Add RAW processing quality settings (fast vs. quality)
- [ ] Support for sidecar XMP files
- [ ] Caching of demosaiced images
- [ ] Parallel RAW processing optimization
- [ ] Support for RAW+JPEG pairs
- [ ] Custom white balance adjustments
- [ ] Exposure compensation during processing
- [ ] Support for additional RAW formats (ORF, RW2, PEF, etc.)

---

## Dependencies & Requirements

### System Requirements
- **LibRaw library** (0.20.0 or later recommended)
  - macOS: `brew install libraw`
  - Ubuntu/Debian: `sudo apt-get install libraw-dev`
  - Fedora/RHEL: `sudo dnf install LibRaw-devel`
- **CGO enabled** (default in Go)
- **C/C++ compiler** (gcc, clang)
- **pkg-config** (for finding LibRaw)

### Go Dependencies
- `github.com/inokone/golibraw` (MIT License)
- LibRaw indirect dependency (LGPL v2.1 or CDDL v1.0)

### Licensing Considerations
- LibRaw is dual-licensed: LGPL v2.1 or CDDL v1.0
- Olsen can choose CDDL v1.0 to avoid LGPL requirements
- Dynamic linking recommended for LGPL compliance
- Document license choice in LICENSE file

---

## Risks & Mitigations

### Risk 1: CGO Complexity
**Impact:** High
**Probability:** Medium
**Mitigation:**
- Use build tags to make CGO optional
- Provide Docker images with pre-built binaries
- Document installation thoroughly
- Provide pre-compiled binaries for common platforms

### Risk 2: Cross-Platform Compatibility
**Impact:** Medium
**Probability:** Medium
**Mitigation:**
- Test on macOS, Linux, Windows
- Use GitHub Actions for multi-platform CI
- Document platform-specific issues
- Consider conditional compilation per platform

### Risk 3: Performance Impact
**Impact:** Medium
**Probability:** Low
**Mitigation:**
- Benchmark early and often
- Make RAW processing optional
- Add configuration for quality vs. speed
- Consider caching decoded images

### Risk 4: Memory Usage
**Impact:** High
**Probability:** Low
**Mitigation:**
- Profile memory usage regularly
- Test with large RAW files
- Implement proper resource cleanup
- Add memory limits if needed

### Risk 5: LibRaw Availability
**Impact:** Low
**Probability:** Low
**Mitigation:**
- Document installation clearly
- Provide Docker images
- Create fallback for non-CGO builds
- Consider vendoring LibRaw (complex)

---

## Success Metrics

### Functional Success
- ✅ All 30 Leica M11 DNG files index successfully
- ✅ Thumbnails generated for all RAW files
- ✅ Colour palettes extracted accurately
- ✅ EXIF metadata extracted completely
- ✅ No crashes or memory leaks
- ✅ Graceful fallback when LibRaw unavailable

### Performance Success
- ✅ RAW indexing > 1 photo/second
- ✅ Memory usage < 500MB per worker
- ✅ No significant slowdown for JPEG/PNG
- ✅ Parallel processing scales linearly

### Quality Success
- ✅ Thumbnails visually acceptable
- ✅ Colours accurately represent RAW content
- ✅ Metadata extraction >95% complete
- ✅ No data loss or corruption

---

## Timeline

| Phase | Duration | Start | End | Status |
|-------|----------|-------|-----|--------|
| 1. Environment Setup & PoC | 1-2 days | TBD | TBD | 🟡 Not Started |
| 2. Architecture Design | 1 day | TBD | TBD | 🟡 Not Started |
| 3. Core Implementation | 3-5 days | TBD | TBD | 🟡 Not Started |
| 4. Testing & Validation | 2-3 days | TBD | TBD | 🟡 Not Started |
| 5. Documentation & Deployment | 1-2 days | TBD | TBD | 🟡 Not Started |
| 6. Optional Enhancements | Variable | TBD | TBD | 🟡 Future |

**Total Estimated Duration:** 8-13 days (excluding optional enhancements)

---

## Next Steps

1. ✅ **Immediate:** Review and approve this plan
2. 🔲 **Next:** Install LibRaw on development machine
3. 🔲 **Then:** Create proof of concept with single DNG file
4. 🔲 **After:** Begin Phase 1 implementation

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-10-05 | Initial plan created | Claude |

---

## References

- [LibRaw Official Documentation](https://www.libraw.org/docs)
- [LibRaw GitHub Repository](https://github.com/LibRaw/LibRaw)
- [golibraw Go Package](https://pkg.go.dev/github.com/inokone/golibraw)
- [Olsen RAW Support Research](./RESEARCH.md)
- [Private Test Data Integration Test](../../internal/indexer/private_testdata_integration_test.go)
