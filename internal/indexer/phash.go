package indexer

import (
	"fmt"
	"image"

	"github.com/corona10/goimagehash"
)

// ComputePerceptualHash computes a perceptual hash (pHash) for an image
func ComputePerceptualHash(img image.Image) (string, error) {
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", fmt.Errorf("failed to compute perceptual hash: %w", err)
	}
	return hash.ToString(), nil
}

// HammingDistance calculates the Hamming distance between two perceptual hashes
func HammingDistance(hash1, hash2 string) (int, error) {
	h1, err := goimagehash.ImageHashFromString(hash1)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hash1: %w", err)
	}

	h2, err := goimagehash.ImageHashFromString(hash2)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hash2: %w", err)
	}

	distance, err := h1.Distance(h2)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	return distance, nil
}

// AreSimilar checks if two images are similar based on their perceptual hashes
// threshold defines the maximum Hamming distance for similarity
// Common thresholds:
//
//	0-5: Identical/near-identical
//	6-10: Very similar
//	11-15: Similar (burst variations)
//	16-20: Somewhat similar
//	21+: Different
func AreSimilar(hash1, hash2 string, threshold int) (bool, error) {
	distance, err := HammingDistance(hash1, hash2)
	if err != nil {
		return false, err
	}
	return distance <= threshold, nil
}
