-- Create artists table
CREATE TABLE IF NOT EXISTS artists (
    artist_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create albums table
CREATE TABLE IF NOT EXISTS albums (
    album_id VARCHAR(36) PRIMARY KEY,
    artist_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    year INT NOT NULL,
    original_image_key VARCHAR(255),
    processed_image_key VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (artist_id) REFERENCES artists(artist_id) ON DELETE CASCADE
);

-- Create reviews table for MySQL implementation (for comparison with DynamoDB)
CREATE TABLE IF NOT EXISTS reviews (
    review_id VARCHAR(36) PRIMARY KEY,
    album_id VARCHAR(36) NOT NULL,
    like_count INT UNSIGNED DEFAULT 0,
    dislike_count INT UNSIGNED DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (album_id) REFERENCES albums(album_id) ON DELETE CASCADE
);

-- Create indexes to improve query performance
CREATE INDEX idx_albums_artist_id ON albums(artist_id);
CREATE INDEX idx_albums_year ON albums(year);
CREATE INDEX idx_reviews_album_id ON reviews(album_id);
