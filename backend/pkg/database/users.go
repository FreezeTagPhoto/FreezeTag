package database

import (
	"database/sql"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserID int64

type PublicUser struct {
	ID        UserID `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`

	PasswordHash string `json:"-"`
}

type UserDatabase interface {
	// Add a new User, return error if username already exists
	AddUser(username string, passwordHash string) (*PublicUser, error)
	// return a User by Username, return error if not found
	GetUserByUsername(username string) (*PublicUser, error)
	// Get User by ID, return error if not found
	GetUserById(id UserID) (*PublicUser, error)
	// Set a User Password, return true if successful, false if user not found
	SetUserPassword(userID UserID, newPasswordHash string) (bool, error)
	// Get Password Hash for a User by ID, return error if ID is not found
	GetPasswordHash(userID UserID) (string, error)
	// List all usernames in the database
	ListUsernames() ([]string, error)
}

type SqliteUserDatabase struct {
	db *sql.DB
}

//go:embed user_schema.sql
var user_schema string

func InitSQLiteUserDatabase(datasource string) (SqliteUserDatabase, error) {
	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		return SqliteUserDatabase{}, err
	}
	_, err = db.Exec(user_schema)
	if err != nil {
		return SqliteUserDatabase{}, err
	}
	return SqliteUserDatabase{db}, nil
}

func (s SqliteUserDatabase) AddUser(username string, passwordHash string) (*PublicUser, error) {
	
	createdAt := time.Now().Unix()
	result, err := s.db.Exec(
		"INSERT INTO Users (username, passwordHash, createdAt) VALUES (?, ?, ?)",
		username,
		passwordHash,
		createdAt,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &PublicUser{
		ID:        UserID(id),
		Username:  username,
		CreatedAt: createdAt,
		PasswordHash: passwordHash,
	}, nil
}

func (s SqliteUserDatabase) GetUserById(id UserID) (*PublicUser, error) {
	var username string
	var passwordHash string
	var createdAt int64
	err := s.db.QueryRow(
		"SELECT username, passwordHash, createdAt FROM Users WHERE id = ?",
		id,
	).Scan(&username, &passwordHash, &createdAt)
	if err != nil {
		return nil, err
	}
	return &PublicUser{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
		PasswordHash: passwordHash,
	}, nil
}

func (s SqliteUserDatabase) GetUserByUsername(username string) (*PublicUser, error) {
	var id UserID
	var createdAt int64
	var passwordHash string
	err := s.db.QueryRow(
		"SELECT id, createdAt, passwordHash FROM Users WHERE username = ?",
		username,
	).Scan(&id, &createdAt, &passwordHash)
	if err != nil {
		return nil, err
	}
	return &PublicUser{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
		PasswordHash: passwordHash,
	}, nil
}

func (s SqliteUserDatabase) GetPasswordHash(userID UserID) (string, error) {
	var passwordHash string
	err := s.db.QueryRow(
		"SELECT passwordHash FROM Users WHERE id = ?",
		userID,
	).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func (s SqliteUserDatabase) SetUserPassword(userID UserID, newPasswordHash string) (bool, error) {
	result, err := s.db.Exec(
		"UPDATE Users SET passwordHash = ? WHERE id = ?",
		newPasswordHash,
		userID,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (s SqliteUserDatabase) ListUsernames() ([]string, error) {
	rows, err := s.db.Query("SELECT username FROM Users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}
