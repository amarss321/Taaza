package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"user-service/database"
	"user-service/models"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CompleteProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logrus.Error("Password hashing error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update user with password and complete status
	_, err = database.DB.Exec("UPDATE users SET password_hash = $1, registration_status = 'completed', updated_at = NOW() WHERE id = $2",
		hashedPassword, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(userID, "profile_complete", "User completed profile setup", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Profile completed successfully", "redirect": "homepage"})
}

func UpdatePersonalInfo(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.PersonalInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if mobile number is already used by another user
	if req.Mobile != "" {
		var existingUserID int
		err := database.DB.QueryRow("SELECT id FROM users WHERE mobile = $1 AND id != $2", req.Mobile, userID).Scan(&existingUserID)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Mobile number is already registered with another account"})
			return
		} else if err != sql.ErrNoRows {
			logrus.Error("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	// Update personal information
	_, err := database.DB.Exec("UPDATE users SET mobile = $1, address = $2, updated_at = NOW() WHERE id = $3",
		req.Mobile, req.Address, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Mobile number is already registered with another account"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Log activity
	utils.LogActivity(userID, "personal_info_update", "User updated mobile/address", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Personal information updated successfully"})
}

func GetProfile(c *gin.Context) {
	userID := c.GetInt("user_id")
	
	logrus.Infof("GetProfile called for user ID: %d", userID)

	var user struct {
		ID                 int    `json:"id"`
		Name               string `json:"name"`
		Email              string `json:"email"`
		Mobile             string `json:"mobile"`
		Address            string `json:"address"`
		Gender             string `json:"gender"`
		DateOfBirth        string `json:"date_of_birth"`
		RegistrationStatus string `json:"registration_status"`
		CreatedAt          string `json:"created_at"`
	}

	err := database.DB.QueryRow("SELECT id, name, email, COALESCE(mobile, '') as mobile, COALESCE(address, '') as address, gender, date_of_birth, registration_status, created_at FROM users WHERE id = $1",
		userID).Scan(&user.ID, &user.Name, &user.Email, &user.Mobile, &user.Address, &user.Gender, &user.DateOfBirth, &user.RegistrationStatus, &user.CreatedAt)

	if err == sql.ErrNoRows {
		logrus.Warnf("User not found with ID: %d", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		logrus.Errorf("Database error in GetProfile for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	logrus.Infof("Successfully retrieved profile for user: %s (ID: %d)", user.Name, user.ID)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec("UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3",
		req.Name, req.Email, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(userID, "profile_update", "User updated profile information", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func CompleteUserProfile(c *gin.Context) {
	type CompleteUserProfileRequest struct {
		Email  string `json:"email" validate:"required,email"`
		Name   string `json:"name" validate:"required,min=2"`
		Mobile string `json:"mobile" validate:"required,len=10"`
	}

	var req CompleteUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if mobile number is already used by another user
	var existingUserID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE mobile = $1 AND email != $2", req.Mobile, req.Email).Scan(&existingUserID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Mobile number is already registered with another account"})
		return
	} else if err != sql.ErrNoRows {
		logrus.Error("CompleteUserProfile - Database error during mobile check:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update user with name, mobile and mark profile as completed
	_, err = database.DB.Exec("UPDATE users SET name = $1, mobile = $2, profile_completed = true, registration_status = 'completed', updated_at = NOW() WHERE email = $3",
		req.Name, req.Mobile, req.Email)
	if err != nil {
		logrus.Error("CompleteUserProfile - Database error during update:", err)
		// Enhanced error detection for constraint violations
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "unique_mobile") ||
			strings.Contains(err.Error(), "mobile_unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "Mobile number is already registered with another account"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get user ID for logging
	var userID int
	database.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	utils.LogActivity(userID, "profile_complete", "User completed profile setup", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Profile completed successfully"})
}

func UpdateUserProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	type UpdateUserProfileRequest struct {
		Name        string `json:"name" validate:"required,min=2"`
		Gender      string `json:"gender"`
		DateOfBirth string `json:"date_of_birth"`
	}

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug logging
	logrus.Infof("Updating user %d with: name=%s, gender=%s, dob=%s", userID, req.Name, req.Gender, req.DateOfBirth)

	// Handle null values properly for optional fields
	var dateOfBirth interface{}
	if req.DateOfBirth == "" {
		dateOfBirth = nil
	} else {
		dateOfBirth = req.DateOfBirth
	}
	
	var gender interface{}
	if req.Gender == "" {
		gender = nil
	} else {
		gender = req.Gender
	}
	
	logrus.Infof("Updating user %d: name=%s, gender=%v, dob=%v", userID, req.Name, gender, dateOfBirth)
	
	_, err := database.DB.Exec("UPDATE users SET name = $1, gender = $2, date_of_birth = $3, updated_at = NOW() WHERE id = $4",
		req.Name, gender, dateOfBirth, userID)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(userID, "personal_info_update", "User updated personal information", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}