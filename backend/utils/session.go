package utils

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "connect.sid"
	sessionDuration  = 24 * time.Hour
)

type Session struct {
	UserID   int64
	Username string
	Expires  time.Time
}

type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var store = &SessionStore{
	sessions: make(map[string]*Session),
}

// StartSession creates a new session and returns the session ID
func StartSession(userID int64, username string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	session := &Session{
		UserID:   userID,
		Username: username,
		Expires:  time.Now().Add(sessionDuration),
	}

	store.mu.Lock()
	store.sessions[sessionID] = session
	store.mu.Unlock()

	// Clean up expired sessions periodically
	go cleanupExpiredSessions()

	return sessionID, nil
}

// GetSession retrieves a session by ID
func GetSession(sessionID string) (*Session, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	session, exists := store.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.Expires) {
		delete(store.sessions, sessionID)
		return nil, false
	}

	return session, true
}

// DeleteSession removes a session
func DeleteSession(sessionID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, sessionID)
}

// SetSessionCookie sets the session cookie on the response
func SetSessionCookie(w http.ResponseWriter, sessionID string, isProduction bool) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		MaxAge:   int(sessionDuration.Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest retrieves the session ID from the request cookie
func GetSessionFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func cleanupExpiredSessions() {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	for id, session := range store.sessions {
		if now.After(session.Expires) {
			delete(store.sessions, id)
		}
	}
}

