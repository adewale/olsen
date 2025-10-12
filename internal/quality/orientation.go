package quality

import (
	"image"
)

// ApplyOrientation applies EXIF orientation transformation to an image
// orientation values follow EXIF standard (1-8)
// Returns the oriented image and true if orientation was applied
func ApplyOrientation(img image.Image, orientation int) (image.Image, bool) {
	if orientation < 2 || orientation > 8 {
		// Orientation 1 or invalid = no transform needed
		return img, false
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var result *image.NRGBA

	switch orientation {
	case 2:
		// Flip horizontal
		result = image.NewNRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(width-1-x, y, img.At(x, y))
			}
		}

	case 3:
		// Rotate 180°
		result = image.NewNRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(width-1-x, height-1-y, img.At(x, y))
			}
		}

	case 4:
		// Flip vertical
		result = image.NewNRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(x, height-1-y, img.At(x, y))
			}
		}

	case 5:
		// Rotate 90° CW and flip horizontal
		result = image.NewNRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(y, width-1-x, img.At(x, y))
			}
		}

	case 6:
		// Rotate 90° CW
		result = image.NewNRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(height-1-y, x, img.At(x, y))
			}
		}

	case 7:
		// Rotate 90° CCW and flip horizontal
		result = image.NewNRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(height-1-y, width-1-x, img.At(x, y))
			}
		}

	case 8:
		// Rotate 90° CCW (or 270° CW)
		result = image.NewNRGBA(image.Rect(0, 0, height, width))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				result.Set(y, x, img.At(x, y))
			}
		}

	default:
		return img, false
	}

	return result, true
}

// OrientationString returns a human-readable description of the orientation
func OrientationString(orientation int) string {
	switch orientation {
	case 1:
		return "Normal (no transform)"
	case 2:
		return "Flip horizontal"
	case 3:
		return "Rotate 180°"
	case 4:
		return "Flip vertical"
	case 5:
		return "Rotate 90° CW + flip horizontal"
	case 6:
		return "Rotate 90° CW"
	case 7:
		return "Rotate 90° CCW + flip horizontal"
	case 8:
		return "Rotate 90° CCW"
	default:
		return "Unknown"
	}
}

// ValidateOrientationAppliedOnce checks that orientation hasn't been applied twice
// This is a helper for guardrails - we track orientation state through the pipeline
type OrientationTracker struct {
	applied bool
	value   int
}

// NewOrientationTracker creates a new orientation tracker
func NewOrientationTracker() *OrientationTracker {
	return &OrientationTracker{
		applied: false,
		value:   1,
	}
}

// Apply marks orientation as applied and records the value
// Returns error if already applied
func (ot *OrientationTracker) Apply(orientation int) error {
	if ot.applied {
		return &OrientationError{
			Message: "orientation already applied once",
			Applied: ot.value,
			Attempt: orientation,
		}
	}
	ot.applied = true
	ot.value = orientation
	return nil
}

// IsApplied returns true if orientation has been applied
func (ot *OrientationTracker) IsApplied() bool {
	return ot.applied
}

// Value returns the orientation value that was applied
func (ot *OrientationTracker) Value() int {
	return ot.value
}

// OrientationError is returned when orientation is applied incorrectly
type OrientationError struct {
	Message string
	Applied int // The orientation value already applied
	Attempt int // The orientation value attempting to be applied
}

func (e *OrientationError) Error() string {
	return e.Message
}
