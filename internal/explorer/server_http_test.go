package explorer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/testsupport"
)

// HTTP-level tests for the explorer. These assert both the positive signal
// (the right content is present) and the negative one (errors, leaks, and
// wrong-status responses are absent), per the "not-empty assertions are weak
// oracles" rule.

func newTestServer(t *testing.T) *Server {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	testsupport.InsertPhotos(t, db,
		testsupport.NewPhoto().
			WithCamera("Canon", "EOS R5").
			WithDateTaken(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)).
			WithColourName("blue").
			WithThumbnail("64").WithThumbnail("256").
			Build(),
		testsupport.NewPhoto().
			WithCamera("Leica Camera AG", "LEICA M11 Monochrom").
			WithDateTaken(time.Date(2023, 11, 5, 9, 0, 0, 0, time.UTC)).
			WithColourName("gray").
			WithThumbnail("64").
			Build(),
	)

	return NewServer(db, "localhost:0")
}

func get(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func TestHTTPPhotosPage(t *testing.T) {
	s := newTestServer(t)

	rec := get(t, s, "/photos")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /photos = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Canon") || !strings.Contains(body, "Leica Camera AG") {
		t.Error("photo grid missing expected camera facet content")
	}
	if strings.Contains(body, "Internal server error") {
		t.Error("page contains an error message despite 200 status")
	}
}

// TestHTTPLimitZeroDoesNotPanic is the HTTP-level regression for the
// division-by-zero panic: the full handler path, not just the URL parser,
// must survive hostile pagination values.
func TestHTTPLimitZeroDoesNotPanic(t *testing.T) {
	s := newTestServer(t)

	for _, q := range []string{"?limit=0", "?limit=-5", "?limit=0&page=3", "?offset=-1"} {
		rec := get(t, s, "/photos"+q)
		if rec.Code != http.StatusOK {
			t.Errorf("GET /photos%s = %d, want 200; body: %s", q, rec.Code, rec.Body.String())
		}
	}
}

// TestHTTPUnknownPathIs404 covers the regression where every unknown URL
// returned a blank 200 page.
func TestHTTPUnknownPathIs404(t *testing.T) {
	s := newTestServer(t)

	for _, path := range []string{"/nonexistent", "/photos/extra", "/api/unknown"} {
		rec := get(t, s, path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
		if rec.Code == http.StatusOK && rec.Body.Len() == 0 {
			t.Errorf("GET %s returned a blank 200 page", path)
		}
	}
}

// TestHTTPMissingPhotoDoesNotLeakInternals asserts the 404 carries a generic
// message: SQL errors and file paths must stay in the server log.
func TestHTTPMissingPhotoDoesNotLeakInternals(t *testing.T) {
	s := newTestServer(t)

	rec := get(t, s, "/photo/99999")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /photo/99999 = %d, want 404", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Photo not found") {
		t.Errorf("404 body missing generic message, got: %q", body)
	}
	for _, leak := range []string{"sql:", "no rows", "sqlite", "/test/"} {
		if strings.Contains(body, leak) {
			t.Errorf("404 body leaks internal detail %q: %q", leak, body)
		}
	}
}

func TestHTTPThumbnail(t *testing.T) {
	s := newTestServer(t)

	rec := get(t, s, "/api/thumbnail/1/256")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/thumbnail/1/256 = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", ct)
	}
	if etag := rec.Header().Get("ETag"); etag == "" {
		t.Error("thumbnail response missing ETag header")
	}
	if rec.Body.Len() == 0 {
		t.Error("thumbnail response has empty body")
	}

	// Conditional request with the returned ETag must yield 304 and no body.
	req := httptest.NewRequest(http.MethodGet, "/api/thumbnail/1/256", nil)
	req.Header.Set("If-None-Match", rec.Header().Get("ETag"))
	rec304 := httptest.NewRecorder()
	s.router.ServeHTTP(rec304, req)
	if rec304.Code != http.StatusNotModified {
		t.Errorf("conditional GET = %d, want 304", rec304.Code)
	}
	if rec304.Body.Len() != 0 {
		t.Errorf("304 response carries a body of %d bytes", rec304.Body.Len())
	}

	// Sad paths: bad size and missing photo are 4xx, never 200 or 500.
	if rec := get(t, s, "/api/thumbnail/1/333"); rec.Code != http.StatusBadRequest {
		t.Errorf("invalid size = %d, want 400", rec.Code)
	}
	if rec := get(t, s, "/api/thumbnail/99999/256"); rec.Code != http.StatusNotFound {
		t.Errorf("missing photo thumbnail = %d, want 404", rec.Code)
	}
}

// TestHTTPGrayFacetClickReturnsFilteredResults is the HTTP-level pin for the
// gray/grey classification bug: clicking the gray colour facet must filter,
// not silently return every photo.
func TestHTTPGrayFacetClickReturnsFilteredResults(t *testing.T) {
	s := newTestServer(t)

	all := get(t, s, "/photos")
	gray := get(t, s, "/photos?color=gray")
	if gray.Code != http.StatusOK {
		t.Fatalf("GET /photos?color=gray = %d, want 200", gray.Code)
	}

	// The gray-filtered page must show the Leica photo's camera but not the
	// Canon one (whose only colour is blue).
	body := gray.Body.String()
	if !strings.Contains(body, "Leica Camera AG") {
		t.Error("gray filter dropped the gray photo")
	}
	if strings.Count(body, "/api/thumbnail/") >= strings.Count(all.Body.String(), "/api/thumbnail/") {
		t.Error("gray filter returned at least as many photos as the unfiltered page — filter not applied")
	}
}
