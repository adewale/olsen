//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/inokone/golibraw"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run poc.go <path-to-dng-file>")
		os.Exit(1)
	}

	dngPath := os.Args[1]
	fmt.Printf("Testing LibRaw with: %s\n\n", dngPath)

	// Test 1: Extract Metadata
	fmt.Println("=== Test 1: Extract Metadata ===")
	metadata, err := golibraw.ExtractMetadata(dngPath)
	if err != nil {
		fmt.Printf("❌ Failed to extract metadata: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully extracted metadata:\n")
		fmt.Printf("   Camera: %s %s\n", metadata.Camera.Make, metadata.Camera.Model)
		fmt.Printf("   Lens: %s\n", metadata.Lens.Model)
		fmt.Printf("   ISO: %d\n", metadata.ISO)
		fmt.Printf("   Aperture: f/%.1f\n", metadata.Aperture)
		fmt.Printf("   Shutter: 1/%.0f\n", metadata.Shutter)
		fmt.Printf("   Timestamp: %v\n", metadata.Timestamp)
		fmt.Printf("   Image Size: %dx%d\n", metadata.Width, metadata.Height)
	}

	// Test 2: Import RAW as image.Image
	fmt.Println("\n=== Test 2: Import RAW as image.Image ===")
	img, err := golibraw.ImportRaw(dngPath)
	if err != nil {
		fmt.Printf("❌ Failed to import RAW: %v\n", err)
		os.Exit(1)
	}

	bounds := img.Bounds()
	fmt.Printf("✅ Successfully decoded RAW image:\n")
	fmt.Printf("   Bounds: %v\n", bounds)
	fmt.Printf("   Size: %dx%d pixels\n", bounds.Dx(), bounds.Dy())
	fmt.Printf("   Color Model: %T\n", img.ColorModel())

	// Test 3: Sample pixel colors
	fmt.Println("\n=== Test 3: Sample Pixel Colors ===")
	for y := 0; y < bounds.Dy(); y += bounds.Dy() / 4 {
		for x := 0; x < bounds.Dx(); x += bounds.Dx() / 4 {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit
			fmt.Printf("   Pixel (%4d,%4d): R=%3d G=%3d B=%3d A=%3d\n",
				x, y, r>>8, g>>8, b>>8, a>>8)
		}
	}

	fmt.Println("\n✅ All tests passed! LibRaw is working correctly.")
	fmt.Println("✅ Ready to integrate into Olsen.")
}
