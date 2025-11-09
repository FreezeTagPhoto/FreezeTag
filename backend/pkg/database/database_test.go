package database

import (
	"crypto/rand"
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
