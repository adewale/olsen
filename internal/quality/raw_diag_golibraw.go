//go:build cgo && !use_seppedelanghe_libraw
// +build cgo,!use_seppedelanghe_libraw

package quality

import (
	"fmt"
	"image"
)

// DecodeRawWithDiag decodes a RAW file and captures diagnostics
// This wraps the inokone/golibraw library (limited diagnostic information)
func DecodeRawWithDiag(path string) (image.Image, *RawDiag, error) {
	diag := &RawDiag{
		LibRawEnabled: true,
		Demosaic:      "unknown", // golibraw doesn't expose this
		OutputBPS:     8,         // golibraw default (assumed)
		OutputColor:   "sRGB",    // golibraw default (assumed)
		UseCameraWB:   true,      // golibraw default (assumed)
		HalfSize:      false,     // We don't use half_size
	}

	// Note: inokone/golibraw doesn't provide configuration options
	// We have to use whatever defaults it provides internally
	// For full diagnostic capabilities, build with use_seppedelanghe_libraw tag

	return nil, diag, fmt.Errorf("DecodeRawWithDiag not implemented for inokone/golibraw - use indexer.DecodeRaw() instead")
}
