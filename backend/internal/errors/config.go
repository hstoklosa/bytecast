package errors

import (
	"fmt"
)

func NewConfigurationError(message string, err error) AppError {
	return AppError{
		Code:       500,
		StatusText: "Configuration Error",
		Message:    message,
		Err:        err,
	}
}

func NewMissingConfigError(configKey string, err error) AppError {
	message := fmt.Sprintf("Missing required configuration: %s", configKey)
	return NewConfigurationError(message, err)
}

func NewInvalidConfigError(configKey string, reason string, err error) AppError {
	message := fmt.Sprintf("Invalid configuration for %s: %s", configKey, reason)
	return NewConfigurationError(message, err)
} 