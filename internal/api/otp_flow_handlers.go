package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/otp"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

type otpFlowService interface {
	SendOTP(ctx context.Context, req otp.SendRequest) (*otp.SendResponse, error)
	VerifyOTP(ctx context.Context, req otp.VerifyRequest) (*otp.VerifyResponse, error)
}

type sendOTPRequest struct {
	Phone    string                 `json:"phone"`
	TenantID int64                  `json:"tenant_id"`
	Token    string                 `json:"token"`
	Metadata map[string]interface{} `json:"metadata"`
}

type verifyOTPRequest struct {
	TenantID int64  `json:"tenant_id"`
	Phone    string `json:"phone"`
	Code     string `json:"code"`
}

// SendOTPHandler handles POST /v1/otp/send.
func SendOTPHandler(service otpFlowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendOTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("Invalid request body"))
			return
		}
		if req.TenantID <= 0 {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant_id is required"))
			return
		}
		if strings.TrimSpace(req.Phone) == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone is required"))
			return
		}

		resp, err := service.SendOTP(c.Request.Context(), otp.SendRequest{
			Phone:    req.Phone,
			TenantID: req.TenantID,
			Token:    req.Token,
			Metadata: req.Metadata,
		})
		if err != nil {
			handleOTPServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// VerifyOTPHandler handles POST /v1/otp/verify.
func VerifyOTPHandler(service otpFlowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req verifyOTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("Invalid request body"))
			return
		}
		if req.TenantID <= 0 {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant_id is required"))
			return
		}
		if strings.TrimSpace(req.Phone) == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone is required"))
			return
		}
		if strings.TrimSpace(req.Code) == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("code is required"))
			return
		}

		resp, err := service.VerifyOTP(c.Request.Context(), otp.VerifyRequest{
			TenantID: req.TenantID,
			Phone:    req.Phone,
			Code:     req.Code,
		})
		if err != nil {
			handleOTPServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func handleOTPServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, otp.ErrTenantDisabled):
		middleware.ErrorHandler(c, apperrors.ErrForbidden("Tenant is disabled"))
	case errors.Is(err, otp.ErrTenantNotFound):
		middleware.ErrorHandler(c, apperrors.ErrNotFound("Tenant not found"))
	case errors.Is(err, otp.ErrOTPAlreadyActive):
		middleware.ErrorHandler(c, apperrors.NewAppError(http.StatusTooManyRequests, "OTP already active"))
	case errors.Is(err, otp.ErrSMSProviderFailed):
		middleware.ErrorHandler(c, apperrors.NewAppError(http.StatusBadGateway, "SMS provider failed"))
	default:
		middleware.ErrorHandler(c, apperrors.ErrInternalServerError("An unexpected error occurred"))
	}
}
