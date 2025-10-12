package indexer

import (
	"time"

	"github.com/adewale/olsen/internal/database"
)

// BurstDetector detects burst photo sequences
type BurstDetector struct {
	db            *database.DB
	maxTimeDelta  time.Duration // Maximum time between burst photos
	maxFocalDelta float64       // Maximum focal length difference (mm)
	minBurstSize  int           // Minimum photos to qualify as burst
}

// NewBurstDetector creates a new burst detector with default settings
func NewBurstDetector(db *database.DB) *BurstDetector {
	return &BurstDetector{
		db:            db,
		maxTimeDelta:  2 * time.Second, // Per spec: within 2 seconds
		maxFocalDelta: 5.0,             // Per spec: ±5mm focal length
		minBurstSize:  3,               // Per spec: 3+ photos
	}
}

// Photo represents a photo for burst detection
type Photo struct {
	ID          int
	FilePath    string
	DateTaken   time.Time
	CameraMake  string
	CameraModel string
	FocalLength float64
}

// DetectBursts finds all burst sequences in the database
func (bd *BurstDetector) DetectBursts() ([][]int, error) {
	// Query all photos ordered by date
	rows, err := bd.db.Query(`
		SELECT id, file_path, date_taken, camera_make, camera_model, focal_length
		FROM photos
		WHERE date_taken IS NOT NULL
		ORDER BY date_taken
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var p Photo
		var dateTakenStr string
		err := rows.Scan(&p.ID, &p.FilePath, &dateTakenStr, &p.CameraMake, &p.CameraModel, &p.FocalLength)
		if err != nil {
			return nil, err
		}

		// Parse date
		p.DateTaken, err = time.Parse("2006-01-02 15:04:05", dateTakenStr)
		if err != nil {
			// Try alternative format
			p.DateTaken, err = time.Parse(time.RFC3339, dateTakenStr)
			if err != nil {
				continue // Skip photos with unparseable dates
			}
		}

		photos = append(photos, p)
	}

	// Detect bursts
	bursts := bd.findBurstSequences(photos)

	return bursts, nil
}

// findBurstSequences finds burst sequences in a list of photos
func (bd *BurstDetector) findBurstSequences(photos []Photo) [][]int {
	if len(photos) < bd.minBurstSize {
		return nil
	}

	var bursts [][]int
	i := 0

	for i < len(photos) {
		burst := []int{i}

		// Try to extend burst from this starting point
		for j := i + 1; j < len(photos); j++ {
			if bd.canExtendBurst(photos, burst, j) {
				burst = append(burst, j)
			}
		}

		// If we found a valid burst, record it
		if len(burst) >= bd.minBurstSize {
			// Convert indices to photo IDs
			burstIDs := make([]int, len(burst))
			for k, idx := range burst {
				burstIDs[k] = photos[idx].ID
			}
			bursts = append(bursts, burstIDs)

			// Skip past this burst
			i = burst[len(burst)-1] + 1
		} else {
			i++
		}
	}

	return bursts
}

// canExtendBurst checks if a photo can extend an existing burst sequence
func (bd *BurstDetector) canExtendBurst(photos []Photo, burst []int, candidateIdx int) bool {
	lastIdx := burst[len(burst)-1]
	last := photos[lastIdx]
	candidate := photos[candidateIdx]

	// Check time delta (from last photo in burst)
	timeDelta := candidate.DateTaken.Sub(last.DateTaken)
	if timeDelta < 0 || timeDelta > bd.maxTimeDelta {
		return false
	}

	// Check camera match (must be same camera)
	if candidate.CameraMake != last.CameraMake || candidate.CameraModel != last.CameraModel {
		return false
	}

	// Check focal length (must be within tolerance)
	focalDelta := abs(candidate.FocalLength - last.FocalLength)
	if focalDelta > bd.maxFocalDelta {
		return false
	}

	return true
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// SaveBursts saves detected burst groups to the database
func (bd *BurstDetector) SaveBursts(bursts [][]int) error {
	for burstIdx, burst := range bursts {
		if len(burst) == 0 {
			continue
		}

		// Generate burst group ID
		burstGroupID := time.Now().Format("20060102150405") + "_" + string(rune('0'+burstIdx))

		// Get first photo's date for burst group metadata
		var dateTaken string
		var cameraMake, cameraModel string
		err := bd.db.QueryRow(`
			SELECT date_taken, camera_make, camera_model
			FROM photos WHERE id = ?
		`, burst[0]).Scan(&dateTaken, &cameraMake, &cameraModel)
		if err != nil {
			return err
		}

		// Calculate time span (in seconds)
		var lastDateTaken string
		err = bd.db.QueryRow(`
			SELECT date_taken FROM photos WHERE id = ?
		`, burst[len(burst)-1]).Scan(&lastDateTaken)
		if err != nil {
			return err
		}

		firstTime, _ := time.Parse("2006-01-02 15:04:05", dateTaken)
		lastTime, _ := time.Parse("2006-01-02 15:04:05", lastDateTaken)
		timeSpan := lastTime.Sub(firstTime).Seconds()

		// Insert burst group
		_, err = bd.db.Exec(`
			INSERT INTO burst_groups (
				id, photo_count, date_taken, camera_make, camera_model,
				representative_photo_id, time_span_seconds
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`, burstGroupID, len(burst), dateTaken, cameraMake, cameraModel, burst[0], timeSpan)
		if err != nil {
			return err
		}

		// Update photos with burst metadata
		for position, photoID := range burst {
			_, err := bd.db.Exec(`
				UPDATE photos
				SET burst_group_id = ?,
				    burst_sequence = ?,
				    burst_count = ?,
				    is_burst_representative = ?
				WHERE id = ?
			`, burstGroupID, position, len(burst), position == 0, photoID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetBurstStats returns statistics about detected bursts
func (bd *BurstDetector) GetBurstStats() (int, int, error) {
	var burstCount int
	err := bd.db.QueryRow("SELECT COUNT(*) FROM burst_groups").Scan(&burstCount)
	if err != nil {
		return 0, 0, err
	}

	var photoCount int
	err = bd.db.QueryRow("SELECT COUNT(*) FROM photos WHERE burst_group_id IS NOT NULL").Scan(&photoCount)
	if err != nil {
		return 0, 0, err
	}

	return burstCount, photoCount, nil
}

// BurstGroup represents a detected burst group
type BurstGroup struct {
	ID        string
	PhotoIDs  []int
	CreatedAt time.Time
}

// GetAllBursts retrieves all burst groups from the database
func (bd *BurstDetector) GetAllBursts() ([]BurstGroup, error) {
	rows, err := bd.db.Query(`
		SELECT id, created_at FROM burst_groups ORDER BY date_taken
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []BurstGroup
	for rows.Next() {
		var g BurstGroup
		var createdAtStr string
		err := rows.Scan(&g.ID, &createdAtStr)
		if err != nil {
			return nil, err
		}

		g.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

		// Get photos for this burst
		photoRows, err := bd.db.Query(`
			SELECT id FROM photos
			WHERE burst_group_id = ?
			ORDER BY burst_sequence
		`, g.ID)
		if err != nil {
			return nil, err
		}

		for photoRows.Next() {
			var photoID int
			photoRows.Scan(&photoID)
			g.PhotoIDs = append(g.PhotoIDs, photoID)
		}
		photoRows.Close()

		groups = append(groups, g)
	}

	return groups, nil
}
