package middleware

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(required ...data.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		JWTpermissions, exists := c.Get("permissions")
		if !exists || JWTpermissions == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, api.StatusBadRequestResponse{Error: "No permissions found"})
			return
		}
		for _, r := range required {
			perms, ok := JWTpermissions.(data.Permissions)
			if !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, api.StatusBadRequestResponse{Error: "Invalid permission type"})
				return
			}
			if !perms.HasPermission(r) {
				c.AbortWithStatusJSON(http.StatusForbidden, api.StatusBadRequestResponse{Error: "Insufficient permissions"})
				return
			}
		}
		c.Next()
	}
}
