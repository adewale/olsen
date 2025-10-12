package indexer

import (
	"testing"
	"time"

	"github.com/adewale/olsen/pkg/models"
)

func TestInferTimeOfDay(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		expected string
	}{
		{"Golden Hour Morning", 6, "golden_hour_morning"},
		{"Morning", 9, "morning"},
		{"Midday", 13, "midday"},
		{"Afternoon", 16, "afternoon"},
		{"Golden Hour Evening", 19, "golden_hour_evening"},
		{"Blue Hour", 21, "blue_hour"},
		{"Night", 23, "night"},
		{"Night Early", 3, "night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := time.Date(2025, 10, 4, tt.hour, 0, 0, 0, time.UTC)
			result := inferTimeOfDay(date)
			if result != tt.expected {
				t.Errorf("inferTimeOfDay(%d) = %s; want %s", tt.hour, result, tt.expected)
			}
		})
	}
}

func TestInferTimeOfDayZero(t *testing.T) {
	result := inferTimeOfDay(time.Time{})
	if result != "" {
		t.Errorf("inferTimeOfDay(zero) = %s; want empty string", result)
	}
}

func TestInferSeason(t *testing.T) {
	tests := []struct {
		name     string
		month    time.Month
		expected string
	}{
		{"Spring March", time.March, "spring"},
		{"Spring April", time.April, "spring"},
		{"Spring May", time.May, "spring"},
		{"Summer June", time.June, "summer"},
		{"Summer July", time.July, "summer"},
		{"Summer August", time.August, "summer"},
		{"Autumn September", time.September, "autumn"},
		{"Autumn October", time.October, "autumn"},
		{"Autumn November", time.November, "autumn"},
		{"Winter December", time.December, "winter"},
		{"Winter January", time.January, "winter"},
		{"Winter February", time.February, "winter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := time.Date(2025, tt.month, 15, 12, 0, 0, 0, time.UTC)
			result := inferSeason(date)
			if result != tt.expected {
				t.Errorf("inferSeason(%s) = %s; want %s", tt.month, result, tt.expected)
			}
		})
	}
}

func TestInferSeasonZero(t *testing.T) {
	result := inferSeason(time.Time{})
	if result != "" {
		t.Errorf("inferSeason(zero) = %s; want empty string", result)
	}
}

func TestInferFocalCategory(t *testing.T) {
	tests := []struct {
		name      string
		focal35mm int
		expected  string
	}{
		{"Wide 24mm", 24, "wide"},
		{"Wide 28mm", 28, "wide"},
		{"Normal 35mm", 35, "normal"},
		{"Normal 50mm", 50, "normal"},
		{"Normal 70mm", 70, "normal"},
		{"Telephoto 85mm", 85, "telephoto"},
		{"Telephoto 135mm", 135, "telephoto"},
		{"Telephoto 200mm", 200, "telephoto"},
		{"Super Telephoto 300mm", 300, "super_telephoto"},
		{"Super Telephoto 600mm", 600, "super_telephoto"},
		{"Zero", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferFocalCategory(tt.focal35mm)
			if result != tt.expected {
				t.Errorf("inferFocalCategory(%d) = %s; want %s", tt.focal35mm, result, tt.expected)
			}
		})
	}
}

func TestInferShootingCondition(t *testing.T) {
	tests := []struct {
		name       string
		iso        int
		flashFired bool
		expected   string
	}{
		{"Bright ISO 100", 100, false, "bright"},
		{"Bright ISO 400", 400, false, "bright"},
		{"Moderate ISO 800", 800, false, "moderate"},
		{"Moderate ISO 1200", 1200, false, "moderate"},
		{"Low Light ISO 1600", 1600, false, "low_light"},
		{"Low Light ISO 3200", 3200, false, "low_light"},
		{"Flash Low ISO", 100, true, "flash"},
		{"Flash High ISO", 3200, true, "flash"},
		{"Zero ISO", 0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferShootingCondition(tt.iso, tt.flashFired)
			if result != tt.expected {
				t.Errorf("inferShootingCondition(%d, %v) = %s; want %s",
					tt.iso, tt.flashFired, result, tt.expected)
			}
		})
	}
}

func TestInferMetadata(t *testing.T) {
	metadata := &models.PhotoMetadata{
		DateTaken:       time.Date(2025, 10, 4, 16, 30, 0, 0, time.UTC), // 16:30 is afternoon
		FocalLength35mm: 85,
		ISO:             800,
		FlashFired:      false,
	}

	InferMetadata(metadata)

	if metadata.TimeOfDay != "afternoon" {
		t.Errorf("TimeOfDay = %s; want afternoon", metadata.TimeOfDay)
	}
	if metadata.Season != "autumn" {
		t.Errorf("Season = %s; want autumn", metadata.Season)
	}
	if metadata.FocalCategory != "telephoto" {
		t.Errorf("FocalCategory = %s; want telephoto", metadata.FocalCategory)
	}
	if metadata.ShootingCondition != "moderate" {
		t.Errorf("ShootingCondition = %s; want moderate", metadata.ShootingCondition)
	}
}
