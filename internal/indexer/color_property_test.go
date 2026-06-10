package indexer

import (
	"testing"

	"github.com/adewale/olsen/pkg/models"
)

// abs8 returns |a-b| for two uint8 channel values.
func abs8(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}

// TestRGBToHSLBoundsProperty sweeps the RGB cube and asserts every conversion
// lands inside the documented HSL ranges. HSL values are stored in the
// database and drive SQL colour filters, so out-of-range values would
// silently break facet queries.
func TestRGBToHSLBoundsProperty(t *testing.T) {
	t.Parallel()

	const step = 15 // 18^3 ≈ 5.8k samples, includes 0 and 255
	for r := 0; r <= 255; r += step {
		for g := 0; g <= 255; g += step {
			for b := 0; b <= 255; b += step {
				hsl := rgbToHSL(models.Colour{R: uint8(r), G: uint8(g), B: uint8(b)})
				if hsl.H < 0 || hsl.H > 360 {
					t.Fatalf("rgbToHSL(%d,%d,%d): H = %d outside [0, 360]", r, g, b, hsl.H)
				}
				if hsl.S < 0 || hsl.S > 100 {
					t.Fatalf("rgbToHSL(%d,%d,%d): S = %d outside [0, 100]", r, g, b, hsl.S)
				}
				if hsl.L < 0 || hsl.L > 100 {
					t.Fatalf("rgbToHSL(%d,%d,%d): L = %d outside [0, 100]", r, g, b, hsl.L)
				}
			}
		}
	}
}

// TestRGBHSLRoundtripProperty asserts HSLToRGB(rgbToHSL(c)) stays close to c.
// HSL components are quantized to integers (H in 360 steps, S/L in 100), so
// an exact roundtrip is impossible; the tolerance below bounds the loss.
func TestRGBHSLRoundtripProperty(t *testing.T) {
	t.Parallel()

	// Worst-case quantization error: L and S each lose up to 1/200 of full
	// scale, hue up to 0.5 degree; combined this stays under ~6 channel
	// steps. 8 gives headroom without masking real conversion bugs.
	const tolerance = 8

	const step = 15
	for r := 0; r <= 255; r += step {
		for g := 0; g <= 255; g += step {
			for b := 0; b <= 255; b += step {
				in := models.Colour{R: uint8(r), G: uint8(g), B: uint8(b)}
				out := HSLToRGB(rgbToHSL(in))

				if abs8(in.R, out.R) > tolerance || abs8(in.G, out.G) > tolerance || abs8(in.B, out.B) > tolerance {
					t.Fatalf("roundtrip drift > %d: in=(%d,%d,%d) out=(%d,%d,%d) hsl=%+v",
						tolerance, in.R, in.G, in.B, out.R, out.G, out.B, rgbToHSL(in))
				}
			}
		}
	}
}

// TestGreysRoundtripExactly pins the achromatic axis: pure greys have S=0 and
// must survive the roundtrip without hue-induced drift.
func TestGreysRoundtripExactly(t *testing.T) {
	t.Parallel()

	for v := 0; v <= 255; v++ {
		in := models.Colour{R: uint8(v), G: uint8(v), B: uint8(v)}
		hsl := rgbToHSL(in)
		if hsl.S != 0 {
			t.Fatalf("grey %d: S = %d, want 0", v, hsl.S)
		}
		out := HSLToRGB(hsl)
		if out.R != out.G || out.G != out.B {
			t.Fatalf("grey %d roundtripped to non-grey (%d,%d,%d)", v, out.R, out.G, out.B)
		}
		if abs8(in.R, out.R) > 2 { // L quantization only
			t.Fatalf("grey %d roundtrip drift: got %d", v, out.R)
		}
	}
}
