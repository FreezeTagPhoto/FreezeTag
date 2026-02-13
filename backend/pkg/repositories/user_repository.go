package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"

	"github.com/mattn/go-sqlite3"
)

type UserRepository interface {
	GetUserByUsername(username string) (*database.PublicUser, error)
	GetUserByID(id database.UserID) (*database.PublicUser, error)
	GetApiPermissions(tokenHash [32]byte) (data.Permissions, error)
	GetUserPermissions(userID database.UserID) (data.Permissions, error)
	AddUser(username string, passwordHash string) (*database.PublicUser, error)

	GetUserPasswordHash(userID database.UserID) (string, error)
	ChangePassword(userID database.UserID, newPasswordHash string) error
	ListUsernames() ([]string, error)
	ListAllUsers() ([]*database.PublicUser, error)

	GrantAdminPermissions(userID database.UserID) error
	RevokeAllPermissions(userID database.UserID) error
	RevokePermissions(userID database.UserID, permissions data.Permissions) error
	GrantPermissions(userID database.UserID, permissions data.Permissions) error
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrDuplicateUsername    = errors.New("username already exists")
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

func (r *DefaultUserRepository) GetUserByID(id database.UserID) (*database.PublicUser, error) {
	user, err := r.GetUserById(id)
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
	users, err := r.ListUsers()
	if err != nil {
		return nil, err
	}
	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}
	return usernames, nil
}

func (r *DefaultUserRepository) ListAllUsers() ([]*database.PublicUser, error) {
	return r.ListUsers()
}

func (r *DefaultUserRepository) GetApiPermissions(tokenHash [32]byte) (data.Permissions, error) {
	permissions, err := r.UserDatabase.GetApiPermissions(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API token: %w", err)
	}
	if len(permissions) == 0 {
		return nil, fmt.Errorf("invalid API token: no permissions found")
	}
	return permissions, nil
}

func (r *DefaultUserRepository) GetUserPermissions(userID database.UserID) (data.Permissions, error) {
	permissions, err := r.UserDatabase.GetUserPermissions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

func (r *DefaultUserRepository) GrantAdminPermissions(userID database.UserID) error {
	allPerms := data.All()
	err := r.UserDatabase.GrantUserPermissions(userID, allPerms)
	if err != nil {
		return fmt.Errorf("failed to grant admin permissions: %w", err)
	}
	return nil
}

func (r *DefaultUserRepository) RevokeAllPermissions(userID database.UserID) error {
	allPerms := data.All()
	err := r.UserDatabase.RevokeUserPermissions(userID, allPerms)
	if err != nil {
		return fmt.Errorf("failed to revoke all permissions: %w", err)
	}
	return nil
}

func (r *DefaultUserRepository) RevokePermissions(userID database.UserID, permissions data.Permissions) error {
	err := r.UserDatabase.RevokeUserPermissions(userID, permissions)
	if err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}
	return nil
}

func (r *DefaultUserRepository) GrantPermissions(userID database.UserID, permissions data.Permissions) error {
	err := r.UserDatabase.GrantUserPermissions(userID, permissions)
	if err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}
	return nil
}
