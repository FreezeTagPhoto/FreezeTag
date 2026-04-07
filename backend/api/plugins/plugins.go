package plugins

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/plugins"
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
	if (sig.Type != plugins.ManualTrigger) && (sig.Type != plugins.GenerateForm) {

		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("hook %s:%s is not a manual hook", plug, hook)})
		return
	}
	var input any
	switch sig.Signature {
	case plugins.ProcessOneImage:
		var id int64
		if err := c.BindJSON(&id); err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("failed to parse request body (expected an ID): %v", err)})
			return
		}
		input = database.ImageId(id)
	case plugins.ProcessImageBatch:
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
	case plugins.ProcessFormData:
		var data map[string]any
		if err := c.BindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: fmt.Sprintf("failed to parse request body (expected json object): %v", err)})
			return
		}
		input = data
	}
	id := pe.jobs.SchedulePluginHook(plug, hook, input)
	c.JSON(http.StatusAccepted, id)
}

// @summary     Get Plugin Configuration
// @description Get the configuration fields of a plugin
// @tags        plugins
// @produce     application/json
// @router      /plugins/config [get]
// @param       plugin query string true "plugin to read configuration from"
// @success     200 {object} map[string]plugins.PublicConfigField "plugin configuration"
// @failure     404 {object} api.NotFoundResponse
// @failure     500 {object} api.ServerErrorResponse
func (pe PluginEndpoint) ReadConfig(c *gin.Context) {
	plug := c.Query("plugin")
	plugin := pe.service.PluginInfo(plug)
	if plugin == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: fmt.Sprintf("plugin %v doesn't exist", plug)})
		return
	}
	config, err := pe.service.ReadConfiguration(plug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	if config == nil {
		config = map[string]plugins.PublicConfigField{}
	}
	c.JSON(http.StatusOK, config)
}

// @summary     Change Plugin Config
// @description Change a map of plugin config values
// @tags        plugins
// @produce     application/json
// @router      /plugins/config [post]
// @param       plugin query string true "plugin to change"
// @param       updates body map[string]any true "field updates"
// @success     200 {object} string
// @failure     404 {object} api.NotFoundResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (pe PluginEndpoint) ChangeConfig(c *gin.Context) {
	plug := c.Query("plugin")
	plugin := pe.service.PluginInfo(plug)
	if plugin == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: fmt.Sprintf("plugin %v doesn't exist", plug)})
		return
	}
	var changes map[string]any
	if err := c.BindJSON(&changes); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse changes, expected a map of new values"})
		return
	}
	err := pe.service.ChangeConfiguration(plug, changes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, "changed configuration")
}

// @summary     Upload Plugin via Git
// @description Add a new plugin by submitting a Git link
// @tags        plugins
// @produce     application/json
// @router      /plugins/upload [post]
// @param       plugin_name query string true "name to give the plugin"
// @param       git_repo body string true "link to pull from"
// @success     200 {object} string
// @failure     404 {object} api.NotFoundResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (pe PluginEndpoint) GitUpload(c *gin.Context) {
	var repo string
	if err := c.BindJSON(&repo); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse changes, expected a string with a link"})
		return
	}
	name := c.Query("plugin_name")
	err := pe.service.DownloadPlugin(name, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, "plugin uploaded")
}
