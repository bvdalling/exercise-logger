package middleware

import (
	"context"
	"net/http"
	"gym-app-backend/utils"
)

type contextKey string

const userIDKey contextKey = "userID"
const usernameKey contextKey = "username"

// RequireAuth middleware verifies that the user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := utils.GetSessionFromRequest(r)
		if !ok {
			http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		session, ok := utils.GetSession(sessionID)
		if !ok {
			http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), userIDKey, session.UserID)
		ctx = context.WithValue(ctx, usernameKey, session.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) int64 {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		return 0
	}
	return userID
}

// GetUsername extracts username from request context
func GetUsername(r *http.Request) string {
	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		return ""
	}
	return username
}

