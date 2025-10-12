package indexer

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/mccutchen/palettor"

	"github.com/adewale/olsen/pkg/models"
)

// ExtractColourPalette extracts the dominant colours from an image
func ExtractColourPalette(img image.Image, numColours int) ([]models.DominantColour, error) {
	if numColours <= 0 {
		return nil, fmt.Errorf("numColours must be positive")
	}

	// Use k-means clustering to find dominant colours
	// palettor.Extract(maxIterations, numColours, img)
	palette, err := palettor.Extract(100, numColours, img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract palette: %w", err)
	}

	// Extract colours and weights from palette
	colours := make([]models.DominantColour, 0, numColours)

	for _, entry := range palette.Entries() {
		r, g, b, _ := entry.Color.RGBA()

		// Convert from 16-bit to 8-bit colour
		colour := models.Colour{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
		}

		// Convert to HSL
		hsl := rgbToHSL(colour)

		colours = append(colours, models.DominantColour{
			Colour: colour,
			HSL:    hsl,
			Weight: entry.Weight,
		})
	}

	return colours, nil
}

// rgbToHSL converts an RGB colour to HSL colour space
func rgbToHSL(c models.Colour) models.ColourHSL {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	delta := max - min

	// Calculate lightness
	l := (max + min) / 2.0

	var h, s float64

	if delta == 0 {
		// Achromatic (grey)
		h = 0
		s = 0
	} else {
		// Calculate saturation
		if l < 0.5 {
			s = delta / (max + min)
		} else {
			s = delta / (2.0 - max - min)
		}

		// Calculate hue
		switch max {
		case r:
			h = ((g - b) / delta)
			if g < b {
				h += 6
			}
		case g:
			h = ((b - r) / delta) + 2
		case b:
			h = ((r - g) / delta) + 4
		}
		h *= 60
	}

	return models.ColourHSL{
		H: int(math.Round(h)),
		S: int(math.Round(s * 100)),
		L: int(math.Round(l * 100)),
	}
}

// HSLToRGB converts an HSL colour to RGB colour space
func HSLToRGB(hsl models.ColourHSL) models.Colour {
	h := float64(hsl.H) / 360.0
	s := float64(hsl.S) / 100.0
	l := float64(hsl.L) / 100.0

	var r, g, b float64

	if s == 0 {
		// Achromatic
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return models.Colour{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// ColourDistance calculates the Euclidean distance between two colours in RGB space
func ColourDistance(c1, c2 models.Colour) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// ToStandardColour converts models.Colour to color.Color interface
func ToStandardColour(c models.Colour) color.Color {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

// FromStandardColour converts color.Color to models.Colour
func FromStandardColour(c color.Color) models.Colour {
	r, g, b, _ := c.RGBA()
	return models.Colour{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
	}
}
