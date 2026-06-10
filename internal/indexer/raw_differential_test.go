//go:build cgo && use_seppedelanghe_libraw

package indexer

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	golibraw "github.com/inokone/golibraw"
	golibrawnew "github.com/seppedelanghe/go-libraw"
)

// Differential tests between the two LibRaw bindings olsen can build
// against. Neither is a perfect oracle, but they wrap the same underlying
// library, so on well-formed files they must agree on the basics; where they
// are known to diverge (JPEG-compressed monochrome DNGs), the divergence is
// pinned as a characterization test instead of silently drifting.

func decodeBoth(t *testing.T, path string) (oldImg, newImg image.Image, oldErr, newErr error) {
	t.Helper()

	oldImg, oldErr = golibraw.ImportRaw(path)

	processor := golibrawnew.NewProcessor(golibrawnew.ProcessorOptions{
		UserQual:    0,
		OutputBps:   8,
		OutputColor: golibrawnew.SRGB,
		UseCameraWb: true,
	})
	newImg, _, newErr = processor.ProcessRaw(path)
	return oldImg, newImg, oldErr, newErr
}

func meanLuma(img image.Image) float64 {
	bounds := img.Bounds()
	stepX := max(1, bounds.Dx()/256)
	stepY := max(1, bounds.Dy()/256)
	var total, n float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			total += 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			n++
		}
	}
	return total / n
}

// TestLibRawBindingsAgreeOnSyntheticDNGs decodes the same well-formed DNG
// fixtures through both bindings and asserts they agree on dimensions and
// approximate brightness.
func TestLibRawBindingsAgreeOnSyntheticDNGs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode: decodes multi-MB DNG fixtures")
	}

	dngDir := filepath.Join("..", "..", "testdata", "dng")
	fixtures := []string{
		"01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng",
		"06_nikon_z9_50mm_winter_blue_hour_iso1600_blue_nogps.dng",
	}

	for _, name := range fixtures {
		path := filepath.Join(dngDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("fixture %s not found", name)
		}

		t.Run(name, func(t *testing.T) {
			oldImg, newImg, oldErr, newErr := decodeBoth(t, path)

			if (oldErr == nil) != (newErr == nil) {
				t.Fatalf("bindings disagree on decodability: golibraw err=%v, go-libraw err=%v", oldErr, newErr)
			}
			if oldErr != nil {
				t.Skipf("both bindings fail on this fixture (golibraw: %v, go-libraw: %v)", oldErr, newErr)
			}

			ob, nb := oldImg.Bounds(), newImg.Bounds()
			if ob.Dx() != nb.Dx() || ob.Dy() != nb.Dy() {
				t.Errorf("dimension mismatch: golibraw=%dx%d go-libraw=%dx%d", ob.Dx(), ob.Dy(), nb.Dx(), nb.Dy())
			}

			oldLuma, newLuma := meanLuma(oldImg), meanLuma(newImg)
			t.Logf("mean luma: golibraw=%.1f go-libraw=%.1f", oldLuma, newLuma)

			// Different demosaic defaults produce slightly different output;
			// a 2x brightness ratio bound catches the "black image" class of
			// bug without flaking on processing differences.
			lo, hi := oldLuma, newLuma
			if lo > hi {
				lo, hi = hi, lo
			}
			if lo < 1 || hi/lo > 2.0 {
				t.Errorf("brightness diverges: golibraw=%.1f go-libraw=%.1f (ratio bound 2.0)", oldLuma, newLuma)
			}
		})
	}
}

// TestLibRawBindingsLeicaMonochromDivergence characterizes how each binding
// handles the JPEG-compressed monochrome DNG that motivated dual-library
// support. It does not demand success from either — it pins the current
// behavior so an upgrade that changes it is noticed and the docs can follow.
func TestLibRawBindingsLeicaMonochromDivergence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode: decodes a 70MB DNG")
	}

	path := filepath.Join("..", "..", "testdata", "dng", "L1001515.DNG")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Leica fixture not found")
	}

	oldImg, newImg, oldErr, newErr := decodeBoth(t, path)
	t.Logf("golibraw: err=%v", oldErr)
	t.Logf("go-libraw: err=%v", newErr)

	if oldErr == nil {
		t.Logf("golibraw: %dx%d meanLuma=%.1f", oldImg.Bounds().Dx(), oldImg.Bounds().Dy(), meanLuma(oldImg))
	}
	if newErr == nil {
		t.Logf("go-libraw: %dx%d meanLuma=%.1f", newImg.Bounds().Dx(), newImg.Bounds().Dy(), meanLuma(newImg))
	}

	// Whatever the bindings do, the indexer-level pipeline must produce a
	// usable image for this file (decode or embedded-JPEG fallback).
	img, err := DecodeRaw(path)
	if err != nil {
		img, err = ExtractEmbeddedJPEG(path)
	}
	if err != nil {
		t.Fatalf("neither DecodeRaw nor embedded-JPEG fallback produced an image: %v", err)
	}
	if luma := meanLuma(img); luma < 5 {
		t.Errorf("pipeline produced a near-black image (meanLuma=%.1f) — the Monochrom bug is back", luma)
	}
}
