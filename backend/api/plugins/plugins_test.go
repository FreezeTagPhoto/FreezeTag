package plugins

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"freezetag/backend/api"
	jsMocks "freezetag/backend/mocks/JobService"
	mocks "freezetag/backend/mocks/PluginService"
	"freezetag/backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func emptyMockJobService(t *testing.T) services.JobService {
	t.Helper()
	js := jsMocks.NewMockJobService(t)
	return js
}

func TestGetAllPlugins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/plugins", nil)
	ctx.Request = req
	m := mocks.NewMockPluginService(t)
	m.EXPECT().Plugins().Return(nil)
	pe := InitPluginEndpoint(m, emptyMockJobService(t))
	pe.ListAll(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDisablePlugin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/plugins?plugin=foo&enabled=false", nil)
	ctx.Request = req
	m := mocks.NewMockPluginService(t)
	m.EXPECT().SetEnabled("foo", false)
	pe := InitPluginEndpoint(m, emptyMockJobService(t))
	pe.SetEnabled(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.PluginDisabledResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.PluginDisabledResponse{Disabled: true}, got)
}

func TestEnablePlugin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/plugins?plugin=foo&enabled=true", nil)
	ctx.Request = req
	m := mocks.NewMockPluginService(t)
	m.EXPECT().SetEnabled("foo", true)
	pe := InitPluginEndpoint(m, emptyMockJobService(t))
	pe.SetEnabled(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.PluginDisabledResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.PluginDisabledResponse{Disabled: false}, got)
}

func TestDisablePluginNoName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/plugins?enabled=false", nil)
	ctx.Request = req
	m := mocks.NewMockPluginService(t)
	pe := InitPluginEndpoint(m, emptyMockJobService(t))
	pe.SetEnabled(ctx)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDisablePluginBadEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/plugins?plugin=foo&enabled=bar", nil)
	ctx.Request = req
	m := mocks.NewMockPluginService(t)
	pe := InitPluginEndpoint(m, emptyMockJobService(t))
	pe.SetEnabled(ctx)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
