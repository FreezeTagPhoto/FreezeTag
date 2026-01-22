package repositories

import (
	"freezetag/backend/pkg/database"
)

type UserRepository interface {
	GetUserIDByUsername(username string) (database.UserID, error)
	GetUserPasswordHashByID(userID database.UserID) (string, error)
	AddUser(username string, passwordHash string) error
	ChangePassword(userID database.UserID, newPasswordHash string) error
	ListUsernames() ([]string, error)	
}

type DefaultUserRepository struct {
	database.SqliteUserDatabase
}

func InitDefaultUserRepository(db database.SqliteUserDatabase) UserRepository {
	return &DefaultUserRepository{
		SqliteUserDatabase: db,
	}
}

func (r *DefaultUserRepository) GetUserIDByUsername(username string) (database.UserID, error) {
	return r.SqliteUserDatabase.GetUserIDByUsername(username)
}

func (r *DefaultUserRepository) GetUserPasswordHashByID(userID database.UserID) (string, error) {
	return r.SqliteUserDatabase.GetUserPasswordHashByID(userID)
}

func (r *DefaultUserRepository) AddUser(username string, passwordHash string) error {
	return r.SqliteUserDatabase.AddUser(username, passwordHash)
}

func (r *DefaultUserRepository) ChangePassword(userID database.UserID, newPasswordHash string) error {
	return r.SqliteUserDatabase.ResetPassword(userID, newPasswordHash)
}

func (r *DefaultUserRepository) ListUsernames() ([]string, error) {
	return r.SqliteUserDatabase.ListUsernames()
}