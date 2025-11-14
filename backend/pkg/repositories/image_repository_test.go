package repositories

import (
	"freezetag/backend/mocks/ImageDatabase"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/formats"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeMockDBSuccess(t *testing.T) *mocks.MockImageDatabase {
	mockdb := mocks.NewMockImageDatabase(t)
	mockdb.EXPECT().
		AddImage(mock.AnythingOfType("string"), mock.AnythingOfType("imagedata.Data")).
		Return(database.ImageId(42), nil)

	mockdb.EXPECT().
		AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)
	return mockdb
}

func initParserCollection() images.Parser {
	parserCollection := images.InitParserCollection()
	err := parserCollection.RegisterParserFunc("*.{cr3,CR3,nef,NEF,dng,DNG}", formats.ParseRaw)
	require.NoError(nil, err, "registering RAW parser should not fail")
	
	err = parserCollection.RegisterParserFunc("*.{png,jpg,jpeg}", formats.ParseBasic)
	require.NoError(nil, err, "registering basic parser should not fail")
	
	return parserCollection
}

func TestStoreImageBytesSuccess(t *testing.T) {

	tempDir := t.TempDir() //tempdir will be cleaned up automatically after the test
	mockdb := makeMockDBSuccess(t)
	parser := initParserCollection()
	repo := InitImageRepository(mockdb, parser, tempDir)
	
	data, err := os.ReadFile("./test_resources/completeGPS.png")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "gopher.png")

	assert.Nil(t, result.Err, "expected no error in result")
	assert.NotNil(t, result.Success, "expected success in result")
	assert.Equal(t, result.Success.Id, database.ImageId(42))

    data2, err := os.ReadFile("test_resources/completeGPS.png")
    assert.NoError(t, err, "failed to read file2")
	assert.Equal(t, data, data2, "written file is not the uploaded file")
}
