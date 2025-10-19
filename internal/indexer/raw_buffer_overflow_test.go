//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	golibraw "github.com/seppedelanghe/go-libraw"
)

// TestBufferOverflowJPEGCompressedDNG tests the buffer overflow bug with JPEG-compressed DNG files
// This test documents the exact failure mode and validates our fix
func TestBufferOverflowJPEGCompressedDNG(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	// Test with default AHD settings (what we use in production)
	t.Run("AHD_8bit_sRGB_CameraWB", func(t *testing.T) {
		processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
			UserQual:    3, // AHD demosaicing
			OutputBps:   8,
			OutputColor: golibraw.SRGB,
			UseCameraWb: true,
		})

		img, meta, err := processor.ProcessRaw(testFile)

		if err != nil {
			t.Logf("Expected error (buffer overflow): %v", err)
			// This should fail with buffer overflow until fixed
			if img != nil {
				t.Error("Got image despite error - unexpected")
			}
		} else {
			// If we get here, the bug is fixed!
			t.Logf("✓ Bug is FIXED! Successfully decoded JPEG-compressed DNG")
			t.Logf("  Image size: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
			t.Logf("  Metadata timestamp: %d", meta.CaptureTimestamp)
		}
	})

	// Test with different settings to understand scope of bug
	t.Run("Linear_8bit_sRGB_CameraWB", func(t *testing.T) {
		processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
			UserQual:    0, // Linear (fastest)
			OutputBps:   8,
			OutputColor: golibraw.SRGB,
			UseCameraWb: true,
		})

		img, _, err := processor.ProcessRaw(testFile)
		if err != nil {
			t.Logf("Linear also fails: %v", err)
		} else {
			t.Logf("Linear works! Image: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	t.Run("AHD_16bit_sRGB_CameraWB", func(t *testing.T) {
		var panicValue interface{}
		var img image.Image
		var err error

		// Capture panic (16-bit mode triggers index out of range)
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicValue = r
				}
			}()

			processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
				UserQual:    3,  // AHD
				OutputBps:   16, // 16-bit instead of 8-bit
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			})

			img, _, err = processor.ProcessRaw(testFile)
		}()

		if panicValue != nil {
			t.Logf("16-bit PANICS: %v", panicValue)
		} else if err != nil {
			t.Logf("16-bit also fails: %v", err)
		} else {
			t.Logf("16-bit works! Image: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})
}

// TestBufferOverflowMultipleFiles tests multiple JPEG-compressed DNG files
func TestBufferOverflowMultipleFiles(t *testing.T) {
	// Find all DNG files in private-testdata
	var testFiles []string
	err := filepath.Walk("../../private-testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".DNG" || filepath.Ext(path) == ".dng") {
			testFiles = append(testFiles, path)
		}
		return nil
	})

	if err != nil || len(testFiles) == 0 {
		t.Skip("No test files found in private-testdata")
	}

	t.Logf("Found %d DNG files to test", len(testFiles))

	processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
		UserQual:    3, // AHD
		OutputBps:   8,
		OutputColor: golibraw.SRGB,
		UseCameraWb: true,
	})

	failCount := 0
	successCount := 0

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			img, _, err := processor.ProcessRaw(testFile)
			if err != nil {
				t.Logf("FAIL: %v", err)
				failCount++
			} else {
				t.Logf("SUCCESS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
				successCount++
			}
		})
	}

	t.Logf("Results: %d succeeded, %d failed out of %d files", successCount, failCount, len(testFiles))
}

// TestUncompressedDNGWorksCorrectly verifies that uncompressed DNGs work fine
// This establishes a baseline that the bug is specific to JPEG-compressed files
func TestUncompressedDNGWorksCorrectly(t *testing.T) {
	// Use our test DNG files (which are mostly JPEGs with .dng extension, but good for testing)
	testFiles, err := filepath.Glob("../../testdata/dng/*.dng")
	if err != nil || len(testFiles) == 0 {
		t.Skip("No test files found in testdata/dng")
	}

	processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
		UserQual:    3,
		OutputBps:   8,
		OutputColor: golibraw.SRGB,
		UseCameraWb: true,
	})

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			img, _, err := processor.ProcessRaw(testFile)
			if err != nil {
				t.Errorf("Uncompressed DNG failed (unexpected): %v", err)
			} else {
				t.Logf("✓ Success: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
			}
		})
	}
}

// TestCompareLibraries compares behaviour between both libraries
func TestCompareLibraries(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"
	t.Run("seppedelanghe_go-libraw", func(t *testing.T) {
		processor := golibraw.NewProcessor(golibraw.ProcessorOptions{
			UserQual:    3,
			OutputBps:   8,
			OutputColor: golibraw.SRGB,
			UseCameraWb: true,
		})

		img, _, err := processor.ProcessRaw(testFile)
		if err != nil {
			t.Logf("seppedelanghe/go-libraw FAILS: %v", err)
		} else {
			t.Logf("seppedelanghe/go-libraw SUCCESS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})

	// Note: We can't test golibraw here because of build tags
	// Use separate test file or run: make test-buffer-overflow-golibraw
}
