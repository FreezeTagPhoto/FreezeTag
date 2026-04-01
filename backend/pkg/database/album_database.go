package database

import (
	"database/sql"
	"fmt"
)

type PrivacyLevel uint8
type VisbilityLevel uint8

const (
	VisPrivate VisbilityLevel = iota
	VisPublic
	VisAdmin
)

type AlbumDatabase interface {
	CreateAlbum(string, UserID, PrivacyLevel) (AlbumId, error)
	SetImageAlbum(ImageId, AlbumId, UserID) error
	RemoveAlbum(string, UserID) error
	RenameAlbum(string, string, UserID) error
	RemoveImageFromAlbum(ImageId, AlbumId, UserID) error
	GetAlbumIds(UserID) ([]AlbumId, error)
	GetAlbumNames(UserID) ([]string, error)
	GetAssociatedAlbumIds(ImageId, UserID) ([]AlbumId, error)
	GetAlbumImages(AlbumId, UserID) ([]ImageId, error)
	GetAlbumTagCounts(AlbumId, UserID) (map[string]int64, error)
	GetAlbumIdByName(string, UserID) (AlbumId, error)
	SetAlbumVisibility(AlbumId, PrivacyLevel, UserID) error
	SetUserAlbumPermission(AlbumId, UserID, PrivacyLevel, UserID) error
}

// who is asking for the data and what are they allowed to see?
const visJoins = `
	CROSS JOIN (SELECT visibility_mode FROM Users WHERE id = ?) AS up
	LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = ?
`

// visibility rules for a given user and album
const visWhere = `(
	up.visibility_mode = 2 -- admin bypass 
	OR a.userId = ?        -- owner bypass
	OR aa.access_level > 0 -- explicit access
	OR (up.visibility_mode = 1 AND a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
)`

type SqliteAlbumDatabase struct {
	db *sql.DB
}

func InitSQLiteAlbumDatabase(datasource string) (SqliteAlbumDatabase, error) {
	registerExtendedSqlite("sqlite3_extrafunc")
	db, err := sql.Open("sqlite3_extrafunc", datasource)
	if err != nil {
		return SqliteAlbumDatabase{}, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return SqliteAlbumDatabase{}, err
	}
	return SqliteAlbumDatabase{db}, nil
}

func (db SqliteAlbumDatabase) getVisibilityMode(userID UserID) (VisbilityLevel, error) {
	var visibility int
	if userID == 0 {
		return VisAdmin, nil
	}
	err := db.db.QueryRow("SELECT visibility_mode FROM Users WHERE id = ?", userID).Scan(&visibility)
	return VisbilityLevel(visibility), err
}

func (db SqliteAlbumDatabase) userAuthorizedForAlbum(albumId AlbumId, userID UserID) (bool, error) {
	var count int

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return false, err
	}
	switch userVisibility {
	case VisAdmin:
		return true, nil
	case VisPublic:
		query := ` 
			SELECT COUNT(*) FROM Albums a
			LEFT JOIN AlbumAccess aa 
			ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.id = ?
			AND (
				(a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
				OR aa.access_level > 0 
				OR a.userId = ?
			)
		`
		args := []any{userID, albumId, userID}
		err := db.db.QueryRow(query, args...).Scan(&count)
		return count > 0, err
	case VisPrivate:
		query := ` 
			SELECT COUNT(*) FROM Albums a
			LEFT JOIN AlbumAccess aa 
			ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.id = ?
			AND (aa.access_level > 0 OR a.userId = ?)
		`
		args := []any{userID, albumId, userID}
		err := db.db.QueryRow(query, args...).Scan(&count)
		return count > 0, err
	default:
		return false, fmt.Errorf("invalid visibility mode for user %v", userID)
	}
}

func (db SqliteAlbumDatabase) CreateAlbum(name string, userId UserID, visibilityMode PrivacyLevel) (AlbumId, error) {
	var id int64
	err := db.db.QueryRow("INSERT INTO Albums (album_name, userId, visibility_mode) VALUES (?, ?, ?) RETURNING id", name, userId, visibilityMode).Scan(&id)
	if err != nil {
		return 0, err
	}
	return AlbumId(id), nil
}

func (db SqliteAlbumDatabase) SetImageAlbum(imageId ImageId, albumId AlbumId, userID UserID) error {
	query := "INSERT INTO AlbumImages (albumId, imageId) VALUES (?, ?)"
	_, err := db.db.Exec(query, albumId, imageId)
	return err
}

func (db SqliteAlbumDatabase) RemoveAlbum(name string, userID UserID) error {
	_, err := db.db.Exec("DELETE FROM Albums WHERE album_name = ? AND userId = ?", name, userID)
	return err
}

func (db SqliteAlbumDatabase) RenameAlbum(oldName string, newName string, userID UserID) error {
	res, err := db.db.Exec("UPDATE Albums SET album_name = ? WHERE album_name = ? AND userId = ?", newName, oldName, userID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("album %q not found or not owned by %v", oldName, userID)
	}

	return nil
}

func (db SqliteAlbumDatabase) RemoveImageFromAlbum(imageId ImageId, albumId AlbumId, userID UserID) error {
	_, err := db.db.Exec("DELETE FROM AlbumImages WHERE albumId = ? AND imageId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)", albumId, imageId, albumId, userID)
	return err
}

func (db SqliteAlbumDatabase) GetAlbumIds(userID UserID) ([]AlbumId, error) {
	var query string
	var args []any

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return nil, err
	}
	switch userVisibility {
	case VisAdmin:
		query = "SELECT id FROM Albums ORDER BY album_name ASC"
	case VisPublic: //album is public or user has explicit access or user is owner of the album
		query = ` 
			SELECT a.id FROM Albums a 
			LEFT JOIN AlbumAccess aa 
			ON aa.albumId = a.id AND aa.userId = :uid
			WHERE (a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
			OR aa.access_level > 0 
			OR a.userId = :uid
			ORDER BY a.album_name ASC
		`
		args = []any{sql.Named("uid", userID)}
	case VisPrivate: //explicitly invited to the album
		query = `
			SELECT a.id FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE aa.access_level > 0 OR a.userId = :uid
			ORDER BY album_name ASC
		`
		args = []any{sql.Named("uid", userID)}
	default:
		return nil, fmt.Errorf("invalid visibility mode for user %v", userID)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []AlbumId
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		albums = append(albums, AlbumId(id))
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumNames(userID UserID) ([]string, error) {
	var query string
	var args []any

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return nil, err
	}
	switch userVisibility {
	case VisAdmin:
		query = "SELECT album_name FROM Albums ORDER BY album_name ASC"
	case VisPublic: //album is public or user has explicit access or user is owner of the album
		query = ` 
			SELECT a.album_name FROM Albums a 
			LEFT JOIN AlbumAccess aa 
			ON aa.albumId = a.id AND aa.userId = :uid
			WHERE (a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
			OR aa.access_level > 0 
			OR a.userId = :uid
			ORDER BY a.album_name ASC
		`
		args = []any{sql.Named("uid", userID)}
	case VisPrivate: //explicitly invited to the album
		query = `
			SELECT album_name FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE aa.access_level > 0 OR a.userId = :uid
			ORDER BY album_name ASC
		`
		args = []any{sql.Named("uid", userID)}
	default:
		return nil, fmt.Errorf("invalid visibility mode for user %v", userID)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		albums = append(albums, name)
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAssociatedAlbumIds(imageID ImageId, userID UserID) ([]AlbumId, error) {
	var query string
	var args []any
	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return nil, err
	}
	switch userVisibility {
	case VisAdmin:
		query = "SELECT DISTINCT albumId FROM AlbumImages WHERE imageId = :id"
		args = []any{sql.Named("id", imageID)}
	case VisPublic:
		query = `
			SELECT DISTINCT ai.albumId FROM AlbumImages ai
			LEFT JOIN AlbumAccess aa ON aa.albumId = ai.albumId AND aa.userId = :uid
			WHERE imageId = :id
			AND ( 
				(a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
				OR aa.access_level > 0 
				OR a.userId = :uid
			)
		`
	case VisPrivate:
		query = `
			SELECT DISTINCT ai.albumId FROM AlbumImages ai
			LEFT JOIN AlbumAccess aa ON aa.albumId = ai.albumId AND aa.userId = :uid
			WHERE imageId = :id 
			AND (aa.access_level > 0 OR a.userId = :uid)
		`
		args = []any{sql.Named("id", imageID), sql.Named("uid", userID)}
	default:
		return nil, fmt.Errorf("invalid visibility mode for user %v", userID)
	}
	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []AlbumId
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		albums = append(albums, AlbumId(id))
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumImages(albumId AlbumId, userID UserID) ([]ImageId, error) {
	ok, err := db.userAuthorizedForAlbum(albumId, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumId)
	}
	query := ` 
		SELECT imageId FROM AlbumImages WHERE albumId = :aid
	`
	args := []any{sql.Named("aid", albumId)}
	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []ImageId
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		images = append(images, ImageId(id))
	}
	return images, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumTagCounts(albumId AlbumId, userID UserID) (map[string]int64, error) {
	var query string
	var args []any

	ok, err := db.userAuthorizedForAlbum(albumId, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumId)
	}

	query = `
		SELECT it.tag, COUNT(*) FROM AlbumImages ai
		JOIN ImageTags it ON it.imageId = ai.imageId
		WHERE ai.albumId = :aid
		GROUP BY it.tag
	`
	args = []any{sql.Named("aid", albumId)}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var tag string
		var count int64
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, err
		}
		counts[tag] = count
	}
	return counts, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumIdByName(name string, userID UserID) (AlbumId, error) {
	var query string
	var args []any
	var id int64

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return 0, err
	}

	switch userVisibility {
	case VisAdmin:
		query = "SELECT id FROM Albums WHERE album_name = ? ORDER BY id ASC LIMIT 1"
		args = []any{name}
	case VisPublic:
		query = `
			SELECT a.id FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.album_name = ?
			AND (
				(a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
				OR aa.access_level > 0
				OR a.userId = ?
			)
			ORDER BY CASE WHEN a.userId = ? THEN 0 ELSE 1 END, a.id ASC
			LIMIT 1
		`
		args = []any{userID, name, userID, userID}
	case VisPrivate:
		query = `
			SELECT a.id FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.album_name = ?
			AND (aa.access_level > 0 OR a.userId = ?)
			ORDER BY CASE WHEN a.userId = ? THEN 0 ELSE 1 END, a.id ASC
			LIMIT 1
		`
		args = []any{userID, name, userID, userID}
	default:
		return 0, fmt.Errorf("invalid visibility mode for user %v", userID)
	}

	err = db.db.QueryRow(query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return AlbumId(id), nil
}

func (db SqliteAlbumDatabase) SetAlbumVisibility(albumId AlbumId, mode PrivacyLevel, userID UserID) error {
	userVis, err := db.getVisibilityMode(userID)
	var query string
	if err != nil {
		return err
	}
	switch userVis {
	case VisAdmin:
		query = "UPDATE Albums SET visibility_mode = :vis WHERE id = :aid"
	default: // user owns the album
		query = `
			UPDATE Albums SET visibility_mode = :vis
			WHERE id = :aid AND userId = :uid
		`
	}
	args := []any{sql.Named("vis", mode), sql.Named("aid", albumId), sql.Named("uid", userID)}
	res, err := db.db.Exec(query, args...)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("album %v not found", albumId)
	}
	return nil
}

func (db SqliteAlbumDatabase) SetUserAlbumPermission(albumId AlbumId, targetUser UserID, permission PrivacyLevel, requesterID UserID) error {
	query := `
        INSERT INTO AlbumAccess (albumId, userId, access_level)
        SELECT ?, ?, ?
        WHERE EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)
        ON CONFLICT(albumId, userId) DO UPDATE SET access_level = excluded.access_level`

	res, err := db.db.Exec(query, albumId, targetUser, permission, albumId, requesterID)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("forbidden: album %v not found or not owned by %v", albumId, requesterID)
	}
	return nil
}
