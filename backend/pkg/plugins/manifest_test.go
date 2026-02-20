package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicManifestImport(t *testing.T) {
	manifest, err := ReadManifest("test_resources/tagger")
	assert.NoError(t, err)
	assert.Equal(t, "tagger", manifest.Name)
	assert.Equal(t, PostUpload, manifest.Hooks["tag_image"].Type)
}

func TestInvalidManifestImport(t *testing.T) {
	manifest, err := ReadManifest("test_resources/invalid_manifest")
	assert.Error(t, err)
	assert.Zero(t, manifest)
}

func TestNonexistentManifestImport(t *testing.T) {
	manifest, err := ReadManifest("test_resources/nonexistent")
	assert.Error(t, err)
	assert.Zero(t, manifest)
}

func TestFilteredPluginInfo(t *testing.T) {
	info := PluginInfo{
		Name:    "foo",
		Version: "1.2.3",
		Enabled: true,
		Hooks: map[string]PluginHook{
			"foo": {Type: PostUpload, Signature: ProcessOneImage},
			"bar": {Type: ManualTrigger, Signature: ProcessImageBatch},
		},
	}
	t.Run("filter by type", func(t *testing.T) {
		hooks := info.HooksWithType(PostUpload)
		assert.Len(t, hooks, 1)
		assert.Contains(t, hooks, HookInfo{
			Name: "foo",
			Hook: PluginHook{Type: PostUpload, Signature: ProcessOneImage},
		})
	})

	t.Run("filter by signature", func(t *testing.T) {
		hooks := info.HooksWithSignature(ProcessImageBatch)
		assert.Len(t, hooks, 1)
		assert.Contains(t, hooks, HookInfo{
			Name: "bar",
			Hook: PluginHook{Type: ManualTrigger, Signature: ProcessImageBatch},
		})
	})

	t.Run("filter by both", func(t *testing.T) {
		hooks := info.FilterHooks(PostUpload, ProcessOneImage)
		assert.Len(t, hooks, 1)
		assert.Contains(t, hooks, HookInfo{
			Name: "foo",
			Hook: PluginHook{Type: PostUpload, Signature: ProcessOneImage},
		})
	})
}

func TestJSONMethods(t *testing.T) {
	type testData struct {
		A HookType      `json:"a"`
		B HookSignature `json:"b"`
	}
	data := testData{
		A: PostUpload,
		B: ProcessImageBatch,
	}
	t.Run("marshalUnmarshalSuccess", func(t *testing.T) {
		marshaled, err := json.Marshal(data)
		require.NoError(t, err)
		var got testData
		require.NoError(t, json.Unmarshal(marshaled, &got))
		assert.Equal(t, data, got)
	})

	t.Run("marshalFailInvalid", func(t *testing.T) {
		_, err := json.Marshal(testData{A: 255, B: ProcessOneImage})
		assert.Error(t, err)
		_, err = json.Marshal(testData{A: PostUpload, B: 255})
		assert.Error(t, err)
	})

	t.Run("unmarshalFailBadType", func(t *testing.T) {
		var got testData
		err := json.Unmarshal([]byte(`{"a":3,"b":"single_image"}`), &got)
		assert.Error(t, err)
		err = json.Unmarshal([]byte(`{"a":"post_upload","b":false}`), &got)
		assert.Error(t, err)
	})

	t.Run("unmarshalFailBadString", func(t *testing.T) {
		var got testData
		err := json.Unmarshal([]byte(`{"a":"foo","b":"single_image"}`), &got)
		assert.Error(t, err)
		err = json.Unmarshal([]byte(`{"a":"post_upload","b":"bar"}`), &got)
		assert.Error(t, err)
	})
}
