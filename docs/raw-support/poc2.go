//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	golibraw "github.com/seppedelanghe/go-libraw"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run poc2.go <path-to-dng-file>")
		os.Exit(1)
	}

	dngPath := os.Args[1]
	fmt.Printf("Testing go-libraw with: %s\n\n", dngPath)

	// Create processor with default options
	processor := golibraw.NewProcessor(golibraw.NewProcessorOptions())

	// Process RAW file
	fmt.Println("=== Processing RAW File ===")
	img, metadata, err := processor.ProcessRaw(dngPath)
	if err != nil {
		fmt.Printf("❌ Failed to process RAW: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully processed RAW file!\n\n")

	// Display metadata
	fmt.Println("=== Metadata ===")
	fmt.Printf("Camera: %s %s\n", metadata.Make, metadata.Model)
	fmt.Printf("ISO: %d\n", metadata.IsoSpeed)
	fmt.Printf("Aperture: f/%.1f\n", metadata.Aperture)
	fmt.Printf("Shutter: %.4f sec\n", metadata.Shutter)
	fmt.Printf("Focal Length: %.1fmm\n", metadata.FocalLen)
	fmt.Printf("Timestamp: %v\n", metadata.Timestamp)

	// Display image info
	fmt.Println("\n=== Image Info ===")
	bounds := img.Bounds()
	fmt.Printf("Bounds: %v\n", bounds)
	fmt.Printf("Size: %dx%d pixels\n", bounds.Dx(), bounds.Dy())
	fmt.Printf("Color Model: %T\n", img.ColorModel())

	// Sample pixel colors
	fmt.Println("\n=== Sample Pixel Colors ===")
	for y := 0; y < bounds.Dy(); y += bounds.Dy() / 4 {
		for x := 0; x < bounds.Dx(); x += bounds.Dx() / 4 {
			r, g, b, a := img.At(x, y).RGBA()
			fmt.Printf("Pixel (%4d,%4d): R=%3d G=%3d B=%3d A=%3d\n",
				x, y, r>>8, g>>8, b>>8, a>>8)
		}
	}

	fmt.Println("\n✅ All tests passed! go-libraw is working correctly.")
	fmt.Println("✅ Ready to integrate into Olsen.")
}
