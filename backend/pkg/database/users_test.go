package database

import (
	"freezetag/backend/pkg/database/data"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempUserDatabase(t *testing.T) UserDatabase {
	tmp := createTempFile(t)
	db, err := InitSQLiteUserDatabase(tmp.Name())
	require.NoError(t, err)
	return db
}

func TestAddUser(t *testing.T) {
	db := createTempUserDatabase(t)
	hashedPassword := "hashedpassword"
	user, err := db.AddUser("1", hashedPassword)

	require.NoError(t, err)
	assert.Equal(t, "1", user.Username)
	assert.NotNil(t, user.ID)
	assert.NotNil(t, user.CreatedAt)
	assert.Equal(t, hashedPassword, user.PasswordHash)
}

func TestAddAndGetUser(t *testing.T) {
	db := createTempUserDatabase(t)
	hashedPassword := "hashedpassword"
	_, err := db.AddUser("1", hashedPassword)
	require.NoError(t, err)
	_, err = db.AddUser("2", hashedPassword)
	require.NoError(t, err)

	user1, err := db.GetUserByUsername("1")
	require.NoError(t, err)
	require.Equal(t, "1", user1.Username)
	assert.Equal(t, user1.ID, user1.ID)
	assert.Equal(t, hashedPassword, user1.PasswordHash)

	user2, err := db.GetUserByUsername("2")
	require.NoError(t, err)
	require.Equal(t, "2", user2.Username)
	assert.Equal(t, user2.ID, user2.ID)
	assert.Equal(t, hashedPassword, user2.PasswordHash)

	assert.NotEqual(t, user1.ID, user2.ID)

}

func TestAddDuplicateUser(t *testing.T) {
	db := createTempUserDatabase(t)
	hashedPassword := "hashedpassword"
	user1, err := db.AddUser("1", hashedPassword)
	require.NoError(t, err)
	_, err = db.AddUser("1", "different password")
	require.Error(t, err)

	user, err := db.GetUserByUsername("1")
	require.NoError(t, err)
	require.Equal(t, "1", user.Username)
	assert.Equal(t, user1.ID, user.ID)
	assert.Equal(t, hashedPassword, user.PasswordHash)
}

func TestGetNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.GetUserByUsername("nonexistent")
	require.Error(t, err)
}

func TestGetPasswordHash(t *testing.T) {
	db := createTempUserDatabase(t)
	hashedPassword := "hashedpassword"
	_, err := db.AddUser("1", hashedPassword)
	require.NoError(t, err)

	user, err := db.GetUserByUsername("1")
	require.NoError(t, err)

	passwordHash, err := db.GetPasswordHash(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "hashedpassword", passwordHash)
}
func TestAddDuplicateUserNoPasswordOverwrite(t *testing.T) {
	db := createTempUserDatabase(t)
	hashedPassword := "hashedpassword"

	user1, err := db.AddUser("1", hashedPassword)
	require.NoError(t, err)
	user2, err := db.AddUser("1", "different password")
	require.Error(t, err)
	assert.Nil(t, user2)

	passwordHash, err := db.GetPasswordHash(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, hashedPassword, passwordHash)
}

func TestGetPasswordHashNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.GetPasswordHash(999)
	require.Error(t, err)
}

func TestListUsernames(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.AddUser("homer", "hash1")
	require.NoError(t, err)
	_, err = db.AddUser("bart", "hash2")
	require.NoError(t, err)

	users, err := db.ListUsers()
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.ElementsMatch(t, []string{"homer", "bart"}, []string{users[0].Username, users[1].Username})
}

func TestSetUserPassword(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.AddUser("1", "hashedpassword")
	require.NoError(t, err)

	user, err := db.GetUserByUsername("1")
	require.NoError(t, err)

	err = db.SetUserPassword(user.ID, "newhashedpassword")
	require.NoError(t, err)

	passwordHash, err := db.GetPasswordHash(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "newhashedpassword", passwordHash)
}

func TestSetUserPasswordNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)
	err := db.SetUserPassword(999, "newhashedpassword")
	require.Error(t, err) // Should error if user does not exist
}

func TestGetUserById(t *testing.T) {
	db := createTempUserDatabase(t)
	user, err := db.AddUser("John Paras", "hash1")
	require.NoError(t, err)

	byName, err := db.GetUserByUsername("John Paras")
	require.NoError(t, err)

	byId, err := db.GetUserById(user.ID)
	require.NoError(t, err)

	assert.Equal(t, byName.ID, byId.ID)
	assert.Equal(t, byName.Username, byId.Username)
}

func TestGetUserByIdNonexistent(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.GetUserById(99999)
	require.Error(t, err)
}

func TestCreatedAtSet(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.AddUser("carol sturka", "hash")
	require.NoError(t, err)

	user, err := db.GetUserByUsername("carol sturka")
	require.NoError(t, err)
	assert.True(t, user.CreatedAt > 0)
}

func TestListUsersEmpty(t *testing.T) {
	db := createTempUserDatabase(t)
	users, err := db.ListUsers()
	require.NoError(t, err)
	assert.Equal(t, 0, len(users))
}

func TestSetUserPasswordDoesNotAffectOthers(t *testing.T) {
	db := createTempUserDatabase(t)
	userA, err := db.AddUser("rickby", "hashedPassA")
	require.NoError(t, err)
	userB, err := db.AddUser("mortycai", "hashedPassB")
	require.NoError(t, err)

	err = db.SetUserPassword(userA.ID, "newA")
	require.NoError(t, err)

	hashA, err := db.GetPasswordHash(userA.ID)
	require.NoError(t, err)
	assert.Equal(t, "newA", hashA)

	hashB, err := db.GetPasswordHash(userB.ID)
	require.NoError(t, err)
	assert.Equal(t, "hashedPassB", hashB)
}

func TestGetUserPermissions(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("permtest", "hash")
	require.NoError(t, err)

	err = db.GrantUserPermissions(user.ID, data.AllPermissions())
	require.NoError(t, err)
	permissions, err := db.GetUserPermissions(user.ID)

	require.NoError(t, err)
	assert.ElementsMatch(t, data.AllPermissions(), permissions)
}

func TestGetUserPermissionsStress(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("permtest", "hash")
	require.NoError(t, err)

	err = db.GrantUserPermissions(user.ID, data.AllPermissions())
	require.NoError(t, err)
	permissions, err := db.GetUserPermissions(user.ID)

	require.NoError(t, err)
	assert.ElementsMatch(t, data.AllPermissions(), permissions)

	err = db.RevokeUserPermissions(user.ID, data.Permissions{data.ReadUser})
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.NotContains(t, permissions, data.ReadUser)
	err = db.RevokeUserPermissions(user.ID, data.AllPermissions())
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.Empty(t, permissions)
	err = db.GrantUserPermissions(user.ID, data.Permissions{data.ReadTags})
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, data.Permissions{data.ReadTags}, permissions)
	err = db.GrantUserPermissions(user.ID, data.AllPermissions())
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, data.AllPermissions(), permissions)
}

func TestAddPermissionNonSeededPermission(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("permtest", "hash")
	require.NoError(t, err)

	err = db.GrantUserPermissions(user.ID, data.Permissions{data.Permission{Slug: "nonexistent:permission"}})
	require.Error(t, err)
}

func TestDeletePermissionUserDoesNotHave(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("permtest", "hash")
	require.NoError(t, err)

	err = db.RevokeUserPermissions(user.ID, data.Permissions{data.Permission{Slug: "nonexistent:permission"}})
	require.NoError(t, err) // should not error even if user does not have the permission
}

func TestDeleteUser(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("usertodelete", "hash")
	require.NoError(t, err)

	err = db.DeleteUser(user.ID)
	require.NoError(t, err)

	_, err = db.GetUserById(user.ID)
	require.Error(t, err)
}

func TestSaveApiTokenSuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)
	tokenHash := [32]byte{1, 2, 3} // example token hash
	label := "test-token"
	id, err := db.SaveApiToken(user.ID, nil, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	retrievedInfo, err := db.GetApiTokenInfo(id)
	require.NoError(t, err)
	assert.Equal(t, label, retrievedInfo.Label)

	retrievedUserID, err := db.GetApiUserID(tokenHash)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUserID)
}

func TestSaveApiTokenNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)
	tokenHash := [32]byte{1, 2, 3} // example token hash
	label := "test-token"
	_, err := db.SaveApiToken(999, nil, tokenHash, label, data.Permissions{}) // user ID 999 does not exist
	require.Error(t, err)
}

func TestSaveApiTokensSuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash1 := [32]byte{1, 2, 3}
	label1 := "test-token-1"
	apiId1, err := db.SaveApiToken(user.ID, nil, tokenHash1, label1, data.Permissions{})
	require.NoError(t, err)

	tokenHash2 := [32]byte{4, 5, 6}
	label2 := "test-token-2"
	apiId2, err := db.SaveApiToken(user.ID, nil, tokenHash2, label2, data.Permissions{})
	require.NoError(t, err)

	retrievedInfo1, err := db.GetApiTokenInfo(apiId1)
	require.NoError(t, err)
	assert.Equal(t, label1, retrievedInfo1.Label)

	retrievedInfo2, err := db.GetApiTokenInfo(apiId2)
	require.NoError(t, err)
	assert.Equal(t, label2, retrievedInfo2.Label)

	retrievedUserID1, err := db.GetApiUserID(tokenHash1)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUserID1)

	retrievedUserID2, err := db.GetApiUserID(tokenHash2)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUserID2)

	retrievedInfo, err := db.GetUserApiTokenInfo(user.ID)
	require.NoError(t, err)
	expected := []ApiTokenInfo{
		{apiId1, label1, TokenStatusActive},
		{apiId2, label2, TokenStatusActive},
	}
	assert.ElementsMatch(t, expected, retrievedInfo)
	assert.ElementsMatch(t, expected, retrievedInfo)
}

func TestRevokeApiKeySuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	apiId, err := db.SaveApiToken(user.ID, nil, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	err = db.AdminRevokeApiToken(apiId)
	require.NoError(t, err)

	info, err := db.GetApiTokenInfo(apiId)
	require.NoError(t, err)
	assert.Equal(t, TokenStatusRevoked, info.Status)

	// revoked tokens shuold not be associated with a user
	_, err = db.GetApiUserID(tokenHash)
	require.Error(t, err)
}

func TestRevokeApiKeyNonexistentToken(t *testing.T) {
	db := createTempUserDatabase(t)

	apiId := TokenID(999) // non-existent token ID
	err := db.AdminRevokeApiToken(apiId)
	require.Error(t, err)
}

func TestGetUserApiTokenLabels(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash1 := [32]byte{1, 2, 3}
	l1 := "test-token-1"
	id1, err := db.SaveApiToken(user.ID, nil, tokenHash1, l1, data.Permissions{})
	require.NoError(t, err)

	tokenHash2 := [32]byte{4, 5, 6}
	label2 := "test-token-2"
	id2, err := db.SaveApiToken(user.ID, nil, tokenHash2, label2, data.Permissions{})
	require.NoError(t, err)

	err = db.AdminRevokeApiToken(id1)
	require.NoError(t, err)

	labels, err := db.GetUserApiTokenInfo(user.ID)
	require.NoError(t, err)
	expected := []ApiTokenInfo{{id2, label2, TokenStatusActive}, {id1, l1, TokenStatusRevoked}}
	assert.ElementsMatch(t, expected, labels)
	tokenHash2 = [32]byte{7, 8, 9}
	label3 := "test-token-3"
	id3, err := db.SaveApiToken(user.ID, nil, tokenHash2, label3, data.Permissions{})
	require.NoError(t, err)

	labels, err = db.GetUserApiTokenInfo(user.ID)
	require.NoError(t, err)
	expected = append(expected, ApiTokenInfo{id3, label3, TokenStatusActive})
	assert.ElementsMatch(t, expected, labels)
}

func TestExpiredTokenIsExpired(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	expiredTime := time.Now().Add(-time.Hour) // expired 1 hour ago
	id, err := db.SaveApiToken(user.ID, &expiredTime, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	result, err := db.GetApiTokenInfo(id)
	require.NoError(t, err)
	assert.Equal(t, TokenStatusExpired, result.Status)

	_, err = db.GetApiUserID(tokenHash)
	require.Error(t, err)
}

func TestAllUsers(t *testing.T) {
	db := createTempUserDatabase(t)
	_, err := db.AddUser("homer", "hash1")
	require.NoError(t, err)
	_, err = db.AddUser("bart", "hash2")
	require.NoError(t, err)

	users, err := db.AllUsers()
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.ElementsMatch(t, []string{"homer", "bart"}, []string{users[0].Username, users[1].Username})
}

func TestAllUsersNoUsers(t *testing.T) {
	db := createTempUserDatabase(t)
	users, err := db.AllUsers()
	require.NoError(t, err)
	assert.Equal(t, 0, len(users))
}

func TestEnsureAdmin(t *testing.T) {
	db := createTempUserDatabase(t)
	user, err := db.AddUser("username", "hash")
	require.NoError(t, err)

	err = db.EnsureAdmin(user.ID)
	require.NoError(t, err)

	permissions, err := db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.True(t, permissions.Contains(data.AllPermissions()))
}

func TestDeleteApiToken(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	apiId, err := db.SaveApiToken(user.ID, nil, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	err = db.DeleteApiToken(apiId)
	require.NoError(t, err)

	_, err = db.GetApiTokenInfo(apiId)
	require.Error(t, err)

	_, err = db.GetApiUserID(tokenHash)
	require.Error(t, err)
}

func TestDeleteUserDeletesToken(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	apiId, err := db.SaveApiToken(user.ID, nil, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	err = db.DeleteUser(user.ID)
	require.NoError(t, err)

	_, err = db.GetApiTokenInfo(apiId)
	require.Error(t, err)

	_, err = db.GetApiUserID(tokenHash)
	require.Error(t, err)
}

func TestRevokeApiTokenSuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	apiId, err := db.SaveApiToken(user.ID, nil, tokenHash, label, data.Permissions{})
	require.NoError(t, err)

	err = db.RevokeApiToken(user.ID, apiId)
	require.NoError(t, err)

	_, err = db.GetApiTokenInfo(apiId)
	require.NoError(t, err)

	_, err = db.GetApiUserID(tokenHash)
	require.Error(t, err)
}

func TestGetApiPermissionsSuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("apitest", "hash")
	require.NoError(t, err)

	tokenHash := [32]byte{1, 2, 3}
	label := "test-token"
	permissions := data.Permissions{data.ReadUser, data.WriteFiles}
	_, err = db.SaveApiToken(user.ID, nil, tokenHash, label, permissions)
	require.NoError(t, err)

	retrievedPermissions, err := db.GetApiPermissions(tokenHash)
	require.NoError(t, err)
	t.Logf("Expected permissions: %v, Retrieved permissions: %v", permissions, retrievedPermissions)
	assert.ElementsMatch(t, permissions, retrievedPermissions)
}

func TestSetUserProfilePictureSuccess(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("profilepicuser", "hash")
	require.NoError(t, err)

	err = db.SetUserProfilePicture(user.ID, []byte("fake image bytes"))
	require.NoError(t, err)

	picData, err := db.GetUserProfilePicture(user.ID)
	require.NoError(t, err)
	// image is now a webp so cant do byte comparison
	assert.NotEmpty(t, picData)
}

func TestSetUserProfilePictureNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)

	err := db.SetUserProfilePicture(999, []byte("fake image bytes"))
	require.Error(t, err)
}	

func TestGetUserProfilePictureNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)

	_, err := db.GetUserProfilePicture(999)
	require.Error(t, err)
}

func TestGetUserProfilePictureDefaultPicture(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("nopictureuser", "hash")
	require.NoError(t, err)

	b, err := db.GetUserProfilePicture(user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, b) // should return default picture, which is not empty
}
