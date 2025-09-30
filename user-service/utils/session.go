package utils

import (
	"time"
	"user-service/database"
)

// StoreSession stores a session in the database with 7-day expiry
func StoreSession(userID int, token string) error {
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour) // 7 days
	
	_, err := database.DB.Exec(`
		INSERT INTO user_sessions (user_id, token, expires_at) 
		VALUES ($1, $2, $3)`,
		userID, token, expiresAt)
	
	return err
}

// ValidateSession checks if session exists and is not expired (no renewal)
func ValidateSession(token string) (int, error) {
	var userID int
	
	// Check if session exists and not expired
	err := database.DB.QueryRow(`
		SELECT user_id FROM user_sessions 
		WHERE token = $1 AND expires_at > NOW()`,
		token).Scan(&userID)
	
	return userID, err
}

// UpdateSessionActivity extends session expiry (call only on specific actions)
func UpdateSessionActivity(token string) error {
	newExpiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	_, err := database.DB.Exec(`
		UPDATE user_sessions 
		SET expires_at = $1 
		WHERE token = $2`,
		newExpiresAt, token)
	
	return err
}

// RevokeSession removes a session from database (logout)
func RevokeSession(token string) error {
	_, err := database.DB.Exec(`
		DELETE FROM user_sessions WHERE token = $1`,
		token)
	
	return err
}

// RevokeAllUserSessions removes all sessions for a user
func RevokeAllUserSessions(userID int) error {
	_, err := database.DB.Exec(`
		DELETE FROM user_sessions WHERE user_id = $1`,
		userID)
	
	return err
}

// ValidateAndRenewSession validates session and extends expiry
func ValidateAndRenewSession(token string) (int, error) {
	userID, err := ValidateSession(token)
	if err != nil {
		return 0, err
	}
	
	// Extend session expiry
	err = UpdateSessionActivity(token)
	if err != nil {
		return 0, err
	}
	
	return userID, nil
}

// CleanupExpiredSessions removes expired sessions (run periodically)
func CleanupExpiredSessions() error {
	_, err := database.DB.Exec(`
		DELETE FROM user_sessions WHERE expires_at < NOW()`)
	
	return err
}



// StartSessionCleanup runs session cleanup every 12 hours
func StartSessionCleanup() {
	// Session cleanup every 12 hours
	sessionTicker := time.NewTicker(12 * time.Hour)
	go func() {
		for range sessionTicker.C {
			if err := CleanupExpiredSessions(); err != nil {
				// Log error but don't stop the cleanup process
				time.Sleep(10 * time.Second) // Wait before next attempt
			}
		}
	}()
	
	// Activity cleanup daily at midnight
	activityTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for range activityTicker.C {
			CleanupOldActivity() // Clean activity logs older than 7 days
		}
	}()
}