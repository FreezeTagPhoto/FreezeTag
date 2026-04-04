package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempAlbumDatabase(t *testing.T) (AlbumDatabase, UserDatabase) {
	tmp := createTempFile(t)
	manager, err := NewDefaultManager(tmp.Name())
	require.NoError(t, err)
	return manager.AlbumDB, manager.UserDB
}

func TestCreateAndGetAlbumForOwner(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Test Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, albumID, album.Id)
	assert.Equal(t, "Test Album", album.Name)
	assert.Equal(t, owner.ID, album.OwnerId)
	assert.Equal(t, ALBUM_PRIVATE, album.AlbumPrivacy)
	assert.Equal(t, VIS_ADMIN, album.VisbilityLevel)
}

func TestRenameAlbum(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Before", owner.ID, ALBUM_PUBLIC)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "After", owner.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "After", album.Name)
}

func TestSetAlbumVisibility(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Visibility Test", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.SetAlbumVisibility(albumID, ALBUM_PUBLIC, owner.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, ALBUM_PUBLIC, album.AlbumPrivacy)
}

func TestSetUserAlbumPermissionGrantsAccess(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)
	viewer, err := userDB.AddUser("viewer", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Shared Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	_, err = albumDB.GetAlbum(albumID, viewer.ID)
	require.Error(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, viewer.ID, ALBUM_PUBLIC, owner.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, viewer.ID)
	require.NoError(t, err)
	assert.Equal(t, albumID, album.Id)
	assert.Equal(t, "Shared Album", album.Name)
	assert.Equal(t, VIS_PUBLIC, album.VisbilityLevel)
}