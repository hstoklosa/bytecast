package utils

import (
	"bytecast/internal/errors"
	"bytecast/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

// HandleError is a helper function to handle errors in controllers
func HandleError(c *gin.Context, err error) {
	_ = c.Error(err) // add to error chain
	c.Abort() // abort the handler chain
}

func HandleValidationError(c *gin.Context, err error, message string) {
	validationErr := utils.HandleValidationError(message, err)
	HandleError(c, validationErr)
}

func LogError(message string, err error) errors.AppError {
	log.Printf("%s: %v", message, err)
	return errors.NewInternal(message, err)
}

// RespondWithError sends an error response directly without going through middleware
// Used only in special cases where middleware error handling is not available
func RespondWithError(c *gin.Context, appErr errors.AppError) {
	c.JSON(appErr.Code, gin.H{
		"status":  appErr.StatusText,
		"message": appErr.Message,
		"details": appErr.Details,
	})
} 