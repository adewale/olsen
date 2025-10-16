//go:build cgo && benchmark_libraw
// +build cgo,benchmark_libraw

package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	golibraw "github.com/inokone/golibraw"
	golibrawnew "github.com/seppedelanghe/go-libraw"

	"github.com/adewale/olsen/internal/quality"
	"github.com/nfnt/resize"
)

func benchmarkLibrawCommand(args []string) error {
	fs := flag.NewFlagSet("benchmark-libraw", flag.ExitOnError)
	inputDir := fs.String("input", "testdata/dng", "Input directory with RAW files")
	output := fs.String("output", "libraw_benchmark.html", "Output HTML report")

	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("LibRaw Library Benchmark")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Input:  %s\n", *inputDir)
	fmt.Printf("Output: %s\n", *output)
	fmt.Println()

	// Find DNG files
	files, err := findDNGFiles(*inputDir)
	if err != nil {
		return fmt.Errorf("failed to find DNG files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no DNG files found in %s", *inputDir)
	}

	fmt.Printf("Found %d DNG files\n\n", len(files))

	// Test approaches
	approaches := []LibRawApproach{
		{
			Name:        "Current (golibraw - defaults)",
			UseGoLibraw: true,
		},
		{
			Name:        "go-libraw (Linear, 8-bit, sRGB, Camera WB)",
			UseGoLibraw: false,
			Config: golibrawnew.ProcessorOptions{
				UserQual:    0, // Linear
				OutputBps:   8,
				OutputColor: golibrawnew.SRGB,
				UseCameraWb: true,
			},
		},
		{
			Name:        "go-libraw (VNG, 8-bit, sRGB, Camera WB)",
			UseGoLibraw: false,
			Config: golibrawnew.ProcessorOptions{
				UserQual:    1, // VNG
				OutputBps:   8,
				OutputColor: golibrawnew.SRGB,
				UseCameraWb: true,
			},
		},
		{
			Name:        "go-libraw (PPG, 8-bit, sRGB, Camera WB)",
			UseGoLibraw: false,
			Config: golibrawnew.ProcessorOptions{
				UserQual:    2, // PPG
				OutputBps:   8,
				OutputColor: golibrawnew.SRGB,
				UseCameraWb: true,
			},
		},
		{
			Name:        "go-libraw (AHD, 8-bit, sRGB, Camera WB)",
			UseGoLibraw: false,
			Config: golibrawnew.ProcessorOptions{
				UserQual:    3, // AHD (best quality)
				OutputBps:   8,
				OutputColor: golibrawnew.SRGB,
				UseCameraWb: true,
			},
		},
		{
			Name:        "go-libraw (AHD, 16-bit, sRGB, Camera WB)",
			UseGoLibraw: false,
			Config: golibrawnew.ProcessorOptions{
				UserQual:    3, // AHD
				OutputBps:   16,
				OutputColor: golibrawnew.SRGB,
				UseCameraWb: true,
			},
		},
	}

	fmt.Printf("Testing %d approaches:\n", len(approaches))
	for i, approach := range approaches {
		fmt.Printf("  %d. %s\n", i+1, approach.Name)
	}
	fmt.Println()

	// Benchmark each file
	results := make(map[string][]*LibRawResult)

	fmt.Println("Processing RAW files...")
	for _, file := range files {
		basename := filepath.Base(file)
		fmt.Printf("  Processing %s... ", basename)

		fileResults := make([]*LibRawResult, 0, len(approaches))

		for _, approach := range approaches {
			result, err := benchmarkApproach(file, approach)
			if err != nil {
				fmt.Printf("\n    ⚠️  %s FAILED: %v\n", approach.Name, err)
				continue
			}
			fileResults = append(fileResults, result)
		}

		results[basename] = fileResults
		fmt.Printf("OK (%d approaches tested)\n", len(fileResults))
	}

	// Generate report
	fmt.Println()
	fmt.Println("Generating report...")
	if err := generateLibRawReport(results, *output); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Print summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ Benchmark complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	printLibRawSummary(results, approaches)

	fmt.Printf("\nReport saved to: %s\n", *output)
	fmt.Println("Open in browser to view detailed comparisons.")

	return nil
}

type LibRawApproach struct {
	Name        string
	UseGoLibraw bool                         // true = inokone/golibraw, false = seppedelanghe/go-libraw
	Config      golibrawnew.ProcessorOptions // Only used if UseGoLibraw == false
}

type LibRawResult struct {
	Approach      LibRawApproach
	DecodeTime    time.Duration
	ThumbnailTime time.Duration
	TotalTime     time.Duration
	ImageWidth    int
	ImageHeight   int
	Metrics       quality.Metrics
	Error         error
}

func benchmarkApproach(filePath string, approach LibRawApproach) (*LibRawResult, error) {
	result := &LibRawResult{
		Approach: approach,
	}

	startTotal := time.Now()

	// Decode RAW
	startDecode := time.Now()
	var img image.Image
	var err error

	if approach.UseGoLibraw {
		// Current library: inokone/golibraw
		img, err = golibraw.ImportRaw(filePath)
	} else {
		// New library: seppedelanghe/go-libraw
		processor := golibrawnew.NewProcessor(approach.Config)
		img, _, err = processor.ProcessRaw(filePath)
	}

	result.DecodeTime = time.Since(startDecode)

	if err != nil {
		result.Error = err
		return result, err
	}

	if img == nil {
		result.Error = fmt.Errorf("decoded image is nil")
		return result, result.Error
	}

	result.ImageWidth = img.Bounds().Dx()
	result.ImageHeight = img.Bounds().Dy()

	// Generate thumbnail and compute quality metrics
	startThumb := time.Now()
	thumbnail := generateTestThumbnail(img, 512)
	result.ThumbnailTime = time.Since(startThumb)

	// Compute metrics vs reference (using Lanczos3 as reference)
	reference := generateReferenceThumbnail(img, 512)
	metrics, err := quality.ComputeAllMetrics(reference, thumbnail)
	if err == nil {
		result.Metrics = metrics
	}

	result.TotalTime = time.Since(startTotal)

	return result, nil
}

func generateTestThumbnail(img image.Image, targetSize uint) image.Image {
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newWidth, newHeight uint
	if width > height {
		newWidth = targetSize
		newHeight = 0
	} else {
		newWidth = 0
		newHeight = targetSize
	}

	// Use Lanczos3 resize
	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

func generateReferenceThumbnail(img image.Image, targetSize uint) image.Image {
	// Same as test for now
	return generateTestThumbnail(img, targetSize)
}

func findDNGFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ext := filepath.Ext(path)
		if !info.IsDir() && (ext == ".dng" || ext == ".DNG") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func printLibRawSummary(results map[string][]*LibRawResult, approaches []LibRawApproach) {
	fmt.Println("Summary Statistics:")
	fmt.Println()
	fmt.Printf("%-50s  %12s  %12s  %10s  %10s\n", "Approach", "Avg Decode", "Avg Total", "Avg SSIM", "Avg PSNR")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────────────")

	for _, approach := range approaches {
		var totalDecode, totalTime time.Duration
		var totalSSIM, totalPSNR float64
		count := 0

		for _, fileResults := range results {
			for _, result := range fileResults {
				if result.Approach.Name == approach.Name && result.Error == nil {
					totalDecode += result.DecodeTime
					totalTime += result.TotalTime
					totalSSIM += result.Metrics.SSIM
					totalPSNR += result.Metrics.PSNR
					count++
				}
			}
		}

		if count > 0 {
			avgDecode := totalDecode / time.Duration(count)
			avgTotal := totalTime / time.Duration(count)
			avgSSIM := totalSSIM / float64(count)
			avgPSNR := totalPSNR / float64(count)

			fmt.Printf("%-50s  %9.1f ms  %9.1f ms  %10.4f  %10.2f\n",
				approach.Name,
				float64(avgDecode.Microseconds())/1000.0,
				float64(avgTotal.Microseconds())/1000.0,
				avgSSIM,
				avgPSNR)
		}
	}
}

func generateLibRawReport(results map[string][]*LibRawResult, outputPath string) error {
	// For now, just create a simple text report
	// TODO: Create HTML report with visual comparisons
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "<html><body><h1>LibRaw Benchmark Results</h1>\n")
	fmt.Fprintf(file, "<p>Generated: %s</p>\n", time.Now().Format(time.RFC3339))

	for filename, fileResults := range results {
		fmt.Fprintf(file, "<h2>%s</h2>\n", filename)
		fmt.Fprintf(file, "<table border='1'>\n")
		fmt.Fprintf(file, "<tr><th>Approach</th><th>Decode Time</th><th>SSIM</th><th>PSNR</th><th>Dimensions</th></tr>\n")

		for _, result := range fileResults {
			if result.Error == nil {
				fmt.Fprintf(file, "<tr><td>%s</td><td>%.1f ms</td><td>%.4f</td><td>%.2f dB</td><td>%dx%d</td></tr>\n",
					result.Approach.Name,
					float64(result.DecodeTime.Microseconds())/1000.0,
					result.Metrics.SSIM,
					result.Metrics.PSNR,
					result.ImageWidth,
					result.ImageHeight)
			} else {
				fmt.Fprintf(file, "<tr><td>%s</td><td colspan='4'>ERROR: %v</td></tr>\n",
					result.Approach.Name,
					result.Error)
			}
		}

		fmt.Fprintf(file, "</table>\n")
	}

	fmt.Fprintf(file, "</body></html>\n")

	return nil
}
