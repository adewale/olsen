package indexer

import (
	"fmt"
	"os"
	"strings"
	"time"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"

	"github.com/adewale/olsen/pkg/models"
)

// ExtractMetadata extracts EXIF metadata using the exif-go library
func ExtractMetadata(filePath string) (*models.PhotoMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file data
	data := make([]byte, fileInfo.Size())
	if _, err := file.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse EXIF
	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract EXIF: %w", err)
	}

	// Parse IFD structure
	entries, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EXIF: %w", err)
	}

	metadata := &models.PhotoMetadata{
		FilePath:     filePath,
		FileSize:     fileInfo.Size(),
		LastModified: fileInfo.ModTime(),
		IndexedAt:    time.Now(),
	}

	// Process all EXIF tags
	for _, entry := range entries {
		tagName := entry.TagName
		val := entry.Value

		if val == nil {
			continue
		}

		switch tagName {
		// Camera metadata
		case "Make":
			metadata.CameraMake = strings.Trim(fmt.Sprintf("%v", val), "\x00 ")
		case "Model":
			metadata.CameraModel = strings.Trim(fmt.Sprintf("%v", val), "\x00 ")
		case "LensMake":
			metadata.LensMake = strings.Trim(fmt.Sprintf("%v", val), "\x00 ")
		case "LensModel":
			metadata.LensModel = strings.Trim(fmt.Sprintf("%v", val), "\x00 ")

		// Exposure metadata
		case "ISOSpeedRatings", "PhotographicSensitivity":
			if iso, ok := val.([]uint16); ok && len(iso) > 0 {
				metadata.ISO = int(iso[0])
			}
		case "FNumber":
			if rats, ok := val.([]exifcommon.Rational); ok && len(rats) > 0 {
				metadata.Aperture = float64(rats[0].Numerator) / float64(rats[0].Denominator)
			}
		case "ExposureTime":
			if rats, ok := val.([]exifcommon.Rational); ok && len(rats) > 0 {
				r := rats[0]
				if r.Denominator == 1 {
					metadata.ShutterSpeed = fmt.Sprintf("%d", r.Numerator)
				} else if r.Numerator == 1 {
					metadata.ShutterSpeed = fmt.Sprintf("1/%d", r.Denominator)
				} else {
					metadata.ShutterSpeed = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
				}
			}
		case "ExposureBiasValue":
			if rats, ok := val.([]exifcommon.SignedRational); ok && len(rats) > 0 {
				metadata.ExposureCompensation = float64(rats[0].Numerator) / float64(rats[0].Denominator)
			}
		case "FocalLength":
			if rats, ok := val.([]exifcommon.Rational); ok && len(rats) > 0 {
				metadata.FocalLength = float64(rats[0].Numerator) / float64(rats[0].Denominator)
			}
		case "FocalLengthIn35mmFilm":
			if v, ok := val.([]uint16); ok && len(v) > 0 {
				metadata.FocalLength35mm = int(v[0])
			}

		// Temporal metadata
		case "DateTimeOriginal", "DateTime":
			if tagName == "DateTimeOriginal" || metadata.DateTaken.IsZero() {
				if dateStr, ok := val.(string); ok {
					if t, err := parseExifDateTime(dateStr); err == nil {
						metadata.DateTaken = t
					}
				}
			}
		case "DateTimeDigitized":
			if dateStr, ok := val.(string); ok {
				if t, err := parseExifDateTime(dateStr); err == nil {
					metadata.DateDigitized = t
				}
			}

		// Image properties
		case "PixelXDimension", "ImageWidth":
			switch v := val.(type) {
			case []uint32:
				if len(v) > 0 {
					metadata.Width = int(v[0])
				}
			case []uint16:
				if len(v) > 0 {
					metadata.Width = int(v[0])
				}
			}
		case "PixelYDimension", "ImageLength":
			switch v := val.(type) {
			case []uint32:
				if len(v) > 0 {
					metadata.Height = int(v[0])
				}
			case []uint16:
				if len(v) > 0 {
					metadata.Height = int(v[0])
				}
			}
		case "Orientation":
			if v, ok := val.([]uint16); ok && len(v) > 0 {
				metadata.Orientation = int(v[0])
			}
		case "ColorSpace":
			metadata.ColourSpace = fmt.Sprintf("%v", val)

		// GPS metadata
		case "GPSLatitude":
			if lat := parseGPSCoordinate(&entry); lat != 0 {
				metadata.Latitude = lat
			}
		case "GPSLongitude":
			if lon := parseGPSCoordinate(&entry); lon != 0 {
				metadata.Longitude = lon
			}
		case "GPSAltitude":
			if rats, ok := val.([]exifcommon.Rational); ok && len(rats) > 0 {
				metadata.Altitude = float64(rats[0].Numerator) / float64(rats[0].Denominator)
			}

		// Flash metadata
		case "Flash":
			switch v := val.(type) {
			case []uint16:
				if len(v) > 0 {
					// Flash tag is a bitmask, bit 0 indicates if flash fired
					metadata.FlashFired = (v[0] & 0x01) != 0
				}
			case uint16:
				metadata.FlashFired = (v & 0x01) != 0
			}

		// White balance
		case "WhiteBalance":
			metadata.WhiteBalance = fmt.Sprintf("%v", val)
		}
	}

	// Apply GPS reference directions
	for _, entry := range entries {
		val := entry.Value
		if val == nil {
			continue
		}

		switch entry.TagName {
		case "GPSLatitudeRef":
			if ref, ok := val.(string); ok && ref == "S" {
				metadata.Latitude = -metadata.Latitude
			}
		case "GPSLongitudeRef":
			if ref, ok := val.(string); ok && (ref == "W" || ref == "w") {
				metadata.Longitude = -metadata.Longitude
			}
		}
	}

	return metadata, nil
}

// parseGPSCoordinate parses GPS coordinate from EXIF rational array
func parseGPSCoordinate(entry *exif.ExifTag) float64 {
	if entry.Value == nil {
		return 0
	}

	rats, ok := entry.Value.([]exifcommon.Rational)
	if !ok || len(rats) < 3 {
		return 0
	}

	// GPS coordinates are stored as [degrees, minutes, seconds]
	degrees := float64(rats[0].Numerator) / float64(rats[0].Denominator)
	minutes := float64(rats[1].Numerator) / float64(rats[1].Denominator)
	seconds := float64(rats[2].Numerator) / float64(rats[2].Denominator)

	return degrees + minutes/60.0 + seconds/3600.0
}

// parseExifDateTime parses EXIF date/time with multiple format support
func parseExifDateTime(s string) (time.Time, error) {
	s = strings.Trim(s, "\x00 ")
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try multiple formats commonly found in EXIF data
	formats := []string{
		"2006:01:02 15:04:05",     // Standard EXIF format
		"2006:01:02 15:04:05.000", // With milliseconds
		"2006-01-02 15:04:05",     // ISO 8601 variant
		"2006-01-02T15:04:05",     // ISO 8601 with T
		"2006-01-02T15:04:05Z",    // ISO 8601 with Z
		"2006:01:02",              // Date only
		"2006-01-02",              // ISO date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
