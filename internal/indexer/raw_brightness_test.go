//go:build cgo && use_seppedelanghe_libraw
// +build cgo,use_seppedelanghe_libraw

package indexer_test

import (
	"image"
	"testing"

	golibraw "github.com/seppedelanghe/go-libraw"
)

// TestRAWBrightnessSettings tests different LibRaw settings to find why images are black
func TestRAWBrightnessSettings(t *testing.T) {
	testFile := "../../testdata/dng/L1001515.DNG"

	testConfigs := []struct {
		name    string
		options golibraw.ProcessorOptions
	}{
		{
			name: "Current settings (8-bit, AHD, sRGB, Camera WB)",
			options: golibraw.ProcessorOptions{
				UserQual:    3, // AHD
				OutputBps:   8,
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			},
		},
		{
			name: "16-bit output",
			options: golibraw.ProcessorOptions{
				UserQual:    3, // AHD
				OutputBps:   16,
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			},
		},
		{
			name: "Linear demosaic (simpler)",
			options: golibraw.ProcessorOptions{
				UserQual:    0, // Linear
				OutputBps:   8,
				OutputColor: golibraw.SRGB,
				UseCameraWb: true,
			},
		},
		{
			name: "Auto white balance instead of camera",
			options: golibraw.ProcessorOptions{
				UserQual:    3, // AHD
				OutputBps:   8,
				OutputColor: golibraw.SRGB,
				UseCameraWb: false,
			},
		},
		{
			name: "No adjustments (raw sensor data)",
			options: golibraw.ProcessorOptions{
				UserQual:     3, // AHD
				OutputBps:    8,
				OutputColor:  golibraw.Raw,
				UseCameraWb:  false,
				NoAutoBright: true,
			},
		},
	}

	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// Recover from LibRaw panic (known bug with JPEG-compressed Monochrom DNGs)
			defer func() {
				if r := recover(); r != nil {
					t.Logf("⚠️  LibRaw panic (known issue with JPEG-compressed monochrome DNGs): %v", r)
					t.Logf("    This is a buffer overflow bug in seppedelanghe/go-libraw")
					t.Logf("    See: https://github.com/seppedelanghe/go-libraw/issues")
				}
			}()

			processor := golibraw.NewProcessor(tc.options)
			img, _, err := processor.ProcessRaw(testFile)
			if err != nil {
				t.Logf("Processing failed (expected for this file): %v", err)
				t.Logf("  LibRaw cannot decode JPEG-compressed Leica M11 Monochrom DNGs")
				return
			}

			brightness := calculateImageBrightness(img)
			histogram := buildImageHistogram(img)

			t.Logf("Settings: %s", tc.name)
			t.Logf("  Brightness: %.1f/255", brightness)
			t.Logf("  Histogram: %s", formatImageHistogram(histogram))
			t.Logf("  Image bounds: %v", img.Bounds())
			t.Logf("  Image type: %T", img)
		})
	}
}

func calculateImageBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	var sum uint64
	var count uint64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (r + g + b) / 3 / 256
			sum += uint64(gray)
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return float64(sum) / float64(count)
}

func buildImageHistogram(img image.Image) [256]int {
	var histogram [256]int
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (r + g + b) / 3 / 256
			if gray < 256 {
				histogram[gray]++
			}
		}
	}

	return histogram
}

func formatImageHistogram(histogram [256]int) string {
	bins := make([]int, 10)
	for i := 0; i < 256; i++ {
		bin := i / 26
		if bin > 9 {
			bin = 9
		}
		bins[bin] += histogram[i]
	}

	maxCount := 0
	for _, count := range bins {
		if count > maxCount {
			maxCount = count
		}
	}

	result := "["
	for i, count := range bins {
		if i > 0 {
			result += " "
		}
		pct := float64(count) / float64(maxCount) * 100
		if pct < 10 {
			result += "_"
		} else if pct < 30 {
			result += "▁"
		} else if pct < 50 {
			result += "▃"
		} else if pct < 70 {
			result += "▅"
		} else if pct < 90 {
			result += "▇"
		} else {
			result += "█"
		}
	}
	result += "]"

	return result
}
