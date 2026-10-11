package query

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/testsupport"
	"github.com/adewale/olsen/pkg/models"
)

// facetFixturePhotos is a small, fixed photo set covering every facet that the
// facet state-machine, lifecycle and engine tests walk: several colours,
// years, cameras (make and model, split on the first space) and times of day,
// with overlapping combinations so that single, dual and triple facet states
// are all non-empty. The tests used to read a developer's local test.db or
// test_query.db and skipped in CI, where neither exists.
//
// Invariants the tests rely on (keep them when editing):
//   - blue is the most common colour, and blue+2024+Canon EOS R5 has 2 photos
//   - 2024 is the most common year; 2023, 2024 and 2025 are all present
//   - more than 5 photos in total (pagination), 4 photos from 2025
//   - ISO 100-400: 6 photos; morning: 4 photos; blue: 5 photos
func facetFixturePhotos() []*models.PhotoMetadata {
	at := func(year int, month time.Month, day, hour int) time.Time {
		return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	}
	photo := testsupport.NewPhoto
	return []*models.PhotoMetadata{
		photo().WithColourName("blue").WithDateTaken(at(2024, time.March, 10, 8)).WithTimeOfDay("morning").WithISO(100).WithAperture(4).WithFocalLength(35).Build(),
		photo().WithColourName("blue").WithDateTaken(at(2024, time.June, 15, 14)).WithTimeOfDay("afternoon").WithISO(400).Build(),
		photo().WithColourName("blue").WithCamera("Nikon", "Z 9").WithDateTaken(at(2024, time.September, 1, 9)).WithTimeOfDay("morning").WithISO(800).WithFocalLength(200).Build(),
		photo().WithColourName("blue").WithDateTaken(at(2025, time.January, 20, 7)).WithTimeOfDay("morning").WithISO(200).WithAperture(5.6).WithFocalLength(70).Build(),
		photo().WithColourName("blue").WithCamera("Sony", "ILCE-7RM5").WithDateTaken(at(2025, time.May, 2, 18)).WithTimeOfDay("evening").WithISO(3200).WithAperture(1.4).Build(),
		photo().WithColourName("red").WithDateTaken(at(2024, time.December, 24, 15)).WithTimeOfDay("afternoon").WithISO(400).Build(),
		photo().WithColourName("red").WithCamera("Nikon", "Z 9").WithDateTaken(at(2023, time.July, 4, 21)).WithTimeOfDay("night").WithISO(6400).Build(),
		photo().WithColourName("green").WithCamera("Sony", "ILCE-7RM5").WithDateTaken(at(2025, time.August, 30, 10)).WithTimeOfDay("morning").WithISO(160).WithFocalLength(24).Build(),
		photo().WithColourName("green").WithDateTaken(at(2023, time.April, 11, 13)).WithTimeOfDay("afternoon").WithISO(250).Build(),
		photo().WithColourName("yellow").WithCamera("Nikon", "Z 9").WithDateTaken(at(2025, time.October, 5, 17)).WithTimeOfDay("golden_hour_evening").WithISO(1600).Build(),
		photo().WithColourName("red").WithColourName("green").WithDateTaken(at(2024, time.August, 8, 12)).WithTimeOfDay("midday").WithISO(1000).Build(),
	}
}

// setupTestDB returns a database in t.TempDir() with the olsen schema,
// seeded with facetFixturePhotos. It never reads files outside the test's
// temporary directory, so results do not depend on local leftovers.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "facets.db"))
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	if _, err := db.Exec(database.Schema); err != nil {
		db.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}
	for _, p := range facetFixturePhotos() {
		testsupport.InsertPhotoSQL(t, db, p)
	}
	return db
}

// openFixtureDatabase returns a *database.DB in t.TempDir() seeded with
// facetFixturePhotos through the production insert path.
func openFixtureDatabase(t testing.TB) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "query.db"))
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	testsupport.InsertPhotos(t, db, facetFixturePhotos()...)
	return db
}
