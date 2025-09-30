package handlers

import (
	"admin-service/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Activity struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
}

type Session struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func GetUserSessions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	query := `
		SELECT id, user_id, token, expires_at, created_at 
		FROM user_sessions 
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		err := rows.Scan(&session.ID, &session.UserID, &session.Token, 
			&session.ExpiresAt, &session.CreatedAt)
		if err != nil {
			continue
		}
		// Hide full token for security
		if len(session.Token) > 10 {
			session.Token = session.Token[:10] + "..."
		}
		sessions = append(sessions, session)
	}

	c.JSON(http.StatusOK, sessions)
}

func GetRecentActivity(c *gin.Context) {
	query := `
		SELECT ua.id, ua.user_id, ua.action, ua.details, ua.ip_address, ua.created_at, u.name
		FROM user_activity ua
		JOIN users u ON ua.user_id = u.id
		ORDER BY ua.created_at DESC 
		LIMIT 100`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var activity Activity
		var userName string
		err := rows.Scan(&activity.ID, &activity.UserID, &activity.Action, 
			&activity.Details, &activity.IPAddress, &activity.CreatedAt, &userName)
		if err != nil {
			continue
		}
		
		activityMap := map[string]interface{}{
			"id":         activity.ID,
			"user_id":    activity.UserID,
			"user_name":  userName,
			"action":     activity.Action,
			"details":    activity.Details,
			"ip_address": activity.IPAddress,
			"created_at": activity.CreatedAt,
		}
		activities = append(activities, activityMap)
	}

	c.JSON(http.StatusOK, activities)
}

func RevokeSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	query := `DELETE FROM user_sessions WHERE id = $1`
	_, err = database.DB.Exec(query, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}