package images

import (
	"freezetag/backend/pkg/images/imagedata"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkValidWebPHeader(data []byte) bool {
	return len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

func TestCreateProfilePicture(t *testing.T) {
	// Create a simple 2x2 RGBA image (red, green, blue, white)
	width, height := 2, 2
	pixels := []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
	data := imagedata.Data{
		PixelsRGBA: pixels,
		Width:      width,
		Height:     height,
	}

	result, err := CreateProfilePicture(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Optionally check that the result is a valid WEBP image by checking the header
	assert.True(t, checkValidWebPHeader(result))
}

func TestCreateProfilePictureNonSquare(t *testing.T) {
	// Create a simple 4x2 RGBA image (red, green, blue, white)
	width, height := 4, 2
	pixels := []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
	data := imagedata.Data{
		PixelsRGBA: pixels,
		Width:      width,
		Height:     height,
	}
	result, err := CreateProfilePicture(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Check that the result is a valid WEBP image by checking the header.
	// I had AI make this check since I dont know the webp header off rip
	assert.True(t, checkValidWebPHeader(result))
}

func TestCreateProfilePictureInvalidData(t *testing.T) {
	// Test with incorrect pixel data length
	data := imagedata.Data{
		PixelsRGBA: []byte{0, 0, 0}, // Not enough data for 2x2 RGBA
		Width:      2,
		Height:     2,
	}

	result, err := CreateProfilePicture(data)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "incorrect image data structure", err.Error())
}

func TestCreateDefaultProfilePicture(t *testing.T) {
	result, err := DefaultProfilePicture("1")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.True(t, checkValidWebPHeader(result))
}
