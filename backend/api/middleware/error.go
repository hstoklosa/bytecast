package middleware

import (
	"bytecast/internal/errors"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorResponse standardizes the error response format
type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler is a middleware that catches panics and handles returned errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				
				// Return a 500 error
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Status:  "Internal Server Error",
					Message: "An unexpected error occurred",
				})
				c.Abort()
			}
		}()

		c.Next()

		// Check if there are any errors set during the request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError inspects the error type and responds accordingly
func handleError(c *gin.Context, err error) {
	// Check if it's our custom AppError
	if appErr, ok := err.(errors.AppError); ok {
		// Use the predefined status code and message
		c.JSON(appErr.Code, ErrorResponse{
			Status:  appErr.StatusText,
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// For non-AppError types, return a generic internal server error
	log.Printf("Unhandled error: %v", err)
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Status:  "Internal Server Error",
		Message: "An unexpected error occurred",
	})
} 