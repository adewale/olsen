// Package explorer provides the web-based photo browser with faceted search.
//
// It implements an HTTP server with embedded HTML templates, state machine-based
// faceted navigation, and dynamic thumbnail serving with ETag caching. The explorer
// provides a read-only view into the photo database with filtering by date, camera,
// color, and other metadata dimensions.
package explorer

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/query"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates *template.Template

func init() {
	templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

// Server represents the HTTP server
type Server struct {
	db        *database.DB
	repo      *Repository
	engine    *query.Engine
	urlMapper *query.URLMapper
	addr      string
	router    *http.ServeMux
}

// NewServer creates a new server instance
func NewServer(db *database.DB, addr string) *Server {
	s := &Server{
		db:        db,
		repo:      NewRepository(db),
		engine:    query.NewEngine(db.DB),
		urlMapper: query.NewURLMapper(),
		addr:      addr,
		router:    http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Photo detail
	s.router.HandleFunc("/photo/", s.handlePhotoDetail)

	// API routes
	s.router.HandleFunc("/api/thumbnail/", s.handleThumbnail)

	// Main photo browsing route - all filtering via query parameters
	s.router.HandleFunc("/photos", s.handleQuery)

	// Legacy browse pages (optional - could redirect to /photos)
	s.router.HandleFunc("/dates", s.handleDates)
	s.router.HandleFunc("/cameras", s.handleCameras)
	s.router.HandleFunc("/lenses", s.handleLenses)

	// Root handler
	s.router.HandleFunc("/", s.handleHome)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting explorer server on http://%s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Clone the template set and add the specific content template as "content"
	tmpl, err := templates.Clone()
	if err != nil {
		log.Printf("Template clone error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the named template and add it as "content"
	contentTmpl := templates.Lookup(name)
	if contentTmpl == nil {
		log.Printf("Template not found: %s", name)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	// Add the content template with the name "content" so layout can find it
	_, err = tmpl.AddParseTree("content", contentTmpl.Tree)
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the layout template
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// Delegate to catch-all handler
		return
	}

	stats, err := s.repo.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	photos, err := s.repo.GetRecentPhotos(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute facets for navigation
	params := query.QueryParams{
		Limit: 100,
	}
	facets, err := s.engine.ComputeFacets(params)
	if err != nil {
		log.Printf("Facet computation error: %v", err)
		facets = nil
	}

	data := map[string]interface{}{
		"Title":  "Home",
		"Stats":  stats,
		"Photos": photos,
		"Facets": facets,
	}

	s.renderTemplate(w, "home", data)
}

func (s *Server) handlePhotoDetail(w http.ResponseWriter, r *http.Request) {
	// Extract photo ID from URL: /photo/:id
	idStr := strings.TrimPrefix(r.URL.Path, "/photo/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := s.repo.GetPhotoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	backLink := r.Referer()
	if backLink == "" {
		backLink = "/"
	}

	data := map[string]interface{}{
		"Title":    "Photo Detail",
		"Photo":    photo,
		"BackLink": backLink,
	}

	s.renderTemplate(w, "detail", data)
}

func (s *Server) handleThumbnail(w http.ResponseWriter, r *http.Request) {
	// Parse: /api/thumbnail/:id/:size
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/thumbnail/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	size := parts[1]
	if size != "64" && size != "256" && size != "512" && size != "1024" {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		return
	}

	thumbnail, indexedAt, err := s.repo.GetThumbnailWithTimestamp(id, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Generate ETag including indexed_at timestamp so it changes when photo is re-indexed
	etag := fmt.Sprintf(`"%d-%s-%d"`, id, size, indexedAt.Unix())

	// Check If-None-Match header for conditional requests
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			// Client has the current version, send 304 Not Modified
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Set cache headers
	// Use a shorter cache time and rely on ETags for efficient caching
	// This prevents stale images when navigating between different filtered views
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
	w.Header().Set("ETag", etag)

	w.Write(thumbnail)
}

func (s *Server) handleDates(w http.ResponseWriter, r *http.Request) {
	years, err := s.repo.GetYears()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "Browse by Date",
		"Years": years,
	}

	s.renderTemplate(w, "years", data)
}

func (s *Server) handleCameras(w http.ResponseWriter, r *http.Request) {
	cameras, err := s.repo.GetCameras()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":       "Browse by Camera",
		"CameraMakes": cameras,
	}

	s.renderTemplate(w, "cameras", data)
}

func (s *Server) handleLenses(w http.ResponseWriter, r *http.Request) {
	lenses, err := s.repo.GetLenses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":  "Browse by Lens",
		"Lenses": lenses,
	}

	s.renderTemplate(w, "lenses", data)
}

func (s *Server) handleDateRoute(w http.ResponseWriter, r *http.Request, parts []string) {
	limit := 100
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
	}
	offset := (page - 1) * limit

	var photos []PhotoCard
	var total int
	var err error
	var title string
	backLink := "/dates"

	switch len(parts) {
	case 1:
		// Year view
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		photos, total, err = s.repo.GetPhotosByYear(year, limit, offset)
		title = fmt.Sprintf("%d", year)
	case 2:
		// Month view
		year, _ := strconv.Atoi(parts[0])
		month, _ := strconv.Atoi(parts[1])
		photos, total, err = s.repo.GetPhotosByMonth(year, month, limit, offset)
		title = fmt.Sprintf("%d/%02d", year, month)
		backLink = fmt.Sprintf("/%d", year)
	case 3:
		// Day view
		year, _ := strconv.Atoi(parts[0])
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[2])
		photos, total, err = s.repo.GetPhotosByDay(year, month, day, limit, offset)
		title = fmt.Sprintf("%d/%02d/%02d", year, month, day)
		backLink = fmt.Sprintf("/%d/%02d", year, month)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination links
	var prevPage, nextPage string
	if page > 1 {
		prevPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page-1)
	}
	if offset+limit < total {
		nextPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page+1)
	}

	data := map[string]interface{}{
		"Title":      title,
		"Photos":     photos,
		"TotalCount": total,
		"Page":       page,
		"PrevPage":   prevPage,
		"NextPage":   nextPage,
		"BackLink":   backLink,
	}

	s.renderTemplate(w, "grid", data)
}

func (s *Server) handleCameraRoute(w http.ResponseWriter, r *http.Request) {
	// Parse: /camera/:make/:model
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/camera/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	make := parts[0]
	model := parts[1]

	limit := 100
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
	}
	offset := (page - 1) * limit

	photos, total, err := s.repo.GetPhotosByCamera(make, model, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination links
	var prevPage, nextPage string
	if page > 1 {
		prevPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page-1)
	}
	if offset+limit < total {
		nextPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page+1)
	}

	data := map[string]interface{}{
		"Title":      fmt.Sprintf("%s %s", make, model),
		"Photos":     photos,
		"TotalCount": total,
		"Page":       page,
		"PrevPage":   prevPage,
		"NextPage":   nextPage,
		"BackLink":   "/cameras",
	}

	s.renderTemplate(w, "grid", data)
}

func (s *Server) handleLensRoute(w http.ResponseWriter, r *http.Request) {
	// Parse: /lens/:model
	lens := strings.TrimPrefix(r.URL.Path, "/lens/")
	if lens == "" {
		http.NotFound(w, r)
		return
	}

	limit := 100
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
	}
	offset := (page - 1) * limit

	photos, total, err := s.repo.GetPhotosByLens(lens, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination links
	var prevPage, nextPage string
	if page > 1 {
		prevPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page-1)
	}
	if offset+limit < total {
		nextPage = fmt.Sprintf("%s?page=%d", r.URL.Path, page+1)
	}

	data := map[string]interface{}{
		"Title":      lens,
		"Photos":     photos,
		"TotalCount": total,
		"Page":       page,
		"PrevPage":   prevPage,
		"NextPage":   nextPage,
		"BackLink":   "/lenses",
	}

	s.renderTemplate(w, "grid", data)
}

// handleQuery handles query-based photo browsing using the query engine
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	// Parse URL path and query string into QueryParams
	params, err := s.urlMapper.ParsePath(r.URL.Path, r.URL.RawQuery)
	if err != nil {
		log.Printf("FACET_404: URL parse failed - path=%s query=%s error=%v", r.URL.Path, r.URL.RawQuery, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		params.Offset = (page - 1) * params.Limit
	}

	// Execute query
	result, err := s.engine.Query(params)
	if err != nil {
		log.Printf("FACET_ERROR: Query execution failed - path=%s params=%+v error=%v",
			r.URL.Path, params, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get facets if requested or if it's a browsing view
	var facets *query.FacetCollection
	if r.URL.Query().Get("facets") != "" || true {
		facets, err = s.engine.ComputeFacets(params)
		if err != nil {
			log.Printf("Facet computation error: %v", err)
			// Don't fail the whole request if facets fail
			facets = nil
		}
	}

	// Log facet state transitions (structured logging for monitoring)
	// This logs all available transitions with their expected result counts
	query.LogTransitionsSummary(params, facets, result.Total)

	// Log when a facet navigation results in no photos (effectively a 404 from user perspective)
	if result.Total == 0 {
		log.Printf("FACET_404: No results found - path=%s query=%s params=%+v",
			r.URL.Path, r.URL.RawQuery, params)
		// Log additional diagnostic information to detect bugs
		query.LogSuspiciousZeroResults(params, facets)
	}

	// Build breadcrumbs
	breadcrumbs := s.urlMapper.BuildBreadcrumbs(params)

	// Build active filters
	activeFilters := s.buildActiveFilters(params)

	// Calculate pagination
	page := (params.Offset / params.Limit) + 1
	var prevPage, nextPage string
	if page > 1 {
		prevParams := params
		prevParams.Offset = (page - 2) * params.Limit
		prevPage = s.urlMapper.BuildFullURL(prevParams)
	}
	if result.HasMore {
		nextParams := params
		nextParams.Offset = page * params.Limit
		nextPage = s.urlMapper.BuildFullURL(nextParams)
	}

	// Build title from params
	title := "Photos"
	if params.Year != nil {
		title = fmt.Sprintf("Photos from %d", *params.Year)
		if params.Month != nil {
			monthNames := []string{"", "January", "February", "March", "April", "May", "June",
				"July", "August", "September", "October", "November", "December"}
			title = fmt.Sprintf("Photos from %s %d", monthNames[*params.Month], *params.Year)
		}
	} else if len(params.CameraMake) > 0 {
		title = params.CameraMake[0]
		if len(params.CameraModel) > 0 {
			title += " " + params.CameraModel[0]
		}
	} else if len(params.ColourName) > 0 {
		title = strings.Title(params.ColourName[0]) + " Photos"
	} else if len(params.TimeOfDay) > 0 {
		title = strings.Title(params.TimeOfDay[0]) + " Photos"
	}

	data := map[string]interface{}{
		"Title":         title,
		"Photos":        result.Photos,
		"TotalCount":    result.Total,
		"Page":          page,
		"PrevPage":      prevPage,
		"NextPage":      nextPage,
		"Facets":        facets,
		"Breadcrumbs":   breadcrumbs,
		"ActiveFilters": activeFilters,
		"BackLink":      "/",
	}

	s.renderTemplate(w, "grid", data)
}

// ActiveFilter represents a currently applied filter
type ActiveFilter struct {
	Type      string // "color", "year", "camera", etc.
	Label     string // "Blue", "2024", "Canon EOS R5"
	RemoveURL string // URL to remove this filter
}

// buildActiveFilters extracts active filters from query params
func (s *Server) buildActiveFilters(params query.QueryParams) []ActiveFilter {
	filters := []ActiveFilter{}

	// Colour filter
	if len(params.ColourName) > 0 {
		for _, colour := range params.ColourName {
			p := params
			p.ColourName = nil
			filters = append(filters, ActiveFilter{
				Type:      "color",
				Label:     strings.Title(colour),
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Year filter
	if params.Year != nil {
		p := params
		p.Year = nil
		// ✅ State machine model: Don't clear Month/Day when removing Year
		// Month and Day are independent dimensions
		filters = append(filters, ActiveFilter{
			Type:      "year",
			Label:     fmt.Sprintf("%d", *params.Year),
			RemoveURL: s.urlMapper.BuildFullURL(p),
		})
	}

	// Month filter
	if params.Month != nil {
		monthNames := []string{"", "January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}
		p := params
		p.Month = nil
		// ✅ State machine model: Don't clear Day when removing Month
		filters = append(filters, ActiveFilter{
			Type:      "month",
			Label:     monthNames[*params.Month],
			RemoveURL: s.urlMapper.BuildFullURL(p),
		})
	}

	// Day filter
	if params.Day != nil {
		p := params
		p.Day = nil
		filters = append(filters, ActiveFilter{
			Type:      "day",
			Label:     fmt.Sprintf("Day %d", *params.Day),
			RemoveURL: s.urlMapper.BuildFullURL(p),
		})
	}

	// Camera filter
	if len(params.CameraMake) > 0 {
		label := params.CameraMake[0]
		if len(params.CameraModel) > 0 {
			label += " " + params.CameraModel[0]
		}
		p := params
		p.CameraMake = nil
		p.CameraModel = nil
		filters = append(filters, ActiveFilter{
			Type:      "camera",
			Label:     label,
			RemoveURL: s.urlMapper.BuildFullURL(p),
		})
	}

	// Lens filter
	if len(params.LensModel) > 0 {
		for _, lens := range params.LensModel {
			p := params
			p.LensModel = nil
			filters = append(filters, ActiveFilter{
				Type:      "lens",
				Label:     lens,
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Time of Day filters
	if len(params.TimeOfDay) > 0 {
		for _, tod := range params.TimeOfDay {
			p := params
			p.TimeOfDay = removeStringFromSlice(p.TimeOfDay, tod)
			filters = append(filters, ActiveFilter{
				Type:      "time_of_day",
				Label:     strings.Title(tod),
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Season filters
	if len(params.Season) > 0 {
		for _, season := range params.Season {
			p := params
			p.Season = removeStringFromSlice(p.Season, season)
			filters = append(filters, ActiveFilter{
				Type:      "season",
				Label:     strings.Title(season),
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Focal Category filters
	if len(params.FocalCategory) > 0 {
		for _, fc := range params.FocalCategory {
			p := params
			p.FocalCategory = removeStringFromSlice(p.FocalCategory, fc)
			filters = append(filters, ActiveFilter{
				Type:      "focal_category",
				Label:     strings.Title(fc),
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Shooting Condition filters
	if len(params.ShootingCondition) > 0 {
		for _, sc := range params.ShootingCondition {
			p := params
			p.ShootingCondition = removeStringFromSlice(p.ShootingCondition, sc)
			filters = append(filters, ActiveFilter{
				Type:      "shooting_condition",
				Label:     strings.ReplaceAll(strings.Title(sc), "_", " "),
				RemoveURL: s.urlMapper.BuildFullURL(p),
			})
		}
	}

	// Burst filter
	if params.InBurst != nil {
		p := params
		p.InBurst = nil
		label := "Not in Burst"
		if *params.InBurst {
			label = "In Burst"
		}
		filters = append(filters, ActiveFilter{
			Type:      "in_burst",
			Label:     label,
			RemoveURL: s.urlMapper.BuildFullURL(p),
		})
	}

	return filters
}

// removeStringFromSlice removes a value from a string slice
func removeStringFromSlice(slice []string, value string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
