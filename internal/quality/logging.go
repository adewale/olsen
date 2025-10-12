package quality

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger handles structured logging of diagnostics
type Logger struct {
	file    *os.File
	enabled bool
}

// NewLogger creates a new diagnostics logger
// If logPath is empty, logging is disabled
func NewLogger(logPath string) (*Logger, error) {
	if logPath == "" {
		return &Logger{enabled: false}, nil
	}

	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file in append mode
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:    file,
		enabled: true,
	}, nil
}

// Log writes a diagnostic entry to the log
func (l *Logger) Log(diag *ImageDiag) error {
	if !l.enabled || l.file == nil {
		return nil
	}

	// Serialize to JSON
	data, err := json.Marshal(diag)
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics: %w", err)
	}

	// Write as single line
	_, err = fmt.Fprintf(l.file, "%s\n", data)
	if err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogToStderr logs diagnostics to stderr in a human-readable format
func LogToStderr(diag *ImageDiag) {
	log.Printf("[THUMB] %s | total=%.1fms decode=%.1fms orient=%.1fms resize=%.1fms encode=%.1fms | "+
		"SSIM=%.3f PSNR=%.1fdB sharpness=%.0f | warnings=%d",
		diag.ImgID,
		diag.TimingMS.Total,
		diag.TimingMS.Decode,
		diag.TimingMS.Orient,
		diag.TimingMS.Resize,
		diag.TimingMS.Encode,
		diag.Metrics.SSIMVsRef,
		diag.Metrics.PSNRVsRefDB,
		diag.Metrics.LapVar,
		len(diag.Warnings))

	if len(diag.Warnings) > 0 {
		for _, warning := range diag.Warnings {
			log.Printf("[THUMB] WARNING: %s: %s", diag.ImgID, warning)
		}
	}
}

// ArtifactManager handles saving QA artifacts
type ArtifactManager struct {
	baseDir string
	enabled bool
}

// NewArtifactManager creates a new artifact manager
func NewArtifactManager(baseDir string, enabled bool) (*ArtifactManager, error) {
	if !enabled || baseDir == "" {
		return &ArtifactManager{enabled: false}, nil
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifact directory: %w", err)
	}

	return &ArtifactManager{
		baseDir: baseDir,
		enabled: true,
	}, nil
}

// SaveArtifacts saves intermediate images and diagnostics for a sample
func (am *ArtifactManager) SaveArtifacts(imgID string, artifacts *Artifacts, diag *ImageDiag) error {
	if !am.enabled {
		return nil
	}

	// Create subdirectory by date
	dateDir := time.Now().Format("2006-01-02")
	targetDir := filepath.Join(am.baseDir, dateDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create date directory: %w", err)
	}

	// Sanitize imgID for filename
	safeID := filepath.Base(imgID)
	if len(safeID) > 50 {
		safeID = safeID[:50]
	}

	// Save diagnostics JSON
	diagPath := filepath.Join(targetDir, fmt.Sprintf("%s_diag.json", safeID))
	diagData, err := json.MarshalIndent(diag, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics: %w", err)
	}
	if err := os.WriteFile(diagPath, diagData, 0644); err != nil {
		return fmt.Errorf("failed to write diagnostics: %w", err)
	}

	// Save images if provided
	if artifacts != nil {
		if artifacts.AfterDecode != nil {
			if err := saveImage(targetDir, safeID, "decode", artifacts.AfterDecode); err != nil {
				log.Printf("Failed to save decode artifact: %v", err)
			}
		}
		if artifacts.AfterOrientColor != nil {
			if err := saveImage(targetDir, safeID, "after_orient_color", artifacts.AfterOrientColor); err != nil {
				log.Printf("Failed to save orient/color artifact: %v", err)
			}
		}
		if artifacts.Resized != nil {
			if err := saveImage(targetDir, safeID, "resized", artifacts.Resized); err != nil {
				log.Printf("Failed to save resize artifact: %v", err)
			}
		}
		if artifacts.Final != nil {
			finalPath := filepath.Join(targetDir, fmt.Sprintf("%s_final.jpg", safeID))
			if err := os.WriteFile(finalPath, artifacts.Final, 0644); err != nil {
				log.Printf("Failed to save final artifact: %v", err)
			}
		}
	}

	return nil
}

// Artifacts holds intermediate images for QA sampling
type Artifacts struct {
	AfterDecode      image.Image
	AfterOrientColor image.Image
	Resized          image.Image
	Final            []byte // Encoded thumbnail
}

func saveImage(dir, id, stage string, img image.Image) error {
	path := filepath.Join(dir, fmt.Sprintf("%s_%s.png", id, stage))
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return encodeImagePNG(file, img)
}

func encodeImagePNG(w *os.File, img image.Image) error {
	return png.Encode(w, img)
}
