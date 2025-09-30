package handlers

import (
	"admin-service/database"
	"admin-service/models"
	"admin-service/utils"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")

	offset := (page - 1) * limit

	query := `SELECT id, name, email, mobile, gender, date_of_birth, registration_status, 
			  is_verified, profile_completed, status, last_login, created_at, updated_at 
			  FROM users WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += " AND (name ILIKE $" + strconv.Itoa(argCount) + " OR email ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		argCount++
		query += " AND status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argCount+1) + " OFFSET $" + strconv.Itoa(argCount+2)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Gender,
			&user.DateOfBirth, &user.RegistrationStatus, &user.IsVerified, &user.ProfileCompleted,
			&user.Status, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			logrus.Error("Row scan error:", err)
			continue
		}
		users = append(users, user)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users WHERE 1=1"
	countArgs := []interface{}{}
	countArgCount := 0

	if search != "" {
		countArgCount++
		countQuery += " AND (name ILIKE $" + strconv.Itoa(countArgCount) + " OR email ILIKE $" + strconv.Itoa(countArgCount) + ")"
		countArgs = append(countArgs, "%"+search+"%")
	}

	if status != "" {
		countArgCount++
		countQuery += " AND status = $" + strconv.Itoa(countArgCount)
		countArgs = append(countArgs, status)
	}

	var total int
	database.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = database.DB.QueryRow(`
		SELECT id, name, email, mobile, gender, date_of_birth, registration_status, 
		       is_verified, profile_completed, status, last_login, created_at, updated_at 
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Gender, &user.DateOfBirth,
		&user.RegistrationStatus, &user.IsVerified, &user.ProfileCompleted,
		&user.Status, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 0

	if req.Status != nil {
		argCount++
		setParts = append(setParts, "status = $"+strconv.Itoa(argCount))
		args = append(args, *req.Status)
	}

	if req.IsVerified != nil {
		argCount++
		setParts = append(setParts, "is_verified = $"+strconv.Itoa(argCount))
		args = append(args, *req.IsVerified)
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	argCount++
	setParts = append(setParts, "updated_at = $"+strconv.Itoa(argCount))
	args = append(args, time.Now())

	argCount++
	args = append(args, userID)

	query := "UPDATE users SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argCount)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogUserActivity(userID, "admin_update", req.Reason, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func GetUserStats(c *gin.Context) {
	var stats models.UserStats

	// Total users
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Active users
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&stats.ActiveUsers)

	// Blocked users
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'blocked'").Scan(&stats.BlockedUsers)

	// Verified users
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_verified = true").Scan(&stats.VerifiedUsers)

	// New users today
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE").Scan(&stats.NewUsersToday)

	// Completed profiles
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE profile_completed = true").Scan(&stats.CompletedProfiles)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func GetUserActivity(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, user_id, action, details, ip_address, created_at 
		FROM user_activity WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var activities []models.UserActivity
	for rows.Next() {
		var activity models.UserActivity
		err := rows.Scan(&activity.ID, &activity.UserID, &activity.Action,
			&activity.Details, &activity.IPAddress, &activity.CreatedAt)
		if err != nil {
			logrus.Error("Row scan error:", err)
			continue
		}
		activities = append(activities, activity)
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}