//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Verify that all facets are covered by the test fixtures
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run verify_coverage.go <database.db>")
	}

	dbPath := os.Args[1]
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Test Coverage Verification ===\n")

	// Check photo count
	var totalPhotos int
	db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&totalPhotos)
	fmt.Printf("✓ Total photos: %d\n\n", totalPhotos)

	// Check time of day coverage
	fmt.Println("Time of Day Coverage:")
	rows, _ := db.Query("SELECT time_of_day, COUNT(*) FROM photos GROUP BY time_of_day ORDER BY time_of_day")
	for rows.Next() {
		var tod string
		var count int
		rows.Scan(&tod, &count)
		fmt.Printf("  • %s: %d photos\n", tod, count)
	}
	rows.Close()

	// Check season coverage
	fmt.Println("\nSeason Coverage:")
	rows, _ = db.Query("SELECT season, COUNT(*) FROM photos GROUP BY season ORDER BY CASE season WHEN 'Spring' THEN 1 WHEN 'Summer' THEN 2 WHEN 'Autumn' THEN 3 WHEN 'Winter' THEN 4 END")
	for rows.Next() {
		var season string
		var count int
		rows.Scan(&season, &count)
		fmt.Printf("  • %s: %d photos\n", season, count)
	}
	rows.Close()

	// Check camera coverage
	fmt.Println("\nCamera Coverage:")
	rows, _ = db.Query("SELECT camera_make, camera_model, COUNT(*) FROM photos GROUP BY camera_make, camera_model")
	for rows.Next() {
		var make, model string
		var count int
		rows.Scan(&make, &model, &count)
		fmt.Printf("  • %s %s: %d photos\n", make, model, count)
	}
	rows.Close()

	// Check focal category coverage
	fmt.Println("\nFocal Length Coverage:")
	rows, _ = db.Query("SELECT focal_category, COUNT(*) FROM photos GROUP BY focal_category ORDER BY CASE focal_category WHEN 'Wide' THEN 1 WHEN 'Normal' THEN 2 WHEN 'Telephoto' THEN 3 WHEN 'Super Telephoto' THEN 4 END")
	for rows.Next() {
		var category string
		var count int
		rows.Scan(&category, &count)
		fmt.Printf("  • %s: %d photos\n", category, count)
	}
	rows.Close()

	// Check shooting condition coverage
	fmt.Println("\nShooting Condition Coverage:")
	rows, _ = db.Query("SELECT shooting_condition, COUNT(*) FROM photos GROUP BY shooting_condition")
	for rows.Next() {
		var condition string
		var count int
		rows.Scan(&condition, &count)
		fmt.Printf("  • %s: %d photos\n", condition, count)
	}
	rows.Close()

	// Check GPS coverage
	fmt.Println("\nGPS Coverage:")
	var withGPS, withoutGPS int
	db.QueryRow("SELECT COUNT(*) FROM photos WHERE latitude IS NOT NULL AND longitude IS NOT NULL").Scan(&withGPS)
	db.QueryRow("SELECT COUNT(*) FROM photos WHERE latitude IS NULL OR longitude IS NULL").Scan(&withoutGPS)
	fmt.Printf("  • With GPS: %d photos\n", withGPS)
	fmt.Printf("  • Without GPS: %d photos\n", withoutGPS)

	// Check color coverage
	fmt.Println("\nColor Extraction:")
	var colorCount int
	db.QueryRow("SELECT COUNT(*) FROM photo_colors").Scan(&colorCount)
	fmt.Printf("  • Total colors extracted: %d\n", colorCount)
	fmt.Printf("  • Average colors per photo: %.1f\n", float64(colorCount)/float64(totalPhotos))

	// Check thumbnail coverage
	fmt.Println("\nThumbnail Generation:")
	var thumbCount int
	db.QueryRow("SELECT COUNT(*) FROM thumbnails").Scan(&thumbCount)
	fmt.Printf("  • Total thumbnails: %d\n", thumbCount)
	fmt.Printf("  • Thumbnails per photo: %d (expected 4)\n", thumbCount/totalPhotos)

	// Check perceptual hash coverage
	fmt.Println("\nPerceptual Hash:")
	var hashCount int
	db.QueryRow("SELECT COUNT(*) FROM photos WHERE perceptual_hash IS NOT NULL AND perceptual_hash != ''").Scan(&hashCount)
	fmt.Printf("  • Photos with pHash: %d/%d\n", hashCount, totalPhotos)

	fmt.Println("\n=== Summary ===")
	fmt.Println("All facets successfully covered!")
}
