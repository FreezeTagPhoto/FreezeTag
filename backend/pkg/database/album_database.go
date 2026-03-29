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

func (db SqliteAlbumDatabase) RemoveImageFromAlbum(imageId ImageId, albumId AlbumId, userID UserID) error {
	_, err := db.db.Exec("DELETE FROM AlbumImages WHERE albumId = ? AND imageId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)", albumId, imageId, albumId, userID)
	return err
}

func (db SqliteAlbumDatabase) GetAlbumIds(userID UserID) ([]AlbumId, error) {
	rows, err := db.db.Query("SELECT id FROM Albums WHERE userId = ?", userID)
	if err != nil {
		return []AlbumId{}, err
	}
	defer rows.Close() //nolint:errcheck
	albums := []AlbumId{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []AlbumId{}, err
		}
		var id int64
		if err := rows.Scan(&id); err != nil {
			return []AlbumId{}, err
		}
		albums = append(albums, AlbumId(id))
	}
	return albums, nil
}

func (db SqliteAlbumDatabase) GetAlbumNames(userID UserID) ([]string, error) {
	query := `
		SELECT a.album_name
		FROM Albums a
		LEFT JOIN AlbumAccess aa
			ON aa.albumId = a.id AND aa.userId = ?
		WHERE
			a.userId = ?
			OR (
				CASE
					WHEN aa.access_level IS NOT NULL THEN aa.access_level > 0
					WHEN a.visibility_mode = 0 THEN 0
					ELSE 1
				END
			)
		ORDER BY a.album_name ASC
		`
	rows, err := db.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var names []string
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (db SqliteAlbumDatabase) GetImageAlbumNames(imageID ImageId, userID UserID) ([]string, error) {
	query := `
		SELECT a.album_name
		FROM Albums a
		JOIN AlbumImages ai
			ON ai.albumId = a.id
		LEFT JOIN AlbumAccess aa
			ON aa.albumId = a.id AND aa.userId = ?
		WHERE
			ai.imageId = ?
			AND (
				a.userId = ?
				OR (
					CASE
						WHEN aa.access_level IS NOT NULL THEN aa.access_level > 0
						WHEN a.visibility_mode = 0 THEN 0
						ELSE 1
					END
				)
			)
		ORDER BY a.album_name ASC
	`

	rows, err := db.db.Query(query, userID, imageID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var names []string
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (db SqliteAlbumDatabase) GetAlbumImages(albumId AlbumId, userID UserID) ([]ImageId, error) {
	rows, err := db.db.Query("SELECT imageId FROM AlbumImages WHERE albumId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)", albumId, albumId, userID)
	if err != nil {
		return []ImageId{}, err
	}
	defer rows.Close() //nolint:errcheck
	images := []ImageId{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []ImageId{}, err
		}
		var id int64
		if err := rows.Scan(&id); err != nil {
			return []ImageId{}, err
		}
		images = append(images, ImageId(id))
	}
	return images, nil
}

func (db SqliteAlbumDatabase) GetAlbumImageCount(albumId AlbumId, userID UserID) (int64, error) {
	var count int64
	err := db.db.QueryRow("SELECT COUNT(*) FROM AlbumImages WHERE albumId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)", albumId, albumId, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db SqliteAlbumDatabase) GetAlbumTagCounts(albumId AlbumId, userID UserID) (map[string]int64, error) {
	query := `
		SELECT tag, COUNT(Tags.id) as count FROM Tags
		LEFT JOIN ImageTags on Tags.id = ImageTags.tagId
		LEFT JOIN AlbumImages on ImageTags.imageId = AlbumImages.imageId
		WHERE AlbumImages.albumId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)
		GROUP BY Tags.tag
		`

	rows, err := db.db.Query(query, albumId, albumId, userID)
	if err != nil {
		return map[string]int64{}, err
	}
	defer rows.Close() //nolint:errcheck
	counts := make(map[string]int64)
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return map[string]int64{}, err
		}
		var tag string
		var count int64
		if err := rows.Scan(&tag, &count); err != nil {
			return map[string]int64{}, err
		}
		counts[tag] = count
	}
	return counts, nil
}

func (db SqliteAlbumDatabase) GetAlbumIdByName(name string, userID UserID) (AlbumId, error) {
	var id int64
	err := db.db.QueryRow("SELECT id FROM Albums WHERE album_name = ? AND userId = ?", name, userID).Scan(&id)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	albumId := AlbumId(id)
	return albumId, nil
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
