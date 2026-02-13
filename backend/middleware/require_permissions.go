package middleware

import (
	"freezetag/backend/pkg/database/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(required ...data.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		JWTpermissions, exists := c.Get("permissions")
		if !exists || JWTpermissions == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permissions not found in context"})
			return
		}
		for _, r := range required {
			if !JWTpermissions.(data.Permissions).HasPermission(r) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				return
			}
		}
		c.Next()
	}
}
