package api

import (
	"net/http"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/otp"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// GenerateOTPCodeHandler handles OTP code generation requests
// POST /v1/otp/code
// Response: { "code": "123456" }
func GenerateOTPCodeHandler(c *gin.Context) {
	code, err := otp.Generate6DigitCode()
	if err != nil {
		// Wrap the error as an internal server error
		appErr := apperrors.ErrInternalServerError("Failed to generate OTP code")
		middleware.ErrorHandler(c, appErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
	})
}

