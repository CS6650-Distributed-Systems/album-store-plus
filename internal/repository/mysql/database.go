package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

// Database represents a MySQL database connection
type Database struct {
	*sql.DB
}

// Connect creates a new MySQL database connection
func Connect(cfg config.MySQLConfig) (*Database, error) {
	// Create the database connection string
	// DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Check if connection is alive
	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			return nil, err
		}
	}

	return &Database{DB: db}, nil
}
