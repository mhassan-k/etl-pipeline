package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// Record represents a raw data record stored in the database
type Record struct {
	ID        int
	Data      string
	Timestamp time.Time
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(connectionString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pgDB := &PostgresDB{db: db}

	// Initialize schema
	if err := pgDB.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return pgDB, nil
}

// initSchema creates the necessary database tables
func (p *PostgresDB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS raw_data (
		id SERIAL PRIMARY KEY,
		data JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS processed_data (
		id SERIAL PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		body TEXT,
		processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_raw_data_created_at ON raw_data(created_at);
	CREATE INDEX IF NOT EXISTS idx_processed_data_processed_at ON processed_data(processed_at);
	CREATE INDEX IF NOT EXISTS idx_processed_data_user_id ON processed_data(user_id);
	`

	_, err := p.db.Exec(schema)
	return err
}

// InsertRawData inserts raw data into the database
func (p *PostgresDB) InsertRawData(data []map[string]interface{}) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO raw_data (data) VALUES ($1)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range data {
		jsonData, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}

		if _, err := stmt.Exec(jsonData); err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertProcessedData inserts processed data into the database
func (p *PostgresDB) InsertProcessedData(records []ProcessedRecord) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO processed_data (user_id, title, body) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		if _, err := stmt.Exec(record.UserID, record.Title, record.Body); err != nil {
			return fmt.Errorf("failed to insert processed record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ProcessedRecord represents a processed data record
type ProcessedRecord struct {
	UserID int    `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// HealthCheck checks if the database connection is healthy
func (p *PostgresDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.db.PingContext(ctx)
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}
