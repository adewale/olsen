package indexer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"

	"github.com/nfnt/resize"

	"github.com/ade/olsen/pkg/models"
)

// GenerateThumbnails generates multiple thumbnail sizes from an image file
// Each thumbnail preserves aspect ratio with longest edge constrained to the specified size
func GenerateThumbnails(filePath string) (map[models.ThumbnailSize][]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	thumbnails := make(map[models.ThumbnailSize][]byte)

	sizes := []struct {
		name         models.ThumbnailSize
		maxDimension uint
	}{
		{models.ThumbnailTiny, 64},
		{models.ThumbnailSmall, 256},
		{models.ThumbnailMedium, 512},
		{models.ThumbnailLarge, 1024},
	}

	for _, size := range sizes {
		// Preserve aspect ratio by constraining longest edge
		bounds := img.Bounds()
		width := uint(bounds.Dx())
		height := uint(bounds.Dy())

		var newWidth, newHeight uint
		if width > height {
			// Landscape: constrain width
			newWidth = size.maxDimension
			newHeight = 0 // resize library will calculate to preserve aspect ratio
		} else {
			// Portrait or square: constrain height
			newWidth = 0 // resize library will calculate to preserve aspect ratio
			newHeight = size.maxDimension
		}

		// Generate thumbnail using Lanczos3 resampling
		thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

		// Encode as JPEG with quality 85
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail %s: %w", size.name, err)
		}

		thumbnails[size.name] = buf.Bytes()
	}

	return thumbnails, nil
}

// GenerateThumbnailsFromImage generates thumbnails from an already-decoded image
func GenerateThumbnailsFromImage(img image.Image) (map[models.ThumbnailSize][]byte, error) {
	thumbnails := make(map[models.ThumbnailSize][]byte)

	sizes := []struct {
		name         models.ThumbnailSize
		maxDimension uint
	}{
		{models.ThumbnailTiny, 64},
		{models.ThumbnailSmall, 256},
		{models.ThumbnailMedium, 512},
		{models.ThumbnailLarge, 1024},
	}

	for _, size := range sizes {
		// Preserve aspect ratio by constraining longest edge
		bounds := img.Bounds()
		width := uint(bounds.Dx())
		height := uint(bounds.Dy())

		var newWidth, newHeight uint
		if width > height {
			newWidth = size.maxDimension
			newHeight = 0
		} else {
			newWidth = 0
			newHeight = size.maxDimension
		}

		// Generate thumbnail using Lanczos3 resampling
		thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

		// Convert Gray to RGBA for JPEG encoding (JPEG encoder doesn't support Gray directly)
		var encodableThumb image.Image = thumb
		if grayImg, ok := thumb.(*image.Gray); ok {
			rgba := image.NewRGBA(grayImg.Bounds())
			for y := grayImg.Bounds().Min.Y; y < grayImg.Bounds().Max.Y; y++ {
				for x := grayImg.Bounds().Min.X; x < grayImg.Bounds().Max.X; x++ {
					rgba.Set(x, y, grayImg.GrayAt(x, y))
				}
			}
			encodableThumb = rgba
		}

		// Encode as JPEG with quality 85
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, encodableThumb, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail %s: %w", size.name, err)
		}

		thumbnails[size.name] = buf.Bytes()
	}

	return thumbnails, nil
}
