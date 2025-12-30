package models

import "time"

type Exercise struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Name         string    `json:"name"`
	ExerciseType string    `json:"exercise_type"`
	MuscleGroup  *string   `json:"muscle_group"`
	Equipment    *string   `json:"equipment"`
	Description  *string   `json:"description"`
	Instructions *string   `json:"instructions"`
	VideoLink    *string   `json:"video_link"`
	ImageLink    *string   `json:"image_link"`
	CreatedAt    time.Time `json:"created_at"`
}

