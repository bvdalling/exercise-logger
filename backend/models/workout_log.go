package models

import (
	"encoding/json"
	"time"
)

type WorkoutLog struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	ExerciseID   int64     `json:"exercise_id"`
	ExerciseName *string   `json:"exercise_name,omitempty"`
	ExerciseType *string   `json:"exercise_type,omitempty"`
	Date         string    `json:"date"`
	Sets         *int      `json:"sets"`
	Reps         *int      `json:"reps"`
	Weight       *float64  `json:"weight"`
	WeightPerSet interface{} `json:"weight_per_set"` // Can be array or null
	RestTime     *int      `json:"rest_time"`
	Distance     *float64  `json:"distance"`
	Duration     *int      `json:"duration"`
	Pace         *float64  `json:"pace"`
	LapTimes     interface{} `json:"lap_times"` // Can be array or null
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
}

// WeightPerSetString returns weight_per_set as JSON string for database storage
func (w *WorkoutLog) WeightPerSetString() *string {
	if w.WeightPerSet == nil {
		return nil
	}
	data, err := json.Marshal(w.WeightPerSet)
	if err != nil {
		return nil
	}
	str := string(data)
	return &str
}

// LapTimesString returns lap_times as JSON string for database storage
func (w *WorkoutLog) LapTimesString() *string {
	if w.LapTimes == nil {
		return nil
	}
	data, err := json.Marshal(w.LapTimes)
	if err != nil {
		return nil
	}
	str := string(data)
	return &str
}

// ParseJSONFields parses JSON fields from database strings
func (w *WorkoutLog) ParseJSONFields() {
	if w.WeightPerSet != nil {
		if str, ok := w.WeightPerSet.(string); ok && str != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(str), &parsed); err == nil {
				w.WeightPerSet = parsed
			} else {
				w.WeightPerSet = nil
			}
		}
	}
	if w.LapTimes != nil {
		if str, ok := w.LapTimes.(string); ok && str != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(str), &parsed); err == nil {
				w.LapTimes = parsed
			} else {
				w.LapTimes = nil
			}
		}
	}
}

