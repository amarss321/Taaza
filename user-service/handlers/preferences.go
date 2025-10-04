package handlers

import (
	"user-service/database"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserPreference struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetUserPreferences(c *gin.Context) {
	userID := c.GetInt("user_id")

	query := `SELECT preference_key, preference_value FROM user_preferences WHERE user_id = $1`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	preferences := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		preferences[key] = value
	}

	utils.LogActivity(userID, "preferences_view", "User viewed preferences", c.ClientIP())
	c.JSON(200, preferences)
}

func SetUserPreference(c *gin.Context) {
	userID := c.GetInt("user_id")

	var pref UserPreference
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO user_preferences (user_id, preference_key, preference_value)
			  VALUES ($1, $2, $3)
			  ON CONFLICT (user_id, preference_key)
			  DO UPDATE SET preference_value = $3, updated_at = CURRENT_TIMESTAMP`

	_, err := database.DB.Exec(query, userID, pref.Key, pref.Value)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(500, gin.H{"error": "Failed to save preference"})
		return
	}

	utils.LogActivity(userID, "preference_set", "User set preference: "+pref.Key, c.ClientIP())
	c.JSON(200, gin.H{"message": "Preference saved successfully"})
}

func SetUserPreferences(c *gin.Context) {
	userID := c.GetInt("user_id")

	var preferences map[string]string
	if err := c.ShouldBindJSON(&preferences); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer tx.Rollback()

	for key, value := range preferences {
		query := `INSERT INTO user_preferences (user_id, preference_key, preference_value)
				  VALUES ($1, $2, $3)
				  ON CONFLICT (user_id, preference_key)
				  DO UPDATE SET preference_value = $3, updated_at = CURRENT_TIMESTAMP`

		_, err := tx.Exec(query, userID, key, value)
		if err != nil {
			logrus.Error("Database error:", err)
			c.JSON(500, gin.H{"error": "Failed to save preferences"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		logrus.Error("Database error:", err)
		c.JSON(500, gin.H{"error": "Failed to save preferences"})
		return
	}

	utils.LogActivity(userID, "preferences_bulk_set", "User updated multiple preferences", c.ClientIP())
	c.JSON(200, gin.H{"message": "Preferences saved successfully"})
}