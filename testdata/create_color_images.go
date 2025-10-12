package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

func main() {
	const width, height = 400, 400

	// Define colors to create
	colors := map[string]color.RGBA{
		"brown_dominant.jpg": {R: 139, G: 69, B: 19, A: 255},   // Brown
		"grey_dominant.jpg":  {R: 128, G: 128, B: 128, A: 255}, // Grey
		"black_dominant.jpg": {R: 10, G: 10, B: 10, A: 255},    // Black
		"white_dominant.jpg": {R: 245, G: 245, B: 245, A: 255}, // White
	}

	for filename, col := range colors {
		filepath := fmt.Sprintf("testdata/color_test/%s", filename)

		// Create a new image
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill with solid color
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, col)
			}
		}

		// Save as JPEG
		f, err := os.Create(filepath)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", filepath, err)
			continue
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
		f.Close()

		if err != nil {
			fmt.Printf("Error encoding %s: %v\n", filepath, err)
			continue
		}

		fmt.Printf("Created %s with RGB(%d, %d, %d)\n", filepath, col.R, col.G, col.B)
	}

	fmt.Printf("\nCreated %d test images\n", len(colors))
}
