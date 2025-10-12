package indexer

import (
	"testing"
	"time"
)

func TestParseExifDateTime(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantYear   int
		wantMonth  time.Month
		wantDay    int
		wantHour   int
		wantMinute int
		wantSecond int
	}{
		{
			name:       "Standard EXIF format",
			input:      "2025:01:15 14:30:45",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "EXIF with milliseconds",
			input:      "2025:01:15 14:30:45.123",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "ISO 8601 with space",
			input:      "2025-01-15 14:30:45",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "ISO 8601 with T",
			input:      "2025-01-15T14:30:45",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "ISO 8601 with Z",
			input:      "2025-01-15T14:30:45Z",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:      "Date only (colon format)",
			input:     "2025:01:15",
			wantErr:   false,
			wantYear:  2025,
			wantMonth: time.January,
			wantDay:   15,
			wantHour:  0,
		},
		{
			name:      "Date only (dash format)",
			input:     "2025-01-15",
			wantErr:   false,
			wantYear:  2025,
			wantMonth: time.January,
			wantDay:   15,
			wantHour:  0,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "Null bytes and whitespace",
			input:   "\x00  \x00",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			input:   "not a date",
			wantErr: true,
		},
		{
			name:    "Wrong date format",
			input:   "15/01/2025",
			wantErr: true,
		},
		{
			name:       "Date with leading/trailing whitespace",
			input:      "  2025:01:15 14:30:45  ",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "Date with null bytes",
			input:      "\x002025:01:15 14:30:45\x00",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   14,
			wantMinute: 30,
			wantSecond: 45,
		},
		{
			name:       "Midnight time",
			input:      "2025:01:15 00:00:00",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
		},
		{
			name:       "End of day",
			input:      "2025:01:15 23:59:59",
			wantErr:    false,
			wantYear:   2025,
			wantMonth:  time.January,
			wantDay:    15,
			wantHour:   23,
			wantMinute: 59,
			wantSecond: 59,
		},
		{
			name:       "Leap year date",
			input:      "2024:02:29 12:00:00",
			wantErr:    false,
			wantYear:   2024,
			wantMonth:  time.February,
			wantDay:    29,
			wantHour:   12,
			wantMinute: 0,
			wantSecond: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExifDateTime(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseExifDateTime(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseExifDateTime(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got.Year() != tt.wantYear {
				t.Errorf("Year = %d, want %d", got.Year(), tt.wantYear)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("Month = %v, want %v", got.Month(), tt.wantMonth)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("Day = %d, want %d", got.Day(), tt.wantDay)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d", got.Hour(), tt.wantHour)
			}
			if tt.wantMinute != 0 && got.Minute() != tt.wantMinute {
				t.Errorf("Minute = %d, want %d", got.Minute(), tt.wantMinute)
			}
			if tt.wantSecond != 0 && got.Second() != tt.wantSecond {
				t.Errorf("Second = %d, want %d", got.Second(), tt.wantSecond)
			}
		})
	}
}

func TestParseExifDateTime_EdgeCases(t *testing.T) {
	t.Run("All formats parse consistently", func(t *testing.T) {
		// All these should parse to the same time
		formats := []string{
			"2025:01:15 14:30:45",
			"2025-01-15 14:30:45",
			"2025-01-15T14:30:45",
			"2025-01-15T14:30:45Z",
		}

		var times []time.Time
		for _, format := range formats {
			parsed, err := parseExifDateTime(format)
			if err != nil {
				t.Errorf("Failed to parse %q: %v", format, err)
				continue
			}
			times = append(times, parsed)
		}

		// Check all times are equivalent (ignoring timezone)
		if len(times) > 1 {
			first := times[0]
			for i, ti := range times[1:] {
				if first.Year() != ti.Year() || first.Month() != ti.Month() ||
					first.Day() != ti.Day() || first.Hour() != ti.Hour() ||
					first.Minute() != ti.Minute() || first.Second() != ti.Second() {
					t.Errorf("Time mismatch between format 0 and %d: %v vs %v",
						i+1, first, ti)
				}
			}
		}
	})

	t.Run("Date-only formats have zero time", func(t *testing.T) {
		formats := []string{
			"2025:01:15",
			"2025-01-15",
		}

		for _, format := range formats {
			parsed, err := parseExifDateTime(format)
			if err != nil {
				t.Errorf("Failed to parse %q: %v", format, err)
				continue
			}

			if parsed.Hour() != 0 || parsed.Minute() != 0 || parsed.Second() != 0 {
				t.Errorf("Date-only format %q should have zero time, got %02d:%02d:%02d",
					format, parsed.Hour(), parsed.Minute(), parsed.Second())
			}
		}
	})
}
