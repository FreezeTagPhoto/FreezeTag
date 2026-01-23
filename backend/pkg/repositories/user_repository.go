package repositories

import (
	"database/sql"
	"errors"
	"freezetag/backend/pkg/database"

	"github.com/mattn/go-sqlite3"
)

type UserRepository interface {
	GetUserByUsername(username string) (*database.PublicUser, error)
	GetUserById(id database.UserID) (*database.PublicUser, error)
	AddUser(username string, passwordHash string) (*database.PublicUser, error)

	GetUserPasswordHash(userID database.UserID) (string, error)
	ChangePassword(userID database.UserID, newPasswordHash string) error
	ListUsernames() ([]string, error)	
}

var ( 
	ErrUserNotFound = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrPasswordChangeFailed = errors.New("password change failed")
)

type DefaultUserRepository struct {
	database.UserDatabase
}

func InitDefaultUserRepository(db database.UserDatabase) UserRepository {
	return &DefaultUserRepository{
		UserDatabase: db,
	}
}

func (r *DefaultUserRepository) GetUserById(id database.UserID) (*database.PublicUser, error) {
	user, err := r.UserDatabase.GetUserById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound 
		}
		return nil, err
	}
	return user, nil
}

func (r *DefaultUserRepository) GetUserByUsername(username string) (*database.PublicUser, error) {
	user, err := r.UserDatabase.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound 
		}
		return nil, err
	}
	return user, nil
}

func (r *DefaultUserRepository) GetUserPasswordHash(userID database.UserID) (string, error) {
	passwordHash, err := r.GetPasswordHash(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUserNotFound 
		}
		return "", err
	}
	return passwordHash, nil
}

func (r *DefaultUserRepository) AddUser(username string, passwordHash string) (*database.PublicUser, error) {
	user, err := r.UserDatabase.AddUser(username, passwordHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil, ErrDuplicateUsername 
		}
		return nil, err
	}
	return user, nil
}

func (r *DefaultUserRepository) ChangePassword(userID database.UserID, newPasswordHash string) error {
	success, err := r.SetUserPassword(userID, newPasswordHash)
	if err != nil {
		return err
	}
	if !success {
		return ErrPasswordChangeFailed
	}
	return nil
}

func (r *DefaultUserRepository) ListUsernames() ([]string, error) {
	return r.UserDatabase.ListUsernames()
}

