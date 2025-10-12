package indexer

import (
	"image"
	"image/color"
	"testing"

	"github.com/adewale/olsen/pkg/models"
)

func TestRGBToHSL(t *testing.T) {
	tests := []struct {
		name     string
		colour   models.Colour
		expected models.ColourHSL
	}{
		{
			name:     "Red",
			colour:   models.Colour{R: 255, G: 0, B: 0},
			expected: models.ColourHSL{H: 0, S: 100, L: 50},
		},
		{
			name:     "Green",
			colour:   models.Colour{R: 0, G: 255, B: 0},
			expected: models.ColourHSL{H: 120, S: 100, L: 50},
		},
		{
			name:     "Blue",
			colour:   models.Colour{R: 0, G: 0, B: 255},
			expected: models.ColourHSL{H: 240, S: 100, L: 50},
		},
		{
			name:     "White",
			colour:   models.Colour{R: 255, G: 255, B: 255},
			expected: models.ColourHSL{H: 0, S: 0, L: 100},
		},
		{
			name:     "Black",
			colour:   models.Colour{R: 0, G: 0, B: 0},
			expected: models.ColourHSL{H: 0, S: 0, L: 0},
		},
		{
			name:     "Grey",
			colour:   models.Colour{R: 128, G: 128, B: 128},
			expected: models.ColourHSL{H: 0, S: 0, L: 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rgbToHSL(tt.colour)
			// Allow small tolerance for floating point calculations
			if !hslClose(result, tt.expected, 2) {
				t.Errorf("rgbToHSL(%+v) = %+v; want %+v", tt.colour, result, tt.expected)
			}
		})
	}
}

func TestHSLToRGB(t *testing.T) {
	tests := []struct {
		name     string
		hsl      models.ColourHSL
		expected models.Colour
	}{
		{
			name:     "Red",
			hsl:      models.ColourHSL{H: 0, S: 100, L: 50},
			expected: models.Colour{R: 255, G: 0, B: 0},
		},
		{
			name:     "Green",
			hsl:      models.ColourHSL{H: 120, S: 100, L: 50},
			expected: models.Colour{R: 0, G: 255, B: 0},
		},
		{
			name:     "Blue",
			hsl:      models.ColourHSL{H: 240, S: 100, L: 50},
			expected: models.Colour{R: 0, G: 0, B: 255},
		},
		{
			name:     "White",
			hsl:      models.ColourHSL{H: 0, S: 0, L: 100},
			expected: models.Colour{R: 255, G: 255, B: 255},
		},
		{
			name:     "Black",
			hsl:      models.ColourHSL{H: 0, S: 0, L: 0},
			expected: models.Colour{R: 0, G: 0, B: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HSLToRGB(tt.hsl)
			if !colourClose(result, tt.expected, 2) {
				t.Errorf("HSLToRGB(%+v) = %+v; want %+v", tt.hsl, result, tt.expected)
			}
		})
	}
}

func TestColourDistance(t *testing.T) {
	tests := []struct {
		name      string
		c1        models.Colour
		c2        models.Colour
		expected  float64
		tolerance float64
	}{
		{
			name:      "Identical colours",
			c1:        models.Colour{R: 100, G: 100, B: 100},
			c2:        models.Colour{R: 100, G: 100, B: 100},
			expected:  0,
			tolerance: 0.01,
		},
		{
			name:      "Black and white",
			c1:        models.Colour{R: 0, G: 0, B: 0},
			c2:        models.Colour{R: 255, G: 255, B: 255},
			expected:  441.67, // sqrt(255^2 + 255^2 + 255^2)
			tolerance: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColourDistance(tt.c1, tt.c2)
			if !floatClose(result, tt.expected, tt.tolerance) {
				t.Errorf("ColourDistance(%+v, %+v) = %f; want %f", tt.c1, tt.c2, result, tt.expected)
			}
		})
	}
}

func TestExtractColourPalette(t *testing.T) {
	// Create a simple test image with known colors
	img := createTestImage(100, 100, []color.RGBA{
		{255, 0, 0, 255}, // Red
		{0, 255, 0, 255}, // Green
		{0, 0, 255, 255}, // Blue
	})

	colours, err := ExtractColourPalette(img, 3)
	if err != nil {
		t.Fatalf("ExtractColourPalette failed: %v", err)
	}

	if len(colours) != 3 {
		t.Errorf("Expected 3 colours, got %d", len(colours))
	}

	// Check that weights sum to approximately 1.0
	totalWeight := 0.0
	for _, c := range colours {
		totalWeight += c.Weight
	}

	if !floatClose(totalWeight, 1.0, 0.01) {
		t.Errorf("Total weight = %f; want 1.0", totalWeight)
	}
}

func TestExtractColourPaletteInvalidCount(t *testing.T) {
	img := createTestImage(10, 10, []color.RGBA{{255, 0, 0, 255}})
	_, err := ExtractColourPalette(img, 0)
	if err == nil {
		t.Error("Expected error for numColors = 0, got nil")
	}
}

func TestToStandardColour(t *testing.T) {
	c := models.Colour{R: 100, G: 150, B: 200}
	stdColor := ToStandardColour(c)
	r, g, b, a := stdColor.RGBA()

	// RGBA() returns 16-bit values
	if r>>8 != uint32(c.R) || g>>8 != uint32(c.G) || b>>8 != uint32(c.B) || a>>8 != 255 {
		t.Errorf("ToStandardColour conversion failed: got (%d,%d,%d,%d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestFromStandardColour(t *testing.T) {
	stdColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	c := FromStandardColour(stdColor)

	if c.R != 100 || c.G != 150 || c.B != 200 {
		t.Errorf("FromStandardColour conversion failed: got %+v", c)
	}
}

// Helper functions

func hslClose(a, b models.ColourHSL, tolerance int) bool {
	return absInt(a.H-b.H) <= int(tolerance) &&
		absInt(a.S-b.S) <= int(tolerance) &&
		absInt(a.L-b.L) <= int(tolerance)
}

func colourClose(a, b models.Colour, tolerance int) bool {
	return absInt(int(a.R)-int(b.R)) <= tolerance &&
		absInt(int(a.G)-int(b.G)) <= tolerance &&
		absInt(int(a.B)-int(b.B)) <= tolerance
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func floatClose(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func createTestImage(width, height int, colors []color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	pixelsPerColor := (width * height) / len(colors)

	colorIndex := 0
	pixelCount := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, colors[colorIndex])
			pixelCount++
			if pixelCount >= pixelsPerColor && colorIndex < len(colors)-1 {
				colorIndex++
				pixelCount = 0
			}
		}
	}

	return img
}
