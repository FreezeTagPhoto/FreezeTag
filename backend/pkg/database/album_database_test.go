package database

import (
	// "freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/images/imagedata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, userDB UserDatabase) (UserID, error) {
	hashedPassword := "hashedpassword"
	user, err := userDB.AddUser("1", hashedPassword)	
	return user.ID, err	
}

func createTempAlbumDatabase(t *testing.T) *Manager {
	tmp := createTempFile(t)
	mgr, err := NewDefaultManager(tmp.Name())
	require.NoError(t, err)
	return mgr
}

func TestCreateAlbum(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	user, err := createTestUser(t, mgr.UserDB)
	require.NoError(t, err)
	_, err = mgr.AlbumDB.CreateAlbum("test description", user, Private)
	require.NoError(t, err)
}

func TestAddImageToAlbum(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	user, err := createTestUser(t, mgr.UserDB)
	require.NoError(t, err)
	iid, err := mgr.ImageDB.AddImage("foo.png", imagedata.Data{
		PixelsRGBA:  []byte{},
		Width:       1280,
		Height:      720,
		DateCreated: nil,
		Geo:         nil,
		Cam:         nil,
	})
	require.NoError(t, err)
	aid, err := mgr.AlbumDB.CreateAlbum("test description", user, Private)
	require.NoError(t, err)
	err = mgr.AlbumDB.SetImageAlbum(iid, aid, user)
	assert.NoError(t, err)
}

func TestGetAlbumPrivate(t *testing.T) {
}

func TestFailImageIsPrivate(t *testing.T) {
}

func TestGetAlbumPublic(t *testing.T) {
}

func TestGetAlbumNotExist(t *testing.T) {
}

func TestGetAlbumImages(t *testing.T) {
}

func TestGetAlbumImagesPrivate(t *testing.T) {
}

func TestGetAlbumImagesPublic(t *testing.T) {
}

func TestGetAlbumImagesNotExist(t *testing.T) {
}