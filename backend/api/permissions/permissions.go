package permissions

import (
	"freezetag/backend/pkg/database/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionEndpoint struct{}

func initPermissionEndpoint() *PermissionEndpoint {
	return &PermissionEndpoint{}
}

func (p *PermissionEndpoint) ListPermissions(c *gin.Context) {
	permissions := data.AllPermissions()
	c.JSON(http.StatusOK, permissions)
}
