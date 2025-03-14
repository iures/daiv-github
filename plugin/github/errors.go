package github

import (
	"fmt"
)

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// ErrorTypeConfiguration represents configuration errors
	ErrorTypeConfiguration ErrorType = "configuration"
	
	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication"
	
	// ErrorTypeAPI represents GitHub API errors
	ErrorTypeAPI ErrorType = "api"
	
	// ErrorTypeInternal represents internal errors
	ErrorTypeInternal ErrorType = "internal"
	
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
)

// Error represents a structured error with context
type Error struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeConfiguration,
		Message: message,
		Cause:   cause,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeAuthentication,
		Message: message,
		Cause:   cause,
	}
}

// NewAPIError creates a new API error
func NewAPIError(message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeAPI,
		Message: message,
		Cause:   cause,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   cause,
	}
} 
