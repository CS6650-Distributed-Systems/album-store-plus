package rds

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config represents database configuration
type Config struct {
	Username     string
	Password     string
	Host         string
	Port         string
	DatabaseName string
	MaxOpen      int
	MaxIdle      int
	MaxLifetime  time.Duration
}

// Connect creates a new database connection
func Connect(config *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DatabaseName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpen)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(config.MaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}

// InitSchema initializes the database schema if not exists
func InitSchema(db *sql.DB) error {
	// Create albums table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS albums (
			id VARCHAR(36) PRIMARY KEY,
			artist VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			year VARCHAR(4) NOT NULL,
			image_id VARCHAR(255) NOT NULL,
			image_size BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			INDEX idx_artist (artist),
			INDEX idx_title (title),
			INDEX idx_year (year)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating albums table: %w", err)
	}

	// Create album_reviews table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS album_reviews (
			id VARCHAR(36) PRIMARY KEY,
			album_id VARCHAR(36) NOT NULL,
			liked BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE,
			INDEX idx_album_id (album_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating album_reviews table: %w", err)
	}

	// Create events table for event sourcing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id VARCHAR(36) PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			aggregate_id VARCHAR(36) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			payload JSON NOT NULL,
			metadata JSON,
			INDEX idx_aggregate_id (aggregate_id),
			INDEX idx_event_type (event_type),
			INDEX idx_timestamp (timestamp)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating events table: %w", err)
	}

	return nil
}
