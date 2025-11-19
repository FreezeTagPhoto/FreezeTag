package repositories

import (
	"fmt"
	mockDatabase "freezetag/backend/mocks/ImageDatabase"
	mockParser "freezetag/backend/mocks/Parser"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"

	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/images/imagedata"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)


// for good data
func initParserCollection() images.Parser {
	parserCollection := images.InitParserCollection()
	err := parserCollection.RegisterParserFunc("*.{cr3,CR3,nef,NEF,dng,DNG}", formats.ParseRaw)
	require.NoError(nil, err, "registering RAW parser should not fail")

	err = parserCollection.RegisterParserFunc("*.{png,jpg,jpeg}", formats.ParseBasic)
	require.NoError(nil, err, "registering basic parser should not fail")

	return parserCollection
}

func TestStoreImageBytesSuccess(t *testing.T) {

	tmpDir := t.TempDir() //tempdir will be cleaned up automatically after the test
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.
		EXPECT().
		AddImage(mock.AnythingOfType("string"), mock.AnythingOfType("imagedata.Data")).
		Return(database.ImageId(42), nil).Times(1)
	mockdb.
		EXPECT().
		AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil).Times(2)

	parser := initParserCollection()
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "gopher.png")

	assert.Nil(t, result.Err, "expected no error in result")
	assert.NotNil(t, result.Success, "expected success in result")
	assert.Equal(t, result.Success.Id, database.ImageId(42))

	data2, err := os.ReadFile("test_resources/gopher1.png")
	assert.NoError(t, err, "failed to read file2")
	assert.Equal(t, data, data2, "written file is not the uploaded file")
}

func TestStoreImageBytesFailParse(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t) // we are expecting a fail on parse, no calls
	mockParser := mockParser.NewMockParser(t)
	mockParser.EXPECT().ParseImage(mock.Anything, mock.Anything).
		Return(imagedata.Data{}, fmt.Errorf("mock error"))

	repo := InitImageRepository(mockdb, mockParser, tmpDir)

	data, err := os.ReadFile("./test_resources/notimage.txt")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "notimage.txt")

	assert.NotNil(t, result.Err, "expected an error in the result")
	assert.Equal(t, result.Err.Filename, "notimage.txt")
	assert.Contains(t, result.Err.Reason, "mock error")
}

func TestStoreImageBytesFailThumbnailGen(t *testing.T) {
	// CreateThumbnail call from StoreImageBytes should fail here,
	// as it is being passed an image that it expects
	// to be 10x10px with 0 pixels
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t) // we are expecting a fail on thumbnail gen, no calls
	mockParser := mockParser.NewMockParser(t)
	mockParser.EXPECT().ParseImage(mock.Anything, mock.Anything).
		Return(imagedata.Data{PixelsRGBA: []byte{}, Height: 10, Width: 10}, nil)
	repo := InitImageRepository(mockdb, mockParser, tmpDir)

	data, err := os.ReadFile("./test_resources/notimage.txt")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "notimage.txt")

	assert.NotNil(t, result.Err, "expected an error in the restult")
	assert.Equal(t, result.Err.Filename, "notimage.txt")
	// not testing error message since we are not trying to test CreateThumbnail directly here
}

func TestStoreImageAddImageFails(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(-1, fmt.Errorf("mock fail"))
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "gopher1.png")

	assert.NotNil(t, result.Err, "expected an error in the restult")
	assert.Equal(t, result.Err.Filename, "gopher1.png")
	assert.Contains(t, result.Err.Reason, "mock fail")
}

func TestStoreImageThumbnailFailsBool(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(1, nil)
	mockdb.EXPECT().AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "gopher1.png")

	assert.NotNil(t, result.Err, "expected an error in the restult")
	assert.Equal(t, result.Err.Filename, "gopher1.png")
	assert.Contains(t, result.Err.Reason, "database returned false when adding thumbnail")
}

func TestStoreImageThumbnailFailsError(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(1, nil)
	mockdb.EXPECT().AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).Return(true, fmt.Errorf("mock error"))
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	result := repo.StoreImageBytes(data, "gopher1.png")

	assert.NotNil(t, result.Err, "expected an error in the restult")
	assert.Equal(t, result.Err.Filename, "gopher1.png")
	assert.Contains(t, result.Err.Reason, "mock error")
}

func TestStoreImageBytesNameCollision(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.
		EXPECT().
		AddImage(mock.AnythingOfType("string"), mock.AnythingOfType("imagedata.Data")).
		Return(database.ImageId(42), nil)
	mockdb.
		EXPECT().
		AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)

	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)

	result := repo.StoreImageBytes(data, "gopher1.png")
	assert.Nil(t, result.Err, "no error in result")
	assert.FileExists(t, tmpDir+"/"+"gopher1.png")

	result2 := repo.StoreImageBytes(data, "gopher1.png")
	assert.Nil(t, result2.Err, "expected error in result")
	assert.FileExists(t, tmpDir+"/"+"copy 1 gopher1.png")
}

func TestGetThumbnailSuccess(t *testing.T) {
	thumbnailBytes := []byte{1, 2, 3}
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImageThumbnail(mock.Anything, mock.Anything).Return(thumbnailBytes, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.RetrieveThumbnail(1, 1)
	assert.Nil(t, err, "error occured when it should have been thumbnailBytes returned")
	assert.Equal(t, result, thumbnailBytes, "thumbnail bytes differ from expected")
}

func TestGetThumbnailFail(t *testing.T) {
	err := fmt.Errorf("mock error")
	thumbnailBytes := []byte{}

	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImageThumbnail(mock.Anything, mock.Anything).Return(thumbnailBytes, err)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.RetrieveThumbnail(1, 1)
	assert.NotNil(t, err, "error didn't occur when it should have when retrieving thumbnail from mockdb")
	assert.Equal(t, result, thumbnailBytes, "thumbnail bytes differ from expected")
}


func TestSearchImageError(t *testing.T) { 
	err := fmt.Errorf("mock error")
	ids := []database.ImageId{}


	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImages(mock.Anything).Return(ids, err)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.SearchImage(queries.CreateImageQuery())
	assert.NotNil(t, err, "error didn't occur when it should have when retrieving imageID from mockdb")
	assert.Equal(t, result, ids)
}

func TestSearchImageNoneReturn(t *testing.T) { 
	ids := []database.ImageId{}

	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImages(mock.Anything).Return(ids, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.SearchImage(queries.CreateImageQuery())
	assert.Nil(t, err, "error occured when it shouldn't have when retrieving imageID from mockdb")
	assert.Equal(t, result, ids)
}

func TestSearchImageSomeReturnedIDs(t *testing.T) { 
	ids := []database.ImageId{1, 2, 3, 4, 5}
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImages(mock.Anything).Return(ids, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.SearchImage(queries.CreateImageQuery())
	assert.Nil(t, err, "error occured when it shouldn't have when retrieving imageID from mockdb")
	assert.Equal(t, result, ids)
}