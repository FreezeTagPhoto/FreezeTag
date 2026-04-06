package database

import (
	"fmt"
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

// since userAuthorizedForAlbum is NOT a part of the 
// AlbumDatabase interface, gotta do some type assertion
func assertSqliteAlbumDatabase(t *testing.T, albumDB AlbumDatabase) SqliteAlbumDatabase {
	var sqliteAlbumDB SqliteAlbumDatabase
	switch db := albumDB.(type) {
	case SqliteAlbumDatabase:
		sqliteAlbumDB = db
	case *SqliteAlbumDatabase:
		sqliteAlbumDB = *db
	default:
		t.Fatalf("unexpected AlbumDatabase implementation: %T", albumDB)
	}
	return sqliteAlbumDB
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

func TestGetVisibilityModeUID0(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)
	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Visibility Mode Test", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, 0) // UID 0 should see everything
	require.NoError(t, err)
	assert.Equal(t, VIS_ADMIN, album.VisbilityLevel)
}

func TestUserAuthorizedForAlbumStressMatrix(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)
	sqliteAlbumDB := assertSqliteAlbumDatabase(t, albumDB)

	owner, err := userDB.AddUser("owner_stress", "hash")
	require.NoError(t, err)
	requester, err := userDB.AddUser("requester_stress", "hash")
	require.NoError(t, err)

	requesterModes := []UserPrivacy{VIS_PRIVATE, VIS_PUBLIC}
	albumModes := []GlobalPrivacy{ALBUM_PRIVATE, ALBUM_PUBLIC}
	accessLevels := []int{-1, 0, 1, 2} // -1 means no AlbumAccess row

	computeExpected := func(requesterMode UserPrivacy, albumMode GlobalPrivacy, accessLevel int) bool {
		switch requesterMode {
		case VIS_PUBLIC:
			globalReadable := albumMode >= ALBUM_PUBLIC && (accessLevel == -1 || accessLevel > 0)
			return globalReadable || accessLevel > 0
		case VIS_PRIVATE:
			return accessLevel > 0
		default:
			return false
		}
	}

	for _, requesterMode := range requesterModes {
		err = userDB.SetUserVisibilityMode(requester.ID, int(requesterMode))
		require.NoError(t, err)

		for _, albumMode := range albumModes {
			for _, accessLevel := range accessLevels {
				albumName := fmt.Sprintf("stress_%d_%d_%d", requesterMode, albumMode, accessLevel)
				albumID, createErr := albumDB.CreateAlbum(albumName, owner.ID, albumMode)
				require.NoError(t, createErr)

				if accessLevel >= 0 {
					permErr := albumDB.SetUserAlbumPermission(albumID, requester.ID, GlobalPrivacy(accessLevel), owner.ID)
					require.NoError(t, permErr)
				}

				got, authErr := sqliteAlbumDB.userAuthorizedForAlbum(albumID, requester.ID)
				require.NoError(t, authErr)

				expected := computeExpected(requesterMode, albumMode, accessLevel)
				assert.Equal(t, expected, got, "mode=%d album=%d access=%d", requesterMode, albumMode, accessLevel)
			}
		}
	}

	// Owner should always be authorized for their own albums.
	ownerAlbumID, err := albumDB.CreateAlbum("owner_album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	ownerAuthorized, err := sqliteAlbumDB.userAuthorizedForAlbum(ownerAlbumID, owner.ID)
	require.NoError(t, err)
	assert.True(t, ownerAuthorized)
}



func TestUserAuthorizedForAlbumStressAdminUserIDZero(t *testing.T) {
	albumDB, userDB := createTempAlbumDatabase(t)

	sqliteAlbumDB := assertSqliteAlbumDatabase(t, albumDB)

	owner, err := userDB.AddUser("owner_admin", "hash")
	require.NoError(t, err)
	albumID, err := albumDB.CreateAlbum("admin_can_access", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	authorized, err := sqliteAlbumDB.userAuthorizedForAlbum(albumID, 0)
	require.NoError(t, err)
	assert.True(t, authorized)
}
