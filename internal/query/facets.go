package query

import (
	"fmt"
	"strings"
)

// ComputeFacets calculates facet counts based on current query parameters
// Facets respect active filters but exclude their own dimension
func (e *Engine) ComputeFacets(params QueryParams) (*FacetCollection, error) {
	facets := &FacetCollection{}

	// Compute each facet dimension
	var err error

	facets.Camera, err = e.computeCameraFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute camera facet: %w", err)
	}

	facets.Lens, err = e.computeLensFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute lens facet: %w", err)
	}

	facets.Year, err = e.computeYearFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute year facet: %w", err)
	}

	facets.Month, err = e.computeMonthFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute month facet: %w", err)
	}

	facets.TimeOfDay, err = e.computeTimeOfDayFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute time of day facet: %w", err)
	}

	facets.Season, err = e.computeSeasonFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute season facet: %w", err)
	}

	facets.FocalCategory, err = e.computeFocalCategoryFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute focal category facet: %w", err)
	}

	facets.ShootingCondition, err = e.computeShootingConditionFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shooting condition facet: %w", err)
	}

	facets.InBurst, err = e.computeBurstFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute burst facet: %w", err)
	}

	facets.ColourName, err = e.computeColourFacet(params)
	if err != nil {
		return nil, fmt.Errorf("failed to compute colour facet: %w", err)
	}

	// Add URLs to all facet values
	builder := NewFacetURLBuilder(NewURLMapper())
	builder.BuildURLsForFacets(facets, params)

	return facets, nil
}

// computeCameraFacet computes camera make/model facet
func (e *Engine) computeCameraFacet(params QueryParams) (*Facet, error) {
	// Exclude camera filters from WHERE clause
	paramsWithoutCamera := params
	paramsWithoutCamera.CameraMake = nil
	paramsWithoutCamera.CameraModel = nil

	where, args := e.buildWhereClause(paramsWithoutCamera)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Add filter for non-NULL cameras
	additionalWhere := "camera_make IS NOT NULL AND camera_model IS NOT NULL"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT camera_make || ' ' || camera_model as camera, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY camera_make, camera_model
		ORDER BY count DESC
		LIMIT 50
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var camera string
		var count int
		if err := rows.Scan(&camera, &count); err != nil {
			return nil, err
		}

		// Check if selected
		selected := false
		for _, make := range params.CameraMake {
			for _, model := range params.CameraModel {
				if camera == make+" "+model {
					selected = true
					break
				}
			}
		}

		values = append(values, FacetValue{
			Value:    camera,
			Label:    camera,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "camera",
		Label:  "Camera",
		Values: values,
	}, nil
}

// computeLensFacet computes lens facet
func (e *Engine) computeLensFacet(params QueryParams) (*Facet, error) {
	paramsWithoutLens := params
	paramsWithoutLens.LensMake = nil
	paramsWithoutLens.LensModel = nil

	where, args := e.buildWhereClause(paramsWithoutLens)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	additionalWhere := "lens_model IS NOT NULL AND lens_model != ''"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT lens_model, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY lens_model
		ORDER BY count DESC
		LIMIT 30
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var lens string
		var count int
		if err := rows.Scan(&lens, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, l := range params.LensModel {
			if lens == l {
				selected = true
				break
			}
		}

		values = append(values, FacetValue{
			Value:    lens,
			Label:    lens,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "lens",
		Label:  "Lens",
		Values: values,
	}, nil
}

// computeYearFacet computes year facet
func (e *Engine) computeYearFacet(params QueryParams) (*Facet, error) {
	paramsWithoutYear := params
	paramsWithoutYear.Year = nil
	// ✅ State machine model: PRESERVE Month and Day filters
	// Month and Day should NOT be cleared - they're independent dimensions
	// The count shown should reflect: "How many photos in this year with current filters?"
	paramsWithoutYear.DateFrom = nil
	paramsWithoutYear.DateTo = nil

	where, args := e.buildWhereClause(paramsWithoutYear)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Query for years (including Unknown for NULL dates)
	query := fmt.Sprintf(`
		SELECT
			COALESCE(strftime('%%Y', date_taken), 'unknown') as year,
			COUNT(*) as count
		FROM photos p
		%s
		GROUP BY year
		ORDER BY
			CASE
				WHEN year = 'unknown' THEN 1
				ELSE 0
			END,
			year DESC
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var year string
		var count int
		if err := rows.Scan(&year, &count); err != nil {
			return nil, err
		}

		selected := false
		label := year

		if year == "unknown" {
			// Check if filtering by unknown year (Year = nil or negative value as marker)
			if params.Year != nil && *params.Year == -1 {
				selected = true
			}
			label = "Unknown"
		} else {
			// Check if this year is selected
			if params.Year != nil && fmt.Sprintf("%04d", *params.Year) == year {
				selected = true
			}
		}

		values = append(values, FacetValue{
			Value:    year,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "year",
		Label:  "Year",
		Values: values,
	}, nil
}

// computeMonthFacet computes month facet (only shown when year is selected)
func (e *Engine) computeMonthFacet(params QueryParams) (*Facet, error) {
	// Only show month facet if year is selected
	if params.Year == nil {
		return &Facet{
			Name:   "month",
			Label:  "Month",
			Values: []FacetValue{},
		}, nil
	}

	paramsWithoutMonth := params
	paramsWithoutMonth.Month = nil
	// ✅ State machine model: PRESERVE Day filter
	// Day should NOT be cleared - it's an independent dimension
	// The count shown should reflect: "How many photos in this month with current filters?"

	where, args := e.buildWhereClause(paramsWithoutMonth)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Query for months within the selected year
	query := fmt.Sprintf(`
		SELECT
			strftime('%%m', date_taken) as month,
			COUNT(*) as count
		FROM photos p
		%s
		GROUP BY month
		ORDER BY month ASC
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monthNames := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	values := []FacetValue{}
	for rows.Next() {
		var monthStr string
		var count int
		if err := rows.Scan(&monthStr, &count); err != nil {
			return nil, err
		}

		// Parse month number (01-12)
		var monthNum int
		if _, err := fmt.Sscanf(monthStr, "%d", &monthNum); err != nil || monthNum < 1 || monthNum > 12 {
			continue
		}

		selected := false
		if params.Month != nil && *params.Month == monthNum {
			selected = true
		}

		values = append(values, FacetValue{
			Value:    monthStr,
			Label:    monthNames[monthNum],
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "month",
		Label:  "Month",
		Values: values,
	}, nil
}

// computeTimeOfDayFacet computes time of day facet
func (e *Engine) computeTimeOfDayFacet(params QueryParams) (*Facet, error) {
	paramsWithoutTOD := params
	paramsWithoutTOD.TimeOfDay = nil

	where, args := e.buildWhereClause(paramsWithoutTOD)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	additionalWhere := "time_of_day IS NOT NULL AND time_of_day != ''"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT time_of_day, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY time_of_day
		ORDER BY
			CASE time_of_day
				WHEN 'morning' THEN 1
				WHEN 'afternoon' THEN 2
				WHEN 'evening' THEN 3
				WHEN 'night' THEN 4
			END
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var tod string
		var count int
		if err := rows.Scan(&tod, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, t := range params.TimeOfDay {
			if tod == t {
				selected = true
				break
			}
		}

		// Capitalize label
		label := strings.Title(tod)

		values = append(values, FacetValue{
			Value:    tod,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "time_of_day",
		Label:  "Time of Day",
		Values: values,
	}, nil
}

// computeSeasonFacet computes season facet
func (e *Engine) computeSeasonFacet(params QueryParams) (*Facet, error) {
	paramsWithoutSeason := params
	paramsWithoutSeason.Season = nil

	where, args := e.buildWhereClause(paramsWithoutSeason)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	additionalWhere := "season IS NOT NULL AND season != ''"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT season, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY season
		ORDER BY
			CASE season
				WHEN 'spring' THEN 1
				WHEN 'summer' THEN 2
				WHEN 'fall' THEN 3
				WHEN 'winter' THEN 4
			END
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var season string
		var count int
		if err := rows.Scan(&season, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, s := range params.Season {
			if season == s {
				selected = true
				break
			}
		}

		// Capitalize label
		label := strings.Title(season)

		values = append(values, FacetValue{
			Value:    season,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "season",
		Label:  "Season",
		Values: values,
	}, nil
}

// computeFocalCategoryFacet computes focal length category facet
func (e *Engine) computeFocalCategoryFacet(params QueryParams) (*Facet, error) {
	paramsWithoutFC := params
	paramsWithoutFC.FocalCategory = nil

	where, args := e.buildWhereClause(paramsWithoutFC)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	additionalWhere := "focal_category IS NOT NULL AND focal_category != ''"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT focal_category, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY focal_category
		ORDER BY
			CASE focal_category
				WHEN 'wide' THEN 1
				WHEN 'normal' THEN 2
				WHEN 'telephoto' THEN 3
			END
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var fc string
		var count int
		if err := rows.Scan(&fc, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, f := range params.FocalCategory {
			if fc == f {
				selected = true
				break
			}
		}

		// Capitalize label
		label := strings.Title(fc)

		values = append(values, FacetValue{
			Value:    fc,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "focal_category",
		Label:  "Focal Length",
		Values: values,
	}, nil
}

// computeShootingConditionFacet computes shooting condition facet
func (e *Engine) computeShootingConditionFacet(params QueryParams) (*Facet, error) {
	paramsWithoutSC := params
	paramsWithoutSC.ShootingCondition = nil

	where, args := e.buildWhereClause(paramsWithoutSC)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	additionalWhere := "shooting_condition IS NOT NULL AND shooting_condition != ''"
	if whereClause != "" {
		whereClause += " AND " + additionalWhere
	} else {
		whereClause = "WHERE " + additionalWhere
	}

	query := fmt.Sprintf(`
		SELECT shooting_condition, COUNT(*) as count
		FROM photos p
		%s
		GROUP BY shooting_condition
		ORDER BY
			CASE shooting_condition
				WHEN 'bright' THEN 1
				WHEN 'normal' THEN 2
				WHEN 'low_light' THEN 3
			END
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var sc string
		var count int
		if err := rows.Scan(&sc, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, s := range params.ShootingCondition {
			if sc == s {
				selected = true
				break
			}
		}

		// Format label
		label := strings.ReplaceAll(strings.Title(sc), "_", " ")

		values = append(values, FacetValue{
			Value:    sc,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "shooting_condition",
		Label:  "Lighting",
		Values: values,
	}, nil
}

// computeBurstFacet computes burst facet
func (e *Engine) computeBurstFacet(params QueryParams) (*Facet, error) {
	paramsWithoutBurst := params
	paramsWithoutBurst.InBurst = nil

	where, args := e.buildWhereClause(paramsWithoutBurst)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count photos in bursts vs not in bursts
	query := fmt.Sprintf(`
		SELECT
			CASE WHEN burst_group_id IS NOT NULL THEN 'yes' ELSE 'no' END as in_burst,
			COUNT(*) as count
		FROM photos p
		%s
		GROUP BY in_burst
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var inBurst string
		var count int
		if err := rows.Scan(&inBurst, &count); err != nil {
			return nil, err
		}

		selected := false
		if params.InBurst != nil {
			if inBurst == "yes" && *params.InBurst {
				selected = true
			} else if inBurst == "no" && !*params.InBurst {
				selected = true
			}
		}

		label := "Not in Burst"
		if inBurst == "yes" {
			label = "In Burst"
		}

		values = append(values, FacetValue{
			Value:    inBurst,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "in_burst",
		Label:  "Burst",
		Values: values,
	}, nil
}

// computeColourFacet computes colour name facet
func (e *Engine) computeColourFacet(params QueryParams) (*Facet, error) {
	paramsWithoutColour := params
	paramsWithoutColour.ColourName = nil

	where, args := e.buildWhereClause(paramsWithoutColour)
	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count photos by dominant colour
	// Classification based on Berlin-Kay universal color terms
	// Priority: Check saturation first (for achromatic colors), then hue ranges
	query := fmt.Sprintf(`
		SELECT
			CASE
				-- Achromatic colors (based on saturation)
				WHEN pc.saturation < 5 AND pc.lightness < 20 THEN 'black'
				WHEN pc.saturation < 5 AND pc.lightness > 80 THEN 'white'
				WHEN pc.saturation < 10 THEN 'gray'
				WHEN pc.saturation < 15 THEN 'bw'

				-- Chromatic colors (based on hue ranges)
				-- Brown: orange hue + low lightness
				WHEN pc.hue BETWEEN 20 AND 40 AND pc.lightness < 50 THEN 'brown'
				WHEN pc.hue BETWEEN 0 AND 15 OR pc.hue BETWEEN 345 AND 360 THEN 'red'
				WHEN pc.hue BETWEEN 16 AND 45 THEN 'orange'
				WHEN pc.hue BETWEEN 46 AND 75 THEN 'yellow'
				WHEN pc.hue BETWEEN 76 AND 165 THEN 'green'
				WHEN pc.hue BETWEEN 166 AND 255 THEN 'blue'
				WHEN pc.hue BETWEEN 256 AND 290 THEN 'purple'
				WHEN pc.hue BETWEEN 291 AND 344 THEN 'pink'
				ELSE 'other'
			END as colour_name,
			COUNT(DISTINCT p.id) as count
		FROM photos p
		JOIN photo_colors pc ON pc.photo_id = p.id
		%s
		GROUP BY colour_name
		ORDER BY count DESC
	`, whereClause)

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []FacetValue{}
	for rows.Next() {
		var colourName string
		var count int
		if err := rows.Scan(&colourName, &count); err != nil {
			return nil, err
		}

		selected := false
		for _, c := range params.ColourName {
			if colourName == c {
				selected = true
				break
			}
		}

		label := strings.Title(colourName)

		values = append(values, FacetValue{
			Value:    colourName,
			Label:    label,
			Count:    count,
			Selected: selected,
		})
	}

	return &Facet{
		Name:   "color",
		Label:  "Colour",
		Values: values,
	}, nil
}
