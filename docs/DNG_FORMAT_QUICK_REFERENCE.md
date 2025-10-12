# DNG Format Quick Reference

**Purpose**: Fast lookup for DNG file handling decisions
**Full Details**: See `DNG_FORMAT_DEEP_DIVE.md`

---

## Critical Facts

### DNG File Structure

```
DNG = TIFF 6.0 + IFD hierarchy + embedded previews + RAW sensor data
```

**Typical Contents**:
- IFD0: Small thumbnail (~256px) for file browser compatibility
- SubIFD1: Medium preview (~1024px) JPEG
- SubIFD2: Full-size preview (~6000px) JPEG (optional)
- SubIFDn: RAW sensor data (CFA/Bayer or Linear Raw)

**NewSubFileType Values**:
- `0` = Full-resolution primary image
- `1` = Preview/thumbnail (primary)
- `10001h` (65537) = Alternate preview

### JPEG Compression Confusion

**Two completely different things**:

1. **JPEG-Compressed RAW** (Lossy DNG):
   - Compression applied to RAW sensor data itself
   - NOT an embedded JPEG image
   - File is no longer "true RAW" (mosaic info removed)
   - LibRaw may fail or only extract thumbnail

2. **Embedded JPEG Previews**:
   - Standard JPEG images in SubIFDs
   - Already demosaiced and tone-mapped
   - Multiple sizes available
   - Fast to extract and decode

### Monochrome DNGs

**Key Difference**: 1 color channel, not 3

```
Color DNG:     width × height × 3 channels × bytes_per_sample
Monochrome DNG: width × height × 1 channel × bytes_per_sample
```

**Critical**: Always use `colors` field from LibRaw, never hardcode `3`

**Preview JPEGs**: Still RGB (R=G=B for grayscale), 3 channels for compatibility

---

## Decision Matrix: Preview vs RAW Decode

| Use Case | Method | Why |
|----------|--------|-----|
| 64px thumbnail | Extract 256px preview, downsize | 60× faster, equal quality |
| 256px thumbnail | Extract 1024px preview, downsize | 60× faster, equal quality |
| 512px thumbnail | Extract 1024px preview, downsize | 60× faster, equal quality |
| 1024px thumbnail | Extract full preview or 1024px preview | 60× faster, equal/better quality |
| Color palette | Extract 256px preview | Sufficient, much faster |
| Perceptual hash | Extract 256-512px preview | Maintains similarity, much faster |
| Full-res editing | Decode RAW | Need full dynamic range, adjustable WB |

**Performance**:
- Preview extraction: 10-20ms
- RAW decode: ~1200ms
- Speedup: **60-120×**

---

## Extraction Methods

### Method 1: ExifTool (Recommended for now)

```bash
exiftool -b -PreviewImage input.dng > preview.jpg      # Medium (~1024px)
exiftool -b -JpgFromRaw input.dng > fullsize.jpg       # Full-size (if exists)
```

**Pros**: Reliable, works with all DNGs
**Cons**: External dependency, slower than direct parsing

### Method 2: TIFF/IFD Parsing (Future Implementation)

```go
// 1. Parse TIFF structure
// 2. Locate SubIFD with NewSubFileType = 1
// 3. Extract JPEG data from IFD
// 4. Decode with image/jpeg
```

**Pros**: Pure Go, fast, no dependencies
**Cons**: Complex, need to handle IFD trees

### Method 3: LibRaw Decode (Current Fallback)

```go
img, err := golibraw.DecodeRaw(path)
```

**Pros**: Full quality, full resolution
**Cons**: 60-120× slower, CGO dependency, buffer overflow risk

---

## Common Pitfalls

### 1. Hardcoding 3 Color Channels

**WRONG**:
```go
size := width * height * 3  // Assumes RGB!
```

**RIGHT**:
```go
size := width * height * colors  // Use actual colors field!
```

### 2. Ignoring Embedded Previews

**WRONG**: Always decode RAW for thumbnails

**RIGHT**: Check preview sizes, use preview if suitable

### 3. Upscaling Small Previews

**WRONG**: Use 256px preview for 1024px thumbnail

**RIGHT**: Decode RAW if preview < target size

### 4. Assuming LibRaw Always Works

**WRONG**: Trust LibRaw for all DNGs

**RIGHT**: Test with JPEG-compressed and monochrome DNGs

---

## LibRaw Limitations

**Requires Compile-time Flags**:
- `--enable-jpeg` for JPEG-compressed DNGs
- `--with-dng-sdk` for lossy/JPEG-XL DNGs

**Known Issues**:
- May refuse lossy-compressed DNGs
- May only extract thumbnail from lossy DNGs
- Monochrome DNGs: 1 channel, not 3 (buffer overflow risk)
- JPEG-XL DNGs need DNG SDK 1.7+ support

**Our Fix** (for monochrome buffer overflow):
- Extract `colors` field from `libraw_processed_image_t`
- Use `colors` in all buffer calculations
- Return `image.Gray` for 1-channel, `image.NRGBA` for 3-channel
- Local fork: `/Users/ade/Documents/projects/go-libraw-fix`

---

## Leica M11 Monochrom Specifics

**Sensor**: 60MP monochrome (9536×6336), no CFA
**Compression**: JPEG-compressed RAW (lossy)
**Bit Depth**: 16-bit
**Color Channels**: 1 (grayscale)
**Previews**: Typically 256px thumbnail + 1024px medium preview

**Why LibRaw Failed** (before fix):
1. JPEG-compressed RAW (needs JPEG support)
2. Monochrome sensor (1 channel, not 3)
3. Buffer overflow (hardcoded 3 channels)

**Now Works**: After fixing buffer overflow bug, all 30 test files decode successfully

---

## Quick Strategy Guide

### For Thumbnail Generation (Olsen)

**Phase 1** (Immediate):
1. Implement proper preview extraction (TIFF/IFD parsing)
2. Check preview size vs target thumbnail sizes
3. Use preview if size >= target
4. Fall back to RAW decode if preview < target

**Phase 2** (Next):
1. Build quality assessment (SSIM, PSNR)
2. Compare preview vs RAW decode quality
3. Document optimal strategy per thumbnail size

**Phase 3** (Future):
1. Cache preview sizes in database
2. Optimize worker pool for I/O-bound preview extraction
3. Consider preview-only mode (no RAW decode at all)

### Expected Performance Gain

**Current** (RAW decode for all thumbnails):
- Processing time: ~1200ms per file
- Throughput: ~2.5 photos/second (with 4 workers)

**After Preview Extraction** (for 1024px and smaller):
- Processing time: ~20ms per file (60× faster)
- Throughput: ~150 photos/second (with 4 workers)

**Realistic** (mixed strategy):
- Preview for thumbnails: ~20ms
- RAW decode for special cases: ~1200ms
- Average: ~50-100ms per file (12-24× faster)
- Throughput: ~40-80 photos/second

---

## Key Takeaway

**For thumbnail generation (64-1024px), embedded DNG previews are usually better than RAW decode:**

- 60-120× faster
- Equal or better quality
- No CGO dependencies
- Better compatibility (avoids lossy-compressed DNG issues)

**Always try preview extraction first, fall back to RAW decode only when necessary.**

---

## References

- **Full Research**: `docs/DNG_FORMAT_DEEP_DIVE.md`
- **LibRaw Fix**: `docs/LIBRAW_FIX_COMPLETE.md`
- **Buffer Overflow Research**: `docs/LIBRAW_BUFFER_OVERFLOW_RESEARCH.md`
- **DNG Spec 1.6**: https://paulbourke.net/dataformats/dng/dng_spec_1_6_0_0.pdf
- **LibRaw Docs**: https://www.libraw.org/docs/
- **ExifTool**: https://exiftool.org/

---

**Last Updated**: 2025-10-12
**Status**: Research complete, implementation pending
