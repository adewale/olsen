//go:build cgo && benchmark_thumbnails
// +build cgo,benchmark_thumbnails

package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/bmp"

	"github.com/adewale/olsen/internal/quality"
)

func benchmarkThumbnailsCommand(args []string) error {
	fs := flag.NewFlagSet("benchmark-thumbnails", flag.ExitOnError)
	inputPath := fs.String("input", "", "Input directory with test images (required)")
	outputFile := fs.String("output", "thumbnail_benchmark.html", "Output HTML report file")
	targetSize := fs.Int("size", 512, "Target thumbnail size (longest edge)")
	useStandard := fs.Bool("standard", true, "Use standard comparison approaches")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *inputPath == "" {
		return fmt.Errorf("--input flag is required\n\nUsage:\n  olsen benchmark-thumbnails --input <directory> [--output report.html] [--size 512]")
	}

	// Check if input exists
	if _, err := os.Stat(*inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input directory not found: %s", *inputPath)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Thumbnail Quality Benchmark")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Input:  %s\n", *inputPath)
	fmt.Printf("Output: %s\n", *outputFile)
	fmt.Printf("Size:   %dpx (longest edge)\n", *targetSize)
	fmt.Println()

	// Find image files
	fmt.Println("Finding image files...")
	imageFiles, err := findImageFiles(*inputPath)
	if err != nil {
		return fmt.Errorf("failed to find images: %w", err)
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("no image files found in %s", *inputPath)
	}

	fmt.Printf("Found %d image files\n", len(imageFiles))
	fmt.Println()

	// Get comparison approaches
	var approaches []quality.ApproachConfig
	if *useStandard {
		approaches = quality.GetStandardApproaches()
	} else {
		return fmt.Errorf("custom approaches not yet implemented; use --standard")
	}

	fmt.Printf("Testing %d approaches:\n", len(approaches))
	for i, approach := range approaches {
		fmt.Printf("  %d. %s\n", i+1, approach.Name)
	}
	fmt.Println()

	// Process each image
	fmt.Println("Processing images...")
	startTime := time.Now()

	var allImageReports []quality.ImageReport
	resultsByApproach := make(map[string][]*quality.ComparisonResult)

	for _, imagePath := range imageFiles {
		fmt.Printf("  Processing %s... ", filepath.Base(imagePath))

		// Load image
		file, err := os.Open(imagePath)
		if err != nil {
			fmt.Printf("FAILED (open error: %v)\n", err)
			continue
		}

		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			fmt.Printf("FAILED (decode error: %v)\n", err)
			continue
		}

		// Run comparisons
		results, err := quality.CompareApproaches(img, approaches, uint(*targetSize))
		if err != nil {
			fmt.Printf("FAILED (comparison error: %v)\n", err)
			continue
		}

		// Store results
		bounds := img.Bounds()
		imageReport := quality.ImageReport{
			ImageName:   filepath.Base(imagePath),
			ImageWidth:  bounds.Dx(),
			ImageHeight: bounds.Dy(),
			Results:     results,
		}
		allImageReports = append(allImageReports, imageReport)

		// Group by approach for summary
		for _, result := range results {
			resultsByApproach[result.Config.Name] = append(resultsByApproach[result.Config.Name], result)
		}

		fmt.Printf("OK (%d approaches tested)\n", len(results))
	}

	duration := time.Since(startTime)
	fmt.Println()
	fmt.Printf("Processing complete: %d images in %v\n", len(allImageReports), duration.Round(time.Millisecond))
	fmt.Println()

	// Compute summaries
	fmt.Println("Computing summary statistics...")
	var summaries []quality.BenchmarkSummary
	for _, approach := range approaches {
		if results, ok := resultsByApproach[approach.Name]; ok {
			summary := quality.SummarizeResults(results)
			summaries = append(summaries, summary)
		}
	}

	// Generate report
	fmt.Printf("Generating HTML report: %s\n", *outputFile)
	reportData := quality.ReportData{
		GeneratedAt:  time.Now(),
		ImageReports: allImageReports,
		Summaries:    summaries,
	}

	outFile, err := os.Create(*outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := quality.GenerateHTMLReport(reportData, outFile); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ Benchmark complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Summary Statistics:")
	fmt.Println()

	// Print summary table
	fmt.Printf("%-35s  %8s  %8s  %10s  %10s  %8s\n",
		"Approach", "Avg SSIM", "Avg PSNR", "Avg Sharp", "Avg Size", "Avg Time")
	fmt.Println(strings.Repeat("─", 100))

	for _, summary := range summaries {
		fmt.Printf("%-35s  %8.4f  %8.2f  %10.0f  %10s  %8s\n",
			truncate(summary.ApproachName, 35),
			summary.AvgSSIM,
			summary.AvgPSNR,
			summary.AvgSharpness,
			formatBytes(summary.AvgFileSize),
			formatDuration(summary.AvgProcessingTime))
	}

	fmt.Println()
	fmt.Printf("Report saved to: %s\n", *outputFile)
	fmt.Println("Open in browser to view visual comparisons and detailed metrics.")
	fmt.Println()

	return nil
}

func findImageFiles(dir string) ([]string, error) {
	var files []string

	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".dng":  true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	kb := float64(b) / 1024.0
	return fmt.Sprintf("%.1f KB", kb)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0f μs", float64(d.Microseconds()))
	}
	return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000.0)
}
