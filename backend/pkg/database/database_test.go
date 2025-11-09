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

func createTempDatabase(t *testing.T) SqliteImageDatabase {
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

func TestAddImages(t *testing.T) {
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
	ids, err := tmp.GetImages(queries.CreateImageQuery().TakenBefore(then))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idA}, ids)
	ids, err = tmp.GetImages(queries.CreateImageQuery().TakenAfter(then))
	assert.NoError(t, err)
	assert.Equal(t, []ImageId{idB}, ids)
}
