package query

import "strings"

// colourClassCaseSQL is the single source of truth for classifying a
// photo_colors row (aliased pc) into one of the Berlin-Kay colour names.
//
// Both the colour facet (GROUP BY this expression) and the colour filter
// (EXISTS ... WHERE this expression = ?) must use it, so that the count a
// facet displays and the rows a click returns can never disagree. The facet
// and the filter previously had independent classifications ("gray" vs
// "grey", different black/white thresholds), which made facet counts lie.
//
// Classification priority: saturation first (achromatic colours), then
// brown (orange hue + low lightness), then hue ranges.
const colourClassCaseSQL = `CASE
	WHEN pc.saturation < 5 AND pc.lightness < 20 THEN 'black'
	WHEN pc.saturation < 5 AND pc.lightness > 80 THEN 'white'
	WHEN pc.saturation < 10 THEN 'gray'
	WHEN pc.saturation < 15 THEN 'bw'
	WHEN pc.hue BETWEEN 20 AND 40 AND pc.lightness < 50 THEN 'brown'
	WHEN pc.hue BETWEEN 0 AND 15 OR pc.hue BETWEEN 345 AND 360 THEN 'red'
	WHEN pc.hue BETWEEN 16 AND 45 THEN 'orange'
	WHEN pc.hue BETWEEN 46 AND 75 THEN 'yellow'
	WHEN pc.hue BETWEEN 76 AND 165 THEN 'green'
	WHEN pc.hue BETWEEN 166 AND 255 THEN 'blue'
	WHEN pc.hue BETWEEN 256 AND 290 THEN 'purple'
	WHEN pc.hue BETWEEN 291 AND 344 THEN 'pink'
	ELSE 'other'
END`

// canonicalColourName maps user-supplied colour names onto the vocabulary
// colourClassCaseSQL produces, accepting common alternate spellings.
func canonicalColourName(name string) string {
	switch n := strings.ToLower(strings.TrimSpace(name)); n {
	case "grey":
		return "gray"
	case "b&w", "bnw", "monochrome":
		return "bw"
	default:
		return n
	}
}
