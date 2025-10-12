package query

import (
	"testing"
)

// TestColorClassification verifies the SQL CASE statement logic for color classification
// This tests the classification rules for the 11 Berlin-Kay universal basic colors

func TestColorClassification_Black(t *testing.T) {
	// Pure black: low saturation, very low lightness
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		{0, 0, 0, "black"},    // Pure black
		{0, 2, 10, "black"},   // Very dark gray
		{180, 4, 15, "black"}, // Dark gray (any hue)
		{0, 5, 20, "gray"},    // Not quite black (S=5 is threshold)
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_White(t *testing.T) {
	// Pure white: low saturation, very high lightness
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		{0, 0, 100, "white"},  // Pure white
		{0, 2, 90, "white"},   // Very light gray
		{180, 4, 85, "white"}, // Light gray (any hue)
		{0, 5, 80, "gray"},    // Not quite white (S=5 is threshold)
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_Gray(t *testing.T) {
	// Gray: low saturation, medium lightness
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		{0, 5, 50, "gray"},   // Mid gray
		{120, 8, 40, "gray"}, // Dark gray
		{240, 9, 70, "gray"}, // Light gray
		{0, 10, 50, "bw"},    // Threshold boundary
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_BW(t *testing.T) {
	// B&W: slightly desaturated, near-grayscale
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		{0, 10, 50, "bw"},      // Near-grayscale
		{30, 12, 40, "bw"},     // Slight sepia tone
		{200, 14, 60, "bw"},    // Very desaturated blue
		{0, 15, 50, "red"},     // At S=15 threshold, hue=0 is red
		{30, 15, 40, "brown"},  // At S=15 threshold, brown hue range + low lightness
		{100, 15, 50, "green"}, // At S=15 threshold, green hue range
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_Brown(t *testing.T) {
	// Brown: orange hue (20-40°) with low lightness (<50%)
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		{25, 50, 30, "brown"},  // Dark orange = brown
		{35, 60, 40, "brown"},  // Brown
		{30, 20, 45, "brown"},  // Desaturated brown
		{25, 50, 50, "orange"}, // Not brown (lightness threshold)
		{15, 50, 30, "red"},    // Not brown (hue too low)
		{45, 50, 30, "orange"}, // Not brown (hue too high)
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_ChromaticColors(t *testing.T) {
	// Test all chromatic color ranges
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		// Red
		{0, 70, 50, "red"},
		{10, 80, 60, "red"},
		{350, 75, 55, "red"},

		// Orange
		{20, 70, 55, "orange"}, // Light orange (not brown)
		{40, 80, 60, "orange"},

		// Yellow
		{50, 90, 60, "yellow"},
		{70, 85, 65, "yellow"},

		// Green
		{90, 60, 50, "green"},
		{120, 70, 55, "green"},
		{160, 65, 45, "green"},

		// Blue
		{180, 70, 50, "blue"},
		{210, 80, 55, "blue"},
		{240, 75, 60, "blue"},

		// Purple
		{270, 70, 50, "purple"},
		{285, 75, 55, "purple"},

		// Pink
		{310, 70, 70, "pink"},
		{330, 75, 65, "pink"},
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

func TestColorClassification_EdgeCases(t *testing.T) {
	// Test boundary conditions
	testCases := []struct {
		hue, sat, light int
		expected        string
	}{
		// Saturation boundaries
		{0, 4, 50, "gray"},     // Just below S=5
		{0, 5, 50, "gray"},     // At S=5
		{0, 9, 50, "gray"},     // Just below S=10
		{0, 10, 50, "bw"},      // At S=10
		{0, 14, 50, "bw"},      // Just below S=15
		{100, 15, 50, "green"}, // At S=15

		// Lightness boundaries for black/white
		{0, 4, 19, "black"}, // Just below L=20
		{0, 4, 20, "gray"},  // At L=20
		{0, 4, 80, "gray"},  // At L=80
		{0, 4, 81, "white"}, // Just above L=80

		// Lightness boundary for brown
		{30, 50, 49, "brown"},  // Just below L=50
		{30, 50, 50, "orange"}, // At L=50

		// Hue boundaries
		{15, 50, 50, "red"},    // Red upper bound
		{16, 50, 50, "orange"}, // Orange lower bound
		{45, 50, 50, "orange"}, // Orange upper bound
		{46, 50, 50, "yellow"}, // Yellow lower bound
	}

	for _, tc := range testCases {
		result := classifyColor(tc.hue, tc.sat, tc.light)
		if result != tc.expected {
			t.Errorf("HSL(%d, %d, %d) = %s, expected %s", tc.hue, tc.sat, tc.light, result, tc.expected)
		}
	}
}

// TestColorClassification_BrownVsOrangeConfusion tests the specific problem where
// brown was being confused for orange before lightness check was added
func TestColorClassification_BrownVsOrangeConfusion(t *testing.T) {
	testCases := []struct {
		name            string
		hue, sat, light int
		expected        string
		reason          string
	}{
		{
			name: "Dark orange = brown",
			hue:  30, sat: 60, light: 30,
			expected: "brown",
			reason:   "Orange hue (30°) with low lightness (30%) should be brown, not orange",
		},
		{
			name: "Chocolate brown",
			hue:  25, sat: 50, light: 25,
			expected: "brown",
			reason:   "Classic brown: orange-ish hue with very low lightness",
		},
		{
			name: "Light orange is NOT brown",
			hue:  30, sat: 60, light: 60,
			expected: "orange",
			reason:   "Same hue but high lightness (60%) should be orange, not brown",
		},
		{
			name: "Exactly at lightness threshold",
			hue:  30, sat: 50, light: 49,
			expected: "brown",
			reason:   "L=49 is below 50% threshold, should be brown",
		},
		{
			name: "Just above lightness threshold",
			hue:  30, sat: 50, light: 50,
			expected: "orange",
			reason:   "L=50 is at/above threshold, should be orange",
		},
		{
			name: "Desaturated brown",
			hue:  35, sat: 20, light: 40,
			expected: "brown",
			reason:   "Low saturation is fine for brown if hue and lightness match",
		},
		{
			name: "Wood/earth tones",
			hue:  28, sat: 45, light: 35,
			expected: "brown",
			reason:   "Natural brown colors in photography (wood, dirt, leather)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyColor(tc.hue, tc.sat, tc.light)
			if result != tc.expected {
				t.Errorf("HSL(%d, %d, %d) = %s, expected %s\nReason: %s",
					tc.hue, tc.sat, tc.light, result, tc.expected, tc.reason)
			}
		})
	}
}

// TestColorClassification_BWMisclassifiedAsRed tests the specific problem where
// B&W photos were being classified as red due to hue=0° being checked before saturation
func TestColorClassification_BWMisclassifiedAsRed(t *testing.T) {
	testCases := []struct {
		name            string
		hue, sat, light int
		expected        string
		reason          string
	}{
		{
			name: "Pure grayscale - was red, now gray",
			hue:  0, sat: 0, light: 50,
			expected: "gray",
			reason:   "Before fix: hue=0° → red. After fix: sat=0% → gray (correct!)",
		},
		{
			name: "B&W portrait with slight noise",
			hue:  0, sat: 3, light: 40,
			expected: "gray",
			reason:   "Film scan with minimal color noise should be gray, not red",
		},
		{
			name: "Dark grayscale - was red, now black",
			hue:  0, sat: 2, light: 15,
			expected: "black",
			reason:   "Very dark gray should be black, not red",
		},
		{
			name: "Light grayscale - was red, now white",
			hue:  0, sat: 3, light: 85,
			expected: "white",
			reason:   "Very light gray should be white, not red",
		},
		{
			name: "Near-grayscale with sepia tone",
			hue:  30, sat: 8, light: 45,
			expected: "gray",
			reason:   "Slight warm tone but S<10% means it's still grayscale",
		},
		{
			name: "Converted B&W with color cast",
			hue:  210, sat: 5, light: 60,
			expected: "gray",
			reason:   "Blue cast (hue=210°) but S<10% means grayscale, not blue",
		},
		{
			name: "True red (not B&W)",
			hue:  0, sat: 70, light: 50,
			expected: "red",
			reason:   "High saturation means this is actually red, not B&W",
		},
		{
			name: "Desaturated red photo",
			hue:  5, sat: 12, light: 55,
			expected: "bw",
			reason:   "S=12% is in B&W range (10-15%), correctly classified as near-grayscale",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyColor(tc.hue, tc.sat, tc.light)
			if result != tc.expected {
				t.Errorf("HSL(%d, %d, %d) = %s, expected %s\nReason: %s",
					tc.hue, tc.sat, tc.light, result, tc.expected, tc.reason)
			}
		})
	}
}

// TestColorClassification_RealWorldExamples tests colors extracted from actual photos
// to ensure classification matches human perception
func TestColorClassification_RealWorldExamples(t *testing.T) {
	testCases := []struct {
		name            string
		hue, sat, light int
		expected        string
		scenario        string
	}{
		{
			name: "Black and white film scan",
			hue:  0, sat: 0, light: 48,
			expected: "gray",
			scenario: "Scanned B&W negative - should never be classified as colored",
		},
		{
			name: "Leather jacket in portrait",
			hue:  28, sat: 35, light: 25,
			expected: "brown",
			scenario: "Dark brown leather - common in fashion photography",
		},
		{
			name: "Sunset orange sky",
			hue:  25, sat: 80, light: 60,
			expected: "orange",
			scenario: "Bright sunset - clearly orange, not brown",
		},
		{
			name: "Wooden furniture",
			hue:  32, sat: 42, light: 38,
			expected: "brown",
			scenario: "Natural wood tones in interior photography",
		},
		{
			name: "Concrete building",
			hue:  0, sat: 5, light: 65,
			expected: "gray",
			scenario: "Urban architecture - achromatic surfaces",
		},
		{
			name: "Overcast sky",
			hue:  200, sat: 8, light: 70,
			expected: "gray",
			scenario: "Cloudy day - desaturated blue-ish gray",
		},
		{
			name: "Dirt/soil in landscape",
			hue:  30, sat: 40, light: 30,
			expected: "brown",
			scenario: "Earth tones in nature photography",
		},
		{
			name: "Pure black shadow",
			hue:  0, sat: 0, light: 5,
			expected: "black",
			scenario: "Deep shadows in high-contrast photography",
		},
		{
			name: "Pure white highlight",
			hue:  0, sat: 2, light: 95,
			expected: "white",
			scenario: "Blown highlights or white backgrounds",
		},
		{
			name: "Sepia-toned photo",
			hue:  35, sat: 15, light: 50,
			expected: "orange",
			scenario: "Vintage sepia effect - S=15% is threshold, becomes orange not brown due to L≥50",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyColor(tc.hue, tc.sat, tc.light)
			if result != tc.expected {
				t.Errorf("HSL(%d, %d, %d) = %s, expected %s\nScenario: %s",
					tc.hue, tc.sat, tc.light, result, tc.expected, tc.scenario)
			}
		})
	}
}

// classifyColor implements the same logic as the SQL CASE statement
// This allows us to test the classification rules in Go
func classifyColor(hue, saturation, lightness int) string {
	// Achromatic colors (check saturation first)
	if saturation < 5 && lightness < 20 {
		return "black"
	}
	if saturation < 5 && lightness > 80 {
		return "white"
	}
	if saturation < 10 {
		return "gray"
	}
	if saturation < 15 {
		return "bw"
	}

	// Chromatic colors (check hue ranges)
	// Brown: orange hue + low lightness
	if hue >= 20 && hue <= 40 && lightness < 50 {
		return "brown"
	}
	if (hue >= 0 && hue <= 15) || (hue >= 345 && hue <= 360) {
		return "red"
	}
	if hue >= 16 && hue <= 45 {
		return "orange"
	}
	if hue >= 46 && hue <= 75 {
		return "yellow"
	}
	if hue >= 76 && hue <= 165 {
		return "green"
	}
	if hue >= 166 && hue <= 255 {
		return "blue"
	}
	if hue >= 256 && hue <= 290 {
		return "purple"
	}
	if hue >= 291 && hue <= 344 {
		return "pink"
	}

	return "other"
}
