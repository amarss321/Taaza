package utils

import (
	"user-service/database"
	"github.com/sirupsen/logrus"
)

// LogActivity logs user activity to the database
func LogActivity(userID int, action, details, ipAddress string) {
	_, err := database.DB.Exec(`
		INSERT INTO user_activity (user_id, action, details, ip_address) 
		VALUES ($1, $2, $3, $4)`,
		userID, action, details, ipAddress)
	
	if err != nil {
		logrus.Error("Failed to log activity:", err)
	}
}

// CleanupOldActivity removes activity logs older than 7 days
func CleanupOldActivity() error {
	_, err := database.DB.Exec(`
		DELETE FROM user_activity 
		WHERE created_at < NOW() - INTERVAL '7 days'`)
	
	if err != nil {
		logrus.Error("Failed to cleanup old activity:", err)
	}
	
	return err
}