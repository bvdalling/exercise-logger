package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gym-app-backend/database"
	"gym-app-backend/handlers"
	"gym-app-backend/middleware"
	"gym-app-backend/services"
)

func main() {
	// Initialize database
	if err := database.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize email service
	if err := services.InitializeEmailService(); err != nil {
		log.Printf("Warning: Failed to initialize email service: %v (email features will be disabled)", err)
	}

	// Get configuration from environment
	port := "3111"
	host := "127.0.0.1" // Default to localhost-only for internal access

	// Setup routes
	mux := http.NewServeMux()

	// Health check (accept both GET and HEAD for Docker health checks)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	})

	// Auth routes
	mux.HandleFunc("/api/auth/register", handlers.Register)
	mux.HandleFunc("/api/auth/login", handlers.Login)
	mux.HandleFunc("/api/auth/logout", handlers.Logout)
	mux.HandleFunc("/api/auth/me", middleware.RequireAuth(http.HandlerFunc(handlers.GetCurrentUser)).ServeHTTP)
	mux.HandleFunc("/api/auth/request-password-reset", handlers.RequestPasswordReset)
	mux.HandleFunc("/api/auth/reset-password", handlers.ResetPassword)
	mux.HandleFunc("/api/auth/setup-totp", middleware.RequireAuth(http.HandlerFunc(handlers.SetupTOTP)).ServeHTTP)
	mux.HandleFunc("/api/auth/verify-totp", handlers.VerifyTOTP)

	// Reports routes
	mux.HandleFunc("/api/reports/weekly", middleware.RequireAuth(http.HandlerFunc(handlers.SendWeeklyReport)).ServeHTTP)

	// Public exercise routes (no auth required)
	mux.HandleFunc("/api/public-exercises", handlers.GetAllPublicExercises)
	mux.HandleFunc("/api/public-exercises/", handlers.GetPublicExerciseById)

	// Exercise routes - exact match for list/create (with auth)
	mux.HandleFunc("/api/exercises", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetAllExercises(w, r)
		case http.MethodPost:
			handlers.CreateExercise(w, r)
		default:
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})).ServeHTTP)

	// Exercise routes with ID - use a pattern matcher (with auth)
	mux.HandleFunc("/api/exercises/", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/progress") {
			handlers.GetExerciseProgress(w, r)
		} else {
			// Handle /api/exercises/:id
			switch r.Method {
			case http.MethodGet:
				handlers.GetExerciseById(w, r)
			case http.MethodPut:
				handlers.UpdateExercise(w, r)
			case http.MethodDelete:
				handlers.DeleteExercise(w, r)
			default:
				http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			}
		}
	})).ServeHTTP)

	// Workout log routes - exact match for list/create (with auth)
	mux.HandleFunc("/api/workout-logs", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetAllWorkoutLogs(w, r)
		case http.MethodPost:
			handlers.CreateWorkoutLog(w, r)
		default:
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})).ServeHTTP)

	// Workout log routes with path - handle /api/workout-logs/exercise/:id/last and /api/workout-logs/:id (with auth)
	mux.HandleFunc("/api/workout-logs/", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if it's the special route /api/workout-logs/exercise/:id/last
		if strings.Contains(path, "/exercise/") && strings.HasSuffix(path, "/last") {
			handlers.GetLastWorkoutValues(w, r)
			return
		}

		// Handle /api/workout-logs/:id
		switch r.Method {
		case http.MethodGet:
			handlers.GetWorkoutLogById(w, r)
		case http.MethodPut:
			handlers.UpdateWorkoutLog(w, r)
		case http.MethodDelete:
			handlers.DeleteWorkoutLog(w, r)
		default:
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})).ServeHTTP)

	// Apply middleware
	handler := middleware.Logging(middleware.CORS(mux))

	// Error handling middleware (apply before 404 handler)
	handler = errorHandler(handler)

	// Wrap with 404 handler (capture handler before reassigning)
	baseHandler := handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response recorder to check if route was handled
		recorder := &statusRecorder{ResponseWriter: w, status: 0}
		baseHandler.ServeHTTP(recorder, r)

		// If status wasn't set (still 0), it's a 404
		// Status 0 means no handler wrote a response
		if recorder.status == 0 {
			log.Printf("404 - Route not found: %s %s", r.Method, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Route not found"})
		}
	})

	// Start server
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Server running on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	if sr.status == 0 {
		sr.status = code
	}
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	// If status hasn't been set yet, default to 200
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	return sr.ResponseWriter.Write(b)
}

func errorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Error: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "Internal server error",
					"details": fmt.Sprintf("%v", err),
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
