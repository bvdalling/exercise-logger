import sqlite3 from 'sqlite3';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { promisify } from 'util';
import fs from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Use data directory if specified (for Docker), otherwise use backend directory
const dataDir = process.env.DATA_DIR || join(__dirname, '..');
const dbPath = join(dataDir, 'gym_app.db');

// Ensure data directory exists
try {
  fs.mkdirSync(dataDir, { recursive: true });
} catch (err) {
  // Directory might already exist, ignore error
}

// Create database connection
const db = new sqlite3.Database(dbPath, (err) => {
  if (err) {
    console.error('Error opening database:', err);
  }
});

// Store original methods before promisifying
const originalRun = db.run.bind(db);

// Promisify database methods
db.get = promisify(db.get.bind(db));
db.all = promisify(db.all.bind(db));
db.exec = promisify(db.exec.bind(db));

// Custom run method that returns lastInsertRowid
db.run = function(sql, ...params) {
  return new Promise((resolve, reject) => {
    originalRun(sql, params, function(err) {
      if (err) {
        reject(err);
      } else {
        resolve({ lastInsertRowid: this.lastID, changes: this.changes });
      }
    });
  });
};

// Enable foreign keys
db.run('PRAGMA foreign_keys = ON').catch(err => {
  console.error('Error enabling foreign keys:', err);
});

// Initialize schema
export async function initializeDatabase() {
  try {
    // Users table
    await db.exec(`
      CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);

    // Exercises table
    await db.exec(`
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
    `);

    // Workout logs table
    await db.exec(`
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
    `);

    // Create indexes for better query performance
    await db.exec(`
      CREATE INDEX IF NOT EXISTS idx_exercises_user_id ON exercises(user_id);
      CREATE INDEX IF NOT EXISTS idx_workout_logs_user_id ON workout_logs(user_id);
      CREATE INDEX IF NOT EXISTS idx_workout_logs_exercise_id ON workout_logs(exercise_id);
      CREATE INDEX IF NOT EXISTS idx_workout_logs_date ON workout_logs(date);
    `);

    // Migrate existing tables to add new columns if they don't exist
    try {
      // Add exercise_type to exercises if it doesn't exist
      await db.run(`
        ALTER TABLE exercises ADD COLUMN exercise_type TEXT DEFAULT 'strength'
      `).catch(() => {
        // Column already exists, ignore error
      });
      
      // Add new columns to workout_logs if they don't exist
      await db.run(`
        ALTER TABLE workout_logs ADD COLUMN weight_per_set TEXT
      `).catch(() => {});
      
      await db.run(`
        ALTER TABLE workout_logs ADD COLUMN rest_time INTEGER
      `).catch(() => {});
      
      await db.run(`
        ALTER TABLE workout_logs ADD COLUMN lap_times TEXT
      `).catch(() => {});
      
      // Add recovery columns to users if they don't exist
      await db.run(`
        ALTER TABLE users ADD COLUMN recovery_uuid TEXT UNIQUE
      `).catch(() => {});
      
      await db.run(`
        ALTER TABLE users ADD COLUMN recovery_secret_hash TEXT
      `).catch(() => {});
    } catch (migrationError) {
      // Migration errors are expected if columns already exist
      console.log('Migration check completed');
    }

    console.log('Database initialized successfully');
  } catch (error) {
    console.error('Error initializing database:', error);
    throw error;
  }
}

export default db;
