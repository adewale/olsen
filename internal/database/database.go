// Package database provides SQLite database operations for the Olsen photo indexer.
//
// It implements a portable catalog design where all metadata, thumbnails, and color
// palettes are stored in a single SQLite file with WAL mode enabled for concurrent
// read access. The database schema supports photos, bursts, duplicates, tags, and
// collections.
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/adewale/olsen/pkg/models"
)

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
}

// Open creates a new database connection and initializes the schema
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set performance pragmas
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Insert facet metadata
	if _, err := db.Exec(FacetMetadataInserts); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to insert facet metadata: %w", err)
	}

	return &DB{db}, nil
}

// InsertPhoto inserts a photo and its related data into the database
func (db *DB) InsertPhoto(photo *models.PhotoMetadata) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert photo record
	result, err := tx.Exec(`
		INSERT INTO photos (
			file_path, file_hash, file_size, last_modified,
			camera_make, camera_model, lens_make, lens_model,
			iso, aperture, shutter_speed, exposure_compensation, focal_length, focal_length_35mm,
			date_taken, date_digitized,
			width, height, orientation, color_space,
			latitude, longitude, altitude,
			dng_version, original_raw_filename,
			flash_fired, white_balance, focus_distance,
			time_of_day, season, focal_category, shooting_condition,
			perceptual_hash
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?
		)`,
		photo.FilePath, photo.FileHash, photo.FileSize, photo.LastModified,
		nullString(photo.CameraMake), nullString(photo.CameraModel), nullString(photo.LensMake), nullString(photo.LensModel),
		nullInt(photo.ISO), nullFloat(photo.Aperture), nullString(photo.ShutterSpeed), nullFloat(photo.ExposureCompensation), nullFloat(photo.FocalLength), nullInt(photo.FocalLength35mm),
		nullTime(photo.DateTaken), nullTime(photo.DateDigitized),
		nullInt(photo.Width), nullInt(photo.Height), nullInt(photo.Orientation), nullString(photo.ColourSpace),
		nullFloat(photo.Latitude), nullFloat(photo.Longitude), nullFloat(photo.Altitude),
		nullString(photo.DNGVersion), nullString(photo.OriginalRawFilename),
		photo.FlashFired, nullString(photo.WhiteBalance), nullFloat(photo.FocusDistance),
		nullString(photo.TimeOfDay), nullString(photo.Season), nullString(photo.FocalCategory), nullString(photo.ShootingCondition),
		nullString(photo.PerceptualHash),
	)
	if err != nil {
		return fmt.Errorf("failed to insert photo: %w", err)
	}

	photoID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get photo ID: %w", err)
	}

	// Insert thumbnails
	for size, data := range photo.Thumbnails {
		_, err := tx.Exec(`
			INSERT INTO thumbnails (photo_id, size, data, format, quality)
			VALUES (?, ?, ?, 'jpeg', 85)
		`, photoID, string(size), data)
		if err != nil {
			return fmt.Errorf("failed to insert thumbnail %s: %w", size, err)
		}
	}

	// Insert colours
	for i, colour := range photo.DominantColours {
		_, err := tx.Exec(`
			INSERT INTO photo_colors (photo_id, color_order, red, green, blue, weight, hue, saturation, lightness)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, photoID, i, colour.Colour.R, colour.Colour.G, colour.Colour.B, colour.Weight, colour.HSL.H, colour.HSL.S, colour.HSL.L)
		if err != nil {
			return fmt.Errorf("failed to insert colour %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PhotoExists checks if a photo with the given file path already exists
func (db *DB) PhotoExists(filePath string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM photos WHERE file_path = ?)", filePath).Scan(&exists)
	return exists, err
}

// GetPhotoHash returns the file hash for a photo by file path
func (db *DB) GetPhotoHash(filePath string) (string, error) {
	var hash string
	err := db.QueryRow("SELECT file_hash FROM photos WHERE file_path = ?", filePath).Scan(&hash)
	return hash, err
}

// DeletePhoto deletes a photo and all related data (thumbnails, colors, etc.) by file path
func (db *DB) DeletePhoto(filePath string) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get photo ID first
	var photoID int
	err = tx.QueryRow("SELECT id FROM photos WHERE file_path = ?", filePath).Scan(&photoID)
	if err != nil {
		return fmt.Errorf("failed to get photo ID: %w", err)
	}

	// Delete related records (thumbnails will cascade delete due to FK constraint)
	// Delete color palette entries
	if _, err := tx.Exec("DELETE FROM color_palette WHERE photo_id = ?", photoID); err != nil {
		return fmt.Errorf("failed to delete color palette: %w", err)
	}

	// Delete the photo itself (thumbnails will cascade)
	if _, err := tx.Exec("DELETE FROM photos WHERE id = ?", photoID); err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPhotoCount returns the total number of photos in the database
func (db *DB) GetPhotoCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&count)
	return count, err
}

// Helper functions to handle NULL values
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}

func nullFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
