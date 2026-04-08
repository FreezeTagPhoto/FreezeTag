package database

import (
	"fmt"
	"freezetag/backend/pkg/images/imagedata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempManager(t *testing.T) *Manager {
	tmp := createTempFile(t)
	manager, err := NewDefaultManager(tmp.Name())
	require.NoError(t, err)
	return manager
}

// since userAuthorizedForAlbum is NOT a part of the
// AlbumDatabase interface, we gotta do some type assertion
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

func TestFailsCanManageAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_manage_fail", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_manage_fail", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Manage Fail Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "Hacked Name", other.ID)
	assert.Error(t, err)

	err = albumDB.SetAlbumVisibility(albumID, ALBUM_PUBLIC, other.ID)
	assert.Error(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, other.ID, USER_PUBLIC, other.ID)
	assert.Error(t, err)

	err = albumDB.RemoveAlbum(albumID, other.ID)
	assert.Error(t, err)

	err = albumDB.SetImageAlbum(1, albumID, other.ID)
	assert.Error(t, err)

	err = albumDB.RemoveImageFromAlbum(1, albumID, other.ID)
	assert.Error(t, err)
}

func TestCreateAndGetAlbumForOwner(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Test Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, albumID, album.ID)
	assert.Equal(t, "Test Album", album.Name)
	assert.Equal(t, owner.ID, album.OwnerID)
	assert.Equal(t, ALBUM_PRIVATE, album.AlbumPrivacy)
	assert.Equal(t, USER_ADMIN, album.VisbilityLevel)
}

func TestRenameAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

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
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

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
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)
	viewer, err := userDB.AddUser("viewer", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Shared Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	_, err = albumDB.GetAlbum(albumID, viewer.ID)
	require.Error(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, viewer.ID, USER_PUBLIC, owner.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, viewer.ID)
	require.NoError(t, err)
	assert.Equal(t, albumID, album.ID)
	assert.Equal(t, "Shared Album", album.Name)
	assert.Equal(t, USER_PUBLIC, album.VisbilityLevel)
}

func TestGetVisibilityModeUID0(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	owner, err := userDB.AddUser("owner", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Visibility Mode Test", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, 0) // UID 0 should see everything
	require.NoError(t, err)
	assert.Equal(t, USER_ADMIN, album.VisbilityLevel)
}

func TestUserAuthorizedForAlbumStressMatrix(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	sqliteAlbumDB := assertSqliteAlbumDatabase(t, albumDB)

	owner, err := userDB.AddUser("owner_stress", "hash")
	require.NoError(t, err)
	requester, err := userDB.AddUser("requester_stress", "hash")
	require.NoError(t, err)

	requesterModes := []UserPrivacy{USER_PRIVATE, USER_PUBLIC}
	albumModes := []GlobalPrivacy{ALBUM_PRIVATE, ALBUM_PUBLIC}
	accessLevels := []int{-1, 0, 1, 2} // -1 means no AlbumAccess row

	computeExpected := func(requesterMode UserPrivacy, albumMode GlobalPrivacy, accessLevel int) bool {
		switch requesterMode {
		case USER_PUBLIC:
			globalReadable := albumMode >= ALBUM_PUBLIC && (accessLevel == -1 || accessLevel > 0)
			return globalReadable || accessLevel > 0
		case USER_PRIVATE:
			return accessLevel > 0
		default:
			return false
		}
	}

	for _, requesterMode := range requesterModes {
		err = userDB.SetUserVisibilityMode(requester.ID, requesterMode)
		require.NoError(t, err)

		for _, albumMode := range albumModes {
			for _, accessLevel := range accessLevels {
				albumName := fmt.Sprintf("stress_%d_%d_%d", requesterMode, albumMode, accessLevel)
				albumID, createErr := albumDB.CreateAlbum(albumName, owner.ID, albumMode)
				require.NoError(t, createErr)

				if accessLevel >= 0 {
					permErr := albumDB.SetUserAlbumPermission(albumID, requester.ID, UserPrivacy(accessLevel), owner.ID)
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
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	sqliteAlbumDB := assertSqliteAlbumDatabase(t, albumDB)

	owner, err := userDB.AddUser("owner_admin", "hash")
	require.NoError(t, err)
	albumID, err := albumDB.CreateAlbum("admin_can_access", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	authorized, err := sqliteAlbumDB.userAuthorizedForAlbum(albumID, 0)
	require.NoError(t, err)
	assert.True(t, authorized)
}

func TestSetImageAlbumSuccess(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_set_image", "hash")
	require.NoError(t, err)
	albumID, err := albumDB.CreateAlbum("Set Image Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	imageID, err := imageDB.AddImage("test.png", imagedata.Data{})
	require.NoError(t, err)
	err = albumDB.SetImageAlbum(imageID, albumID, owner.ID)
	require.NoError(t, err)
}

func TestSetImageAlbumCantManage(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_no_manage", "hash")
	require.NoError(t, err)
	requester, err := userDB.AddUser("requester_no_manage", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("No Manage Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	imageID, err := imageDB.AddImage("test.png", imagedata.Data{})
	require.NoError(t, err)

	err = albumDB.SetImageAlbum(imageID, albumID, requester.ID)
	assert.Error(t, err)
}

func TestRenameAlbumNotOwner(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_rename", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_rename", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Original", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "Hijacked", other.ID)
	require.Error(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "Original", album.Name)
}

func TestSetAlbumVisibilityNotOwnerFails(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_visibility", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_visibility", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Visibility Owner", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.SetAlbumVisibility(albumID, ALBUM_PUBLIC, other.ID)
	require.Error(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, ALBUM_PRIVATE, album.AlbumPrivacy)
}

func TestSetUserAlbumPermissionCannotManage(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_perm", "hash")
	require.NoError(t, err)
	target, err := userDB.AddUser("target_perm", "hash")
	require.NoError(t, err)
	requester, err := userDB.AddUser("requester_perm", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("No Manage Perm", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, target.ID, USER_PUBLIC, requester.ID)
	require.Error(t, err)
}

func TestGetAlbumForbiddenForPrivateAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_get_forbidden", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_get_forbidden", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Private Get", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	_, err = albumDB.GetAlbum(albumID, other.ID)
	require.Error(t, err)
}

func TestGetAlbumImagesForbiddenForPrivateAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_images_forbidden", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_images_forbidden", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Private Images", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	imageID, err := imageDB.AddImage("private.png", imagedata.Data{})
	require.NoError(t, err)
	require.NoError(t, albumDB.SetImageAlbum(imageID, albumID, owner.ID))

	_, err = albumDB.GetAlbumImages(albumID, other.ID)
	require.Error(t, err)
}

func TestGetAssociatedAlbumsRespectsPermissions(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_assoc", "hash")
	require.NoError(t, err)
	viewer, err := userDB.AddUser("viewer_assoc", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Assoc Private", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	imageID, err := imageDB.AddImage("assoc.png", imagedata.Data{})
	require.NoError(t, err)
	require.NoError(t, albumDB.SetImageAlbum(imageID, albumID, owner.ID))

	albums, err := albumDB.GetAssociatedAlbums(imageID, viewer.ID)
	require.NoError(t, err)
	assert.Len(t, albums, 0)

	require.NoError(t, albumDB.SetUserAlbumPermission(albumID, viewer.ID, USER_PUBLIC, owner.ID))
	albums, err = albumDB.GetAssociatedAlbums(imageID, viewer.ID)
	require.NoError(t, err)
	assert.Len(t, albums, 1)
	assert.Equal(t, albumID, albums[0].ID)
}

func TestGetAlbumSharedUsersForbidden(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_shared_forbidden", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_shared_forbidden", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Shared Forbidden", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	_, err = albumDB.GetAlbumSharedUsers(albumID, other.ID)
	require.Error(t, err)
}

func TestRemoveImageFromAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_remove_image", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Remove Image Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	imageID, err := imageDB.AddImage("toremove.png", imagedata.Data{})
	require.NoError(t, err)

	require.NoError(t, albumDB.SetImageAlbum(imageID, albumID, owner.ID))

	err = albumDB.RemoveImageFromAlbum(imageID, albumID, owner.ID)
	require.NoError(t, err)

	images, err := albumDB.GetAlbumImages(albumID, owner.ID)
	require.NoError(t, err)
	assert.Len(t, images, 0)
}

func TestRemoveImageFromAlbumUnauthorized(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB
	imageDB := manager.ImageDB

	owner, err := userDB.AddUser("owner_noremove_image", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Remove Image Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	imageID, err := imageDB.AddImage("toremove.png", imagedata.Data{})
	require.NoError(t, err)

	require.NoError(t, albumDB.SetImageAlbum(imageID, albumID, owner.ID))

	err = albumDB.RemoveImageFromAlbum(imageID, albumID, owner.ID+1)
	require.Error(t, err)

	images, err := albumDB.GetAlbumImages(albumID, owner.ID)
	require.NoError(t, err)
	assert.Len(t, images, 1)
}

func TestRemoveAlbum(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_remove_album", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_remove_album", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("To Remove", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	err = albumDB.RemoveAlbum(albumID, owner.ID)
	require.NoError(t, err)

	_, err = albumDB.GetAlbum(albumID, owner.ID)
	require.Error(t, err)

	err = albumDB.RemoveAlbum(albumID, other.ID)
	require.Error(t, err)
}

func TestRemoveAlbumOwnerHasAuthorized(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_remove_auth", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_remove_auth", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("To Remove", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, other.ID, USER_ADMIN, owner.ID)
	require.NoError(t, err)
	err = albumDB.RemoveAlbum(albumID, other.ID)
	require.NoError(t, err)
}

func TestRemoveAlbumUnauthorized(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_remove_unauth", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_remove_unauth", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("To Remove", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.RemoveAlbum(albumID, other.ID)
	require.Error(t, err)
}

func TestRenameAlbumSuccess(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_rename_success", "hash")
	require.NoError(t, err)
	albumID, err := albumDB.CreateAlbum("Original Name", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "New Name", owner.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", album.Name)
}

func TestRenameAlbumOwnerHasAuthorized(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_rename_auth", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_rename_auth", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Original Name", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.SetUserAlbumPermission(albumID, other.ID, USER_ADMIN, owner.ID)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "New Name", other.ID)
	require.NoError(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", album.Name)
}

func TestRenameAlbumUnauthorized(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_rename_unauth", "hash")
	require.NoError(t, err)
	other, err := userDB.AddUser("other_rename_unauth", "hash")
	require.NoError(t, err)

	albumID, err := albumDB.CreateAlbum("Original Name", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)

	err = albumDB.RenameAlbum(albumID, "New Name", other.ID)
	assert.Error(t, err)

	album, err := albumDB.GetAlbum(albumID, owner.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Original Name", album.Name)
}

func TestGetAlbumsAdmin(t *testing.T) {
	manager := createTempManager(t)
	albumDB := manager.AlbumDB
	userDB := manager.UserDB

	owner, err := userDB.AddUser("owner_get_admin", "hash")
	require.NoError(t, err)
	err = userDB.SetUserVisibilityMode(owner.ID, USER_ADMIN)
	require.NoError(t, err)
	albumID, err := albumDB.CreateAlbum("Admin Album", owner.ID, ALBUM_PRIVATE)
	require.NoError(t, err)
	albumID2, err := albumDB.CreateAlbum("Admin Album 2", owner.ID, ALBUM_PUBLIC)
	require.NoError(t, err)

	albums, err := albumDB.GetAlbums(owner.ID)
	require.NoError(t, err)
	assert.Len(t, albums, 2)
	assert.ElementsMatch(t,
		[]AlbumID{albumID, albumID2},
		[]AlbumID{albums[0].ID, albums[1].ID})
}
