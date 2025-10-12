# Explorer Interface Specification

**Olsen Photo Explorer - Minimal Web Interface**
**Version:** 1.0
**Last Updated:** October 2025

---

## Overview

A minimal, spartan web interface for browsing the photo catalog. The explorer provides essential functionality with zero dependencies beyond Go's standard library.

### Design Principles

1. **Minimal**: No JavaScript frameworks, no CSS frameworks, no build tools
2. **Fast**: Server-side rendering, optimized queries, efficient thumbnails
3. **Functional**: Cover core browsing needs without feature bloat
4. **Portable**: Single binary, works anywhere Go runs
5. **Clean**: Simple HTML/CSS, keyboard-friendly, responsive

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Server                          │
│                  (net/http package)                     │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
   ┌────▼────┐              ┌────▼────┐
   │ HTML    │              │  API    │
   │ Views   │              │ Routes  │
   └────┬────┘              └────┬────┘
        │                         │
        └────────────┬────────────┘
                     │
        ┌────────────▼────────────┐
        │   Repository Layer      │
        │   (Query Engine)        │
        └────────────┬────────────┘
                     │
        ┌────────────▼────────────┐
        │   SQLite Database       │
        └─────────────────────────┘
```

---

## Features

### Phase 1: Core Browsing (MVP)

1. **Home Page**
   - Total photo count
   - Quick stats (cameras, date range, bursts, duplicates)
   - Recent photos grid (last 50)

2. **Photo Grid View**
   - Responsive grid layout (3-5 columns)
   - 256px thumbnails
   - Date taken, camera info overlay on hover
   - Click to view detail

3. **Photo Detail View**
   - 1024px preview image
   - Full EXIF metadata display
   - Color palette visualization
   - Navigation (prev/next in current query)

4. **Temporal Browsing**
   - Year listing (/dates)
   - Month view (/YYYY/MM)
   - Day view (/YYYY/MM/DD)

5. **Equipment Browsing**
   - Camera make listing (/cameras)
   - Camera model view (/camera/:make/:model)
   - Lens listing (/lenses)
   - Lens view (/lens/:model)

6. **Simple Search**
   - Basic search form with common filters
   - Camera, lens, date range, ISO range

### Phase 2: Advanced Features (Future)

- Burst/duplicate browsing
- Color search
- Full faceted interface
- Keyboard shortcuts
- Export functionality

---

## Routes

### HTML Routes (Server-Side Rendered)

```
GET  /                         → Home page
GET  /photos                   → Recent photos grid
GET  /photo/:id                → Photo detail view

GET  /dates                    → Year listing
GET  /:year                    → Year view (photos from year)
GET  /:year/:month             → Month view
GET  /:year/:month/:day        → Day view

GET  /cameras                  → Camera make listing
GET  /camera/:make             → Camera make view
GET  /camera/:make/:model      → Camera model view

GET  /lenses                   → Lens listing
GET  /lens/:model              → Lens view

GET  /search                   → Search form
GET  /search?...               → Search results
```

### API Routes (JSON)

```
GET  /api/photo/:id            → Photo metadata JSON
GET  /api/thumbnail/:id/:size  → Thumbnail image (JPEG)
GET  /api/stats                → Database statistics
```

---

## Page Templates

### Home Page

```
┌─────────────────────────────────────────────────┐
│  OLSEN PHOTO EXPLORER                           │
├─────────────────────────────────────────────────┤
│  📊 Stats:                                      │
│  • 1,234 photos indexed                         │
│  • 5 cameras, 8 lenses                          │
│  • 2020-2025 (5 years)                          │
│  • 45 bursts, 12 duplicate clusters             │
├─────────────────────────────────────────────────┤
│  Recent Photos:                                 │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐                │
│  │img│ │img│ │img│ │img│ │img│                │
│  └───┘ └───┘ └───┘ └───┘ └───┘                │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐                │
│  │img│ │img│ │img│ │img│ │img│                │
│  └───┘ └───┘ └───┘ └───┘ └───┘                │
├─────────────────────────────────────────────────┤
│  Browse:                                        │
│  • By Date    • By Camera    • Search          │
└─────────────────────────────────────────────────┘
```

### Photo Grid View

```
┌─────────────────────────────────────────────────┐
│  < Back to Home        Canon EOS R5 (234)       │
├─────────────────────────────────────────────────┤
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │
│  │        │ │        │ │        │ │        │  │
│  │  img   │ │  img   │ │  img   │ │  img   │  │
│  │        │ │        │ │        │ │        │  │
│  │ Canon  │ │ Canon  │ │ Canon  │ │ Canon  │  │
│  │ Oct 15 │ │ Oct 15 │ │ Oct 14 │ │ Oct 14 │  │
│  └────────┘ └────────┘ └────────┘ └────────┘  │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │
│  │        │ │        │ │        │ │        │  │
│  │  img   │ │  img   │ │  img   │ │  img   │  │
│  │        │ │        │ │        │ │        │  │
│  └────────┘ └────────┘ └────────┘ └────────┘  │
├─────────────────────────────────────────────────┤
│  Showing 1-50 of 234         [Prev] [Next]     │
└─────────────────────────────────────────────────┘
```

### Photo Detail View

```
┌─────────────────────────────────────────────────┐
│  < Back to Grid          [Prev] [Next]          │
├─────────────────────────────────────────────────┤
│                                                 │
│              ┌──────────────────┐               │
│              │                  │               │
│              │                  │               │
│              │   Large Preview  │               │
│              │    (1024px)      │               │
│              │                  │               │
│              │                  │               │
│              └──────────────────┘               │
│                                                 │
├─────────────────────────────────────────────────┤
│  📷 Canon EOS R5 + RF24-70mm F2.8 L            │
│  📅 October 15, 2025 at 2:34 PM                │
│  ⚙️  ISO 100, f/2.8, 1/1000s, 50mm (50mm)     │
│  📍 San Francisco, CA                           │
│  🎨 Colors: ████ ████ ████ ████ ████          │
│                                                 │
│  File: /photos/2025/10/DSC_1234.dng            │
│  Hash: abc123...                                │
│  Size: 45.2 MB                                  │
└─────────────────────────────────────────────────┘
```

---

## Technical Implementation

### Server Structure

```go
package explorer

type Server struct {
    db       *database.DB
    repo     *repository.Repository
    addr     string
    router   *http.ServeMux
}

func NewServer(db *database.DB, addr string) *Server {
    s := &Server{
        db:     db,
        repo:   repository.New(db),
        addr:   addr,
        router: http.NewServeMux(),
    }

    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    // HTML routes
    s.router.HandleFunc("/", s.handleHome)
    s.router.HandleFunc("/photos", s.handlePhotoGrid)
    s.router.HandleFunc("/photo/", s.handlePhotoDetail)
    s.router.HandleFunc("/dates", s.handleDates)
    s.router.HandleFunc("/cameras", s.handleCameras)
    s.router.HandleFunc("/lenses", s.handleLenses)
    s.router.HandleFunc("/search", s.handleSearch)

    // API routes
    s.router.HandleFunc("/api/thumbnail/", s.handleThumbnail)
    s.router.HandleFunc("/api/photo/", s.handlePhotoAPI)
    s.router.HandleFunc("/api/stats", s.handleStats)
}

func (s *Server) Start() error {
    return http.ListenAndServe(s.addr, s.router)
}
```

### HTML Templates

Use Go's `html/template` package for server-side rendering:

```go
var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    err := templates.ExecuteTemplate(w, name+".html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

### CSS Styling

Minimal inline CSS with:
- CSS Grid for responsive layouts
- Simple color scheme (dark mode friendly)
- No external dependencies
- ~200 lines total

---

## Data Flow

### Photo Grid Request

```
1. Browser → GET /camera/Canon/EOS-R5
2. Server → Parse route parameters
3. Server → Build query (camera_make=Canon, camera_model=EOS-R5)
4. Repository → Execute query
5. Database → Return photos + metadata
6. Server → Render template with photo list
7. Server → HTML response
8. Browser → Display grid
9. Browser → Request thumbnails (async): GET /api/thumbnail/:id/256
10. Server → Query thumbnail from database
11. Server → Return JPEG bytes
12. Browser → Display thumbnail
```

### Photo Detail Request

```
1. Browser → GET /photo/123
2. Server → Query photo ID 123
3. Database → Return full metadata + 1024px thumbnail
4. Server → Render detail template
5. Server → Embedded data URL for thumbnail (inline base64)
6. Browser → Display photo detail page
```

---

## Performance Optimizations

1. **Thumbnail Caching**
   - Set Cache-Control headers (1 year)
   - ETags based on photo ID + size
   - Conditional GET support

2. **Query Optimization**
   - Limit results to reasonable page sizes (50-100)
   - Use database indexes effectively
   - Pre-compute facet counts where possible

3. **Connection Pooling**
   - Reuse database connections
   - Set reasonable pool limits

4. **Response Compression**
   - Gzip HTML responses
   - Thumbnails already compressed (JPEG)

---

## Security Considerations

1. **No Authentication** (single-user, local access assumed)
2. **Path Traversal Protection** - Validate all file paths
3. **SQL Injection Protection** - Use parameterized queries
4. **XSS Protection** - Escape all user input in templates
5. **Rate Limiting** - Basic protection against abuse

---

## Configuration

```go
type Config struct {
    DatabasePath string // Path to SQLite database
    ListenAddr   string // e.g., "localhost:8080"
    MaxPageSize  int    // Maximum photos per page (default: 100)
}
```

---

## Command-Line Interface

```bash
# Start explorer server
olsen explore --db /path/to/photos.db --addr localhost:8080

# Options
--db PATH        Database file path (required)
--addr ADDR      Listen address (default: localhost:8080)
--limit N        Max photos per page (default: 100)
--open           Open browser automatically
```

---

## Success Criteria

### Must Have (MVP)
- ✅ View photo grid
- ✅ View photo detail
- ✅ Browse by date (year/month/day)
- ✅ Browse by camera
- ✅ Browse by lens
- ✅ Display thumbnails efficiently
- ✅ Show full EXIF metadata
- ✅ Pagination support

### Should Have
- ⚠️  Search with basic filters
- ⚠️  Responsive layout (mobile-friendly)
- ⚠️  Keyboard navigation
- ⚠️  Breadcrumb navigation

### Nice to Have
- ⭕ Burst/duplicate browsing
- ⭕ Color search
- ⭕ Export functionality
- ⭕ Slideshow mode

---

## File Structure

```
internal/explorer/
├── server.go          # HTTP server setup
├── handlers.go        # Route handlers
├── templates.go       # Template rendering
├── repository.go      # Query logic (reuse from main repo)
└── templates/
    ├── layout.html    # Base layout
    ├── home.html      # Home page
    ├── grid.html      # Photo grid
    ├── detail.html    # Photo detail
    ├── years.html     # Year listing
    ├── cameras.html   # Camera listing
    └── search.html    # Search form

cmd/olsen/
└── explore.go         # CLI command for explorer
```

---

## Estimated Implementation

- **Core server setup**: 2-3 hours
- **HTML templates**: 2-3 hours
- **Route handlers**: 3-4 hours
- **CSS styling**: 1-2 hours
- **Testing**: 2-3 hours

**Total**: ~10-15 hours for MVP

---

## Future Enhancements

1. **WebSocket updates** - Live updates when new photos indexed
2. **Advanced search** - Full faceted interface
3. **Collections** - User-defined photo collections
4. **Editing** - Basic metadata editing
5. **Sharing** - Generate shareable links
6. **Export** - Batch export functionality
7. **Statistics** - Visual charts and graphs
