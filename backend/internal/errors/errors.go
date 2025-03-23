package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors with metadata
type AppError struct {
	Code       int         `json:"-"`
	StatusText string      `json:"status"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Err        error       `json:"-"`
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// WithDetails adds additional details to the error
func (e AppError) WithDetails(details interface{}) AppError {
	e.Details = details
	return e
}

// Unwrap implements the errors.Unwrap interface
func (e AppError) Unwrap() error {
	return e.Err
}

func NewBadRequest(message string, err error) AppError {
	return AppError{
		Code:       http.StatusBadRequest,
		StatusText: "Bad Request",
		Message:    message,
		Err:        err,
	}
}

func NewUnauthorized(message string, err error) AppError {
	return AppError{
		Code:       http.StatusUnauthorized,
		StatusText: "Unauthorized",
		Message:    message,
		Err:        err,
	}
}

func NewForbidden(message string, err error) AppError {
	return AppError{
		Code:       http.StatusForbidden,
		StatusText: "Forbidden",
		Message:    message,
		Err:        err,
	}
}

func NewNotFound(message string, err error) AppError {
	return AppError{
		Code:       http.StatusNotFound,
		StatusText: "Not Found",
		Message:    message,
		Err:        err,
	}
}

func NewConflict(message string, err error) AppError {
	return AppError{
		Code:       http.StatusConflict,
		StatusText: "Conflict",
		Message:    message,
		Err:        err,
	}
}

func NewValidationError(message string, err error, details interface{}) AppError {
	return AppError{
		Code:       http.StatusUnprocessableEntity,
		StatusText: "Validation Error",
		Message:    message,
		Err:        err,
		Details:    details,
	}
}

func NewInternal(message string, err error) AppError {
	return AppError{
		Code:       http.StatusInternalServerError,
		StatusText: "Internal Server Error",
		Message:    message,
		Err:        err,
	}
}

func NewServiceUnavailable(message string, err error) AppError {
	return AppError{
		Code:       http.StatusServiceUnavailable,
		StatusText: "Service Unavailable",
		Message:    message,
		Err:        err,
	}
} 