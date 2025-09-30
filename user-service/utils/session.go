package utils

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
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



// UserSession represents a multi-device session
type UserSession struct {
	UserID    int       `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// GenerateDeviceID creates a unique device identifier
func GenerateDeviceID(ip, userAgent string) string {
	data := fmt.Sprintf("%s-%s-%d", ip, userAgent, time.Now().UnixNano())
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// StoreMultiDeviceSession stores session with device tracking
func StoreMultiDeviceSession(userID int, deviceID, token string) error {
	session := UserSession{
		UserID:    userID,
		DeviceID:  deviceID,
		Token:     token,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
	
	// Store in Redis with 30-day expiration
	sessionKey := fmt.Sprintf("session:%d:%s", userID, deviceID)
	sessionData, _ := json.Marshal(session)
	
	err := RedisClient.Set(context.Background(), sessionKey, sessionData, 30*24*time.Hour).Err()
	if err != nil {
		return err
	}
	
	// Also store in database for persistence
	return StoreSession(userID, token)
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