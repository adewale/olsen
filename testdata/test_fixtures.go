//go:build ignore
// +build ignore

package main

import (
	"log"
	"os"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/indexer"
)

func main() {
	// Create test database
	dbPath := "/tmp/olsen_fixtures_test.db"
	os.Remove(dbPath) // Clean up old test db

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Index the DNG fixtures
	engine := indexer.NewEngine(db, 4)
	if err := engine.IndexDirectory("testdata/dng"); err != nil {
		log.Fatalf("Failed to index: %v", err)
	}

	stats := engine.GetStats()
	log.Printf("\nIndexing complete!")
	log.Printf("  Files found: %d", stats.FilesFound)
	log.Printf("  Files processed: %d", stats.FilesProcessed)
	log.Printf("  Files failed: %d", stats.FilesFailed)
	log.Printf("  Thumbnails: %d", stats.ThumbnailsGenerated)
	log.Printf("  Duration: %v", stats.Duration())
	log.Printf("\nDatabase created: %s", dbPath)
	log.Printf("Run: go run testdata/verify_coverage.go %s", dbPath)
}
