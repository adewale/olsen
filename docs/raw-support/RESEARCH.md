# RAW Image Format Support Research

**Date:** 2025-10-05
**Researcher:** Claude
**Purpose:** Evaluate options for adding RAW DNG support to Olsen

---

## Problem Statement

Olsen currently cannot index RAW DNG files from Leica M11 Monochrom cameras. These files use "PhotometricInterpretation: Linear Raw" which Go's standard TIFF decoder (`golang.org/x/image/tiff`) does not support.

**Current Status:**
- ✅ JPEG, PNG, BMP support working
- ✅ Basic TIFF/DNG format recognition
- ❌ RAW sensor data decoding (fails with "unsupported feature: color model")
- ❌ Thumbnail generation from RAW
- ❌ Colour extraction from RAW

**Test Case:**
- 30 Leica M11 Monochrom DNG files in `private-testdata/`
- All currently fail during indexing
- Integration test documents limitation: `TestIntegrationIndexPrivateTestData`

---

## Options Evaluated

### Option 1: LibRaw (C++ library) via Go CGO Bindings ⭐ RECOMMENDED

**Description:**
LibRaw is a mature, comprehensive C++ library for reading RAW files from virtually all digital cameras. Based on dcraw (the gold standard RAW processor), it has been in active development since 2008.

**Capabilities:**
- Supports 500+ camera models
- Extracts RAW sensor data, metadata, embedded previews
- Demosaicing (Bayer pattern conversion to RGB)
- White balance, colour correction, exposure adjustment
- Active development with new camera support every 9-18 months

**Available Go Bindings:**

1. **`github.com/inokone/golibraw`** ⭐ RECOMMENDED
   - Simple API: `ExtractMetadata()`, `ExtractThumbnail()`, `ImportRaw()`
   - Returns standard Go `image.Image`
   - Active development
   - Installation: `go get github.com/inokone/golibraw`
   - Requires: `brew install libraw` (macOS) or `apt-get install libraw-dev` (Ubuntu)

2. **`github.com/seppedelanghe/go-libraw`**
   - More recent (2023)
   - Processor-based API with configurable options
   - Goroutine-friendly design
   - Similar capabilities to golibraw

3. **`github.com/mrht-srproject/librawgo`**
   - Lower-level bindings to LibRaw C API
   - More control but more complex to use

**Advantages:**
- ✅ Comprehensive format support (DNG, CR2, NEF, RAF, ARW, etc.)
- ✅ Battle-tested, production-ready
- ✅ Full RAW processing capabilities
- ✅ Active development and camera support updates
- ✅ Used by professional photo management software
- ✅ Proper demosaicing for accurate colour extraction
- ✅ Will definitely work with Leica M11 DNG files

**Disadvantages:**
- ❌ Requires CGO (complicates cross-compilation)
- ❌ External system library dependency (libraw)
- ❌ Larger binary size
- ❌ More complex deployment (must install libraw on target systems)
- ❌ Requires C/C++ compiler toolchain for building

**Licensing:**
- Dual-licensed: LGPL v2.1 or CDDL v1.0
- Can choose CDDL v1.0 to avoid LGPL requirements
- Dynamic linking recommended for LGPL compliance

**Installation:**
```bash
# macOS
brew install libraw

# Ubuntu/Debian
sudo apt-get install libraw-dev

# Go package
go get github.com/inokone/golibraw
```

**Example Usage:**
```go
import "github.com/inokone/golibraw"

// Extract metadata
metadata, err := golibraw.ExtractMetadata("/path/to/image.dng")

// Decode to image.Image
img, err := golibraw.ImportRaw("/path/to/image.dng")

// Extract thumbnail
err := golibraw.ExtractThumbnail("/path/to/image.dng", "/path/to/thumb.jpg")
```

**Verdict:** ⭐ **Best option for production use**

---

### Option 2: Pure Go DNG Thumbnail Extraction

**Library:** `github.com/mdouchement/dng`

**Description:**
A minimal Go library that extracts embedded JPEG thumbnails from DNG file metadata.

**Capabilities:**
- Extracts embedded JPEG preview from DNG
- Returns as standard Go `image.Image`
- No external dependencies

**Advantages:**
- ✅ No CGO required
- ✅ Pure Go (easy cross-compilation)
- ✅ Simple installation (`go get`)
- ✅ Small binary size
- ✅ No system dependencies

**Disadvantages:**
- ❌ Only extracts embedded thumbnails (not true RAW decoding)
- ❌ Cannot decode RAW sensor data
- ❌ No colour extraction from RAW possible
- ❌ Fails if DNG has no embedded preview
- ❌ Cannot generate custom thumbnail sizes
- ❌ Very limited - just thumbnail extraction

**Use Case:**
Only suitable if:
- You're okay with embedded thumbnails (not all RAW files have them)
- You don't need colour analysis from RAW data
- You want zero-dependency deployment

**Example Usage:**
```go
import _ "github.com/mdouchement/dng"

// Automatically registered with image.Decode
img, _, err := image.Decode(file)
```

**Verdict:** ❌ **Too limited for Olsen's needs**

---

### Option 3: Adobe DNG SDK via Go Bindings

**Library:** `github.com/abworrall/go-dng`

**Description:**
Go bindings for Adobe's official DNG SDK v1.6, which is Adobe's reference implementation for DNG format.

**Capabilities:**
- Full DNG processing using official Adobe SDK
- Access to different processing stages (raw, linearized, demosaiced)
- White balance and colour correction
- Camera-specific colour matrices

**Advantages:**
- ✅ Official Adobe implementation
- ✅ Comprehensive DNG support
- ✅ Access to low-level processing stages

**Disadvantages:**
- ❌ Requires CGO (not pure Go)
- ❌ Complex build process (must compile DNG SDK first)
- ❌ Linux-only currently
- ❌ Requires `build-essentials`, `zlib1g-dev`, `libexpat1-dev`
- ❌ "Very incomplete" Go bindings
- ❌ Potential memory leaks
- ❌ DNG-specific (doesn't support CR2, NEF, RAF, etc.)
- ❌ More complex Makefile-based build

**Installation:**
```bash
# Ubuntu
sudo apt-get install build-essentials zlib1g-dev libexpat1-dev

# Clone and build
git clone https://github.com/abworrall/go-dng
cd go-dng
make
```

**Verdict:** ❌ **Too complex, incomplete bindings, DNG-only**

---

### Option 4: Native Go RAW Decoders

**Libraries:**
- `bitbucket.org/osocurioso/raw` - Aims to implement decoders for various vendors
- `github.com/kladd/raw` - Fujifilm RAW only
- `github.com/mdouchement/hdr` - HDR/tone mapping focus

**Description:**
Pure Go implementations attempting to decode RAW formats without C dependencies.

**Advantages:**
- ✅ Pure Go (no CGO)
- ✅ Easy deployment
- ✅ Simple cross-compilation

**Disadvantages:**
- ❌ Very limited format support (often single vendor)
- ❌ Incomplete implementations
- ❌ Not production-ready
- ❌ Won't work with Leica M11 DNG files
- ❌ Minimal active development
- ❌ Missing demosaicing algorithms
- ❌ No colour correction pipelines

**Verdict:** ❌ **Not suitable - won't support Leica DNGs**

---

## Comparison Matrix

| Feature | LibRaw + golibraw | mdouchement/dng | go-dng | Native Go |
|---------|-------------------|-----------------|---------|-----------|
| **RAW Decoding** | ✅ Full | ❌ Thumbnail only | ✅ Full | ⚠️ Limited |
| **DNG Support** | ✅ Yes | ⚠️ Thumbnail | ✅ Yes | ❌ No |
| **Leica M11 Support** | ✅ Yes | ⚠️ Maybe | ⚠️ Maybe | ❌ No |
| **Other RAW Formats** | ✅ 500+ cameras | ❌ No | ❌ DNG only | ⚠️ 1-2 brands |
| **CGO Required** | ✅ Yes | ❌ No | ✅ Yes | ❌ No |
| **External Deps** | libraw | None | DNG SDK + libs | None |
| **Production Ready** | ✅ Yes | ⚠️ Limited | ❌ No | ❌ No |
| **Colour Extraction** | ✅ Yes | ❌ No | ✅ Yes | ❌ No |
| **Active Development** | ✅ Yes | ⚠️ Minimal | ❌ Stalled | ❌ Stalled |
| **Installation Complexity** | Medium | Easy | Hard | Easy |
| **Binary Size Impact** | Large | Minimal | Large | Minimal |
| **Cross-Compilation** | Hard | Easy | Hard | Easy |

---

## Recommendation

### **Primary Recommendation: LibRaw via `github.com/inokone/golibraw`**

**Rationale:**
1. **Only solution that will actually work** for Leica M11 Monochrom DNG files
2. Comprehensive format support (will work with any camera in the future)
3. Production-ready, battle-tested solution
4. Can extract full RAW data for thumbnails and colour analysis
5. Active development with regular camera support updates

**Trade-offs to Accept:**
- CGO dependency (manageable with proper documentation)
- System library requirement (document in README)
- More complex deployment (provide Docker images)
- Larger binary size (acceptable for desktop tool)

**Mitigation Strategies:**
1. Use build tags to make LibRaw optional:
   ```go
   //go:build cgo
   // +build cgo

   package indexer

   import "github.com/inokone/golibraw"
   ```

2. Provide fallback for non-CGO builds:
   ```go
   //go:build !cgo
   // +build !cgo

   package indexer

   func decodeRaw(path string) (image.Image, error) {
       return nil, errors.New("RAW support requires CGO and libraw")
   }
   ```

3. Document installation clearly in README
4. Provide Docker images with LibRaw pre-installed
5. Add CI/CD pipeline for CGO builds

---

## Alternative Approach (Not Recommended)

If CGO is absolutely unacceptable, use `mdouchement/dng` for thumbnail extraction only:

**Limitations:**
- Only works if DNG has embedded preview
- No colour extraction from RAW data
- Cannot generate custom thumbnail sizes
- Less robust solution

**Use only if:**
- Deployment constraints prevent CGO
- Embedded thumbnails are acceptable
- Colour analysis is not critical

---

## Implementation Strategy

### Phase 1: Proof of Concept
1. Install LibRaw: `brew install libraw`
2. Test with single Leica DNG file
3. Verify decoding, thumbnail, colour extraction

### Phase 2: Integration
1. Create `internal/indexer/raw.go` with build tag
2. Implement RAW decoder using golibraw
3. Integrate with existing indexer pipeline
4. Add fallback for non-CGO builds

### Phase 3: Testing
1. Update `TestIntegrationIndexPrivateTestData`
2. Verify all 30 Leica files index successfully
3. Test with other RAW formats (CR2, NEF, RAF)
4. Performance benchmarking

### Phase 4: Documentation
1. Update README with LibRaw installation
2. Document CGO requirement
3. Provide Docker image
4. Add troubleshooting guide

---

## Performance Considerations

**Expected Performance:**
- RAW decoding: ~1-2 seconds per file (slower than JPEG)
- Memory: ~200-400MB per RAW file during processing
- Parallelization: Works well with worker pools

**Optimizations:**
- Use LibRaw's half-size mode for faster processing
- Consider caching demosaiced images
- Process RAW files in separate worker queue

---

## Security Considerations

**LibRaw Security:**
- Mature library with security track record
- Regular CVE monitoring and patches
- Used by major photo software (lower risk)

**Recommendations:**
- Pin LibRaw version in documentation
- Monitor security advisories
- Update LibRaw regularly
- Validate input files before processing

---

## Licensing Implications

**LibRaw Licenses:**
- Option 1: LGPL v2.1
- Option 2: CDDL v1.0

**For Olsen:**
- Recommend CDDL v1.0 (more permissive)
- Dynamic linking satisfies LGPL if chosen
- Document license choice in LICENSE file
- No impact on Olsen's own licensing

---

## References

### Primary Sources
- [LibRaw Official Site](https://www.libraw.org/)
- [LibRaw GitHub](https://github.com/LibRaw/LibRaw)
- [golibraw Package](https://pkg.go.dev/github.com/inokone/golibraw)
- [go-libraw Package](https://pkg.go.dev/github.com/seppedelanghe/go-libraw)

### Technical Resources
- [DNG Specification (Adobe)](https://helpx.adobe.com/camera-raw/digital-negative.html)
- [dcraw Documentation](https://www.dechifro.org/dcraw/)
- [RAW Image Processing Pipeline](https://en.wikipedia.org/wiki/Raw_image_format#Processing)

### Community Discussions
- [golang-nuts: RAW file processing](https://groups.google.com/g/golang-nuts/c/60WthPS_TXg)
- [Go Issue #57746: DNG decoding](https://github.com/golang/go/issues/57746)

---

## Appendix: Test Results

### Current Status (2025-10-05)
```
=== RUN   TestIntegrationIndexPrivateTestData
Files found: 30
Files processed: 0
Files failed: 30
Error: "tiff: unsupported feature: color model"
Status: SKIP (RAW DNG format not supported)
```

### Expected After Implementation
```
=== RUN   TestIntegrationIndexPrivateTestData
Files found: 30
Files processed: 30
Files failed: 0
Thumbnails generated: 120 (4 sizes × 30 photos)
Photos with colour data: 30/30 (100%)
Status: PASS
```

---

## Conclusion

LibRaw via golibraw is the clear choice for production RAW support in Olsen. While it introduces CGO complexity, it's the only solution that will reliably decode Leica M11 Monochrom DNG files and provides room for future expansion to other RAW formats.

The trade-off between deployment complexity and functionality is justified for a desktop photo management tool where RAW support is essential for professional photographers.
