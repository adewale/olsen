// Package query implements the faceted search query engine with state machine navigation.
//
// It provides SQL-based photo queries with support for filtering by temporal, equipment,
// visual, and technical dimensions. The engine implements a state machine model where users
// navigate through valid filter combinations, ensuring no transitions result in zero photos.
// All filters are preserved during transitions unless explicitly removed.
package query

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Engine handles query execution
type Engine struct {
	db *sql.DB
}

// NewEngine creates a new query engine
func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// Query executes a query with the given parameters
func (e *Engine) Query(params QueryParams) (*QueryResult, error) {
	startTime := time.Now()

	// Set defaults
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.SortBy == "" {
		params.SortBy = "date_taken"
		params.SortOrder = "desc"
	}

	// Build query
	query, args := e.buildQuery(params)

	// Execute count query
	countQuery := e.buildCountQuery(params)
	var total int
	err := e.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count results: %w", err)
	}

	// Execute main query
	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse results
	photos := []PhotoSummary{}
	for rows.Next() {
		photo, err := e.scanPhotoSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, photo)
	}

	queryTime := time.Since(startTime).Milliseconds()

	return &QueryResult{
		Photos:      photos,
		Total:       total,
		Limit:       params.Limit,
		Offset:      params.Offset,
		HasMore:     params.Offset+params.Limit < total,
		QueryTimeMs: queryTime,
	}, nil
}

// buildQuery constructs the SQL query from parameters
func (e *Engine) buildQuery(params QueryParams) (string, []interface{}) {
	var where []string
	var args []interface{}

	// Build WHERE clauses
	where, args = e.buildWhereClause(params)

	// Build ORDER BY
	orderBy := e.buildOrderBy(params)

	// Construct full query
	query := `
		SELECT
			p.id, p.file_path, p.date_taken,
			p.camera_make, p.camera_model, p.lens_model,
			p.iso, p.aperture, p.shutter_speed, p.focal_length, p.focal_length_35mm,
			p.width, p.height,
			p.time_of_day, p.season, p.focal_category,
			p.burst_group_id, p.is_burst_representative,
			p.latitude, p.longitude,
			p.indexed_at
		FROM photos p
	`

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " " + orderBy
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, params.Offset)

	return query, args
}

// buildCountQuery constructs the count query
func (e *Engine) buildCountQuery(params QueryParams) string {
	where, _ := e.buildWhereClause(params)

	query := "SELECT COUNT(*) FROM photos p"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	return query
}

// buildWhereClause builds WHERE conditions and arguments
func (e *Engine) buildWhereClause(params QueryParams) ([]string, []interface{}) {
	var where []string
	var args []interface{}

	// Temporal filters
	if params.Year != nil {
		if *params.Year == -1 {
			// Filter for photos without dates (unknown year)
			where = append(where, "p.date_taken IS NULL")
		} else {
			where = append(where, "strftime('%Y', p.date_taken) = ?")
			args = append(args, fmt.Sprintf("%04d", *params.Year))
		}
	}
	if params.Month != nil {
		// ✅ State machine model: Month is independent of Year
		// Apply month filter even when Year is not set (for facet computation)
		where = append(where, "strftime('%m', p.date_taken) = ?")
		args = append(args, fmt.Sprintf("%02d", *params.Month))
	}
	if params.Day != nil {
		// ✅ State machine model: Day is independent of Month and Year
		// Apply day filter even when Month/Year are not set (for facet computation)
		where = append(where, "strftime('%d', p.date_taken) = ?")
		args = append(args, fmt.Sprintf("%02d", *params.Day))
	}
	if params.DateFrom != nil {
		where = append(where, "p.date_taken >= ?")
		args = append(args, params.DateFrom.Format("2006-01-02 15:04:05"))
	}
	if params.DateTo != nil {
		where = append(where, "p.date_taken <= ?")
		args = append(args, params.DateTo.Format("2006-01-02 15:04:05"))
	}
	if len(params.TimeOfDay) > 0 {
		placeholders := make([]string, len(params.TimeOfDay))
		for i, tod := range params.TimeOfDay {
			placeholders[i] = "?"
			args = append(args, tod)
		}
		where = append(where, fmt.Sprintf("p.time_of_day IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.Season) > 0 {
		placeholders := make([]string, len(params.Season))
		for i, season := range params.Season {
			placeholders[i] = "?"
			args = append(args, season)
		}
		where = append(where, fmt.Sprintf("p.season IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Equipment filters
	if len(params.CameraMake) > 0 {
		placeholders := make([]string, len(params.CameraMake))
		for i, make := range params.CameraMake {
			placeholders[i] = "?"
			args = append(args, make)
		}
		where = append(where, fmt.Sprintf("p.camera_make IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.CameraModel) > 0 {
		placeholders := make([]string, len(params.CameraModel))
		for i, model := range params.CameraModel {
			placeholders[i] = "?"
			args = append(args, model)
		}
		where = append(where, fmt.Sprintf("p.camera_model IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.LensMake) > 0 {
		placeholders := make([]string, len(params.LensMake))
		for i, make := range params.LensMake {
			placeholders[i] = "?"
			args = append(args, make)
		}
		where = append(where, fmt.Sprintf("p.lens_make IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.LensModel) > 0 {
		placeholders := make([]string, len(params.LensModel))
		for i, model := range params.LensModel {
			placeholders[i] = "?"
			args = append(args, model)
		}
		where = append(where, fmt.Sprintf("p.lens_model IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Technical range filters
	if params.ISOMin != nil {
		where = append(where, "p.iso >= ?")
		args = append(args, *params.ISOMin)
	}
	if params.ISOMax != nil {
		where = append(where, "p.iso <= ?")
		args = append(args, *params.ISOMax)
	}
	if params.ApertureMin != nil {
		where = append(where, "p.aperture >= ?")
		args = append(args, *params.ApertureMin)
	}
	if params.ApertureMax != nil {
		where = append(where, "p.aperture <= ?")
		args = append(args, *params.ApertureMax)
	}
	if params.FocalLengthMin != nil {
		where = append(where, "p.focal_length >= ?")
		args = append(args, *params.FocalLengthMin)
	}
	if params.FocalLengthMax != nil {
		where = append(where, "p.focal_length <= ?")
		args = append(args, *params.FocalLengthMax)
	}
	if params.FocalLength35mmMin != nil {
		where = append(where, "p.focal_length_35mm >= ?")
		args = append(args, *params.FocalLength35mmMin)
	}
	if params.FocalLength35mmMax != nil {
		where = append(where, "p.focal_length_35mm <= ?")
		args = append(args, *params.FocalLength35mmMax)
	}

	// Categorical filters
	if len(params.FocalCategory) > 0 {
		placeholders := make([]string, len(params.FocalCategory))
		for i, cat := range params.FocalCategory {
			placeholders[i] = "?"
			args = append(args, cat)
		}
		where = append(where, fmt.Sprintf("p.focal_category IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.ShootingCondition) > 0 {
		placeholders := make([]string, len(params.ShootingCondition))
		for i, cond := range params.ShootingCondition {
			placeholders[i] = "?"
			args = append(args, cond)
		}
		where = append(where, fmt.Sprintf("p.shooting_condition IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Location filters
	if params.LatMin != nil {
		where = append(where, "p.latitude >= ?")
		args = append(args, *params.LatMin)
	}
	if params.LatMax != nil {
		where = append(where, "p.latitude <= ?")
		args = append(args, *params.LatMax)
	}
	if params.LonMin != nil {
		where = append(where, "p.longitude >= ?")
		args = append(args, *params.LonMin)
	}
	if params.LonMax != nil {
		where = append(where, "p.longitude <= ?")
		args = append(args, *params.LonMax)
	}
	if params.HasGPS != nil {
		if *params.HasGPS {
			where = append(where, "(p.latitude IS NOT NULL AND p.longitude IS NOT NULL)")
		} else {
			where = append(where, "(p.latitude IS NULL OR p.longitude IS NULL)")
		}
	}

	// Colour filters (requires join with photo_colors table)
	if len(params.ColourName) > 0 {
		colourConditions := []string{}
		for _, colourName := range params.ColourName {
			if hueRange, ok := ColourNameToHueRange[strings.ToLower(colourName)]; ok {
				// Handle red which wraps around (0-15 and 345-360)
				if colourName == "red" {
					colourConditions = append(colourConditions,
						"EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND ((pc.hue >= 0 AND pc.hue <= 15) OR (pc.hue >= 345 AND pc.hue <= 360)))")
				} else if colourName == "grey" || colourName == "black" || colourName == "white" {
					// Special handling for achromatic colours
					if colourName == "grey" {
						colourConditions = append(colourConditions,
							"EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.saturation < 20 AND pc.lightness BETWEEN 20 AND 80)")
					} else if colourName == "black" {
						colourConditions = append(colourConditions,
							"EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.lightness < 20)")
					} else if colourName == "white" {
						colourConditions = append(colourConditions,
							"EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.lightness > 80)")
					}
				} else {
					colourConditions = append(colourConditions,
						fmt.Sprintf("EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.hue BETWEEN %d AND %d)", hueRange[0], hueRange[1]))
				}
			}
		}
		if len(colourConditions) > 0 {
			where = append(where, "("+strings.Join(colourConditions, " OR ")+")")
		}
	}

	// Hue range filter
	if params.HueMin != nil && params.HueMax != nil {
		where = append(where, "EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.hue BETWEEN ? AND ?)")
		args = append(args, *params.HueMin, *params.HueMax)
	}

	// Saturation range
	if params.SatMin != nil {
		where = append(where, "EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.saturation >= ?)")
		args = append(args, *params.SatMin)
	}
	if params.SatMax != nil {
		where = append(where, "EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.saturation <= ?)")
		args = append(args, *params.SatMax)
	}

	// Lightness range
	if params.LightMin != nil {
		where = append(where, "EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.lightness >= ?)")
		args = append(args, *params.LightMin)
	}
	if params.LightMax != nil {
		where = append(where, "EXISTS (SELECT 1 FROM photo_colors pc WHERE pc.photo_id = p.id AND pc.lightness <= ?)")
		args = append(args, *params.LightMax)
	}

	// Burst filters
	if params.InBurst != nil {
		if *params.InBurst {
			where = append(where, "p.burst_group_id IS NOT NULL")
		} else {
			where = append(where, "p.burst_group_id IS NULL")
		}
	}
	if params.BurstGroupID != nil {
		where = append(where, "p.burst_group_id = ?")
		args = append(args, *params.BurstGroupID)
	}
	if params.IsBurstRep != nil {
		where = append(where, "p.is_burst_representative = ?")
		args = append(args, *params.IsBurstRep)
	}

	// Image properties
	if params.WidthMin != nil {
		where = append(where, "p.width >= ?")
		args = append(args, *params.WidthMin)
	}
	if params.WidthMax != nil {
		where = append(where, "p.width <= ?")
		args = append(args, *params.WidthMax)
	}
	if params.HeightMin != nil {
		where = append(where, "p.height >= ?")
		args = append(args, *params.HeightMin)
	}
	if params.HeightMax != nil {
		where = append(where, "p.height <= ?")
		args = append(args, *params.HeightMax)
	}
	if params.Orientation != nil {
		where = append(where, "p.orientation = ?")
		args = append(args, *params.Orientation)
	}
	if params.IsLandscape != nil {
		if *params.IsLandscape {
			where = append(where, "p.width > p.height")
		} else {
			where = append(where, "p.width <= p.height")
		}
	}
	if params.IsPortrait != nil {
		if *params.IsPortrait {
			where = append(where, "p.height > p.width")
		} else {
			where = append(where, "p.height <= p.width")
		}
	}

	// Other filters
	if params.FlashFired != nil {
		where = append(where, "p.flash_fired = ?")
		args = append(args, *params.FlashFired)
	}
	if len(params.WhiteBalance) > 0 {
		placeholders := make([]string, len(params.WhiteBalance))
		for i, wb := range params.WhiteBalance {
			placeholders[i] = "?"
			args = append(args, wb)
		}
		where = append(where, fmt.Sprintf("p.white_balance IN (%s)", strings.Join(placeholders, ", ")))
	}
	if len(params.ColourSpace) > 0 {
		placeholders := make([]string, len(params.ColourSpace))
		for i, cs := range params.ColourSpace {
			placeholders[i] = "?"
			args = append(args, cs)
		}
		where = append(where, fmt.Sprintf("p.color_space IN (%s)", strings.Join(placeholders, ", ")))
	}

	return where, args
}

// buildOrderBy constructs ORDER BY clause
func (e *Engine) buildOrderBy(params QueryParams) string {
	order := "DESC"
	if params.SortOrder == "asc" {
		order = "ASC"
	}

	switch params.SortBy {
	case "date_taken":
		return fmt.Sprintf("ORDER BY p.date_taken %s", order)
	case "camera":
		return fmt.Sprintf("ORDER BY p.camera_make %s, p.camera_model %s", order, order)
	case "focal_length":
		return fmt.Sprintf("ORDER BY p.focal_length %s", order)
	case "iso":
		return fmt.Sprintf("ORDER BY p.iso %s", order)
	case "aperture":
		return fmt.Sprintf("ORDER BY p.aperture %s", order)
	default:
		return "ORDER BY p.date_taken DESC"
	}
}

// scanPhotoSummary scans a row into PhotoSummary
func (e *Engine) scanPhotoSummary(rows *sql.Rows) (PhotoSummary, error) {
	var p PhotoSummary
	var dateTaken sql.NullString
	var cameraMake, cameraModel, lensModel sql.NullString
	var iso, width, height sql.NullInt64
	var aperture, focalLength sql.NullFloat64
	var focalLength35mm sql.NullInt64
	var shutterSpeed sql.NullString
	var timeOfDay, season, focalCategory sql.NullString
	var burstGroupID sql.NullString
	var isBurstRep sql.NullBool
	var latitude, longitude sql.NullFloat64
	var indexedAt sql.NullString

	err := rows.Scan(
		&p.ID, &p.FilePath, &dateTaken,
		&cameraMake, &cameraModel, &lensModel,
		&iso, &aperture, &shutterSpeed, &focalLength, &focalLength35mm,
		&width, &height,
		&timeOfDay, &season, &focalCategory,
		&burstGroupID, &isBurstRep,
		&latitude, &longitude,
		&indexedAt,
	)
	if err != nil {
		return p, err
	}

	// Parse nullable fields
	if dateTaken.Valid {
		p.DateTaken, _ = time.Parse(time.RFC3339, dateTaken.String)
	}
	if cameraMake.Valid {
		p.CameraMake = cameraMake.String
	}
	if cameraModel.Valid {
		p.CameraModel = cameraModel.String
	}
	if lensModel.Valid {
		p.LensModel = lensModel.String
	}
	if iso.Valid {
		p.ISO = int(iso.Int64)
	}
	if aperture.Valid {
		p.Aperture = aperture.Float64
	}
	if shutterSpeed.Valid {
		p.ShutterSpeed = shutterSpeed.String
	}
	if focalLength.Valid {
		p.FocalLength = focalLength.Float64
	}
	if focalLength35mm.Valid {
		p.FocalLength35mm = int(focalLength35mm.Int64)
	}
	if width.Valid {
		p.Width = int(width.Int64)
	}
	if height.Valid {
		p.Height = int(height.Int64)
	}
	if timeOfDay.Valid {
		p.TimeOfDay = timeOfDay.String
	}
	if season.Valid {
		p.Season = season.String
	}
	if focalCategory.Valid {
		p.FocalCategory = focalCategory.String
	}
	if burstGroupID.Valid {
		p.InBurst = true
		p.BurstGroupID = burstGroupID.String
	}
	if isBurstRep.Valid {
		p.IsBurstRep = isBurstRep.Bool
	}
	if latitude.Valid && longitude.Valid {
		p.HasGPS = true
		p.Latitude = latitude.Float64
		p.Longitude = longitude.Float64
	}
	if indexedAt.Valid {
		p.IndexedAt, _ = time.Parse(time.RFC3339, indexedAt.String)
	}

	return p, nil
}
