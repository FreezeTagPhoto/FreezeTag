package services

import (
	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/plugins"
	"freezetag/backend/pkg/repositories"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreatePluginService(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	plugs := serv.AllPlugins()
	assert.Contains(t, plugs, "foo")
	assert.Contains(t, plugs, "bar")
	assert.Equal(t, 2, len(plugs))
}

func TestListPluginHooks(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	hooks := serv.AllHooks("foo")
	assert.Contains(t, hooks, "tag_image")
	assert.Contains(t, hooks, "tag_image_2")
}

func TestFilterPluginHooks(t *testing.T) {
	imgRepo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	filteredTypeHooks := serv.HooksWithType("bar", plugins.PreUpload)
	assert.Contains(t, filteredTypeHooks, "locate_image")
	assert.Equal(t, 1, len(filteredTypeHooks))
	filteredSigHooks := serv.HooksWithSignature("bar", plugins.ImageProcess)
	assert.Contains(t, filteredSigHooks, "tag_image")
	assert.Equal(t, 1, len(filteredSigHooks))
	filteredBothHooks := serv.Hooks("bar", plugins.PreUpload, plugins.ImageProcess)
	assert.Empty(t, filteredBothHooks)
}

func TestPhonyPostUploadJob(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/bar/.venv") //nolint:errcheck
	})
	imgRepo := mocks.NewMockImageRepository(t)
	data, err := os.ReadFile("test_resources/gopher.webp")
	require.NoError(t, err)
	imgRepo.EXPECT().RetrieveThumbnail(mock.AnythingOfType("ImageId"), uint(2)).Return(data, nil)
	imgRepo.EXPECT().AddImageTags(database.ImageId(1), []string{"1"}).Return(repositories.ImageTagResult{})
	imgRepo.EXPECT().AddImageTags(database.ImageId(2), []string{"2"}).Return(repositories.ImageTagResult{})
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	fakeUploadJob := []*repositories.ImageUploadSuccess{
		{Id: database.ImageId(1), Filename: "foo.png"},
		{Id: database.ImageId(2), Filename: "bar.jpg"},
	}
	results, err := serv.RunPostUpload("bar", t.Context(), fakeUploadJob)
	require.NoError(t, err)
	for {
		err, ok := <-results
		if !ok {
			break
		}
		assert.NoError(t, err)
	}
}

func TestPhonyPostUploadJobMultiHooks(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/foo/.venv") //nolint:errcheck
	})
	imgRepo := mocks.NewMockImageRepository(t)
	data, err := os.ReadFile("test_resources/gopher.webp")
	require.NoError(t, err)
	imgRepo.EXPECT().RetrieveThumbnail(mock.AnythingOfType("ImageId"), uint(2)).Return(data, nil)
	imgRepo.EXPECT().AddImageTags(database.ImageId(1), []string{"1"}).Return(repositories.ImageTagResult{})
	imgRepo.EXPECT().AddImageTags(database.ImageId(2), []string{"2"}).Return(repositories.ImageTagResult{})
	imgRepo.EXPECT().AddImageTags(database.ImageId(1), []string{"3"}).Return(repositories.ImageTagResult{})
	imgRepo.EXPECT().AddImageTags(database.ImageId(2), []string{"6"}).Return(repositories.ImageTagResult{})
	serv, err := InitDefaultPluginService("./test_resources", imgRepo)
	assert.NoError(t, err)
	fakeUploadJob := []*repositories.ImageUploadSuccess{
		{Id: database.ImageId(1), Filename: "foo.png"},
		{Id: database.ImageId(2), Filename: "bar.jpg"},
	}
	results, err := serv.RunPostUpload("foo", t.Context(), fakeUploadJob)
	require.NoError(t, err)
	for {
		err, ok := <-results
		if !ok {
			break
		}
		assert.NoError(t, err)
	}
}

func TestNonexistentPluginDirectory(t *testing.T) {
	repo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./nonexistent", repo)
	assert.Error(t, err)
	assert.Zero(t, serv)
}

func TestNonexistentPluginJob(t *testing.T) {
	repo := mocks.NewMockImageRepository(t)
	serv, err := InitDefaultPluginService("./test_resources", repo)
	assert.NoError(t, err)
	fakeUploadJob := []*repositories.ImageUploadSuccess{
		{Id: database.ImageId(1), Filename: "foo.png"},
		{Id: database.ImageId(2), Filename: "bar.jpg"},
	}
	results, err := serv.RunPostUpload("nonexistent", t.Context(), fakeUploadJob)
	assert.Error(t, err)
	assert.Zero(t, results)
}
