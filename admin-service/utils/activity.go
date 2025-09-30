package utils

import (
	"admin-service/database"

	"github.com/sirupsen/logrus"
)

func LogUserActivity(userID int, action, details, ipAddress string) {
	_, err := database.DB.Exec(
		"INSERT INTO user_activity (user_id, action, details, ip_address) VALUES ($1, $2, $3, $4)",
		userID, action, details, ipAddress)

	if err != nil {
		logrus.Error("Failed to log user activity:", err)
	}
}