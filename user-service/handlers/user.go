package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"user-service/database"
	"user-service/models"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var validate = validator.New()

func StartRegistration(c *gin.Context) {
	var req models.StartRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Insert user with pending status
	_, err = database.DB.Exec("INSERT INTO users (name, email, registration_status) VALUES ($1, $2, 'pending')",
		req.Name, req.Email)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Generate and store OTP
	otp := utils.GenerateOTP()
	if err := utils.StoreOTP(req.Email, otp); err != nil {
		logrus.Error("OTP storage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Send OTP via email
	if err := utils.SendOTPEmail(req.Email, req.Name, otp); err != nil {
		logrus.Error("Email sending error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration started. OTP sent to email"})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := database.DB.QueryRow("SELECT id, name, email, password_hash, registration_status, is_verified FROM users WHERE email = $1",
		req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.RegistrationStatus, &user.IsVerified)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if user.Password == "" || !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not verified"})
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		logrus.Error("JWT generation error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Store session in database
	if err := utils.StoreSession(user.ID, token); err != nil {
		logrus.Error("Session storage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(user.ID, "login", "User logged in successfully", c.ClientIP())

	response := gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
			"registration_status": user.RegistrationStatus,
		},
	}

	if user.RegistrationStatus == "verified" {
		response["redirect"] = "complete-profile"
	} else {
		response["redirect"] = "homepage"
	}

	c.JSON(http.StatusOK, response)
}

func LoginWithOTP(c *gin.Context) {
	var req models.LoginWithOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists, if not create new user
	var userExists bool
	var userName string
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1), COALESCE((SELECT name FROM users WHERE email = $1), '')", req.Email).Scan(&userExists, &userName)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !userExists {
		// Create new user
		defaultName := req.Email[:strings.Index(req.Email, "@")]
		_, err = database.DB.Exec("INSERT INTO users (name, email, registration_status, is_verified) VALUES ($1, $2, 'pending', false)",
			defaultName, req.Email)
		if err != nil {
			logrus.Error("Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		userName = defaultName
	}

	// Generate and store OTP
	otp := utils.GenerateOTP()
	if err := utils.StoreOTP(req.Email, otp); err != nil {
		logrus.Error("OTP storage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Send OTP via email
	if err := utils.SendOTPEmail(req.Email, userName, otp); err != nil {
		logrus.Error("Email sending error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to email"})
}

func LoginVerifyOTP(c *gin.Context) {
	var req models.LoginVerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.VerifyOTP(req.Email, req.OTP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// Get user details
	var user models.User
	var mobile, address sql.NullString
	err := database.DB.QueryRow("SELECT id, name, email, registration_status, is_verified, mobile, address, profile_completed FROM users WHERE email = $1",
		req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.RegistrationStatus, &user.IsVerified, &mobile, &address, &user.ProfileCompleted)

	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update verification status
	_, err = database.DB.Exec("UPDATE users SET is_verified = true WHERE email = $1", req.Email)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		logrus.Error("JWT generation error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Store session in database
	if err := utils.StoreSession(user.ID, token); err != nil {
		logrus.Error("Session storage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log activity
	utils.LogActivity(user.ID, "login_otp", "User logged in via OTP", c.ClientIP())

	// Check if user is new (profile not completed)
	isNewUser := !user.ProfileCompleted || !mobile.Valid || mobile.String == ""

	response := gin.H{
		"token": token,
		"isNewUser": isNewUser,
	}

	if !isNewUser {
		response["user"] = gin.H{
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
			"mobile": mobile.String,
		}
	}

	c.JSON(http.StatusOK, response)
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.VerifyOTP(req.Email, req.OTP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// Update user verification status
	_, err := database.DB.Exec("UPDATE users SET is_verified = true, registration_status = 'verified' WHERE email = $1", req.Email)
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully", "redirect": "complete-profile"})
}

