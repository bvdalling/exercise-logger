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

type ExerciseResponse struct {
	Exercise models.Exercise `json:"exercise"`
}

type ExercisesResponse struct {
	Exercises []models.Exercise `json:"exercises"`
}

type ProgressResponse struct {
	Progress []models.WorkoutLog `json:"progress"`
}

// GetAllExercises returns all exercises for the authenticated user
func GetAllExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	rows, err := database.DB.Query(
		"SELECT * FROM exercises WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		fmt.Printf("Get exercises error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		var createdAtStr string
		err := rows.Scan(
			&ex.ID, &ex.UserID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
			&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
			&ex.ImageLink, &createdAtStr,
		)
		if err != nil {
			fmt.Printf("Error scanning exercise: %v\n", err)
			continue
		}
		// Parse created_at if needed (simplified - you may want proper time parsing)
		exercises = append(exercises, ex)
	}

	response := ExercisesResponse{Exercises: exercises}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExerciseById returns a single exercise by ID
func GetExerciseById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	exerciseID, err := strconv.ParseInt(r.URL.Path[len("/api/exercises/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	var ex models.Exercise
	var createdAtStr string
	err = database.DB.QueryRow(
		"SELECT * FROM exercises WHERE id = ? AND user_id = ?",
		exerciseID, userID,
	).Scan(
		&ex.ID, &ex.UserID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
		&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
		&ex.ImageLink, &createdAtStr,
	)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := ExerciseResponse{Exercise: ex}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type CreateExerciseRequest struct {
	Name         string  `json:"name"`
	ExerciseType *string `json:"exercise_type"`
	MuscleGroup  *string `json:"muscle_group"`
	Equipment    *string `json:"equipment"`
	Description  *string `json:"description"`
	Instructions *string `json:"instructions"`
	VideoLink    *string `json:"video_link"`
	ImageLink    *string `json:"image_link"`
}

// CreateExercise creates a new exercise
func CreateExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"Exercise name is required"}`, http.StatusBadRequest)
		return
	}

	// Validate exercise_type
	exerciseType := "strength"
	if req.ExerciseType != nil {
		if *req.ExerciseType == "strength" || *req.ExerciseType == "cardio" {
			exerciseType = *req.ExerciseType
		}
	}

	result, err := database.DB.Exec(
		`INSERT INTO exercises (user_id, name, exercise_type, muscle_group, equipment, description, instructions, video_link, image_link)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.Name, exerciseType, req.MuscleGroup, req.Equipment,
		req.Description, req.Instructions, req.VideoLink, req.ImageLink,
	)
	if err != nil {
		fmt.Printf("Create exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	exerciseID, _ := result.LastInsertId()

	var ex models.Exercise
	var createdAtStr string
	err = database.DB.QueryRow(
		"SELECT * FROM exercises WHERE id = ?",
		exerciseID,
	).Scan(
		&ex.ID, &ex.UserID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
		&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
		&ex.ImageLink, &createdAtStr,
	)
	if err != nil {
		fmt.Printf("Error fetching created exercise: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := ExerciseResponse{Exercise: ex}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type UpdateExerciseRequest struct {
	Name         *string `json:"name"`
	ExerciseType *string `json:"exercise_type"`
	MuscleGroup  *string `json:"muscle_group"`
	Equipment    *string `json:"equipment"`
	Description  *string `json:"description"`
	Instructions *string `json:"instructions"`
	VideoLink    *string `json:"video_link"`
	ImageLink    *string `json:"image_link"`
}

// UpdateExercise updates an existing exercise
func UpdateExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	exerciseID, err := strconv.ParseInt(r.URL.Path[len("/api/exercises/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	// Verify exercise belongs to user
	var existingID int64
	err = database.DB.QueryRow(
		"SELECT id FROM exercises WHERE id = ? AND user_id = ?",
		exerciseID, userID,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Update exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var req UpdateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Build update query dynamically
	updates := []string{}
	values := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		values = append(values, *req.Name)
	}
	if req.ExerciseType != nil {
		validTypes := []string{"strength", "cardio"}
		exerciseType := "strength"
		for _, vt := range validTypes {
			if *req.ExerciseType == vt {
				exerciseType = *req.ExerciseType
				break
			}
		}
		updates = append(updates, "exercise_type = ?")
		values = append(values, exerciseType)
	}
	if req.MuscleGroup != nil {
		updates = append(updates, "muscle_group = ?")
		values = append(values, *req.MuscleGroup)
	}
	if req.Equipment != nil {
		updates = append(updates, "equipment = ?")
		values = append(values, *req.Equipment)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		values = append(values, *req.Description)
	}
	if req.Instructions != nil {
		updates = append(updates, "instructions = ?")
		values = append(values, *req.Instructions)
	}
	if req.VideoLink != nil {
		updates = append(updates, "video_link = ?")
		values = append(values, *req.VideoLink)
	}
	if req.ImageLink != nil {
		updates = append(updates, "image_link = ?")
		values = append(values, *req.ImageLink)
	}

	if len(updates) > 0 {
		values = append(values, exerciseID, userID)
		query := fmt.Sprintf("UPDATE exercises SET %s WHERE id = ? AND user_id = ?", strings.Join(updates, ", "))

		_, err = database.DB.Exec(query, values...)
		if err != nil {
			fmt.Printf("Update exercise error: %v\n", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	var ex models.Exercise
	var createdAtStr string
	err = database.DB.QueryRow(
		"SELECT * FROM exercises WHERE id = ?",
		exerciseID,
	).Scan(
		&ex.ID, &ex.UserID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
		&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
		&ex.ImageLink, &createdAtStr,
	)
	if err != nil {
		fmt.Printf("Error fetching updated exercise: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := ExerciseResponse{Exercise: ex}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteExercise deletes an exercise
func DeleteExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	exerciseID, err := strconv.ParseInt(r.URL.Path[len("/api/exercises/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	// Verify exercise belongs to user
	var existingID int64
	err = database.DB.QueryRow(
		"SELECT id FROM exercises WHERE id = ? AND user_id = ?",
		exerciseID, userID,
	).Scan(&existingID)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Delete exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("DELETE FROM exercises WHERE id = ? AND user_id = ?", exerciseID, userID)
	if err != nil {
		fmt.Printf("Delete exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Exercise deleted successfully"})
}

// GetExerciseProgress returns workout logs for an exercise ordered by date
func GetExerciseProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	// Extract exercise ID from path like /api/exercises/1/progress
	path := r.URL.Path
	// Find the last two slashes
	lastSlash := -1
	secondLastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if lastSlash == -1 {
				lastSlash = i
			} else if secondLastSlash == -1 {
				secondLastSlash = i
				break
			}
		}
	}
	if secondLastSlash == -1 {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}
	exerciseIDStr := path[secondLastSlash+1 : lastSlash]
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
		fmt.Printf("Get progress error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Get all workout logs for this exercise
	rows, err := database.DB.Query(
		`SELECT id, date, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, sets, reps, notes
		 FROM workout_logs
		 WHERE exercise_id = ? AND user_id = ?
		 ORDER BY date ASC`,
		exerciseID, userID,
	)
	if err != nil {
		fmt.Printf("Get progress error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.WorkoutLog
	for rows.Next() {
		var log models.WorkoutLog
		var weightPerSetStr, lapTimesStr sql.NullString
		err := rows.Scan(
			&log.ID, &log.Date, &log.Weight, &weightPerSetStr, &log.RestTime,
			&log.Distance, &log.Duration, &log.Pace, &lapTimesStr,
			&log.Sets, &log.Reps, &log.Notes,
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

	response := ProgressResponse{Progress: logs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
