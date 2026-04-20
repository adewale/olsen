package explorer

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/pkg/models"
)

// Repository provides query methods for the explorer
type Repository struct {
	db *database.DB
}

// NewRepository creates a new repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Stats contains homepage statistics
type Stats struct {
	TotalPhotos   int
	CameraCount   int
	LensCount     int
	DateRangeFrom time.Time
	DateRangeTo   time.Time
	BurstCount    int
}

// PhotoCard represents a photo in grid view
type PhotoCard struct {
	ID          int
	DateTaken   time.Time
	CameraMake  string
	CameraModel string
	IndexedAt   time.Time // Used for cache busting in thumbnail URLs
}

// PhotoDetail represents full photo details
type PhotoDetail struct {
	ID              int
	Thumbnail       []byte
	ThumbnailBase64 string
	DateTaken       time.Time
	CameraMake      string
	CameraModel     string
	LensModel       string
	ISO             int
	Aperture        float64
	ShutterSpeed    string
	FocalLength     float64
	FocalLength35mm int
	FilePath        string
	FileHash        string
	FileSize        int64
	FileSizeMB      string
	Width           int
	Height          int
	Latitude        float64
	Longitude       float64
	DominantColours []models.DominantColour

	// Navigation
	PrevID int
	NextID int
}

// YearInfo represents a year with photo count
type YearInfo struct {
	Year  int
	Count int
}

// CameraMakeInfo represents a camera make with models
type CameraMakeInfo struct {
	Make       string
	TotalCount int
	Models     []CameraModelInfo
}

// CameraModelInfo represents a camera model with count
type CameraModelInfo struct {
	Model string
	Count int
}

// LensInfo represents a lens with count
type LensInfo struct {
	Model string
	Count int
}

// GetStats returns homepage statistics
func (r *Repository) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Total photos
	err := r.db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&stats.TotalPhotos)
	if err != nil {
		return nil, err
	}

	// Camera count
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT camera_make || ' ' || camera_model)
		FROM photos
		WHERE camera_make != '' AND camera_model != ''
	`).Scan(&stats.CameraCount)
	if err != nil {
		return nil, err
	}

	// Lens count
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT lens_model)
		FROM photos
		WHERE lens_model != ''
	`).Scan(&stats.LensCount)
	if err != nil {
		return nil, err
	}

	// Date range
	var minDate, maxDate sql.NullString
	err = r.db.QueryRow(`
		SELECT MIN(date_taken), MAX(date_taken)
		FROM photos
		WHERE date_taken IS NOT NULL
	`).Scan(&minDate, &maxDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if minDate.Valid {
		stats.DateRangeFrom, _ = time.Parse(time.RFC3339, minDate.String)
	}
	if maxDate.Valid {
		stats.DateRangeTo, _ = time.Parse(time.RFC3339, maxDate.String)
	}

	// Burst count
	err = r.db.QueryRow("SELECT COUNT(*) FROM burst_groups").Scan(&stats.BurstCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return stats, nil
}

// GetRecentPhotos returns the most recent photos
func (r *Repository) GetRecentPhotos(limit int) ([]PhotoCard, error) {
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model, indexed_at
		FROM photos
		WHERE date_taken IS NOT NULL
		ORDER BY date_taken DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		var cameraMake sql.NullString
		var cameraModel sql.NullString
		var indexedAt sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &cameraMake, &cameraModel, &indexedAt)
		if err != nil {
			return nil, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}
		if cameraMake.Valid {
			p.CameraMake = cameraMake.String
		}
		if cameraModel.Valid {
			p.CameraModel = cameraModel.String
		}
		if indexedAt.Valid {
			p.IndexedAt, _ = time.Parse(time.RFC3339, indexedAt.String)
		}

		photos = append(photos, p)
	}

	return photos, nil
}

// GetPhotoByID returns detailed photo information
func (r *Repository) GetPhotoByID(id int) (*PhotoDetail, error) {
	photo := &PhotoDetail{}

	var dateTaken sql.NullString
	var cameraMake, cameraModel, lensModel, shutterSpeed, fileHash sql.NullString
	var iso, width, height sql.NullInt64
	var aperture, focalLength, focalLength35mm sql.NullFloat64
	var latitude, longitude sql.NullFloat64
	var fileSize int64

	err := r.db.QueryRow(`
		SELECT id, date_taken, camera_make, camera_model, lens_model,
		       iso, aperture, shutter_speed, focal_length, focal_length_35mm,
		       file_path, file_hash, file_size, width, height,
		       latitude, longitude
		FROM photos
		WHERE id = ?
	`, id).Scan(
		&photo.ID, &dateTaken, &cameraMake, &cameraModel, &lensModel,
		&iso, &aperture, &shutterSpeed, &focalLength, &focalLength35mm,
		&photo.FilePath, &fileHash, &fileSize, &width, &height,
		&latitude, &longitude,
	)
	if err != nil {
		return nil, err
	}

	// Populate string fields
	if dateTaken.Valid {
		photo.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
	}
	if cameraMake.Valid {
		photo.CameraMake = cameraMake.String
	}
	if cameraModel.Valid {
		photo.CameraModel = cameraModel.String
	}
	if lensModel.Valid {
		photo.LensModel = lensModel.String
	}
	if shutterSpeed.Valid {
		photo.ShutterSpeed = shutterSpeed.String
	}
	if fileHash.Valid {
		photo.FileHash = fileHash.String
	}

	// Populate numeric fields
	if iso.Valid {
		photo.ISO = int(iso.Int64)
	}
	if aperture.Valid {
		photo.Aperture = aperture.Float64
	}
	if focalLength.Valid {
		photo.FocalLength = focalLength.Float64
	}
	if focalLength35mm.Valid {
		photo.FocalLength35mm = int(focalLength35mm.Float64)
	}
	if width.Valid {
		photo.Width = int(width.Int64)
	}
	if height.Valid {
		photo.Height = int(height.Int64)
	}
	if latitude.Valid {
		photo.Latitude = latitude.Float64
	}
	if longitude.Valid {
		photo.Longitude = longitude.Float64
	}

	photo.FileSize = fileSize

	// Get 1024px thumbnail
	thumbnail, err := r.GetThumbnail(id, "1024")
	if err == nil {
		photo.Thumbnail = thumbnail
		photo.ThumbnailBase64 = base64.StdEncoding.EncodeToString(thumbnail)
	}

	// Get dominant colors
	colorRows, err := r.db.Query(`
		SELECT red, green, blue, hue, saturation, lightness, weight
		FROM photo_colors
		WHERE photo_id = ?
		ORDER BY color_order
		LIMIT 5
	`, id)
	if err == nil {
		defer colorRows.Close()
		for colorRows.Next() {
			var dc models.DominantColour
			colorRows.Scan(
				&dc.Colour.R, &dc.Colour.G, &dc.Colour.B,
				&dc.HSL.H, &dc.HSL.S, &dc.HSL.L,
				&dc.Weight,
			)
			photo.DominantColours = append(photo.DominantColours, dc)
		}
	}

	// Get prev/next photo IDs
	r.db.QueryRow(`
		SELECT id FROM photos
		WHERE date_taken < (SELECT date_taken FROM photos WHERE id = ?)
		ORDER BY date_taken DESC
		LIMIT 1
	`, id).Scan(&photo.PrevID)

	r.db.QueryRow(`
		SELECT id FROM photos
		WHERE date_taken > (SELECT date_taken FROM photos WHERE id = ?)
		ORDER BY date_taken ASC
		LIMIT 1
	`, id).Scan(&photo.NextID)

	// Format file size
	photo.FileSizeMB = fmt.Sprintf("%.1f", float64(photo.FileSize)/(1024*1024))

	return photo, nil
}

// GetThumbnail returns thumbnail data for a photo
// If the requested size doesn't exist, it falls back to the next smaller size
func (r *Repository) GetThumbnail(photoID int, size string) ([]byte, error) {
	// Define size fallback order: try requested size, then smaller sizes
	var sizePriority []string
	switch size {
	case "1024":
		sizePriority = []string{"1024", "512", "256", "64"}
	case "512":
		sizePriority = []string{"512", "256", "64"}
	case "256":
		sizePriority = []string{"256", "64"}
	case "64":
		sizePriority = []string{"64"}
	default:
		// Unknown size, try it anyway
		sizePriority = []string{size}
	}

	var data []byte
	var lastErr error

	// Try each size in priority order
	for _, trySize := range sizePriority {
		err := r.db.QueryRow(`
			SELECT data FROM thumbnails
			WHERE photo_id = ? AND size = ?
		`, photoID, trySize).Scan(&data)

		if err == nil {
			// Found a thumbnail!
			return data, nil
		}
		lastErr = err
	}

	// No thumbnail found at any size
	return nil, lastErr
}

// GetThumbnailWithTimestamp returns thumbnail data and indexed_at timestamp for a photo
// If the requested size doesn't exist, it falls back to the next smaller size
func (r *Repository) GetThumbnailWithTimestamp(photoID int, size string) ([]byte, time.Time, error) {
	// Define size fallback order: try requested size, then smaller sizes
	var sizePriority []string
	switch size {
	case "1024":
		sizePriority = []string{"1024", "512", "256", "64"}
	case "512":
		sizePriority = []string{"512", "256", "64"}
	case "256":
		sizePriority = []string{"256", "64"}
	case "64":
		sizePriority = []string{"64"}
	default:
		// Unknown size, try it anyway
		sizePriority = []string{size}
	}

	var data []byte
	var indexedAt sql.NullString
	var lastErr error

	// Try each size in priority order
	for _, trySize := range sizePriority {
		err := r.db.QueryRow(`
			SELECT t.data, p.indexed_at
			FROM thumbnails t
			JOIN photos p ON t.photo_id = p.id
			WHERE t.photo_id = ? AND t.size = ?
		`, photoID, trySize).Scan(&data, &indexedAt)

		if err == nil {
			// Found a thumbnail!
			var timestamp time.Time
			if indexedAt.Valid {
				timestamp, _ = time.Parse(time.RFC3339, indexedAt.String)
			}
			return data, timestamp, nil
		}
		lastErr = err
	}

	// No thumbnail found at any size
	return nil, time.Time{}, lastErr
}

// GetPhotosByYear returns photos from a specific year
func (r *Repository) GetPhotosByYear(year int, limit, offset int) ([]PhotoCard, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE strftime('%Y', date_taken) = ?
	`, fmt.Sprintf("%04d", year)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get photos
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model
		FROM photos
		WHERE strftime('%Y', date_taken) = ?
		ORDER BY date_taken DESC
		LIMIT ? OFFSET ?
	`, fmt.Sprintf("%04d", year), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &p.CameraMake, &p.CameraModel)
		if err != nil {
			return nil, 0, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}

		photos = append(photos, p)
	}

	return photos, total, nil
}

// GetPhotosByMonth returns photos from a specific month
func (r *Repository) GetPhotosByMonth(year, month int, limit, offset int) ([]PhotoCard, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE strftime('%Y-%m', date_taken) = ?
	`, fmt.Sprintf("%04d-%02d", year, month)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get photos
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model
		FROM photos
		WHERE strftime('%Y-%m', date_taken) = ?
		ORDER BY date_taken DESC
		LIMIT ? OFFSET ?
	`, fmt.Sprintf("%04d-%02d", year, month), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &p.CameraMake, &p.CameraModel)
		if err != nil {
			return nil, 0, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}

		photos = append(photos, p)
	}

	return photos, total, nil
}

// GetPhotosByDay returns photos from a specific day
func (r *Repository) GetPhotosByDay(year, month, day int, limit, offset int) ([]PhotoCard, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE strftime('%Y-%m-%d', date_taken) = ?
	`, fmt.Sprintf("%04d-%02d-%02d", year, month, day)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get photos
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model
		FROM photos
		WHERE strftime('%Y-%m-%d', date_taken) = ?
		ORDER BY date_taken DESC
		LIMIT ? OFFSET ?
	`, fmt.Sprintf("%04d-%02d-%02d", year, month, day), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &p.CameraMake, &p.CameraModel)
		if err != nil {
			return nil, 0, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}

		photos = append(photos, p)
	}

	return photos, total, nil
}

// GetYears returns all years with photo counts
func (r *Repository) GetYears() ([]YearInfo, error) {
	rows, err := r.db.Query(`
		SELECT strftime('%Y', date_taken) as year, COUNT(*) as count
		FROM photos
		WHERE date_taken IS NOT NULL
		GROUP BY year
		ORDER BY year DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []YearInfo
	for rows.Next() {
		var y YearInfo
		var yearStr string
		err := rows.Scan(&yearStr, &y.Count)
		if err != nil {
			return nil, err
		}
		fmt.Sscanf(yearStr, "%d", &y.Year)
		years = append(years, y)
	}

	return years, nil
}

// GetCameras returns all camera makes with models
func (r *Repository) GetCameras() ([]CameraMakeInfo, error) {
	// Get all makes
	makeRows, err := r.db.Query(`
		SELECT camera_make, COUNT(*) as count
		FROM photos
		WHERE camera_make != ''
		GROUP BY camera_make
		ORDER BY camera_make
	`)
	if err != nil {
		return nil, err
	}
	defer makeRows.Close()

	var makes []CameraMakeInfo
	for makeRows.Next() {
		var make CameraMakeInfo
		makeRows.Scan(&make.Make, &make.TotalCount)

		// Get models for this make
		modelRows, err := r.db.Query(`
			SELECT camera_model, COUNT(*) as count
			FROM photos
			WHERE camera_make = ? AND camera_model != ''
			GROUP BY camera_model
			ORDER BY count DESC
		`, make.Make)
		if err != nil {
			continue
		}

		for modelRows.Next() {
			var model CameraModelInfo
			modelRows.Scan(&model.Model, &model.Count)
			make.Models = append(make.Models, model)
		}
		modelRows.Close()

		makes = append(makes, make)
	}

	return makes, nil
}

// GetLenses returns all lenses with counts
func (r *Repository) GetLenses() ([]LensInfo, error) {
	rows, err := r.db.Query(`
		SELECT lens_model, COUNT(*) as count
		FROM photos
		WHERE lens_model != ''
		GROUP BY lens_model
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lenses []LensInfo
	for rows.Next() {
		var lens LensInfo
		rows.Scan(&lens.Model, &lens.Count)
		lenses = append(lenses, lens)
	}

	return lenses, nil
}

// GetPhotosByCamera returns photos for a specific camera
func (r *Repository) GetPhotosByCamera(make, model string, limit, offset int) ([]PhotoCard, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE camera_make = ? AND camera_model = ?
	`, make, model).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get photos
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model
		FROM photos
		WHERE camera_make = ? AND camera_model = ?
		ORDER BY date_taken DESC
		LIMIT ? OFFSET ?
	`, make, model, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &p.CameraMake, &p.CameraModel)
		if err != nil {
			return nil, 0, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}

		photos = append(photos, p)
	}

	return photos, total, nil
}

// GetPhotosByLens returns photos for a specific lens
func (r *Repository) GetPhotosByLens(lens string, limit, offset int) ([]PhotoCard, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE lens_model = ?
	`, lens).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get photos
	rows, err := r.db.Query(`
		SELECT id, date_taken, camera_make, camera_model
		FROM photos
		WHERE lens_model = ?
		ORDER BY date_taken DESC
		LIMIT ? OFFSET ?
	`, lens, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var photos []PhotoCard
	for rows.Next() {
		var p PhotoCard
		var dateTaken sql.NullString
		err := rows.Scan(&p.ID, &dateTaken, &p.CameraMake, &p.CameraModel)
		if err != nil {
			return nil, 0, err
		}

		if dateTaken.Valid {
			p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}

		photos = append(photos, p)
	}

	return photos, total, nil
}

// TimelineGroup represents a month/year group of photos for the timeline view
type TimelineGroup struct {
	Year   int
	Month  int
	Label  string
	Count  int
	Photos []PhotoCard
}

// GetTimelineGroups returns photos grouped by year/month
func (r *Repository) GetTimelineGroups() ([]TimelineGroup, error) {
	rows, err := r.db.Query(`
		SELECT
			CAST(strftime('%Y', date_taken) AS INTEGER) as year,
			CAST(strftime('%m', date_taken) AS INTEGER) as month,
			COUNT(*) as count
		FROM photos
		WHERE date_taken IS NOT NULL
		GROUP BY year, month
		ORDER BY year DESC, month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	var groups []TimelineGroup
	for rows.Next() {
		var g TimelineGroup
		if err := rows.Scan(&g.Year, &g.Month, &g.Count); err != nil {
			return nil, err
		}
		if g.Month >= 1 && g.Month <= 12 {
			g.Label = fmt.Sprintf("%s %d", monthNames[g.Month], g.Year)
		}
		groups = append(groups, g)
	}

	for i := range groups {
		photoRows, err := r.db.Query(`
			SELECT id, date_taken, camera_make, camera_model, indexed_at
			FROM photos
			WHERE strftime('%Y-%m', date_taken) = ?
			ORDER BY date_taken DESC
			LIMIT 8
		`, fmt.Sprintf("%04d-%02d", groups[i].Year, groups[i].Month))
		if err != nil {
			continue
		}
		for photoRows.Next() {
			var p PhotoCard
			var dateTaken, cameraMake, cameraModel, indexedAt sql.NullString
			photoRows.Scan(&p.ID, &dateTaken, &cameraMake, &cameraModel, &indexedAt)
			if dateTaken.Valid {
				p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
			}
			if cameraMake.Valid {
				p.CameraMake = cameraMake.String
			}
			if cameraModel.Valid {
				p.CameraModel = cameraModel.String
			}
			if indexedAt.Valid {
				p.IndexedAt, _ = time.Parse(time.RFC3339, indexedAt.String)
			}
			groups[i].Photos = append(groups[i].Photos, p)
		}
		photoRows.Close()
	}

	return groups, nil
}

// PaletteGroup represents a color group of photos for the palette view
type PaletteGroup struct {
	ColorName string
	Label     string
	HexColor  string
	Count     int
	Photos    []PhotoCard
}

// GetPaletteGroups returns photos grouped by dominant color
func (r *Repository) GetPaletteGroups() ([]PaletteGroup, error) {
	colorOrder := []string{"red", "orange", "yellow", "green", "blue", "purple", "pink", "brown", "white", "gray", "black", "bw"}
	colorLabels := map[string]string{
		"red": "Red", "orange": "Orange", "yellow": "Yellow", "green": "Green",
		"blue": "Blue", "purple": "Purple", "pink": "Pink", "brown": "Brown",
		"white": "White", "gray": "Gray", "black": "Black", "bw": "B&W",
	}
	colorHex := map[string]string{
		"red": "#e74c3c", "orange": "#e67e22", "yellow": "#f1c40f", "green": "#27ae60",
		"blue": "#3498db", "purple": "#9b59b6", "pink": "#e91e63", "brown": "#8B4513",
		"white": "#ffffff", "gray": "#888888", "black": "#111111",
	}

	countRows, err := r.db.Query(`
		SELECT colour_name, COUNT(DISTINCT photo_id) as count
		FROM photo_colors
		WHERE colour_name != ''
		GROUP BY colour_name
	`)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()

	countByColor := map[string]int{}
	for countRows.Next() {
		var name string
		var count int
		countRows.Scan(&name, &count)
		countByColor[name] = count
	}

	var groups []PaletteGroup
	for _, name := range colorOrder {
		count := countByColor[name]
		if count == 0 {
			continue
		}
		hex := colorHex[name]
		if hex == "" {
			hex = "#888888"
		}
		g := PaletteGroup{
			ColorName: name,
			Label:     colorLabels[name],
			HexColor:  hex,
			Count:     count,
		}

		photoRows, err := r.db.Query(`
			SELECT DISTINCT p.id, p.date_taken, p.camera_make, p.camera_model, p.indexed_at
			FROM photos p
			JOIN photo_colors pc ON pc.photo_id = p.id
			WHERE pc.colour_name = ?
			ORDER BY p.date_taken DESC
			LIMIT 8
		`, name)
		if err == nil {
			for photoRows.Next() {
				var p PhotoCard
				var dateTaken, cameraMake, cameraModel, indexedAt sql.NullString
				photoRows.Scan(&p.ID, &dateTaken, &cameraMake, &cameraModel, &indexedAt)
				if dateTaken.Valid {
					p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
				}
				if cameraMake.Valid {
					p.CameraMake = cameraMake.String
				}
				if cameraModel.Valid {
					p.CameraModel = cameraModel.String
				}
				if indexedAt.Valid {
					p.IndexedAt, _ = time.Parse(time.RFC3339, indexedAt.String)
				}
				g.Photos = append(g.Photos, p)
			}
			photoRows.Close()
		}

		groups = append(groups, g)
	}

	return groups, nil
}

// GPSPhoto represents a photo with GPS coordinates for the map view
type GPSPhoto struct {
	ID        int     `json:"id"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	DateTaken string  `json:"date"`
	Camera    string  `json:"camera"`
}

// GetGPSPhotos returns all photos with GPS coordinates
func (r *Repository) GetGPSPhotos() ([]GPSPhoto, error) {
	rows, err := r.db.Query(`
		SELECT id, latitude, longitude, date_taken, camera_make, camera_model
		FROM photos
		WHERE latitude IS NOT NULL AND longitude IS NOT NULL
		  AND latitude != 0 AND longitude != 0
		ORDER BY date_taken DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []GPSPhoto
	for rows.Next() {
		var p GPSPhoto
		var dateTaken, cameraMake, cameraModel sql.NullString
		rows.Scan(&p.ID, &p.Latitude, &p.Longitude, &dateTaken, &cameraMake, &cameraModel)
		if dateTaken.Valid {
			p.DateTaken = dateTaken.String
		}
		camera := ""
		if cameraMake.Valid {
			camera = cameraMake.String
		}
		if cameraModel.Valid && cameraModel.String != "" {
			if camera != "" {
				camera += " "
			}
			camera += cameraModel.String
		}
		p.Camera = camera
		photos = append(photos, p)
	}

	return photos, nil
}

// GetGPSPhotoCount returns the count of GPS-tagged photos
func (r *Repository) GetGPSPhotoCount() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM photos
		WHERE latitude IS NOT NULL AND longitude IS NOT NULL
		  AND latitude != 0 AND longitude != 0
	`).Scan(&count)
	return count, err
}

// BurstGroup represents a group of photos taken in burst mode
type BurstGroup struct {
	GroupID          string
	RepresentativeID int
	PhotoCount       int
	DateTaken        time.Time
	CameraMake       string
	CameraModel      string
	Photos           []PhotoCard
}

// GetBurstGroups returns all burst groups with their photos
func (r *Repository) GetBurstGroups() ([]BurstGroup, error) {
	rows, err := r.db.Query(`
		SELECT bg.group_id, bg.representative_id, bg.photo_count,
		       p.date_taken, p.camera_make, p.camera_model
		FROM burst_groups bg
		JOIN photos p ON p.id = bg.representative_id
		ORDER BY p.date_taken DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []BurstGroup
	for rows.Next() {
		var g BurstGroup
		var dateTaken, cameraMake, cameraModel sql.NullString
		rows.Scan(&g.GroupID, &g.RepresentativeID, &g.PhotoCount, &dateTaken, &cameraMake, &cameraModel)
		if dateTaken.Valid {
			g.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
		}
		if cameraMake.Valid {
			g.CameraMake = cameraMake.String
		}
		if cameraModel.Valid {
			g.CameraModel = cameraModel.String
		}
		groups = append(groups, g)
	}

	for i := range groups {
		photoRows, err := r.db.Query(`
			SELECT p.id, p.date_taken, p.camera_make, p.camera_model, p.indexed_at
			FROM photos p
			JOIN burst_group_members bgm ON bgm.photo_id = p.id
			WHERE bgm.group_id = ?
			ORDER BY p.date_taken ASC
		`, groups[i].GroupID)
		if err != nil {
			continue
		}
		for photoRows.Next() {
			var p PhotoCard
			var dateTaken, cameraMake, cameraModel, indexedAt sql.NullString
			photoRows.Scan(&p.ID, &dateTaken, &cameraMake, &cameraModel, &indexedAt)
			if dateTaken.Valid {
				p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
			}
			if cameraMake.Valid {
				p.CameraMake = cameraMake.String
			}
			if cameraModel.Valid {
				p.CameraModel = cameraModel.String
			}
			if indexedAt.Valid {
				p.IndexedAt, _ = time.Parse(time.RFC3339, indexedAt.String)
			}
			groups[i].Photos = append(groups[i].Photos, p)
		}
		photoRows.Close()
	}

	return groups, nil
}
