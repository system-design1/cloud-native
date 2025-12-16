package middleware

import (
	"net/http"

	apperrors "go-backend-service/pkg/errors"
	"go-backend-service/internal/logger"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware creates a global error handler middleware
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get correlation ID from context
			correlationIDValue, _ := c.Get("correlation_id")
			var correlationID string
			if correlationIDValue != nil {
				correlationID = correlationIDValue.(string)
			}
			log := logger.Get(correlationID)

			// Get the last error
			err := c.Errors.Last()

			// Determine error type and create appropriate response
			var appErr *apperrors.AppError
			var statusCode int
			var message string
			var details string

			// Check if it's an AppError
			var ok bool
			if appErr, ok = err.Err.(*apperrors.AppError); ok {
				statusCode = appErr.HTTPStatus()
				message = appErr.Message
				details = appErr.Details
			} else {
				// Generic error - don't expose internal details
				statusCode = http.StatusInternalServerError
				message = "Internal server error"
				details = "" // Don't expose internal error details to client
			}

			// Log error with correlation ID and relevant details
			// Note: correlation_id is already included by logger.Get()
			log.Error().
				Str("error", err.Error()).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Int("status_code", statusCode).
				Msg("Request error occurred")

			// Send standardized error response
			c.JSON(statusCode, apperrors.ErrorResponse{
				Error:     http.StatusText(statusCode),
				Message:   message,
				Code:      statusCode,
				Details:   details,
				RequestID: correlationID,
			})

			// Abort the request
			c.Abort()
		}
	}
}

// ErrorHandler is a helper function to set errors in Gin context
func ErrorHandler(c *gin.Context, err error) {
	// Check if it's already an AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.Error(appErr)
		return
	}

	// Wrap generic errors as internal server errors
	// Don't expose internal error details to client
	appErr := apperrors.ErrInternalServerError("An unexpected error occurred")
	c.Error(appErr)

	// Log the actual error internally
	correlationIDValue, _ := c.Get("correlation_id")
	var correlationID string
	if correlationIDValue != nil {
		correlationID = correlationIDValue.(string)
	}
	log := logger.Get(correlationID)
	// Note: correlation_id is already included by logger.Get()
	log.Error().
		Err(err).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("Internal error occurred")
}

