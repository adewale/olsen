//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"

	"golang.org/x/image/bmp"
)

func main() {
	// Create JPEG test image
	jpegImg := createTestImage(800, 600, color.RGBA{100, 150, 200, 255})
	jpegFile, err := os.Create("testdata/photos/test1.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer jpegFile.Close()

	if err := jpeg.Encode(jpegFile, jpegImg, &jpeg.Options{Quality: 90}); err != nil {
		log.Fatal(err)
	}
	log.Println("Created test1.jpg")

	// Create another JPEG
	jpegImg2 := createTestImage(1200, 800, color.RGBA{200, 100, 50, 255})
	jpegFile2, err := os.Create("testdata/photos/test2.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer jpegFile2.Close()

	if err := jpeg.Encode(jpegFile2, jpegImg2, &jpeg.Options{Quality: 90}); err != nil {
		log.Fatal(err)
	}
	log.Println("Created test2.jpg")

	// Create BMP test image (simulating a scan)
	bmpImg := createTestImage(600, 800, color.RGBA{250, 240, 230, 255})
	bmpFile, err := os.Create("testdata/photos/scan1.bmp")
	if err != nil {
		log.Fatal(err)
	}
	defer bmpFile.Close()

	if err := bmp.Encode(bmpFile, bmpImg); err != nil {
		log.Fatal(err)
	}
	log.Println("Created scan1.bmp")

	// Create subfolder with more images
	os.MkdirAll("testdata/photos/subfolder", 0755)

	jpegImg3 := createGradientImage(640, 480)
	jpegFile3, err := os.Create("testdata/photos/subfolder/test3.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer jpegFile3.Close()

	if err := jpeg.Encode(jpegFile3, jpegImg3, &jpeg.Options{Quality: 85}); err != nil {
		log.Fatal(err)
	}
	log.Println("Created subfolder/test3.jpg")

	log.Println("All test fixtures created successfully!")
}

func createTestImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Add some variation
			variation := uint8((x + y) % 50)
			img.Set(x, y, color.RGBA{
				R: c.R + variation,
				G: c.G + variation,
				B: c.B + variation,
				A: 255,
			})
		}
	}
	return img
}

func createGradientImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			intensity := uint8((x * 255) / width)
			img.Set(x, y, color.RGBA{intensity, intensity, intensity, 255})
		}
	}
	return img
}
