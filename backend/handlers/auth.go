package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"gym-app-backend/database"
	"gym-app-backend/middleware"
	"gym-app-backend/models"
	"gym-app-backend/utils"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResetPasswordRequest struct {
	RecoveryUUID   string `json:"recoveryUuid"`
	RecoverySecret string `json:"recoverySecret"`
	NewPassword    string `json:"newPassword"`
}

type RegisterResponse struct {
	Message  string      `json:"message"`
	User     models.User `json:"user"`
	Recovery *struct {
		UUID   string `json:"uuid"`
		Secret string `json:"secret"`
	} `json:"recovery,omitempty"`
}

type LoginResponse struct {
	Message string      `json:"message"`
	User    models.User `json:"user"`
}

type UserResponse struct {
	User models.User `json:"user"`
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"Username and password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingID int64
	err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err == nil {
		http.Error(w, `{"error":"Username already exists"}`, http.StatusBadRequest)
		return
	} else if err != sql.ErrNoRows {
		fmt.Printf("Registration error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate recovery credentials
	recoveryUUID, err := generateUUID()
	if err != nil {
		fmt.Printf("Failed to generate recovery UUID: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	recoverySecret, err := utils.GenerateRecoverySecret()
	if err != nil {
		fmt.Printf("Failed to generate recovery secret: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	recoverySecretHash, err := utils.HashRecoverySecret(recoverySecret)
	if err != nil {
		fmt.Printf("Failed to hash recovery secret: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Try to insert with recovery columns, fallback if they don't exist
	var result sql.Result
	result, err = database.DB.Exec(
		"INSERT INTO users (username, password_hash, recovery_uuid, recovery_secret_hash) VALUES (?, ?, ?, ?)",
		req.Username, passwordHash, recoveryUUID, recoverySecretHash,
	)
	if err != nil {
		// If recovery columns don't exist, insert without them
		if err.Error() != "" {
			result, err = database.DB.Exec(
				"INSERT INTO users (username, password_hash) VALUES (?, ?)",
				req.Username, passwordHash,
			)
			if err != nil {
				fmt.Printf("Registration error: %v\n", err)
				http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
				return
			}
			recoveryUUID = ""
			recoverySecret = ""
		} else {
			fmt.Printf("Registration error: %v\n", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	userID, _ := result.LastInsertId()

	// Create session
	sessionID, err := utils.StartSession(userID, req.Username)
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	isProduction := os.Getenv("NODE_ENV") == "production"
	utils.SetSessionCookie(w, sessionID, isProduction)

	// Prepare response
	user := models.User{
		ID:       userID,
		Username: req.Username,
	}

	response := RegisterResponse{
		Message: "User registered successfully",
		User:    user,
	}

	if recoveryUUID != "" && recoverySecret != "" {
		response.Recovery = &struct {
			UUID   string `json:"uuid"`
			Secret string `json:"secret"`
		}{
			UUID:   recoveryUUID,
			Secret: recoverySecret,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user login
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"Username and password are required"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Printf("Login error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Verify password
	if !utils.ComparePassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID, err := utils.StartSession(user.ID, user.Username)
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	isProduction := os.Getenv("NODE_ENV") == "production"
	utils.SetSessionCookie(w, sessionID, isProduction)

	response := LoginResponse{
		Message: "Login successful",
		User: models.User{
			ID:       user.ID,
			Username: user.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	sessionID, ok := utils.GetSessionFromRequest(r)
	if ok {
		utils.DeleteSession(sessionID)
	}

	utils.ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, `{"error":"Not authenticated"}`, http.StatusUnauthorized)
		return
	}

	var user models.User
	var createdAtStr string
	err := database.DB.QueryRow(
		"SELECT id, username, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &createdAtStr)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Get current user error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Parse created_at
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

	response := UserResponse{User: user}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPassword handles password reset using recovery credentials
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.RecoveryUUID == "" || req.RecoverySecret == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"Recovery UUID, recovery secret, and new password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Find user by recovery UUID
	var user models.User
	var recoverySecretHash sql.NullString
	err := database.DB.QueryRow(
		"SELECT id, username, recovery_uuid, recovery_secret_hash FROM users WHERE recovery_uuid = ?",
		req.RecoveryUUID,
	).Scan(&user.ID, &user.Username, &user.RecoveryUUID, &recoverySecretHash)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Invalid recovery credentials"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Printf("Reset password error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Verify recovery secret
	if !recoverySecretHash.Valid || !utils.CompareRecoverySecret(req.RecoverySecret, recoverySecretHash.String) {
		http.Error(w, `{"error":"Invalid recovery credentials"}`, http.StatusUnauthorized)
		return
	}

	// Hash new password and update
	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, user.ID)
	if err != nil {
		fmt.Printf("Failed to update password: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Message: "Password reset successfully",
		User: models.User{
			ID:       user.ID,
			Username: user.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateUUID generates a UUID v4 using crypto/rand
func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10

	// Format as UUID string: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
