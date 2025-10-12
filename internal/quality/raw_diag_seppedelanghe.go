//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package quality

import (
	"fmt"
	"image"

	golibraw "github.com/seppedelanghe/go-libraw"
)

// DecodeRawWithDiag decodes a RAW file and captures diagnostics
// This wraps the go-libraw library and extracts diagnostic information
func DecodeRawWithDiag(path string, opts *golibraw.ProcessorOptions) (image.Image, *RawDiag, error) {
	// Use default options if not provided
	if opts == nil {
		opts = &golibraw.ProcessorOptions{
			UserQual:    3, // AHD demosaicing (highest quality)
			OutputBps:   8, // 8-bit output
			OutputColor: golibraw.SRGB,
			UseCameraWb: true,
		}
	}

	// Build diagnostics from known configuration
	diag := &RawDiag{
		LibRawEnabled: true,
		Demosaic:      demosaicName(opts.UserQual),
		OutputBPS:     opts.OutputBps,
		OutputColor:   colorSpaceName(opts.OutputColor),
		UseCameraWB:   opts.UseCameraWb,
		HalfSize:      false, // We don't use half_size
	}

	// Process the RAW file
	processor := golibraw.NewProcessor(*opts)
	img, _, err := processor.ProcessRaw(path)
	if err != nil {
		return nil, diag, fmt.Errorf("libraw decode failed: %w", err)
	}

	return img, diag, nil
}

// Helper functions to convert enum values to human-readable names

func demosaicName(userQual int) string {
	switch userQual {
	case 0:
		return "Linear"
	case 1:
		return "VNG"
	case 2:
		return "PPG"
	case 3:
		return "AHD"
	case 4:
		return "DCB"
	case 11:
		return "DHT"
	case 12:
		return "AAHD"
	default:
		return fmt.Sprintf("unknown(%d)", userQual)
	}
}

func colorSpaceName(color golibraw.OutputColor) string {
	switch color {
	case golibraw.Raw:
		return "Raw"
	case golibraw.SRGB:
		return "sRGB"
	case golibraw.AdobeRGB:
		return "AdobeRGB"
	case golibraw.WideGamutRGB:
		return "WideGamutRGB"
	case golibraw.ProPhotoRGB:
		return "ProPhotoRGB"
	case golibraw.XYZ:
		return "XYZ"
	case golibraw.ACES:
		return "ACES"
	case golibraw.DciP3:
		return "DCI-P3"
	case golibraw.Rec2020:
		return "Rec2020"
	default:
		return fmt.Sprintf("unknown(%d)", color)
	}
}
