package database

import (
	"fmt"
	"freezetag/backend/pkg/images/imagedata"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, userDB UserDatabase) (UserID, error) {
	hashedPassword := "hashedpassword"
	user, err := userDB.AddUser(fmt.Sprintf("user_%s", t.Name()), hashedPassword)
	return user.ID, err
}

func createNamedTestUser(t *testing.T, userDB UserDatabase, username string) UserID {
	t.Helper()
	user, err := userDB.AddUser(username, "hashedpassword")
	require.NoError(t, err)
	return user.ID
}

func createTempAlbumDatabase(t *testing.T) *Manager {
	tmp := createTempFile(t)
	mgr, err := NewDefaultManager(tmp.Name())
	require.NoError(t, err)
	return mgr
}

func TestGetAlbumPrivate(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_private")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_private")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	albumID, err := mgr.AlbumDB.CreateAlbum("shared-private", owner, PrivacyLevel(0))
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetUserAlbumPermission(albumID, viewer, PrivacyLevel(1), owner))

	ids, err := mgr.AlbumDB.GetAlbumIds(viewer)
	require.NoError(t, err)
	assert.Contains(t, ids, albumID)
}

func TestFailImageIsPrivate(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_no_share")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_no_share")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	albumID, err := mgr.AlbumDB.CreateAlbum("owner-private", owner, PrivacyLevel(0))
	require.NoError(t, err)
	imageID, err := mgr.ImageDB.AddImage("private.png", imagedata.Data{Width: 10, Height: 10})
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetImageAlbum(imageID, albumID, owner))

	_, err = mgr.AlbumDB.GetAlbumImages(albumID, viewer)
	require.Error(t, err)
}

func TestGetAlbumPublic(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_public")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_public")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 1))

	albumID, err := mgr.AlbumDB.CreateAlbum("public-album", owner, PrivacyLevel(1))
	require.NoError(t, err)

	ids, err := mgr.AlbumDB.GetAlbumIds(viewer)
	require.NoError(t, err)
	assert.Contains(t, ids, albumID)
}

func TestGetAlbumNotExist(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	user := createNamedTestUser(t, mgr.UserDB, "lookup_user")

	id, err := mgr.AlbumDB.GetAlbumIdByName("missing-album", user)
	require.NoError(t, err)
	assert.Equal(t, AlbumId(0), id)
}

func TestGetAlbumImages(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_images")

	albumID, err := mgr.AlbumDB.CreateAlbum("owner-album", owner, PrivacyLevel(0))
	require.NoError(t, err)
	imageID, err := mgr.ImageDB.AddImage("owner.png", imagedata.Data{Width: 20, Height: 20})
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetImageAlbum(imageID, albumID, owner))

	images, err := mgr.AlbumDB.GetAlbumImages(albumID, owner)
	require.NoError(t, err)
	assert.Equal(t, []ImageId{imageID}, images)
}

func TestGetAlbumImagesPrivate(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_private_images")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_private_images")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	albumID, err := mgr.AlbumDB.CreateAlbum("shared-images", owner, PrivacyLevel(0))
	require.NoError(t, err)
	imageID, err := mgr.ImageDB.AddImage("shared.png", imagedata.Data{Width: 30, Height: 30})
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetImageAlbum(imageID, albumID, owner))
	require.NoError(t, mgr.AlbumDB.SetUserAlbumPermission(albumID, viewer, PrivacyLevel(1), owner))

	images, err := mgr.AlbumDB.GetAlbumImages(albumID, viewer)
	require.NoError(t, err)
	assert.Equal(t, []ImageId{imageID}, images)
}

func TestGetAlbumImagesPublic(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_public_images")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_public_images")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 1))

	albumID, err := mgr.AlbumDB.CreateAlbum("public-images", owner, PrivacyLevel(1))
	require.NoError(t, err)
	imageID, err := mgr.ImageDB.AddImage("public.png", imagedata.Data{Width: 40, Height: 40})
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetImageAlbum(imageID, albumID, owner))

	images, err := mgr.AlbumDB.GetAlbumImages(albumID, viewer)
	require.NoError(t, err)
	assert.Equal(t, []ImageId{imageID}, images)
}

func TestGetAlbumImagesNotExist(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	user := createNamedTestUser(t, mgr.UserDB, "viewer_missing_album")

	_, err := mgr.AlbumDB.GetAlbumImages(AlbumId(99999), user)
	require.Error(t, err)
}

func TestGetAlbumIdByNamePrivateExplicitAccess(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_lookup_private_grant")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_lookup_private_grant")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	albumID, err := mgr.AlbumDB.CreateAlbum("lookup-private-grant", owner, PrivacyLevel(0))
	require.NoError(t, err)
	require.NoError(t, mgr.AlbumDB.SetUserAlbumPermission(albumID, viewer, PrivacyLevel(1), owner))

	found, err := mgr.AlbumDB.GetAlbumIdByName("lookup-private-grant", viewer)
	require.NoError(t, err)
	assert.Equal(t, albumID, found)
}

func TestGetAlbumIdByNamePrivateNoAccess(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_lookup_private_denied")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_lookup_private_denied")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	_, err := mgr.AlbumDB.CreateAlbum("lookup-private-denied", owner, PrivacyLevel(0))
	require.NoError(t, err)

	found, err := mgr.AlbumDB.GetAlbumIdByName("lookup-private-denied", viewer)
	require.NoError(t, err)
	assert.Equal(t, AlbumId(0), found)
}

func TestGetAlbumIdByNamePublicBlockedForPrivateUserMode(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_lookup_public")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_lookup_public_private_mode")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 0))

	_, err := mgr.AlbumDB.CreateAlbum("lookup-public-hidden", owner, PrivacyLevel(1))
	require.NoError(t, err)

	found, err := mgr.AlbumDB.GetAlbumIdByName("lookup-public-hidden", viewer)
	require.NoError(t, err)
	assert.Equal(t, AlbumId(0), found)
}

func TestGetAlbumIdByNamePrefersOwnedAlbumOnDuplicateName(t *testing.T) {
	mgr := createTempAlbumDatabase(t)
	owner := createNamedTestUser(t, mgr.UserDB, "owner_dup_name")
	viewer := createNamedTestUser(t, mgr.UserDB, "viewer_dup_name")
	require.NoError(t, mgr.UserDB.SetUserVisibilityMode(viewer, 1))

	ownerAlbum, err := mgr.AlbumDB.CreateAlbum("same-name", owner, PrivacyLevel(1))
	require.NoError(t, err)
	viewerAlbum, err := mgr.AlbumDB.CreateAlbum("same-name", viewer, PrivacyLevel(0))
	require.NoError(t, err)

	found, err := mgr.AlbumDB.GetAlbumIdByName("same-name", viewer)
	require.NoError(t, err)
	assert.Equal(t, viewerAlbum, found)
	assert.NotEqual(t, ownerAlbum, found)
}
