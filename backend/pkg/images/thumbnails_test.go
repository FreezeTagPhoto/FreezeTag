package images

import (
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/images/imagedata"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCreateThumbnailGivenValidRGBA(t *testing.T) {
	mockData := imagedata.Data{
		PixelsRGBA: []byte{
			0xFF, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF,
			0x00, 0x00, 0x00, 0xFF,
		},
		Width:  2,
		Height: 2,
	}
	data, err := CreateThumbnail(mockData, 512, 0.8)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestShouldNotCreateThumbnailGivenInvalidRGBA(t *testing.T) {
	mockData := imagedata.Data{
		PixelsRGBA: []byte{
			0xFF, 0x69, 0x42, 0x00,
			0x67, // obviously bad
		},
		Width:  2, // not the correct width anyway
		Height: 2,
	}
	data, err := CreateThumbnail(mockData, 512, 0.8)
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestShrinkingThumbnail(t *testing.T) {
	data, err := os.ReadFile("formats/test_resources/tree.CR3")
	require.NoError(t, err)
	imageData, err := formats.ParseBasic("tree.CR3", data)
	require.NoError(t, err)
	thumbnail, err := CreateThumbnail(imageData, 512, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, thumbnail)
}

func TestNonShrinkingThumbnail(t *testing.T) {
	data, err := os.ReadFile("formats/test_resources/tree.CR3")
	require.NoError(t, err)
	imageData, err := formats.ParseBasic("tree.CR3", data)
	require.NoError(t, err)
	thumbnail, err := CreateThumbnail(imageData, 0, 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, thumbnail)
}
