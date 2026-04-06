package permissions

import (
	"freezetag/backend/pkg/database/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionEndpoint struct{}

func InitPermissionEndpoint() *PermissionEndpoint {
	return &PermissionEndpoint{}
}

// @summary List all permissions
// @description Retrieve a list of all available permissions in the system.
// @tags permissions
// @router /permissions/list [get]
// @produce application/json
// @success 200 {array} data.Permission "List of permission names"
func (p *PermissionEndpoint) ListPermissions(c *gin.Context) {
	permissions := data.AllPermissions()

	c.JSON(http.StatusOK, permissions)
}
