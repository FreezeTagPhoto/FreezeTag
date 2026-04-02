package database

import (
	"database/sql"
	"fmt"
)

type PrivacyLevel uint8
type VisbilityLevel uint8

type AlbumSharedUser struct {
	UserID     UserID
	Permission PrivacyLevel
}

type Album struct {
	Id      AlbumID `json:"id"`
	Name    string  `json:"name"`
	OwnerId UserID  `json:"owner_id"`
}

const (
	VisPrivate VisbilityLevel = iota
	VisPublic
	VisAdmin
)

type AlbumDatabase interface {
	CreateAlbum(string, UserID, PrivacyLevel) (AlbumID, error)
	SetImageAlbum(ImageId, AlbumID, UserID) error
	RemoveAlbum(AlbumID, UserID) error
	RenameAlbum(AlbumID, string, UserID) error
	RemoveImageFromAlbum(ImageId, AlbumID, UserID) error
	GetAssociatedAlbums(ImageId, UserID) ([]Album, error)
	GetAlbumImages(AlbumID, UserID) ([]ImageId, error)
	GetAlbumTagCounts(AlbumID, UserID) (map[string]int64, error)

	GetAlbums(UserID) ([]Album, error)
	GetAlbum(AlbumID, UserID) (Album, error)
	// SetAlbumName(AlbumID, string, UserID) error

	GetAlbumSharedUsers(AlbumID, UserID) ([]AlbumSharedUser, error)
	SetAlbumVisibility(AlbumID, PrivacyLevel, UserID) error
	SetUserAlbumPermission(AlbumID, UserID, PrivacyLevel, UserID) error
}

type SqliteAlbumDatabase struct {
	db *sql.DB
}

// func InitSQLiteAlbumDatabase(datasource string) (SqliteAlbumDatabase, error) {
// 	registerExtendedSqlite("sqlite3_extrafunc")
// 	db, err := sql.Open("sqlite3_extrafunc", datasource)
// 	if err != nil {
// 		return SqliteAlbumDatabase{}, err
// 	}
// 	_, err = db.Exec(schema)
// 	if err != nil {
// 		return SqliteAlbumDatabase{}, err
// 	}
// 	return SqliteAlbumDatabase{db}, nil
// }

func (db SqliteAlbumDatabase) getVisibilityMode(userID UserID) (VisbilityLevel, error) {
	var visibility int
	if userID == 0 {
		return VisAdmin, nil
	}
	err := db.db.QueryRow("SELECT visibility_mode FROM Users WHERE id = ?", userID).Scan(&visibility)
	return VisbilityLevel(visibility), err
}

func (db SqliteAlbumDatabase) userAuthorizedForAlbum(albumId AlbumID, userID UserID) (bool, error) {
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

func (db SqliteAlbumDatabase) CreateAlbum(name string, userId UserID, visibilityMode PrivacyLevel) (AlbumID, error) {
	var id int64
	err := db.db.QueryRow("INSERT INTO Albums (album_name, userId, visibility_mode) VALUES (?, ?, ?) RETURNING id", name, userId, visibilityMode).Scan(&id)
	if err != nil {
		return 0, err
	}
	return AlbumID(id), nil
}

func (db SqliteAlbumDatabase) SetImageAlbum(imageId ImageId, albumId AlbumID, userID UserID) error {
	canManage, err := db.CanManageAlbum(albumId, userID)
	if err != nil {
		return err
	}
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumId, userID)
	}

	query := "INSERT INTO AlbumImages (albumId, imageId) VALUES (?, ?)"
	_, err = db.db.Exec(query, albumId, imageId)
	return err
}

func (db SqliteAlbumDatabase) RemoveAlbum(albumId AlbumID, userID UserID) error {
	_, err := db.db.Exec("DELETE FROM Albums WHERE id = ? AND userId = ?", albumId, userID)
	return err
}

func (db SqliteAlbumDatabase) RenameAlbum(albumId AlbumID, newName string, userID UserID) error {
	res, err := db.db.Exec("UPDATE Albums SET album_name = ? WHERE id = ? AND userId = ?", newName, albumId, userID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("album id %q not found or not owned by %v", newName, userID)
	}

	return nil
}

func (db SqliteAlbumDatabase) RemoveImageFromAlbum(imageId ImageId, albumId AlbumID, userID UserID) error {
	_, err := db.db.Exec("DELETE FROM AlbumImages WHERE albumId = ? AND imageId = ? AND EXISTS (SELECT 1 FROM Albums WHERE id = ? AND userId = ?)", albumId, imageId, albumId, userID)
	return err
}

func (db SqliteAlbumDatabase) GetAlbums(userID UserID) ([]Album, error) {
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
			SELECT a.id, a.album_name, a.userId FROM Albums a 
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
			SELECT a.id, a.album_name, a.userId FROM Albums a
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
	defer rows.Close() //nolint:errcheck

	var albums []Album
	for rows.Next() {
		var id AlbumID
		var name string
		var ownerId UserID
		if err := rows.Scan(&id, &name, &ownerId); err != nil {
			return nil, err
		}
		albums = append(albums, Album{Id: id, Name: name, OwnerId: ownerId})
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAssociatedAlbums(imageID ImageId, userID UserID) ([]Album, error) {
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
	defer rows.Close() //nolint:errcheck

	var albums []Album
	for rows.Next() {
		var id AlbumID
		var name string
		var ownerId UserID
		if err := rows.Scan(&id, &name, &ownerId); err != nil {
			return nil, err
		}
		albums = append(albums, Album{Id: id, Name: name, OwnerId: ownerId})
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumImages(albumId AlbumID, userID UserID) ([]ImageId, error) {
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
	defer rows.Close() //nolint:errcheck

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

func (db SqliteAlbumDatabase) GetAlbumTagCounts(albumId AlbumID, userID UserID) (map[string]int64, error) {
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
	defer rows.Close() //nolint:errcheck

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

func (db SqliteAlbumDatabase) GetAlbumNameById(albumId AlbumID, userID UserID) (string, error) {
	var name string

	if userID == 0 {
		err := db.db.QueryRow("SELECT album_name FROM Albums WHERE id = ?", albumId).Scan(&name)
		if err == sql.ErrNoRows {
			return "", nil
		}
		return name, err
	}

	ok, err := db.userAuthorizedForAlbum(albumId, userID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	err = db.db.QueryRow("SELECT album_name FROM Albums WHERE id = ?", albumId).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return name, err
}

func (db SqliteAlbumDatabase) GetAlbumIdByName(name string, userID UserID) (AlbumID, error) {
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
	return AlbumID(id), nil
}

func (db SqliteAlbumDatabase) CanManageAlbum(albumId AlbumID, userID UserID) (bool, error) {
	if userID == 0 {
		return true, nil
	}

	var count int
	query := `
		SELECT COUNT(*) FROM Albums a
		LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = ?
		WHERE a.id = ?
		AND (a.userId = ? OR aa.access_level = 2)
	`
	err := db.db.QueryRow(query, userID, albumId, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db SqliteAlbumDatabase) GetAlbumOwner(albumId AlbumID) (UserID, error) {
	var ownerID int64
	err := db.db.QueryRow("SELECT userId FROM Albums WHERE id = ?", albumId).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return UserID(ownerID), nil
}

func (db SqliteAlbumDatabase) GetAlbumSharedUserIDs(albumId AlbumID) ([]UserID, error) {
	rows, err := db.db.Query("SELECT userId FROM AlbumAccess WHERE albumId = ? AND access_level > 0 ORDER BY userId ASC", albumId)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var userIDs []UserID
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, UserID(id))
	}
	return userIDs, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumSharedUsers(albumId AlbumID, userId UserID) ([]AlbumSharedUser, error) {
	ok, err := db.userAuthorizedForAlbum(albumId, userId) // any user can see shared users if they have access to the album
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userId, albumId)
	}
	query := ` 
		SELECT userId, access_level FROM AlbumAccess WHERE albumId = ? AND access_level > 0 ORDER BY userId ASC
	`
	rows, err := db.db.Query(query, albumId)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	sharedUsers := make([]AlbumSharedUser, 0)
	for rows.Next() {
		var id int64
		var permission int
		if err := rows.Scan(&id, &permission); err != nil {
			return nil, err
		}
		sharedUsers = append(sharedUsers, AlbumSharedUser{
			UserID:     UserID(id),
			Permission: PrivacyLevel(permission),
		})
	}

	return sharedUsers, rows.Err()
}

func (db SqliteAlbumDatabase) SetAlbumVisibility(albumId AlbumID, mode PrivacyLevel, userID UserID) error {
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

func (db SqliteAlbumDatabase) SetUserAlbumPermission(albumId AlbumID, targetUser UserID, permission PrivacyLevel, requesterID UserID) error {
	canManage, err := db.CanManageAlbum(albumId, requesterID)
	if err != nil {
		return err
	}
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumId, requesterID)
	}

	query := `
        INSERT INTO AlbumAccess (albumId, userId, access_level)
        VALUES (?, ?, ?)
        ON CONFLICT(albumId, userId) DO UPDATE SET access_level = excluded.access_level`

	res, err := db.db.Exec(query, albumId, targetUser, permission)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("forbidden: failed to update album %v permissions", albumId)
	}
	return nil
}

func (db SqliteAlbumDatabase) GetAlbum(albumId AlbumID, userID UserID) (Album, error) {
	var name string
	var ownerId UserID
	ok, err := db.userAuthorizedForAlbum(albumId, userID) // any user can see shared users if they have access to the album
	if err != nil {
		return Album{}, err
	}
	if !ok {
		return Album{}, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumId)
	}
	err = db.db.QueryRow("SELECT album_name, userId FROM Albums WHERE id = ?", albumId).Scan(&name, &ownerId)
	if err == sql.ErrNoRows {
		return Album{}, nil
	}
	if err != nil {
		return Album{}, err
	}
	return Album{
		Id:      albumId,
		Name:    name,
		OwnerId: ownerId,
	}, nil
}
