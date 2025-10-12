# Explorer Implementation Plan

**Olsen Photo Explorer - Development Plan**
**Version:** 1.0
**Last Updated:** October 2025

---

## Implementation Phases

### Phase 1: Foundation (Steps 1-3)
Setup core server infrastructure and basic routing

### Phase 2: Repository Layer (Steps 4-5)
Create query interface for the explorer

### Phase 3: Templates (Steps 6-8)
Build HTML templates and styling

### Phase 4: Handlers (Steps 9-12)
Implement route handlers

### Phase 5: Integration (Steps 13-14)
Wire everything together and test

---

## Detailed Steps

### Step 1: Create Explorer Package Structure

**Files to create:**
```
internal/explorer/
├── server.go
├── handlers.go
├── repository.go
└── templates/
    └── (HTML files)
```

**server.go** - Core server setup:
- Server struct
- Route registration
- Start/Stop methods

---

### Step 2: Implement Repository Interface

**repository.go** - Query layer for explorer:
- GetStats() - Homepage statistics
- GetRecentPhotos(limit) - Recent photos
- GetPhotoByID(id) - Single photo detail
- GetPhotosByYear(year) - Year view
- GetPhotosByMonth(year, month) - Month view
- GetPhotosByDay(year, month, day) - Day view
- GetPhotosByCamera(make, model) - Camera view
- GetPhotosByLens(lens) - Lens view
- GetYears() - List all years
- GetCameras() - List all cameras
- GetLenses() - List all lenses
- GetThumbnail(photoID, size) - Thumbnail data

**Data structures:**
```go
type Stats struct {
    TotalPhotos   int
    CameraCount   int
    LensCount     int
    DateRangeFrom time.Time
    DateRangeTo   time.Time
    BurstCount    int
    ClusterCount  int
}

type PhotoGrid struct {
    Photos     []PhotoCard
    TotalCount int
    Page       int
    PerPage    int
}

type PhotoCard struct {
    ID          int
    ThumbnailID string  // For lazy loading
    DateTaken   time.Time
    CameraMake  string
    CameraModel string
}

type PhotoDetail struct {
    ID              int
    Thumbnail       []byte  // 1024px thumbnail
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
    Width           int
    Height          int
    Latitude        float64
    Longitude       float64
    DominantColors  []models.DominantColor

    // Navigation
    PrevID int
    NextID int
}
```

---

### Step 3: Create Base Template

**templates/layout.html** - Base HTML structure:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Olsen Photo Explorer</title>
    <style>
        /* Inline CSS - minimal, spartan design */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #1a1a1a;
            color: #e0e0e0;
            line-height: 1.6;
        }
        header {
            background: #2d2d2d;
            padding: 1rem 2rem;
            border-bottom: 2px solid #404040;
        }
        header h1 {
            font-size: 1.5rem;
            font-weight: 300;
            letter-spacing: 0.05em;
        }
        nav {
            margin-top: 0.5rem;
        }
        nav a {
            color: #888;
            text-decoration: none;
            margin-right: 1.5rem;
            font-size: 0.9rem;
        }
        nav a:hover { color: #fff; }
        main {
            max-width: 1400px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 1rem;
            margin-top: 2rem;
        }
        .card {
            background: #2d2d2d;
            border-radius: 4px;
            overflow: hidden;
            transition: transform 0.2s;
        }
        .card:hover {
            transform: translateY(-4px);
        }
        .card img {
            width: 100%;
            height: 250px;
            object-fit: cover;
            display: block;
        }
        .card-info {
            padding: 0.75rem;
            font-size: 0.85rem;
            color: #aaa;
        }
        a { color: #6ab7ff; }
        a:hover { color: #8ec9ff; }
    </style>
</head>
<body>
    <header>
        <h1>OLSEN PHOTO EXPLORER</h1>
        <nav>
            <a href="/">Home</a>
            <a href="/dates">By Date</a>
            <a href="/cameras">By Camera</a>
            <a href="/lenses">By Lens</a>
            <a href="/search">Search</a>
        </nav>
    </header>
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
```

---

### Step 4: Create Home Page Template

**templates/home.html**:

```html
{{define "content"}}
<section>
    <h2>Statistics</h2>
    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin: 2rem 0;">
        <div style="background: #2d2d2d; padding: 1.5rem; border-radius: 4px;">
            <div style="font-size: 2rem; font-weight: bold;">{{.Stats.TotalPhotos}}</div>
            <div style="color: #888; font-size: 0.9rem;">Photos</div>
        </div>
        <div style="background: #2d2d2d; padding: 1.5rem; border-radius: 4px;">
            <div style="font-size: 2rem; font-weight: bold;">{{.Stats.CameraCount}}</div>
            <div style="color: #888; font-size: 0.9rem;">Cameras</div>
        </div>
        <div style="background: #2d2d2d; padding: 1.5rem; border-radius: 4px;">
            <div style="font-size: 2rem; font-weight: bold;">{{.Stats.LensCount}}</div>
            <div style="color: #888; font-size: 0.9rem;">Lenses</div>
        </div>
        <div style="background: #2d2d2d; padding: 1.5rem; border-radius: 4px;">
            <div style="font-size: 2rem; font-weight: bold;">{{.Stats.BurstCount}}</div>
            <div style="color: #888; font-size: 0.9rem;">Bursts</div>
        </div>
    </div>
</section>

<section>
    <h2>Recent Photos</h2>
    <div class="grid">
        {{range .Photos}}
        <a href="/photo/{{.ID}}" class="card">
            <img src="/api/thumbnail/{{.ID}}/256" alt="Photo" loading="lazy">
            <div class="card-info">
                <div>{{.CameraMake}} {{.CameraModel}}</div>
                <div style="font-size: 0.8rem; color: #666;">{{.DateTaken.Format "Jan 2, 2006"}}</div>
            </div>
        </a>
        {{end}}
    </div>
</section>
{{end}}
```

---

### Step 5: Create Photo Grid Template

**templates/grid.html**:

```html
{{define "content"}}
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
    <div>
        <a href="{{.BackLink}}" style="color: #888;">← Back</a>
        <h2 style="display: inline; margin-left: 1rem;">{{.Title}}</h2>
        <span style="color: #666; margin-left: 1rem;">({{.TotalCount}} photos)</span>
    </div>
</div>

<div class="grid">
    {{range .Photos}}
    <a href="/photo/{{.ID}}" class="card">
        <img src="/api/thumbnail/{{.ID}}/256" alt="Photo" loading="lazy">
        <div class="card-info">
            <div>{{.CameraMake}} {{.CameraModel}}</div>
            <div style="font-size: 0.8rem; color: #666;">{{.DateTaken.Format "Jan 2, 2006 3:04 PM"}}</div>
        </div>
    </a>
    {{end}}
</div>

{{if or .PrevPage .NextPage}}
<div style="text-align: center; margin: 3rem 0;">
    {{if .PrevPage}}<a href="{{.PrevPage}}">← Previous</a>{{end}}
    <span style="margin: 0 2rem; color: #666;">Page {{.Page}}</span>
    {{if .NextPage}}<a href="{{.NextPage}}">Next →</a>{{end}}
</div>
{{end}}
{{end}}
```

---

### Step 6: Create Photo Detail Template

**templates/detail.html**:

```html
{{define "content"}}
<div style="display: flex; justify-content: space-between; margin-bottom: 1rem;">
    <a href="{{.BackLink}}" style="color: #888;">← Back to Grid</a>
    <div>
        {{if .PrevID}}<a href="/photo/{{.PrevID}}">← Prev</a>{{end}}
        {{if and .PrevID .NextID}}<span style="margin: 0 1rem; color: #666;">|</span>{{end}}
        {{if .NextID}}<a href="/photo/{{.NextID}}">Next →</a>{{end}}
    </div>
</div>

<div style="text-align: center; margin: 2rem 0;">
    <img src="data:image/jpeg;base64,{{.ThumbnailBase64}}"
         style="max-width: 100%; max-height: 70vh; border-radius: 4px;">
</div>

<div style="background: #2d2d2d; padding: 2rem; border-radius: 4px; margin-top: 2rem;">
    <h3>Details</h3>
    <table style="width: 100%; margin-top: 1rem;">
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Camera</td>
            <td>{{.CameraMake}} {{.CameraModel}}</td>
        </tr>
        {{if .LensModel}}
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Lens</td>
            <td>{{.LensModel}}</td>
        </tr>
        {{end}}
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Date</td>
            <td>{{.DateTaken.Format "January 2, 2006 at 3:04 PM"}}</td>
        </tr>
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Settings</td>
            <td>ISO {{.ISO}}, f/{{.Aperture}}, {{.ShutterSpeed}}, {{.FocalLength}}mm ({{.FocalLength35mm}}mm equiv.)</td>
        </tr>
        {{if .Latitude}}
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Location</td>
            <td>{{printf "%.4f" .Latitude}}, {{printf "%.4f" .Longitude}}</td>
        </tr>
        {{end}}
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Dimensions</td>
            <td>{{.Width}} × {{.Height}}</td>
        </tr>
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">File</td>
            <td style="font-family: monospace; font-size: 0.85rem;">{{.FilePath}}</td>
        </tr>
        <tr>
            <td style="color: #888; padding: 0.5rem 0;">Size</td>
            <td>{{.FileSizeMB}} MB</td>
        </tr>
    </table>

    {{if .DominantColors}}
    <div style="margin-top: 2rem;">
        <h4 style="color: #888; margin-bottom: 0.5rem;">Dominant Colors</h4>
        <div style="display: flex; gap: 0.5rem;">
            {{range .DominantColors}}
            <div style="width: 50px; height: 50px; background: rgb({{.Color.R}},{{.Color.G}},{{.Color.B}}); border-radius: 4px;"></div>
            {{end}}
        </div>
    </div>
    {{end}}
</div>
{{end}}
```

---

### Step 7: Create List Templates

**templates/years.html**:

```html
{{define "content"}}
<h2>Browse by Year</h2>
<div class="grid" style="grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));">
    {{range .Years}}
    <a href="/{{.Year}}" class="card">
        <div style="padding: 3rem 1rem; text-align: center;">
            <div style="font-size: 2rem; font-weight: bold;">{{.Year}}</div>
            <div style="color: #888; font-size: 0.9rem; margin-top: 0.5rem;">{{.Count}} photos</div>
        </div>
    </a>
    {{end}}
</div>
{{end}}
```

**templates/cameras.html**:

```html
{{define "content"}}
<h2>Browse by Camera</h2>
<div style="margin-top: 2rem;">
    {{range .CameraMakes}}
    <div style="background: #2d2d2d; padding: 1.5rem; border-radius: 4px; margin-bottom: 1rem;">
        <h3>{{.Make}} <span style="color: #666; font-weight: normal; font-size: 0.9rem;">({{.TotalCount}} photos)</span></h3>
        <div style="margin-top: 1rem;">
            {{range .Models}}
            <a href="/camera/{{$.Make}}/{{.Model}}" style="display: inline-block; background: #3d3d3d; padding: 0.5rem 1rem; border-radius: 4px; margin: 0.25rem; text-decoration: none;">
                {{.Model}} <span style="color: #666;">({{.Count}})</span>
            </a>
            {{end}}
        </div>
    </div>
    {{end}}
</div>
{{end}}
```

---

### Step 8: Implement Core Handlers

**handlers.go** - Route handler implementations:

```go
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
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

    data := map[string]interface{}{
        "Title":  "Home",
        "Stats":  stats,
        "Photos": photos,
    }

    s.renderTemplate(w, "home", data)
}

func (s *Server) handlePhotoDetail(w http.ResponseWriter, r *http.Request) {
    // Extract photo ID from URL
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

    // Encode thumbnail as base64
    photo.ThumbnailBase64 = base64.StdEncoding.EncodeToString(photo.Thumbnail)
    photo.FileSizeMB = fmt.Sprintf("%.1f", float64(photo.FileSize)/(1024*1024))

    data := map[string]interface{}{
        "Title":    "Photo Detail",
        "Photo":    photo,
        "BackLink": r.Referer(),
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

    thumbnail, err := s.repo.GetThumbnail(id, size)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Set cache headers
    w.Header().Set("Content-Type", "image/jpeg")
    w.Header().Set("Cache-Control", "public, max-age=31536000")
    w.Header().Set("ETag", fmt.Sprintf(`"%d-%s"`, id, size))

    w.Write(thumbnail)
}
```

---

### Step 9: Implement Template Rendering

**templates.go**:

```go
package explorer

import (
    "embed"
    "html/template"
    "net/http"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates *template.Template

func init() {
    templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    err := templates.ExecuteTemplate(w, name+".html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

---

### Step 10: Implement CLI Command

**cmd/olsen/explore.go**:

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os/exec"
    "runtime"

    "github.com/adewale/olsen/internal/database"
    "github.com/adewale/olsen/internal/explorer"
)

func exploreCommand(args []string) error {
    fs := flag.NewFlagSet("explore", flag.ExitOnError)
    dbPath := fs.String("db", "", "Path to database file (required)")
    addr := fs.String("addr", "localhost:8080", "Listen address")
    openBrowser := fs.Bool("open", false, "Open browser automatically")

    fs.Parse(args)

    if *dbPath == "" {
        return fmt.Errorf("--db flag is required")
    }

    // Open database
    db, err := database.Open(*dbPath)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }

    // Create server
    server := explorer.NewServer(db, *addr)

    // Open browser if requested
    if *openBrowser {
        go openBrowserURL("http://" + *addr)
    }

    log.Printf("Starting explorer server on http://%s", *addr)
    return server.Start()
}

func openBrowserURL(url string) {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "darwin":
        cmd = "open"
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start"}
    default:
        cmd = "xdg-open"
    }

    args = append(args, url)
    exec.Command(cmd, args...).Start()
}
```

---

### Step 11: Testing Strategy

1. **Unit Tests**
   - Repository query functions
   - Template rendering
   - Handler logic

2. **Integration Tests**
   - Full request/response cycle
   - Use test database with fixtures

3. **Manual Testing**
   - Test with real database
   - Check all routes
   - Verify thumbnails load
   - Test pagination

---

### Step 12: Build and Run

```bash
# Build
go build -o olsen cmd/olsen/*.go

# Run
./olsen explore --db /path/to/photos.db --addr localhost:8080 --open
```

---

## Implementation Checklist

### Core Infrastructure
- [ ] Create explorer package structure
- [ ] Implement Server struct and routing
- [ ] Setup template system
- [ ] Create base layout template

### Repository Layer
- [ ] Implement Stats query
- [ ] Implement Recent photos query
- [ ] Implement Photo by ID query
- [ ] Implement Year/Month/Day queries
- [ ] Implement Camera/Lens queries
- [ ] Implement Thumbnail query

### Templates
- [ ] Create home.html
- [ ] Create grid.html
- [ ] Create detail.html
- [ ] Create years.html
- [ ] Create cameras.html
- [ ] Create lenses.html

### Handlers
- [ ] Implement handleHome
- [ ] Implement handlePhotoDetail
- [ ] Implement handlePhotoGrid
- [ ] Implement handleYears
- [ ] Implement handleCameras
- [ ] Implement handleLenses
- [ ] Implement handleThumbnail (API)

### CLI Integration
- [ ] Add explore command to main CLI
- [ ] Add command-line flags
- [ ] Add browser auto-open

### Testing
- [ ] Unit tests for repository
- [ ] Handler tests
- [ ] Integration test with test DB
- [ ] Manual testing with real data

---

## Success Metrics

- [ ] Server starts successfully
- [ ] Homepage loads and shows stats
- [ ] Recent photos grid displays
- [ ] Photo detail page works
- [ ] Thumbnails load efficiently
- [ ] Navigation between views works
- [ ] Pagination functions correctly
- [ ] All routes respond without errors
- [ ] Page load time < 500ms
- [ ] Thumbnail load time < 100ms

---

## Notes

- Keep it simple: No JavaScript initially
- Focus on functionality over aesthetics
- Use Go standard library wherever possible
- Optimize database queries
- Cache thumbnails aggressively
- Test with large databases (10K+ photos)
