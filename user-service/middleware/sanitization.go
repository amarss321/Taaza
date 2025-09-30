package middleware

import (
	"html"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func SanitizeInputAdvanced() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = sanitizeString(value)
			}
			c.Request.URL.Query()[key] = values
		}

		// Sanitize form data
		if c.Request.Form != nil {
			for key, values := range c.Request.Form {
				for i, value := range values {
					values[i] = sanitizeString(value)
				}
				c.Request.Form[key] = values
			}
		}

		c.Next()
	}
}

func sanitizeString(input string) string {
	// Remove potential XSS
	sanitized := html.EscapeString(input)
	// Remove potential SQL injection characters
	sanitized = strings.ReplaceAll(sanitized, "'", "")
	sanitized = strings.ReplaceAll(sanitized, "\"", "")
	sanitized = strings.ReplaceAll(sanitized, ";", "")
	sanitized = strings.ReplaceAll(sanitized, "--", "")
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	return sanitized
}

func SanitizeStructData(data interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return
	}
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			field.SetString(sanitizeString(field.String()))
		}
	}
}