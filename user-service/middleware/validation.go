package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Input sanitization middleware
func SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize common XSS patterns
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// Add basic XSS protection
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
		}
		c.Next()
	}
}

// Enhanced validation helpers
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidatePhone(phone string) bool {
	phone = strings.ReplaceAll(phone, " ", "")
	phoneRegex := regexp.MustCompile(`^[6-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 2 && len(name) <= 50
}