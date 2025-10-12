# LibRaw Dual Library Support

**Date**: 2025-10-12
**Status**: Implemented

## Overview

Olsen now supports **two** LibRaw Go bindings with seamless switching at build time:

1. **`seppedelanghe/go-libraw`** - Full-featured, configurable (default)
2. **`inokone/golibraw`** - Simple, minimal configuration

## Quick Start

### Building with Default (seppedelanghe)
```bash
make build-raw
# or explicitly:
make build-seppedelanghe
```

### Building with golibraw
```bash
make build-golibraw
```

### Checking Which Library Is Active
```bash
./bin/olsen version
# Output:
# Olsen Photo Indexer
# Go version: go1.25.1
# OS/Arch: darwin/arm64
# RAW support: enabled (seppedelanghe/go-libraw)
```

## Comparison

| Feature | seppedelanghe/go-libraw | inokone/golibraw |
|---------|-------------------------|------------------|
| **Configuration Control** | ✅ Full | ❌ None |
| **Demosaic Algorithm** | ✅ Selectable (Linear, VNG, PPG, AHD, DCB, DHT, AAHD) | ❌ Fixed |
| **Output Bit Depth** | ✅ 8-bit or 16-bit | ❌ Fixed (8-bit assumed) |
| **Color Space Control** | ✅ sRGB, AdobeRGB, ProPhoto, etc. | ❌ Fixed (sRGB assumed) |
| **White Balance** | ✅ Camera WB or Auto WB | ❌ Fixed |
| **Diagnostics Capture** | ✅ Complete | ❌ No visibility |
| **Goroutine Safety** | ✅ Safe | ⚠️ Unknown |
| **API Complexity** | ⚠️ More complex | ✅ Very simple |
| **Quality Instrumentation** | ✅ Full support | ❌ Limited |
| **JPEG-compressed DNG** | 🐛 Buffer overflow (fixable) | ❌ Fails with unclear error |

## Build Tags

The library selection is controlled by build tags:

- **No tags** (or `cgo` only) → `inokone/golibraw`
- **`cgo use_seppedelanghe_libraw`** → `seppedelanghe/go-libraw`

## File Structure

```
internal/indexer/
├── raw_golibraw.go          # inokone/golibraw implementation
├── raw_seppedelanghe.go     # seppedelanghe/go-libraw implementation
└── raw_nocgo.go             # Stub for non-CGO builds

internal/quality/
├── raw_diag_golibraw.go     # Limited diagnostics for golibraw
└── raw_diag_seppedelanghe.go # Full diagnostics for go-libraw

cmd/olsen/
├── version.go               # Reports which library is active
└── benchmark_libraw.go      # Benchmarks both libraries
```

## Benchmarking

### Compare Both Libraries
```bash
make benchmark-libraw
```

This builds with both libraries and generates comparison reports:
- `libraw_benchmark_golibraw.html`
- `libraw_benchmark_seppedelanghe.html`

### Benchmark Individual Libraries
```bash
# Just golibraw
make benchmark-libraw-golibraw

# Just go-libraw
make benchmark-libraw-seppedelanghe
```

## Why Two Libraries?

### Rationale for Dual Support

1. **Quality Instrumentation Requirements**
   - `seppedelanghe/go-libraw` is **required** for full thumbnail quality diagnostics
   - Can capture demosaic algorithm, bit depth, color space, etc.
   - Essential for performance optimization and quality analysis

2. **Fallback Option**
   - `inokone/golibraw` provides a simpler, stable fallback
   - Useful if configuration complexity causes issues
   - Faster to build and test during development

3. **Performance Comparison**
   - Easy to benchmark both approaches
   - Validate that configuration actually improves quality
   - Ensure no significant performance regression

4. **Risk Mitigation**
   - If one library has a critical bug, we can switch immediately
   - Currently: `seppedelanghe` has buffer overflow bug (being fixed)
   - Having both = zero downtime if fix takes time

### Why seppedelanghe is Default

Despite the current buffer overflow bug:
- Provides **essential** configuration capabilities
- Required for quality instrumentation goals
- Bug is **fixable** (see `docs/LIBRAW_FORK_PLAN.md`)
- More capable and modern architecture

## Current Status

### seppedelanghe/go-libraw ⚠️
- **Status**: Default, has known bug
- **Issue**: Buffer overflow with JPEG-compressed DNG files
- **Workaround**: Use golibraw for JPEG-compressed DNGs
- **Fix**: In progress (see `LIBRAW_FORK_PLAN.md`)
- **Quality**: Excellent when working
- **Configuration**: Full control

### inokone/golibraw ✅
- **Status**: Stable, fully working
- **Limitation**: No configuration options
- **Limitation**: Cannot capture diagnostics
- **Quality**: Good (unknown exact settings)
- **Configuration**: None available

## Migration Path

### When Bug is Fixed

1. Update dependency to fixed version
2. Remove workarounds from documentation
3. Keep dual support (still valuable for comparison)
4. Update default recommendation in docs

### If Bug Cannot Be Fixed

1. Make `golibraw` the default
2. Keep `seppedelanghe` available for those who need configuration
3. Accept limited diagnostics capture
4. Document the tradeoff clearly

## Testing

### Test Both Libraries
```bash
# Test with golibraw
CGO_ENABLED=1 go test -tags cgo -v ./internal/indexer/

# Test with go-libraw
CGO_ENABLED=1 go test -tags "cgo use_seppedelanghe_libraw" -v ./internal/indexer/
```

### Integration Tests
```bash
make test-integration-raw
```

## API Differences

### golibraw Implementation
```go
// Simple, no configuration
img, err := golibraw.ImportRaw(path)
```

### go-libraw Implementation
```go
// Full configuration control
processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
    UserQual:    3,              // AHD demosaicing
    OutputBps:   8,              // 8-bit output
    OutputColor: golibraw.SRGB,  // sRGB color space
    UseCameraWb: true,           // Camera white balance
})
img, meta, err := processor.ProcessRaw(path)
```

### Diagnostics Capture

**golibraw** - Limited:
```go
// Cannot capture actual LibRaw settings
diag := &RawDiag{
    Demosaic:    "unknown",
    OutputBPS:   8,          // assumed
    OutputColor: "sRGB",     // assumed
    UseCameraWB: true,       // assumed
}
```

**go-libraw** - Complete:
```go
// Full diagnostic information
img, diag, err := DecodeRawWithDiag(path, &golibraw.ProcessorOptions{
    UserQual: 3,  // Captured as "AHD"
    // ... all settings captured
})
// diag.Demosaic = "AHD"
// diag.OutputBPS = 8
// diag.OutputColor = "sRGB"
```

## Environment Variables

No environment variables needed - library selection is compile-time only via build tags.

## Recommendations

### For Development
- Use `make build-golibraw` for faster iteration
- Simple, stable, works on all formats (except JPEG-compressed DNG)

### For Production
- Use `make build-seppedelanghe` (default)
- Better quality control and diagnostics
- **Caveat**: Avoid JPEG-compressed DNGs until bug is fixed
- Monitor for fix: https://github.com/seppedelanghe/go-libraw/issues

### For Benchmarking
- Always use `make benchmark-libraw` to compare both
- Generates side-by-side HTML reports
- Essential for validating configuration improvements

## Future Plans

1. **Fix Buffer Overflow** (In Progress)
   - Fork `seppedelanghe/go-libraw`
   - Create minimal test case
   - Fix bug and submit PR
   - See `LIBRAW_FORK_PLAN.md`

2. **Optimize Configuration** (Future)
   - Benchmark different demosaic algorithms
   - Test 8-bit vs 16-bit output
   - Measure quality vs performance tradeoffs

3. **Advanced Features** (Future)
   - ICC color profile handling
   - Linear-light resizing
   - Per-camera optimal settings

## Related Documentation

- `docs/LIBRAW_FORK_PLAN.md` - Plan for fixing buffer overflow bug
- `docs/THUMBNAIL_QUALITY_IMPLEMENTATION_STATUS.md` - Quality instrumentation status
- `CLAUDE.md` - Build instructions and architecture overview

## Questions?

Run `make help` to see all available build and benchmark targets.
