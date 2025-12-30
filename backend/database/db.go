package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitializeDatabase sets up the database connection and schema
func InitializeDatabase() error {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "gym_app.db")

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Initialize schema
	if err := createSchema(); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Database initialized successfully")
	return nil
}

func createSchema() error {
	// Users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Exercises table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			exercise_type TEXT DEFAULT 'strength',
			muscle_group TEXT,
			equipment TEXT,
			description TEXT,
			instructions TEXT,
			video_link TEXT,
			image_link TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create exercises table: %w", err)
	}

	// Workout logs table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS workout_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			exercise_id INTEGER NOT NULL,
			date DATE NOT NULL,
			sets INTEGER,
			reps INTEGER,
			weight REAL,
			weight_per_set TEXT,
			rest_time INTEGER,
			distance REAL,
			duration INTEGER,
			pace REAL,
			lap_times TEXT,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create workout_logs table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_exercises_user_id ON exercises(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_workout_logs_user_id ON workout_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_workout_logs_exercise_id ON workout_logs(exercise_id)",
		"CREATE INDEX IF NOT EXISTS idx_workout_logs_date ON workout_logs(date)",
	}

	for _, idx := range indexes {
		if _, err := DB.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func runMigrations() error {
	// Add exercise_type to exercises if it doesn't exist
	_, err := DB.Exec("ALTER TABLE exercises ADD COLUMN exercise_type TEXT DEFAULT 'strength'")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add exercise_type column: %w", err)
	}

	// Add new columns to workout_logs if they don't exist
	workoutLogColumns := []string{
		"ALTER TABLE workout_logs ADD COLUMN weight_per_set TEXT",
		"ALTER TABLE workout_logs ADD COLUMN rest_time INTEGER",
		"ALTER TABLE workout_logs ADD COLUMN lap_times TEXT",
	}

	for _, col := range workoutLogColumns {
		_, err := DB.Exec(col)
		if err != nil && !isColumnExistsError(err) {
			return fmt.Errorf("failed to add column: %w", err)
		}
	}

	// Add recovery columns to users if they don't exist
	// Note: SQLite doesn't allow adding UNIQUE constraint directly when adding a column
	// So we add the column without UNIQUE, then create a unique index separately
	_, err = DB.Exec("ALTER TABLE users ADD COLUMN recovery_uuid TEXT")
	if err != nil && !isColumnExistsError(err) && !strings.Contains(err.Error(), "Cannot add a UNIQUE column") {
		return fmt.Errorf("failed to add recovery_uuid column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN recovery_secret_hash TEXT")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add recovery_secret_hash column: %w", err)
	}

	// Create unique index on recovery_uuid if it doesn't exist
	// This ensures uniqueness even though we couldn't add it as a column constraint
	_, err = DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_recovery_uuid ON users(recovery_uuid)")
	if err != nil {
		// Ignore errors - index might already exist or column might not exist yet
		// This is a best-effort attempt
	}

	return nil
}

func isColumnExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// SQLite returns "duplicate column name: <column_name>" for existing columns
	return errStr != "" && (errStr == "duplicate column name" || strings.Contains(errStr, "duplicate column name"))
}

