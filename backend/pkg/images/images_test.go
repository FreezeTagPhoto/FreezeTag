package images

import (
	"errors"
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/images/imagedata"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserCollectionRegisterOne(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("*.png", func(name string, data []byte) (imagedata.Data, error) {
		return imagedata.Data{
			PixelsRGBA:  []byte{},
			Width:       69,
			Height:      420,
			DateCreated: nil,
			Geo:         nil,
			Cam:         nil,
		}, nil
	})
	assert.NoError(t, err, "registering a parser with a valid glob should not fail")
	info, err := parser.ParseImage("test.png", []byte{})
	require.NoError(t, err, "parsing an image should not fail")
	assert.Equal(t, 69, info.Width)
	assert.Equal(t, 420, info.Height)
	assert.Nil(t, info.DateCreated)
}

func TestParserCollectionRegisterMultiple(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("*.png", func(name string, data []byte) (imagedata.Data, error) {
		return imagedata.Data{}, errors.New("png")
	})
	assert.NoError(t, err, "registering a parser with a valid glob should not fail")
	err = parser.RegisterParserFunc("{*.jpg,*.jpeg}", func(name string, data []byte) (imagedata.Data, error) {
		return imagedata.Data{}, errors.New("jpg")
	})
	assert.NoError(t, err, "registering a parser with a valid glob should not fail")
	_, err = parser.ParseImage("test.jpeg", []byte{})
	assert.ErrorContains(t, err, "jpg", "the 'jpg' parser function should be called")
	_, err = parser.ParseImage("test.png", []byte{})
	assert.ErrorContains(t, err, "png", "the 'png' parser function should be called")
}

func TestParserCollectionNoMatch(t *testing.T) {
	parser := InitParserCollection()
	_, err := parser.ParseImage("test.png", []byte{})
	assert.ErrorContains(t, err, "no parser for file")
}

func TestParserCollectionInvalidGlob(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("^^[{bad}", func(name string, data []byte) (imagedata.Data, error) {
		return imagedata.Data{}, errors.New("foo")
	})
	assert.Error(t, err)
}

type imageParserTestCase struct {
	name     string
	file     string
	fail     bool
	failText string
	width    int
	height   int
}

func TestParserCollectionFormatIntegration(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("*.{cr3,CR3}", formats.ParseRaw)
	require.NoError(t, err)
	err = parser.RegisterParserFunc("*.{png,jpg,jpeg}", formats.ParseBasic)
	require.NoError(t, err)
	cases := []imageParserTestCase{
		{name: "Parse JPG", file: "formats/test_resources/gopher.jpg", width: 32, height: 32},
		{name: "Parse PNG", file: "formats/test_resources/gopher.png", width: 32, height: 32},
		{name: "Parse CR3", file: "formats/test_resources/tree.CR3", width: 8191, height: 5463},
		{name: "Parse HEIC (fail)", file: "formats/test_resources/gopher.heic", fail: true, failText: "no parser for file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			require.NoError(t, err)
			imageData, err := parser.ParseImage(tc.file, data)
			if tc.fail {
				assert.ErrorContains(t, err, tc.failText)
				assert.Zero(t, imageData)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, imageData.PixelsRGBA)
				assert.Equal(t, tc.width, imageData.Width)
				assert.Equal(t, tc.height, imageData.Height)
			}
		})
	}
}
