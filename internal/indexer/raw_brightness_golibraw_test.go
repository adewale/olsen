//go:build cgo && !use_seppedelanghe_libraw
// +build cgo,!use_seppedelanghe_libraw

package indexer_test

// This test is commented out because:
// 1. It requires private test data that doesn't exist in the repo
// 2. The helper functions (calculateImageBrightness, etc.) need to be moved to a shared test helper file
// 3. This is a diagnostic test for the golibraw implementation, not critical functionality

// See raw_brightness_test.go for the seppedelanghe version that works
