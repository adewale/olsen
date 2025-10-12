package database

const Schema = `
-- ============================================================
-- PHOTOS TABLE (Core metadata)
-- ============================================================
CREATE TABLE IF NOT EXISTS photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT UNIQUE NOT NULL,
    file_hash TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_modified DATETIME NOT NULL,

    -- Camera metadata
    camera_make TEXT,
    camera_model TEXT,
    lens_make TEXT,
    lens_model TEXT,

    -- Exposure metadata
    iso INTEGER,
    aperture REAL,
    shutter_speed TEXT,
    exposure_compensation REAL,
    focal_length REAL,
    focal_length_35mm INTEGER,

    -- Temporal metadata
    date_taken DATETIME,
    date_digitized DATETIME,

    -- Image properties
    width INTEGER,
    height INTEGER,
    orientation INTEGER,
    color_space TEXT,

    -- Location metadata
    latitude REAL,
    longitude REAL,
    altitude REAL,

    -- DNG-specific
    dng_version TEXT,
    original_raw_filename TEXT,

    -- Lighting metadata
    flash_fired BOOLEAN,
    white_balance TEXT,
    focus_distance REAL,

    -- Inferred metadata
    time_of_day TEXT,
    season TEXT,
    focal_category TEXT,
    shooting_condition TEXT,

    -- Perceptual hash
    perceptual_hash TEXT,

    -- Burst metadata
    burst_group_id TEXT,
    burst_sequence INTEGER,
    burst_count INTEGER,
    is_burst_representative BOOLEAN DEFAULT FALSE
);

-- ============================================================
-- THUMBNAILS TABLE (Multiple sizes stored)
-- ============================================================
CREATE TABLE IF NOT EXISTS thumbnails (
    photo_id INTEGER NOT NULL,
    size TEXT NOT NULL,  -- "64", "256", "512", "1024" (longest edge)
    data BLOB NOT NULL,
    format TEXT DEFAULT 'jpeg',
    quality INTEGER DEFAULT 85,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (photo_id, size),
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE
);

-- ============================================================
-- PHOTO COLORS TABLE (Dominant color palette)
-- ============================================================
CREATE TABLE IF NOT EXISTS photo_colors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    photo_id INTEGER NOT NULL,
    color_order INTEGER NOT NULL,
    red INTEGER NOT NULL,
    green INTEGER NOT NULL,
    blue INTEGER NOT NULL,
    weight REAL NOT NULL,
    hue INTEGER,
    saturation INTEGER,
    lightness INTEGER,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    UNIQUE(photo_id, color_order)
);

-- ============================================================
-- BURST GROUPS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS burst_groups (
    id TEXT PRIMARY KEY,
    photo_count INTEGER NOT NULL,
    date_taken DATETIME,
    camera_make TEXT,
    camera_model TEXT,
    representative_photo_id INTEGER,
    time_span_seconds REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (representative_photo_id) REFERENCES photos(id)
);

-- ============================================================
-- TAGS TABLE (User-defined)
-- ============================================================
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS photo_tags (
    photo_id INTEGER,
    tag_id INTEGER,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (photo_id, tag_id)
);

-- ============================================================
-- COLLECTIONS TABLE (Virtual collections)
-- ============================================================
CREATE TABLE IF NOT EXISTS collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT CHECK(type IN ('manual', 'smart')) DEFAULT 'manual',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collection_photos (
    collection_id INTEGER,
    photo_id INTEGER,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
    FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, photo_id)
);

-- ============================================================
-- FACET METADATA (For display configuration)
-- ============================================================
CREATE TABLE IF NOT EXISTS facet_metadata (
    facet_type TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    facet_order INTEGER,
    allow_multiple BOOLEAN DEFAULT FALSE,
    hierarchical BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE
);

-- ============================================================
-- PERFORMANCE INDEXES
-- ============================================================
-- Core queries
CREATE INDEX IF NOT EXISTS idx_photos_date_taken ON photos(date_taken);
CREATE INDEX IF NOT EXISTS idx_photos_camera ON photos(camera_make, camera_model);
CREATE INDEX IF NOT EXISTS idx_photos_lens ON photos(lens_model);
CREATE INDEX IF NOT EXISTS idx_photos_gps ON photos(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_photos_hash ON photos(file_hash);
CREATE INDEX IF NOT EXISTS idx_photos_phash ON photos(perceptual_hash);

-- Faceted browsing
CREATE INDEX IF NOT EXISTS idx_photos_iso ON photos(iso);
CREATE INDEX IF NOT EXISTS idx_photos_aperture ON photos(aperture);
CREATE INDEX IF NOT EXISTS idx_photos_focal_length ON photos(focal_length);
CREATE INDEX IF NOT EXISTS idx_photos_time_of_day ON photos(time_of_day);
CREATE INDEX IF NOT EXISTS idx_photos_season ON photos(season);
CREATE INDEX IF NOT EXISTS idx_photos_focal_category ON photos(focal_category);
CREATE INDEX IF NOT EXISTS idx_photos_shooting_condition ON photos(shooting_condition);

-- Burst queries
CREATE INDEX IF NOT EXISTS idx_photos_burst ON photos(burst_group_id);
CREATE INDEX IF NOT EXISTS idx_burst_groups_date ON burst_groups(date_taken);

-- Color search
CREATE INDEX IF NOT EXISTS idx_colors_hue ON photo_colors(hue);
CREATE INDEX IF NOT EXISTS idx_colors_saturation ON photo_colors(saturation);
CREATE INDEX IF NOT EXISTS idx_colors_lightness ON photo_colors(lightness);
CREATE INDEX IF NOT EXISTS idx_colors_rgb ON photo_colors(red, green, blue);
CREATE INDEX IF NOT EXISTS idx_colors_photo ON photo_colors(photo_id, color_order);
`

const FacetMetadataInserts = `
INSERT OR IGNORE INTO facet_metadata VALUES
('camera_make', 'Camera', 1, 0, 1, 1),
('lens_model', 'Lens', 2, 0, 0, 1),
('time_of_day', 'Time of Day', 3, 1, 0, 1),
('season', 'Season', 4, 1, 0, 1),
('color', 'Dominant Color', 5, 1, 0, 1),
('iso', 'ISO', 6, 0, 0, 1),
('aperture', 'Aperture', 7, 0, 0, 1),
('focal_category', 'Focal Length', 8, 1, 0, 1),
('burst_group', 'Bursts', 9, 0, 0, 1);
`
