package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"time"
)

// SessionTTL controls how long a session stays valid after login.
const SessionTTL = 24 * time.Hour

// DB must be set (from main, the same *sql.DB used by the repositories)
// before any of the functions in this file are called.
var DB *sql.DB

// CreateTable creates the sessions table if it doesn't already exist.
// Sessions are persisted so they survive server restarts, unlike an
// in-memory map which would silently log out every user on redeploy.
func CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		expires_at DATETIME NOT NULL
	)`
	_, err := DB.Exec(query)
	return err
}

func GenerateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession generates a new session ID for the given email, valid for
// SessionTTL, and returns the ID to be set as the session cookie value.
func CreateSession(email string) string {
	sessionID := GenerateSessionID()
	expiresAt := time.Now().Add(SessionTTL)

	_, err := DB.Exec(
		`INSERT INTO sessions (id, email, expires_at) VALUES (?, ?, ?)`,
		sessionID, email, expiresAt,
	)
	if err != nil {
		log.Println("auth: failed to persist session:", err)
	}
	return sessionID
}

// GetSessionEmail returns the email tied to a session ID and whether it's
// still valid. An expired session is deleted on lookup rather than kept
// around indefinitely.
func GetSessionEmail(sessionID string) (string, bool) {
	var email string
	var expiresAt time.Time

	err := DB.QueryRow(
		`SELECT email, expires_at FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&email, &expiresAt)

	if err != nil {
		return "", false
	}

	if time.Now().After(expiresAt) {
		DeleteSession(sessionID)
		return "", false
	}

	return email, true
}

// DeleteSession removes a session immediately — used on logout and account
// deletion.
func DeleteSession(sessionID string) {
	DB.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
}

// CleanupExpiredSessions deletes every session whose expiry has passed.
// Without this, sessions that are never presented again (e.g. a user logs
// in once and never returns) would accumulate in the table forever, since
// GetSessionEmail only reaps a session when it's actually looked up.
func CleanupExpiredSessions() error {
	_, err := DB.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	return err
}

// StartCleanupLoop runs CleanupExpiredSessions immediately and then on the
// given interval, for as long as the process runs. Intended to be started
// once from main via `go auth.StartCleanupLoop(time.Hour)`.
func StartCleanupLoop(interval time.Duration) {
	if err := CleanupExpiredSessions(); err != nil {
		log.Println("auth: session cleanup failed:", err)
	}

	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := CleanupExpiredSessions(); err != nil {
			log.Println("auth: session cleanup failed:", err)
		}
	}
}
