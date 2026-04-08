package database

import (
	"database/sql"
	"fmt"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "embed"
)

type ImageID uint64
type TagID uint64
type AlbumID uint64

type ImageDatabase interface {
	// Get a list of all image IDs corresponding to the
	//
	// Only returns a list of image IDs that the user has access to.
	GetImages(queries.DatabaseQuery, UserID) ([]ImageID, error)
	// Get a list of image IDs corresponding to the provided query and order
	GetImagesOrder(
		queries.DatabaseQuery,
		queries.SortField,
		queries.SortOrder,
		UserID) ([]ImageID, error)
	// Get a list of image IDs corresponding to the provided query, order, page size, and page number
	GetImagesOrderPaged(
		queries.DatabaseQuery,
		queries.SortField,
		queries.SortOrder,
		uint,
		uint,
		UserID) ([]ImageID, error)
	// Get the image filename pointed to by the image ID
	GetImageFile(ImageID) (*string, error)
	// Get the lowest suffix number that doesn't overlap with an existing image
	GetNonOverlappingSuffix(string) (int, error)
	// Get the thumbnail data at the given thumbnail level for the image with the given ID
	GetImageThumbnail(ImageID, uint) ([]byte, error)
	// Get the tags attached to an image
	GetImageTags(ImageID) ([]string, error)
	// Get the metadata for an image
	GetImageMetadata(ImageID) (imagedata.Metadata, error)
	// Get the resolution of an image
	GetImageResolution(ImageID) (int, int, error)
	// Get all tags matching the given names in the database
	//
	// Note: This function WILL fail silently if you ask for tags that haven't been added. Only use this if you don't care about the tags absolutely existing.
	GetTags([]string) ([]TagID, error)
	// Get all tags present in the database
	GetAllTags() (map[string]int64, error)
	// Get the thumbnail sizes an image has
	GetImageThumbnailSizes(ImageID) ([]int, error)
	// Add an image file and its metadata to the database
	//
	// returns: the database ID of the image
	AddImage(string, imagedata.Data) (ImageID, error)
	// Add a thumbnail of the given size with the given data to the database
	//
	// returns: whether the thumbnail was added successfully
	AddImageThumbnail(ImageID, int, []byte) (bool, error)
	// Add a set of tags to the image with the given id
	//
	// returns: the number of tags successfully added
	AddImageTags(ImageID, []string) (int, error)
	// Add a set of tags by name to the database
	//
	// returns: the IDs of the tags with those names
	AddTags([]string) ([]TagID, error)
	// Remove a set of tags by name from the database
	//
	// returns: the number of tags actually removed
	RemoveTags([]string) (int, error)
	// Remove an image from the database
	//
	// returns: whether an image was removed
	RemoveImage(ImageID) (bool, error)
	// Remove tags from an image
	//
	// returns: the number of tags successfully removed
	RemoveImageTags(ImageID, []string) (int, error)
	// Remove a thumbnail with the given size from an image
	//
	// returns: whether the thumbnail was removed
	RemoveImageThumbnail(ImageID, int) (bool, error)
	// Gets the tag count for each image ID in the provided list
	//
	// returns: a map of tag name to count
	GetTagCounts([]ImageID) (map[string]int64, error)
}

type SqliteImageDatabase struct {
	db *sql.DB
}

func (db SqliteImageDatabase) GetImages(q queries.DatabaseQuery, UserID UserID) ([]ImageID, error) {
	return db.GetImagesOrder(q, queries.DateAdded, queries.Descending, UserID)
}

func (db SqliteImageDatabase) GetImagesOrder(q queries.DatabaseQuery, sf queries.SortField, so queries.SortOrder, UserID UserID) ([]ImageID, error) {
	return db.GetImagesOrderPaged(q, sf, so, 0, 0, UserID)
}

func (db SqliteImageDatabase) GetImagesOrderPaged(q queries.DatabaseQuery, sf queries.SortField, so queries.SortOrder, pageSize uint, pageNum uint, userID UserID) ([]ImageID, error) {
	stmt, stmtArgs := q.StatementWithArgs()
	var args []any
	var query string

	if userID == 0 {
		// If userID is 0, we are assuming this is a plugin/other internal tool
		// and dont apply filters because that would be a pain and this needs to be shipped
		query = fmt.Sprintf("SELECT id FROM Images WHERE %s", stmt)
		args = append(args, stmtArgs...)
	} else {
		query = `
        WITH UserPermissions AS (
            SELECT visibility_mode FROM Users WHERE id = ?
        ),
        AccessibleAlbums AS (
            SELECT albumId 
            FROM AlbumAccess 
            WHERE userId = ? AND access_level > 0
        )
        SELECT DISTINCT i.id 
        FROM Images i
        CROSS JOIN UserPermissions up
        LEFT JOIN AlbumImages ai ON i.id = ai.imageId
        WHERE (` + stmt + `) AND (
            up.visibility_mode > 0
            OR 
            ai.albumId IN (SELECT albumId FROM AccessibleAlbums)
        )`
		args = append(args, userID, userID)
		args = append(args, stmtArgs...)
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sf.String(), so.String())
	if pageSize != 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, pageSize, pageSize*pageNum)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return []ImageID{}, err
	}
	defer rows.Close() //nolint:errcheck
	ids := []ImageID{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []ImageID{}, err
		}
		var id int64
		if err := rows.Scan(&id); err != nil {
			return []ImageID{}, err
		}
		ids = append(ids, ImageID(id))
	}
	return ids, nil
}

func (db SqliteImageDatabase) GetImageFile(id ImageID) (*string, error) {
	rows, err := db.db.Query("SELECT imageFile FROM Images WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		return &name, nil
	}
	return nil, nil
}

func (db SqliteImageDatabase) GetNonOverlappingSuffix(name string) (int, error) {
	var exp string
	if ext := filepath.Ext(name); ext == "" {
		exp = fmt.Sprintf("%s([0-9]*)", regexp.QuoteMeta(name))
	} else {
		base := strings.TrimSuffix(name, ext)
		exp = fmt.Sprintf("%s([0-9]*)%s", regexp.QuoteMeta(base), regexp.QuoteMeta(ext))
	}
	rows, err := db.db.Query("SELECT CAST(rextract(?, imageFile) AS INTEGER) FROM Images WHERE imageFile REGEXP ? ORDER BY CAST(rextract(?, imageFile) AS INTEGER) ASC", exp, exp, exp)
	if err != nil {
		return 0, err
	}
	defer rows.Close() //nolint:errcheck
	lowest := 0
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		var suffix int
		if err := rows.Scan(&suffix); err != nil {
			return 0, err
		}
		if suffix != lowest {
			break
		}
		lowest++
	}
	return lowest, nil
}

func (db SqliteImageDatabase) GetImageThumbnail(id ImageID, size uint) ([]byte, error) {
	rows, err := db.db.Query("SELECT thumbnailData FROM Thumbnails WHERE imageId = ? AND thumbnailSize = ?", id, size)
	if err != nil {
		return []byte{}, err
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		if err := rows.Err(); err != nil {
			return []byte{}, err
		}
		data := []byte{}
		if err := rows.Scan(&data); err != nil {
			return []byte{}, err
		}
		return data, nil
	}
	return nil, nil
}

func (db SqliteImageDatabase) GetImageTags(id ImageID) ([]string, error) {
	rows, err := db.db.Query("SELECT tag FROM Tags WHERE id IN (SELECT tagId FROM ImageTags WHERE imageId = ?)", id)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close() //nolint:errcheck
	tags := []string{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []string{}, err
		}
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return []string{}, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (db SqliteImageDatabase) GetTags(tags []string) ([]TagID, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return []TagID{}, err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt, err := tx.Prepare("SELECT id FROM Tags WHERE tag = ?")
	if err != nil {
		return []TagID{}, err
	}
	defer stmt.Close() //nolint:errcheck
	ids := make([]TagID, 0, len(tags))
	for _, tag := range tags {
		var id int64
		if err := stmt.QueryRow(tag).Scan(&id); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return []TagID{}, fmt.Errorf("failed to get a tag ID: %w", err)
		}
		ids = append(ids, TagID(id))
	}
	if err := tx.Commit(); err != nil {
		return []TagID{}, err
	}
	return ids, nil
}

func (db SqliteImageDatabase) AddTags(tags []string) ([]TagID, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return []TagID{}, err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO Tags (tag) VALUES (?) RETURNING id")
	if err != nil {
		return []TagID{}, err
	}
	defer stmt.Close() //nolint:errcheck
	ids := make([]TagID, 0, len(tags))
	for _, tag := range tags {
		var id int64
		if err := stmt.QueryRow(tag).Scan(&id); err != nil {
			if err == sql.ErrNoRows {
				continue
			} 
			return []TagID{}, err
		}
		ids = append(ids, TagID(id))
	}
	if err := tx.Commit(); err != nil {
		return []TagID{}, err
	}
	return ids, nil
}

func (db SqliteImageDatabase) RemoveTags(tags []string) (int, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt, err := tx.Prepare("DELETE FROM Tags WHERE tag = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	modified := 0
	for _, tag := range tags {
		res, err := stmt.Exec(tag)
		if err != nil {
			return 0, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		modified += int(rows)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return modified, nil
}

func (db SqliteImageDatabase) GetAllTags() (map[string]int64, error) {
	rows, err := db.db.Query("SELECT tag, COUNT(ImageTags.tagId) as count FROM Tags LEFT JOIN ImageTags ON Tags.id = ImageTags.tagId GROUP BY tag")
	if err != nil {
		return map[string]int64{}, err
	}
	defer rows.Close() //nolint:errcheck
	tags := map[string]int64{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return map[string]int64{}, err
		}
		var tag string
		var count int64
		if err := rows.Scan(&tag, &count); err != nil {
			return map[string]int64{}, err
		}
		tags[tag] = count
	}
	return tags, nil
}

func (db SqliteImageDatabase) GetImageMetadata(id ImageID) (imagedata.Metadata, error) {
	rows, err := db.db.Query("SELECT imageFile, dateTaken, dateUploaded, cameraMake, cameraModel, latitude, longitude FROM Images WHERE id = ?", id)
	if err != nil {
		return imagedata.Metadata{}, err
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		if err := rows.Err(); err != nil {
			return imagedata.Metadata{}, err
		}
		var fileName sql.NullString
		var dateTaken sql.NullInt64
		var dateUploaded sql.NullInt64
		var cameraMake sql.NullString
		var cameraModel sql.NullString
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		if err := rows.Scan(&fileName, &dateTaken, &dateUploaded, &cameraMake, &cameraModel, &latitude, &longitude); err != nil {
			return imagedata.Metadata{}, err
		}
		return imagedata.Metadata{
			FileName:     nullStringPtr(fileName),
			DateTaken:    nullInt64Ptr(dateTaken),
			DateUploaded: nullInt64Ptr(dateUploaded),
			CameraMake:   nullStringPtr(cameraMake),
			CameraModel:  nullStringPtr(cameraModel),
			Latitude:     nullFloat64Ptr(latitude),
			Longitude:    nullFloat64Ptr(longitude),
		}, nil
	}
	return imagedata.Metadata{}, nil
}

func (db SqliteImageDatabase) GetImageResolution(id ImageID) (w int, h int, err error) {
	err = db.db.QueryRow("SELECT width, height FROM Images WHERE id = ?", id).Scan(&w, &h)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return
}

func (db SqliteImageDatabase) GetImageThumbnailSizes(id ImageID) ([]int, error) {
	rows, err := db.db.Query("SELECT thumbnailSize FROM Thumbnails WHERE imageId = ?", id)
	if err != nil {
		return []int{}, err
	}
	defer rows.Close() //nolint:errcheck
	sizes := []int{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []int{}, err
		}
		var size int
		if err := rows.Scan(&size); err != nil {
			return []int{}, err
		}
		sizes = append(sizes, size)
	}
	return sizes, nil
}

func (db SqliteImageDatabase) AddImage(file string, data imagedata.Data) (ImageID, error) {
	var datetaken *int64
	dateuploaded := time.Now().Unix()
	var deviceMake *string
	var model *string
	var latitude *float64
	var longitude *float64
	if data.DateCreated != nil {
		formatted := data.DateCreated.Unix()
		datetaken = &formatted
	}
	if data.Cam != nil {
		deviceMake = &data.Cam.Manufacturer
		model = &data.Cam.Model
	}
	if data.Geo != nil {
		latitude = &data.Geo.Lat
		longitude = &data.Geo.Lon
	}
	var id int64
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	if err := tx.QueryRow(
		"INSERT INTO Images (imageFile, dateTaken, dateUploaded, cameraMake, cameraModel, latitude, longitude, width, height) values (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id",
		file, datetaken, &dateuploaded, deviceMake, model, latitude, longitude, data.Width, data.Height,
	).Scan(&id); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return ImageID(id), nil
}

func (db SqliteImageDatabase) AddImageThumbnail(id ImageID, size int, data []byte) (bool, error) {
	res, err := db.db.Exec("INSERT OR IGNORE INTO Thumbnails (imageId, thumbnailSize, thumbnailData) VALUES (?, ?, ?)", id, size, data)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows != 0, nil
}

func (db SqliteImageDatabase) AddImageTags(id ImageID, tags []string) (int, error) {
	modified := 0
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err = db.AddTags(tags); err != nil {
		return 0, err
	}
	tagIDs, err := db.GetTags(tags)
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO ImageTags (imageId, tagId) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	for _, tagID := range tagIDs {
		res, err := stmt.Exec(id, tagID)
		if err != nil {
			return 0, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		modified += int(rows)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return modified, nil
}

func (db SqliteImageDatabase) RemoveImage(id ImageID) (bool, error) {
	res, err := db.db.Exec("DELETE FROM Images WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows != 0, nil
}

func (db SqliteImageDatabase) RemoveImageTags(id ImageID, tags []string) (int, error) {
	if len(tags) == 0 {
		return 0, nil
	}
	tagIDs, err := db.GetTags(tags)
	if err != nil {
		return 0, err
	}
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt, err := tx.Prepare("DELETE FROM ImageTags WHERE imageId = ? AND tagId = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	modified := 0
	for _, tagID := range tagIDs {
		res, err := stmt.Exec(id, tagID)
		if err != nil {
			return 0, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		modified += int(rows)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return modified, nil
}

func (db SqliteImageDatabase) RemoveImageThumbnail(id ImageID, size int) (bool, error) {
	res, err := db.db.Exec("DELETE FROM Thumbnails WHERE imageId = ? AND thumbnailSize = ?", id, size)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows != 0, nil
}

func (db SqliteImageDatabase) GetTagCounts(imageIds []ImageID) (map[string]int64, error) {
	if len(imageIds) == 0 {
		return map[string]int64{}, nil
	}
	var value strings.Builder
	params := make([]any, len(imageIds))

	value.WriteByte('(')
	for i, id := range imageIds {
		value.WriteString("?")
		if i < len(imageIds)-1 {
			value.WriteString(", ")
		}
		params[i] = id
	}
	value.WriteByte(')')

	query := "SELECT tag, COUNT(Tags.id) as count FROM Tags LEFT JOIN ImageTags on Tags.id = ImageTags.tagId WHERE ImageTags.imageId IN " + value.String() + " GROUP BY Tags.tag"
	rows, err := db.db.Query(query, params...)
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

// helper functions to convert sql.Null* to pointers
func nullInt64Ptr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

func nullStringPtr(n sql.NullString) *string {
	if n.Valid {
		return &n.String
	}
	return nil
}

func nullFloat64Ptr(n sql.NullFloat64) *float64 {
	if n.Valid {
		return &n.Float64
	}
	return nil
}
