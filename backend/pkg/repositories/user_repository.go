package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"freezetag/backend/pkg/database"

	"github.com/mattn/go-sqlite3"
)

type UserRepository interface {
	GetUserByUsername(username string) (*database.PublicUser, error)
	GetUserById(username string) (database.UserID, error)
	GetUserPasswordHash(userID database.UserID) (string, error)
	AddUser(username string, passwordHash string) (database.UserID, error)
	ChangePassword(userID database.UserID, newPasswordHash string) error
	ListUsernames() ([]string, error)	
}

var ( 
	ErrUserNotFound = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
)

type DefaultUserRepository struct {
	database.SqliteUserDatabase
}

func InitDefaultUserRepository(db database.SqliteUserDatabase) UserRepository {
	return &DefaultUserRepository{
		SqliteUserDatabase: db,
	}
}

func (r *DefaultUserRepository) GetUserById(username string) (database.UserID, error) {
	user, err := r.SqliteUserDatabase.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrUserNotFound 
		}
		return 0, err
	}
	return user.ID, nil
}

func (r *DefaultUserRepository) GetUserByUsername(username string) (*database.PublicUser, error) {
	user, err := r.SqliteUserDatabase.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound 
		}
		return nil, err
	}
	return user, nil
}

func (r *DefaultUserRepository) GetUserPasswordHash(userID database.UserID) (string, error) {
	passwordHash, err := r.SqliteUserDatabase.GetPasswordHash(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUserNotFound 
		}
		return "", err
	}
	return passwordHash, nil
}

func (r *DefaultUserRepository) AddUser(username string, passwordHash string) (database.UserID, error) {
	id, err := r.SqliteUserDatabase.AddUser(username, passwordHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return 0, ErrDuplicateUsername 
		}
	}
	return id, nil
}

func (r *DefaultUserRepository) ChangePassword(userID database.UserID, newPasswordHash string) error {
	return fmt.Errorf("not implemented")
}

func (r *DefaultUserRepository) ListUsernames() ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

