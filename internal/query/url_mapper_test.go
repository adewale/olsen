package query

import (
	"testing"
)

func TestParsePathLegacyRoutes(t *testing.T) {
	mapper := NewURLMapper()

	tests := []struct {
		name        string
		path        string
		queryString string
		want        QueryParams
	}{
		{
			name: "Camera make only",
			path: "/camera/Canon",
			want: QueryParams{
				CameraMake: []string{"Canon"},
				Limit:      50,
			},
		},
		{
			name: "Camera make and model",
			path: "/camera/Canon/EOS-R5",
			want: QueryParams{
				CameraMake:  []string{"Canon"},
				CameraModel: []string{"EOS R5"},
				Limit:       50,
			},
		},
		{
			name: "Lens",
			path: "/lens/Canon-RF-24-70",
			want: QueryParams{
				LensModel: []string{"Canon RF 24 70"},
				Limit:     50,
			},
		},
		{
			name: "Color",
			path: "/color/blue",
			want: QueryParams{
				ColourName: []string{"blue"},
				Limit:      50,
			},
		},
		{
			name: "Bursts",
			path: "/bursts",
			want: QueryParams{
				InBurst: boolPtr(true),
				Limit:   50,
			},
		},
		{
			name: "Time of day - morning",
			path: "/morning",
			want: QueryParams{
				TimeOfDay: []string{"morning"},
				Limit:     50,
			},
		},
		{
			name: "Time of day - blue hour",
			path: "/blue_hour",
			want: QueryParams{
				TimeOfDay: []string{"blue_hour"},
				Limit:     50,
			},
		},
		{
			name: "Time of day - golden hour morning",
			path: "/golden_hour_morning",
			want: QueryParams{
				TimeOfDay: []string{"golden_hour_morning"},
				Limit:     50,
			},
		},
		{
			name: "Season - spring",
			path: "/spring",
			want: QueryParams{
				Season: []string{"spring"},
				Limit:  50,
			},
		},
		{
			name: "Season - fall",
			path: "/fall",
			want: QueryParams{
				Season: []string{"fall"},
				Limit:  50,
			},
		},
		{
			name: "Focal category - wide",
			path: "/wide",
			want: QueryParams{
				FocalCategory: []string{"wide"},
				Limit:         50,
			},
		},
		{
			name: "Focal category - telephoto",
			path: "/telephoto",
			want: QueryParams{
				FocalCategory: []string{"telephoto"},
				Limit:         50,
			},
		},
		{
			name: "Year only",
			path: "/2025",
			want: QueryParams{
				Year:  intPtr(2025),
				Limit: 50,
			},
		},
		{
			name: "Year and month",
			path: "/2025/10",
			want: QueryParams{
				Year:  intPtr(2025),
				Month: intPtr(10),
				Limit: 50,
			},
		},
		{
			name: "Year, month, and day",
			path: "/2025/10/04",
			want: QueryParams{
				Year:  intPtr(2025),
				Month: intPtr(10),
				Day:   intPtr(4),
				Limit: 50,
			},
		},
		{
			name: "Invalid year (too low)",
			path: "/1899",
			want: QueryParams{
				Limit: 50,
			},
		},
		{
			name: "Invalid year (too high)",
			path: "/2101",
			want: QueryParams{
				Limit: 50,
			},
		},
		{
			name: "Invalid month",
			path: "/2025/13",
			want: QueryParams{
				Year:  intPtr(2025),
				Limit: 50,
			},
		},
		{
			name: "Invalid day",
			path: "/2025/10/32",
			want: QueryParams{
				Year:  intPtr(2025),
				Month: intPtr(10),
				Limit: 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapper.ParsePath(tt.path, tt.queryString)
			if err != nil {
				t.Fatalf("ParsePath() error = %v", err)
			}

			// Compare fields
			if !equalIntPtr(got.Year, tt.want.Year) {
				t.Errorf("Year = %v, want %v", got.Year, tt.want.Year)
			}
			if !equalIntPtr(got.Month, tt.want.Month) {
				t.Errorf("Month = %v, want %v", got.Month, tt.want.Month)
			}
			if !equalIntPtr(got.Day, tt.want.Day) {
				t.Errorf("Day = %v, want %v", got.Day, tt.want.Day)
			}
			if !equalStringSlice(got.CameraMake, tt.want.CameraMake) {
				t.Errorf("CameraMake = %v, want %v", got.CameraMake, tt.want.CameraMake)
			}
			if !equalStringSlice(got.CameraModel, tt.want.CameraModel) {
				t.Errorf("CameraModel = %v, want %v", got.CameraModel, tt.want.CameraModel)
			}
			if !equalStringSlice(got.LensModel, tt.want.LensModel) {
				t.Errorf("LensModel = %v, want %v", got.LensModel, tt.want.LensModel)
			}
			if !equalStringSlice(got.ColourName, tt.want.ColourName) {
				t.Errorf("ColourName = %v, want %v", got.ColourName, tt.want.ColourName)
			}
			if !equalStringSlice(got.TimeOfDay, tt.want.TimeOfDay) {
				t.Errorf("TimeOfDay = %v, want %v", got.TimeOfDay, tt.want.TimeOfDay)
			}
			if !equalStringSlice(got.Season, tt.want.Season) {
				t.Errorf("Season = %v, want %v", got.Season, tt.want.Season)
			}
			if !equalStringSlice(got.FocalCategory, tt.want.FocalCategory) {
				t.Errorf("FocalCategory = %v, want %v", got.FocalCategory, tt.want.FocalCategory)
			}
			if !equalBoolPtr(got.InBurst, tt.want.InBurst) {
				t.Errorf("InBurst = %v, want %v", got.InBurst, tt.want.InBurst)
			}
			if got.Limit != tt.want.Limit {
				t.Errorf("Limit = %v, want %v", got.Limit, tt.want.Limit)
			}
		})
	}
}

func TestParseQueryString(t *testing.T) {
	mapper := NewURLMapper()

	tests := []struct {
		name        string
		path        string
		queryString string
		want        QueryParams
	}{
		{
			name:        "All temporal filters",
			path:        "/photos",
			queryString: "year=2025&month=10&day=4",
			want: QueryParams{
				Year:  intPtr(2025),
				Month: intPtr(10),
				Day:   intPtr(4),
				Limit: 50,
			},
		},
		{
			name:        "Pagination",
			path:        "/photos",
			queryString: "limit=100&offset=50",
			want: QueryParams{
				Limit:  100,
				Offset: 50,
			},
		},
		{
			name:        "Sorting",
			path:        "/photos",
			queryString: "sort=focal_length&order=asc",
			want: QueryParams{
				SortBy:    "focal_length",
				SortOrder: "asc",
				Limit:     50,
			},
		},
		{
			name:        "Multiple camera makes",
			path:        "/photos",
			queryString: "camera_make=Canon&camera_make=Nikon",
			want: QueryParams{
				CameraMake: []string{"Canon", "Nikon"},
				Limit:      50,
			},
		},
		{
			name:        "Multiple colors",
			path:        "/photos",
			queryString: "color=red&color=blue&color=green",
			want: QueryParams{
				ColourName: []string{"red", "blue", "green"},
				Limit:      50,
			},
		},
		{
			name:        "Technical ranges - ISO",
			path:        "/photos",
			queryString: "iso_min=100&iso_max=3200",
			want: QueryParams{
				ISOMin: intPtr(100),
				ISOMax: intPtr(3200),
				Limit:  50,
			},
		},
		{
			name:        "Technical ranges - Aperture",
			path:        "/photos",
			queryString: "aperture_min=1.4&aperture_max=5.6",
			want: QueryParams{
				ApertureMin: float64Ptr(1.4),
				ApertureMax: float64Ptr(5.6),
				Limit:       50,
			},
		},
		{
			name:        "Technical ranges - Focal length",
			path:        "/photos",
			queryString: "focal_min=24&focal_max=70",
			want: QueryParams{
				FocalLengthMin: float64Ptr(24),
				FocalLengthMax: float64Ptr(70),
				Limit:          50,
			},
		},
		{
			name:        "Boolean - in burst true",
			path:        "/photos",
			queryString: "in_burst=true",
			want: QueryParams{
				InBurst: boolPtr(true),
				Limit:   50,
			},
		},
		{
			name:        "Boolean - in burst false",
			path:        "/photos",
			queryString: "in_burst=false",
			want: QueryParams{
				InBurst: boolPtr(false),
				Limit:   50,
			},
		},
		{
			name:        "Boolean - has GPS",
			path:        "/photos",
			queryString: "has_gps=1",
			want: QueryParams{
				HasGPS: boolPtr(true),
				Limit:  50,
			},
		},
		{
			name:        "Complex combination",
			path:        "/photos",
			queryString: "year=2025&color=red&camera_make=Canon&time_of_day=morning&limit=100",
			want: QueryParams{
				Year:       intPtr(2025),
				ColourName: []string{"red"},
				CameraMake: []string{"Canon"},
				TimeOfDay:  []string{"morning"},
				Limit:      100,
			},
		},
		{
			name:        "Legacy path with query string override",
			path:        "/2024",
			queryString: "year=2025",
			want: QueryParams{
				Year:  intPtr(2025), // Query string overrides path
				Limit: 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapper.ParsePath(tt.path, tt.queryString)
			if err != nil {
				t.Fatalf("ParsePath() error = %v", err)
			}

			// Compare all fields
			compareQueryParams(t, got, tt.want)
		})
	}
}

func TestBuildQueryString(t *testing.T) {
	mapper := NewURLMapper()

	tests := []struct {
		name   string
		params QueryParams
		want   string
	}{
		{
			name: "Empty params",
			params: QueryParams{
				Limit: 50, // Default limit doesn't appear in query string
			},
			want: "",
		},
		{
			name: "Year only",
			params: QueryParams{
				Year:  intPtr(2025),
				Limit: 50,
			},
			want: "?year=2025",
		},
		{
			name: "Multiple filters",
			params: QueryParams{
				Year:       intPtr(2025),
				ColourName: []string{"red", "blue"},
				CameraMake: []string{"Canon"},
				Limit:      50,
			},
			want: "?camera_make=Canon&color=red&color=blue&year=2025",
		},
		{
			name: "Pagination non-default",
			params: QueryParams{
				Limit:  100,
				Offset: 50,
			},
			want: "?limit=100&offset=50",
		},
		{
			name: "Technical ranges",
			params: QueryParams{
				ISOMin:         intPtr(100),
				ISOMax:         intPtr(3200),
				ApertureMin:    float64Ptr(1.4),
				ApertureMax:    float64Ptr(5.6),
				FocalLengthMin: float64Ptr(24.0),
				FocalLengthMax: float64Ptr(70.0),
				Limit:          50,
			},
			want: "?aperture_max=5.6&aperture_min=1.4&focal_max=70&focal_min=24&iso_max=3200&iso_min=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.BuildQueryString(tt.params)
			if got != tt.want {
				t.Errorf("BuildQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFullURL(t *testing.T) {
	mapper := NewURLMapper()

	params := QueryParams{
		Year:       intPtr(2025),
		ColourName: []string{"red"},
		CameraMake: []string{"Canon"},
		Limit:      100,
	}

	got := mapper.BuildFullURL(params)
	// Should be /photos?...
	if got[:7] != "/photos" {
		t.Errorf("BuildFullURL() path = %v, want /photos", got[:7])
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalBoolPtr(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareQueryParams(t *testing.T, got, want QueryParams) {
	t.Helper()

	if !equalIntPtr(got.Year, want.Year) {
		t.Errorf("Year = %v, want %v", got.Year, want.Year)
	}
	if !equalIntPtr(got.Month, want.Month) {
		t.Errorf("Month = %v, want %v", got.Month, want.Month)
	}
	if !equalIntPtr(got.Day, want.Day) {
		t.Errorf("Day = %v, want %v", got.Day, want.Day)
	}
	if !equalStringSlice(got.CameraMake, want.CameraMake) {
		t.Errorf("CameraMake = %v, want %v", got.CameraMake, want.CameraMake)
	}
	if !equalStringSlice(got.CameraModel, want.CameraModel) {
		t.Errorf("CameraModel = %v, want %v", got.CameraModel, want.CameraModel)
	}
	if !equalStringSlice(got.LensModel, want.LensModel) {
		t.Errorf("LensModel = %v, want %v", got.LensModel, want.LensModel)
	}
	if !equalStringSlice(got.ColourName, want.ColourName) {
		t.Errorf("ColourName = %v, want %v", got.ColourName, want.ColourName)
	}
	if !equalStringSlice(got.TimeOfDay, want.TimeOfDay) {
		t.Errorf("TimeOfDay = %v, want %v", got.TimeOfDay, want.TimeOfDay)
	}
	if !equalStringSlice(got.Season, want.Season) {
		t.Errorf("Season = %v, want %v", got.Season, want.Season)
	}
	if !equalStringSlice(got.FocalCategory, want.FocalCategory) {
		t.Errorf("FocalCategory = %v, want %v", got.FocalCategory, want.FocalCategory)
	}
	if !equalBoolPtr(got.InBurst, want.InBurst) {
		t.Errorf("InBurst = %v, want %v", got.InBurst, want.InBurst)
	}
	if !equalBoolPtr(got.HasGPS, want.HasGPS) {
		t.Errorf("HasGPS = %v, want %v", got.HasGPS, want.HasGPS)
	}
	if got.Limit != want.Limit {
		t.Errorf("Limit = %v, want %v", got.Limit, want.Limit)
	}
	if got.Offset != want.Offset {
		t.Errorf("Offset = %v, want %v", got.Offset, want.Offset)
	}
	if got.SortBy != want.SortBy {
		t.Errorf("SortBy = %v, want %v", got.SortBy, want.SortBy)
	}
	if got.SortOrder != want.SortOrder {
		t.Errorf("SortOrder = %v, want %v", got.SortOrder, want.SortOrder)
	}
	if !equalIntPtr(got.ISOMin, want.ISOMin) {
		t.Errorf("ISOMin = %v, want %v", got.ISOMin, want.ISOMin)
	}
	if !equalIntPtr(got.ISOMax, want.ISOMax) {
		t.Errorf("ISOMax = %v, want %v", got.ISOMax, want.ISOMax)
	}
	if !equalFloat64Ptr(got.ApertureMin, want.ApertureMin) {
		t.Errorf("ApertureMin = %v, want %v", got.ApertureMin, want.ApertureMin)
	}
	if !equalFloat64Ptr(got.ApertureMax, want.ApertureMax) {
		t.Errorf("ApertureMax = %v, want %v", got.ApertureMax, want.ApertureMax)
	}
	if !equalFloat64Ptr(got.FocalLengthMin, want.FocalLengthMin) {
		t.Errorf("FocalLengthMin = %v, want %v", got.FocalLengthMin, want.FocalLengthMin)
	}
	if !equalFloat64Ptr(got.FocalLengthMax, want.FocalLengthMax) {
		t.Errorf("FocalLengthMax = %v, want %v", got.FocalLengthMax, want.FocalLengthMax)
	}
}

func equalFloat64Ptr(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
