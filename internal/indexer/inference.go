package indexer

import (
	"time"

	"github.com/adewale/olsen/pkg/models"
)

// InferMetadata adds inferred metadata based on extracted EXIF data
func InferMetadata(metadata *models.PhotoMetadata) {
	metadata.TimeOfDay = inferTimeOfDay(metadata.DateTaken)
	metadata.Season = inferSeason(metadata.DateTaken)
	metadata.FocalCategory = inferFocalCategory(metadata.FocalLength35mm)
	metadata.ShootingCondition = inferShootingCondition(metadata.ISO, metadata.FlashFired)
}

// inferTimeOfDay classifies the time of day based on the hour of capture
func inferTimeOfDay(dateTaken time.Time) string {
	if dateTaken.IsZero() {
		return ""
	}

	hour := dateTaken.Hour()

	switch {
	case hour >= 5 && hour < 7:
		return "golden_hour_morning"
	case hour >= 7 && hour < 11:
		return "morning"
	case hour >= 11 && hour < 15:
		return "midday"
	case hour >= 15 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 20:
		return "golden_hour_evening"
	case hour >= 20 && hour < 22:
		return "blue_hour"
	default:
		return "night"
	}
}

// inferSeason classifies the season based on the month (Northern Hemisphere)
func inferSeason(dateTaken time.Time) string {
	if dateTaken.IsZero() {
		return ""
	}

	month := dateTaken.Month()

	switch month {
	case time.March, time.April, time.May:
		return "spring"
	case time.June, time.July, time.August:
		return "summer"
	case time.September, time.October, time.November:
		return "autumn"
	case time.December, time.January, time.February:
		return "winter"
	default:
		return ""
	}
}

// inferFocalCategory classifies focal length into categories
func inferFocalCategory(focalLength35mm int) string {
	if focalLength35mm == 0 {
		return ""
	}

	switch {
	case focalLength35mm < 35:
		return "wide"
	case focalLength35mm >= 35 && focalLength35mm <= 70:
		return "normal"
	case focalLength35mm >= 71 && focalLength35mm <= 200:
		return "telephoto"
	case focalLength35mm > 200:
		return "super_telephoto"
	default:
		return ""
	}
}

// inferShootingCondition classifies shooting conditions based on ISO and flash
func inferShootingCondition(iso int, flashFired bool) string {
	if iso == 0 {
		return ""
	}

	if flashFired {
		return "flash"
	}

	switch {
	case iso <= 400:
		return "bright"
	case iso >= 401 && iso <= 1599:
		return "moderate"
	case iso >= 1600:
		return "low_light"
	default:
		return ""
	}
}
