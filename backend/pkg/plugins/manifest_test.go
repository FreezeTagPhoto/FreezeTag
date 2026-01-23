package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
