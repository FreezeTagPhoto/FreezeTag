package database

import (
	"freezetag/backend/pkg/database/data"
	"testing"

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

	result, err := db.SetUserPassword(user.ID, "newhashedpassword")
	require.NoError(t, err)
	assert.True(t, result)

	passwordHash, err := db.GetPasswordHash(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "newhashedpassword", passwordHash)
}

func TestSetUserPasswordNonexistentUser(t *testing.T) {
	db := createTempUserDatabase(t)
	result, err := db.SetUserPassword(999, "newhashedpassword")
	require.NoError(t, err) // Should not error even if user does not exist
	assert.False(t, result)
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

	ok, err := db.SetUserPassword(userA.ID, "newA")
	require.NoError(t, err)
	assert.True(t, ok)

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

	err = db.GrantUserPermissions(user.ID, data.All())
	require.NoError(t, err)
	permissions, err := db.GetUserPermissions(user.ID)

	require.NoError(t, err)
	assert.ElementsMatch(t, data.All(), permissions)
}

func TestGetUserPermissionsStress(t *testing.T) {
	db := createTempUserDatabase(t)

	user, err := db.AddUser("permtest", "hash")
	require.NoError(t, err)

	err = db.GrantUserPermissions(user.ID, data.All())
	require.NoError(t, err)
	permissions, err := db.GetUserPermissions(user.ID)

	require.NoError(t, err)
	assert.ElementsMatch(t, data.All(), permissions)

	err = db.RevokeUserPermissions(user.ID, data.Permissions{data.ReadUser})
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.NotContains(t, permissions, data.ReadUser)
	err = db.RevokeUserPermissions(user.ID, data.All())
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.Empty(t, permissions)
	err = db.GrantUserPermissions(user.ID, data.Permissions{data.ReadTags})
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, data.Permissions{data.ReadTags}, permissions)
	err = db.GrantUserPermissions(user.ID, data.All())
	require.NoError(t, err)
	permissions, err = db.GetUserPermissions(user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, data.All(), permissions)
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
