package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionTTL controls how long a session stays valid after login.
const SessionTTL = 24 * time.Hour

type sessionEntry struct {
	Email     string
	ExpiresAt time.Time
}

var (
	mu       sync.Mutex
	sessions = map[string]sessionEntry{}
)

func GenerateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession generates a new session ID for the given email, valid for
// SessionTTL, and returns the ID to be set as the session cookie value.
func CreateSession(email string) string {
	sessionID := GenerateSessionID()
	mu.Lock()
	sessions[sessionID] = sessionEntry{
		Email:     email,
		ExpiresAt: time.Now().Add(SessionTTL),
	}
	mu.Unlock()
	return sessionID
}

// GetSessionEmail returns the email tied to a session ID and whether it's
// still valid. An expired session is deleted on lookup rather than kept
// around indefinitely.
func GetSessionEmail(sessionID string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	entry, exists := sessions[sessionID]
	if !exists {
		return "", false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(sessions, sessionID)
		return "", false
	}
	return entry.Email, true
}

// DeleteSession removes a session immediately — used on logout and account
// deletion.
func DeleteSession(sessionID string) {
	mu.Lock()
	delete(sessions, sessionID)
	mu.Unlock()
}
