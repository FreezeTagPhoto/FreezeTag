package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	mockUserDatabase "freezetag/backend/mocks/UserDatabase"
	"freezetag/backend/pkg/database"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserByUsername(t *testing.T) {
	time := time.Now().Unix()
	user := &database.PublicUser{
		ID:           1,
		Username:     "testuser",
		CreatedAt:    time,
		PasswordHash: "hashedpassword",
	}
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserByUsername("testuser").
		Return(user, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	userGot, err := userRepo.GetUserByUsername("testuser")
	require.NoError(t, err)
	assert.Equal(t, user, userGot)
}

func TestGetUserById(t *testing.T) {
	time := time.Now().Unix()
	user := &database.PublicUser{
		ID:           100,
		Username:     "testuser",
		CreatedAt:    time,
		PasswordHash: "hashedpassword",
	}
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserById(database.UserID(100)).
		Return(user, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	userGot, err := userRepo.GetUserByID(database.UserID(100))
	require.NoError(t, err)
	assert.Equal(t, user, userGot)
}

func TestGetUserByUsernameNotFound(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserByUsername("nonexistent").
		Return(nil, sql.ErrNoRows).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err := userRepo.GetUserByUsername("nonexistent")
	require.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestGetUserByIdNotFound(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserById(database.UserID(1)).
		Return(nil, sql.ErrNoRows).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err := userRepo.GetUserByID(database.UserID(1))
	require.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestGetUserByUsernameInternalError(t *testing.T) {
	err := fmt.Errorf("internal error")
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserByUsername("someuser").
		Return(nil, err).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err2 := userRepo.GetUserByUsername("someuser")
	require.Error(t, err2)
	assert.Equal(t, err, err2)
}

func TestGetUserByIdInternalError(t *testing.T) {
	err := fmt.Errorf("internal error")
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetUserById(database.UserID(1)).
		Return(nil, err).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err2 := userRepo.GetUserByID(database.UserID(1))
	require.Error(t, err2)
	assert.Equal(t, err, err2)
}

func TestListUsers(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	expectedUsers := []*database.PublicUser{
		{Username: "cant"},
		{Username: "think"},
		{Username: "of"},
		{Username: "more"},
		{Username: "usernames"},
	}
	mockDB.EXPECT().
		ListUsers().
		Return(expectedUsers, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	users, err := userRepo.ListAllUsers()
	require.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
}

func TestAddUserDuplicateUsername(t *testing.T) {
	user := &database.PublicUser{
		ID:           database.UserID(0),
		Username:     "duplicateuser",
		CreatedAt:    time.Now().Unix(),
		PasswordHash: "hashedpassword",
	}
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		AddUser("duplicateuser", "hashedpassword").
		Return(user, sqlite3.Error{Code: sqlite3.ErrConstraint}).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err := userRepo.AddUser("duplicateuser", "hashedpassword")
	require.Error(t, err)
	assert.Equal(t, ErrDuplicateUsername, err)
}

func TestAddUserInternalError(t *testing.T) {
	user := &database.PublicUser{
		ID:           database.UserID(0),
		Username:     "duplicateuser",
		CreatedAt:    time.Now().Unix(),
		PasswordHash: "hashedpassword",
	}
	err := fmt.Errorf("internal error")
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		AddUser("duplicateuser", "hashedpassword").
		Return(user, err).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err = userRepo.AddUser("duplicateuser", "hashedpassword")
	require.Error(t, err)
	assert.Equal(t, err, err)
}

func TestChangePasswordTrue(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		SetUserPassword(database.UserID(1), "newhashedpassword").
		Return(true, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	err := userRepo.ChangePassword(database.UserID(1), "newhashedpassword")
	require.NoError(t, err)
}

func TestChangePasswordFalse(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		SetUserPassword(database.UserID(1), "newhashedpassword").
		Return(false, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	err := userRepo.ChangePassword(database.UserID(1), "newhashedpassword")
	require.Error(t, err)
	assert.Equal(t, ErrPasswordChangeFailed, err)
}

func TestChangePasswordInternalError(t *testing.T) {
	err := fmt.Errorf("internal error")
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		SetUserPassword(database.UserID(1), "newhashedpassword").
		Return(false, err).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	err2 := userRepo.ChangePassword(database.UserID(1), "newhashedpassword")
	require.Error(t, err2)
	assert.Equal(t, err, err2)
}

func TestGetUserPasswordHash(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetPasswordHash(database.UserID(1)).
		Return("hashedpassword", nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	passwordHash, err := userRepo.GetUserPasswordHash(database.UserID(1))
	require.NoError(t, err)
	assert.Equal(t, "hashedpassword", passwordHash)
}

func TestGetUserPasswordHashUserNotFound(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetPasswordHash(database.UserID(1)).
		Return("", sql.ErrNoRows).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err := userRepo.GetUserPasswordHash(database.UserID(1))
	require.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestGetUserPasswordHasInternalServerError(t *testing.T) {
	err := fmt.Errorf("internal error")
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		GetPasswordHash(database.UserID(1)).
		Return("", err).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	_, err2 := userRepo.GetUserPasswordHash(database.UserID(1))
	require.Error(t, err2)
	assert.Equal(t, err, err2)
}

func TestAddUserSuccess(t *testing.T) {
	user := &database.PublicUser{
		ID:           database.UserID(0),
		Username:     "duplicateuser",
		CreatedAt:    time.Now().Unix(),
		PasswordHash: "hashedpassword",
	}
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	mockDB.EXPECT().
		AddUser("newuser", "hashedpassword").
		Return(user, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	userGot, err := userRepo.AddUser("newuser", "hashedpassword")
	require.NoError(t, err)
	assert.Equal(t, user, userGot)
}

func TestListUsernames(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)
	expectedUsers := []*database.PublicUser{
		{Username: "cant"},
		{Username: "think"},
		{Username: "of"},
		{Username: "more"},
	}
	mockDB.EXPECT().
		ListUsers().
		Return(expectedUsers, nil).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	users, err := userRepo.ListUsernames()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"cant", "think", "of", "more"}, users)
}

func TestListUsernamesErr(t *testing.T) {
	mockDB := mockUserDatabase.NewMockUserDatabase(t)

	mockDB.EXPECT().
		ListUsers().
		Return(nil, errors.New("an error")).
		Once()

	userRepo := InitDefaultUserRepository(mockDB)
	users, err := userRepo.ListUsernames()
	require.Error(t, err)
	assert.Nil(t, users)
}
