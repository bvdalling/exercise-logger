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

	// Seed public exercises
	if err := seedPublicExercises(); err != nil {
		return fmt.Errorf("failed to seed public exercises: %w", err)
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

	// Public exercises table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS public_exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			exercise_type TEXT DEFAULT 'strength',
			muscle_group TEXT,
			equipment TEXT,
			description TEXT,
			instructions TEXT,
			video_link TEXT,
			image_link TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create public_exercises table: %w", err)
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

	// Add email columns to users if they don't exist
	_, err = DB.Exec("ALTER TABLE users ADD COLUMN email TEXT")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add email column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN password_reset_token TEXT")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add password_reset_token column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN password_reset_expires DATETIME")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add password_reset_expires column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN weekly_report_enabled BOOLEAN DEFAULT 1")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add weekly_report_enabled column: %w", err)
	}

	// Create unique index on email if it doesn't exist
	_, err = DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	if err != nil {
		// Ignore errors - index might already exist
	}

	// Add TOTP columns to users if they don't exist
	_, err = DB.Exec("ALTER TABLE users ADD COLUMN totp_secret TEXT")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add totp_secret column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN totp_enabled BOOLEAN DEFAULT 0")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add totp_enabled column: %w", err)
	}

	_, err = DB.Exec("ALTER TABLE users ADD COLUMN totp_backup_codes TEXT")
	if err != nil && !isColumnExistsError(err) {
		return fmt.Errorf("failed to add totp_backup_codes column: %w", err)
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

// seedPublicExercises seeds the database with common public exercises
func seedPublicExercises() error {
	// Check if public exercises already exist
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM public_exercises").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check public exercises: %w", err)
	}

	// If exercises already exist, skip seeding
	if count > 0 {
		return nil
	}

	// Common strength exercises
	strengthExercises := []struct {
		name         string
		muscleGroup  string
		equipment    string
		description  string
		instructions string
	}{
		{
			name:        "Bench Press",
			muscleGroup: "Chest, Triceps, Shoulders",
			equipment:   "Barbell, Bench",
			description: "A compound exercise that targets the chest, shoulders, and triceps.",
			instructions: "Lie on bench, grip bar slightly wider than shoulders. Lower bar to chest, then press up.",
		},
		{
			name:        "Squat",
			muscleGroup: "Quadriceps, Glutes, Hamstrings",
			equipment:   "Barbell",
			description: "A fundamental lower body exercise targeting the quadriceps, glutes, and hamstrings.",
			instructions: "Stand with feet shoulder-width apart, bar on upper back. Lower by bending knees and hips, then stand up.",
		},
		{
			name:        "Deadlift",
			muscleGroup: "Back, Glutes, Hamstrings",
			equipment:   "Barbell",
			description: "A compound exercise that works the entire posterior chain.",
			instructions: "Stand with feet hip-width apart, bar over mid-foot. Hinge at hips, grip bar, then lift by extending hips and knees.",
		},
		{
			name:        "Overhead Press",
			muscleGroup: "Shoulders, Triceps",
			equipment:   "Barbell",
			description: "A shoulder-focused exercise that also works the triceps and core.",
			instructions: "Stand with feet shoulder-width apart, bar at shoulder height. Press bar overhead until arms are fully extended.",
		},
		{
			name:        "Barbell Row",
			muscleGroup: "Back, Biceps",
			equipment:   "Barbell",
			description: "A pulling exercise that targets the back muscles and biceps.",
			instructions: "Bend at hips, grip bar with overhand grip. Pull bar to lower chest/upper abdomen, then lower with control.",
		},
		{
			name:        "Pull-ups",
			muscleGroup: "Back, Biceps",
			equipment:   "Pull-up Bar",
			description: "A bodyweight exercise that targets the back and biceps.",
			instructions: "Hang from bar with palms facing away. Pull body up until chin is over bar, then lower with control.",
		},
		{
			name:        "Dips",
			muscleGroup: "Triceps, Chest, Shoulders",
			equipment:   "Parallel Bars",
			description: "A bodyweight exercise targeting the triceps, chest, and shoulders.",
			instructions: "Support body on parallel bars. Lower by bending arms, then press up to starting position.",
		},
		{
			name:        "Bicep Curls",
			muscleGroup: "Biceps",
			equipment:   "Dumbbells, Barbell",
			description: "An isolation exercise targeting the biceps.",
			instructions: "Stand holding weights at sides. Curl weights up by flexing biceps, then lower with control.",
		},
	}

	// Common cardio exercises
	cardioExercises := []struct {
		name         string
		muscleGroup  string
		equipment    string
		description  string
		instructions string
	}{
		{
			name:        "Running",
			muscleGroup: "Full Body",
			equipment:   "None",
			description: "A cardiovascular exercise that improves endurance and burns calories.",
			instructions: "Start with a warm-up walk, then gradually increase to running pace. Maintain steady breathing.",
		},
		{
			name:        "Cycling",
			muscleGroup: "Legs, Cardiovascular",
			equipment:   "Bicycle",
			description: "A low-impact cardiovascular exercise that strengthens the legs.",
			instructions: "Adjust seat height so leg is almost fully extended at bottom of pedal stroke. Maintain steady cadence.",
		},
		{
			name:        "Rowing",
			muscleGroup: "Full Body",
			equipment:   "Rowing Machine",
			description: "A full-body cardiovascular exercise that works legs, core, and upper body.",
			instructions: "Start with legs extended, lean back slightly, pull handle to chest. Return to starting position in reverse order.",
		},
		{
			name:        "Swimming",
			muscleGroup: "Full Body",
			equipment:   "Pool",
			description: "A full-body, low-impact cardiovascular exercise.",
			instructions: "Use proper stroke technique. Focus on breathing rhythm and efficient movement through the water.",
		},
	}

	// Insert strength exercises
	for _, ex := range strengthExercises {
		_, err := DB.Exec(
			`INSERT INTO public_exercises (name, exercise_type, muscle_group, equipment, description, instructions)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			ex.name, "strength", ex.muscleGroup, ex.equipment, ex.description, ex.instructions,
		)
		if err != nil {
			return fmt.Errorf("failed to insert public exercise %s: %w", ex.name, err)
		}
	}

	// Insert cardio exercises
	for _, ex := range cardioExercises {
		_, err := DB.Exec(
			`INSERT INTO public_exercises (name, exercise_type, muscle_group, equipment, description, instructions)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			ex.name, "cardio", ex.muscleGroup, ex.equipment, ex.description, ex.instructions,
		)
		if err != nil {
			return fmt.Errorf("failed to insert public exercise %s: %w", ex.name, err)
		}
	}

	fmt.Println("Public exercises seeded successfully")
	return nil
}

