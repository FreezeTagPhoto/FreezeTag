package database

import (
	"database/sql"
	"fmt"
)

type PrivacyLevel uint8

const (
	Private PrivacyLevel = iota
	Public
	PublicEditable
)

type AlbumDatabase interface {
	CreateAlbum(string, UserID, PrivacyLevel) (AlbumId, error)
	SetImageAlbum(ImageId, AlbumId, UserID) error
	RemoveAlbum(string, UserID) error
	RenameAlbum(string, string, UserID) error
	RemoveImageFromAlbum(ImageId, AlbumId, UserID) error
	GetAlbumIds(UserID) ([]AlbumId, error)
	GetAlbumNames(UserID) ([]string, error)
	GetImageAlbumNames(ImageId, UserID) ([]string, error)
	GetAlbumImages(AlbumId, UserID) ([]ImageId, error)
	GetAlbumImageCount(AlbumId, UserID) (int64, error)
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

	if userID == 0 {
		query = "SELECT id FROM Albums"
	} else {
		query = fmt.Sprintf("SELECT a.id FROM Albums a %s WHERE %s", visJoins, visWhere)
		args = []any{userID, userID, userID}
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

	if userID == 0 {
		query = "SELECT album_name FROM Albums ORDER BY album_name ASC"
	} else {
		query = fmt.Sprintf("SELECT a.album_name FROM Albums a %s WHERE %s ORDER BY a.album_name ASC", visJoins, visWhere)
		args = []any{userID, userID, userID}
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (db SqliteAlbumDatabase) GetImageAlbumNames(imageID ImageId, userID UserID) ([]string, error) {
	var query string
	var args []any

	if userID == 0 {
		query = `
			SELECT a.album_name FROM Albums a 
			JOIN AlbumImages ai ON a.id = ai.albumId 
			WHERE ai.imageId = ? ORDER BY a.album_name ASC`
		args = []any{imageID}
	} else {
		query = fmt.Sprintf(`
			SELECT a.album_name FROM Albums a 
			JOIN AlbumImages ai ON a.id = ai.albumId 
			%s WHERE ai.imageId = ? AND %s ORDER BY a.album_name ASC`, visJoins, visWhere)
		args = []any{userID, userID, imageID, userID}
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumImages(albumId AlbumId, userID UserID) ([]ImageId, error) {
	var query string
	var args []any

	if userID == 0 {
		query = "SELECT imageId FROM AlbumImages WHERE albumId = ? ORDER BY imageId ASC"
		args = []any{albumId}
	} else {
		query = fmt.Sprintf(`
			SELECT ai.imageId FROM AlbumImages ai 
			JOIN Albums a ON a.id = ai.albumId 
			%s WHERE ai.albumId = ? AND %s ORDER BY ai.imageId ASC`, visJoins, visWhere)
		args = []any{userID, userID, albumId, userID}
	}

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

func (db SqliteAlbumDatabase) GetAlbumImageCount(albumId AlbumId, userID UserID) (int64, error) {
	var query string
	var args []any
	var count int64

	if userID == 0 {
		query = "SELECT COUNT(*) FROM AlbumImages WHERE albumId = ?"
		args = []any{albumId}
	} else {
		query = fmt.Sprintf(`
			SELECT COUNT(ai.imageId) FROM AlbumImages ai 
			JOIN Albums a ON a.id = ai.albumId 
			%s WHERE ai.albumId = ? AND %s`, visJoins, visWhere)
		args = []any{userID, userID, albumId, userID}
	}

	err := db.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (db SqliteAlbumDatabase) GetAlbumTagCounts(albumId AlbumId, userID UserID) (map[string]int64, error) {
	var query string
	var args []any

	if userID == 0 {
		query = `
			SELECT Tags.tag, COUNT(Tags.id) FROM Tags
			LEFT JOIN ImageTags ON Tags.id = ImageTags.tagId
			LEFT JOIN AlbumImages ON ImageTags.imageId = AlbumImages.imageId
			WHERE AlbumImages.albumId = ? GROUP BY Tags.tag`
		args = []any{albumId}
	} else {
		query = fmt.Sprintf(`
			SELECT t.tag, COUNT(t.id) FROM Tags t
			LEFT JOIN ImageTags it ON t.id = it.tagId
			LEFT JOIN AlbumImages ai ON it.imageId = ai.imageId
			JOIN Albums a ON a.id = ai.albumId
			%s WHERE ai.albumId = ? AND %s GROUP BY t.tag`, visJoins, visWhere)
		args = []any{userID, userID, albumId, userID}
	}

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

	if userID == 0 {
		query = "SELECT id FROM Albums WHERE album_name = ? LIMIT 1"
		args = []any{name}
	} else {
		query = fmt.Sprintf("SELECT a.id FROM Albums a %s WHERE a.album_name = ? AND %s LIMIT 1", visJoins, visWhere)
		args = []any{userID, userID, name, userID}
	}

	err := db.db.QueryRow(query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return AlbumId(id), err
}

func (db SqliteAlbumDatabase) SetAlbumVisibility(albumId AlbumId, mode PrivacyLevel, userID UserID) error {
	res, err := db.db.Exec("UPDATE Albums SET visibility_mode = ? WHERE id = ? AND userId = ?", mode, albumId, userID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("forbidden: album %v not found or not owned by %v", albumId, userID)
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
