package utils

import (
	"bytecast/internal/errors"

	"github.com/go-playground/validator/v10"
)

// ValidationErrors converts validator errors to a map of field-error pairs
func ValidationErrors(err error) map[string]string {
	validationErrors := make(map[string]string)
	
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			// Format the error based on the validation tag
			validationErrors[e.Field()] = formatValidationError(e)
		}
	}
	
	return validationErrors
}

// formatValidationError generates a user-friendly error message based on validation tag
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "alphanum":
		return "Only alphanumeric characters are allowed"
	default:
		return "Invalid value"
	}
}

// HandleValidationError creates a standardized validation error
func HandleValidationError(message string, err error) errors.AppError {
	validationErrs := ValidationErrors(err)
	return errors.NewValidationError(message, err, validationErrs)
} 