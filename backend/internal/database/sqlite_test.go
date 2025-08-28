package database

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{
		Path: tmpfile.Name(),
	}

	db, err := New(config)
	require.NoError(t, err)
	assert.NotNil(t, db)
	defer func() { _ = db.Close() }()

	// Test connection
	err = db.Health()
	assert.NoError(t, err)
}

func TestNew_InvalidPath(t *testing.T) {
	config := Config{
		Path: "/nonexistent/path/test.db",
	}

	db, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestDatabase_Health(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = db.Health()
	assert.NoError(t, err)
}

func TestDatabase_Execute(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test table
	err = db.Execute(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	assert.NoError(t, err)

	// Insert test data
	err = db.Execute(`
		INSERT INTO test_table (name, created_at) 
		VALUES (?, ?)
	`, "test_name", time.Now().Format(time.RFC3339))
	assert.NoError(t, err)
}

func TestDatabase_ExecuteInTx(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test table
	err = db.Execute(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test successful transaction
	err = db.ExecuteInTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test_table (name) VALUES (?)", "tx_test")
		return err
	})
	assert.NoError(t, err)

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table WHERE name = ?", "tx_test").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDatabase_ExecuteInTx_Rollback(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test table
	err = db.Execute(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test failed transaction (should rollback)
	err = db.ExecuteInTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test_table (name) VALUES (?)", "rollback_test")
		if err != nil {
			return err
		}
		return assert.AnError // Force error to trigger rollback
	})
	assert.Error(t, err)

	// Verify data was not inserted due to rollback
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table WHERE name = ?", "rollback_test").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDatabase_IsTableEmpty(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test table
	err = db.Execute(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test empty table
	isEmpty, err := db.IsTableEmpty("test_table")
	assert.NoError(t, err)
	assert.True(t, isEmpty)

	// Add data
	err = db.Execute("INSERT INTO test_table (name) VALUES (?)", "test")
	assert.NoError(t, err)

	// Test non-empty table
	isEmpty, err = db.IsTableEmpty("test_table")
	assert.NoError(t, err)
	assert.False(t, isEmpty)
}

func TestDatabase_IsTableEmpty_NonexistentTable(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	isEmpty, err := db.IsTableEmpty("nonexistent_table")
	assert.Error(t, err)
	assert.False(t, isEmpty)
}

func TestDatabase_GetLastInsertID(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create test table
	err = db.Execute(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Insert data
	err = db.Execute("INSERT INTO test_table (name) VALUES (?)", "test")
	assert.NoError(t, err)

	// Get last insert ID
	lastID, err := db.GetLastInsertID()
	assert.NoError(t, err)
	assert.Greater(t, lastID, int64(0))
}

func TestDatabase_Vacuum(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := Config{Path: tmpfile.Name()}
	db, err := New(config)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = db.Vacuum()
	assert.NoError(t, err)
}
