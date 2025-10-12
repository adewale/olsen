# Current Status: RAW DNG Support in Olsen

**Date:** 2025-10-05
**Status:** ❌ Not Working - RAW DNG files cannot be indexed

---

## Problem Description

When attempting to index RAW DNG files (specifically Leica M11 Monochrom files from `private-testdata/`), Olsen fails with the following error:

```
2025/10/05 23:40:20 Worker 11: Failed to process private-testdata/2024-12-23/L1001531.DNG:
failed to decode image: tiff: unsupported feature: color model
```

### Technical Details

**Why This Happens:**

1. **DNG files are TIFF-based**, so Go's image decoder recognizes them as TIFF files
2. **Go's TIFF decoder is limited** to standard color models:
   - RGB
   - RGBA
   - Grayscale
   - Paletted (indexed color)
3. **RAW DNG files use special color models** not supported by standard TIFF:
   - `PhotometricInterpretation: Linear Raw` (what Leica uses)
   - CFA (Color Filter Array) / Bayer pattern data
   - Unprocessed sensor data requiring demosaicing

**What's Missing:**
- RAW sensor data decoding
- Bayer pattern demosaicing (converting sensor CFA to RGB)
- Color space conversion
- White balance application
- Gamma correction

---

## Current Code Path

### 1. File Discovery
```
internal/indexer/indexer.go: IndexDirectory()
  ↓
  Finds: L1001531.DNG
  ↓
  Recognizes extension: .DNG
  ↓
  Routes to: processFile()
```

### 2. Image Decoding Attempt
```
internal/indexer/indexer.go: processFile()
  ↓
  Opens file
  ↓
  Calls: image.Decode(file)
  ↓
  Go's image decoder checks registered formats
  ↓
  Matches: golang.org/x/image/tiff (registered in indexer.go:10)
  ↓
  TIFF decoder attempts to parse
  ↓
  Reads TIFF header: ✅ Valid TIFF
  ↓
  Reads PhotometricInterpretation tag: "Linear Raw" (34892)
  ↓
  ❌ ERROR: "unsupported feature: color model"
```

### 3. Error Propagation
```
image.Decode() returns error
  ↓
  processFile() catches error
  ↓
  Logs: "Worker X: Failed to process ..."
  ↓
  Increments stats.FilesFailed
  ↓
  Photo is NOT indexed
```

---

## What We're Missing

### Decoder Capabilities Needed

| Capability | Standard Go TIFF | LibRaw | Status |
|------------|------------------|--------|--------|
| Read TIFF header | ✅ Yes | ✅ Yes | ✅ Working |
| Read EXIF metadata | ✅ Yes | ✅ Yes | ✅ Working |
| Decode RGB TIFF | ✅ Yes | ✅ Yes | ✅ Working |
| Decode RAW sensor data | ❌ No | ✅ Yes | ❌ Missing |
| Demosaic Bayer pattern | ❌ No | ✅ Yes | ❌ Missing |
| Apply white balance | ❌ No | ✅ Yes | ❌ Missing |
| Color space conversion | ❌ No | ✅ Yes | ❌ Missing |
| Generate RGB output | ❌ No | ✅ Yes | ❌ Missing |

---

## File Analysis: L1001531.DNG

### File Properties
```bash
$ file L1001531.DNG
L1001531.DNG: TIFF image data, little-endian, direntries=40, height=6336,
bps=16, compression=JPEG, PhotometricIntepretation=(unknown=0xffff884c),
manufacturer=Leica Camera AG, model=LEICA M11 Monochrom, orientation=upper-left,
width=9536
```

### EXIF Data (Excerpt)
```bash
$ exiftool L1001531.DNG | head -20
File Type                       : DNG
File Type Extension             : dng
MIME Type                       : image/x-adobe-dng
Image Width                     : 9536
Image Height                    : 6336
Bits Per Sample                 : 16
Compression                     : JPEG
Photometric Interpretation      : Linear Raw    ← THIS IS THE PROBLEM
Make                            : Leica Camera AG
Camera Model Name               : LEICA M11 Monochrom
```

**Key Issue:** `Photometric Interpretation: Linear Raw`
- This is a RAW sensor data format
- Contains unprocessed CFA (Color Filter Array) data
- Requires demosaicing to convert to RGB
- Not supported by Go's standard image decoders

---

## Current Workarounds (None Effective)

### ❌ Workaround 1: Use Embedded Thumbnail
**Approach:** Extract embedded JPEG preview from DNG metadata
**Library:** `github.com/mdouchement/dng`
**Problem:**
- Only extracts thumbnails (doesn't decode RAW)
- No access to RAW sensor data for color analysis
- Cannot generate custom thumbnail sizes
- Not all DNG files have embedded previews

### ❌ Workaround 2: Skip RAW Files
**Approach:** Ignore .DNG files entirely
**Problem:**
- Defeats the purpose of Olsen
- Private-testdata contains 30 Leica photos (100% of test data)
- Users with RAW workflow cannot use Olsen

### ❌ Workaround 3: Convert RAW to JPEG First
**Approach:** Use external tool to convert DNG → JPEG before indexing
**Problem:**
- Manual preprocessing step
- Loses RAW data fidelity
- Doubles storage requirements
- Not practical for users

---

## Solution: LibRaw Integration

The only viable solution is to integrate LibRaw, which provides:

### What LibRaw Does

1. **Reads RAW Sensor Data**
   - Understands CFA/Bayer pattern
   - Extracts unprocessed sensor values
   - Handles Linear Raw photometric interpretation

2. **Demosaics to RGB**
   - Converts CFA data to full RGB image
   - Multiple demosaicing algorithms
   - Quality vs. speed options

3. **Applies Processing Pipeline**
   - White balance correction
   - Color space conversion
   - Gamma correction
   - Exposure adjustment

4. **Outputs Standard RGB**
   - Returns processed RGB image data
   - Can be converted to Go's `image.Image`
   - Ready for thumbnail generation and color analysis

### How It Will Work (After Implementation)

```
processFile()
  ↓
  Detects .DNG extension
  ↓
  Checks if RAW decoder available
  ↓
  YES: LibRaw installed
  ↓
  golibraw.ImportRaw(path)
  ↓
  LibRaw decodes RAW → RGB
  ↓
  Returns: image.Image ✅
  ↓
  Generate thumbnails ✅
  ↓
  Extract color palette ✅
  ↓
  Store in database ✅
```

---

## Impact on Olsen

### Current State
- **Files Found:** 30
- **Files Processed:** 0
- **Files Failed:** 30
- **Success Rate:** 0%
- **User Impact:** Cannot use Olsen with RAW files

### Expected After LibRaw Integration
- **Files Found:** 30
- **Files Processed:** 30
- **Files Failed:** 0
- **Success Rate:** 100%
- **User Impact:** Full RAW support for all camera brands

---

## Next Steps

1. ✅ **Document current failure mode** (this document)
2. 🔲 **Begin Phase 1: Environment Setup**
   - Install LibRaw: `brew install libraw`
   - Create proof of concept
   - Test with single DNG file
3. 🔲 **Follow implementation plan**
   - See: `docs/raw-support/PLAN.md`
   - Track progress in: `docs/raw-support/PROGRESS.md`

---

## References

- [PLAN.md](./PLAN.md) - Full implementation plan
- [RESEARCH.md](./RESEARCH.md) - Technical research on options
- [PROGRESS.md](./PROGRESS.md) - Live progress tracker
- [Integration Test](../../internal/indexer/private_testdata_integration_test.go) - Test that documents this limitation

---

## Error Log

### Latest Error (2025-10-05)
```
2025/10/05 23:40:20 Worker 11: Failed to process private-testdata/2024-12-23/L1001531.DNG:
failed to decode image: tiff: unsupported feature: color model
```

**Stack Trace (Conceptual):**
```
github.com/adewale/olsen/internal/indexer.(*Engine).processFile
  ↓
image.Decode(file)
  ↓
golang.org/x/image/tiff.Decode
  ↓
golang.org/x/image/tiff.(*decoder).decode
  ↓
Error: "unsupported feature: color model"
```

**Source Location:**
- Package: `golang.org/x/image/tiff`
- Error occurs when decoder encounters PhotometricInterpretation value it doesn't recognize
- Go TIFF decoder supports PhotometricInterpretation values: 0, 1, 2, 3 (WhiteIsZero, BlackIsZero, RGB, Paletted)
- Does NOT support: 34892 (Linear Raw) or 32803 (CFA)

---

## Conclusion

This is **not a bug in Olsen** - it's a fundamental limitation of Go's standard image decoders. LibRaw integration is the only practical solution to enable RAW DNG support.

The error message `"tiff: unsupported feature: color model"` is expected and confirms our analysis. No amount of tweaking Olsen's current code will fix this - we need a RAW-capable decoder library.
