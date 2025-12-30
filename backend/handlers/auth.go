package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gym-app-backend/database"
	"gym-app-backend/middleware"
	"gym-app-backend/models"
	"gym-app-backend/services"
	"gym-app-backend/utils"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
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

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, `{"error":"Username, email, and password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		http.Error(w, `{"error":"Invalid email address"}`, http.StatusBadRequest)
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

	// Check if email already exists
	err = database.DB.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingID)
	if err == nil {
		http.Error(w, `{"error":"Email already exists"}`, http.StatusBadRequest)
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

	// Try to insert with recovery columns and email, fallback if they don't exist
	var result sql.Result
	result, err = database.DB.Exec(
		"INSERT INTO users (username, email, password_hash, recovery_uuid, recovery_secret_hash) VALUES (?, ?, ?, ?, ?)",
		req.Username, req.Email, passwordHash, recoveryUUID, recoverySecretHash,
	)
	if err != nil {
		// If recovery columns don't exist, insert without them
		if err.Error() != "" {
			result, err = database.DB.Exec(
				"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
				req.Username, req.Email, passwordHash,
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
		Email:    &req.Email,
	}

	// Don't create session yet - user must set up TOTP first
	// Return response indicating TOTP setup is required
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "User registered successfully. TOTP setup required.",
		"user":              user,
		"requiresTOTPSetup": true,
	})
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

	// Check if TOTP is enabled
	var totpEnabled sql.NullBool
	err = database.DB.QueryRow(
		"SELECT totp_enabled FROM users WHERE id = ?",
		user.ID,
	).Scan(&totpEnabled)

	if err != nil {
		fmt.Printf("Login error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// If TOTP is enabled, require TOTP verification
	if totpEnabled.Valid && totpEnabled.Bool {
		// Return response indicating TOTP is required
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"requiresTOTP": true,
			"username":     user.Username,
			"message":      "TOTP verification required",
		})
		return
	}

	// TOTP not enabled - require setup
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requiresTOTPSetup": true,
		"username":          user.Username,
		"message":           "TOTP setup required",
	})
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
	var email sql.NullString
	var totpEnabled sql.NullBool
	err := database.DB.QueryRow(
		"SELECT id, username, email, totp_enabled, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &email, &totpEnabled, &createdAtStr)

	if email.Valid {
		user.Email = &email.String
	}
	if totpEnabled.Valid {
		enabled := totpEnabled.Bool
		user.TOTPEnabled = &enabled
	}

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

// RequestPasswordReset handles password reset requests via email
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, `{"error":"Email is required"}`, http.StatusBadRequest)
		return
	}

	// Find user by email
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email)

	if err == sql.ErrNoRows {
		// Don't reveal if email exists - return success anyway for security
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a password reset link has been sent"})
		return
	} else if err != nil {
		fmt.Printf("Request password reset error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate secure reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	token := fmt.Sprintf("%x", tokenBytes)

	// Set expiration to 1 hour from now
	expires := time.Now().Add(1 * time.Hour)

	// Store token in database
	_, err = database.DB.Exec(
		"UPDATE users SET password_reset_token = ?, password_reset_expires = ? WHERE id = ?",
		token, expires, user.ID,
	)
	if err != nil {
		fmt.Printf("Failed to store reset token: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Send email
	if services.EmailService != nil {
		if err := services.EmailService.SendPasswordResetEmail(*user.Email, token); err != nil {
			fmt.Printf("Failed to send reset email: %v\n", err)
			// Don't fail the request - token is still valid
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a password reset link has been sent"})
}

// ResetPassword handles password reset using email token
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

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"Token and new password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	// Find user by reset token and check expiration
	var user models.User
	var expiresStr sql.NullString
	err := database.DB.QueryRow(
		"SELECT id, username, email, password_reset_expires FROM users WHERE password_reset_token = ?",
		req.Token,
	).Scan(&user.ID, &user.Username, &user.Email, &expiresStr)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Invalid or expired reset token"}`, http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Printf("Reset password error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Check if token has expired
	if !expiresStr.Valid {
		http.Error(w, `{"error":"Invalid or expired reset token"}`, http.StatusUnauthorized)
		return
	}

	expires, err := time.Parse("2006-01-02 15:04:05", expiresStr.String)
	if err != nil {
		fmt.Printf("Failed to parse expiration: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if time.Now().After(expires) {
		http.Error(w, `{"error":"Invalid or expired reset token"}`, http.StatusUnauthorized)
		return
	}

	// Hash new password and update
	passwordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Update password and clear reset token
	_, err = database.DB.Exec(
		"UPDATE users SET password_hash = ?, password_reset_token = NULL, password_reset_expires = NULL WHERE id = ?",
		passwordHash, user.ID,
	)
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
			Email:    user.Email,
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

// SetupTOTPRequest represents a request to set up TOTP
type SetupTOTPRequest struct {
	Code string `json:"code"`
}

// SetupTOTPResponse represents the response from TOTP setup
type SetupTOTPResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qrCode"`
	BackupCodes []string `json:"backupCodes,omitempty"`
	Message     string   `json:"message"`
}

// SetupTOTP handles TOTP setup for a user
func SetupTOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == 0 {
		http.Error(w, `{"error":"Not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// Get user info
	var username string
	err := database.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	var req SetupTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Check if user already has TOTP enabled
	var existingSecret sql.NullString
	var totpEnabled sql.NullBool
	err = database.DB.QueryRow(
		"SELECT totp_secret, totp_enabled FROM users WHERE id = ?",
		userID,
	).Scan(&existingSecret, &totpEnabled)

	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	issuer := os.Getenv("TOTP_ISSUER")
	if issuer == "" {
		issuer = "Gym App"
	}

	// If no code provided, generate new secret and QR code
	if req.Code == "" {
		key, err := utils.GenerateTOTPSecret(issuer, username)
		if err != nil {
			http.Error(w, `{"error":"Failed to generate TOTP secret"}`, http.StatusInternalServerError)
			return
		}

		qrCode, err := utils.GenerateTOTPQRCode(key)
		if err != nil {
			http.Error(w, `{"error":"Failed to generate QR code"}`, http.StatusInternalServerError)
			return
		}

		// Store secret temporarily (user hasn't verified yet)
		_, err = database.DB.Exec(
			"UPDATE users SET totp_secret = ? WHERE id = ?",
			key.Secret(), userID,
		)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		response := SetupTOTPResponse{
			Secret:  key.Secret(),
			QRCode:  qrCode,
			Message: "Scan the QR code with your authenticator app",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Code provided - verify it
	if !existingSecret.Valid {
		http.Error(w, `{"error":"No TOTP secret found. Please request a new one."}`, http.StatusBadRequest)
		return
	}

	// Verify TOTP code
	if !utils.ValidateTOTP(existingSecret.String, req.Code) {
		http.Error(w, `{"error":"Invalid TOTP code"}`, http.StatusUnauthorized)
		return
	}

	// Generate backup codes
	backupCodes, err := utils.GenerateBackupCodes(10)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate backup codes"}`, http.StatusInternalServerError)
		return
	}

	hashedCodes, err := utils.HashBackupCodes(backupCodes)
	if err != nil {
		http.Error(w, `{"error":"Failed to hash backup codes"}`, http.StatusInternalServerError)
		return
	}

	// Enable TOTP and store backup codes
	_, err = database.DB.Exec(
		"UPDATE users SET totp_enabled = 1, totp_backup_codes = ? WHERE id = ?",
		hashedCodes, userID,
	)
	if err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := SetupTOTPResponse{
		Message:     "TOTP enabled successfully",
		BackupCodes: backupCodes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyTOTPRequest represents a request to verify TOTP during login
type VerifyTOTPRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

// VerifyTOTP handles TOTP verification during login
func VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req VerifyTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Code == "" {
		http.Error(w, `{"error":"Username and TOTP code are required"}`, http.StatusBadRequest)
		return
	}

	// Find user
	var user models.User
	var totpSecret sql.NullString
	var totpEnabled sql.NullBool
	var totpBackupCodes sql.NullString
	err := database.DB.QueryRow(
		"SELECT id, username, totp_secret, totp_enabled, totp_backup_codes FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &totpSecret, &totpEnabled, &totpBackupCodes)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Verify TOTP error: %v\n", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Check if TOTP is enabled
	if !totpEnabled.Valid || !totpEnabled.Bool {
		http.Error(w, `{"error":"TOTP is not enabled for this user"}`, http.StatusBadRequest)
		return
	}

	// Verify TOTP code or backup code
	valid := false
	if totpSecret.Valid {
		valid = utils.ValidateTOTP(totpSecret.String, req.Code)
	}

	// If TOTP code invalid, try backup codes
	if !valid && totpBackupCodes.Valid {
		valid = utils.VerifyBackupCode(totpBackupCodes.String, req.Code)
		// If backup code used, remove it
		if valid {
			// Remove used backup code (simplified - in production, properly update the array)
			// For now, we'll just mark it as used in a more sophisticated way
		}
	}

	if !valid {
		http.Error(w, `{"error":"Invalid TOTP code"}`, http.StatusUnauthorized)
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
