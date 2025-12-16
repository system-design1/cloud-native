package errors

import (
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// HTTPStatus returns the HTTP status code for the error
func (e *AppError) HTTPStatus() int {
	return e.Code
}

// NewAppError creates a new application error
func NewAppError(code int, message string, details ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// Common error constructors
var (
	ErrBadRequest          = func(message string) *AppError { return NewAppError(http.StatusBadRequest, message) }
	ErrUnauthorized        = func(message string) *AppError { return NewAppError(http.StatusUnauthorized, message) }
	ErrForbidden           = func(message string) *AppError { return NewAppError(http.StatusForbidden, message) }
	ErrNotFound            = func(message string) *AppError { return NewAppError(http.StatusNotFound, message) }
	ErrInternalServerError = func(message string) *AppError { return NewAppError(http.StatusInternalServerError, message) }
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

