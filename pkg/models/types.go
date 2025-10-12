// Package models defines the core data structures used throughout the Olsen photo indexer.
//
// It includes photo metadata, thumbnail specifications, color representations, and
// statistics tracking types that are shared between the indexer, database, and query
// components.
package models

import (
	"time"
)

// ThumbnailSize represents the longest edge dimension for thumbnails
type ThumbnailSize string

const (
	ThumbnailTiny   ThumbnailSize = "64"   // Grid view (longest edge)
	ThumbnailSmall  ThumbnailSize = "256"  // List view (longest edge)
	ThumbnailMedium ThumbnailSize = "512"  // Preview (longest edge)
	ThumbnailLarge  ThumbnailSize = "1024" // Large preview (longest edge)
)

// Colour represents an RGB colour
type Colour struct {
	R, G, B uint8
}

// ColourHSL represents a colour in HSL colour space
type ColourHSL struct {
	H int // 0-360 degrees
	S int // 0-100%
	L int // 0-100%
}

// DominantColour represents a colour with its weight in the image
type DominantColour struct {
	Colour Colour
	HSL    ColourHSL
	Weight float64
}

// PhotoMetadata contains all metadata for a photo
type PhotoMetadata struct {
	ID           int
	FilePath     string
	FileHash     string
	FileSize     int64
	LastModified time.Time
	IndexedAt    time.Time

	// Camera & Lens
	CameraMake  string
	CameraModel string
	LensMake    string
	LensModel   string

	// Exposure Settings
	ISO                  int
	Aperture             float64
	ShutterSpeed         string
	ExposureCompensation float64
	FocalLength          float64
	FocalLength35mm      int

	// Temporal
	DateTaken     time.Time
	DateDigitized time.Time

	// Image Properties
	Width       int
	Height      int
	Orientation int
	ColourSpace string

	// Location
	Latitude  float64
	Longitude float64
	Altitude  float64

	// DNG-Specific
	DNGVersion          string
	OriginalRawFilename string

	// Lighting
	FlashFired    bool
	WhiteBalance  string
	FocusDistance float64

	// Inferred Metadata
	TimeOfDay         string
	Season            string
	FocalCategory     string
	ShootingCondition string

	// Visual Analysis
	Thumbnails      map[ThumbnailSize][]byte
	DominantColours []DominantColour

	// Perceptual Hash
	PerceptualHash string

	// Burst Detection
	BurstGroupID          string
	BurstSequence         int
	BurstCount            int
	IsBurstRepresentative bool

	// Duplicate Clustering
	DuplicateClusterID      string
	ClusterSize             int
	IsClusterRepresentative bool
	SimilarityScore         float64
}

// IndexStats tracks indexing progress
type IndexStats struct {
	FilesFound          int
	FilesProcessed      int
	FilesSkipped        int
	FilesUpdated        int
	FilesFailed         int
	ThumbnailsGenerated int
	HashesComputed      int
	StartTime           time.Time
	EndTime             time.Time
}

// Duration returns the total indexing duration
func (s *IndexStats) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

// PhotosPerSecond returns the indexing rate
func (s *IndexStats) PhotosPerSecond() float64 {
	duration := s.Duration()
	if duration == 0 {
		return 0
	}
	return float64(s.FilesProcessed) / duration.Seconds()
}

// PerfStats tracks detailed performance metrics for a single photo
type PerfStats struct {
	FilePath           string
	TotalTime          time.Duration
	HashTime           time.Duration
	MetadataTime       time.Duration
	ImageDecodeTime    time.Duration
	ThumbnailTime      time.Duration
	ColorTime          time.Duration
	PerceptualHashTime time.Duration
	InferenceTime      time.Duration
	DatabaseTime       time.Duration
	FileSize           int64
	WasSkipped         bool
	WasUpdated         bool
	Error              string
}

// PerfSummary tracks aggregate performance statistics
type PerfSummary struct {
	TotalPhotos     int
	ProcessedPhotos int
	SkippedPhotos   int
	UpdatedPhotos   int
	FailedPhotos    int

	TotalTime          time.Duration
	HashTime           time.Duration
	MetadataTime       time.Duration
	ImageDecodeTime    time.Duration
	ThumbnailTime      time.Duration
	ColorTime          time.Duration
	PerceptualHashTime time.Duration
	InferenceTime      time.Duration
	DatabaseTime       time.Duration

	TotalBytes int64

	// Running averages (milliseconds)
	AvgTotalMs          float64
	AvgHashMs           float64
	AvgMetadataMs       float64
	AvgImageDecodeMs    float64
	AvgThumbnailMs      float64
	AvgColorMs          float64
	AvgPerceptualHashMs float64
	AvgInferenceMs      float64
	AvgDatabaseMs       float64

	AvgThroughputMBps float64
}
