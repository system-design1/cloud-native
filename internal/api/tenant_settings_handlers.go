package api

import (
	"net/http"
	"strconv"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// GetTenantSettingsByIDHandler handles requests to fetch tenant settings by ID
// GET /v1/otp/tenant-settings/:id
// Returns 200 with tenant settings JSON on success
// Returns 404 if tenant settings not found
// Returns 400 if id is invalid
func GetTenantSettingsByIDHandler(repo *repository.TenantSettingsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse id from path parameter
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			appErr := apperrors.ErrBadRequest("Invalid tenant settings id: must be a positive integer")
			middleware.ErrorHandler(c, appErr)
			return
		}

		// Validate id is positive
		if id <= 0 {
			appErr := apperrors.ErrBadRequest("Invalid tenant settings id: must be a positive integer")
			middleware.ErrorHandler(c, appErr)
			return
		}

		// Get context from request
		ctx := c.Request.Context()

		// Call repository to get tenant settings
		tenantSettings, err := repo.GetTenantSettingsByID(ctx, id)
		if err != nil {
			// Repository already returns AppError for not found, so just pass it through
			middleware.ErrorHandler(c, err)
			return
		}

		// Return JSON response with all fields
		c.JSON(http.StatusOK, tenantSettings)
	}
}

