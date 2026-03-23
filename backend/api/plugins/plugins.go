package plugins

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	plugs "freezetag/backend/pkg/plugins"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PluginEndpoint struct {
	service services.PluginService
	jobs    services.JobService
}

func InitPluginEndpoint(service services.PluginService, jobs services.JobService) PluginEndpoint {
	return PluginEndpoint{service, jobs}
}

// @summary     List Plugins
// @description List plugins along with their hooks and enabled status
// @tags        plugins
// @produce     application/json
// @router      /plugins/list [get]
// @success     200 {array} plugins.PluginInfo
func (pe PluginEndpoint) ListAll(c *gin.Context) {
	c.JSON(http.StatusOK, pe.service.Plugins())
}

// @summary     Set Enabled
// @description Set the enabled status of a plugin by name (plugin names are unique)
// @tags        plugins
// @produce     application/json
// @router      /plugins/enable [post]
// @param       plugin  query string true "plugin to enable/disable"
// @param       enabled query bool true "whether the plugin should be enabled"
// @success     200 {array}  api.PluginDisabledResponse
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

// @summary     Run Manual Plugin Hook
// @description Start running a manual plugin hook by name
// @tags        plugins
// @produce     application/json
// @router      /plugins/run [post]
// @param       plugin query string true "plugin to run"
// @param       hook query string true "hook to run"
// @param       input body any true "input to hook"
// @success     202 {object} string "the UUID of the created job for the plugin run"
// @failure     400 {object} api.BadRequestResponse
func (pe PluginEndpoint) RunManual(c *gin.Context) {
	plug := c.Query("plugin")
	hook := c.Query("hook")
	info := pe.service.PluginInfo(plug)
	if info == nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("no plugin with name '%s'", plug)})
		return
	}
	sig, ok := info.Hooks[hook]
	if !ok {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("plugin %v has no hook '%s'", plug, hook)})
		return
	}
	if sig.Type != plugs.ManualTrigger {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("hook %s:%s is not a manual hook", plug, hook)})
		return
	}
	var input any
	switch sig.Signature {
	case plugs.ProcessOneImage:
		var id int64
		if err := c.BindJSON(&id); err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("failed to parse request body (expected an ID): %v", err)})
			return
		}
		input = database.ImageId(id)
	case plugs.ProcessImageBatch:
		var in []int64
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("failed to parse request body (expected an array of IDs): %v", err)})
			return
		}
		ids := make([]database.ImageId, len(in))
		for i, id := range in {
			ids[i] = database.ImageId(id)
		}
		input = ids
	}
	id := pe.jobs.SchedulePluginHook(plug, hook, input)
	c.JSON(http.StatusAccepted, id)
}
