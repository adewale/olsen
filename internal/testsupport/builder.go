// Package testsupport provides test data builders and insertion helpers
// shared by tests across packages.
//
// The builder follows the test-data-builder pattern: sensible defaults for
// every field so tests state only what matters to their assertions. It
// deliberately avoids importing internal/database (tests inside that package
// could not use it otherwise); insertion goes through the small PhotoInserter
// interface, which *database.DB satisfies, or through raw SQL for tests that
// hold a bare *sql.DB.
package testsupport

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adewale/olsen/pkg/models"
)

// photoCounter makes generated file paths and hashes unique per process.
var photoCounter atomic.Int64

// PhotoBuilder builds models.PhotoMetadata values with sensible defaults.
type PhotoBuilder struct {
	photo models.PhotoMetadata
}

// NewPhoto returns a builder for a valid, fully-populated photo. Every photo
// gets a unique file path and hash; override only the fields the test cares
// about.
func NewPhoto() *PhotoBuilder {
	n := photoCounter.Add(1)
	return &PhotoBuilder{photo: models.PhotoMetadata{
		FilePath:     fmt.Sprintf("/test/photo_%06d.jpg", n),
		FileHash:     fmt.Sprintf("hash_%06d", n),
		FileSize:     1_000_000,
		LastModified: time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
		CameraMake:   "Canon",
		CameraModel:  "EOS R5",
		LensModel:    "RF24-70mm F2.8 L IS USM",
		ISO:          400,
		Aperture:     2.8,
		ShutterSpeed: "1/250",
		FocalLength:  50,
		DateTaken:    time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
		Width:        6000,
		Height:       4000,
	}}
}

// WithPath sets the file path (and derives a matching hash).
func (b *PhotoBuilder) WithPath(path string) *PhotoBuilder {
	b.photo.FilePath = path
	return b
}

// WithHash sets the file hash.
func (b *PhotoBuilder) WithHash(hash string) *PhotoBuilder {
	b.photo.FileHash = hash
	return b
}

// WithCamera sets camera make and model.
func (b *PhotoBuilder) WithCamera(make, model string) *PhotoBuilder {
	b.photo.CameraMake = make
	b.photo.CameraModel = model
	return b
}

// WithLens sets the lens model.
func (b *PhotoBuilder) WithLens(model string) *PhotoBuilder {
	b.photo.LensModel = model
	return b
}

// WithDateTaken sets the capture time.
func (b *PhotoBuilder) WithDateTaken(t time.Time) *PhotoBuilder {
	b.photo.DateTaken = t
	return b
}

// WithoutDate clears the capture time (photos with unknown dates).
func (b *PhotoBuilder) WithoutDate() *PhotoBuilder {
	b.photo.DateTaken = time.Time{}
	return b
}

// WithISO sets the ISO value.
func (b *PhotoBuilder) WithISO(iso int) *PhotoBuilder {
	b.photo.ISO = iso
	return b
}

// WithAperture sets the aperture.
func (b *PhotoBuilder) WithAperture(f float64) *PhotoBuilder {
	b.photo.Aperture = f
	return b
}

// WithFocalLength sets the focal length in mm.
func (b *PhotoBuilder) WithFocalLength(mm float64) *PhotoBuilder {
	b.photo.FocalLength = mm
	return b
}

// WithGPS sets latitude and longitude.
func (b *PhotoBuilder) WithGPS(lat, lon float64) *PhotoBuilder {
	b.photo.Latitude = lat
	b.photo.Longitude = lon
	return b
}

// WithTimeOfDay sets the inferred time-of-day facet value.
func (b *PhotoBuilder) WithTimeOfDay(tod string) *PhotoBuilder {
	b.photo.TimeOfDay = tod
	return b
}

// WithSeason sets the inferred season facet value.
func (b *PhotoBuilder) WithSeason(season string) *PhotoBuilder {
	b.photo.Season = season
	return b
}

// WithFocalCategory sets the inferred focal-length category facet value.
func (b *PhotoBuilder) WithFocalCategory(fc string) *PhotoBuilder {
	b.photo.FocalCategory = fc
	return b
}

// WithShootingCondition sets the inferred shooting-condition facet value.
func (b *PhotoBuilder) WithShootingCondition(sc string) *PhotoBuilder {
	b.photo.ShootingCondition = sc
	return b
}

// WithDominantColour appends a dominant colour by HSL with a representative
// RGB value and the given weight.
func (b *PhotoBuilder) WithDominantColour(h, s, l int, weight float64) *PhotoBuilder {
	b.photo.DominantColours = append(b.photo.DominantColours, models.DominantColour{
		Colour: models.Colour{R: 128, G: 128, B: 128},
		HSL:    models.ColourHSL{H: h, S: s, L: l},
		Weight: weight,
	})
	return b
}

// ColourHSL returns a representative HSL value that the query engine's
// classification (saturation-first, then hue) assigns to the given colour
// name. Unknown names panic so test typos fail loud.
func ColourHSL(name string) (h, s, l int) {
	switch name {
	case "red":
		return 5, 80, 50
	case "orange":
		return 30, 80, 60
	case "yellow":
		return 60, 80, 50
	case "green":
		return 120, 80, 50
	case "blue":
		return 210, 80, 50
	case "purple":
		return 270, 80, 50
	case "pink":
		return 320, 80, 50
	case "brown":
		return 30, 60, 30 // orange hue, low lightness
	case "gray", "grey":
		return 0, 5, 50 // saturation < 10
	case "bw":
		return 0, 12, 50 // saturation 10-14: near-grayscale
	case "black":
		return 0, 2, 10 // saturation < 5, lightness < 20
	case "white":
		return 0, 2, 90 // saturation < 5, lightness > 80
	default:
		panic(fmt.Sprintf("testsupport.ColourHSL: unknown colour %q", name))
	}
}

// WithColourName appends a dominant colour whose HSL is representative of the
// named colour (see ColourHSL).
func (b *PhotoBuilder) WithColourName(name string) *PhotoBuilder {
	h, s, l := ColourHSL(name)
	return b.WithDominantColour(h, s, l, 1.0)
}

// WithThumbnail adds a thumbnail of the given size with placeholder bytes.
func (b *PhotoBuilder) WithThumbnail(size models.ThumbnailSize) *PhotoBuilder {
	if b.photo.Thumbnails == nil {
		b.photo.Thumbnails = map[models.ThumbnailSize][]byte{}
	}
	b.photo.Thumbnails[size] = []byte("thumb_" + string(size))
	return b
}

// Build returns the constructed photo.
func (b *PhotoBuilder) Build() *models.PhotoMetadata {
	p := b.photo // copy so the builder can be reused
	return &p
}

// PhotoInserter is the subset of *database.DB the helpers need. Declared here
// instead of importing internal/database so that package database's own tests
// can use testsupport without an import cycle.
type PhotoInserter interface {
	InsertPhoto(photo *models.PhotoMetadata) error
}

// InsertPhotos inserts the given photos, failing the test on error.
func InsertPhotos(t testing.TB, db PhotoInserter, photos ...*models.PhotoMetadata) {
	t.Helper()
	for _, p := range photos {
		if err := db.InsertPhoto(p); err != nil {
			t.Fatalf("InsertPhotos: failed to insert %s: %v", p.FilePath, err)
		}
	}
}

// InsertPhotoSQL inserts a photo (and its dominant colours) via raw SQL for
// tests that hold a bare *sql.DB with the olsen schema loaded.
func InsertPhotoSQL(t testing.TB, db *sql.DB, p *models.PhotoMetadata) int64 {
	t.Helper()

	var dateTaken interface{}
	if !p.DateTaken.IsZero() {
		dateTaken = p.DateTaken.Format("2006-01-02 15:04:05")
	}
	nullable := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}
	res, err := db.Exec(`
		INSERT INTO photos (file_path, file_hash, file_size, last_modified,
		                    camera_make, camera_model, lens_model, date_taken,
		                    iso, aperture, focal_length,
		                    time_of_day, season, focal_category, shooting_condition,
		                    latitude, longitude)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.FilePath, p.FileHash, p.FileSize, p.LastModified.Format("2006-01-02 15:04:05"),
		nullable(p.CameraMake), nullable(p.CameraModel), nullable(p.LensModel), dateTaken,
		p.ISO, p.Aperture, p.FocalLength,
		nullable(p.TimeOfDay), nullable(p.Season), nullable(p.FocalCategory), nullable(p.ShootingCondition),
		p.Latitude, p.Longitude,
	)
	if err != nil {
		t.Fatalf("InsertPhotoSQL: failed to insert %s: %v", p.FilePath, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("InsertPhotoSQL: LastInsertId: %v", err)
	}

	for i, c := range p.DominantColours {
		if _, err := db.Exec(`
			INSERT INTO photo_colors (photo_id, color_order, red, green, blue, weight, hue, saturation, lightness)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, i, c.Colour.R, c.Colour.G, c.Colour.B, c.Weight, c.HSL.H, c.HSL.S, c.HSL.L,
		); err != nil {
			t.Fatalf("InsertPhotoSQL: failed to insert colour %d for %s: %v", i, p.FilePath, err)
		}
	}
	return id
}
