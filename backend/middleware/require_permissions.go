package middleware

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(required ...data.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := hasPermissions(c, required...); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, api.BadRequestResponse{Error: err.Error()})
			return
		}
		c.Next()
	}
}

func RequirePermissionOrSelf(required ...data.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSelf(c) {
			c.Next()
			return
		}
		if err := hasPermissions(c, required...); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, api.BadRequestResponse{Error: err.Error()})
			return
		}
		c.Next()
	}
}

func isSelf(c *gin.Context) bool {
	rawID, idExists := c.Get("userID")
	if !idExists {
		return false
	}
	requesterID, ok := rawID.(string)
	if !ok || requesterID == "" {
		return false
	}
	if c.Param("id") == requesterID {
		return true
	}
	return false
}

func hasPermissions(c *gin.Context, required ...data.Permission) error {
	rawPerms, exists := c.Get("permissions")
	if !exists || rawPerms == nil {
		return fmt.Errorf("no permissions found")
	}

	perms, ok := rawPerms.(data.Permissions)
	if !ok {
		return fmt.Errorf("invalid permission type")
	}

	for _, r := range required {
		if !perms.HasPermission(r) {
			return fmt.Errorf("insufficient permissions")
		}
	}
	return nil
}
