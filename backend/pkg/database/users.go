package database

import (
	"database/sql"
	"errors"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/images"

	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserID uint64
type TokenID uint64

type PublicUser struct {
	ID        UserID `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`

	PasswordHash string `json:"-"`
}

type TokenStatus string

type ProfilePicture []byte

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusExpired TokenStatus = "expired"
	TokenStatusRevoked TokenStatus = "revoked"
)

type ApiTokenInfo struct {
	Id     TokenID     `json:"id"`
	Label  string      `json:"label"`
	Status TokenStatus `json:"status"` // active, expired, revoked
}

type UserDatabase interface {
	// Add a new User, return error if username already exists
	AddUser(username string, passwordHash string) (*PublicUser, error)
	// return a User by Username, return error if not found
	GetUserByUsername(username string) (*PublicUser, error)
	// Get User by ID, return error if not found
	GetUserById(id UserID) (*PublicUser, error)
	// Set a User Password, return error if user not found or issue occurs
	SetUserPassword(userID UserID, newPasswordHash string) error
	// Get Password Hash for a User by ID, return error if ID is not found
	GetPasswordHash(userID UserID) (string, error)
	// List all users in the database
	ListUsers() ([]*PublicUser, error)
	// Get User permissions by user ID, return error if not found
	GetUserPermissions(userID UserID) (data.Permissions, error)
	// revoke permissions for a user by user ID
	RevokeUserPermissions(userID UserID, permissions data.Permissions) error
	// grant permissions for a user by user ID
	GrantUserPermissions(userID UserID, permissions data.Permissions) error
	// delete a user by ID
	DeleteUser(userID UserID) error
	// save a new API token for a user, return the token hash and error if user not found. expiresAt can be nil for no expiration
	SaveApiToken(userID UserID, expiresAt *time.Time, tokenHash [32]byte, label string, permissions data.Permissions) (TokenID, error)
	// Get API permissions for a user by token hash, return error if not found
	GetApiPermissions(tokenHash [32]byte) (data.Permissions, error)
	// get user ID associated with an API token hash, return error if not found
	GetApiUserID(tokenHash [32]byte) (UserID, error)
	// soft delete an API token by its id. This does NOT delete the token from the database, but it does remove permissions/userID/etc
	RevokeApiToken(userId UserID, tokenId TokenID) error
	// Admin level operation to revoke an API token by its id, regardless of associated user
	AdminRevokeApiToken(tokenId TokenID) error
	// delete an API token by its id from the database
	DeleteApiToken(tokenId TokenID) error
	// get the label for an API token by its id, return error if not found.
	// revoked and expired tokens are still returned by this function, but not deleted tokens
	GetApiTokenInfo(tokenId TokenID) (ApiTokenInfo, error)
	// get all API token labels for a user by user ID, return error if user not found
	GetUserApiTokenInfo(userID UserID) ([]ApiTokenInfo, error)
	// ensure an admin has all permissions
	EnsureAdmin(userID UserID) error
	// Get all users in the database with their ID, username, and createdAt fields. Does not include password hashes. Return error if issue occurs
	AllUsers() ([]PublicUser, error)
	// Set user profile picture, return error if user not found or issue occurs
	SetUserProfilePicture(userID UserID, pictureData []byte) error
	// Get user profile picture, return error if user not found or issue occurs
	GetUserProfilePicture(userID UserID) (ProfilePicture, error)
}

type SqliteUserDatabase struct {
	db *sql.DB
}

//go:embed schema.sql
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

	for _, p := range data.AllPermissions() {
		if _, err := stmt.Exec(p.Slug, p.Name, p.Description); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s SqliteUserDatabase) AddUser(username string, passwordHash string) (*PublicUser, error) {

	createdAt := time.Now().Unix()
	defaultPicture, err := images.DefaultProfilePicture(username)
	if err != nil {
		return nil, err
	}
	result, err := s.db.Exec(
		"INSERT INTO Users (username, passwordHash, createdAt, profilePicture) VALUES (?, ?, ?, ?)",
		username,
		passwordHash,
		createdAt,
		defaultPicture,
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

func (s SqliteUserDatabase) SetUserPassword(userID UserID, newPasswordHash string) error {
	result, err := s.db.Exec(
		"UPDATE Users SET passwordHash = ? WHERE id = ?",
		newPasswordHash,
		userID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("failed to update password: user not found")
	}
	return nil
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

// API token related methods
func (s SqliteUserDatabase) SaveApiToken(userID UserID, expiresAt *time.Time, tokenHash [32]byte, label string, permissions data.Permissions) (TokenID, error) {
	var expiresUnix any
	if expiresAt != nil {
		expiresUnix = expiresAt.Unix()
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.Exec(`
		INSERT INTO API_Token (userId, tokenHash, createdAt, expiresAt, label) 
		VALUES (?, ?, ?, ?, ?)`,
		userID, tokenHash[:], time.Now().Unix(), expiresUnix, label,
	)
	if err != nil {
		return 0, err
	}
	tokenID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	permStmt, err := tx.Prepare(`
		INSERT INTO Token_Permissions (tokenId, permissionId)
		VALUES (?, (SELECT id FROM App_Permissions WHERE slug = ?))`)
	if err != nil {
		return 0, err
	}
	defer permStmt.Close() //nolint:errcheck

	for _, perm := range permissions {
		_, err := permStmt.Exec(tokenID, perm.Slug)
		if err != nil {
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return TokenID(tokenID), nil
}

func (s SqliteUserDatabase) GetApiPermissions(tokenHash [32]byte) (data.Permissions, error) {
	query := `
		SELECT p.slug, p.name, p.description
		FROM API_Token t
		JOIN Token_Permissions up ON t.id = up.tokenId
		JOIN App_Permissions p  ON up.permissionId = p.id
		WHERE t.tokenHash = ? 
		AND t.revoked = 0 
		AND (t.expiresAt > ? OR t.expiresAt IS NULL)
	`
	rows, err := s.db.Query(
		query,
		tokenHash[:],
		time.Now().Unix(),
	)
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

func (s SqliteUserDatabase) GetApiUserID(tokenHash [32]byte) (UserID, error) {
	var userID UserID
	query := `
		SELECT userId 
		FROM API_Token 
		WHERE tokenHash = ? 
		AND (expiresAt IS NULL OR expiresAt > ?)
		AND revoked = 0 
	`
	err := s.db.QueryRow(
		query,
		tokenHash[:],
		time.Now().Unix(),
	).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s SqliteUserDatabase) RevokeApiToken(userID UserID, tokenID TokenID) error {
	query := `
		UPDATE API_Token 
		SET revoked = 1 
		WHERE id = ? AND userId = ?
	`
	result, err := s.db.Exec(query, tokenID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("failed to revoke API token")
	}
	return nil
}

func (s SqliteUserDatabase) AdminRevokeApiToken(tokenID TokenID) error {
	query := `
		UPDATE API_Token 
		SET revoked = 1 
		WHERE id = ?
	`
	result, err := s.db.Exec(query, tokenID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("failed to revoke API token")
	}
	return nil
}

func (s SqliteUserDatabase) DeleteApiToken(tokenID TokenID) error {
	query := `
		DELETE FROM API_Token 
		WHERE id = ?
	`
	result, err := s.db.Exec(query, tokenID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if affected == 0 || affected > 1 || err != nil {
		return errors.New("an issue occurred while deleting API token")
	}
	return nil
}

func (s SqliteUserDatabase) GetUserApiTokenInfo(userID UserID) ([]ApiTokenInfo, error) {
	var info ApiTokenInfo
	var revoked int
	var expiresAt sql.NullInt64
	query := `
		SELECT label, id, expiresAt, revoked
		FROM API_Token
		WHERE userId = ?
	`
	rows, err := s.db.Query(
		query,
		userID,
		time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var labels []ApiTokenInfo
	for rows.Next() {
		if err := rows.Scan(&info.Label, &info.Id, &expiresAt, &revoked); err != nil {
			return nil, err
		}
		info.Status = getTokenStatus(revoked, expiresAt)
		labels = append(labels, info)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return labels, nil
}

func (s SqliteUserDatabase) GetApiTokenInfo(tokenID TokenID) (ApiTokenInfo, error) {
	var info ApiTokenInfo
	var revoked int
	var expiresAt sql.NullInt64
	query := `
		SELECT label, expiresAt, revoked
		FROM API_Token
		WHERE id = ?
	`
	err := s.db.QueryRow(
		query,
		tokenID,
		time.Now().Unix(),
	).Scan(&info.Label, &expiresAt, &revoked)
	if err != nil {
		return ApiTokenInfo{}, err
	}
	info.Status = getTokenStatus(revoked, expiresAt)
	return info, nil
}

func (s SqliteUserDatabase) EnsureAdmin(userID UserID) error {

	err := s.GrantUserPermissions(userID, data.AllPermissions())
	if err != nil {
		return err
	}
	return nil
}

func getTokenStatus(revoked int, expiresAt sql.NullInt64) TokenStatus {
	if revoked == 1 {
		return TokenStatusRevoked
	}
	if expiresAt.Valid && expiresAt.Int64 < time.Now().Unix() {
		return TokenStatusExpired
	}
	return TokenStatusActive
}

func (s SqliteUserDatabase) AllUsers() ([]PublicUser, error) {
	rows, err := s.db.Query("SELECT id, username, createdAt FROM Users")
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var users []PublicUser
	for rows.Next() {
		var user PublicUser
		err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s SqliteUserDatabase) SetUserProfilePicture(userID UserID, pictureData []byte) error {
	result, err := s.db.Exec(
		"UPDATE Users SET profilePicture = ? WHERE id = ?",
		pictureData,
		userID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("failed to update profile picture: user not found")
	}
	return nil
}
	

func (s SqliteUserDatabase) GetUserProfilePicture(userID UserID) (ProfilePicture, error) {
	var pictureData ProfilePicture
	err := s.db.QueryRow(
		"SELECT profilePicture FROM Users WHERE id = ?",
		userID,
	).Scan(&pictureData)
	if err != nil {
		return nil, err
	}
	return pictureData, nil
}
