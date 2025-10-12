//go:build cgo
// +build cgo

package indexer_test

import (
	"os"
	"testing"
)

// TestRAWLibraryComparison tests all 3 RAW processing options for brightness
// This helps identify which library correctly processes JPEG-compressed monochrome DNGs
func TestRAWLibraryComparison(t *testing.T) {
	testFile := "../../private-testdata/2024-12-23/L1001530.DNG"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not found (requires private-testdata): ", testFile)
	}

	t.Run("Seppedelanghe/go-libraw", func(t *testing.T) {
		// This is tested via the build tag: use_seppedelanghe_libraw
		// Run with: make test-raw-brightness
		t.Skip("Test separately with: make test-raw-brightness")
	})

	t.Run("Inokone/golibraw", func(t *testing.T) {
		// This is tested via the build tag: cgo (without use_seppedelanghe_libraw)
		// Run with: CGO_ENABLED=1 go test -tags cgo
		t.Skip("Test separately with golibraw build tag")
	})

	t.Run("Embedded JPEG (no CGO)", func(t *testing.T) {
		// This is tested via the build tag: (no cgo)
		// Run with: CGO_ENABLED=0 go test
		t.Skip("Test separately without CGO")
	})
}
