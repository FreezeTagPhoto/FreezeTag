package database

import (
	"crypto/rand"
	"fmt"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempFile(t *testing.T) *os.File {
	tmp, err := os.CreateTemp("", "*.db")
	require.NoError(t, err)
	return tmp
}

func createTempDatabase(t *testing.T) ImageDatabase {
	tmp := createTempFile(t)
	db, err := InitSQLiteImageDatabase(tmp.Name())
	require.NoError(t, err)
	return db
}

func TestOpenDatabase(t *testing.T) {
	tmp := createTempFile(t)
	_, err := InitSQLiteImageDatabase(tmp.Name())
	assert.NoError(t, err)
}

func TestOpenDatabaseBrokenFile(t *testing.T) {
	tmp := createTempFile(t)
	// fill it with garbage
	reader := io.LimitReader(rand.Reader, 1024)
	_, err := io.Copy(tmp, reader)
	require.NoError(t, err)
	// try to open it and it will fail integrity
	_, err = InitSQLiteImageDatabase(tmp.Name())
	assert.Error(t, err)
}

func TestAddAndRetrieveImages(t *testing.T) {
	tmp := createTempDatabase(t)
	id, err := tmp.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	assert.NoError(t, err)
	assert.NotZero(t, id)
	testTime := time.Now()
	id2, err := tmp.AddImage("bar.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       480,
		Height:      320,
		DateCreated: &testTime,
		Geo:         nil,
		Cam:         nil,
	})
	assert.NoError(t, err)
	assert.NotZero(t, id2)
	assert.NotEqual(t, id, id2)
	name, err := tmp.GetImageFile(id)
	assert.NoError(t, err)
	assert.NotNil(t, name)
	assert.Equal(t, "foo.png", *name)
}

func TestRetrieveImagesSortedByDateCreated(t *testing.T) {
	tmp := createTempDatabase(t)
	today := time.Now()
	yesterday := today.Add(-24 * time.Hour)
	halfway := today.Add(-12 * time.Hour)
	id1, err := tmp.AddImage("foo.png", imagedata.Data{
		Width:       1280,
		Height:      720,
		DateCreated: &yesterday,
	})
	assert.NoError(t, err)
	assert.NotZero(t, id1)
	id2, err := tmp.AddImage("bar.png", imagedata.Data{
		Width:       480,
		Height:      320,
		DateCreated: &today,
	})
	assert.NoError(t, err)
	assert.NotZero(t, id2)
	id3, err := tmp.AddImage("baz.png", imagedata.Data{
		Width:       720,
		Height:      1280,
		DateCreated: &halfway,
	})
	assert.NoError(t, err)
	assert.NotZero(t, id3)
	ids, err := tmp.GetImagesOrder(queries.CreateImageQuery(), queries.DateCreated, queries.Descending)
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{id2, id3, id1}, ids)
}

func TestRetrieveNoImage(t *testing.T) {
	tmp := createTempDatabase(t)
	name, err := tmp.GetImageFile(ImageId(1))
	assert.NoError(t, err)
	assert.Nil(t, name)
}

func TestRetrieveAllImages(t *testing.T) {
	tmp := createTempDatabase(t)
	_, err := tmp.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	_, err = tmp.AddImage("bar.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       480,
		Height:      320,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	ids, err := tmp.GetImages(queries.CreateImageQuery())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ids))
	for _, id := range ids {
		assert.NotZero(t, id)
	}
}

func TestRetrieveImageWithMakeAndModel(t *testing.T) {
	tmp := createTempDatabase(t)
	idNikon, err := tmp.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam: &struct {
			Manufacturer string
			Model        string
		}{
			Manufacturer: "Nikon",
			Model:        "D7500 DSLR",
		},
	})
	require.NoError(t, err)
	idCanon, err := tmp.AddImage("bar.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       480,
		Height:      320,
		DateCreated: nil,
		Geo:         nil,
		Cam: &struct {
			Manufacturer string
			Model        string
		}{
			Manufacturer: "Canon",
			Model:        "EOS R6 Mark II",
		},
	})
	require.NoError(t, err)
	ids, err := tmp.GetImages(queries.CreateImageQuery().WithMake("Nikon"))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idNikon}, ids)
	ids, err = tmp.GetImages(queries.CreateImageQuery().WithModelLike("EOS"))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idCanon}, ids)
}

func TestRetrieveImageByGeoDegrees(t *testing.T) {
	tmp := createTempDatabase(t)
	idA, err := tmp.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo: &struct {
			Lat float64
			Lon float64
			Alt float64
		}{
			Lat: 0.0,
			Lon: 0.0,
			Alt: 0.0,
		},
		Cam: nil,
	})
	require.NoError(t, err)
	idB, err := tmp.AddImage("bar.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo: &struct {
			Lat float64
			Lon float64
			Alt float64
		}{
			Lat: 12.0,
			Lon: 4.0,
			Alt: 0.0,
		},
		Cam: nil,
	})
	_ = insertTestImage(t, tmp)
	require.NoError(t, err)
	ids, err := tmp.GetImages(queries.CreateImageQuery().WithLocation(0., 0., 1.))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idA}, ids)
	ids, err = tmp.GetImages(queries.CreateImageQuery().WithLocation(11.0, 4.0, 1.1))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idB}, ids)
}

func TestRetrieveImageByDateRange(t *testing.T) {
	tmp := createTempDatabase(t)
	now := time.Now()
	then := time.Now().Add(-24 * time.Hour)
	beforeThen := then.Add(-1 * time.Hour)
	afterThen := then.Add(1 * time.Hour)
	idA, err := tmp.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: &then,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	idB, err := tmp.AddImage("bar.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: &now,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	ids, err := tmp.GetImages(queries.CreateImageQuery().TakenAfter(beforeThen).TakenBefore(afterThen))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idA}, ids)
	ids, err = tmp.GetImages(queries.CreateImageQuery().TakenAfter(afterThen))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idB}, ids)
}

func insertTestImage(t *testing.T, db ImageDatabase) ImageId {
	id, err := db.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	return id
}

func insertTestImageNamed(t *testing.T, db ImageDatabase, name string) ImageId {
	id, err := db.AddImage(name, imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	return id
}

func TestAddAndRetrieveThumbnails(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	success, err := tmp.AddImageThumbnail(id, 1, []byte("foo"))
	assert.NoError(t, err)
	assert.True(t, success)
	success, err = tmp.AddImageThumbnail(id, 2, []byte("bar"))
	assert.NoError(t, err)
	assert.True(t, success)
	sizes, err := tmp.GetImageThumbnailSizes(id)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 2}, sizes)
	data, err := tmp.GetImageThumbnail(id, 1)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), data)
}

func TestRetrieveNoThumbnail(t *testing.T) {
	tmp := createTempDatabase(t)
	data, err := tmp.GetImageThumbnail(ImageId(1), 1)
	assert.NoError(t, err)
	assert.Zero(t, data)
}

func TestRemoveImageThumbnail(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	_, _ = tmp.AddImageThumbnail(id, 1, []byte("foo"))
	_, _ = tmp.AddImageThumbnail(id, 2, []byte("bar"))
	success, err := tmp.RemoveImageThumbnail(id, 1)
	assert.NoError(t, err)
	assert.True(t, success)
	success, err = tmp.RemoveImageThumbnail(id, 3)
	assert.NoError(t, err)
	assert.False(t, success)
}

func TestAddDuplicateThumbnail(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	success, err := tmp.AddImageThumbnail(id, 1, []byte("foo"))
	assert.NoError(t, err)
	assert.True(t, success)
	success, err = tmp.AddImageThumbnail(id, 1, []byte("bar"))
	assert.NoError(t, err)
	assert.False(t, success)
}

func TestAddGetImageTags(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	num, err := tmp.AddImageTags(id, []string{"foo", "bar", "baz"})
	assert.NoError(t, err)
	assert.Equal(t, 3, num)
	tags, err := tmp.GetImageTags(id)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"foo", "bar", "baz"}, tags)
}

func TestGetAllTags(t *testing.T) {
	tmp := createTempDatabase(t)

	id := insertTestImageNamed(t, tmp, "foo.png")
	num, err := tmp.AddImageTags(id, []string{"foo", "bar", "baz"})
	assert.NoError(t, err)
	assert.Equal(t, 3, num)

	id = insertTestImageNamed(t, tmp, "bar.jpeg")
	num, err = tmp.AddImageTags(id, []string{"baz", "bat"})
	assert.NoError(t, err)
	assert.Equal(t, 2, num)

	tagCounts, err := tmp.GetAllTags()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int64{"foo": 1, "bar": 1, "baz": 2, "bat": 1}, tagCounts)
}

func TestDeleteImage(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(id, []string{"foo", "bar"})
	_, _ = tmp.AddImageThumbnail(id, 1, []byte("baz"))
	success, err := tmp.RemoveImage(id)
	assert.NoError(t, err)
	assert.True(t, success)
	tags, _ := tmp.GetImageTags(id)
	assert.Empty(t, tags)
	data, _ := tmp.GetImageThumbnail(id, 1)
	assert.Empty(t, data)
	success, err = tmp.RemoveImage(ImageId(3))
	assert.NoError(t, err)
	assert.False(t, success)
}

func TestRemoveImageTags(t *testing.T) {
	tmp := createTempDatabase(t)
	id := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(id, []string{"foo", "bar", "baz", "snap", "crackle", "pop"})
	tags, _ := tmp.GetImageTags(id)
	assert.ElementsMatch(t, []string{"foo", "bar", "baz", "snap", "crackle", "pop"}, tags)
	count, err := tmp.RemoveImageTags(id, []string{"foo", "pop", "notatag"})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	tags, _ = tmp.GetImageTags(id)
	assert.ElementsMatch(t, []string{"bar", "baz", "snap", "crackle"}, tags)
}

func TestAddTags(t *testing.T) {
	tmp := createTempDatabase(t)
	ids, err := tmp.AddTags([]string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, 3, len(ids))
	ids, err = tmp.AddTags([]string{"c", "d", "e"})
	require.NoError(t, err)
	assert.Equal(t, 2, len(ids))
}

func TestRemoveTags(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	idB := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(idA, []string{"a", "b", "c"})
	_, _ = tmp.AddImageTags(idB, []string{"b", "c", "d"})
	del, err := tmp.RemoveTags([]string{"b", "notatag"})
	require.NoError(t, err)
	assert.Equal(t, 1, del)
	tags, err := tmp.GetImageTags(idA)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "c"}, tags)
	tags, err = tmp.GetImageTags(idB)
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "d"}, tags)

	tagCounts, err := tmp.GetAllTags()
	require.NoError(t, err)
	assert.Equal(t, map[string]int64{"a": 1, "c": 2, "d": 1}, tagCounts)
}

func TestGetMetadata(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	data, err := tmp.GetImageMetadata(idA)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Nil(t, data.DateTaken)
	assert.NotNil(t, data.DateUploaded)
	assert.Nil(t, data.CameraMake)
	assert.Nil(t, data.CameraModel)
	assert.Nil(t, data.Latitude)
	assert.Nil(t, data.Longitude)
}

func TestGetMetadataNoId(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	idB := idA + 1
	data, err := tmp.GetImageMetadata(idB)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, imagedata.Metadata{}, data)
}

func TestGetMetadataPopulated(t *testing.T) {
	tmp := createTempDatabase(t)
	testTime := time.Now()
	lat := 12.34
	lon := 56.78
	idA, err := tmp.AddImage("meta.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       800,
		Height:      600,
		DateCreated: &testTime,
		Geo: &struct {
			Lat float64
			Lon float64
			Alt float64
		}{
			Lat: lat,
			Lon: lon,
			Alt: 0,
		},
	})
	require.NoError(t, err)
	data, err := tmp.GetImageMetadata(idA)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, testTime.Unix(), *data.DateTaken)
	assert.NotNil(t, data.DateUploaded)
	assert.Nil(t, data.CameraMake)
	assert.Nil(t, data.CameraModel)
	assert.Equal(t, &lat, data.Latitude)
	assert.Equal(t, &lon, data.Longitude)
}

func TestGetResolution(t *testing.T) {
	tmp := createTempDatabase(t)
	idA, err := tmp.AddImage("foo.png", imagedata.Data{
		Width:  800,
		Height: 600,
	})
	require.NoError(t, err)
	w, h, err := tmp.GetImageResolution(idA)
	assert.NoError(t, err)
	assert.Equal(t, 800, w)
	assert.Equal(t, 600, h)
}

func TestGetResolutionNoImage(t *testing.T) {
	tmp := createTempDatabase(t)
	w, h, err := tmp.GetImageResolution(1)
	assert.NoError(t, err)
	assert.Zero(t, w)
	assert.Zero(t, h)
}

func TestGetTagCounts(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	idB := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(idA, []string{"tag1", "tag2", "tag3"})
	_, _ = tmp.AddImageTags(idB, []string{"tag2", "tag3", "tag4"})
	counts, err := tmp.GetTagCounts([]string{fmt.Sprint(idA), fmt.Sprint(idB)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts["tag1"])
	assert.Equal(t, int64(2), counts["tag2"])
	assert.Equal(t, int64(2), counts["tag3"])
	assert.Equal(t, int64(1), counts["tag4"])
}

func TestGetTagCountsNoTags(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	idB := insertTestImage(t, tmp)
	counts, err := tmp.GetTagCounts([]string{fmt.Sprint(idA), fmt.Sprint(idB)})
	require.NoError(t, err)
	assert.Empty(t, counts)
}

func TestGetTagCountsNoIds(t *testing.T) {
	tmp := createTempDatabase(t)
	counts, err := tmp.GetTagCounts([]string{})
	require.NoError(t, err)
	assert.Empty(t, counts)
}

func TestGetTagCountsSomeIds(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(idA, []string{"tag1", "tag2"})
	counts, err := tmp.GetTagCounts([]string{fmt.Sprint(idA), "9999"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts["tag1"])
	assert.Equal(t, int64(1), counts["tag2"])
}

func TestGetTagCountsDuplicateTagsSingleId(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(idA, []string{"tag1", "tag1", "tag2"})
	counts, err := tmp.GetTagCounts([]string{fmt.Sprint(idA)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts["tag1"])
	assert.Equal(t, int64(1), counts["tag2"])
}

func TestMoreDuplicates(t *testing.T) {
	tmp := createTempDatabase(t)
	idA := insertTestImage(t, tmp)
	idB := insertTestImage(t, tmp)
	idC := insertTestImage(t, tmp)
	_, _ = tmp.AddImageTags(idA, []string{"A", "B", "C"})
	_, _ = tmp.AddImageTags(idB, []string{"A", "B", "C"})
	_, _ = tmp.AddImageTags(idC, []string{"A", "C", "D"})
	counts, err := tmp.GetTagCounts([]string{fmt.Sprint(idA), fmt.Sprint(idB), fmt.Sprint(idC)})
	require.NoError(t, err)
	assert.Equal(t, map[string]int64{
		"A": 3,
		"B": 2,
		"C": 3,
		"D": 1,
	}, counts)
}

func TestGetImageFile(t *testing.T) {
	tmp := createTempDatabase(t)
	idA, err := tmp.AddImage("abc", imagedata.Data{})
	require.NoError(t, err)
	name, err := tmp.GetImageFile(idA)
	assert.NoError(t, err)
	require.NotNil(t, name)
	assert.Equal(t, "abc", *name)
}

func TestGetImageNameCollisionNoImages(t *testing.T) {
	tmp := createTempDatabase(t)
	suffix, err := tmp.GetNonOverlappingSuffix("abc")
	assert.NoError(t, err)
	assert.Equal(t, 0, suffix)
}

func TestGetImageNameCollisionOnce(t *testing.T) {
	tmp := createTempDatabase(t)
	_, err := tmp.AddImage("abc", imagedata.Data{})
	require.NoError(t, err)
	suffix, err := tmp.GetNonOverlappingSuffix("abc")
	assert.NoError(t, err)
	assert.Equal(t, 1, suffix)
}

func TestGetImageNameCollisionSeveral(t *testing.T) {
	tmp := createTempDatabase(t)
	_, err := tmp.AddImage("abc", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc1", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc2", imagedata.Data{})
	require.NoError(t, err)
	suffix, err := tmp.GetNonOverlappingSuffix("abc")
	assert.NoError(t, err)
	assert.Equal(t, 3, suffix)
}

func TestGetImageNameCollisionGap(t *testing.T) {
	tmp := createTempDatabase(t)
	_, err := tmp.AddImage("abc", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc1", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc3", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc4", imagedata.Data{})
	require.NoError(t, err)
	suffix, err := tmp.GetNonOverlappingSuffix("abc")
	assert.NoError(t, err)
	assert.Equal(t, 2, suffix)
}

func TestGetImageNameCollisionGap2(t *testing.T) {
	tmp := createTempDatabase(t)
	_, err := tmp.AddImage("abc1", imagedata.Data{})
	require.NoError(t, err)
	_, err = tmp.AddImage("abc3", imagedata.Data{})
	require.NoError(t, err)
	suffix, err := tmp.GetNonOverlappingSuffix("abc")
	assert.NoError(t, err)
	assert.Equal(t, 0, suffix)
}
