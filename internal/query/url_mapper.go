package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// URLMapper handles conversion between URLs and QueryParams
type URLMapper struct{}

// NewURLMapper creates a new URL mapper
func NewURLMapper() *URLMapper {
	return &URLMapper{}
}

// ParsePath converts a URL path to QueryParams
// Supports patterns like:
//
//	/2025/10/04          - year/month/day
//	/2025/10             - year/month
//	/2025                - year
//	/camera/Canon/EOS-R5 - camera make/model
//	/lens/Canon-RF-24-70 - lens
//	/color/blue          - colour search
//	/morning             - time of day
//	/bursts              - photos in bursts
func (m *URLMapper) ParsePath(path string, queryString string) (QueryParams, error) {
	params := QueryParams{
		Limit: 50, // default
	}

	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	if path == "" || path == "photos" {
		// Parse query string for filters
		if queryString != "" {
			values, err := url.ParseQuery(queryString)
			if err == nil {
				m.parseQueryString(values, &params)
			}
		}
		return params, nil
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	// Parse based on first segment
	switch segments[0] {
	case "camera":
		if len(segments) >= 3 {
			params.CameraMake = []string{segments[1]}
			params.CameraModel = []string{strings.ReplaceAll(segments[2], "-", " ")}
		} else if len(segments) == 2 {
			params.CameraMake = []string{segments[1]}
		}

	case "lens":
		if len(segments) >= 2 {
			params.LensModel = []string{strings.ReplaceAll(segments[1], "-", " ")}
		}

	case "color":
		if len(segments) >= 2 {
			params.ColourName = []string{segments[1]}
		}

	case "bursts":
		inBurst := true
		params.InBurst = &inBurst

	case "morning", "afternoon", "evening", "night", "blue_hour", "golden_hour_morning", "golden_hour_evening", "midday":
		params.TimeOfDay = []string{segments[0]}

	case "spring", "summer", "fall", "winter":
		params.Season = []string{segments[0]}

	case "wide", "normal", "telephoto":
		params.FocalCategory = []string{segments[0]}

	default:
		// Try to parse as year/month/day
		if year, err := strconv.Atoi(segments[0]); err == nil && year >= 1900 && year <= 2100 {
			params.Year = &year

			if len(segments) >= 2 {
				if month, err := strconv.Atoi(segments[1]); err == nil && month >= 1 && month <= 12 {
					params.Month = &month

					if len(segments) >= 3 {
						if day, err := strconv.Atoi(segments[2]); err == nil && day >= 1 && day <= 31 {
							params.Day = &day
						}
					}
				}
			}
		}
	}

	// Parse query string for additional filters
	if queryString != "" {
		values, err := url.ParseQuery(queryString)
		if err == nil {
			m.parseQueryString(values, &params)
		}
	}

	return params, nil
}

// parseQueryString parses URL query parameters into QueryParams
func (m *URLMapper) parseQueryString(values url.Values, params *QueryParams) {
	// Temporal filters
	if year := values.Get("year"); year != "" {
		if year == "unknown" {
			// Use -1 as a marker for "unknown" year (photos without dates)
			unknownYear := -1
			params.Year = &unknownYear
		} else if y, err := strconv.Atoi(year); err == nil && y >= 1900 && y <= 2100 {
			params.Year = &y
		}
	}
	if month := values.Get("month"); month != "" {
		if m, err := strconv.Atoi(month); err == nil && m >= 1 && m <= 12 {
			params.Month = &m
		}
	}
	if day := values.Get("day"); day != "" {
		if d, err := strconv.Atoi(day); err == nil && d >= 1 && d <= 31 {
			params.Day = &d
		}
	}

	// Pagination
	if limit := values.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}
	if offset := values.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			params.Offset = o
		}
	}

	// Sorting
	if sortBy := values.Get("sort"); sortBy != "" {
		params.SortBy = sortBy
	}
	if order := values.Get("order"); order != "" {
		params.SortOrder = order
	}

	// Equipment filters
	if make := values["camera_make"]; len(make) > 0 {
		params.CameraMake = append(params.CameraMake, make...)
	}
	if model := values["camera_model"]; len(model) > 0 {
		params.CameraModel = append(params.CameraModel, model...)
	}
	if lens := values["lens"]; len(lens) > 0 {
		params.LensModel = append(params.LensModel, lens...)
	}

	// Time filters
	if tod := values["time_of_day"]; len(tod) > 0 {
		params.TimeOfDay = append(params.TimeOfDay, tod...)
	}
	if season := values["season"]; len(season) > 0 {
		params.Season = append(params.Season, season...)
	}

	// Technical filters
	if isoMin := values.Get("iso_min"); isoMin != "" {
		if v, err := strconv.Atoi(isoMin); err == nil {
			params.ISOMin = &v
		}
	}
	if isoMax := values.Get("iso_max"); isoMax != "" {
		if v, err := strconv.Atoi(isoMax); err == nil {
			params.ISOMax = &v
		}
	}
	if apMin := values.Get("aperture_min"); apMin != "" {
		if v, err := strconv.ParseFloat(apMin, 64); err == nil {
			params.ApertureMin = &v
		}
	}
	if apMax := values.Get("aperture_max"); apMax != "" {
		if v, err := strconv.ParseFloat(apMax, 64); err == nil {
			params.ApertureMax = &v
		}
	}
	if flMin := values.Get("focal_min"); flMin != "" {
		if v, err := strconv.ParseFloat(flMin, 64); err == nil {
			params.FocalLengthMin = &v
		}
	}
	if flMax := values.Get("focal_max"); flMax != "" {
		if v, err := strconv.ParseFloat(flMax, 64); err == nil {
			params.FocalLengthMax = &v
		}
	}

	// Categorical filters
	if fc := values["focal_category"]; len(fc) > 0 {
		params.FocalCategory = append(params.FocalCategory, fc...)
	}
	if sc := values["shooting_condition"]; len(sc) > 0 {
		params.ShootingCondition = append(params.ShootingCondition, sc...)
	}

	// Colour filters
	if color := values["color"]; len(color) > 0 {
		params.ColourName = append(params.ColourName, color...)
	}

	// Burst filter
	if burst := values.Get("in_burst"); burst != "" {
		if burst == "true" || burst == "1" {
			inBurst := true
			params.InBurst = &inBurst
		} else if burst == "false" || burst == "0" {
			inBurst := false
			params.InBurst = &inBurst
		}
	}

	// GPS filter
	if hasGPS := values.Get("has_gps"); hasGPS != "" {
		if hasGPS == "true" || hasGPS == "1" {
			gps := true
			params.HasGPS = &gps
		} else if hasGPS == "false" || hasGPS == "0" {
			gps := false
			params.HasGPS = &gps
		}
	}
}

// BuildPath converts QueryParams to a URL path
// Always returns /photos - all filtering is done via query parameters
func (m *URLMapper) BuildPath(params QueryParams) string {
	return "/photos"
}

// BuildQueryString converts QueryParams to URL query parameters
// All filters are included in query string
func (m *URLMapper) BuildQueryString(params QueryParams) string {
	values := url.Values{}

	// Temporal filters
	if params.Year != nil {
		if *params.Year == -1 {
			values.Set("year", "unknown")
		} else {
			values.Set("year", strconv.Itoa(*params.Year))
		}
	}
	if params.Month != nil {
		values.Set("month", strconv.Itoa(*params.Month))
	}
	if params.Day != nil {
		values.Set("day", strconv.Itoa(*params.Day))
	}

	// Camera filters
	for _, m := range params.CameraMake {
		values.Add("camera_make", m)
	}
	for _, m := range params.CameraModel {
		values.Add("camera_model", m)
	}

	// Lens filters
	for _, l := range params.LensModel {
		values.Add("lens", l)
	}

	// Colour filters
	for _, c := range params.ColourName {
		values.Add("color", c)
	}

	// Time of day filters
	for _, t := range params.TimeOfDay {
		values.Add("time_of_day", t)
	}

	// Season filters
	for _, s := range params.Season {
		values.Add("season", s)
	}

	// Focal category filters
	for _, f := range params.FocalCategory {
		values.Add("focal_category", f)
	}

	// Shooting condition filters
	for _, sc := range params.ShootingCondition {
		values.Add("shooting_condition", sc)
	}

	// Burst filter
	if params.InBurst != nil {
		values.Set("in_burst", strconv.FormatBool(*params.InBurst))
	}

	// Pagination
	if params.Limit != 50 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		values.Set("offset", strconv.Itoa(params.Offset))
	}

	// Sorting
	if params.SortBy != "" && params.SortBy != "date_taken" {
		values.Set("sort", params.SortBy)
	}
	if params.SortOrder != "" && params.SortOrder != "desc" {
		values.Set("order", params.SortOrder)
	}

	// Technical ranges
	if params.ISOMin != nil {
		values.Set("iso_min", strconv.Itoa(*params.ISOMin))
	}
	if params.ISOMax != nil {
		values.Set("iso_max", strconv.Itoa(*params.ISOMax))
	}
	if params.ApertureMin != nil {
		values.Set("aperture_min", fmt.Sprintf("%.1f", *params.ApertureMin))
	}
	if params.ApertureMax != nil {
		values.Set("aperture_max", fmt.Sprintf("%.1f", *params.ApertureMax))
	}
	if params.FocalLengthMin != nil {
		values.Set("focal_min", fmt.Sprintf("%.0f", *params.FocalLengthMin))
	}
	if params.FocalLengthMax != nil {
		values.Set("focal_max", fmt.Sprintf("%.0f", *params.FocalLengthMax))
	}

	// GPS filter
	if params.HasGPS != nil {
		values.Set("has_gps", strconv.FormatBool(*params.HasGPS))
	}

	if len(values) == 0 {
		return ""
	}

	return "?" + values.Encode()
}

// BuildFullURL builds complete URL from QueryParams
func (m *URLMapper) BuildFullURL(params QueryParams) string {
	path := m.BuildPath(params)
	query := m.BuildQueryString(params)
	return path + query
}

// Breadcrumb represents a navigation breadcrumb
type Breadcrumb struct {
	Label string
	URL   string
}

// BuildBreadcrumbs generates breadcrumb trail from QueryParams
func (m *URLMapper) BuildBreadcrumbs(params QueryParams) []Breadcrumb {
	crumbs := []Breadcrumb{
		{Label: "Home", URL: "/"},
	}

	// Temporal breadcrumbs - ✅ State machine model: filters are independent
	// Show breadcrumbs for each active temporal filter regardless of others

	if params.Year != nil {
		crumbs = append(crumbs, Breadcrumb{
			Label: strconv.Itoa(*params.Year),
			URL:   m.BuildFullURL(QueryParams{Year: params.Year, Limit: params.Limit}),
		})
	}

	if params.Month != nil {
		monthNames := []string{"", "January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}
		// Month breadcrumb works with or without Year
		crumbs = append(crumbs, Breadcrumb{
			Label: monthNames[*params.Month],
			URL:   m.BuildFullURL(QueryParams{Year: params.Year, Month: params.Month, Limit: params.Limit}),
		})
	}

	if params.Day != nil {
		// Day breadcrumb works with or without Month/Year
		crumbs = append(crumbs, Breadcrumb{
			Label: fmt.Sprintf("Day %d", *params.Day),
			URL:   m.BuildFullURL(QueryParams{Year: params.Year, Month: params.Month, Day: params.Day, Limit: params.Limit}),
		})
	}

	// Camera breadcrumbs
	if len(params.CameraMake) > 0 {
		if params.Year == nil {
			crumbs = append(crumbs, Breadcrumb{
				Label: params.CameraMake[0],
				URL:   fmt.Sprintf("/camera/%s", strings.ReplaceAll(params.CameraMake[0], " ", "-")),
			})
		}

		if len(params.CameraModel) > 0 {
			crumbs = append(crumbs, Breadcrumb{
				Label: params.CameraModel[0],
				URL:   m.BuildPath(params),
			})
		}
	}

	// Other single breadcrumbs
	if len(params.ColourName) > 0 && params.Year == nil && len(params.CameraMake) == 0 {
		crumbs = append(crumbs, Breadcrumb{
			Label: strings.Title(params.ColourName[0]),
			URL:   fmt.Sprintf("/color/%s", params.ColourName[0]),
		})
	}

	if len(params.TimeOfDay) > 0 && params.Year == nil && len(params.CameraMake) == 0 {
		crumbs = append(crumbs, Breadcrumb{
			Label: strings.Title(params.TimeOfDay[0]),
			URL:   fmt.Sprintf("/%s", params.TimeOfDay[0]),
		})
	}

	return crumbs
}
