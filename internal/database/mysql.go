package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"employee-management-platform/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

// Connect initializes a MySQL database connection pool and verifies connectivity.
func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.DSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	// Verify connectivity with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database at %s:%s (database: %s): %w", cfg.DBHost, cfg.DBPort, cfg.DBName, err)
	}

	log.Printf("[DATABASE] Successfully connected to MySQL database %q on %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort)
	return db, nil
}

// Close closes the database connection cleanly.
func Close(db *sql.DB) error {
	if db != nil {
		log.Println("[DATABASE] Closing MySQL database connection pool...")
		return db.Close()
	}
	return nil
}
