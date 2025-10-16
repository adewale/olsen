//go:build cgo && use_golibraw
// +build cgo,use_golibraw

package indexer_test

import (
	"os"
	"path/filepath"
	"testing"

	golibraw "github.com/inokone/golibraw"
)

// TestGolibrawJPEGCompressedDNG tests how inokone/golibraw handles JPEG-compressed DNG
// This provides a comparison baseline
func TestGolibrawJPEGCompressedDNG(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found: ", testFile)
	}

	t.Run("golibraw_default_settings", func(t *testing.T) {
		img, err := golibraw.ImportRaw(testFile)

		if err != nil {
			t.Logf("golibraw FAILS on JPEG-compressed DNG: %v", err)
		} else if img == nil {
			t.Error("golibraw returned nil image without error")
		} else {
			t.Logf("✓ golibraw SUCCESS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}
	})
}

// TestGolibrawMultipleFiles tests multiple DNG files with golibraw
func TestGolibrawMultipleFiles(t *testing.T) {
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
		t.Skip("No test files found")
	}

	t.Logf("Testing %d DNG files with golibraw", len(testFiles))

	failCount := 0
	successCount := 0

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			img, err := golibraw.ImportRaw(testFile)
			if err != nil {
				t.Logf("FAIL: %v", err)
				failCount++
			} else if img == nil {
				t.Error("Got nil image without error")
				failCount++
			} else {
				t.Logf("SUCCESS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
				successCount++
			}
		})
	}

	t.Logf("golibraw results: %d succeeded, %d failed out of %d files", successCount, failCount, len(testFiles))
}

// TestGolibrawUncompressedDNG tests with regular (non-JPEG-compressed) DNG files
func TestGolibrawUncompressedDNG(t *testing.T) {
	testFiles, err := filepath.Glob("../../testdata/dng/*.dng")
	if err != nil || len(testFiles) == 0 {
		t.Skip("No test files found in testdata/dng")
	}

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			img, err := golibraw.ImportRaw(testFile)
			if err != nil {
				t.Logf("FAIL (these are fake DNGs): %v", err)
			} else if img != nil {
				t.Logf("SUCCESS: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
			}
		})
	}
}
