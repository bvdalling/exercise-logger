package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gym-app-backend/database"
	"gym-app-backend/middleware"
	"gym-app-backend/models"
	"gym-app-backend/services"
)

// SendWeeklyReport sends a weekly workout report via email
func SendWeeklyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	// Get user email and weekly report preference
	var email sql.NullString
	var weeklyReportEnabled sql.NullBool
	err := database.DB.QueryRow(
		"SELECT email, weekly_report_enabled FROM users WHERE id = ?",
		userID,
	).Scan(&email, &weeklyReportEnabled)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get user error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if !email.Valid || email.String == "" {
		http.Error(w, `{"error":"Email not set for user"}`, http.StatusBadRequest)
		return
	}

	if weeklyReportEnabled.Valid && !weeklyReportEnabled.Bool {
		http.Error(w, `{"error":"Weekly reports are disabled for this account"}`, http.StatusBadRequest)
		return
	}

	// Calculate week range (Sunday to Saturday)
	now := time.Now()
	weekStart := now
	for weekStart.Weekday() != time.Sunday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 0, weekEnd.Location())

	startDate := weekStart.Format("2006-01-02")
	endDate := weekEnd.Format("2006-01-02")

	// Get workout logs for the week
	rows, err := database.DB.Query(
		`SELECT wl.*, 
		       COALESCE(e.name, pe.name) as exercise_name,
		       COALESCE(e.exercise_type, pe.exercise_type) as exercise_type
		 FROM workout_logs wl
		 LEFT JOIN exercises e ON wl.exercise_id = e.id AND wl.user_id = e.user_id
		 LEFT JOIN public_exercises pe ON wl.exercise_id = pe.id
		 WHERE wl.user_id = ? AND wl.date >= ? AND wl.date <= ?
		 ORDER BY wl.date ASC, wl.created_at ASC`,
		userID, startDate, endDate,
	)
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

	// Generate HTML report
	reportHTML := generateWeeklyReportHTML(logs, weekStart, weekEnd)

	// Send email
	if services.EmailService != nil {
		if err := services.EmailService.SendWeeklyReport(email.String, reportHTML); err != nil {
			fmt.Printf("Failed to send weekly report: %v\n", err)
			http.Error(w, `{"error":"Failed to send email"}`, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, `{"error":"Email service not configured"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Weekly report sent successfully"})
}

func generateWeeklyReportHTML(logs []models.WorkoutLog, weekStart, weekEnd time.Time) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Weekly Workout Report</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #4CAF50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
		.workout { background: #f9f9f9; padding: 15px; margin-bottom: 10px; border-radius: 5px; border-left: 4px solid #4CAF50; }
		.exercise-name { font-weight: bold; font-size: 1.1em; margin-bottom: 5px; }
		.stats { color: #666; font-size: 0.9em; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; font-size: 0.9em; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Your Weekly Workout Report</h1>
			<p>%s - %s</p>
		</div>
	`, weekStart.Format("January 2, 2006"), weekEnd.Format("January 2, 2006"))

	if len(logs) == 0 {
		html += `<p>No workouts logged this week. Keep pushing!</p>`
	} else {
		html += fmt.Sprintf(`<p><strong>Total Workouts:</strong> %d</p>`, len(logs))
		
		// Group by date
		logsByDate := make(map[string][]models.WorkoutLog)
		for _, log := range logs {
			date := log.Date
			logsByDate[date] = append(logsByDate[date], log)
		}

		for date, dayLogs := range logsByDate {
			html += fmt.Sprintf(`<h2>%s</h2>`, formatDateForReport(date))
			for _, log := range dayLogs {
				html += `<div class="workout">`
				html += fmt.Sprintf(`<div class="exercise-name">%s</div>`, *log.ExerciseName)
				
				if log.ExerciseType != nil && *log.ExerciseType == "strength" {
					if log.WeightPerSet != nil {
						if sets, ok := log.WeightPerSet.([]interface{}); ok {
							for i, set := range sets {
								if setMap, ok := set.(map[string]interface{}); ok {
									html += fmt.Sprintf(`<div class="stats">Set %d: `, i+1)
									if reps, ok := setMap["reps"].(float64); ok {
										html += fmt.Sprintf(`%.0f reps`, reps)
									}
									if weight, ok := setMap["weight"].(float64); ok {
										html += fmt.Sprintf(` @ %.1flbs`, weight)
									}
									html += `</div>`
								}
							}
						}
					} else {
						if log.Sets != nil && log.Reps != nil {
							html += fmt.Sprintf(`<div class="stats">Sets: %d, Reps: %d</div>`, *log.Sets, *log.Reps)
						}
						if log.Weight != nil {
							html += fmt.Sprintf(`<div class="stats">Weight: %.1flbs</div>`, *log.Weight)
						}
					}
				} else {
					if log.Distance != nil {
						html += fmt.Sprintf(`<div class="stats">Distance: %.2f miles</div>`, *log.Distance)
					}
					if log.Duration != nil {
						hours := *log.Duration / 60
						minutes := *log.Duration % 60
						html += fmt.Sprintf(`<div class="stats">Duration: %dh %dm</div>`, hours, minutes)
					}
					if log.Pace != nil {
						html += fmt.Sprintf(`<div class="stats">Pace: %.1f min/mile</div>`, *log.Pace)
					}
				}
				
				if log.Notes != nil && *log.Notes != "" {
					html += fmt.Sprintf(`<div class="stats" style="margin-top: 5px; font-style: italic;">Notes: %s</div>`, *log.Notes)
				}
				
				html += `</div>`
			}
		}
	}

	html += `
		<div class="footer">
			<p>Keep up the great work! 💪</p>
			<p>This is an automated email from your Gym App.</p>
		</div>
	</div>
</body>
</html>`

	return html
}

func formatDateForReport(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Monday, January 2, 2006")
}

