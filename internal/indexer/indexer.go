// Package indexer implements the core photo indexing engine with concurrent processing.
//
// It extracts EXIF metadata, generates aspect-ratio-preserving thumbnails, analyzes
// color palettes, computes perceptual hashes, and infers additional metadata like
// time of day and season. The engine uses a worker pool pattern for parallel file
// processing and guarantees read-only access to photo files.
package indexer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adewale/olsen/internal/database"
	"github.com/adewale/olsen/internal/quality"
	"github.com/adewale/olsen/pkg/models"
)

// Processing limits to prevent resource exhaustion
const (
	maxImagePixels         = 100_000_000       // 100 megapixels
	maxFileSizeBytes       = 500 * 1024 * 1024 // 500 MB
	fileProcessingTimeout  = 60 * time.Second  // Timeout per file
	estimatedBytesPerPhoto = 250 * 1024        // 250KB estimate for database (metadata + thumbnails + colors)
)

// ProgressCallback is called with progress updates
type ProgressCallback func(processed, total int)

// Engine is the main indexer engine
type Engine struct {
	db               *database.DB
	workerCount      int
	stats            models.IndexStats
	mu               sync.Mutex
	progressCallback ProgressCallback
	perfTracking     bool
	perfStats        []models.PerfStats
	perfSummary      models.PerfSummary
	qualityConfig    quality.ThumbnailConfig
	qualityLogger    *quality.Logger
	artifactManager  *quality.ArtifactManager
}

// NewEngine creates a new indexer engine
func NewEngine(db *database.DB, workerCount int) *Engine {
	if workerCount <= 0 {
		workerCount = 4
	}

	// Initialize quality configuration from environment variables
	qualityConfig := quality.DefaultThumbnailConfig()

	// Check for QA sampling environment variable
	qaSampleRate := os.Getenv("THUMB_QA_SAMPLE")
	if qaSampleRate != "" {
		var rate float64
		if _, err := fmt.Sscanf(qaSampleRate, "%f", &rate); err == nil {
			qualityConfig.QASample = rate
			log.Printf("Quality sampling enabled: %.1f%%", rate*100)
		}
	}

	// Check for QA directory
	qaDir := os.Getenv("THUMB_QA_DIR")
	if qaDir != "" {
		qualityConfig.QADir = qaDir
	}

	// Check for disable artifacts flag
	if os.Getenv("THUMB_QA_DISABLE_ARTIFACTS") == "1" {
		qualityConfig.QADisableArtifacts = true
	}

	// Initialize quality logger
	logPath := os.Getenv("THUMB_LOG_PATH")
	qualityLogger, err := quality.NewLogger(logPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize quality logger: %v", err)
	}

	// Initialize artifact manager
	artifactManager, err := quality.NewArtifactManager(
		qualityConfig.QADir,
		qualityConfig.QASample > 0 && !qualityConfig.QADisableArtifacts,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize artifact manager: %v", err)
	}

	return &Engine{
		db:              db,
		workerCount:     workerCount,
		qualityConfig:   qualityConfig,
		qualityLogger:   qualityLogger,
		artifactManager: artifactManager,
		stats: models.IndexStats{
			StartTime: time.Now(),
		},
	}
}

// SetProgressCallback sets the progress callback function
func (e *Engine) SetProgressCallback(callback ProgressCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.progressCallback = callback
}

// EnablePerfTracking enables detailed performance tracking
func (e *Engine) EnablePerfTracking() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.perfTracking = true
	e.perfStats = make([]models.PerfStats, 0)
}

// IndexDirectory recursively indexes all DNG files in a directory
func (e *Engine) IndexDirectory(rootPath string) error {
	return e.IndexDirectoryContext(context.Background(), rootPath)
}

// IndexDirectoryContext recursively indexes all supported files in a
// directory. Cancelling ctx stops feeding work to the pool; in-flight files
// are abandoned at the next pipeline stage boundary and the function returns
// ctx.Err(). Already-committed photos are unaffected (each photo is a single
// transaction), so an interrupted run can simply be re-run.
func (e *Engine) IndexDirectoryContext(ctx context.Context, rootPath string) error {
	log.Printf("Starting indexing of %s with %d workers\n", rootPath, e.workerCount)

	// Find all DNG files
	files, err := e.findDNGFiles(rootPath)
	if err != nil {
		return fmt.Errorf("failed to find DNG files: %w", err)
	}

	e.mu.Lock()
	e.stats.FilesFound = len(files)
	e.mu.Unlock()

	log.Printf("Found %d DNG files\n", len(files))

	if len(files) == 0 {
		return nil
	}

	// Check disk space before starting
	estimatedSize := uint64(len(files)) * estimatedBytesPerPhoto
	dbPath := e.db.GetPath()
	if err := checkDiskSpace(dbPath, estimatedSize); err != nil {
		return err
	}

	// Create work channel and worker pool
	workChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < e.workerCount; i++ {
		wg.Add(1)
		go e.worker(ctx, i, workChan, &wg)
	}

	// Send work to workers, stopping early if cancelled
feed:
	for _, file := range files {
		select {
		case workChan <- file:
		case <-ctx.Done():
			break feed
		}
	}
	close(workChan)

	// Wait for all workers to finish
	wg.Wait()

	e.mu.Lock()
	e.stats.EndTime = time.Now()
	e.mu.Unlock()

	if err := ctx.Err(); err != nil {
		log.Printf("Indexing interrupted: %v", err)
		return err
	}

	// Print summary
	log.Printf("\nIndexing complete!")
	log.Printf("  Files found: %d\n", e.stats.FilesFound)
	log.Printf("  Files processed: %d\n", e.stats.FilesProcessed)
	log.Printf("  Files skipped: %d\n", e.stats.FilesSkipped)
	log.Printf("  Files updated: %d\n", e.stats.FilesUpdated)
	log.Printf("  Files failed: %d\n", e.stats.FilesFailed)
	log.Printf("  Thumbnails generated: %d\n", e.stats.ThumbnailsGenerated)
	log.Printf("  Duration: %v\n", e.stats.Duration())
	log.Printf("  Rate: %.2f photos/second\n", e.stats.PhotosPerSecond())

	return nil
}

// worker processes files from the work channel
func (e *Engine) worker(ctx context.Context, id int, workChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range workChan {
		// Drain remaining work without processing once cancelled
		if ctx.Err() != nil {
			continue
		}

		// Process file with timeout
		perfStats, err := e.processFileWithTimeout(ctx, filePath)
		if err != nil && ctx.Err() != nil {
			continue // interrupted, not a real failure
		}
		if err != nil {
			log.Printf("Worker %d: Failed to process %s: %v\n", id, filePath, err)
		}

		e.mu.Lock()
		if err != nil {
			e.stats.FilesFailed++
			if e.perfTracking {
				perfStats.Error = err.Error()
				e.perfStats = append(e.perfStats, perfStats)
				e.perfSummary.FailedPhotos++
			}
		} else {
			// Skipped files are counted in FilesSkipped (inside processFile),
			// not double-counted as processed.
			if !perfStats.WasSkipped {
				e.stats.FilesProcessed++
			}
			if e.perfTracking {
				e.perfStats = append(e.perfStats, perfStats)
				e.updatePerfSummary(perfStats)
			}
		}

		// Progress covers every file with a final disposition
		// (processed, skipped, or failed).
		done := e.stats.FilesProcessed + e.stats.FilesSkipped + e.stats.FilesFailed
		total := e.stats.FilesFound
		callback := e.progressCallback

		// Report progress every 100 files (legacy logging)
		if done%100 == 0 {
			log.Printf("Progress: %d/%d files processed (%.1f%%)\n",
				done, total, float64(done)/float64(total)*100)
		}
		e.mu.Unlock()

		// Call progress callback if set
		if callback != nil {
			callback(done, total)
		}
	}
}

// processFileWithTimeout wraps processFile with a per-file timeout. The
// timeout cancels the context handed to processFile, so the worker goroutine
// aborts at its next stage boundary instead of continuing (and inserting into
// the database) after the file has already been reported as failed.
func (e *Engine) processFileWithTimeout(ctx context.Context, filePath string) (models.PerfStats, error) {
	type result struct {
		perf models.PerfStats
		err  error
	}

	ctx, cancel := context.WithTimeout(ctx, fileProcessingTimeout)
	defer cancel()

	resultChan := make(chan result, 1)

	// Run processing in goroutine
	go func() {
		perf, err := e.processFile(ctx, filePath)
		resultChan <- result{perf, err}
	}()

	// Wait for completion or timeout
	select {
	case res := <-resultChan:
		return res.perf, res.err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return models.PerfStats{FilePath: filePath},
				fmt.Errorf("⏱️  timeout after %v", fileProcessingTimeout)
		}
		return models.PerfStats{FilePath: filePath}, ctx.Err()
	}
}

// processFile processes a single DNG file. It checks ctx at each pipeline
// stage boundary so a timeout or interrupt stops the work (most importantly,
// before the database insert) instead of running to completion in the
// background.
func (e *Engine) processFile(ctx context.Context, filePath string) (models.PerfStats, error) {
	startTime := time.Now()
	perf := models.PerfStats{
		FilePath: filePath,
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		perf.FileSize = fileInfo.Size()
	}

	// Calculate file hash first to check if file has changed
	hashStart := time.Now()
	currentHash, err := calculateFileHash(filePath)
	perf.HashTime = time.Since(hashStart)
	if err != nil {
		return perf, fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Check if already indexed
	exists, err := e.db.PhotoExists(filePath)
	if err != nil {
		return perf, fmt.Errorf("failed to check if photo exists: %w", err)
	}

	if exists {
		// Check if file has been modified by comparing hashes
		existingHash, err := e.db.GetPhotoHash(filePath)
		if err != nil {
			return perf, fmt.Errorf("failed to get existing photo hash: %w", err)
		}

		if existingHash == currentHash {
			// File unchanged, skip
			e.mu.Lock()
			e.stats.FilesSkipped++
			e.mu.Unlock()
			perf.WasSkipped = true
			perf.TotalTime = time.Since(startTime)
			return perf, nil
		}

		// File has been modified, delete the old entry and re-index
		log.Printf("File modified, re-indexing: %s", filePath)
		if err := e.db.DeletePhoto(filePath); err != nil {
			return perf, fmt.Errorf("failed to delete old photo entry: %w", err)
		}
		e.mu.Lock()
		e.stats.FilesUpdated++
		e.mu.Unlock()
		perf.WasUpdated = true
	}

	if err := ctx.Err(); err != nil {
		return perf, err
	}

	// Check if this is a RAW file
	ext := strings.ToLower(filepath.Ext(filePath))
	isRawFile := ext == ".dng" || ext == ".cr2" || ext == ".nef" || ext == ".raf" || ext == ".arw"

	var metadata *models.PhotoMetadata
	var img image.Image

	// Metadata extraction
	metadataStart := time.Now()

	// Extract EXIF metadata using go-exif (works for both RAW and JPEG)
	metadata, err = ExtractMetadata(filePath)
	if err != nil {
		// If EXIF extraction fails, create basic metadata from file info
		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil {
			return perf, fmt.Errorf("failed to stat file: %w", statErr)
		}

		metadata = &models.PhotoMetadata{
			FilePath:     filePath,
			FileSize:     fileInfo.Size(),
			LastModified: fileInfo.ModTime(),
			IndexedAt:    time.Now(),
		}
	}

	// Use the hash we already calculated
	metadata.FileHash = currentHash
	perf.MetadataTime = time.Since(metadataStart)

	// Check file size limits before decoding
	if metadata.FileSize > maxFileSizeBytes {
		log.Printf("⚠️  Skipping oversized file: %s (%.1f MB exceeds %.1f MB limit)",
			filepath.Base(filePath),
			float64(metadata.FileSize)/1024/1024,
			float64(maxFileSizeBytes)/1024/1024)

		// Store metadata-only, no thumbnails
		if err := e.db.InsertPhoto(metadata); err != nil {
			return perf, fmt.Errorf("failed to insert photo metadata: %w", err)
		}
		perf.TotalTime = time.Since(startTime)
		return perf, nil
	}

	// Check image dimensions if available (after EXIF extraction)
	if metadata.Width > 0 && metadata.Height > 0 {
		pixels := int64(metadata.Width) * int64(metadata.Height)
		if pixels > maxImagePixels {
			log.Printf("⚠️  Skipping oversized image: %s (%dx%d = %.1f MP exceeds %.0f MP limit)",
				filepath.Base(filePath),
				metadata.Width, metadata.Height,
				float64(pixels)/1_000_000,
				float64(maxImagePixels)/1_000_000)

			// Store metadata-only, no thumbnails
			if err := e.db.InsertPhoto(metadata); err != nil {
				return perf, fmt.Errorf("failed to insert photo metadata: %w", err)
			}
			perf.TotalTime = time.Since(startTime)
			return perf, nil
		}
	}

	// Image decoding
	decodeStart := time.Now()

	if err := ctx.Err(); err != nil {
		return perf, err
	}

	// Try RAW decode if applicable
	if isRawFile && IsRawSupported() {
		// Try to decode RAW image
		var decodeErr error
		img, decodeErr = DecodeRaw(filePath)
		if decodeErr != nil {
			log.Printf("RAW image decode failed for %s: %v, trying embedded JPEG", filepath.Base(filePath), decodeErr)
			// Try to extract embedded JPEG preview as fallback
			img, decodeErr = ExtractEmbeddedJPEG(filePath)
			if decodeErr != nil {
				log.Printf("Embedded JPEG extraction also failed for %s: %v, will use metadata-only", filepath.Base(filePath), decodeErr)
			} else {
				log.Printf("Successfully extracted embedded JPEG preview for %s", filepath.Base(filePath))
			}
		}
	}

	// If we don't have an image yet, try standard image decode
	if img == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return perf, fmt.Errorf("failed to open image: %w", err)
		}
		defer file.Close()

		// Guard the decode with a cheap header-only dimension check. The
		// EXIF-based megapixel check above only works when EXIF dimensions
		// are present; without this a crafted or EXIF-less image could force
		// an arbitrarily large allocation.
		if cfg, _, err := image.DecodeConfig(file); err == nil {
			pixels := int64(cfg.Width) * int64(cfg.Height)
			if pixels > maxImagePixels {
				log.Printf("⚠️  Skipping oversized image: %s (%dx%d = %.1f MP exceeds %.0f MP limit)",
					filepath.Base(filePath),
					cfg.Width, cfg.Height,
					float64(pixels)/1_000_000,
					float64(maxImagePixels)/1_000_000)

				// Store metadata-only, no thumbnails
				if err := e.db.InsertPhoto(metadata); err != nil {
					return perf, fmt.Errorf("failed to insert photo metadata: %w", err)
				}
				perf.TotalTime = time.Since(startTime)
				return perf, nil
			}
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return perf, fmt.Errorf("failed to rewind image file: %w", err)
		}

		var decodeErr error
		img, _, decodeErr = image.Decode(file)
		if decodeErr != nil {
			// For RAW files that can't be decoded, we can still store metadata
			if isRawFile {
				log.Printf("RAW file %s indexed with metadata only (no thumbnail)", filepath.Base(filePath))
				perf.ImageDecodeTime = time.Since(decodeStart)

				// Store metadata without thumbnails/colours
				dbStart := time.Now()
				if err := e.db.InsertPhoto(metadata); err != nil {
					return perf, fmt.Errorf("failed to insert photo: %w", err)
				}
				perf.DatabaseTime = time.Since(dbStart)
				perf.TotalTime = time.Since(startTime)
				return perf, nil
			}
			return perf, fmt.Errorf("failed to decode image: %w", decodeErr)
		}
	}
	perf.ImageDecodeTime = time.Since(decodeStart)

	// Backfill dimensions from the decoded image when EXIF lacked them
	if metadata.Width == 0 || metadata.Height == 0 {
		metadata.Width = img.Bounds().Dx()
		metadata.Height = img.Bounds().Dy()
	}

	if err := ctx.Err(); err != nil {
		return perf, err
	}

	// Generate thumbnails with quality instrumentation
	thumbnailStart := time.Now()

	// Prepare image metadata for quality pipeline
	imgMeta := quality.ImageMetadata{
		FilePath:      filePath,
		Orientation:   metadata.Orientation,
		ColorSpace:    metadata.ColourSpace,
		HasICCProfile: false, // TODO: Detect ICC profiles
		Width:         img.Bounds().Dx(),
		Height:        img.Bounds().Dy(),
	}

	// Generate thumbnails with diagnostics
	thumbnails, diag, err := quality.GenerateThumbnailsWithDiag(ctx, img, imgMeta, e.qualityConfig)
	if err != nil {
		return perf, fmt.Errorf("failed to generate thumbnails: %w", err)
	}

	// If no thumbnails were generated (e.g., image too small, upscaling prevented),
	// store the original image as the tiny thumbnail
	if len(thumbnails) == 0 {
		log.Printf("No thumbnails generated for %s (image too small), storing original as TINY thumbnail", filepath.Base(filePath))
		// Encode original image as JPEG
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return perf, fmt.Errorf("failed to encode original as thumbnail: %w", err)
		}
		thumbnails = map[models.ThumbnailSize][]byte{
			models.ThumbnailTiny: buf.Bytes(),
		}
	}

	metadata.Thumbnails = thumbnails
	perf.ThumbnailTime = time.Since(thumbnailStart)

	e.mu.Lock()
	e.stats.ThumbnailsGenerated += len(thumbnails)
	e.mu.Unlock()

	// Log diagnostics if logger is enabled
	if e.qualityLogger != nil {
		if err := e.qualityLogger.Log(diag); err != nil {
			log.Printf("Warning: Failed to log quality diagnostics: %v", err)
		}
	}

	// Log warnings to stderr if any
	if len(diag.Warnings) > 0 {
		quality.LogToStderr(diag)
	}

	// Extract color palette from the smallest available thumbnail for efficiency
	colorStart := time.Now()

	// Find the smallest available thumbnail (in case some were skipped due to upscaling)
	var thumbData []byte
	for _, size := range []models.ThumbnailSize{models.ThumbnailTiny, models.ThumbnailSmall, models.ThumbnailMedium, models.ThumbnailLarge} {
		if data, ok := thumbnails[size]; ok && len(data) > 0 {
			thumbData = data
			break
		}
	}

	var thumbImg image.Image
	if len(thumbData) == 0 {
		// No thumbnails available, use original image for color extraction and perceptual hash
		thumbImg = img
		colours, err := ExtractColourPalette(img, 5)
		if err != nil {
			return perf, fmt.Errorf("failed to extract colours from original image: %w", err)
		}
		metadata.DominantColours = colours
	} else {
		// Decode thumbnail for color extraction
		var err error
		thumbImg, _, err = image.Decode(bytes.NewReader(thumbData))
		if err != nil {
			return perf, fmt.Errorf("failed to decode thumbnail for color extraction: %w", err)
		}

		colours, err := ExtractColourPalette(thumbImg, 5)
		if err != nil {
			return perf, fmt.Errorf("failed to extract colours: %w", err)
		}
		metadata.DominantColours = colours
	}
	perf.ColorTime = time.Since(colorStart)

	// Compute perceptual hash
	phashStart := time.Now()
	phash, err := ComputePerceptualHash(thumbImg)
	if err != nil {
		return perf, fmt.Errorf("failed to compute perceptual hash: %w", err)
	}
	metadata.PerceptualHash = phash
	perf.PerceptualHashTime = time.Since(phashStart)

	e.mu.Lock()
	e.stats.HashesComputed++
	e.mu.Unlock()

	// Infer metadata
	inferStart := time.Now()
	InferMetadata(metadata)
	perf.InferenceTime = time.Since(inferStart)

	// Last boundary check before the write: a timed-out file must not be
	// inserted after it was already reported as failed.
	if err := ctx.Err(); err != nil {
		return perf, err
	}

	// Store in database
	dbStart := time.Now()
	if err := e.db.InsertPhoto(metadata); err != nil {
		return perf, fmt.Errorf("failed to insert photo: %w", err)
	}
	perf.DatabaseTime = time.Since(dbStart)

	perf.TotalTime = time.Since(startTime)
	return perf, nil
}

// findDNGFiles recursively finds all supported image files in a directory
// Supports: DNG, JPEG, JPG, BMP
func (e *Engine) findDNGFiles(rootPath string) ([]string, error) {
	var files []string

	supportedExts := map[string]bool{
		".dng":  true,
		".jpg":  true,
		".jpeg": true,
		".bmp":  true,
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log and continue: one unreadable directory must not abort
			// indexing of an entire library.
			log.Printf("⚠️  Skipping %s: %v", path, err)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// calculateFileHash calculates SHA-256 hash of a file
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetStats returns the current indexing statistics
func (e *Engine) GetStats() models.IndexStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stats
}

// GetPerfStats returns the collected performance statistics
func (e *Engine) GetPerfStats() []models.PerfStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.perfStats
}

// GetPerfSummary returns the aggregate performance summary
func (e *Engine) GetPerfSummary() models.PerfSummary {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.perfSummary
}

// Close closes the engine and any associated resources
func (e *Engine) Close() error {
	if e.qualityLogger != nil {
		return e.qualityLogger.Close()
	}
	return nil
}

// updatePerfSummary updates the running performance summary (must be called with lock held)
func (e *Engine) updatePerfSummary(perf models.PerfStats) {
	e.perfSummary.TotalPhotos++

	if perf.WasSkipped {
		e.perfSummary.SkippedPhotos++
		return // Don't include skipped files in processing stats
	}

	if perf.WasUpdated {
		e.perfSummary.UpdatedPhotos++
	}

	e.perfSummary.ProcessedPhotos++

	// Add to totals
	e.perfSummary.TotalTime += perf.TotalTime
	e.perfSummary.HashTime += perf.HashTime
	e.perfSummary.MetadataTime += perf.MetadataTime
	e.perfSummary.ImageDecodeTime += perf.ImageDecodeTime
	e.perfSummary.ThumbnailTime += perf.ThumbnailTime
	e.perfSummary.ColorTime += perf.ColorTime
	e.perfSummary.PerceptualHashTime += perf.PerceptualHashTime
	e.perfSummary.InferenceTime += perf.InferenceTime
	e.perfSummary.DatabaseTime += perf.DatabaseTime
	e.perfSummary.TotalBytes += perf.FileSize

	// Calculate running averages (only for processed photos)
	n := float64(e.perfSummary.ProcessedPhotos)
	e.perfSummary.AvgTotalMs = float64(e.perfSummary.TotalTime.Milliseconds()) / n
	e.perfSummary.AvgHashMs = float64(e.perfSummary.HashTime.Milliseconds()) / n
	e.perfSummary.AvgMetadataMs = float64(e.perfSummary.MetadataTime.Milliseconds()) / n
	e.perfSummary.AvgImageDecodeMs = float64(e.perfSummary.ImageDecodeTime.Milliseconds()) / n
	e.perfSummary.AvgThumbnailMs = float64(e.perfSummary.ThumbnailTime.Milliseconds()) / n
	e.perfSummary.AvgColorMs = float64(e.perfSummary.ColorTime.Milliseconds()) / n
	e.perfSummary.AvgPerceptualHashMs = float64(e.perfSummary.PerceptualHashTime.Milliseconds()) / n
	e.perfSummary.AvgInferenceMs = float64(e.perfSummary.InferenceTime.Milliseconds()) / n
	e.perfSummary.AvgDatabaseMs = float64(e.perfSummary.DatabaseTime.Milliseconds()) / n

	// Calculate throughput (MB/s)
	if e.perfSummary.TotalTime > 0 {
		totalMB := float64(e.perfSummary.TotalBytes) / (1024 * 1024)
		totalSec := e.perfSummary.TotalTime.Seconds()
		e.perfSummary.AvgThroughputMBps = totalMB / totalSec
	}
}

// checkDiskSpace checks if there's sufficient disk space for indexing
func checkDiskSpace(dbPath string, estimatedBytes uint64) error {
	available, err := availableDiskSpace(filepath.Dir(dbPath))
	if err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}
	if available == 0 {
		// Platform without a disk space probe; skip the pre-flight check.
		return nil
	}

	// Need 20% safety margin
	needed := uint64(float64(estimatedBytes) * 1.2)

	if available < needed {
		return fmt.Errorf("insufficient disk space: need %.1f GB, have %.1f GB available",
			float64(needed)/(1024*1024*1024),
			float64(available)/(1024*1024*1024))
	}

	log.Printf("Disk space check: %.1f GB needed, %.1f GB available ✓",
		float64(needed)/(1024*1024*1024),
		float64(available)/(1024*1024*1024))

	return nil
}
