package quality

import (
	"encoding/json"
	"time"
)

// RawDiag contains diagnostics about RAW file decode
type RawDiag struct {
	LibRawEnabled bool   `json:"libraw_enabled"`
	Demosaic      string `json:"demosaic"`     // "AHD","DCB","PPG","unknown"
	OutputBPS     int    `json:"output_bps"`   // 8|16
	OutputColor   string `json:"output_color"` // "sRGB","AdobeRGB","linear","unknown"
	UseCameraWB   bool   `json:"use_camera_wb"`
	HalfSize      bool   `json:"half_size"`
}

// SharpenDiag contains sharpening configuration
type SharpenDiag struct {
	Enabled bool    `json:"enabled"`
	Amount  float64 `json:"amount"`
	Radius  float64 `json:"radius"`
}

// ResizeDiag contains diagnostics about resize operation
type ResizeDiag struct {
	TargetLongEdge int         `json:"target_long_edge"`
	Filter         string      `json:"filter"`
	PreSharpen     SharpenDiag `json:"pre_sharpen"`
	PostSharpen    SharpenDiag `json:"post_sharpen"`
	Upscale        bool        `json:"upscale"`
}

// EncodeDiag contains diagnostics about encoding
type EncodeDiag struct {
	Format      string `json:"format"` // "jpeg","webp","avif"
	Quality     int    `json:"quality"`
	Chroma      string `json:"chroma"` // "420","422","444"
	Progressive bool   `json:"progressive"`
	Bytes       int    `json:"bytes"`
}

// PipelineDiag contains diagnostics about the processing pipeline
type PipelineDiag struct {
	OrientationApplied bool       `json:"orientation_applied"`
	ColorspaceIn       string     `json:"colorspace_in"`
	ColorspaceOut      string     `json:"colorspace_out"`
	GammaLinearized    bool       `json:"gamma_linearized"`
	Resize             ResizeDiag `json:"resize"`
	Encode             EncodeDiag `json:"encode"`
}

// SourceDiag contains diagnostics about the source image
type SourceDiag struct {
	Format          string  `json:"format"`
	InputW          int     `json:"input_w"`
	InputH          int     `json:"input_h"`
	HasICC          bool    `json:"has_icc"`
	ICCDesc         string  `json:"icc_desc"`
	EXIFOrientation int     `json:"exif_orientation"`
	Raw             RawDiag `json:"raw,omitempty"`
	FallbackReason  string  `json:"fallback_reason"` // "none|no_cgo|decode_error|no_raw|embedded_only"
}

// MetricsDiag contains quality metrics
type MetricsDiag struct {
	SSIMVsRef         float64 `json:"ssim_vs_ref"`
	PSNRVsRefDB       float64 `json:"psnr_vs_ref_db"`
	LapVar            float64 `json:"lap_var"`
	DeltaEMean        float64 `json:"delta_e_mean"`
	ClippedPixelsLow  int     `json:"clipped_pixels_low"`
	ClippedPixelsHigh int     `json:"clipped_pixels_high"`
	BandingScore      float64 `json:"histogram_banding_score"`
}

// TimingDiag contains timing information for each stage
type TimingDiag struct {
	Decode  float64 `json:"decode"`
	Orient  float64 `json:"orient"`
	Color   float64 `json:"color"`
	Resize  float64 `json:"resize"`
	Sharpen float64 `json:"sharpen"`
	Encode  float64 `json:"encode"`
	Store   float64 `json:"store"`
	Total   float64 `json:"total"`
}

// VersionDiag contains version information
type VersionDiag struct {
	ThumbPipeline string `json:"thumb_pipeline"`
	LibRaw        string `json:"libraw"`
	Encoder       string `json:"encoder"`
}

// ImageDiag contains comprehensive diagnostics for a single thumbnail generation
type ImageDiag struct {
	ImgID       string       `json:"img_id"`
	Source      SourceDiag   `json:"source"`
	Pipeline    PipelineDiag `json:"pipeline"`
	Metrics     MetricsDiag  `json:"metrics"`
	TimingMS    TimingDiag   `json:"timing_ms"`
	Warnings    []string     `json:"warnings"`
	Version     VersionDiag  `json:"version"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// ToJSON serializes the diagnostics to JSON
func (d *ImageDiag) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// ToJSONString serializes the diagnostics to a JSON string
func (d *ImageDiag) ToJSONString() (string, error) {
	b, err := d.ToJSON()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromJSON deserializes diagnostics from JSON
func FromJSON(data []byte) (*ImageDiag, error) {
	var diag ImageDiag
	err := json.Unmarshal(data, &diag)
	return &diag, err
}

// AddWarning adds a warning to the diagnostics
func (d *ImageDiag) AddWarning(warning string) {
	d.Warnings = append(d.Warnings, warning)
}

// HasWarnings returns true if there are any warnings
func (d *ImageDiag) HasWarnings() bool {
	return len(d.Warnings) > 0
}

// IsUpscale returns true if the image was upscaled
func (d *ImageDiag) IsUpscale() bool {
	return d.Pipeline.Resize.Upscale
}

// IsFallback returns true if a fallback decode path was used
func (d *ImageDiag) IsFallback() bool {
	return d.Source.FallbackReason != "none" && d.Source.FallbackReason != ""
}

// NewImageDiag creates a new ImageDiag with default values
func NewImageDiag(imgID string) *ImageDiag {
	return &ImageDiag{
		ImgID:       imgID,
		Warnings:    []string{},
		GeneratedAt: time.Now(),
		Version: VersionDiag{
			ThumbPipeline: "2.0.0",
		},
		Source: SourceDiag{
			FallbackReason: "none",
		},
	}
}
