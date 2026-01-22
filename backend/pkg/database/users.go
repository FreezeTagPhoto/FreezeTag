package database

import (
	"database/sql"
	_ "embed"
)

type UserID int64

type UserDatabase interface {
	AddUser(username string, passwordHash string) error
	GetUserIDByUsername(username string) (UserID, error)
	AuthenticateUser(username string, passwordHash string) (bool, error)
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
	_, err := s.db.Exec(
		"INSERT INTO Users (username, passwordHash, createdAt) VALUES (?, ?, strftime('%s','now'))",
		username, passwordHash,
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

func (s SqliteUserDatabase) AuthenticateUser(username string, passwordHash string) (bool, error) {
	var storedHash string
	err := s.db.QueryRow(
		"SELECT passwordHash FROM Users WHERE username = ?",
		username,
	).Scan(&storedHash)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return storedHash == passwordHash, nil
}