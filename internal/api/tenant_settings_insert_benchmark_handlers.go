package api

import (
	"net/http"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// InsertTenantSettingsBenchmarkHandler handles POST /v1/otp/tenant-settings-insert-benchmark
func InsertTenantSettingsBenchmarkHandler(repo *repository.TenantSettingsInsertRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCode, _ := c.Get("correlation_id")
		code, ok := tenantCode.(string)
		if !ok || code == "" {
			code = c.GetHeader("X-Correlation-ID")
		}
		if code == "" {
			middleware.ErrorHandler(c, apperrors.ErrInternalServerError("missing correlation_id"))
			return
		}

		ctx := c.Request.Context()
		id, err := repo.InsertTenantSettingsForInsertNew(ctx, code)
		if err != nil {
			middleware.ErrorHandler(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}
