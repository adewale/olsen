package quality

import (
	"image"
	"image/color"
	"math"
)

// Metrics holds quality assessment metrics for a thumbnail
type Metrics struct {
	SSIM      float64 // Structural Similarity Index (0-1, higher is better)
	PSNR      float64 // Peak Signal-to-Noise Ratio (dB, higher is better)
	Sharpness float64 // Laplacian variance (higher is sharper)
	MSE       float64 // Mean Squared Error (lower is better)
}

// ComputeSSIM calculates the Structural Similarity Index between two images
// Based on the Wang et al. 2004 paper "Image Quality Assessment: From Error Visibility to Structural Similarity"
// Returns value between 0 (completely different) and 1 (identical)
func ComputeSSIM(img1, img2 image.Image) (float64, error) {
	// Ensure images are same size
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		// If images aren't the same size, resize to match
		// For now, we'll compute on the intersection
		minWidth := min(bounds1.Dx(), bounds2.Dx())
		minHeight := min(bounds1.Dy(), bounds2.Dy())
		bounds1 = image.Rect(0, 0, minWidth, minHeight)
		bounds2 = image.Rect(0, 0, minWidth, minHeight)
	}

	// Constants from the paper
	const (
		k1 = 0.01
		k2 = 0.03
		L  = 255.0 // dynamic range of pixel values (8-bit)
	)

	c1 := (k1 * L) * (k1 * L)
	c2 := (k2 * L) * (k2 * L)

	// Convert to grayscale and compute statistics
	var meanX, meanY, varX, varY, covXY float64
	var n float64

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			grayX := rgbaToGray(img1.At(x, y))
			grayY := rgbaToGray(img2.At(x, y))

			meanX += grayX
			meanY += grayY
			n++
		}
	}

	meanX /= n
	meanY /= n

	// Compute variance and covariance
	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			grayX := rgbaToGray(img1.At(x, y))
			grayY := rgbaToGray(img2.At(x, y))

			diffX := grayX - meanX
			diffY := grayY - meanY

			varX += diffX * diffX
			varY += diffY * diffY
			covXY += diffX * diffY
		}
	}

	varX /= (n - 1)
	varY /= (n - 1)
	covXY /= (n - 1)

	// SSIM formula
	numerator := (2*meanX*meanY + c1) * (2*covXY + c2)
	denominator := (meanX*meanX + meanY*meanY + c1) * (varX + varY + c2)

	ssim := numerator / denominator

	return ssim, nil
}

// ComputePSNR calculates Peak Signal-to-Noise Ratio between two images
// Returns value in dB (typically 20-50 dB, higher is better)
// Infinite PSNR means images are identical
func ComputePSNR(img1, img2 image.Image) (float64, error) {
	mse := ComputeMSE(img1, img2)

	if mse == 0 {
		return math.Inf(1), nil // Perfect match
	}

	maxPixelValue := 255.0
	psnr := 20 * math.Log10(maxPixelValue/math.Sqrt(mse))

	return psnr, nil
}

// ComputeMSE calculates Mean Squared Error between two images
// Lower values indicate more similar images
func ComputeMSE(img1, img2 image.Image) float64 {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	minWidth := min(bounds1.Dx(), bounds2.Dx())
	minHeight := min(bounds1.Dy(), bounds2.Dy())

	var sumSquaredError float64
	var n float64

	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			diffR := float64(r1) - float64(r2)
			diffG := float64(g1) - float64(g2)
			diffB := float64(b1) - float64(b2)

			sumSquaredError += diffR*diffR + diffG*diffG + diffB*diffB
			n += 3 // Three channels
		}
	}

	return sumSquaredError / n
}

// ComputeSharpness calculates the Laplacian variance as a measure of sharpness
// Higher values indicate sharper images
// Based on the "variance of Laplacian" method commonly used in focus detection
func ComputeSharpness(img image.Image) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert to grayscale first
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	// Apply Laplacian kernel
	// Kernel:  0  1  0
	//          1 -4  1
	//          0  1  0
	var laplacianSum float64
	var n float64

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			center := float64(gray.GrayAt(x, y).Y)
			top := float64(gray.GrayAt(x, y-1).Y)
			bottom := float64(gray.GrayAt(x, y+1).Y)
			left := float64(gray.GrayAt(x-1, y).Y)
			right := float64(gray.GrayAt(x+1, y).Y)

			laplacian := top + bottom + left + right - 4*center
			laplacianSum += laplacian
			n++
		}
	}

	mean := laplacianSum / n

	// Calculate variance
	var variance float64
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			center := float64(gray.GrayAt(x, y).Y)
			top := float64(gray.GrayAt(x, y-1).Y)
			bottom := float64(gray.GrayAt(x, y+1).Y)
			left := float64(gray.GrayAt(x-1, y).Y)
			right := float64(gray.GrayAt(x+1, y).Y)

			laplacian := top + bottom + left + right - 4*center
			diff := laplacian - mean
			variance += diff * diff
		}
	}

	return variance / n
}

// ComputeAllMetrics computes all quality metrics between a reference and test image
func ComputeAllMetrics(reference, test image.Image) (Metrics, error) {
	ssim, err := ComputeSSIM(reference, test)
	if err != nil {
		return Metrics{}, err
	}

	psnr, err := ComputePSNR(reference, test)
	if err != nil {
		return Metrics{}, err
	}

	mse := ComputeMSE(reference, test)
	sharpness := ComputeSharpness(test)

	return Metrics{
		SSIM:      ssim,
		PSNR:      psnr,
		Sharpness: sharpness,
		MSE:       mse,
	}, nil
}

// ComputeDeltaE computes the average Delta-E (CIE76) color difference between two images
// Returns the mean Delta-E across all pixels
func ComputeDeltaE(img1, img2 image.Image) float64 {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	minWidth := min(bounds1.Dx(), bounds2.Dx())
	minHeight := min(bounds1.Dy(), bounds2.Dy())

	var totalDeltaE float64
	var n int

	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			// Convert RGB to LAB (simplified)
			l1, a1, b_1 := rgbToLab(float64(r1), float64(g1), float64(b1))
			l2, a2, b_2 := rgbToLab(float64(r2), float64(g2), float64(b2))

			// Delta-E (CIE76 formula)
			deltaE := math.Sqrt((l2-l1)*(l2-l1) + (a2-a1)*(a2-a1) + (b_2-b_1)*(b_2-b_1))
			totalDeltaE += deltaE
			n++
		}
	}

	if n == 0 {
		return 0
	}

	return totalDeltaE / float64(n)
}

// rgbToLab converts RGB (0-255) to LAB color space (simplified)
// This is a simplified conversion for Delta-E calculation
func rgbToLab(r, g, b float64) (l, a, b_ float64) {
	// Normalize to 0-1
	r /= 255.0
	g /= 255.0
	b /= 255.0

	// Apply sRGB gamma correction
	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r /= 12.92
	}
	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g /= 12.92
	}
	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b /= 12.92
	}

	// RGB to XYZ (using D65 illuminant)
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041

	// XYZ to LAB (using D65 reference white)
	x /= 0.95047
	y /= 1.00000
	z /= 1.08883

	if x > 0.008856 {
		x = math.Pow(x, 1.0/3.0)
	} else {
		x = (7.787 * x) + (16.0 / 116.0)
	}
	if y > 0.008856 {
		y = math.Pow(y, 1.0/3.0)
	} else {
		y = (7.787 * y) + (16.0 / 116.0)
	}
	if z > 0.008856 {
		z = math.Pow(z, 1.0/3.0)
	} else {
		z = (7.787 * z) + (16.0 / 116.0)
	}

	l = (116.0 * y) - 16.0
	a = 500.0 * (x - y)
	b_ = 200.0 * (y - z)

	return l, a, b_
}

// CountClippedPixels counts pixels that are clipped (pure black or pure white)
func CountClippedPixels(img image.Image) (low, high int) {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit
			r, g, b = r>>8, g>>8, b>>8

			// Check for clipping
			if r == 0 && g == 0 && b == 0 {
				low++
			} else if r == 255 && g == 255 && b == 255 {
				high++
			}
		}
	}

	return low, high
}

// rgbaToGray converts a color to grayscale using luminance formula
func rgbaToGray(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit and apply luminance weights
	// Using Rec. 709 luma coefficients
	gray := 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)
	return gray
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
