package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adewale/olsen/pkg/models"
)

// PrintPerfStats outputs performance statistics in a human-friendly, machine-readable format
func PrintPerfStats(summary models.PerfSummary, detailed []models.PerfStats) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 PERFORMANCE STATISTICS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Summary statistics
	fmt.Println("SUMMARY:")
	fmt.Printf("  Total Photos:      %d\n", summary.TotalPhotos)
	fmt.Printf("  Processed:         %d\n", summary.ProcessedPhotos)
	fmt.Printf("  Skipped:           %d (unchanged)\n", summary.SkippedPhotos)
	fmt.Printf("  Updated:           %d (re-indexed)\n", summary.UpdatedPhotos)
	fmt.Printf("  Failed:            %d\n", summary.FailedPhotos)
	fmt.Printf("  Total Data:        %.2f MB\n", float64(summary.TotalBytes)/(1024*1024))
	fmt.Printf("  Throughput:        %.2f MB/s\n", summary.AvgThroughputMBps)
	fmt.Println()

	// Average timings per photo
	fmt.Println("AVERAGE TIMINGS PER PHOTO (processed only):")
	fmt.Printf("  Total:             %8.2f ms  (100.00%%)\n", summary.AvgTotalMs)
	fmt.Printf("  Hash:              %8.2f ms  (%6.2f%%)\n", summary.AvgHashMs, summary.AvgHashMs/summary.AvgTotalMs*100)
	fmt.Printf("  Metadata:          %8.2f ms  (%6.2f%%)\n", summary.AvgMetadataMs, summary.AvgMetadataMs/summary.AvgTotalMs*100)
	fmt.Printf("  Image Decode:      %8.2f ms  (%6.2f%%)\n", summary.AvgImageDecodeMs, summary.AvgImageDecodeMs/summary.AvgTotalMs*100)
	fmt.Printf("  Thumbnails:        %8.2f ms  (%6.2f%%)\n", summary.AvgThumbnailMs, summary.AvgThumbnailMs/summary.AvgTotalMs*100)
	fmt.Printf("  Color Extraction:  %8.2f ms  (%6.2f%%)\n", summary.AvgColorMs, summary.AvgColorMs/summary.AvgTotalMs*100)
	fmt.Printf("  Perceptual Hash:   %8.2f ms  (%6.2f%%)\n", summary.AvgPerceptualHashMs, summary.AvgPerceptualHashMs/summary.AvgTotalMs*100)
	fmt.Printf("  Inference:         %8.2f ms  (%6.2f%%)\n", summary.AvgInferenceMs, summary.AvgInferenceMs/summary.AvgTotalMs*100)
	fmt.Printf("  Database:          %8.2f ms  (%6.2f%%)\n", summary.AvgDatabaseMs, summary.AvgDatabaseMs/summary.AvgTotalMs*100)
	fmt.Println()

	// Pipeline breakdown visualization
	fmt.Println("PIPELINE BREAKDOWN (by time %):")
	printBar("Hash", summary.AvgHashMs, summary.AvgTotalMs)
	printBar("Metadata", summary.AvgMetadataMs, summary.AvgTotalMs)
	printBar("Decode", summary.AvgImageDecodeMs, summary.AvgTotalMs)
	printBar("Thumbnails", summary.AvgThumbnailMs, summary.AvgTotalMs)
	printBar("Color", summary.AvgColorMs, summary.AvgTotalMs)
	printBar("PHash", summary.AvgPerceptualHashMs, summary.AvgTotalMs)
	printBar("Inference", summary.AvgInferenceMs, summary.AvgTotalMs)
	printBar("Database", summary.AvgDatabaseMs, summary.AvgTotalMs)
	fmt.Println()

	// Top 10 slowest files
	if len(detailed) > 0 {
		// Find slowest processed (non-skipped) files
		slowest := make([]models.PerfStats, 0, len(detailed))
		for _, p := range detailed {
			if !p.WasSkipped && p.Error == "" {
				slowest = append(slowest, p)
			}
		}

		// Simple bubble sort for top 10
		for i := 0; i < len(slowest) && i < 10; i++ {
			for j := i + 1; j < len(slowest); j++ {
				if slowest[j].TotalTime > slowest[i].TotalTime {
					slowest[i], slowest[j] = slowest[j], slowest[i]
				}
			}
		}

		if len(slowest) > 0 {
			fmt.Println("TOP 10 SLOWEST PHOTOS:")
			limit := 10
			if len(slowest) < limit {
				limit = len(slowest)
			}
			for i := 0; i < limit; i++ {
				p := slowest[i]
				filename := filepath.Base(p.FilePath)
				if len(filename) > 50 {
					filename = filename[:47] + "..."
				}
				fmt.Printf("  %2d. %-50s  %8.2f ms\n", i+1, filename, float64(p.TotalTime.Milliseconds()))
			}
			fmt.Println()
		}

		// Failed files
		failed := make([]models.PerfStats, 0)
		for _, p := range detailed {
			if p.Error != "" {
				failed = append(failed, p)
			}
		}

		if len(failed) > 0 {
			fmt.Println("FAILED FILES:")
			for i, p := range failed {
				if i >= 20 {
					fmt.Printf("  ... and %d more\n", len(failed)-20)
					break
				}
				filename := filepath.Base(p.FilePath)
				fmt.Printf("  - %-50s  %s\n", filename, p.Error)
			}
			fmt.Println()
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printBar prints a horizontal bar chart
func printBar(label string, value, total float64) {
	percentage := value / total * 100
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	barWidth := 50
	filled := int(percentage / 100.0 * float64(barWidth))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("  %-12s [%s] %6.2f%%\n", label, bar, percentage)
}

// ExportPerfStatsJSON exports detailed performance statistics to a JSON file
func ExportPerfStatsJSON(filename string, summary models.PerfSummary, detailed []models.PerfStats) error {
	data := struct {
		Summary  models.PerfSummary `json:"summary"`
		Detailed []models.PerfStats `json:"detailed"`
	}{
		Summary:  summary,
		Detailed: detailed,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Performance stats exported to: %s\n", filename)
	return nil
}
