package plugins

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PluginEndpoint struct {
	service services.PluginService
}

func InitPluginEndpoint(service services.PluginService) PluginEndpoint {
	return PluginEndpoint{service}
}

// @summary     List Plugins
// @description List plugins along with their hooks and enabled status
// @tags        plugins
// @produce     application/json
// @router      /plugins [get]
// @success     200 {array} plugins.PluginInfo
func (pe PluginEndpoint) ListAll(c *gin.Context) {
	c.JSON(http.StatusOK, pe.service.Plugins())
}

// @summary     Set Enabled
// @description Set the enabled status of a plugin by name (plugin names are unique)
// @tags        plugins
// @produce     application/json
// @router      /plugins [post]
// @param       plugin  query string true "plugin to enable/disable"
// @param       enabled query bool true "whether the plugin should be enabled"
// @success     200 {array}  plugins.PluginInfo
// @failure     400 {object} api.BadRequestResponse
func (pe PluginEndpoint) SetEnabled(c *gin.Context) {
	plug := c.Query("plugin")
	if plug == "" {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no plugin parameter"})
		return
	}
	enabledParam := c.Query("enabled")
	var enabled bool
	switch enabledParam {
	case "true":
		enabled = true
	case "false":
		enabled = false
	default:
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad enabled parameter"})
		return
	}
	pe.service.SetEnabled(plug, enabled)
	c.JSON(http.StatusOK, api.PluginDisabledResponse{
		Disabled: !enabled,
	})
}
