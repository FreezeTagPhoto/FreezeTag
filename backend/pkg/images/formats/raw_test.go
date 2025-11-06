package formats

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rawTestCase struct {
	name   string
	file   string
	width  int
	height int
}

func TestInvalidRawImage(t *testing.T) {
	data := []byte{0x69, 0x4, 0x20} // trivially wrong
	imageData, err := ParseRaw("gopher.heic", data)
	assert.ErrorContains(t, err, "failed to convert")
	assert.Zero(t, imageData)
}

func TestRawImageMissingGeo(t *testing.T) {
	data, err := os.ReadFile("test_resources/incompleteGPS.heic")
	require.NoError(t, err)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	imageData, err := ParseRaw("incompleteGPS.heic", data)
	assert.NoError(t, err)
	assert.Nil(t, imageData.Geo)
	assert.Contains(t, buf.String(), "incomplete GPS data")
}

func TestVariousRawFormats(t *testing.T) {
	tests := []rawTestCase{
		{name: "Parse HEIC 1", file: "test_resources/gopher.heic", width: 32, height: 32},
		{name: "Parse HEIC 2", file: "test_resources/gopher.heif", width: 32, height: 32},
		{name: "Parse CR3", file: "test_resources/tree.CR3", width: 8191, height: 5463},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			require.NoError(t, err)
			imageData, err := ParseRaw(tc.file, data)
			assert.NoError(t, err)
			assert.Equal(t, tc.width, imageData.Width)
			assert.Equal(t, tc.height, imageData.Height)
		})
	}
}
