# Thumbnail Quality Research Briefing

**Status**: Research planning phase
**Priority**: Medium (quality improvement, not blocking)
**Owner**: TBD (agent research task)

## Executive Summary

Investigation into thumbnail generation quality ceiling and optimization opportunities. Current implementation uses LibRaw → Lanczos3 resize → JPEG 85% pipeline. Need to determine if we're hitting quality limits at RAW decode, resize, or encode stages, and what improvements are viable within portability/performance constraints.

## Current Implementation

### RAW Processing Pipeline

**Primary Decoder** (CGO builds):
- Library: `github.com/inokone/golibraw` (LibRaw C binding)
- Location: `internal/indexer/raw.go:21`
- Process: Full RAW demosaicing → RGB image
- Unknown config: demosaicing algorithm, color space output, bit depth

**Fallback Chain**:
1. LibRaw full decode (`DecodeRaw()`)
2. Embedded JPEG extraction (`ExtractEmbeddedJPEG()`) - manual JPEG marker scan
3. Standard Go decoders (jpeg/png/bmp/tiff via `golang.org/x/image`)

**Non-CGO Builds**:
- RAW files: Falls back to embedded JPEG or fails
- JPEG/BMP/PNG: Standard decoders only

### Thumbnail Generation

**Location**: `internal/indexer/thumbnail.go:76`

**Algorithm**:
```go
resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
```

**Sizes**: 64px, 256px, 512px, 1024px (longest edge, aspect-ratio preserved)

**Encoding**:
```go
jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85})
```

**Storage**: SQLite BLOBs (portability requirement)

### Performance Context

- ~62ms per photo (full pipeline: metadata + thumbnails + color + hash)
- ~431ms per photo with 4 workers (includes I/O)
- Target: 10-30 photos/second throughput
- Memory: ~500MB peak with 4 workers

## Critical Gaps

### 1. RAW Library Quality Ceiling

**Known Issues**:
- LibRaw defaults unknown (demosaicing algo, color space, bit depth)
- No explicit color management (profile conversion, gamma correction)
- Unclear if getting full sensor dynamic range or 8-bit lossy output

**Research Questions**:
- What demosaicing algorithm is golibraw using? (AHD, DCB, PPG, or LibRaw's default?)
- Is LibRaw outputting sRGB or linear RGB?
- Are we getting 8-bit or 16-bit output before resize?
- Is auto white balance enabled?
- How does LibRaw decode quality compare to embedded JPEG previews?
- What's the quality ceiling imposed by LibRaw configuration?

### 2. Quality Assessment Mechanism [BUILD THIS FIRST]

**Currently Missing**:
- No objective quality metrics (SSIM, PSNR, sharpness)
- No visual comparison tooling
- No regression testing
- No way to validate if current approach is adequate

**Needed Infrastructure**:

**A. Metrics Package** (`internal/quality/metrics.go`):
- SSIM (Structural Similarity Index) - perceptual quality
- PSNR (Peak Signal-to-Noise Ratio) - technical quality
- Sharpness metrics (Laplacian variance, gradient magnitude)
- Color accuracy (delta-E in CIELAB space)
- Edge preservation scores

**B. Comparison Framework** (`internal/quality/compare.go`):
- Generate thumbnails with multiple approaches
- Score each with objective metrics
- Create visual comparison HTML reports
- Track database size impact
- Measure processing time

**C. Benchmark CLI**:
```bash
./olsen benchmark-thumbnails \
  --input testdata/representative_samples/ \
  --approaches "current,lanczos2,mitchell,catmull-rom" \
  --qualities "75,85,90,95" \
  --formats "jpeg,webp,avif" \
  --output report.html
```

**D. Representative Test Suite**:
- High-contrast scenes (sharpness preservation test)
- Low-light/high-ISO (noise handling)
- Fine detail (foliage, fabric, text)
- Wide-gamut colors (sunset, saturated)
- Various aspect ratios
- Different camera models (LibRaw consistency)

### 3. Pipeline Quality Issues

**Color Management**:
- Is LibRaw outputting sRGB or linear RGB?
- Are embedded color profiles being preserved/converted?
- Should wide-gamut RAW files be explicitly converted to sRGB for thumbnails?
- Is gamma correction happening at the right stage?
- Risk: Color shifts between full-res and thumbnail

**Resizing Quality**:
- Is Lanczos3 optimal for photographic content?
- Alternative kernels: Mitchell-Netravali, Catmull-Rom, Lanczos2
- Should resizing happen in linear light (gamma-correct resizing)?
- Pre-sharpening before downsampling? (prevents softness)
- Post-sharpening after resize? (unsharp mask ~0.3-0.5 strength)
- Risk: Soft/blurry thumbnails, aliasing artifacts

**Encoding Quality**:
- Is JPEG 85% optimal? (need SSIM scores across 75-95% range)
- Should quality vary by thumbnail size? (e.g., 80% for 64px, 90% for 1024px)
- Progressive JPEG vs. baseline?
- WebP lossy/lossless: better quality-to-size ratio?
- AVIF: cutting-edge quality but browser support?
- Risk: Visible compression artifacts, excessive database bloat

**LibRaw Configuration Unknowns**:
- `raw.output_color` setting? (0=raw, 1=sRGB, 2=Adobe RGB)
- `raw.use_camera_wb` vs `raw.use_auto_wb`?
- `raw.user_qual` (demosaicing quality: 0=linear, 1=VNG, 2=PPG, 3=AHD, 4=DCB)
- `raw.output_bps` (8 or 16 bits per channel)?
- `raw.half_size` (fast but lower quality)?
- Risk: Low-quality RAW decode limiting entire pipeline

### 4. Fallback Path Quality

**Embedded JPEG Extraction** (`raw.go:98`):
- Current: Manual JPEG marker scanning (0xFFD8/0xFFD9)
- Risk: Embedded previews may be lower resolution than target sizes
- Risk: May not represent final RAW processing (different WB, exposure)
- Need to compare: LibRaw full decode vs embedded JPEG quality

**Non-CGO Build Limitations**:
- Standard `image.Decode()` cannot handle RAW sensor data
- DNG files depend on embedded JPEG extraction or fail
- Quality ceiling: embedded preview quality only

## Hard Constraints

1. **Portability**: Thumbnails MUST stay in SQLite database (single-file catalog)
2. **Read-only**: Cannot modify source photo files
3. **Format support**: DNG, JPEG, BMP input files required
4. **Memory budget**: ~500MB peak with 4 workers
5. **Performance target**: 10-30 photos/second throughput
6. **Aspect ratio preservation**: Critical for composition

## Research Plan

### Phase 1: Build Assessment Infrastructure [CRITICAL PATH]

**Deliverables**:
1. `internal/quality/metrics.go` - SSIM, PSNR, sharpness, color accuracy
2. `internal/quality/compare.go` - Multi-approach comparison framework
3. Representative test dataset (15-20 diverse photos)
4. Baseline quality scores for current approach
5. HTML visual comparison report generator

**Success Criteria**:
- Can generate objective quality scores for any thumbnail approach
- Can visually compare approaches side-by-side
- Have quantified baseline to improve against

### Phase 2: LibRaw Decode Quality Audit

**Research Tasks**:
1. Document golibraw's actual LibRaw configuration (examine source/docs)
2. Test different demosaicing algorithms (AHD, DCB, PPG, VNG, linear)
3. Test different color space outputs (linear, sRGB, Adobe RGB)
4. Test different bit depths (8-bit vs 16-bit)
5. Compare: full LibRaw decode vs embedded JPEG vs standard decoder
6. Measure dynamic range preservation through pipeline

**Deliverables**:
- Configuration audit report
- Quality scores for each LibRaw setting combination
- Recommendation for optimal LibRaw configuration
- Performance impact analysis

**Questions to Answer**:
- Are we hitting a quality ceiling at decode stage?
- Can we improve LibRaw output without performance penalty?
- Is embedded JPEG fallback acceptable quality?

### Phase 3: Resizing Algorithm Comparison

**Test Matrix**:
- Algorithms: Lanczos3 (current), Lanczos2, Mitchell-Netravali, Catmull-Rom, Bilinear
- With/without gamma-correct resizing (linear light conversion)
- With/without pre-sharpening (before resize)
- With/without post-sharpening (after resize, before encode)

**Measurements**:
- SSIM scores (perceptual quality)
- Sharpness scores (edge preservation)
- Processing time impact
- Visual comparison grids

**Deliverables**:
- Quality vs performance trade-off analysis
- Recommendation for resize algorithm + sharpening
- Implementation plan if change needed

**Questions to Answer**:
- Is Lanczos3 optimal or causing softness?
- Does sharpening improve perceived quality?
- Can we improve quality within performance budget?

### Phase 4: Encoding Optimization

**Test Matrix**:
- JPEG quality levels: 75, 80, 85 (current), 90, 95
- Per-size quality variation (e.g., 80% for 64px, 90% for 1024px)
- Progressive vs baseline JPEG
- WebP lossy (quality 70-95)
- WebP lossless
- AVIF (quality 60-90) - if browser support acceptable

**Measurements**:
- SSIM scores at each quality level
- Database size impact
- Encoding time impact
- Browser support matrix (for WebP/AVIF)

**Deliverables**:
- Optimal JPEG quality recommendation
- WebP/AVIF viability assessment
- Database size projections
- Migration strategy if format changes

**Questions to Answer**:
- Is 85% JPEG optimal or overkill/insufficient?
- Can alternative formats improve quality-to-size ratio?
- What's the user impact (database size, compatibility)?

### Phase 5: Integration & Recommendations

**Deliverables**:
1. Comprehensive recommendation report with data
2. Proposed implementation changes (if any)
3. Migration strategy (schema versioning if format/quality changes)
4. Updated performance benchmarks (quality + speed + size)
5. Regression test suite (quality metrics in CI)

**Decision Framework**:
- Must quantify quality improvement (SSIM delta)
- Must stay within performance budget (10-30 photos/sec)
- Must consider database size impact (portability)
- Must be validated by visual human review

## Expected Outcomes

**Best Case**:
- Identify 20-30% SSIM improvement opportunity at minimal cost
- LibRaw configuration tweaks yield better decode quality
- Optimized resize + sharpening improve perceived quality
- Stay within performance/size budget

**Realistic Case**:
- Identify 10-15% SSIM improvement opportunity
- Current approach is mostly optimal, minor tweaks possible
- Trade-offs between quality/performance/size documented
- Clear understanding of quality ceiling

**Worst Case**:
- Current approach is already near-optimal
- Improvements require unacceptable performance/size trade-offs
- At least have quantified baseline and testing infrastructure
- Can detect future regressions

## Dependencies

**Technical**:
- Need representative photo dataset (various cameras, scenes, conditions)
- May need additional Go libraries for SSIM/PSNR metrics
- May need WebP/AVIF encoder libraries for evaluation

**Tooling**:
- HTML/CSS for visual comparison reports
- Charting library for metric visualization
- Benchmark harness integration

**Knowledge**:
- LibRaw documentation and configuration options
- Image quality metrics theory (SSIM, PSNR, perceptual sharpness)
- Color management and gamma-correct processing
- RAW processing pipelines (demosaicing, color spaces)

## References

**Code Locations**:
- RAW decoding: `internal/indexer/raw.go` (CGO), `raw_nocgo.go` (fallback)
- Thumbnail generation: `internal/indexer/thumbnail.go:76`
- Full pipeline: `internal/indexer/indexer.go:170` (processFile)

**Dependencies**:
- `github.com/inokone/golibraw` - LibRaw wrapper
- `github.com/nfnt/resize` - Lanczos3 resizing
- `image/jpeg` - JPEG encoding

**Related Docs**:
- `docs/raw-support/RESEARCH.md` - RAW processing background
- `CLAUDE.md` - Architecture overview
- `specs/olsen_requirements.md` - System requirements

## Next Steps

1. **Immediate**: Create `internal/quality/` package structure
2. **Week 1**: Implement metrics (SSIM, PSNR, sharpness)
3. **Week 1**: Build comparison framework and HTML report generator
4. **Week 1**: Assemble representative test dataset
5. **Week 2**: Run Phase 2 (LibRaw audit)
6. **Week 3**: Run Phase 3 (resizing comparison)
7. **Week 4**: Run Phase 4 (encoding optimization)
8. **Week 5**: Synthesize findings and recommendations

**Estimated Effort**: 3-5 weeks research + 1-2 weeks implementation (if changes warranted)

---

**Document Status**: Initial research briefing
**Last Updated**: 2025-10-11
**Next Review**: After Phase 1 completion
