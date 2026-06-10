package quality

import (
	"image"
	"image/color"
	"testing"

	"github.com/nfnt/resize"
	xdraw "golang.org/x/image/draw"
)

// Differential test: nfnt/resize (archived since 2018, the current resampler)
// against golang.org/x/image/draw (maintained). Each implementation acts as
// the other's oracle. This pins how close the two are today and is the
// safety harness for an eventual migration off the archived library: if the
// two ever drift apart — from a behavior change in either — this fails
// before users see different thumbnails.

// makeStructuredImage produces an image with edges, texture, and gradients so
// SSIM comparisons are meaningful (flat images score ~1.0 trivially).
func makeStructuredImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Gradient base
			r := uint8(x * 255 / w)
			g := uint8(y * 255 / h)
			// Checkerboard texture for high-frequency content
			b := uint8(40)
			if (x/8+y/8)%2 == 0 {
				b = 215
			}
			// A few hard diagonal edges
			if (x+y)%97 < 3 {
				r, g, b = 255, 255, 255
			}
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func resizeWithXDraw(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func meanAbsLumaDiff(a, b image.Image) float64 {
	bounds := a.Bounds()
	var total float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			la := rgbaToGray(a.At(x, y))
			lb := rgbaToGray(b.At(x, y))
			d := la - lb
			if d < 0 {
				d = -d
			}
			total += d
		}
	}
	return total / float64(bounds.Dx()*bounds.Dy())
}

func TestResizeLibrariesAgree(t *testing.T) {
	t.Parallel()

	src := makeStructuredImage(1024, 768)

	cases := []struct{ w, h int }{
		{512, 384},
		{256, 192},
		{64, 48},
	}

	for _, tc := range cases {
		oldImg := resize.Resize(uint(tc.w), uint(tc.h), src, resize.Lanczos3)
		newImg := resizeWithXDraw(src, tc.w, tc.h)

		if oldImg.Bounds().Dx() != newImg.Bounds().Dx() || oldImg.Bounds().Dy() != newImg.Bounds().Dy() {
			t.Fatalf("%dx%d: dimension mismatch: nfnt=%v xdraw=%v", tc.w, tc.h, oldImg.Bounds(), newImg.Bounds())
		}

		ssim, err := ComputeSSIM(oldImg, newImg)
		if err != nil {
			t.Fatalf("%dx%d: ComputeSSIM failed: %v", tc.w, tc.h, err)
		}
		lumaDiff := meanAbsLumaDiff(oldImg, newImg)
		t.Logf("%dx%d: SSIM=%.4f meanLumaDiff=%.2f", tc.w, tc.h, ssim, lumaDiff)

		// Lanczos3 and CatmullRom are different kernels, so pixel-perfect
		// agreement is not expected; structural similarity must stay high
		// and average luminance must stay close.
		if ssim < 0.90 {
			t.Errorf("%dx%d: SSIM %.4f below 0.90 — resamplers have drifted apart", tc.w, tc.h, ssim)
		}
		if lumaDiff > 5.0 {
			t.Errorf("%dx%d: mean |luma| diff %.2f above 5.0 — resamplers have drifted apart", tc.w, tc.h, lumaDiff)
		}
	}
}
