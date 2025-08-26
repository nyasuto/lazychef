package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps sql.DB with LazyChef-specific functionality
type Database struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	Path string
}

// New creates a new database connection
func New(config Config) (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Open database connection
	db, err := sql.Open("sqlite3", config.Path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	
	return &Database{DB: db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.DB.Close()
}

// Health checks database health
func (db *Database) Health() error {
	return db.Ping()
}

// GetStats returns database statistics
func (db *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Recipe count
	var recipeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&recipeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe count: %w", err)
	}
	stats["recipe_count"] = recipeCount
	
	// Meal plan count
	var mealPlanCount int
	err = db.QueryRow("SELECT COUNT(*) FROM meal_plans").Scan(&mealPlanCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get meal plan count: %w", err)
	}
	stats["meal_plan_count"] = mealPlanCount
	
	// User preferences count
	var userPrefCount int
	err = db.QueryRow("SELECT COUNT(*) FROM user_preferences").Scan(&userPrefCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences count: %w", err)
	}
	stats["user_preferences_count"] = userPrefCount
	
	// Database size (approximate)
	var pageCount, pageSize int
	err = db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err == nil {
		err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
		if err == nil {
			stats["database_size_bytes"] = pageCount * pageSize
		}
	}
	
	return stats, nil
}

// Execute executes a query with optional transaction
func (db *Database) Execute(query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	return err
}

// ExecuteInTx executes a function within a transaction
func (db *Database) ExecuteInTx(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// Vacuum performs database maintenance
func (db *Database) Vacuum() error {
	_, err := db.Exec("VACUUM")
	return err
}

// Backup creates a backup of the database
func (db *Database) Backup(backupPath string) error {
	// Simple file copy backup
	backupDB, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup database: %w", err)
	}
	defer backupDB.Close()
	
	// Use SQLite's backup API through ATTACH
	_, err = db.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS backup", backupPath))
	if err != nil {
		return fmt.Errorf("failed to attach backup database: %w", err)
	}
	defer db.Exec("DETACH DATABASE backup")
	
	// Copy tables
	tables := []string{"recipes", "meal_plans", "user_preferences"}
	for _, table := range tables {
		query := fmt.Sprintf("CREATE TABLE backup.%s AS SELECT * FROM main.%s", table, table)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to backup table %s: %w", table, err)
		}
	}
	
	return nil
}

// IsTableEmpty checks if a table is empty
func (db *Database) IsTableEmpty(tableName string) (bool, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if table %s is empty: %w", tableName, err)
	}
	return count == 0, nil
}

// GetLastInsertID returns the ID of the last inserted row
func (db *Database) GetLastInsertID() (int64, error) {
	var id int64
	err := db.QueryRow("SELECT last_insert_rowid()").Scan(&id)
	return id, err
}