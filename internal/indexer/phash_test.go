package indexer

import (
	"image"
	"image/color"
	"testing"
)

func TestComputePerceptualHash(t *testing.T) {
	// Create a simple test image
	img := createSolidColorImage(100, 100, color.RGBA{255, 0, 0, 255})

	hash, err := ComputePerceptualHash(img)
	if err != nil {
		t.Fatalf("ComputePerceptualHash failed: %v", err)
	}

	// Hash should be a non-empty string
	if len(hash) == 0 {
		t.Error("Hash is empty")
	}

	// Hash should be reasonably long (at least 16 characters)
	if len(hash) < 16 {
		t.Errorf("Hash length = %d; want >= 16", len(hash))
	}
}

func TestHammingDistance(t *testing.T) {
	// Create two identical images
	img1 := createSolidColorImage(100, 100, color.RGBA{255, 0, 0, 255})
	hash1, _ := ComputePerceptualHash(img1)

	img2 := createSolidColorImage(100, 100, color.RGBA{255, 0, 0, 255})
	hash2, _ := ComputePerceptualHash(img2)

	// Distance between identical images should be 0
	distance, err := HammingDistance(hash1, hash2)
	if err != nil {
		t.Fatalf("HammingDistance failed: %v", err)
	}

	if distance != 0 {
		t.Errorf("Distance between identical images = %d; want 0", distance)
	}
}

func TestHammingDistanceDifferentImages(t *testing.T) {
	// Create two different images with more structure
	img1 := createGradientImage(100, 100)
	hash1, _ := ComputePerceptualHash(img1)

	img2 := createCheckerboardImage(100, 100)
	hash2, _ := ComputePerceptualHash(img2)

	// Distance between different images should be non-zero
	distance, err := HammingDistance(hash1, hash2)
	if err != nil {
		t.Fatalf("HammingDistance failed: %v", err)
	}

	// We expect some distance for different patterns
	if distance == 0 {
		t.Errorf("Distance between different images = 0; want > 0")
	}
}

func TestHammingDistanceInvalidHash(t *testing.T) {
	hash1 := "0123456789abcdef"
	invalidHash := "invalid"

	_, err := HammingDistance(hash1, invalidHash)
	if err == nil {
		t.Error("Expected error for invalid hash, got nil")
	}
}

func TestAreSimilar(t *testing.T) {
	// Create identical images
	img1 := createSolidColorImage(100, 100, color.RGBA{128, 128, 128, 255})
	hash1, _ := ComputePerceptualHash(img1)

	img2 := createSolidColorImage(100, 100, color.RGBA{128, 128, 128, 255})
	hash2, _ := ComputePerceptualHash(img2)

	// Should be similar with threshold 5
	similar, err := AreSimilar(hash1, hash2, 5)
	if err != nil {
		t.Fatalf("AreSimilar failed: %v", err)
	}

	if !similar {
		t.Error("Identical images should be similar")
	}
}

func TestAreSimilarDifferentImages(t *testing.T) {
	// Create different images
	img1 := createGradientImage(100, 100)
	hash1, _ := ComputePerceptualHash(img1)

	img2 := createCheckerboardImage(100, 100)
	hash2, _ := ComputePerceptualHash(img2)

	// Check distance first
	distance, _ := HammingDistance(hash1, hash2)

	// Should not be similar with threshold lower than actual distance
	if distance > 0 {
		similar, err := AreSimilar(hash1, hash2, 0)
		if err != nil {
			t.Fatalf("AreSimilar failed: %v", err)
		}

		if similar {
			t.Error("Different images should not be similar with threshold 0")
		}
	}
}

func TestPerceptualHashConsistency(t *testing.T) {
	// Same image should produce the same hash
	img := createSolidColorImage(100, 100, color.RGBA{200, 100, 50, 255})

	hash1, err := ComputePerceptualHash(img)
	if err != nil {
		t.Fatalf("First hash computation failed: %v", err)
	}

	hash2, err := ComputePerceptualHash(img)
	if err != nil {
		t.Fatalf("Second hash computation failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Same image produced different hashes: %s vs %s", hash1, hash2)
	}
}

func TestPerceptualHashDifferentSizes(t *testing.T) {
	// pHash should be relatively invariant to size changes
	// Create same content at different sizes
	img1 := createGradientImage(100, 100)
	hash1, _ := ComputePerceptualHash(img1)

	img2 := createGradientImage(200, 200)
	hash2, _ := ComputePerceptualHash(img2)

	// Distance should be low for same content at different sizes
	distance, err := HammingDistance(hash1, hash2)
	if err != nil {
		t.Fatalf("HammingDistance failed: %v", err)
	}

	// Should be similar (distance < 15 is typical for resized versions)
	if distance > 15 {
		t.Errorf("Distance between different-sized same-content images = %d; want < 15", distance)
	}
}

// Helper function to create a gradient image
func createGradientImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a gradient from black to white
			intensity := uint8((x * 255) / width)
			img.Set(x, y, color.RGBA{intensity, intensity, intensity, 255})
		}
	}
	return img
}

// Helper function to create a checkerboard pattern
func createCheckerboardImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	squareSize := 10
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a checkerboard pattern
			if ((x/squareSize)+(y/squareSize))%2 == 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0, 255}) // Black
			}
		}
	}
	return img
}
