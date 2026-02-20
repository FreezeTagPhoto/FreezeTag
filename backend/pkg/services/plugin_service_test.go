package services

import (
	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/plugins"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fooInfo plugins.PluginInfo = plugins.PluginInfo{Name: "foo", Version: "0.0.1", Enabled: true, Hooks: map[string]plugins.PluginHook{
	"tag_image":   {Type: plugins.PostUpload, Signature: plugins.ProcessOneImage},
	"tag_image_2": {Type: plugins.PostUpload, Signature: plugins.ProcessOneImage},
}}

var barInfo plugins.PluginInfo = plugins.PluginInfo{Name: "bar", Version: "0", Enabled: true, Hooks: map[string]plugins.PluginHook{
	"tag_image":    {Type: plugins.PostUpload, Signature: plugins.ProcessOneImage},
	"locate_image": {Type: plugins.PostUpload, Signature: plugins.ProcessImageBatch},
}}

func TestCreatePluginService(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	plugs := serv.Plugins()
	assert.Contains(t, plugs, fooInfo)
	assert.Contains(t, plugs, barInfo)
	assert.Equal(t, 2, len(plugs))
}

func TestNonexistentPluginDirectory(t *testing.T) {
	repo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./nonexistent", repo)
	assert.Error(t, err)
	assert.Zero(t, serv)
}

func TestEnableDisablePlugin(t *testing.T) {
	repo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", repo)
	require.NoError(t, err)
	serv.SetEnabled("bar", false)
	plugs := serv.Plugins()
	expected := barInfo
	expected.Enabled = false
	assert.Contains(t, plugs, expected)
	assert.Equal(t, 2, len(plugs))
}

func TestGetPluginInfo(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	info := serv.PluginInfo("foo")
	require.NotNil(t, info)
	assert.Equal(t, fooInfo, *info)
}

func TestGetNoPluginInfo(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	info := serv.PluginInfo("baz")
	assert.Nil(t, info)
}

func TestLaunchNoPlugin(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	_, err = serv.LaunchPlugin("bar", t.Context())
	assert.Error(t, err)
}
