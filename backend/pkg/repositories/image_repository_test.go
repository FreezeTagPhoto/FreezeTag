package repositories

import (
	"fmt"
	mockDatabase "freezetag/backend/mocks/ImageDatabase"
	mockParser "freezetag/backend/mocks/Parser"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"path"

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

func TestFolderPath(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	parser := initParserCollection()
	repo := InitImageRepository(mockdb, parser, "/some/test/folder")
	assert.Equal(t, "/some/test/folder/", repo.folderPath, "folder path should have trailing slash")
	repo = InitImageRepository(mockdb, parser, "/some/test/folder/")
	assert.Equal(t, "/some/test/folder/", repo.folderPath, "folder path should a SINGLE have trailing slash")
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
	mockdb.
		EXPECT().
		GetNonOverlappingSuffix(mock.Anything).
		Return(0, nil)

	parser := initParserCollection()
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	id, err := repo.StoreImageBytes(data, "gopher.png")

	assert.NoError(t, err, "expected no error in result")
	assert.Equal(t, id, database.ImageId(42))
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
	_, err = repo.StoreImageBytes(data, "notimage.txt")

	assert.Error(t, err, "expected an error in the result")
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
	_, err = repo.StoreImageBytes(data, "notimage.txt")

	assert.Error(t, err, "expected an error in the restult")
	// not testing error message since we are not trying to test CreateThumbnail directly here
}

func TestStoreImageAddImageFails(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(-1, fmt.Errorf("mock fail"))
	mockdb.
		EXPECT().
		GetNonOverlappingSuffix(mock.Anything).
		Return(0, nil)
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	_, err = repo.StoreImageBytes(data, "gopher1.png")

	assert.Error(t, err, "expected an error in the restult")
	assert.Equal(t, "gopher1.png", "gopher1.png")
	assert.Contains(t, err.Error(), "mock fail")
}

func TestStoreImageThumbnailFailsBool(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(1, nil)
	mockdb.EXPECT().AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mockdb.
		EXPECT().
		GetNonOverlappingSuffix(mock.Anything).
		Return(0, nil)
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	_, err = repo.StoreImageBytes(data, "gopher1.png")

	assert.Error(t, err, "expected an error in the restult")
	assert.Equal(t, "gopher1.png", "gopher1.png")
	assert.Contains(t, err.Error(), "database returned false when adding thumbnail")
}

func TestStoreImageThumbnailFailsError(t *testing.T) {
	tmpDir := t.TempDir()
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImage(mock.Anything, mock.Anything).Return(1, nil)
	mockdb.EXPECT().AddImageThumbnail(mock.Anything, mock.Anything, mock.Anything).Return(true, fmt.Errorf("mock error"))
	mockdb.
		EXPECT().
		GetNonOverlappingSuffix(mock.Anything).
		Return(0, nil)
	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	_, err = repo.StoreImageBytes(data, "gopher1.png")

	assert.Error(t, err, "expected an error in the restult")
	assert.Equal(t, "gopher1.png", "gopher1.png")
	assert.Contains(t, err.Error(), "mock error")
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
	mockdb.
		EXPECT().
		GetNonOverlappingSuffix(path.Join(tmpDir, "gopher.png")).
		Return(1, nil)

	parser := initParserCollection() // we need good data here
	repo := InitImageRepository(mockdb, parser, tmpDir)

	data, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)

	_, err = repo.StoreImageBytes(data, "gopher.png")
	assert.NoError(t, err, "no error in result")
	assert.FileExists(t, path.Join(tmpDir, "gopher1.png"))
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

func TestSearchImageOrderedSomeReturnedIDs(t *testing.T) {
	ids := []database.ImageId{1, 3, 2, 4, 5}
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImagesOrder(mock.Anything, mock.Anything, mock.Anything).Return(ids, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.SearchImageOrdered(queries.CreateImageQuery(), queries.DateCreated, queries.Ascending)
	assert.NoError(t, err)
	assert.Equal(t, result, ids)
}

func TestRetrieveAllTagsSuccess(t *testing.T) {
	expected := map[string]int64{"tag1": 1, "tag2": 1}
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetAllTags().Return(expected, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.RetrieveAllTags()
	assert.Nil(t, err, "error occured when it shouldn't have when retrieving imageID from mockdb")
	assert.Equal(t, expected, result)
}

func TestRetrieveAllTagsFail(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetAllTags().Return(map[string]int64{}, fmt.Errorf("mock error"))
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	_, err := repo.RetrieveAllTags()
	assert.NotNil(t, err, "no error when there should be an error")
	assert.Equal(t, "mock error", err.Error())
}

func TestRetrieveImageTagsSuccess(t *testing.T) {
	expected := []string{"tag1", "tag2"}
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImageTags(mock.Anything).Return(expected, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.RetrieveImageTags(1)
	assert.Nil(t, err, "error occured when it shouldn't have when retrieving imageID from mockdb")
	assert.Equal(t, expected, result)
}

func TestRetrievImageTagsFail(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().GetImageTags(mock.Anything).Return([]string{}, fmt.Errorf("mock error"))
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	_, err := repo.RetrieveImageTags(1)
	assert.NotNil(t, err, "no error when there should be an error")
	assert.Equal(t, "mock error", err.Error())
}

func TestAddImageTagsSuccess(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImageTags(mock.Anything, mock.Anything).Return(1, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result := repo.AddImageTags(1, []string{"tag"})
	assert.Nil(t, result.Err, "error should be nil")
	assert.NotNil(t, result.Success, "success should not be nil")
	assert.Equal(t, ImageTagSuccess{Id: 1, Count: 1}, *result.Success)
}

func TestAddImageTagsFail(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().AddImageTags(mock.Anything, mock.Anything).Return(1, fmt.Errorf("mock error"))
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result := repo.AddImageTags(1, []string{"tag"})
	assert.NotNil(t, result.Err, "error should be nil")
	assert.Nil(t, result.Success, "success should be nil")
	assert.Equal(t, ImageTagFail{Reason: "mock error", Id: 1}, *result.Err)
}

func TestRemoveImageTagsSuccess(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().RemoveImageTags(mock.Anything, mock.Anything).Return(1, nil)
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result := repo.RemoveImageTags(1, []string{"tag"})
	assert.Nil(t, result.Err, "error should be nil")
	assert.NotNil(t, result.Success, "success should not be nil")
	assert.Equal(t, ImageTagSuccess{Id: 1, Count: 1}, *result.Success)
}

func TestRemoveImageTagsFail(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().RemoveImageTags(mock.Anything, mock.Anything).Return(1, fmt.Errorf("mock error"))
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result := repo.RemoveImageTags(1, []string{"tag"})
	assert.NotNil(t, result.Err, "error should be nil")
	assert.Nil(t, result.Success, "success should be nil")
	assert.Equal(t, ImageTagFail{Reason: "mock error", Id: 1}, *result.Err)
}

func TestGetImageFilepathSuccess(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	s := "filename"
	mockdb.EXPECT().
		GetImageFile(mock.Anything).
		Return(&s, nil)

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	result, err := repo.GetImageFilepath(1)
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, result, "result should be valid")
	assert.Equal(t, "/"+s, result)
}

func TestGetImageFilepathNilString(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)

	mockdb.EXPECT().
		GetImageFile(mock.Anything).
		Return(nil, nil)

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	_, err := repo.GetImageFilepath(1)
	assert.NotNil(t, err, "result should be valid")
	assert.Equal(t, err.Error(), "nil or empty file")
}

func TestGetImageFilepathError(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().
		GetImageFile(mock.Anything).
		Return(nil, fmt.Errorf("mock error"))

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "")

	_, err := repo.GetImageFilepath(1)
	assert.NotNil(t, err, "result should be valid")
	assert.Equal(t, err.Error(), "mock error")
}

func ptrFloat64(f float64) *float64 {
	return &f
}
func ptrString(s string) *string {
	return &s
}
func ptrInt64(i int64) *int64 {
	return &i
}

func TestGetImageMetadataSuccess(tt *testing.T) {
	expected := imagedata.Metadata{
		FileName:     ptrString("mockfile.jpg"),
		DateTaken:    ptrInt64(0),
		DateUploaded: ptrInt64(0),
		CameraMake:   ptrString("mockCamera"),
		CameraModel:  ptrString("mockModel"),
		Latitude:     ptrFloat64(67.41),
		Longitude:    ptrFloat64(-9.1021),
	}

	mockdb := mockDatabase.NewMockImageDatabase(tt)
	mockdb.EXPECT().
		GetImageMetadata(mock.Anything).
		Return(expected, nil)

	parser := mockParser.NewMockParser(tt)
	repo := InitImageRepository(mockdb, parser, "/this/is/a/folder/")

	result, err := repo.GetImageMetadata(1)
	assert.Nil(tt, err, "error should be nil")
	assert.Equal(tt, expected, result)
}

func TestGetImageMetadataMissingFields(t *testing.T) {
	expected := imagedata.Metadata{
		FileName:     nil,
		DateTaken:    ptrInt64(0),
		DateUploaded: ptrInt64(0),
		CameraMake:   nil,
		CameraModel:  nil,
		Latitude:     ptrFloat64(67.41),
		Longitude:    ptrFloat64(-9.1021),
	}

	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().
		GetImageMetadata(mock.Anything).
		Return(expected, nil)

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "/this/is/a/folder/")

	result, err := repo.GetImageMetadata(1)
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, expected, result)
}

func TestGetImageMetadataFail9(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().
		GetImageMetadata(mock.Anything).
		Return(imagedata.Metadata{}, fmt.Errorf("mock error"))
	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "/this/is/a/folder/")

	_, err := repo.GetImageMetadata(1)
	assert.NotNil(t, err, "error should not be nil")
	assert.Equal(t, "mock error", err.Error())
}

func TestGetTagCount(t *testing.T) {
	expected := map[string]int64{
		"tag1": 5,
		"tag2": 10,
	}

	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().
		GetTagCounts(mock.Anything).
		Return(expected, nil)

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "/this/is/a/folder/")

	result, err := repo.GetTagCounts([]string{"tag1", "tag2"})
	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, expected, result)
}

func TestGetTagCountFail(t *testing.T) {
	mockdb := mockDatabase.NewMockImageDatabase(t)
	mockdb.EXPECT().
		GetTagCounts(mock.Anything).
		Return(nil, fmt.Errorf("mock error"))

	parser := mockParser.NewMockParser(t)
	repo := InitImageRepository(mockdb, parser, "/this/is/a/folder/")

	_, err := repo.GetTagCounts([]string{"tag1", "tag2"})
	assert.NotNil(t, err, "error should not be nil")
	assert.Equal(t, "mock error", err.Error())
}
