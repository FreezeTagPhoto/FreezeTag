package formats

import (
	"bytes"
	"freezetag/backend/pkg/images"
	"log"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePNGInvalid(t *testing.T) {
	data, err := ParsePNG("test.png", []byte{0x69, 0x4, 0x20}) // obviously not a PNG
	assert.Equal(t, data, images.Data{})
	assert.ErrorContains(t, err, "failed to convert")
}

func TestParsePNGIncompleteGPS(t *testing.T) {
	data, err := os.ReadFile("test/incompleteGPS.png")
	require.NoError(t, err)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	imageData, err := ParsePNG("incompleteGPS.png", data)
	assert.NoError(t, err)
	assert.Nil(t, imageData.Geo)
	assert.Contains(t, buf.String(), "incomplete GPS data")
}

func TestParsePNGWithGPSData(t *testing.T) {
	data, err := os.ReadFile("test/completeGPS.png")
	require.NoError(t, err)
	imageData, err := ParsePNG("completeGPS.png", data)
	assert.NoError(t, err)
	assert.NotNil(t, imageData.Geo)
	assert.Equal(t, 6.9, imageData.Geo.Lat)
	assert.Equal(t, 42.0, imageData.Geo.Lon)
	assert.Equal(t, 67.0, imageData.Geo.Alt)
}

func TestParsePNGWithEXIFDate(t *testing.T) {
	data, err := os.ReadFile("test/withExifDate.png")
	require.NoError(t, err)
	imageData, err := ParsePNG("withExifDate.png", data)
	assert.NoError(t, err)
	if assert.NotNil(t, imageData.DateCreated) {
		assert.Equal(t, 2077, imageData.DateCreated.Year(), "exif data should have a funny year")
	}
}

func TestParsePNGWithNoEXIFDate(t *testing.T) {
	data, err := os.ReadFile("test/gopher.png")
	require.NoError(t, err)
	imageData, err := ParsePNG("gopher.png", data)
	assert.NoError(t, err)
	assert.Nil(t, imageData.DateCreated)
}

func TestParsePNGWithManufacturerInfo(t *testing.T) {
	data, err := os.ReadFile("test/completeManufacturer.png")
	require.NoError(t, err)
	imageData, err := ParsePNG("completeManufacturer.png", data)
	assert.NoError(t, err)
	if assert.NotNil(t, imageData.Cam) {
		assert.Equal(t, "google", imageData.Cam.Manufacturer)
		assert.Equal(t, "gopher", imageData.Cam.Model)
	}
}

func TestParsePNGGivesCorrectBytes(t *testing.T) {
	data, err := os.ReadFile("test/teenytiny.png")
	assert.NoError(t, err)
	imageData, err := ParsePNG("teenytiny.png", data)
	assert.NoError(t, err)
	assert.True(t, slices.Equal(imageData.PixelsRGBA, []byte{
		0x48, 0x6C, 0x55, 0xFF,
		0x48, 0x6C, 0x55, 0xFF,
		0x48, 0x6C, 0xAA, 0xFF,
		0x48, 0x6C, 0x55, 0xFF,
	}), "teenytiny.png should have a specific RGBA conversion")
	assert.Equal(t, 2, imageData.Width)
	assert.Equal(t, 2, imageData.Height)
}
