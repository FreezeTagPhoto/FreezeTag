package formats

import (
	"bytes"
	"log"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePNGInvalid(t *testing.T) {
	data, err := ParseBasic("test.png", []byte{0x69, 0x4, 0x20}) // obviously not a PNG
	assert.ErrorContains(t, err, "failed to convert")
	assert.Zero(t, data)
}

func TestParsePNGIncompleteGPS(t *testing.T) {
	data, err := os.ReadFile("test_resources/incompleteGPS.png")
	require.NoError(t, err)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	imageData, err := ParseBasic("incompleteGPS.png", data)
	assert.NoError(t, err)
	assert.Nil(t, imageData.Geo)
	assert.Contains(t, buf.String(), "incomplete GPS data")
}

func TestParsePNGWithGPSData(t *testing.T) {
	data, err := os.ReadFile("test_resources/completeGPS.png")
	require.NoError(t, err)
	imageData, err := ParseBasic("completeGPS.png", data)
	assert.NoError(t, err)
	assert.NotNil(t, imageData.Geo)
	assert.Equal(t, 6.9, imageData.Geo.Lat)
	assert.Equal(t, 42.0, imageData.Geo.Lon)
	assert.Equal(t, 67.0, imageData.Geo.Alt)
}

func TestParsePNGWithEXIFDate(t *testing.T) {
	data, err := os.ReadFile("test_resources/withExifDate.png")
	require.NoError(t, err)
	imageData, err := ParseBasic("withExifDate.png", data)
	assert.NoError(t, err)
	if assert.NotNil(t, imageData.DateCreated) {
		assert.Equal(t, 2077, imageData.DateCreated.Year(), "exif data should have a funny year")
	}
}

func TestParsePNGWithNoEXIFDate(t *testing.T) {
	data, err := os.ReadFile("test_resources/gopher.png")
	require.NoError(t, err)
	imageData, err := ParseBasic("gopher.png", data)
	assert.NoError(t, err)
	assert.Nil(t, imageData.DateCreated)
}

func TestParsePNGWithManufacturerInfo(t *testing.T) {
	data, err := os.ReadFile("test_resources/completeManufacturer.png")
	require.NoError(t, err)
	imageData, err := ParseBasic("completeManufacturer.png", data)
	assert.NoError(t, err)
	if assert.NotNil(t, imageData.Cam) {
		assert.Equal(t, "google", imageData.Cam.Manufacturer)
		assert.Equal(t, "gopher", imageData.Cam.Model)
	}
}

func TestParsePNGGivesCorrectBytes(t *testing.T) {
	data, err := os.ReadFile("test_resources/teenytiny.png")
	require.NoError(t, err)
	imageData, err := ParseBasic("teenytiny.png", data)
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

type basicParseTestCase struct {
	name string
	file string
}

func TestVariousBasicFormats(t *testing.T) {
	tests := []basicParseTestCase{
		{name: "Parse JPG", file: "test_resources/gopher.jpg"},
		{name: "Parse TIFF", file: "test_resources/gopher.tiff"},
		{name: "Parse WEBP", file: "test_resources/gopher.webp"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			require.NoError(t, err)
			imageData, err := ParseBasic(tc.file, data)
			assert.NoError(t, err)
			assert.Equal(t, 32, imageData.Width)
			assert.Equal(t, 32, imageData.Height)
		})
	}
}
