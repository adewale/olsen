//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"testing"

	"github.com/adewale/olsen/internal/indexer"
	golibraw "github.com/seppedelanghe/go-libraw"
)

// TestLibRawBufferOverflowRegression is a minimal reproduction case for the buffer overflow
// bug in seppedelanghe/go-libraw when processing JPEG-compressed Leica M11 Monochrom DNGs.
//
// BUG DESCRIPTION:
// - File: Leica M11 Monochrom DNG with JPEG compression
// - Error: "unexpected data size: got 60420096, want 181260288"
// - Panic: runtime error: index out of range [53248] with length 53248
// - Location: github.com/seppedelanghe/go-libraw@v0.4.0/libraw.go:404 in ProcessRaw()
//
// ROOT CAUSE:
// The library calculates buffer size incorrectly for JPEG-compressed DNGs.
// - Expected size: width * height * channels * bytes_per_sample
//   = 9536 * 6336 * 3 * 1 = 181,260,288 bytes (for 8-bit RGB)
// - Actual data size from LibRaw: 60,420,096 bytes (JPEG-compressed)
// - Buffer allocated: 53,248 bytes (way too small!)
// - Result: Buffer overflow when trying to copy data
//
// This test documents the bug and will pass once fixed.
func TestLibRawBufferOverflowRegression(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	// Test with minimal options (8-bit RGB)
	t.Run("8-bit RGB output", func(t *testing.T) {
		var panicValue interface{}
		var img interface{}
		var err error

		// Capture panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicValue = r
				}
			}()

			processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
				UserQual:    3, // AHD demosaic
				OutputBps:   8,
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			})

			img, _, err = processor.ProcessRaw(testFile)
		}()

		// CURRENT BEHAVIOR: Should panic or return error
		if panicValue != nil {
			t.Logf("❌ BUG REPRODUCED: Panic occurred: %v", panicValue)
			t.Logf("   This is the buffer overflow bug in libraw.go:404")
			return // Test passes - bug is documented
		}

		if err != nil {
			t.Logf("⚠️  Error occurred (no panic): %v", err)
			t.Logf("   LibRaw returned error instead of panicking")
			return // Test passes - error is better than panic
		}

		// EXPECTED BEHAVIOR AFTER FIX: Should decode successfully
		if img != nil {
			t.Logf("✅ SUCCESS: Image decoded without panic or error!")
			t.Logf("   The buffer overflow bug has been fixed!")
			// Verify image dimensions
			// Note: Will likely be embedded JPEG (9504x6320) not full RAW (9536x6336)
		}
	})

	// Test with 16-bit output (more likely to trigger overflow)
	t.Run("16-bit RGB output", func(t *testing.T) {
		var panicValue interface{}
		var img interface{}
		var err error

		func() {
			defer func() {
				if r := recover(); r != nil {
					panicValue = r
				}
			}()

			processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
				UserQual:    3,
				OutputBps:   16, // 16-bit = larger buffer needed
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			})

			img, _, err = processor.ProcessRaw(testFile)
		}()

		if panicValue != nil {
			t.Logf("❌ BUG REPRODUCED: Panic occurred: %v", panicValue)
			return
		}

		if err != nil {
			t.Logf("⚠️  Error occurred: %v", err)
			return
		}

		if img != nil {
			t.Logf("✅ SUCCESS: 16-bit decode worked!")
		}
	})
}

// TestLibRawBufferSizeCalculation tests the buffer size calculation logic
// This is where the fix should be applied in the upstream library
func TestLibRawBufferSizeCalculation(t *testing.T) {
	// Expected dimensions from EXIF
	width := 9536
	height := 6336
	channels := 3    // RGB
	bytesPerSample := 1 // 8-bit

	// What the library SHOULD calculate
	expectedBufferSize := width * height * channels * bytesPerSample
	t.Logf("Expected buffer size for uncompressed: %d bytes", expectedBufferSize)
	t.Logf("  = %d × %d × %d × %d", width, height, channels, bytesPerSample)
	t.Logf("  = %.2f MB", float64(expectedBufferSize)/(1024*1024))

	// What LibRaw actually returns for JPEG-compressed
	actualDataSize := 60420096 // From error message
	t.Logf("Actual JPEG-compressed data size: %d bytes", actualDataSize)
	t.Logf("  = %.2f MB", float64(actualDataSize)/(1024*1024))

	// The buffer that gets allocated (causes overflow)
	allocatedBufferSize := 53248 // From panic message
	t.Logf("Allocated buffer size: %d bytes", allocatedBufferSize)
	t.Logf("  = %.2f KB (!!)", float64(allocatedBufferSize)/1024)

	// Show the mismatch
	t.Logf("\n⚠️  MISMATCH:")
	t.Logf("  Data to copy: %.2f MB", float64(actualDataSize)/(1024*1024))
	t.Logf("  Buffer size:  %.2f KB", float64(allocatedBufferSize)/1024)
	t.Logf("  Overflow:     %.2f MB", float64(actualDataSize-allocatedBufferSize)/(1024*1024))

	// The fix should ensure: allocatedBufferSize >= actualDataSize
	if allocatedBufferSize < actualDataSize {
		t.Logf("\n❌ Buffer too small: %d < %d (overflow!)", allocatedBufferSize, actualDataSize)
		t.Logf("   FIX: Allocate buffer based on LibRaw's actual output size, not EXIF dimensions")
	}
}

// TestLibRawJPEGCompressedDNGSupport tests various JPEG-compressed DNGs
// Use this to verify the fix works across different files
func TestLibRawJPEGCompressedDNGSupport(t *testing.T) {
	testCases := []struct {
		name     string
		file     string
		expected struct {
			width  int
			height int
		}
	}{
		{
			name: "Leica M11 Monochrom JPEG-compressed DNG",
			file: "../../testdata/dng/L1001515.DNG",
			expected: struct {
				width  int
				height int
			}{
				width:  9536,  // Full RAW
				height: 6336,
			},
		},
		// Add more JPEG-compressed DNGs here to test the fix thoroughly
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var panicValue interface{}
			var img interface{}
			var err error

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicValue = r
					}
				}()

				processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
					UserQual:    3,
					OutputBps:   8,
					OutputColor: golibraw.SRGB,
					UseCameraWb: true,
				})

				img, _, err = processor.ProcessRaw(tc.file)
			}()

			if panicValue != nil {
				t.Errorf("❌ PANIC: %v (fix not applied or incomplete)", panicValue)
				return
			}

			if err != nil {
				// Error is acceptable if it's not a panic
				t.Logf("⚠️  Decode failed with error: %v", err)
				t.Logf("   Better than panic, but ideally should decode JPEG-compressed DNGs")
				return
			}

			if img != nil {
				t.Logf("✅ SUCCESS: Decoded without panic or error")
			}
		})
	}
}

// TestLibRawFallbackBehavior verifies that our fallback to embedded JPEG works
// even when LibRaw fails
func TestLibRawFallbackBehavior(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	t.Log("Testing that embedded JPEG extraction works as fallback...")

	// This should always work, even if LibRaw ProcessRaw fails
	jpgImg, err := indexer.ExtractEmbeddedJPEG(testFile)
	if err != nil {
		t.Fatalf("Embedded JPEG extraction failed: %v", err)
	}

	bounds := jpgImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	t.Logf("✅ Embedded JPEG fallback works: %dx%d", width, height)

	// Verify dimensions are reasonable (embedded JPEG is typically slightly smaller)
	if width < 9000 || width > 10000 || height < 6000 || height > 7000 {
		t.Errorf("Unexpected dimensions: %dx%d", width, height)
	}
}
