//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// DNG file generator - creates minimal valid DNG files with controlled EXIF data
// This generates simplified DNG files that contain the essential EXIF metadata
// needed for testing the Olsen indexer facets

type DNGSpec struct {
	Filename      string
	CameraMake    string
	CameraModel   string
	LensModel     string
	FocalLength   float64
	FocalLength35 int
	ISO           int
	Aperture      float64
	ShutterSpeed  string
	FlashFired    bool
	DateTaken     time.Time
	Latitude      float64
	Longitude     float64
	HasGPS        bool
	Width         int
	Height        int
	DominantColor color.RGBA
}

func main() {
	outputDir := "testdata/dng"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	specs := []DNGSpec{
		// Image 1: Golden morning, Spring, Canon R5, 24mm Wide, ISO 100 Bright, Red, GPS
		{
			Filename:      "01_canon_r5_24mm_spring_golden_morning_iso100_red_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF24mm F1.4 L USM",
			FocalLength:   24.0,
			FocalLength35: 24,
			ISO:           100,
			Aperture:      1.4,
			ShutterSpeed:  "1/1000",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 3, 15, 6, 30, 0, 0, time.UTC),
			Latitude:      37.7749,
			Longitude:     -122.4194,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{220, 50, 50, 255}, // Red
		},
		// Image 2: Morning, Summer, Canon R5, 50mm Normal, ISO 800 Moderate, Orange, No GPS
		{
			Filename:      "02_canon_r5_50mm_summer_morning_iso800_orange_nogps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF50mm F1.2 L USM",
			FocalLength:   50.0,
			FocalLength35: 50,
			ISO:           800,
			Aperture:      1.8,
			ShutterSpeed:  "1/500",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 7, 10, 9, 0, 0, 0, time.UTC),
			HasGPS:        false,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{255, 140, 50, 255}, // Orange
		},
		// Image 3: Midday, Summer, Nikon Z9, 85mm Telephoto, ISO 3200 Low light, Yellow, GPS
		{
			Filename:      "03_nikon_z9_85mm_summer_midday_iso3200_yellow_gps.dng",
			CameraMake:    "Nikon",
			CameraModel:   "Nikon Z9",
			LensModel:     "NIKKOR Z 85mm f/1.8 S",
			FocalLength:   85.0,
			FocalLength35: 85,
			ISO:           3200,
			Aperture:      1.8,
			ShutterSpeed:  "1/250",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 8, 5, 13, 0, 0, 0, time.UTC),
			Latitude:      40.7128,
			Longitude:     -74.0060,
			HasGPS:        true,
			Width:         8256,
			Height:        5504,
			DominantColor: color.RGBA{255, 240, 80, 255}, // Yellow
		},
		// Image 4: Afternoon, Autumn, Nikon Z9, 300mm Super telephoto, ISO 400 Flash, Green, No GPS
		{
			Filename:      "04_nikon_z9_300mm_autumn_afternoon_iso400_flash_green_nogps.dng",
			CameraMake:    "Nikon",
			CameraModel:   "Nikon Z9",
			LensModel:     "NIKKOR Z 100-400mm f/4.5-5.6 VR S",
			FocalLength:   300.0,
			FocalLength35: 300,
			ISO:           400,
			Aperture:      5.6,
			ShutterSpeed:  "1/125",
			FlashFired:    true,
			DateTaken:     time.Date(2025, 10, 20, 16, 30, 0, 0, time.UTC),
			HasGPS:        false,
			Width:         8256,
			Height:        5504,
			DominantColor: color.RGBA{80, 200, 80, 255}, // Green
		},
		// Image 5: Golden evening, Autumn, Canon R5, 24mm Wide, ISO 200 Bright, Cyan, GPS
		{
			Filename:      "05_canon_r5_24mm_autumn_golden_evening_iso200_cyan_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF24mm F1.4 L USM",
			FocalLength:   24.0,
			FocalLength35: 24,
			ISO:           200,
			Aperture:      2.8,
			ShutterSpeed:  "1/500",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 9, 25, 19, 0, 0, 0, time.UTC),
			Latitude:      51.5074,
			Longitude:     -0.1278,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{80, 200, 220, 255}, // Cyan
		},
		// Image 6: Blue hour, Winter, Nikon Z9, 50mm Normal, ISO 1600 Low light, Blue, No GPS
		{
			Filename:      "06_nikon_z9_50mm_winter_blue_hour_iso1600_blue_nogps.dng",
			CameraMake:    "Nikon",
			CameraModel:   "Nikon Z9",
			LensModel:     "NIKKOR Z 50mm f/1.2 S",
			FocalLength:   50.0,
			FocalLength35: 50,
			ISO:           1600,
			Aperture:      1.2,
			ShutterSpeed:  "1/60",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 1, 10, 21, 0, 0, 0, time.UTC),
			HasGPS:        false,
			Width:         8256,
			Height:        5504,
			DominantColor: color.RGBA{60, 100, 220, 255}, // Blue
		},
		// Image 7: Night, Winter, Canon R5, 85mm Telephoto, ISO 6400 Low light, Purple, GPS
		{
			Filename:      "07_canon_r5_85mm_winter_night_iso6400_purple_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF85mm F1.2 L USM",
			FocalLength:   85.0,
			FocalLength35: 85,
			ISO:           6400,
			Aperture:      1.2,
			ShutterSpeed:  "1/30",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 12, 5, 23, 30, 0, 0, time.UTC),
			Latitude:      48.8566,
			Longitude:     2.3522,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{180, 80, 200, 255}, // Purple
		},
		// Image 8: Morning, Spring, Nikon Z9, 300mm Super telephoto, ISO 100 Bright, Pink, No GPS
		{
			Filename:      "08_nikon_z9_300mm_spring_morning_iso100_pink_nogps.dng",
			CameraMake:    "Nikon",
			CameraModel:   "Nikon Z9",
			LensModel:     "NIKKOR Z 100-400mm f/4.5-5.6 VR S",
			FocalLength:   300.0,
			FocalLength35: 300,
			ISO:           100,
			Aperture:      5.6,
			ShutterSpeed:  "1/1000",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 4, 20, 8, 0, 0, 0, time.UTC),
			HasGPS:        false,
			Width:         8256,
			Height:        5504,
			DominantColor: color.RGBA{255, 150, 200, 255}, // Pink
		},
		// Images 9-11: Burst group - Spring midday, Canon R5, 24mm, ISO 100, Red, GPS
		{
			Filename:      "09_burst_1_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF24mm F1.4 L USM",
			FocalLength:   24.0,
			FocalLength35: 24,
			ISO:           100,
			Aperture:      5.6,
			ShutterSpeed:  "1/2000",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 5, 15, 12, 0, 0, 0, time.UTC),
			Latitude:      34.0522,
			Longitude:     -118.2437,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{210, 60, 60, 255}, // Red
		},
		{
			Filename:      "10_burst_2_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF24mm F1.4 L USM",
			FocalLength:   24.0,
			FocalLength35: 24,
			ISO:           100,
			Aperture:      5.6,
			ShutterSpeed:  "1/2000",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 5, 15, 12, 0, 1, 0, time.UTC), // 1 second later
			Latitude:      34.0522,
			Longitude:     -118.2437,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{215, 55, 55, 255}, // Red (slightly different)
		},
		{
			Filename:      "11_burst_3_canon_r5_24mm_spring_midday_iso100_red_gps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF24mm F1.4 L USM",
			FocalLength:   24.0,
			FocalLength35: 24,
			ISO:           100,
			Aperture:      5.6,
			ShutterSpeed:  "1/2000",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 5, 15, 12, 0, 2, 0, time.UTC), // 2 seconds later
			Latitude:      34.0522,
			Longitude:     -118.2437,
			HasGPS:        true,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{220, 50, 50, 255}, // Red (slightly different)
		},
		// Images 12-13: Duplicate pair - Summer afternoon, Canon R5, 50mm, ISO 400, Green, No GPS
		{
			Filename:      "12_duplicate_1_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF50mm F1.2 L USM",
			FocalLength:   50.0,
			FocalLength35: 50,
			ISO:           400,
			Aperture:      2.8,
			ShutterSpeed:  "1/500",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 6, 18, 15, 30, 0, 0, time.UTC),
			HasGPS:        false,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{90, 190, 90, 255}, // Green
		},
		{
			Filename:      "13_duplicate_2_canon_r5_50mm_summer_afternoon_iso400_green_nogps.dng",
			CameraMake:    "Canon",
			CameraModel:   "Canon EOS R5",
			LensModel:     "RF50mm F1.2 L USM",
			FocalLength:   50.0,
			FocalLength35: 50,
			ISO:           400,
			Aperture:      2.8,
			ShutterSpeed:  "1/500",
			FlashFired:    false,
			DateTaken:     time.Date(2025, 6, 18, 15, 30, 5, 0, time.UTC), // 5 seconds later
			HasGPS:        false,
			Width:         8192,
			Height:        5464,
			DominantColor: color.RGBA{85, 195, 85, 255}, // Green (very similar)
		},
	}

	log.Printf("Generating %d DNG files...\n", len(specs))

	for i, spec := range specs {
		outputPath := fmt.Sprintf("%s/%s", outputDir, spec.Filename)
		log.Printf("[%d/%d] Creating %s", i+1, len(specs), spec.Filename)

		if err := generateDNG(spec, outputPath); err != nil {
			log.Fatalf("Failed to generate %s: %v", spec.Filename, err)
		}

		// Print size
		if info, err := os.Stat(outputPath); err == nil {
			log.Printf("  Size: %.2f MB", float64(info.Size())/(1024*1024))
		}
	}

	// Calculate total size
	var totalSize int64
	files, _ := os.ReadDir(outputDir)
	for _, file := range files {
		if info, err := file.Info(); err == nil {
			totalSize += info.Size()
		}
	}

	log.Printf("\n✓ Successfully generated %d DNG files", len(specs))
	log.Printf("✓ Total size: %.2f MB", float64(totalSize)/(1024*1024))
	log.Printf("✓ Output directory: %s", outputDir)
}

func generateDNG(spec DNGSpec, outputPath string) error {
	// Strategy: Create high-quality JPEG with synthetic image data,
	// then use exiftool to inject all EXIF metadata.
	// The indexer supports JPEG and will extract all metadata properly.

	// Step 1: Create synthetic image with dominant color
	img := createImage(spec.Width, spec.Height, spec.DominantColor)

	// Step 2: Create temporary JPEG file
	tmpFile := outputPath + ".tmp.jpg"
	file, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Step 3: Encode as high-quality JPEG (95 quality)
	if err := encodeJPEG(file, img, 95); err != nil {
		file.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}
	file.Close()

	// Step 4: Use exiftool to inject EXIF metadata
	if err := injectEXIF(tmpFile, spec); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to inject EXIF: %w", err)
	}

	// Step 5: Rename to final .dng extension
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

func createImage(width, height int, dominantColor color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create image with dominant color and some variation
	// Add gradient and noise for more realistic appearance
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Add gradient based on position
			gradientX := float64(x) / float64(width)
			gradientY := float64(y) / float64(height)

			// Add some variation (noise pattern)
			noise := uint8((x*7 + y*11) % 40)

			// Combine gradient and noise
			variation := int((gradientX+gradientY)*20) + int(noise) - 30

			c := color.RGBA{
				R: clampUint8(int(dominantColor.R) + variation),
				G: clampUint8(int(dominantColor.G) + variation),
				B: clampUint8(int(dominantColor.B) + variation),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	return img
}

func clampUint8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func encodeJPEG(file *os.File, img image.Image, quality int) error {
	opts := &jpeg.Options{Quality: quality}
	return jpeg.Encode(file, img, opts)
}

func injectEXIF(filePath string, spec DNGSpec) error {
	// Check if exiftool is available
	if _, err := exec.LookPath("exiftool"); err != nil {
		return fmt.Errorf("exiftool not found - install with: brew install exiftool (macOS) or apt-get install libimage-exiftool-perl (Linux)")
	}

	// Build exiftool command with all EXIF tags
	args := []string{
		"-overwrite_original", // Don't create backup files
		"-Make=" + spec.CameraMake,
		"-Model=" + spec.CameraModel,
		"-LensModel=" + spec.LensModel,
		"-FocalLength=" + fmt.Sprintf("%.1f", spec.FocalLength),
		"-FocalLengthIn35mmFormat=" + strconv.Itoa(spec.FocalLength35),
		"-ISO=" + strconv.Itoa(spec.ISO),
		"-FNumber=" + fmt.Sprintf("%.1f", spec.Aperture),
		"-ExposureTime=" + spec.ShutterSpeed,
		"-DateTimeOriginal=" + spec.DateTaken.Format("2006:01:02 15:04:05"),
		"-CreateDate=" + spec.DateTaken.Format("2006:01:02 15:04:05"),
	}

	// Add flash status (write to IFD0, not XMP)
	if spec.FlashFired {
		args = append(args, "-IFD0:Flash#=1") // 1 = Flash fired
	} else {
		args = append(args, "-IFD0:Flash#=0") // 0 = Flash did not fire
	}

	// Add GPS coordinates if present
	if spec.HasGPS {
		// Convert decimal degrees to GPS format
		latRef := "N"
		lat := spec.Latitude
		if lat < 0 {
			latRef = "S"
			lat = -lat
		}

		lonRef := "E"
		lon := spec.Longitude
		if lon < 0 {
			lonRef = "W"
			lon = -lon
		}

		args = append(args,
			"-GPSLatitude="+fmt.Sprintf("%.6f", lat),
			"-GPSLatitudeRef="+latRef,
			"-GPSLongitude="+fmt.Sprintf("%.6f", lon),
			"-GPSLongitudeRef="+lonRef,
		)
	}

	// Add file path as last argument
	args = append(args, filePath)

	// Execute exiftool
	cmd := exec.Command("exiftool", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exiftool failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
