package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gym-app-backend/database"
	"gym-app-backend/middleware"
	"gym-app-backend/models"
)

type WorkoutLogResponse struct {
	Log models.WorkoutLog `json:"log"`
}

type WorkoutLogsResponse struct {
	Logs []models.WorkoutLog `json:"logs"`
}

type LastWorkoutResponse struct {
	LastLog *models.WorkoutLog `json:"lastLog"`
}

type CreateWorkoutLogRequest struct {
	ExerciseID   int64        `json:"exercise_id"`
	Date         string       `json:"date"`
	Sets         *int         `json:"sets"`
	Reps         *int         `json:"reps"`
	Weight       *float64     `json:"weight"`
	WeightPerSet interface{} `json:"weight_per_set"`
	RestTime     *int         `json:"rest_time"`
	Distance     *float64     `json:"distance"`
	Duration     *int         `json:"duration"`
	Pace         *float64     `json:"pace"`
	LapTimes     interface{}  `json:"lap_times"`
	Notes        *string      `json:"notes"`
}

type UpdateWorkoutLogRequest struct {
	ExerciseID   *int64       `json:"exercise_id"`
	Date         *string      `json:"date"`
	Sets         *int          `json:"sets"`
	Reps         *int         `json:"reps"`
	Weight       *float64     `json:"weight"`
	WeightPerSet interface{}  `json:"weight_per_set"`
	RestTime     *int         `json:"rest_time"`
	Distance     *float64     `json:"distance"`
	Duration     *int         `json:"duration"`
	Pace         *float64     `json:"pace"`
	LapTimes     interface{}  `json:"lap_times"`
	Notes        *string      `json:"notes"`
}

// GetAllWorkoutLogs returns all workout logs for the authenticated user
func GetAllWorkoutLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	query := `
		SELECT wl.*, e.name as exercise_name, e.exercise_type
		FROM workout_logs wl
		JOIN exercises e ON wl.exercise_id = e.id
		WHERE wl.user_id = ?
	`
	params := []interface{}{userID}

	// Add filters
	if exerciseIDStr := r.URL.Query().Get("exercise_id"); exerciseIDStr != "" {
		if exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64); err == nil {
			query += " AND wl.exercise_id = ?"
			params = append(params, exerciseID)
		}
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		query += " AND wl.date >= ?"
		params = append(params, startDate)
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		query += " AND wl.date <= ?"
		params = append(params, endDate)
	}

	query += " ORDER BY wl.date DESC, wl.created_at DESC"

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			query += " LIMIT ?"
			params = append(params, limit)
		}
	}

	rows, err := database.DB.Query(query, params...)
	if err != nil {
		fmt.Printf("Get workout logs error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.WorkoutLog
	for rows.Next() {
		var log models.WorkoutLog
		var weightPerSetStr, lapTimesStr sql.NullString
		var createdAtStr string
		err := rows.Scan(
			&log.ID, &log.UserID, &log.ExerciseID, &log.Date, &log.Sets, &log.Reps,
			&log.Weight, &weightPerSetStr, &log.RestTime, &log.Distance, &log.Duration,
			&log.Pace, &lapTimesStr, &log.Notes, &createdAtStr, &log.ExerciseName, &log.ExerciseType,
		)
		if err != nil {
			fmt.Printf("Error scanning log: %v\n", err)
			continue
		}

		// Parse JSON fields
		if weightPerSetStr.Valid && weightPerSetStr.String != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(weightPerSetStr.String), &parsed); err == nil {
				log.WeightPerSet = parsed
			}
		}
		if lapTimesStr.Valid && lapTimesStr.String != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(lapTimesStr.String), &parsed); err == nil {
				log.LapTimes = parsed
			}
		}

		logs = append(logs, log)
	}

	response := WorkoutLogsResponse{Logs: logs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWorkoutLogById returns a single workout log by ID
func GetWorkoutLogById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	logID, err := strconv.ParseInt(r.URL.Path[len("/api/workout-logs/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid log ID"}`, http.StatusBadRequest)
		return
	}

	var log models.WorkoutLog
	var weightPerSetStr, lapTimesStr sql.NullString
	var createdAtStr string
	err = database.DB.QueryRow(
		`SELECT wl.*, e.name as exercise_name, e.exercise_type
		 FROM workout_logs wl
		 JOIN exercises e ON wl.exercise_id = e.id
		 WHERE wl.id = ? AND wl.user_id = ?`,
		logID, userID,
	).Scan(
		&log.ID, &log.UserID, &log.ExerciseID, &log.Date, &log.Sets, &log.Reps,
		&log.Weight, &weightPerSetStr, &log.RestTime, &log.Distance, &log.Duration,
		&log.Pace, &lapTimesStr, &log.Notes, &createdAtStr, &log.ExerciseName, &log.ExerciseType,
	)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Workout log not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Parse JSON fields
	if weightPerSetStr.Valid && weightPerSetStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(weightPerSetStr.String), &parsed); err == nil {
			log.WeightPerSet = parsed
		}
	}
	if lapTimesStr.Valid && lapTimesStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(lapTimesStr.String), &parsed); err == nil {
			log.LapTimes = parsed
		}
	}

	response := WorkoutLogResponse{Log: log}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateWorkoutLog creates a new workout log
func CreateWorkoutLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req CreateWorkoutLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ExerciseID == 0 || req.Date == "" {
		http.Error(w, `{"error":"Exercise ID and date are required"}`, http.StatusBadRequest)
		return
	}

	// Verify exercise belongs to user and get exercise type
	var exerciseType string
	err := database.DB.QueryRow(
		"SELECT exercise_type FROM exercises WHERE id = ? AND user_id = ?",
		req.ExerciseID, userID,
	).Scan(&exerciseType)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Create workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Validate: distance/time should only be used for cardio
	if exerciseType != "cardio" && (req.Distance != nil || req.Duration != nil || req.Pace != nil || req.LapTimes != nil) {
		http.Error(w, `{"error":"Distance, duration, pace, and lap times can only be used for cardio exercises"}`, http.StatusBadRequest)
		return
	}

	// Validate: weight/weight_per_set should only be used for strength
	if exerciseType != "strength" && (req.Weight != nil || req.WeightPerSet != nil) {
		http.Error(w, `{"error":"Weight and weight per set can only be used for strength exercises"}`, http.StatusBadRequest)
		return
	}

	// Serialize JSON fields
	var weightPerSetStr, lapTimesStr sql.NullString
	if req.WeightPerSet != nil {
		data, err := json.Marshal(req.WeightPerSet)
		if err == nil {
			weightPerSetStr = sql.NullString{String: string(data), Valid: true}
		}
	}
	if req.LapTimes != nil {
		data, err := json.Marshal(req.LapTimes)
		if err == nil {
			lapTimesStr = sql.NullString{String: string(data), Valid: true}
		}
	}

	result, err := database.DB.Exec(
		`INSERT INTO workout_logs (user_id, exercise_id, date, sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.ExerciseID, req.Date, req.Sets, req.Reps, req.Weight,
		weightPerSetStr, req.RestTime, req.Distance, req.Duration, req.Pace, lapTimesStr, req.Notes,
	)
	if err != nil {
		fmt.Printf("Create workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	logID, _ := result.LastInsertId()

	var log models.WorkoutLog
	var weightPerSetStr2, lapTimesStr2 sql.NullString
	var createdAtStr string
	err = database.DB.QueryRow(
		`SELECT wl.*, e.name as exercise_name, e.exercise_type
		 FROM workout_logs wl
		 JOIN exercises e ON wl.exercise_id = e.id
		 WHERE wl.id = ?`,
		logID,
	).Scan(
		&log.ID, &log.UserID, &log.ExerciseID, &log.Date, &log.Sets, &log.Reps,
		&log.Weight, &weightPerSetStr2, &log.RestTime, &log.Distance, &log.Duration,
		&log.Pace, &lapTimesStr2, &log.Notes, &createdAtStr, &log.ExerciseName, &log.ExerciseType,
	)
	if err != nil {
		fmt.Printf("Error fetching created log: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Parse JSON fields
	if weightPerSetStr2.Valid && weightPerSetStr2.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(weightPerSetStr2.String), &parsed); err == nil {
			log.WeightPerSet = parsed
		}
	}
	if lapTimesStr2.Valid && lapTimesStr2.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(lapTimesStr2.String), &parsed); err == nil {
			log.LapTimes = parsed
		}
	}

	response := WorkoutLogResponse{Log: log}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateWorkoutLog updates an existing workout log
func UpdateWorkoutLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	logID, err := strconv.ParseInt(r.URL.Path[len("/api/workout-logs/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid log ID"}`, http.StatusBadRequest)
		return
	}

	// Verify log belongs to user
	var existingExerciseID int64
	err = database.DB.QueryRow(
		"SELECT exercise_id FROM workout_logs WHERE id = ? AND user_id = ?",
		logID, userID,
	).Scan(&existingExerciseID)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Workout log not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Update workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var req UpdateWorkoutLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Get exercise type for validation
	currentExerciseID := existingExerciseID
	if req.ExerciseID != nil {
		currentExerciseID = *req.ExerciseID
	}

	var exerciseType string
	err = database.DB.QueryRow(
		"SELECT exercise_type FROM exercises WHERE id = ? AND user_id = ?",
		currentExerciseID, userID,
	).Scan(&exerciseType)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Update workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Validate: distance/time should only be used for cardio
	if exerciseType != "cardio" && (req.Distance != nil || req.Duration != nil || req.Pace != nil || req.LapTimes != nil) {
		http.Error(w, `{"error":"Distance, duration, pace, and lap times can only be used for cardio exercises"}`, http.StatusBadRequest)
		return
	}

	// Validate: weight/weight_per_set should only be used for strength
	if exerciseType != "strength" && (req.Weight != nil || req.WeightPerSet != nil) {
		http.Error(w, `{"error":"Weight and weight per set can only be used for strength exercises"}`, http.StatusBadRequest)
		return
	}

	// Build update query dynamically
	updates := []string{}
	values := []interface{}{}

	if req.ExerciseID != nil {
		updates = append(updates, "exercise_id = ?")
		values = append(values, *req.ExerciseID)
	}
	if req.Date != nil {
		updates = append(updates, "date = ?")
		values = append(values, *req.Date)
	}
	if req.Sets != nil {
		updates = append(updates, "sets = ?")
		values = append(values, *req.Sets)
	}
	if req.Reps != nil {
		updates = append(updates, "reps = ?")
		values = append(values, *req.Reps)
	}
	if req.Weight != nil {
		updates = append(updates, "weight = ?")
		values = append(values, *req.Weight)
	}
	if req.WeightPerSet != nil {
		data, err := json.Marshal(req.WeightPerSet)
		if err == nil {
			updates = append(updates, "weight_per_set = ?")
			values = append(values, string(data))
		}
	}
	if req.RestTime != nil {
		updates = append(updates, "rest_time = ?")
		values = append(values, *req.RestTime)
	}
	if req.Distance != nil {
		updates = append(updates, "distance = ?")
		values = append(values, *req.Distance)
	}
	if req.Duration != nil {
		updates = append(updates, "duration = ?")
		values = append(values, *req.Duration)
	}
	if req.Pace != nil {
		updates = append(updates, "pace = ?")
		values = append(values, *req.Pace)
	}
	if req.LapTimes != nil {
		data, err := json.Marshal(req.LapTimes)
		if err == nil {
			updates = append(updates, "lap_times = ?")
			values = append(values, string(data))
		}
	}
	if req.Notes != nil {
		updates = append(updates, "notes = ?")
		values = append(values, *req.Notes)
	}

	if len(updates) > 0 {
		values = append(values, logID, userID)
		query := fmt.Sprintf("UPDATE workout_logs SET %s WHERE id = ? AND user_id = ?", strings.Join(updates, ", "))
		_, err = database.DB.Exec(query, values...)
		if err != nil {
			fmt.Printf("Update workout log error: %v\n", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	var log models.WorkoutLog
	var weightPerSetStr, lapTimesStr sql.NullString
	var createdAtStr string
	err = database.DB.QueryRow(
		`SELECT wl.*, e.name as exercise_name, e.exercise_type
		 FROM workout_logs wl
		 JOIN exercises e ON wl.exercise_id = e.id
		 WHERE wl.id = ?`,
		logID,
	).Scan(
		&log.ID, &log.UserID, &log.ExerciseID, &log.Date, &log.Sets, &log.Reps,
		&log.Weight, &weightPerSetStr, &log.RestTime, &log.Distance, &log.Duration,
		&log.Pace, &lapTimesStr, &log.Notes, &createdAtStr, &log.ExerciseName, &log.ExerciseType,
	)
	if err != nil {
		fmt.Printf("Error fetching updated log: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Parse JSON fields
	if weightPerSetStr.Valid && weightPerSetStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(weightPerSetStr.String), &parsed); err == nil {
			log.WeightPerSet = parsed
		}
	}
	if lapTimesStr.Valid && lapTimesStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(lapTimesStr.String), &parsed); err == nil {
			log.LapTimes = parsed
		}
	}

	response := WorkoutLogResponse{Log: log}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteWorkoutLog deletes a workout log
func DeleteWorkoutLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	logID, err := strconv.ParseInt(r.URL.Path[len("/api/workout-logs/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid log ID"}`, http.StatusBadRequest)
		return
	}

	// Verify log belongs to user
	var existingID int64
	err = database.DB.QueryRow(
		"SELECT id FROM workout_logs WHERE id = ? AND user_id = ?",
		logID, userID,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Workout log not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Delete workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("DELETE FROM workout_logs WHERE id = ? AND user_id = ?", logID, userID)
	if err != nil {
		fmt.Printf("Delete workout log error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Workout log deleted successfully"})
}

// GetLastWorkoutValues returns the most recent workout log for an exercise
func GetLastWorkoutValues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	// Extract exercise ID from path like /api/workout-logs/exercise/1/last
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var exerciseIDStr string
	for i, part := range parts {
		if part == "exercise" && i+1 < len(parts) {
			exerciseIDStr = parts[i+1]
			break
		}
	}

	if exerciseIDStr == "" {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	// Verify exercise belongs to user
	var exerciseExists int64
	err = database.DB.QueryRow(
		"SELECT id FROM exercises WHERE id = ? AND user_id = ?",
		exerciseID, userID,
	).Scan(&exerciseExists)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get last workout values error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Get the most recent workout log for this exercise
	var log models.WorkoutLog
	var weightPerSetStr, lapTimesStr sql.NullString
	err = database.DB.QueryRow(
		`SELECT sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, date
		 FROM workout_logs
		 WHERE exercise_id = ? AND user_id = ?
		 ORDER BY date DESC, created_at DESC
		 LIMIT 1`,
		exerciseID, userID,
	).Scan(
		&log.Sets, &log.Reps, &log.Weight, &weightPerSetStr, &log.RestTime,
		&log.Distance, &log.Duration, &log.Pace, &lapTimesStr, &log.Date,
	)

	if err == sql.ErrNoRows {
		response := LastWorkoutResponse{LastLog: nil}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	} else if err != nil {
		fmt.Printf("Get last workout values error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Parse JSON fields
	if weightPerSetStr.Valid && weightPerSetStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(weightPerSetStr.String), &parsed); err == nil {
			log.WeightPerSet = parsed
		}
	}
	if lapTimesStr.Valid && lapTimesStr.String != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(lapTimesStr.String), &parsed); err == nil {
			log.LapTimes = parsed
		}
	}

	response := LastWorkoutResponse{LastLog: &log}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

