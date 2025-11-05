package images

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserCollectionRegisterOne(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("*.png", func(name string, data []byte) (Info, error) {
		return Info{
			struct {
				pixelsRGBA []byte
				width      int
				height     int
			}{
				[]byte{69, 4, 20},
				69,
				420,
			},
			nil,
		}, nil
	})
	assert.NoError(t, err, "registering a parser with a valid glob should not fail")
	info, err := parser.ParseImage("test.png", []byte{})
	require.NoError(t, err, "parsing an image should not fail")
	assert.Equal(t, 69, info.image.width)
	assert.Equal(t, 420, info.image.height)
	assert.Nil(t, info.meta)
}

func TestParserCollectionRegisterMultiple(t *testing.T) {
	parser := InitParserCollection()
	err := parser.RegisterParserFunc("*.png", func(name string, data []byte) (Info, error) {
		return Info{}, errors.New("png")
	})
	assert.NoError(t, err, "registering a parser with a valid glob should not fail")
	err = parser.RegisterParserFunc("{*.jpg,*.jpeg}", func(name string, data []byte) (Info, error) {
		return Info{}, errors.New("jpg")
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
	err := parser.RegisterParserFunc("^^[{bad}", func(name string, data []byte) (Info, error) {
		return Info{}, errors.New("foo")
	})
	assert.Error(t, err)
}
