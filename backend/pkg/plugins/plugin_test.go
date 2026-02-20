package plugins

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/repositories"
)

func TestEchoedPluginInitializes(t *testing.T) {
	// this works because cat echoes stdin
	// and the simplest protocol that succeeds is READY -> READY -> SHUTDOWN -> SHUTDOWN
	ctx, cancel := context.WithCancel(t.Context())
	cmd := exec.CommandContext(ctx, "cat")
	plugin, err := PluginFromProcess("cat", cmd, cancel)
	assert.NoError(t, err)
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestEchoedPluginEchoes(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cmd := exec.CommandContext(ctx, "cat")
	plugin, err := PluginFromProcess("cat", cmd, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "cat", plugin.Name())
	plugin.IO().In <- PluginMessage{BIN, []byte{1, 2, 3}}
	select {
	case msg := <-plugin.IO().Out:
		assert.Equal(t, PluginMessage{BIN, []byte{1, 2, 3}}, msg)
	case <-time.After(time.Second):
		assert.Fail(t, "capturing output didn't work")
	}
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestPluginWithBadInitHook(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/badinit/.venv") //nolint:errcheck
	})
	manifest, err := ReadManifest("test_resources/badinit")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	assert.Error(t, err)
	assert.Zero(t, plugin)
}

func TestPluginWithBadTeardownHook(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/badteardown/.venv") //nolint:errcheck
	})
	manifest, err := ReadManifest("test_resources/badteardown")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	err = plugin.Shutdown()
	assert.Error(t, err)
}

func createMockRepo(t *testing.T) *mocks.MockImageRepository {
	data, err := os.ReadFile("test_resources/gopher.webp")
	require.NoError(t, err)
	repo := mocks.NewMockImageRepository(t)
	repo.EXPECT().RetrieveThumbnail(mock.AnythingOfType("ImageId"), uint(2)).Return(data, nil)
	return repo
}

func TestTaggingPluginWithManifest(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	repo.EXPECT().AddImageTags(database.ImageId(42), []string{"foo", "bar"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 42, Count: 2}})
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	for _, hook := range plugin.GetHooks(PostUpload, ProcessOneImage) {
		res, err := plugin.processOneImage(hook, database.ImageId(42), repo)
		assert.NoError(t, err)
		assert.NotZero(t, res)
	}
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestTaggingPluginWithManifestAndRequirements(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger_numpy/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	repo.EXPECT().AddImageTags(database.ImageId(42), []string{"2", "3", "4", "5", "6"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 42, Count: 5}})
	manifest, err := ReadManifest("test_resources/tagger_numpy")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	for _, hook := range plugin.GetHooks(PostUpload, ProcessOneImage) {
		res, err := plugin.processOneImage(hook, database.ImageId(42), repo)
		assert.NoError(t, err)
		assert.NotZero(t, res)
	}
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestReuseSameVenv(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	repo.EXPECT().AddImageTags(database.ImageId(1), []string{"foo", "bar"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 1, Count: 2}})
	repo.EXPECT().AddImageTags(database.ImageId(2), []string{"foo", "bar"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 2, Count: 2}})
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	for i := 1; i <= 2; i++ {
		plugin, err := PluginFromManifest(manifest, t.Context())
		require.NoError(t, err)
		for _, hook := range plugin.GetHooks(PostUpload, ProcessOneImage) {
			res, err := plugin.RunHook(hook, repositories.ImageUploadSuccess{Id: database.ImageId(i)}, repo)
			assert.NoError(t, err)
			assert.NotZero(t, res)
		}
		err = plugin.Shutdown()
		assert.NoError(t, err)
	}
}

func TestMultipleActionsOnePlugin(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	for i := 1; i <= 4; i++ {
		repo.EXPECT().AddImageTags(database.ImageId(i), []string{"foo", "bar"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: database.ImageId(i), Count: 2}})
	}
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	for i := 1; i <= 4; i++ {
		res, err := plugin.RunHook(plugin.GetHooks(PostUpload, ProcessOneImage)[0], repositories.ImageUploadSuccess{Id: database.ImageId(i)}, repo)
		assert.NoError(t, err)
		assert.NotZero(t, res)
	}
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestPluginProcessError(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	_, err = plugin.RunHook(plugin.GetHooks(PostUpload, ProcessOneImage)[0], repositories.ImageUploadSuccess{Id: 67}, repo)
	assert.Error(t, err)
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestPluginMetadataRequest(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	filename := "abc.png"
	repo.EXPECT().GetImageMetadata(database.ImageId(76)).Return(imagedata.Metadata{FileName: &filename}, nil)
	repo.EXPECT().AddImageTags(database.ImageId(76), []string{"abc.png"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 76, Count: 1}})
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	res, err := plugin.RunHook(plugin.GetHooks(PostUpload, ProcessOneImage)[0], repositories.ImageUploadSuccess{Id: 76}, repo)
	assert.NoError(t, err)
	assert.NotZero(t, res)
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestImageBatchPlugin(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/tagger/.venv") //nolint:errcheck
	})
	repo := mocks.NewMockImageRepository(t)
	var in []repositories.ImageUploadSuccess
	for i := range 10 {
		repo.EXPECT().AddImageTags(database.ImageId(i), []string{"foo"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: database.ImageId(i), Count: 1}})
		in = append(in, repositories.ImageUploadSuccess{Id: database.ImageId(i)})
	}
	manifest, err := ReadManifest("test_resources/tagger")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	res, err := plugin.RunHook("tag_batch", in, repo)
	assert.NoError(t, err)
	assert.NotZero(t, res)
}

func TestEveryHook(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("test_resources/every_hook/.venv") //nolint:errcheck
	})
	repo := createMockRepo(t)
	var in []repositories.ImageUploadSuccess
	for i := range 11 {
		in = append(in, repositories.ImageUploadSuccess{Id: database.ImageId(i)})
	}
	repo.EXPECT().AddImageTags(database.ImageId(1), []string{"foo"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 1, Count: 1}})
	repo.EXPECT().RemoveImageTags(database.ImageId(1), []string{"foo"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 1, Count: 1}})
	repo.EXPECT().DeleteTags([]string{"foo"}).Return(1, nil)
	repo.EXPECT().DeleteImage(database.ImageId(1)).Return("foo", nil)
	repo.EXPECT().StoreImageBytes(mock.AnythingOfType("[]uint8"), "foo.webp").Return(database.ImageId(2), nil)
	repo.EXPECT().SearchImage(mock.Anything).Return([]database.ImageId{4, 20}, nil)
	repo.EXPECT().RetrieveImageTags(database.ImageId(1)).Return([]string{"foo"}, nil)
	repo.EXPECT().RetrieveAllTags().Return(map[string]int64{"foo": 1, "bar": 2}, nil)
	repo.EXPECT().GetQueryTagCounts(mock.Anything).Return(map[string]int64{"foo": 1, "bar": 2}, nil)
	manifest, err := ReadManifest("test_resources/every_hook")
	require.NoError(t, err)
	plugin, err := PluginFromManifest(manifest, t.Context())
	require.NoError(t, err)
	for _, input := range in {
		res, err := plugin.RunHook("every", input, repo)
		assert.NoError(t, err)
		assert.NotZero(t, res)
	}
	assert.NoError(t, plugin.Shutdown())
}
