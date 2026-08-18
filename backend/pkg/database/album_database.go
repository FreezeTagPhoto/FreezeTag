package database

import (
	"database/sql"
	"fmt"
	"log"
)

type GlobalPrivacy uint8
type UserPrivacy uint8

const (
	USER_PRIVATE UserPrivacy = iota
	USER_PUBLIC
	USER_ADMIN
)

const (
	ALBUM_PRIVATE GlobalPrivacy = iota
	ALBUM_PUBLIC
)

type AlbumSharedUser struct {
	UserID     UserID        `json:"user_id"`
	Permission GlobalPrivacy `json:"permission"`
}

type Album struct {
	ID             AlbumID       `json:"id"`
	Name           string        `json:"name"`
	OwnerID        UserID        `json:"owner_id"`
	AlbumPrivacy   GlobalPrivacy `json:"album_privacy"`
	VisbilityLevel UserPrivacy   `json:"visibility_level"`
}

type AlbumDatabase interface {
	CreateAlbum(string, UserID, GlobalPrivacy) (AlbumID, error)
	SetImageAlbum(ImageID, AlbumID, UserID) error
	RemoveAlbum(AlbumID, UserID) error
	RenameAlbum(AlbumID, string, UserID) error
	RemoveImageFromAlbum(ImageID, AlbumID, UserID) error
	GetAssociatedAlbums(ImageID, UserID) ([]Album, error)
	GetAlbumImages(AlbumID, UserID) ([]ImageID, error)
	GetAlbumTagCounts(AlbumID, UserID) (map[string]int64, error)

	GetAlbums(UserID) ([]Album, error)
	GetAlbum(AlbumID, UserID) (Album, error)

	GetAlbumSharedUsers(AlbumID, UserID) ([]AlbumSharedUser, error)
	SetAlbumVisibility(AlbumID, GlobalPrivacy, UserID) error
	SetUserAlbumPermission(AlbumID, UserID, UserPrivacy, UserID) error
}

type SqliteAlbumDatabase struct {
	db *sql.DB
}

func (db SqliteAlbumDatabase) getVisibilityMode(userID UserID) (UserPrivacy, error) {
	var visibility int
	if userID == 0 {
		return USER_ADMIN, nil
	}
	err := db.db.QueryRow("SELECT visibility_mode FROM Users WHERE id = ?", userID).Scan(&visibility)
	return UserPrivacy(visibility), err
}

func (db SqliteAlbumDatabase) userAuthorizedForAlbum(albumID AlbumID, userID UserID) (bool, error) {
	var count int

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return false, err
	}
	switch userVisibility {
	case USER_ADMIN:
		return true, nil
	case USER_PUBLIC:
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
		args := []any{userID, albumID, userID}
		err := db.db.QueryRow(query, args...).Scan(&count)
		return count > 0, err
	case USER_PRIVATE:
		query := ` 
			SELECT COUNT(*) FROM Albums a
			LEFT JOIN AlbumAccess aa 
			ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.id = ?
			AND (aa.access_level > 0 OR a.userId = ?)
		`
		args := []any{userID, albumID, userID}
		err := db.db.QueryRow(query, args...).Scan(&count)
		return count > 0, err
	default:
		return false, fmt.Errorf("invalid visibility mode for user %v", userID)
	}
}

func (db SqliteAlbumDatabase) CreateAlbum(name string, userID UserID, visibilityMode GlobalPrivacy) (AlbumID, error) {
	var id int64
	err := db.db.QueryRow("INSERT INTO Albums (album_name, userId, visibility_mode) VALUES (?, ?, ?) RETURNING id", name, userID, visibilityMode).Scan(&id)
	if err != nil {
		log.Printf("[WARN] failed to create an album: %v", err)
		return 0, fmt.Errorf("could not create album: %s", name)
	}
	return AlbumID(id), nil
}

func (db SqliteAlbumDatabase) SetImageAlbum(imageID ImageID, albumID AlbumID, userID UserID) error {
	canManage, err := db.CanManageAlbum(albumID, userID)
	if err != nil {
		return err
	}
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumID, userID)
	}

	query := "INSERT INTO AlbumImages (albumId, imageId) VALUES (?, ?)"
	_, err = db.db.Exec(query, albumID, imageID)
	return err
}

func (db SqliteAlbumDatabase) RemoveAlbum(albumID AlbumID, userID UserID) error {
	canManage, err := db.CanManageAlbum(albumID, userID)
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumID, userID)
	}
	if err != nil {
		return err
	}
	_, err = db.db.Exec("DELETE FROM Albums WHERE id = ? AND userId = ?", albumID, userID)
	return err
}

func (db SqliteAlbumDatabase) RenameAlbum(albumID AlbumID, newName string, userID UserID) error {
	canManage, err := db.CanManageAlbum(albumID, userID)
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumID, userID)
	}
	if err != nil {
		return err
	}
	res, err := db.db.Exec("UPDATE Albums SET album_name = ? WHERE id = ?", newName, albumID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("album id %q not found or not owned by %v", newName, userID)
	}

	return nil
}

func (db SqliteAlbumDatabase) RemoveImageFromAlbum(imageID ImageID, albumID AlbumID, userID UserID) error {
	canManage, err := db.CanManageAlbum(albumID, userID)
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumID, userID)
	}
	if err != nil {
		return err
	}
	query := `
		DELETE FROM AlbumImages
		WHERE albumId = :aid AND imageId = :iid
		AND EXISTS (SELECT 1 FROM Albums WHERE id = :aid AND userId = :uid)
	`
	_, err = db.db.Exec(query,
		sql.Named("aid", albumID),
		sql.Named("iid", imageID),
		sql.Named("uid", userID))
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
	case USER_ADMIN:
		query = "SELECT id, album_name, userId, visibility_mode, 2 AS vis_level FROM Albums ORDER BY album_name ASC"
	case USER_PUBLIC:
		query = ` 
			SELECT a.id, a.album_name, a.userId, a.visibility_mode,
				   CASE 
					   WHEN a.userId = :uid THEN 2
					   WHEN aa.access_level IS NOT NULL THEN aa.access_level
					   WHEN a.visibility_mode = 2 THEN 2
					   WHEN a.visibility_mode = 1 THEN 1
					   ELSE 0 
				   END AS vis_level
			FROM Albums a 
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE (a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
			   OR aa.access_level > 0 
			   OR a.userId = :uid
			ORDER BY a.album_name ASC
		`
		args = []any{sql.Named("uid", userID)}
	case USER_PRIVATE:
		query = `
			SELECT a.id, a.album_name, a.userId, a.visibility_mode,
				   CASE 
					   WHEN a.userId = :uid THEN 2
					   WHEN aa.access_level IS NOT NULL THEN aa.access_level
					   ELSE 0 
				   END AS vis_level
			FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE aa.access_level > 0 OR a.userId = :uid
			ORDER BY a.album_name ASC
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
		var ownerID UserID
		var albumPrivacy int
		var visLevel int

		if err := rows.Scan(&id, &name, &ownerID, &albumPrivacy, &visLevel); err != nil {
			return nil, err
		}
		albums = append(albums, Album{
			ID:             id,
			Name:           name,
			OwnerID:        ownerID,
			AlbumPrivacy:   GlobalPrivacy(albumPrivacy),
			VisbilityLevel: UserPrivacy(visLevel),
		})
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAssociatedAlbums(imageID ImageID, userID UserID) ([]Album, error) {
	var query string
	var args []any

	userVisibility, err := db.getVisibilityMode(userID)
	if err != nil {
		return nil, err
	}

	switch userVisibility {
	case USER_ADMIN:
		query = `
			SELECT a.id, a.album_name, a.userId, a.visibility_mode, 2 AS vis_level
			FROM AlbumImages ai
			JOIN Albums a ON ai.albumId = a.id
			WHERE ai.imageId = :id
		`
		args = []any{sql.Named("id", imageID)}
	case USER_PUBLIC:
		query = `
			SELECT a.id, a.album_name, a.userId, a.visibility_mode,
				   CASE 
					   WHEN a.userId = :uid THEN 2
					   WHEN aa.access_level IS NOT NULL THEN aa.access_level
					   WHEN a.visibility_mode = 2 THEN 2
					   WHEN a.visibility_mode = 1 THEN 1
					   ELSE 0 
				   END AS vis_level
			FROM AlbumImages ai
			JOIN Albums a ON ai.albumId = a.id
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE ai.imageId = :id
			  AND ( 
				  (a.visibility_mode >= 1 AND (aa.access_level IS NULL OR aa.access_level > 0))
				  OR aa.access_level > 0 
				  OR a.userId = :uid
			  )
		`
		args = []any{sql.Named("id", imageID), sql.Named("uid", userID)}
	case USER_PRIVATE:
		query = `
			SELECT a.id, a.album_name, a.userId, a.visibility_mode,
				   CASE 
					   WHEN a.userId = :uid THEN 2
					   WHEN aa.access_level IS NOT NULL THEN aa.access_level
					   ELSE 0 
				   END AS vis_level
			FROM AlbumImages ai
			JOIN Albums a ON ai.albumId = a.id
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = :uid
			WHERE ai.imageId = :id 
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
		var ownerID UserID
		var albumPrivacy int
		var visLevel int

		if err := rows.Scan(&id, &name, &ownerID, &albumPrivacy, &visLevel); err != nil {
			return nil, err
		}
		albums = append(albums, Album{
			ID:             id,
			Name:           name,
			OwnerID:        ownerID,
			AlbumPrivacy:   GlobalPrivacy(albumPrivacy),
			VisbilityLevel: UserPrivacy(visLevel),
		})
	}
	return albums, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumImages(albumID AlbumID, userID UserID) ([]ImageID, error) {
	ok, err := db.userAuthorizedForAlbum(albumID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumID)
	}
	query := ` 
		SELECT imageId FROM AlbumImages WHERE albumId = :aid
	`
	args := []any{sql.Named("aid", albumID)}
	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var images []ImageID
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		images = append(images, ImageID(id))
	}
	return images, rows.Err()
}

func (db SqliteAlbumDatabase) GetAlbumTagCounts(albumID AlbumID, userID UserID) (map[string]int64, error) {
	var query string
	var args []any

	ok, err := db.userAuthorizedForAlbum(albumID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumID)
	}

	query = `
		SELECT t.tag, COUNT(*) FROM AlbumImages ai
		JOIN ImageTags it ON it.imageId = ai.imageId
		JOIN Tags t ON t.id = it.tagId
		WHERE ai.albumId = :aid
		GROUP BY t.tag
	`
	args = []any{sql.Named("aid", albumID)}

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

func (db SqliteAlbumDatabase) GetAlbumNameByID(albumID AlbumID, userID UserID) (string, error) {
	var name string

	if userID == 0 {
		err := db.db.QueryRow("SELECT album_name FROM Albums WHERE id = ?", albumID).Scan(&name)
		if err == sql.ErrNoRows {
			return "", nil
		}
		return name, err
	}

	ok, err := db.userAuthorizedForAlbum(albumID, userID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	err = db.db.QueryRow("SELECT album_name FROM Albums WHERE id = ?", albumID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return name, err
}

func (db SqliteAlbumDatabase) CanManageAlbum(albumID AlbumID, userID UserID) (bool, error) {
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
	err := db.db.QueryRow(query, userID, albumID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db SqliteAlbumDatabase) GetAlbumOwner(albumID AlbumID) (UserID, error) {
	var ownerID int64
	err := db.db.QueryRow("SELECT userId FROM Albums WHERE id = ?", albumID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("no owner for album with ID: %v", albumID)
	}
	if err != nil {
		return 0, err
	}
	return UserID(ownerID), nil
}

func (db SqliteAlbumDatabase) GetAlbumSharedUserIDs(albumID AlbumID) ([]UserID, error) {
	rows, err := db.db.Query("SELECT userId FROM AlbumAccess WHERE albumId = ? AND access_level > 0 ORDER BY userId ASC", albumID)
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

func (db SqliteAlbumDatabase) GetAlbumSharedUsers(albumID AlbumID, userID UserID) ([]AlbumSharedUser, error) {
	ok, err := db.userAuthorizedForAlbum(albumID, userID) // any user can see shared users if they have access to the album
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumID)
	}
	query := ` 
		SELECT u.id,
			CASE 
				WHEN a.userId = u.id THEN 2 
				WHEN aa.access_level IS NOT NULL THEN aa.access_level 
				WHEN u.visibility_mode = 0 THEN 0 
				WHEN a.visibility_mode = 0 THEN 0 
				WHEN a.visibility_mode = 1 THEN 1 
				ELSE 0 
			END AS permission
		FROM Users u
		JOIN Albums a ON a.id = ?
		LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = u.id
		ORDER BY u.id ASC
	`
	rows, err := db.db.Query(query, albumID)
	if err != nil {
		log.Printf("Error querying shared users for album %v: %v", albumID, err)
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
			Permission: GlobalPrivacy(permission),
		})
	}

	return sharedUsers, rows.Err()
}

func (db SqliteAlbumDatabase) SetAlbumVisibility(albumID AlbumID, mode GlobalPrivacy, userID UserID) error {
	userVis, err := db.getVisibilityMode(userID)
	var query string
	if err != nil {
		return err
	}
	switch userVis {
	case USER_ADMIN:
		query = "UPDATE Albums SET visibility_mode = :vis WHERE id = :aid"
	default: // user owns the album
		query = `
			UPDATE Albums SET visibility_mode = :vis
			WHERE id = :aid AND userId = :uid
		`
	}
	args := []any{sql.Named("vis", mode), sql.Named("aid", albumID), sql.Named("uid", userID)}
	res, err := db.db.Exec(query, args...)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("album %v not found", albumID)
	}
	return nil
}

func (db SqliteAlbumDatabase) SetUserAlbumPermission(albumID AlbumID, targetUser UserID, permission UserPrivacy, requesterID UserID) error {
	canManage, err := db.CanManageAlbum(albumID, requesterID)
	if err != nil {
		return err
	}
	if !canManage {
		return fmt.Errorf("forbidden: album %v not found or not manageable by %v", albumID, requesterID)
	}

	query := `
        INSERT INTO AlbumAccess (albumId, userId, access_level)
        VALUES (?, ?, ?)
        ON CONFLICT(albumId, userId) DO UPDATE SET access_level = excluded.access_level`

	res, err := db.db.Exec(query, albumID, targetUser, permission)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("forbidden: failed to update album %v permissions", albumID)
	}
	return nil
}

func (db SqliteAlbumDatabase) GetAlbum(albumID AlbumID, userID UserID) (Album, error) {
	ok, err := db.userAuthorizedForAlbum(albumID, userID)
	if err != nil {
		return Album{}, err
	}
	if !ok {
		return Album{}, fmt.Errorf("forbidden: user %v does not have access to album %v", userID, albumID)
	}

	userVis, err := db.getVisibilityMode(userID)
	if err != nil {
		return Album{}, err
	}

	var query string
	if userVis == USER_ADMIN {
		query = "SELECT album_name, userId, visibility_mode, 2 AS vis_level FROM Albums WHERE id = ?"
	} else {
		query = `
			SELECT a.album_name, a.userId, a.visibility_mode,
				   CASE 
					   WHEN a.userId = ? THEN 2
					   WHEN aa.access_level IS NOT NULL THEN aa.access_level
					   WHEN a.visibility_mode = 2 THEN 2
					   WHEN a.visibility_mode = 1 THEN 1
					   ELSE 0 
				   END AS vis_level
			FROM Albums a
			LEFT JOIN AlbumAccess aa ON aa.albumId = a.id AND aa.userId = ?
			WHERE a.id = ?
		`
	}

	var name string
	var ownerID UserID
	var albumPrivacy int
	var visLevel int

	if userVis == USER_ADMIN {
		err = db.db.QueryRow(query, albumID).Scan(&name, &ownerID, &albumPrivacy, &visLevel)
	} else {
		err = db.db.QueryRow(query, userID, userID, albumID).Scan(&name, &ownerID, &albumPrivacy, &visLevel)
	}

	if err == sql.ErrNoRows {
		return Album{}, nil
	}
	if err != nil {
		return Album{}, err
	}

	return Album{
		ID:             albumID,
		Name:           name,
		OwnerID:        ownerID,
		AlbumPrivacy:   GlobalPrivacy(albumPrivacy),
		VisbilityLevel: UserPrivacy(visLevel),
	}, nil
}
