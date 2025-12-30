package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gym-app-backend/database"
	"gym-app-backend/models"
)

type PublicExerciseResponse struct {
	Exercise models.PublicExercise `json:"exercise"`
}

type PublicExercisesResponse struct {
	Exercises []models.PublicExercise `json:"exercises"`
}

// GetAllPublicExercises returns all public exercises
func GetAllPublicExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query(
		"SELECT * FROM public_exercises ORDER BY name ASC",
	)
	if err != nil {
		fmt.Printf("Get public exercises error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var exercises []models.PublicExercise
	for rows.Next() {
		var ex models.PublicExercise
		var createdAtStr string
		err := rows.Scan(
			&ex.ID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
			&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
			&ex.ImageLink, &createdAtStr,
		)
		if err != nil {
			fmt.Printf("Error scanning public exercise: %v\n", err)
			continue
		}
		exercises = append(exercises, ex)
	}

	response := PublicExercisesResponse{Exercises: exercises}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPublicExerciseById returns a single public exercise by ID
func GetPublicExerciseById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	exerciseID, err := strconv.ParseInt(r.URL.Path[len("/api/public-exercises/"):], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid exercise ID"}`, http.StatusBadRequest)
		return
	}

	var ex models.PublicExercise
	var createdAtStr string
	err = database.DB.QueryRow(
		"SELECT * FROM public_exercises WHERE id = ?",
		exerciseID,
	).Scan(
		&ex.ID, &ex.Name, &ex.ExerciseType, &ex.MuscleGroup,
		&ex.Equipment, &ex.Description, &ex.Instructions, &ex.VideoLink,
		&ex.ImageLink, &createdAtStr,
	)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Public exercise not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get public exercise error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := PublicExerciseResponse{Exercise: ex}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

