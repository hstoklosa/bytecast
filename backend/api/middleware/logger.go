package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogFormat represents a structured log entry format
type LogFormat struct {
	RequestID    string        `json:"request_id"`
	ClientIP     string        `json:"client_ip"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	Latency      time.Duration `json:"latency"`
	ErrorMessage string        `json:"error,omitempty"`
	UserID       string        `json:"user_id,omitempty"`
}

// Logger returns a middleware that logs requests with detailed information
func Logger(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Get or create request ID
		requestID := c.GetString("requestID")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
				c.Set("requestID", requestID)
			}
		}
		
		// Set X-Request-ID header in response
		c.Header("X-Request-ID", requestID)
		
		// Process request
		c.Next()
		
		// Calculate request duration
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		
		// Get user ID if available
		userID := ""
		if user, exists := c.Get("userID"); exists {
			if id, ok := user.(string); ok {
				userID = id
			}
		}
		
		// Get error if any
		errorMessage := ""
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}
		
		// Log request details
		logger.Printf("[%s] %s %s | %d | %v | %s | UserID: %s | Errors: %s",
			requestID,
			method,
			path,
			statusCode,
			latency,
			clientIP,
			userID,
			errorMessage,
		)
	}
}

// RequestID generates a unique request ID for each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
} 