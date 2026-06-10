package query

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// FuzzParsePath asserts ParsePath never panics and always yields pagination
// values inside the documented bounds, no matter what the client sends.
//
// Regression context: ?limit=0 used to flow through unclamped and crash the
// explorer with a division by zero; negative limits became SQLite "no LIMIT".
func FuzzParsePath(f *testing.F) {
	seeds := []struct{ path, query string }{
		{"/photos", ""},
		{"/photos", "limit=0"},
		{"/photos", "limit=-5&offset=-3"},
		{"/photos", "limit=999999999999999999999"}, // overflows int
		{"/2024/11/05", "color=blue&color=red"},
		{"/camera/Leica%20Camera%20AG/LEICA-M11-Monochrom", ""},
		{"/lens/RF24-70mm-F2.8-L-IS-USM", "page=2"},
		{"/color/blue", "sort=iso&order=asc"},
		{"/bursts", "in_burst=true"},
		{"/photos", "year=unknown&month=13&day=32"},
		{"/photos", "camera_make=a%26b%3Dc&lens=%2B%2B"},
		{"/photos", "iso_min=abc&aperture_max=ƒ"},
		{"//..//photos", "%zz=%zz"}, // malformed escapes
		{"/photos", strings.Repeat("color=red&", 100)},
	}
	for _, s := range seeds {
		f.Add(s.path, s.query)
	}

	mapper := NewURLMapper()
	f.Fuzz(func(t *testing.T, path, query string) {
		params, err := mapper.ParsePath(path, query)
		if err != nil {
			return // a structured error is an acceptable outcome
		}
		if params.Limit < 1 || params.Limit > MaxLimit {
			t.Errorf("ParsePath(%q, %q): Limit %d outside [1, %d]", path, query, params.Limit, MaxLimit)
		}
		if params.Offset < 0 || params.Offset > MaxOffset {
			t.Errorf("ParsePath(%q, %q): Offset %d outside [0, %d]", path, query, params.Offset, MaxOffset)
		}
		if params.Month != nil && (*params.Month < 1 || *params.Month > 12) {
			t.Errorf("ParsePath(%q, %q): Month %d outside [1, 12]", path, query, *params.Month)
		}
		if params.Day != nil && (*params.Day < 1 || *params.Day > 31) {
			t.Errorf("ParsePath(%q, %q): Day %d outside [1, 31]", path, query, *params.Day)
		}

		// Whatever was parsed must serialize and re-parse to the same state
		// (URLs shown to users must be stable under reload).
		url := mapper.BuildFullURL(params)
		reparsed, err := mapper.ParsePath(splitURL(url))
		if err != nil {
			t.Fatalf("re-parsing built URL %q failed: %v", url, err)
		}
		if !reflect.DeepEqual(normalizeParams(params), normalizeParams(reparsed)) {
			t.Errorf("parse(build(parse(x))) != parse(x)\n  input: %q %q\n  built: %q\n  first:  %+v\n  second: %+v",
				path, query, url, params, reparsed)
		}
	})
}

// TestURLRoundtripProperty generates random-but-valid QueryParams and asserts
// ParsePath(BuildFullURL(p)) restores them exactly.
//
// Regression context: the camera facet bug ("Leica Camera AG LEICA M11
// Monochrom" split on the first space, sending users to zero-result pages)
// was precisely a violation of this property.
func TestURLRoundtripProperty(t *testing.T) {
	t.Parallel()

	mapper := NewURLMapper()

	for seed := int64(1); seed <= 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		for i := 0; i < 50; i++ {
			params := randomParams(rng)

			url := mapper.BuildFullURL(params)
			reparsed, err := mapper.ParsePath(splitURL(url))
			if err != nil {
				t.Fatalf("seed %d iter %d: ParsePath(%q) failed: %v", seed, i, url, err)
			}

			got, want := normalizeParams(reparsed), normalizeParams(params)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("seed %d iter %d: roundtrip mismatch\n  url:  %q\n  want: %+v\n  got:  %+v",
					seed, i, url, want, got)
			}
		}
	}
}

// splitURL splits "/photos?a=b" into the (path, rawQuery) pair ParsePath takes.
func splitURL(url string) (string, string) {
	if i := strings.IndexByte(url, '?'); i >= 0 {
		return url[:i], url[i+1:]
	}
	return url, ""
}

// normalizeParams maps semantically-equal states to one representation:
// BuildQueryString elides the default sort (date_taken desc), so parsing the
// built URL yields "" where the input may have said it explicitly.
func normalizeParams(p QueryParams) QueryParams {
	if p.SortBy == "date_taken" {
		p.SortBy = ""
	}
	if p.SortOrder == "desc" {
		p.SortOrder = ""
	}
	return p
}

// Vocabularies deliberately include the adversarial cases from past bugs:
// multi-word makes, models with spaces, slashes, '&', '=', '+', and unicode.
var (
	genMakes = []string{
		"Canon", "Leica Camera AG", "Phase One A/S", "Hasselblad AB",
		"a&b=c", "trailing space ",
	}
	genModels = []string{
		"EOS R5", "LEICA M11 Monochrom", "X2D 100C", "iPhone 15 Pro Max", "IQ4 150MP",
	}
	genLenses = []string{
		"RF24-70mm F2.8 L IS USM", "Summilux-M 1:1.4/50 ASPH.", "ƒ=35mm + hood",
	}
	genColours    = []string{"red", "orange", "yellow", "green", "blue", "purple", "pink", "brown", "grey", "black", "white"}
	genTimesOfDay = []string{"morning", "midday", "afternoon", "evening", "night", "golden_hour_morning", "blue_hour"}
	genSeasons    = []string{"spring", "summer", "fall", "winter"}
	genFocalCats  = []string{"ultra_wide", "wide", "normal", "telephoto", "super_telephoto"}
	genConditions = []string{"bright", "low_light", "flash", "long_exposure"}
	genSorts      = []string{"", "iso", "aperture", "focal_length", "file_size"}
)

func pick(rng *rand.Rand, vocab []string) string { return vocab[rng.Intn(len(vocab))] }

func pickSome(rng *rand.Rand, vocab []string, max int) []string {
	n := rng.Intn(max + 1)
	if n == 0 {
		return nil
	}
	out := make([]string, 0, n)
	seen := map[string]bool{}
	for len(out) < n {
		v := pick(rng, vocab)
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// intPtr, float64Ptr, and boolPtr are shared with url_mapper_test.go.
func maybe(rng *rand.Rand, p float64) bool { return rng.Float64() < p }

// randomParams generates QueryParams restricted to the field set and value
// shapes that BuildQueryString serializes (apertures have one decimal, focal
// lengths are whole numbers, etc.).
func randomParams(rng *rand.Rand) QueryParams {
	p := QueryParams{Limit: DefaultLimit}

	if maybe(rng, 0.5) {
		if maybe(rng, 0.1) {
			p.Year = intPtr(-1) // "unknown" year marker
		} else {
			p.Year = intPtr(1990 + rng.Intn(40))
		}
	}
	if maybe(rng, 0.4) {
		p.Month = intPtr(1 + rng.Intn(12))
	}
	if maybe(rng, 0.2) {
		p.Day = intPtr(1 + rng.Intn(31))
	}

	if maybe(rng, 0.4) {
		p.CameraMake = []string{pick(rng, genMakes)}
		if maybe(rng, 0.8) {
			p.CameraModel = []string{pick(rng, genModels)}
		}
	}
	p.LensModel = pickSome(rng, genLenses, 2)
	p.ColourName = pickSome(rng, genColours, 3)
	p.TimeOfDay = pickSome(rng, genTimesOfDay, 2)
	p.Season = pickSome(rng, genSeasons, 2)
	p.FocalCategory = pickSome(rng, genFocalCats, 2)
	p.ShootingCondition = pickSome(rng, genConditions, 2)

	if maybe(rng, 0.3) {
		p.ISOMin = intPtr(50 << rng.Intn(6))
	}
	if maybe(rng, 0.3) {
		p.ISOMax = intPtr(800 << rng.Intn(5))
	}
	if maybe(rng, 0.3) {
		p.ApertureMin = float64Ptr(float64(10+rng.Intn(80)) / 10) // one decimal place
	}
	if maybe(rng, 0.3) {
		p.ApertureMax = float64Ptr(float64(20+rng.Intn(200)) / 10)
	}
	if maybe(rng, 0.2) {
		p.FocalLengthMin = float64Ptr(float64(8 + rng.Intn(100))) // whole mm
	}
	if maybe(rng, 0.2) {
		p.FocalLengthMax = float64Ptr(float64(50 + rng.Intn(550)))
	}

	if maybe(rng, 0.2) {
		p.InBurst = boolPtr(maybe(rng, 0.5))
	}
	if maybe(rng, 0.2) {
		p.HasGPS = boolPtr(maybe(rng, 0.5))
	}

	if maybe(rng, 0.3) {
		p.Limit = 1 + rng.Intn(MaxLimit)
	}
	if maybe(rng, 0.3) {
		p.Offset = rng.Intn(1000)
	}

	p.SortBy = pick(rng, genSorts)
	if p.SortBy != "" && maybe(rng, 0.5) {
		p.SortOrder = "asc"
	}

	return p
}
