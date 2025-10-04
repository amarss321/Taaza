package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ProxyToUserService(c *gin.Context) {
	proxyRequest(c, os.Getenv("USER_SERVICE_URL"))
}

func ProxyToAdminService(c *gin.Context) {
	proxyRequest(c, os.Getenv("ADMIN_SERVICE_URL"))
}

func ProxyToInventoryService(c *gin.Context) {
	proxyRequest(c, os.Getenv("INVENTORY_SERVICE_URL"))
}

func proxyRequest(c *gin.Context, targetURL string) {
	// Build target URL
	url := targetURL + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		url += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create new request
	req, err := http.NewRequest(c.Request.Method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logrus.Error("Failed to create request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user context from middleware
	if userID, exists := c.Get("user_id"); exists {
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID.(int)))
	}
	if role, exists := c.Get("role"); exists {
		req.Header.Set("X-User-Role", role.(string))
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("Failed to proxy request:", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

func RouteBasedOnJWT(c *gin.Context) {
	path := c.Request.URL.Path
	
	// Admin routes
	if strings.HasPrefix(path, "/api/v1/admin/") {
		ProxyToAdminService(c)
		return
	}
	
	// User routes
	if strings.HasPrefix(path, "/api/v1/users/") {
		ProxyToUserService(c)
		return
	}
	
	// Inventory routes
	if strings.HasPrefix(path, "/api/v1/inventory/") {
		ProxyToInventoryService(c)
		return
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
}