package database

import (
	"database/sql"
	"freezetag/backend/pkg/database/data"

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
	// List all users in the database
	ListUsers() ([]*PublicUser, error)
	// Get API permissions for a user by token hash, return error if not found
	GetApiPermissions(tokenHash [32]byte) (data.Permissions, error)
	// Get User permissions by user ID, return error if not found
	GetUserPermissions(userID UserID) (data.Permissions, error)
	// revoke permissions for a user by user ID
	RevokeUserPermissions(userID UserID, permissions data.Permissions) error
	// grant permissions for a user by user ID
	GrantUserPermissions(userID UserID, permissions data.Permissions) error
	// delete a user by ID
	DeleteUser(userID UserID) error
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
	defaultDB := SqliteUserDatabase{db}
	if err := defaultDB.seedPermissions(); err != nil {
		return SqliteUserDatabase{}, err
	}
	return defaultDB, nil
}

// Seed the App_Permissions table with all defined permissions
func (s SqliteUserDatabase) seedPermissions() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// calling rollback after Commit is a no-op
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO App_Permissions (slug, name, description) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	for _, p := range data.All() {
		if _, err := stmt.Exec(p.Slug, p.Name, p.Description); err != nil {
			return err
		}
	}
	return tx.Commit()
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
		ID:           UserID(id),
		Username:     username,
		CreatedAt:    createdAt,
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
		ID:           id,
		Username:     username,
		CreatedAt:    createdAt,
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
		ID:           id,
		Username:     username,
		CreatedAt:    createdAt,
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

func (s SqliteUserDatabase) ListUsers() ([]*PublicUser, error) {
	rows, err := s.db.Query("SELECT id, username, createdAt, passwordHash FROM Users")
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var users []*PublicUser
	for rows.Next() {
		var user PublicUser
		err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.PasswordHash)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (s SqliteUserDatabase) GetApiPermissions(tokenHash [32]byte) (data.Permissions, error) {
	query := `
		SELECT p.slug, p.name, p.description
		FROM API_Token t
		JOIN User_Permissions up ON t.userId = up.userId
		JOIN App_Permissions p  ON up.permissionId = p.id
		WHERE t.tokenHash = ? 
		AND t.revoked = 0 
		AND t.expiresAt > strftime('%s', 'now')
	`
	rows, err := s.db.Query(query, tokenHash[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var permissions data.Permissions
	for rows.Next() {
		var slug, name, description string
		if err := rows.Scan(&slug, &name, &description); err != nil {
			return nil, err
		}
		permissions = append(permissions, data.Permission{Slug: slug, Name: name, Description: description})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s SqliteUserDatabase) GetUserPermissions(userID UserID) (data.Permissions, error) {
	query := `
		SELECT p.slug, p.name, p.description
		FROM User_Permissions up
		JOIN App_Permissions p ON up.permissionId = p.id
		WHERE up.userId = ?
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var permissions data.Permissions
	for rows.Next() {
		var slug, name, description string
		if err := rows.Scan(&slug, &name, &description); err != nil {
			return nil, err
		}
		permissions = append(permissions, data.Permission{Slug: slug, Name: name, Description: description})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s SqliteUserDatabase) RevokeUserPermissions(userID UserID, permissions data.Permissions) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	query := `
		DELETE FROM User_Permissions 
		WHERE userId = ? 
		AND permissionId IN (SELECT id FROM App_Permissions WHERE slug = ?)
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	for _, p := range permissions {
		if _, err := stmt.Exec(userID, string(p.Slug)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s SqliteUserDatabase) GrantUserPermissions(userID UserID, permissions data.Permissions) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	query := `
		INSERT INTO User_Permissions (userId, permissionId) 
		VALUES (?, (SELECT id FROM App_Permissions WHERE slug = ?))
		ON CONFLICT (userId, permissionId) DO NOTHING`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	for _, p := range permissions {
		if _, err := stmt.Exec(userID, string(p.Slug)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s SqliteUserDatabase) DeleteUser(userID UserID) error {
	_, err := s.db.Exec("DELETE FROM Users WHERE id = ?", userID)
	return err
}
