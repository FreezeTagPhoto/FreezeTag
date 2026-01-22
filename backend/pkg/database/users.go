package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserID int64

type UserDatabase interface {
	AddUser(username string, passwordHash string) error
	GetUserIDByUsername(username string) (UserID, error)
	GetUserPasswordHashByID(userID UserID) (string, error)
	ResetPassword(userID UserID, newPasswordHash string) error
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

func (s SqliteUserDatabase) AddUser(username string, passwordHash string) error {

	dateCreated := time.Now().Unix()
	_, err := s.db.Exec(
		"INSERT INTO Users (username, passwordHash, createdAt) VALUES (?, ?, ?)",
		username,
		passwordHash,
		dateCreated,
	)
	return err
}

func (s SqliteUserDatabase) GetUserIDByUsername(username string) (UserID, error) {
	var id UserID
	err := s.db.QueryRow(
		"SELECT id FROM Users WHERE username = ?",
		username,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s SqliteUserDatabase) GetUserPasswordHashByID(userID UserID) (string, error) {
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

func (s SqliteUserDatabase) ResetPassword(userID UserID, newPasswordHash string) error {
	return fmt.Errorf("not implemented")
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