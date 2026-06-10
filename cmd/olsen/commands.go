package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/explorer"
	"github.com/adewale/olsen/internal/indexer"
	"github.com/adewale/olsen/pkg/models"
)

// indexCommand performs actual photo indexing
func indexCommand(photoDir, dbPath string, workers int, perfstats bool) error {
	// Validate photo directory
	if info, err := os.Stat(photoDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("photo directory does not exist: %s", photoDir)
		}
		return fmt.Errorf("cannot access photo directory: %v", err)
	} else if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", photoDir)
	}

	// Open/create database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create indexer engine
	engine := indexer.NewEngine(db, workers)

	// Index directory
	fmt.Println("Indexing photos...")
	fmt.Printf("  Directory: %s\n", photoDir)
	fmt.Printf("  Database: %s\n", dbPath)
	fmt.Printf("  Workers: %d\n", workers)
	fmt.Println()

	startTime := time.Now()

	// Set up progress callback with inline progress bar
	var lastUpdate time.Time
	engine.SetProgressCallback(func(processed, total int) {
		// Throttle updates to avoid excessive terminal writes (max 4 updates/sec)
		now := time.Now()
		if now.Sub(lastUpdate) < 250*time.Millisecond && processed < total {
			return
		}
		lastUpdate = now

		// Calculate progress
		percent := float64(processed) / float64(total) * 100
		elapsed := time.Since(startTime).Seconds()
		rate := float64(processed) / elapsed

		// Draw progress bar (40 chars wide)
		barWidth := 40
		filled := int(float64(barWidth) * float64(processed) / float64(total))
		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filled {
				bar += "="
			} else if i == filled {
				bar += ">"
			} else {
				bar += " "
			}
		}

		// Print progress line (use \r to overwrite)
		fmt.Printf("\r[%s] %6.2f%% | %d/%d photos | %.1f photos/sec",
			bar, percent, processed, total, rate)
	})

	// Stop cleanly on Ctrl+C / SIGTERM: in-flight photos are abandoned at the
	// next stage boundary, committed photos are kept, and a re-run resumes.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = engine.IndexDirectoryContext(ctx, photoDir)
	interrupted := errors.Is(err, context.Canceled)
	if err != nil && !interrupted {
		return fmt.Errorf("indexing failed: %v", err)
	}

	// Clear progress line and print newline
	fmt.Println()

	// Get final stats
	stats := engine.GetStats()

	if interrupted {
		fmt.Printf("\n\nIndexing interrupted after %s (already-indexed photos were kept; re-run to resume)\n",
			time.Since(startTime).Round(time.Millisecond))
	} else {
		fmt.Printf("\n\nIndexing complete in %s\n", time.Since(startTime).Round(time.Millisecond))
	}
	fmt.Printf("  Found: %d files\n", stats.FilesFound)
	fmt.Printf("  Processed: %d photos\n", stats.FilesProcessed)
	fmt.Printf("  Skipped: %d photos\n", stats.FilesSkipped)
	if stats.FilesFailed > 0 {
		fmt.Printf("  Failed: %d photos\n", stats.FilesFailed)
	}
	fmt.Printf("  Database: %s\n", dbPath)

	return nil
}

// statsCommand displays database statistics
func statsCommand(dbPath string) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Get photo count
	var photoCount int
	err = db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&photoCount)
	if err != nil {
		return fmt.Errorf("failed to query photo count: %v", err)
	}

	fmt.Println("Database Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Total photos: %d\n", photoCount)

	// Get camera counts
	rows, err := db.Query(`
		SELECT camera_make || ' ' || camera_model as camera, COUNT(*) as count
		FROM photos
		WHERE camera_make IS NOT NULL
		GROUP BY camera_make, camera_model
		ORDER BY count DESC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()

		fmt.Println("\nTop 5 Cameras:")
		for rows.Next() {
			var camera string
			var count int
			if err := rows.Scan(&camera, &count); err == nil {
				fmt.Printf("  %s: %d photos\n", camera, count)
			}
		}
	}

	// Get year distribution
	rows, err = db.Query(`
		SELECT strftime('%Y', date_taken) as year, COUNT(*) as count
		FROM photos
		WHERE date_taken IS NOT NULL
		GROUP BY year
		ORDER BY year DESC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()

		fmt.Println("\nPhotos by Year:")
		for rows.Next() {
			var year string
			var count int
			if err := rows.Scan(&year, &count); err == nil {
				fmt.Printf("  %s: %d photos\n", year, count)
			}
		}
	}

	return nil
}

// analyzeCommand performs burst detection and duplicate analysis
func analyzeCommand(dbPath string) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Println("Analyzing photos...")

	// Detect bursts
	fmt.Println("  Detecting burst sequences...")
	burstDetector := indexer.NewBurstDetector(db)
	burstCount, err := burstDetector.DetectBursts()
	if err != nil {
		return fmt.Errorf("burst detection failed: %v", err)
	}

	fmt.Printf("\nAnalysis complete\n")
	fmt.Printf("  Burst groups detected: %d\n", burstCount)

	return nil
}

// exploreCommand starts the web explorer server
func exploreCommand(dbPath, addr string, openBrowser bool) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Start server
	fmt.Println("Starting Olsen Photo Explorer...")
	fmt.Printf("  Database: %s\n", dbPath)
	fmt.Printf("  Address: http://%s\n", addr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	server := explorer.NewServer(db, addr)

	// Shut down gracefully on Ctrl+C / SIGTERM, letting in-flight
	// requests finish.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go func() { errChan <- server.Start() }()

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server failed: %v", err)
		}
		return nil
	case <-ctx.Done():
		fmt.Println("\nShutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown failed: %v", err)
		}
		return nil
	}
}

// showCommand displays metadata for a specific photo
func showCommand(dbPath string, photoID int) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Query photo metadata
	var filePath, cameraMake, cameraModel, lensModel string
	var dateTaken, indexedAt sql.NullString
	var iso sql.NullInt64
	var aperture, shutterSpeed, focalLength sql.NullFloat64

	err = db.QueryRow(`
		SELECT file_path, camera_make, camera_model, lens_model,
		       date_taken, iso, aperture, shutter_speed, focal_length, indexed_at
		FROM photos
		WHERE id = ?
	`, photoID).Scan(
		&filePath, &cameraMake, &cameraModel, &lensModel,
		&dateTaken, &iso, &aperture, &shutterSpeed, &focalLength, &indexedAt,
	)

	if err == sql.ErrNoRows {
		return fmt.Errorf("photo not found: %d", photoID)
	}
	if err != nil {
		return fmt.Errorf("failed to query photo: %v", err)
	}

	// Display metadata
	fmt.Printf("Photo #%d\n", photoID)
	fmt.Println("━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("File: %s\n", filePath)
	if dateTaken.Valid {
		fmt.Printf("Date taken: %s\n", dateTaken.String)
	}
	fmt.Printf("Camera: %s %s\n", cameraMake, cameraModel)
	if lensModel != "" {
		fmt.Printf("Lens: %s\n", lensModel)
	}
	if iso.Valid {
		fmt.Printf("ISO: %d\n", iso.Int64)
	}
	if aperture.Valid {
		fmt.Printf("Aperture: f/%.1f\n", aperture.Float64)
	}
	if shutterSpeed.Valid {
		fmt.Printf("Shutter speed: %.4fs\n", shutterSpeed.Float64)
	}
	if focalLength.Valid {
		fmt.Printf("Focal length: %.1fmm\n", focalLength.Float64)
	}
	if indexedAt.Valid {
		fmt.Printf("Indexed: %s\n", indexedAt.String)
	}

	return nil
}

// thumbnailCommand extracts a thumbnail from a photo
func thumbnailCommand(dbPath string, photoID int, outputPath string, size int) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Validate the size at the boundary
	thumbnailSize, err := models.ParseThumbnailSize(strconv.Itoa(size))
	if err != nil {
		return err
	}

	// Query thumbnail
	var thumbnailData []byte
	err = db.QueryRow(`
		SELECT thumbnail_data
		FROM thumbnails
		WHERE photo_id = ? AND size = ?
	`, photoID, thumbnailSize).Scan(&thumbnailData)

	if err == sql.ErrNoRows {
		return fmt.Errorf("thumbnail not found for photo %d at size %d", photoID, size)
	}
	if err != nil {
		return fmt.Errorf("failed to query thumbnail: %v", err)
	}

	// Write thumbnail to file
	if err := os.WriteFile(outputPath, thumbnailData, 0644); err != nil {
		return fmt.Errorf("failed to write thumbnail: %v", err)
	}

	fmt.Printf("Thumbnail saved to: %s\n", outputPath)
	return nil
}

// verifyCommand verifies database integrity
func verifyCommand(dbPath string) error {
	// Check database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database not found: %s", dbPath)
	}

	// Open database
	db, err := database.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Println("Verifying database integrity...")

	// Check photo count
	var photoCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&photoCount); err != nil {
		return fmt.Errorf("failed to query photos: %v", err)
	}

	// Check for photos without thumbnails
	var missingThumbnails int
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT p.id)
		FROM photos p
		LEFT JOIN thumbnails t ON p.id = t.photo_id
		WHERE t.photo_id IS NULL
	`).Scan(&missingThumbnails)
	if err != nil {
		return fmt.Errorf("failed to check thumbnails: %v", err)
	}

	// Check for orphaned thumbnails
	var orphanedThumbnails int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM thumbnails t
		LEFT JOIN photos p ON t.photo_id = p.id
		WHERE p.id IS NULL
	`).Scan(&orphanedThumbnails)
	if err != nil {
		return fmt.Errorf("failed to check orphaned thumbnails: %v", err)
	}

	// Display results
	fmt.Println("\nVerification Results:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total photos: %d\n", photoCount)
	fmt.Printf("Photos without thumbnails: %d\n", missingThumbnails)
	fmt.Printf("Orphaned thumbnails: %d\n", orphanedThumbnails)

	if missingThumbnails == 0 && orphanedThumbnails == 0 {
		fmt.Println("\n✓ Database is healthy")
		return nil
	} else {
		fmt.Println("\n⚠ Database has issues")
		return fmt.Errorf("database verification found %d issues", missingThumbnails+orphanedThumbnails)
	}
}
