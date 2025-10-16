//go:build cgo && !use_golibraw && !use_seppedelanghe_libraw
// +build cgo,!use_golibraw,!use_seppedelanghe_libraw

package indexer

import (
	"errors"
	"image"
)

// LibRawImpl identifies that RAW support is disabled in non-CGO builds
const LibRawImpl = "disabled (CGO required)"

// DecodeRaw stub for non-CGO builds
func DecodeRaw(path string) (image.Image, error) {
	return nil, errors.New("RAW support requires CGO and LibRaw (build with CGO_ENABLED=1)")
}

// IsRawSupported returns false in non-CGO builds
func IsRawSupported() bool {
	return false
}

// ExtractEmbeddedJPEG stub for non-CGO builds
func ExtractEmbeddedJPEG(path string) (image.Image, error) {
	return nil, errors.New("embedded JPEG extraction requires CGO build")
}
