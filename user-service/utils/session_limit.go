package utils

import (
	"user-service/database"
)

// StoreSessionWithLimit stores session and enforces device limit
func StoreSessionWithLimit(userID int, token string, maxSessions int) error {
	// Count current active sessions
	var sessionCount int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM user_sessions 
		WHERE user_id = $1 AND expires_at > NOW()`,
		userID).Scan(&sessionCount)
	
	if err != nil {
		return err
	}
	
	// If at limit, remove oldest session
	if sessionCount >= maxSessions {
		_, err = database.DB.Exec(`
			DELETE FROM user_sessions 
			WHERE user_id = $1 AND id IN (
				SELECT id FROM user_sessions 
				WHERE user_id = $1 AND expires_at > NOW()
				ORDER BY created_at ASC 
				LIMIT 1
			)`, userID)
		
		if err != nil {
			return err
		}
	}
	
	// Store new session
	return StoreSession(userID, token)
}

// GetActiveSessionCount returns number of active sessions for user
func GetActiveSessionCount(userID int) (int, error) {
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM user_sessions 
		WHERE user_id = $1 AND expires_at > NOW()`,
		userID).Scan(&count)
	
	return count, err
}