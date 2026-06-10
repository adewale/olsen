// Package explorer provides the web-based photo browser with faceted search.
//
// It implements an HTTP server with embedded HTML templates, state machine-based
// faceted navigation, and dynamic thumbnail serving with ETag caching. The explorer
// provides a read-only view into the photo database with filtering by date, camera,
// color, and other metadata dimensions.
package explorer

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	httpSrv   *http.Server
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

// Start starts the HTTP server and blocks until it exits.
func (s *Server) Start() error {
	log.Printf("Starting explorer server on http://%s", s.addr)
	s.httpSrv = &http.Server{
		Addr:              s.addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	err := s.httpSrv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown gracefully stops the HTTP server, waiting for in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// serverError logs the detailed error and sends a generic 500 to the client,
// avoiding leaking internal paths and SQL errors.
func serverError(w http.ResponseWriter, err error) {
	log.Printf("Internal error: %v", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

// titleCase upper-cases the first letter of each space-separated word.
// Replaces the deprecated strings.Title; our inputs are known ASCII facet
// values ("blue", "golden hour"), not arbitrary Unicode text.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// Clone the template set and add the specific content template as "content"
	tmpl, err := templates.Clone()
	if err != nil {
		serverError(w, fmt.Errorf("template clone: %w", err))
		return
	}

	// Get the named template and add it as "content"
	contentTmpl := templates.Lookup(name)
	if contentTmpl == nil {
		serverError(w, fmt.Errorf("template not found: %s", name))
		return
	}

	// Add the content template with the name "content" so layout can find it
	_, err = tmpl.AddParseTree("content", contentTmpl.Tree)
	if err != nil {
		serverError(w, fmt.Errorf("template parse: %w", err))
		return
	}

	// Render to a buffer first so a mid-render failure produces a clean 500
	// instead of a partial page followed by a superfluous header write.
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		serverError(w, fmt.Errorf("template execution: %w", err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Response write error: %v", err)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// The "/" pattern matches every otherwise-unhandled path; anything
		// that is not exactly the home page is a 404 (previously this
		// returned a blank 200).
		http.NotFound(w, r)
		return
	}

	stats, err := s.repo.GetStats()
	if err != nil {
		serverError(w, err)
		return
	}

	photos, err := s.repo.GetRecentPhotos(50)
	if err != nil {
		serverError(w, err)
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
		http.Error(w, "Photo not found", http.StatusNotFound)
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
		http.Error(w, "Thumbnail not found", http.StatusNotFound)
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

	if _, err := w.Write(thumbnail); err != nil {
		log.Printf("Thumbnail write error: %v", err)
	}
}

func (s *Server) handleDates(w http.ResponseWriter, r *http.Request) {
	years, err := s.repo.GetYears()
	if err != nil {
		serverError(w, err)
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
		serverError(w, err)
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
		serverError(w, err)
		return
	}

	data := map[string]interface{}{
		"Title":  "Browse by Lens",
		"Lenses": lenses,
	}

	s.renderTemplate(w, "lenses", data)
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

	// Defensive: ParsePath clamps the limit, but pagination math below
	// divides by it, so never allow a non-positive value through.
	if params.Limit <= 0 {
		params.Limit = query.DefaultLimit
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
		serverError(w, err)
		return
	}

	// Facets are always computed for browsing views; a failure degrades the
	// page (no facet rail) rather than failing the whole request.
	facets, err := s.engine.ComputeFacets(params)
	if err != nil {
		log.Printf("Facet computation error: %v", err)
		facets = nil
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
		title = titleCase(params.ColourName[0]) + " Photos"
	} else if len(params.TimeOfDay) > 0 {
		title = titleCase(params.TimeOfDay[0]) + " Photos"
	}

	// Determine current sort value for dropdown
	currentSort := "date_taken_desc" // default
	if params.SortBy != "" {
		if params.SortBy == "date_taken" {
			if params.SortOrder == "asc" {
				currentSort = "date_taken_asc"
			}
			// else stays date_taken_desc
		} else {
			currentSort = params.SortBy
		}
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
		"CurrentSort":   currentSort,
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
				Label:     titleCase(colour),
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
				Label:     titleCase(tod),
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
				Label:     titleCase(season),
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
				Label:     titleCase(fc),
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
				Label:     strings.ReplaceAll(titleCase(sc), "_", " "),
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
