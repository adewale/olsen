package query

import "time"

// QueryParams represents all possible query filters
type QueryParams struct {
	// Temporal filters
	Year      *int
	Month     *int
	Day       *int
	DateFrom  *time.Time
	DateTo    *time.Time
	TimeOfDay []string // morning, afternoon, evening, night
	Season    []string // spring, summer, fall, winter

	// Equipment filters
	CameraMake  []string
	CameraModel []string
	LensMake    []string
	LensModel   []string

	// Technical filters (ranges)
	ISOMin             *int
	ISOMax             *int
	ApertureMin        *float64
	ApertureMax        *float64
	FocalLengthMin     *float64
	FocalLengthMax     *float64
	FocalLength35mmMin *int
	FocalLength35mmMax *int

	// Categorical filters
	FocalCategory     []string // wide, normal, telephoto
	ShootingCondition []string // bright, normal, low_light

	// Location filters
	LatMin *float64
	LatMax *float64
	LonMin *float64
	LonMax *float64
	HasGPS *bool

	// Colour filters
	ColourName []string // red, orange, yellow, green, blue, purple, pink, brown, grey, black, white
	ColourHex  *string  // exact colour with tolerance
	HueMin     *int     // 0-360
	HueMax     *int
	SatMin     *int // 0-100
	SatMax     *int
	LightMin   *int // 0-100
	LightMax   *int

	// Burst filters
	InBurst      *bool
	BurstGroupID *string
	IsBurstRep   *bool // only burst representatives

	// Image properties
	WidthMin         *int
	WidthMax         *int
	HeightMin        *int
	HeightMax        *int
	Orientation      *int     // 1=normal, 6=90CW, 8=90CCW, 3=180
	ImageOrientation []string // landscape, portrait, square
	IsLandscape      *bool
	IsPortrait       *bool

	// Other filters
	FlashFired   *bool
	WhiteBalance []string
	ColourSpace  []string

	// Pagination
	Limit  int
	Offset int

	// Sorting
	SortBy    string // date_taken, date_taken_desc, camera, focal_length, iso, aperture
	SortOrder string // asc, desc
}

// PhotoSummary is a lightweight photo representation for query results
type PhotoSummary struct {
	ID              int
	FilePath        string
	DateTaken       time.Time
	CameraMake      string
	CameraModel     string
	LensModel       string
	ISO             int
	Aperture        float64
	ShutterSpeed    string
	FocalLength     float64
	FocalLength35mm int
	Width           int
	Height          int
	TimeOfDay       string
	Season          string
	FocalCategory   string
	InBurst         bool
	BurstGroupID    string
	IndexedAt       time.Time // Used for cache busting in thumbnail URLs
	IsBurstRep      bool
	HasGPS          bool
	Latitude        float64
	Longitude       float64
}

// QueryResult contains the query results and metadata
type QueryResult struct {
	Photos      []PhotoSummary
	Total       int
	Limit       int
	Offset      int
	HasMore     bool
	Facets      *FacetCollection
	QueryTimeMs int64
}

// Facet represents a single facet dimension
type Facet struct {
	Name     string
	Label    string
	Values   []FacetValue
	Selected []string
}

// FacetValue represents a single value within a facet
type FacetValue struct {
	Value    string
	Label    string
	Count    int
	Selected bool
	URL      string

	// Camera-specific fields (to avoid string parsing bugs)
	CameraMake  string // Only populated for camera facets
	CameraModel string // Only populated for camera facets
}

// FacetCollection contains all available facets
type FacetCollection struct {
	Camera            *Facet
	Lens              *Facet
	Year              *Facet
	Month             *Facet
	TimeOfDay         *Facet
	Season            *Facet
	FocalCategory     *Facet
	ShootingCondition *Facet
	InBurst           *Facet
	ColourName        *Facet
	ImageOrientation  *Facet
	ISO               *Facet
	Aperture          *Facet
}

// RangeFilter represents a min/max range
type RangeFilter struct {
	Min *float64
	Max *float64
}

// ColourNameToHueRange maps colour names to hue ranges (degrees 0-360)
var ColourNameToHueRange = map[string][2]int{
	"red":    {0, 15}, // and 345-360
	"orange": {16, 45},
	"yellow": {46, 75},
	"green":  {76, 165},
	"blue":   {166, 255},
	"purple": {256, 290},
	"pink":   {291, 344},
	"brown":  {16, 45}, // same as orange but low saturation/lightness
	"grey":   {0, 360}, // any hue, very low saturation
	"black":  {0, 360}, // any hue, very low lightness
	"white":  {0, 360}, // any hue, very high lightness
}
