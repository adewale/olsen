package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

// HSLToRGB converts HSL color to RGB
func HSLToRGB(h, s, l float64) color.RGBA {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l // achromatic
	} else {
		hue2rgb := func(p, q, t float64) float64 {
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

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// ColorBlock represents a colored rectangle in the image
type ColorBlock struct {
	Hue        float64 // 0-1
	Saturation float64 // 0-1
	Lightness  float64 // 0-1
	Width      int     // pixels
}

// GenerateColorImage creates an image with specified color blocks
func GenerateColorImage(filename string, width, height int, blocks []ColorBlock) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with color blocks horizontally
	x := 0
	for _, block := range blocks {
		c := HSLToRGB(block.Hue, block.Saturation, block.Lightness)

		endX := x + block.Width
		if endX > width {
			endX = width
		}

		for px := x; px < endX; px++ {
			for py := 0; py < height; py++ {
				img.Set(px, py, c)
			}
		}
		x = endX
	}

	// Save as JPEG
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
}

// GenerateGradientImage creates an image with a gradient of colors
func GenerateGradientImage(filename string, width, height int, blocks []ColorBlock) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for i, block := range blocks {
		endX := x + block.Width
		if endX > width {
			endX = width
		}

		// If there's a next block, create gradient
		if i < len(blocks)-1 {
			nextBlock := blocks[i+1]
			for px := x; px < endX; px++ {
				t := float64(px-x) / float64(block.Width)

				// Interpolate HSL values
				h := block.Hue + (nextBlock.Hue-block.Hue)*t
				s := block.Saturation + (nextBlock.Saturation-block.Saturation)*t
				l := block.Lightness + (nextBlock.Lightness-block.Lightness)*t

				c := HSLToRGB(h, s, l)

				for py := 0; py < height; py++ {
					img.Set(px, py, c)
				}
			}
		} else {
			// Last block, solid color
			c := HSLToRGB(block.Hue, block.Saturation, block.Lightness)
			for px := x; px < endX; px++ {
				for py := 0; py < height; py++ {
					img.Set(px, py, c)
				}
			}
		}
		x = endX
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
}

func main() {
	outputDir := "testdata/color_test"

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	width, height := 800, 600

	// Image 1: Dominant Red with some Orange and Yellow
	log.Println("Generating red_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "red_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.0, Saturation: 0.8, Lightness: 0.5, Width: 500},  // Red
		{Hue: 0.05, Saturation: 0.7, Lightness: 0.5, Width: 200}, // Orange
		{Hue: 0.15, Saturation: 0.7, Lightness: 0.5, Width: 100}, // Yellow
	})

	// Image 2: Dominant Orange with Red and Yellow
	log.Println("Generating orange_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "orange_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.05, Saturation: 0.85, Lightness: 0.5, Width: 450},  // Orange
		{Hue: 0.0, Saturation: 0.7, Lightness: 0.5, Width: 200},    // Red
		{Hue: 0.15, Saturation: 0.75, Lightness: 0.55, Width: 150}, // Yellow
	})

	// Image 3: Dominant Yellow with Green and Orange
	log.Println("Generating yellow_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "yellow_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.15, Saturation: 0.9, Lightness: 0.55, Width: 500}, // Yellow
		{Hue: 0.3, Saturation: 0.6, Lightness: 0.4, Width: 200},   // Green
		{Hue: 0.05, Saturation: 0.7, Lightness: 0.5, Width: 100},  // Orange
	})

	// Image 4: Dominant Green with Blue and Yellow
	log.Println("Generating green_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "green_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.33, Saturation: 0.7, Lightness: 0.4, Width: 500},  // Green
		{Hue: 0.55, Saturation: 0.7, Lightness: 0.5, Width: 200},  // Blue
		{Hue: 0.15, Saturation: 0.8, Lightness: 0.55, Width: 100}, // Yellow
	})

	// Image 5: Dominant Blue with Purple and Green
	log.Println("Generating blue_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "blue_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.58, Saturation: 0.8, Lightness: 0.5, Width: 500},  // Blue
		{Hue: 0.75, Saturation: 0.6, Lightness: 0.5, Width: 200},  // Purple
		{Hue: 0.35, Saturation: 0.6, Lightness: 0.45, Width: 100}, // Green
	})

	// Image 6: Dominant Purple with Pink and Blue
	log.Println("Generating purple_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "purple_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.75, Saturation: 0.7, Lightness: 0.5, Width: 500}, // Purple
		{Hue: 0.92, Saturation: 0.7, Lightness: 0.6, Width: 200}, // Pink
		{Hue: 0.6, Saturation: 0.7, Lightness: 0.5, Width: 100},  // Blue
	})

	// Image 7: Dominant Pink with Red and Purple
	log.Println("Generating pink_dominant.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "pink_dominant.jpg"), width, height, []ColorBlock{
		{Hue: 0.92, Saturation: 0.7, Lightness: 0.65, Width: 500}, // Pink
		{Hue: 0.0, Saturation: 0.7, Lightness: 0.5, Width: 200},   // Red
		{Hue: 0.78, Saturation: 0.6, Lightness: 0.5, Width: 100},  // Purple
	})

	// Image 8: Rainbow gradient (all colors)
	log.Println("Generating rainbow.jpg...")
	GenerateGradientImage(filepath.Join(outputDir, "rainbow.jpg"), width, height, []ColorBlock{
		{Hue: 0.0, Saturation: 0.8, Lightness: 0.5, Width: 114},   // Red
		{Hue: 0.05, Saturation: 0.8, Lightness: 0.5, Width: 114},  // Orange
		{Hue: 0.15, Saturation: 0.8, Lightness: 0.5, Width: 114},  // Yellow
		{Hue: 0.33, Saturation: 0.7, Lightness: 0.45, Width: 114}, // Green
		{Hue: 0.58, Saturation: 0.8, Lightness: 0.5, Width: 114},  // Blue
		{Hue: 0.75, Saturation: 0.7, Lightness: 0.5, Width: 115},  // Purple
		{Hue: 0.92, Saturation: 0.7, Lightness: 0.6, Width: 115},  // Pink
	})

	// Image 9: Mixed warm colors (Red, Orange, Yellow dominant)
	log.Println("Generating warm_mix.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "warm_mix.jpg"), width, height, []ColorBlock{
		{Hue: 0.0, Saturation: 0.8, Lightness: 0.5, Width: 300},   // Red
		{Hue: 0.05, Saturation: 0.85, Lightness: 0.5, Width: 300}, // Orange
		{Hue: 0.15, Saturation: 0.9, Lightness: 0.55, Width: 200}, // Yellow
	})

	// Image 10: Mixed cool colors (Blue, Green, Purple dominant)
	log.Println("Generating cool_mix.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "cool_mix.jpg"), width, height, []ColorBlock{
		{Hue: 0.58, Saturation: 0.8, Lightness: 0.5, Width: 300},  // Blue
		{Hue: 0.33, Saturation: 0.7, Lightness: 0.45, Width: 300}, // Green
		{Hue: 0.75, Saturation: 0.7, Lightness: 0.5, Width: 200},  // Purple
	})

	// Image 11: Balanced spectrum (equal amounts of all major hues)
	log.Println("Generating balanced_spectrum.jpg...")
	GenerateColorImage(filepath.Join(outputDir, "balanced_spectrum.jpg"), width, height, []ColorBlock{
		{Hue: 0.0, Saturation: 0.8, Lightness: 0.5, Width: 100},   // Red
		{Hue: 0.05, Saturation: 0.8, Lightness: 0.5, Width: 100},  // Orange
		{Hue: 0.15, Saturation: 0.8, Lightness: 0.5, Width: 100},  // Yellow
		{Hue: 0.33, Saturation: 0.7, Lightness: 0.45, Width: 100}, // Green
		{Hue: 0.5, Saturation: 0.7, Lightness: 0.5, Width: 100},   // Cyan
		{Hue: 0.58, Saturation: 0.8, Lightness: 0.5, Width: 100},  // Blue
		{Hue: 0.75, Saturation: 0.7, Lightness: 0.5, Width: 100},  // Purple
		{Hue: 0.92, Saturation: 0.7, Lightness: 0.6, Width: 100},  // Pink
	})

	// Image 12: Sunset gradient (Red->Orange->Yellow->Purple)
	log.Println("Generating sunset.jpg...")
	GenerateGradientImage(filepath.Join(outputDir, "sunset.jpg"), width, height, []ColorBlock{
		{Hue: 0.75, Saturation: 0.6, Lightness: 0.3, Width: 200},  // Deep Purple
		{Hue: 0.0, Saturation: 0.85, Lightness: 0.5, Width: 250},  // Red
		{Hue: 0.05, Saturation: 0.9, Lightness: 0.55, Width: 200}, // Orange
		{Hue: 0.15, Saturation: 0.9, Lightness: 0.6, Width: 150},  // Yellow
	})

	log.Println("✓ Generated 12 color test images in", outputDir)
	log.Println()
	log.Println("To index these images:")
	log.Println("  ./olsen index", outputDir, "--db color_test.db")
	log.Println()
	log.Println("To test color queries:")
	log.Println("  ./olsen query --db color_test.db --color red --facets")
	log.Println("  ./olsen query --db color_test.db --color blue --facets")
	log.Println("  ./olsen query --db color_test.db --color green,yellow")
}
