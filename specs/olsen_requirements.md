# Requirements Document: DNG Photo Indexer

**Version:** 1.0  
**Date:** October 2025  
**Status:** Final

## 1. Executive Summary

A command-line photo management system that recursively indexes DNG (Digital Negative) photographs, extracting comprehensive metadata, generating thumbnails, and analyzing color palettes. The system stores all data in a SQLite database optimized for fast querying across large collections (100,000+ images).

## 2. Business Requirements

### BR-1: Core Functionality
- **BR-1.1**: System SHALL recursively scan directories for DNG files
- **BR-1.2**: System SHALL extract all available EXIF metadata from DNG files
- **BR-1.3**: System SHALL generate thumbnail images for quick preview
- **BR-1.4**: System SHALL extract dominant color palettes from images
- **BR-1.5**: System SHALL store all data in a searchable database
- **BR-1.6**: System SHALL provide query capabilities for finding photos

### BR-2: Performance
- **BR-2.1**: System SHALL handle collections of 100,000+ photos efficiently
- **BR-2.2**: Indexing SHALL process at least 10 photos per second on modern hardware
- **BR-2.3**: Database queries SHALL return results in under 1 second for indexed fields
- **BR-2.4**: System SHALL support concurrent processing to maximize CPU utilization

### BR-3: Data Integrity
- **BR-3.1**: System SHALL never modify original photo files
- **BR-3.2**: System SHALL calculate file hashes for duplicate detection
- **BR-3.3**: System SHALL handle missing or corrupted EXIF data gracefully
- **BR-3.4**: Database operations SHALL be transactional to prevent partial updates

### BR-4: Usability
- **BR-4.1**: System SHALL provide clear progress reporting during indexing
- **BR-4.2**: System SHALL provide meaningful error messages
- **BR-4.3**: System SHALL offer statistics about the photo collection
- **BR-4.4**: Color search SHALL support both hex codes and human-readable names

## 3. Functional Requirements

### FR-1: Metadata Extraction

#### FR-1.1: Camera Metadata
- Extract camera make and model
- Extract lens make and model
- Store equipment information for filtering

#### FR-1.2: Exposure Metadata
- Extract ISO speed rating
- Extract aperture (f-number)
- Extract shutter speed
- Extract exposure compensation
- Extract focal length (actual and 35mm equivalent)
- All numeric values stored as appropriate data types

#### FR-1.3: Temporal Metadata
- Extract date/time photo was taken
- Extract date/time photo was digitized
- Store timestamps in ISO 8601 format
- Support timezone information

#### FR-1.4: Location Metadata
- Extract GPS latitude
- Extract GPS longitude
- Extract GPS altitude
- Store coordinates as decimal degrees

#### FR-1.5: Image Properties
- Extract image width and height
- Extract orientation (rotation)
- Extract color space information
- Extract bit depth where available

#### FR-1.6: Flash and Lighting
- Extract flash fired status
- Extract white balance mode
- Extract focus distance where available

#### FR-1.7: DNG-Specific Metadata
- Extract DNG version
- Extract original RAW filename (if converted)

### FR-2: Intelligent Inference

#### FR-2.1: Time of Day Classification
Based on hour of capture, classify as:
- Golden hour (morning): 5:00-7:00
- Morning: 7:00-11:00
- Midday: 11:00-15:00
- Afternoon: 15:00-18:00
- Golden hour (evening): 18:00-20:00
- Blue hour: 20:00-22:00
- Night: 22:00-5:00

#### FR-2.2: Season Classification
Based on month of capture (Northern Hemisphere):
- Spring: March-May
- Summer: June-August
- Autumn: September-November
- Winter: December-February

#### FR-2.3: Focal Length Categories
- Wide: < 35mm
- Normal: 35-70mm
- Telephoto: 71-200mm
- Super telephoto: > 200mm

#### FR-2.4: Shooting Conditions
Based on ISO and flash:
- Bright: ISO ≤ 400, no flash
- Moderate: ISO 401-1599, no flash
- Low light: ISO ≥ 1600, no flash
- Flash: Any ISO with flash fired

### FR-3: Visual Analysis

#### FR-3.1: Thumbnail Generation
- Generate 256×256 pixel thumbnails
- Use high-quality Lanczos3 resampling
- Store as JPEG with quality setting 85
- Store thumbnail data as BLOB in database

#### FR-3.2: Color Palette Extraction
- Extract top 5 dominant colors using k-means clustering
- Calculate proportional weight for each color
- Store colors in both RGB and HSL color spaces
- Run clustering for maximum 100 iterations
- Extract palette from thumbnail for efficiency

#### FR-3.3: Color Space Conversion
- Convert RGB (0-255 per channel) to HSL
- Hue: 0-360 degrees
- Saturation: 0-100%
- Lightness: 0-100%
- Store all values for flexible querying

### FR-4: Database Design

#### FR-4.1: Schema Requirements
- **photos** table: Store all photo metadata
- **photo_colors** table: Store color palette data (1-to-many)
- **tags** table: User-defined tags
- **photo_tags** table: Many-to-many tag relationships
- **collections** table: Virtual photo collections
- **collection_photos** table: Collection membership

#### FR-4.2: Indexing Requirements
Create indexes on frequently queried fields:
- date_taken
- camera_make, camera_model (composite)
- lens_model
- latitude, longitude (composite)
- file_hash
- iso
- aperture
- time_of_day
- hue, saturation, lightness (color table)
- red, green, blue (color table, composite)

#### FR-4.3: Data Types
- Text fields: UTF-8 encoded
- Timestamps: ISO 8601 datetime
- Coordinates: REAL (decimal degrees)
- Colors: INTEGER (0-255 for RGB, 0-360 for hue, 0-100 for S/L)
- File hash: TEXT (SHA-256 hex string)
- Thumbnails: BLOB

### FR-5: Query Capabilities

#### FR-5.1: Statistics Queries
- Total photo count
- Top 10 cameras by usage
- Top 10 lenses by usage
- Photos by time of day distribution
- Dominant color distribution
- Total storage used
- Date range of collection

#### FR-5.2: Color Search
- Search by hex color code (e.g., #FF5733)
- RGB tolerance: ±30 in each channel
- Search by hue name: red, orange, yellow, green, cyan, blue, purple, pink
- Results ordered by color similarity

#### FR-5.3: Hue Mapping
Map color names to hue ranges:
- Red: 0-15° and 345-360°
- Orange: 16-45°
- Yellow: 46-75°
- Green: 76-165°
- Cyan: 166-195°
- Blue: 196-255°
- Purple: 256-285°
- Pink: 286-344°

### FR-6: Command-Line Interface

#### FR-6.1: Scan Command
```
./indexer -scan <path> [-workers N] [-db path]
```
- `path`: Directory to scan (required)
- `workers`: Number of concurrent workers (default: 4)
- `db`: Database file path (default: photos.db)

#### FR-6.2: Query Commands
```
./indexer -query stats
./indexer -query "color:#RRGGBB"
./indexer -query "hue:colorname"
```

#### FR-6.3: Progress Reporting
- Display total files found
- Display progress every 100 files processed
- Display summary at completion: processed, failed, duration

### FR-7: Concurrent Processing

#### FR-7.1: Worker Pool
- Configurable number of workers
- Channel-based work distribution
- Each worker processes one file at a time
- Shared database access with transaction safety

#### FR-7.2: Error Handling
- Log errors without stopping indexing
- Track failed file count
- Continue processing remaining files
- Report failures in summary

## 4. Non-Functional Requirements

### NFR-1: Performance
- **NFR-1.1**: Indexing throughput ≥ 10 photos/second
- **NFR-1.2**: Color search queries < 500ms on 100K photos
- **NFR-1.3**: Statistics queries < 1000ms on 100K photos
- **NFR-1.4**: Memory usage < 500MB during indexing
- **NFR-1.5**: Database size ≤ 40KB per photo (including thumbnails)

### NFR-2: Reliability
- **NFR-2.1**: No data corruption on unexpected termination
- **NFR-2.2**: Graceful handling of unreadable files
- **NFR-2.3**: Transaction-based database updates
- **NFR-2.4**: Validation of all user inputs

### NFR-3: Maintainability
- **NFR-3.1**: Code organized into logical modules
- **NFR-3.2**: Clear separation of concerns
- **NFR-3.3**: Comprehensive error logging
- **NFR-3.4**: Well-documented functions and data structures

### NFR-4: Portability
- **NFR-4.1**: Cross-platform support (Linux, macOS, Windows)
- **NFR-4.2**: No platform-specific dependencies
- **NFR-4.3**: Single binary deployment
- **NFR-4.4**: Standard SQLite database format

### NFR-5: Scalability
- **NFR-5.1**: Linear scaling with worker count (up to CPU cores)
- **NFR-5.2**: Support for collections up to 1M photos
- **NFR-5.3**: Efficient database indexing strategy
- **NFR-5.4**: Constant memory usage regardless of collection size

## 5. Constraints

### C-1: Technical Constraints
- **C-1.1**: Must use Go programming language
- **C-1.2**: Must use SQLite for storage
- **C-1.3**: Must support DNG file format
- **C-1.4**: No external API dependencies

### C-2: Resource Constraints
- **C-2.1**: Target systems: 4+ CPU cores
- **C-2.2**: Target systems: 8GB+ RAM
- **C-2.3**: Target systems: SSD storage recommended

### C-3: Implementation Constraints
- **C-3.1**: Command-line interface only (no GUI in v1.0)
- **C-3.2**: Read-only access to photos (no editing)
- **C-3.3**: Local processing only (no cloud features)

## 6. Dependencies

### D-1: External Libraries
- **goexif**: EXIF metadata extraction
- **go-sqlite3**: SQLite database driver
- **resize**: Image resizing library
- **palettor**: K-means color palette extraction

### D-2: System Requirements
- Go 1.16 or later
- CGo enabled (for SQLite)
- libjpeg/libpng support (standard)

## 7. Success Criteria

### SC-1: Functional Success
- Successfully indexes 10,000 DNG files without errors
- Extracts metadata from 95%+ of test files
- Color search returns relevant results
- Statistics accurately reflect collection

### SC-2: Performance Success
- Processes 100 photos in < 10 seconds (10 photos/sec)
- Color search returns in < 500ms
- Database size meets < 40KB per photo target
- Memory usage stays under 500MB

### SC-3: Quality Success
- Zero data corruption incidents
- Graceful handling of all error cases
- Clear error messages for failures
- Successful operation on Windows, macOS, Linux

## 8. Future Enhancements (Out of Scope for v1.0)

### FE-1: Advanced Features
- Perceptual hashing for near-duplicate detection
- Burst sequence detection and grouping
- Sharpness scoring for identifying best shots
- Face detection and recognition
- ML-based object and scene detection
- Multi-size thumbnail generation
- Full-text search with SQLite FTS5

### FE-2: User Interface
- Web-based UI for browsing
- Visual color picker for searches
- Interactive filtering interface
- Gallery view with virtual scrolling

### FE-3: Extended Format Support
- RAW formats: CR2, NEF, ARW, ORF
- Video file indexing with frame extraction
- JPEG/PNG for testing purposes
- XMP sidecar file reading/writing

### FE-4: Advanced Organization
- Smart collection rules engine
- Automatic tagging suggestions
- Duplicate management tools
- Export to HTML galleries
- Integration with editing software

## 9. Assumptions

### A-1: User Environment
- Users have Go development environment
- Users have permissions to read photo directories
- Users have write permissions for database location
- Photo files are stored on accessible filesystems

### A-2: Data Assumptions
- DNG files contain standard EXIF metadata
- Photos have embedded thumbnails or can be decoded
- File timestamps are reliable
- GPS data (when present) is accurate

### A-3: Usage Assumptions
- Collections are indexed on single machine
- Photos are not modified during indexing
- Users accept read-only database access during queries
- Command-line interface is acceptable for v1.0

## 10. Risks and Mitigations

### R-1: Performance Risks
- **Risk**: Slow processing on large images
- **Mitigation**: Extract palette from thumbnails, not full-res
- **Risk**: Database locks with high concurrency
- **Mitigation**: Transaction batching, configurable workers

### R-2: Data Quality Risks
- **Risk**: Missing EXIF data in some files
- **Mitigation**: Graceful fallback, partial indexing
- **Risk**: Corrupt image files
- **Mitigation**: Error handling, continue processing

### R-3: Compatibility Risks
- **Risk**: Platform-specific issues
- **Mitigation**: Standard library usage, cross-platform testing
- **Risk**: SQLite version differences
- **Mitigation**: Use widely supported SQL features

## 11. Glossary

- **DNG**: Digital Negative, Adobe's RAW image format
- **EXIF**: Exchangeable Image File Format, metadata standard
- **HSL**: Hue, Saturation, Lightness color model
- **K-means**: Clustering algorithm for color quantization
- **SQLite**: Embedded relational database
- **Perceptual hash**: Content-based image fingerprint
- **BLOB**: Binary Large Object (database storage)
- **FTS5**: Full-Text Search version 5 (SQLite extension)

## 12. Approval

**Product Owner**: _______________  
**Technical Lead**: _______________  
**Date**: _______________