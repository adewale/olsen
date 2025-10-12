# Processing Flow Diagrams

## Complete Indexing Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     User Invocation                              │
│                  olsen index <path>                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Initialize Engine                              │
│  • Open/create SQLite database                                   │
│  • Set worker count (default: 4)                                │
│  • Initialize statistics counters                                │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Scan Filesystem                               │
│  • Recursive directory walk                                      │
│  • Filter for: .dng, .jpg, .jpeg, .bmp                          │
│  • Collect file paths                                            │
│  • Report: "Found N files"                                       │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    Files Found > 0?
                             │
                    ┌────────┴────────┐
                    │                 │
                   YES               NO
                    │                 │
                    │                 ▼
                    │        ┌──────────────┐
                    │        │ Exit (0)     │
                    │        └──────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Create Worker Pool                              │
│  • Create buffered channel (size: 100)                           │
│  • Spawn N worker goroutines                                     │
│  • Each worker waits on channel                                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Distribute Work                                 │
│  • Push each file path to channel                                │
│  • Close channel when done                                       │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
    ┌────────┐          ┌────────┐          ┌────────┐
    │Worker 1│          │Worker 2│          │Worker N│
    │        │          │        │          │        │
    │ ┌──────┴──────┐  │ ┌──────┴──────┐  │ ┌──────┴──────┐
    │ │processFile()│  │ │processFile()│  │ │processFile()│
    │ └──────┬──────┘  │ └──────┬──────┘  │ └──────┬──────┘
    └────────┘          └────────┘          └────────┘
         │                   │                   │
         └───────────────────┴───────────────────┘
                             │
                    All workers complete
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Report Summary                                 │
│  • Files found: N                                                │
│  • Files processed: N                                            │
│  • Files failed: N                                               │
│  • Duration: Xs                                                  │
│  • Rate: X photos/second                                         │
└─────────────────────────────────────────────────────────────────┘
```

## processFile() Detailed Flow

```
                    processFile(path)
                          │
                          ▼
              ┌───────────────────────┐
              │ Already indexed?      │
              │ (check database)      │
              └───────┬───────────────┘
                      │
              ┌───────┴────────┐
              │                │
             YES              NO
              │                │
              ▼                │
        Skip (return)          │
                               │
                               ▼
              ┌────────────────────────────┐
              │ Step 1: Extract EXIF       │
              │ • Open file                │
              │ • Parse EXIF data          │
              │ • Populate PhotoMetadata   │
              └────────┬───────────────────┘
                       │
                  Success?
                       │
              ┌────────┴────────┐
              │                 │
             NO                YES
              │                 │
              ▼                 │
      ┌──────────────┐          │
      │ Log error    │          │
      │ Return error │          │
      └──────────────┘          │
                                │
                                ▼
              ┌────────────────────────────┐
              │ Step 2: Calculate Hash     │
              │ • Read file bytes          │
              │ • Compute SHA-256          │
              │ • Store in metadata        │
              └────────┬───────────────────┘
                       │
                       ▼
              ┌────────────────────────────┐
              │ Step 3: Decode Image       │
              │ • Open file                │
              │ • Detect format            │
              │ • Decode to image.Image    │
              └────────┬───────────────────┘
                       │
                  Success?
                       │
              ┌────────┴────────┐
              │                 │
             NO                YES
              │                 │
              ▼                 │
      ┌──────────────┐          │
      │ Log error    │          │
      │ Return error │          │
      └──────────────┘          │
                                │
                                ▼
              ┌────────────────────────────┐
              │ Step 4: Generate Thumbnails│
              │ • For each size:           │
              │   - Calculate dimensions   │
              │   - Resize (Lanczos3)      │
              │   - Encode JPEG (Q=85)     │
              │ • Store 4 BLOBs            │
              └────────┬───────────────────┘
                       │
                       ▼
              ┌────────────────────────────┐
              │ Step 5: Extract Colors     │
              │ • Use 256px thumbnail      │
              │ • Decode thumbnail         │
              │ • K-means clustering       │
              │ • Extract 5 colors         │
              │ • Convert RGB → HSL        │
              │ • Calculate weights        │
              └────────┬───────────────────┘
                       │
                       ▼
              ┌────────────────────────────┐
              │ Step 6: Compute pHash      │
              │ • Use 256px thumbnail      │
              │ • Resize to 32×32 gray     │
              │ • Apply DCT                │
              │ • Extract 64-bit hash      │
              └────────┬───────────────────┘
                       │
                       ▼
              ┌────────────────────────────┐
              │ Step 7: Infer Metadata     │
              │ • Time of day (from hour)  │
              │ • Season (from month)      │
              │ • Focal category (from mm) │
              │ • Conditions (from ISO)    │
              └────────┬───────────────────┘
                       │
                       ▼
              ┌────────────────────────────┐
              │ Step 8: Database Insert    │
              │ • BEGIN TRANSACTION        │
              │ • INSERT INTO photos       │
              │ • INSERT INTO thumbnails×4 │
              │ • INSERT INTO photo_colors×5│
              │ • COMMIT                   │
              └────────┬───────────────────┘
                       │
                  Success?
                       │
              ┌────────┴────────┐
              │                 │
             NO                YES
              │                 │
              ▼                 ▼
      ┌──────────────┐  ┌──────────────┐
      │ ROLLBACK     │  │ Increment    │
      │ Log error    │  │ success count│
      │ Return error │  │ Return nil   │
      └──────────────┘  └──────────────┘
```

## Thumbnail Generation Flow

```
                generateThumbnails(img)
                          │
                          ▼
              ┌───────────────────────┐
              │ Get image dimensions  │
              │ width × height        │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ For each size:        │
              │ • 64px                │
              │ • 256px               │
              │ • 512px               │
              │ • 1024px              │
              └───────┬───────────────┘
                      │
            ┌─────────┴─────────┐
            │                   │
         width > height?     height ≥ width?
            │                   │
           YES                 YES
            │                   │
            ▼                   ▼
    ┌──────────────┐    ┌──────────────┐
    │ Landscape    │    │ Portrait/Sq  │
    │              │    │              │
    │ newWidth=N   │    │ newWidth=0   │
    │ newHeight=0  │    │ newHeight=N  │
    └──────┬───────┘    └──────┬───────┘
           │                   │
           └─────────┬─────────┘
                     │
                     ▼
           ┌─────────────────────┐
           │ resize.Resize()     │
           │ • Lanczos3 filter   │
           │ • Auto-calculate    │
           │   other dimension   │
           └─────────┬───────────┘
                     │
                     ▼
           ┌─────────────────────┐
           │ Encode as JPEG      │
           │ • Quality: 85       │
           │ • Write to buffer   │
           └─────────┬───────────┘
                     │
                     ▼
           ┌─────────────────────┐
           │ Store BLOB          │
           │ map[size][]byte     │
           └─────────────────────┘
                     │
                     ▼
              More sizes?
                     │
              ┌──────┴──────┐
              │             │
             YES           NO
              │             │
              ▼             ▼
         (loop back)   Return map
```

## Color Palette Extraction Flow

```
              extractColorPalette(img, numColors=5)
                          │
                          ▼
              ┌───────────────────────┐
              │ K-means Clustering    │
              │ • maxIterations: 100  │
              │ • k: 5 colors         │
              │ • Input: image pixels │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Palette with entries  │
              │ Each entry:           │
              │ • color.Color         │
              │ • weight (float64)    │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ For each entry:       │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Extract RGB           │
              │ • Get RGBA()          │
              │ • Convert 16→8 bit    │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ RGB → HSL Conversion  │
              │                       │
              │ r,g,b → [0,1]        │
              │ max = max(r,g,b)     │
              │ min = min(r,g,b)     │
              │ Δ = max - min        │
              │                       │
              │ L = (max+min)/2      │
              │                       │
              │ if Δ=0: S=0, H=0     │
              │ else:                 │
              │   S = Δ/(2-max-min)  │
              │   H = calc from RGB   │
              │                       │
              │ Output:               │
              │ • H: 0-360°          │
              │ • S: 0-100%          │
              │ • L: 0-100%          │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Create DominantColor  │
              │ • Color (RGB)         │
              │ • ColorHSL (HSL)      │
              │ • Weight (from entry) │
              └───────┬───────────────┘
                      │
                      ▼
                 More colors?
                      │
              ┌───────┴──────┐
              │              │
             YES            NO
              │              │
              ▼              ▼
         (loop back)   Return array
```

## Perceptual Hash Computation Flow

```
              computePerceptualHash(img)
                          │
                          ▼
              ┌───────────────────────┐
              │ Resize to 32×32       │
              │ • Convert to grayscale│
              │ • Bicubic interpolate │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Apply DCT             │
              │ (Discrete Cosine      │
              │  Transform)           │
              │ • Converts to freq    │
              │   domain              │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Extract Low Freq      │
              │ • Take top-left 8×8   │
              │ • These are the most  │
              │   significant         │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Compute Median        │
              │ • Of 64 DCT values    │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Generate Hash         │
              │ For each of 64 values:│
              │ • If > median: 1      │
              │ • If ≤ median: 0      │
              │                       │
              │ Result: 64-bit hash   │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ Convert to String     │
              │ • Hex representation  │
              │ • 16-18 characters    │
              └───────┬───────────────┘
                      │
                      ▼
                 Return hash string
```

## Metadata Inference Flow

```
                  inferMetadata(metadata)
                          │
                          ▼
            ┌─────────────────────────────┐
            │ Time of Day Classification  │
            │                             │
            │ hour := dateTaken.Hour()    │
            │                             │
            │  5-7:  golden_hour_morning  │
            │  7-11: morning              │
            │ 11-15: midday               │
            │ 15-18: afternoon            │
            │ 18-20: golden_hour_evening  │
            │ 20-22: blue_hour            │
            │ 22-5:  night                │
            └─────────────┬───────────────┘
                          │
                          ▼
            ┌─────────────────────────────┐
            │ Season Classification       │
            │                             │
            │ month := dateTaken.Month()  │
            │                             │
            │ Mar-May:   spring           │
            │ Jun-Aug:   summer           │
            │ Sep-Nov:   autumn           │
            │ Dec-Feb:   winter           │
            └─────────────┬───────────────┘
                          │
                          ▼
            ┌─────────────────────────────┐
            │ Focal Length Category       │
            │                             │
            │ focal := focalLength35mm    │
            │                             │
            │ < 35mm:    wide             │
            │ 35-70mm:   normal           │
            │ 71-200mm:  telephoto        │
            │ > 200mm:   super_telephoto  │
            └─────────────┬───────────────┘
                          │
                          ▼
            ┌─────────────────────────────┐
            │ Shooting Conditions         │
            │                             │
            │ if flashFired:              │
            │   → "flash"                 │
            │                             │
            │ else by ISO:                │
            │   ≤400:     bright          │
            │   401-1599: moderate        │
            │   ≥1600:    low_light       │
            └─────────────┬───────────────┘
                          │
                          ▼
                  metadata.TimeOfDay = ...
                  metadata.Season = ...
                  metadata.FocalCategory = ...
                  metadata.ShootingCondition = ...
```

## Database Transaction Flow

```
                  db.InsertPhoto(metadata)
                          │
                          ▼
              ┌───────────────────────┐
              │ BEGIN TRANSACTION     │
              └───────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │ INSERT INTO photos    │
              │ • 50+ fields          │
              │ • Returns photo_id    │
              └───────┬───────────────┘
                      │
                  Success?
                      │
              ┌───────┴──────┐
              │              │
             NO             YES
              │              │
              ▼              │
      ┌──────────────┐       │
      │ ROLLBACK     │       │
      │ Return error │       │
      └──────────────┘       │
                             │
                             ▼
              ┌───────────────────────┐
              │ INSERT INTO thumbnails│
              │ • 4 rows (sizes)      │
              │ • photo_id (FK)       │
              │ • BLOB data           │
              └───────┬───────────────┘
                      │
                  Success?
                      │
              ┌───────┴──────┐
              │              │
             NO             YES
              │              │
              ▼              │
      ┌──────────────┐       │
      │ ROLLBACK     │       │
      │ Return error │       │
      └──────────────┘       │
                             │
                             ▼
              ┌───────────────────────┐
              │ INSERT INTO           │
              │ photo_colors          │
              │ • 5 rows (colors)     │
              │ • photo_id (FK)       │
              │ • RGB + HSL values    │
              └───────┬───────────────┘
                      │
                  Success?
                      │
              ┌───────┴──────┐
              │              │
             NO             YES
              │              │
              ▼              ▼
      ┌──────────────┐  ┌──────────────┐
      │ ROLLBACK     │  │ COMMIT       │
      │ Return error │  │ Return nil   │
      └──────────────┘  └──────────────┘
```

## Error Handling Flow

```
                Any operation fails
                        │
                        ▼
            ┌───────────────────────┐
            │ Log error with:       │
            │ • File path           │
            │ • Error message       │
            │ • Stack trace (debug) │
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ Increment             │
            │ stats.FilesFailed     │
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ If in transaction:    │
            │ • ROLLBACK            │
            │ • Free resources      │
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ Return error to worker│
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ Worker continues with │
            │ next file in queue    │
            └───────────────────────┘
                    │
                    ▼
            Main indexing continues
            (no crash, graceful degradation)
```

## Progress Reporting Flow

```
         Worker processes file successfully
                        │
                        ▼
            ┌───────────────────────┐
            │ Acquire mutex lock    │
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ stats.FilesProcessed++│
            └───────┬───────────────┘
                    │
                    ▼
            ┌───────────────────────┐
            │ FilesProcessed % 100  │
            │ == 0?                 │
            └───────┬───────────────┘
                    │
            ┌───────┴──────┐
            │              │
           YES            NO
            │              │
            ▼              │
┌───────────────────┐      │
│ Print progress:   │      │
│                   │      │
│ "Progress: N/M    │      │
│  files (X%)"      │      │
└───────────────────┘      │
            │              │
            └──────┬───────┘
                   │
                   ▼
       ┌───────────────────────┐
       │ Release mutex lock    │
       └───────────────────────┘
                   │
                   ▼
            Worker continues
```
