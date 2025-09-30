package middleware

import (
	"compress/gzip"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Compression middleware
func Compression() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !shouldCompress(c.Request) {
			c.Next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Writer = &gzipWriter{c.Writer, gz}
		c.Next()
	})
}

// Caching middleware
func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cache static assets
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Header("Cache-Control", "public, max-age=86400") // 24 hours
		} else if c.Request.Method == "GET" && strings.Contains(c.Request.URL.Path, "/profile") {
			c.Header("Cache-Control", "private, max-age=300") // 5 minutes
		} else {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		c.Next()
	}
}

// Request timeout middleware
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx := c.Request.Context()
		
		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(timeoutCtx)
		
		done := make(chan struct{})
		go func() {
			c.Next()
			done <- struct{}{}
		}()
		
		select {
		case <-done:
			return
		case <-timeoutCtx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
			c.Abort()
		}
	})
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func shouldCompress(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
}