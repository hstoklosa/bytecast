package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig contains the configuration for the CORS middleware
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS returns a middleware that adds CORS headers to every response
// Using the default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		origin := c.GetHeader("Origin")
		allowOrigin := "*"
		
		// If specific origins are defined and the request origin matches one of them,
		// use that specific origin instead of the wildcard
		if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] != "*" {
			allowOrigin = ""
			for _, o := range config.AllowOrigins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
			if allowOrigin == "" {
				allowOrigin = config.AllowOrigins[0]
			}
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		if len(config.ExposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", 
				joinStrings(config.ExposeHeaders, ", "))
		}
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Methods", 
				joinStrings(config.AllowMethods, ", "))
			c.Writer.Header().Set("Access-Control-Allow-Headers", 
				joinStrings(config.AllowHeaders, ", "))
			c.Writer.Header().Set("Access-Control-Max-Age", 
				time.Duration(config.MaxAge.Seconds()).String())
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// Helper function to join strings with separator
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	
	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	
	return result
} 