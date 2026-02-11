package database

import (
	"database/sql"
	"fmt"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"regexp"
	"strings"
	"time"

	_ "embed"
)

type ImageId int64
type TagId int64

type ImageDatabase interface {
	// Get a list of image IDs corresponding to the provided query
	GetImages(queries.DatabaseQuery) ([]ImageId, error)
	// Get a list of image IDs corresponding to the provided query and order
	GetImagesOrder(queries.DatabaseQuery, queries.SortField, queries.SortOrder) ([]ImageId, error)
	// Get the image filename pointed to by the image ID
	GetImageFile(ImageId) (*string, error)
	// Get the lowest suffix number that doesn't overlap with an existing image
	GetNonOverlappingSuffix(string) (int, error)
	// Get the thumbnail data at the given thumbnail level for the image with the given ID
	GetImageThumbnail(ImageId, uint) ([]byte, error)
	// Get the tags attached to an image
	GetImageTags(ImageId) ([]string, error)
	// Get the metadata for an image
	GetImageMetadata(ImageId) (imagedata.Metadata, error)
	// Get the resolution of an image
	GetImageResolution(ImageId) (int, int, error)
	// Get all tags matching the given names in the database
	//
	// Note: This function WILL fail silently if you ask for tags that haven't been added. Only use this if you don't care about the tags absolutely existing.
	GetTags([]string) ([]TagId, error)
	// Get all tags present in the database
	GetAllTags() (map[string]int64, error)
	// Get the thumbnail sizes an image has
	GetImageThumbnailSizes(ImageId) ([]int, error)
	// Add an image file and its metadata to the database
	//
	// returns: the database ID of the image
	AddImage(string, imagedata.Data) (ImageId, error)
	// Add a thumbnail of the given size with the given data to the database
	//
	// returns: whether the thumbnail was added successfully
	AddImageThumbnail(ImageId, int, []byte) (bool, error)
	// Add a set of tags to the image with the given id
	//
	// returns: the number of tags successfully added
	AddImageTags(ImageId, []string) (int, error)
	// Add a set of tags by name to the database
	//
	// returns: the IDs of the tags with those names
	AddTags([]string) ([]TagId, error)
	// Remove a set of tags by name from the database
	//
	// returns: the number of tags actually removed
	RemoveTags([]string) (int, error)
	// Remove an image from the database
	//
	// returns: whether an image was removed
	RemoveImage(ImageId) (bool, error)
	// Remove tags from an image
	//
	// returns: the number of tags successfully removed
	RemoveImageTags(ImageId, []string) (int, error)
	// Remove a thumbnail with the given size from an image
	//
	// returns: whether the thumbnail was removed
	RemoveImageThumbnail(ImageId, int) (bool, error)
	// Gets the counts of each tag in the provided list
	//
	// returns: a map of tag name to count
	GetTagCounts([]string) (map[string]int64, error)
}

type SqliteImageDatabase struct {
	db *sql.DB
}

//go:embed schema.sql
var schema string

func InitSQLiteImageDatabase(datasource string) (SqliteImageDatabase, error) {
	registerExtendedSqlite("sqlite3_extrafunc")
	db, err := sql.Open("sqlite3_extrafunc", datasource)
	if err != nil {
		return SqliteImageDatabase{}, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return SqliteImageDatabase{}, err
	}
	return SqliteImageDatabase{db}, nil
}

func (db SqliteImageDatabase) GetImages(q queries.DatabaseQuery) ([]ImageId, error) {
	return db.GetImagesOrder(q, queries.DateAdded, queries.Descending)
}

func (db SqliteImageDatabase) GetImagesOrder(q queries.DatabaseQuery, sf queries.SortField, so queries.SortOrder) ([]ImageId, error) {
	s, as := queries.ImageIdPreparable(q, sf, so)
	stmt, err := db.db.Prepare(s)
	if err != nil {
		return []ImageId{}, err
	}
	defer stmt.Close() //nolint:errcheck
	rows, err := stmt.Query(as...)
	if err != nil {
		return []ImageId{}, err
	}
	defer rows.Close() //nolint:errcheck
	ids := []ImageId{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []ImageId{}, err
		}
		var id int
		if err := rows.Scan(&id); err != nil {
			return []ImageId{}, err
		}
		ids = append(ids, ImageId(id))
	}
	return ids, nil
}

func (db SqliteImageDatabase) GetImageFile(id ImageId) (*string, error) {
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
	exp := fmt.Sprintf("%s[0-9]*", regexp.QuoteMeta(name))
	rows, err := db.db.Query("SELECT CAST(SUBSTR(imageFile, ?) AS INTEGER) FROM Images WHERE imageFile REGEXP ? ORDER BY CAST(SUBSTR(imageFile, ?) AS INTEGER) ASC", len(name)+1, exp, len(name)+1)
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

func (db SqliteImageDatabase) GetImageThumbnail(id ImageId, size uint) ([]byte, error) {
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

func (db SqliteImageDatabase) GetImageTags(id ImageId) ([]string, error) {
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

func (db SqliteImageDatabase) GetTags(tags []string) ([]TagId, error) {
	var value strings.Builder
	value.WriteByte('(')
	for i := range tags {
		value.WriteString("?")
		if i < len(tags)-1 {
			value.WriteString(", ")
		}
	}
	value.WriteByte(')')
	params := make([]any, len(tags))
	for i, tag := range tags {
		params[i] = tag
	}
	rows, err := db.db.Query("SELECT id FROM Tags WHERE tag IN "+value.String(), params...)
	if err != nil {
		return []TagId{}, err
	}
	defer rows.Close() //nolint:errcheck
	ids := []TagId{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []TagId{}, err
		}
		var id TagId
		if err := rows.Scan(&id); err != nil {
			return []TagId{}, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (db SqliteImageDatabase) AddTags(tags []string) ([]TagId, error) {
	var value strings.Builder
	for i := range tags {
		value.WriteString("(?)")
		if i < len(tags)-1 {
			value.WriteString(", ")
		}
	}
	params := make([]any, len(tags))
	for i, tag := range tags {
		params[i] = tag
	}
	rows, err := db.db.Query("INSERT OR IGNORE INTO Tags (tag) VALUES "+value.String()+" RETURNING id", params...)
	if err != nil {
		return []TagId{}, err
	}
	defer rows.Close() //nolint:errcheck
	ids := []TagId{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return []TagId{}, err
		}
		var id TagId
		if err := rows.Scan(&id); err != nil {
			return []TagId{}, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (db SqliteImageDatabase) RemoveTags(tags []string) (int, error) {
	var value strings.Builder
	value.WriteByte('(')
	for i := range tags {
		value.WriteString("?")
		if i < len(tags)-1 {
			value.WriteString(", ")
		}
	}
	value.WriteByte(')')
	params := make([]any, len(tags))
	for i, tag := range tags {
		params[i] = tag
	}
	res, err := db.db.Exec("DELETE FROM Tags WHERE tag IN "+value.String(), params...)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func (db SqliteImageDatabase) GetAllTags() (map[string]int64, error) {
	rows, err := db.db.Query("SELECT tag, COUNT(Tags.id) as count FROM Tags LEFT JOIN ImageTags ON Tags.id = ImageTags.tagId GROUP BY tag")
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

func (db SqliteImageDatabase) GetImageMetadata(id ImageId) (imagedata.Metadata, error) {
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

func (db SqliteImageDatabase) GetImageResolution(id ImageId) (w int, h int, err error) {
	rows, err := db.db.Query("SELECT width, height FROM Images WHERE id = ?", id)
	if err != nil {
		return
	}
	defer rows.Close() //nolint:errcheck
	if rows.Next() {
		var width sql.NullInt32
		var height sql.NullInt32
		if err = rows.Scan(&width, &height); err != nil {
			return
		}
		if !width.Valid || !height.Valid {
			return 0, 0, fmt.Errorf("image has null resolution (this really shouldn't happen)")
		}
		return int(width.Int32), int(height.Int32), nil
	}
	return 0, 0, nil
}

func (db SqliteImageDatabase) GetImageThumbnailSizes(id ImageId) ([]int, error) {
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

func (db SqliteImageDatabase) AddImage(file string, data imagedata.Data) (ImageId, error) {
	var datetaken *int64
	dateuploaded := time.Now().Unix()
	var make *string
	var model *string
	var latitude *float64
	var longitude *float64
	if data.DateCreated != nil {
		formatted := data.DateCreated.Unix()
		datetaken = &formatted
	}
	if data.Cam != nil {
		make = &data.Cam.Manufacturer
		model = &data.Cam.Model
	}
	if data.Geo != nil {
		latitude = &data.Geo.Lat
		longitude = &data.Geo.Lon
	}

	stmt, err := db.db.Prepare(`INSERT INTO Images (imageFile, dateTaken, dateUploaded, cameraMake, cameraModel, latitude, longitude, width, height) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	res, err := stmt.Exec(file, datetaken, &dateuploaded, make, model, latitude, longitude, data.Width, data.Height)
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return ImageId(id), nil
}

func (db SqliteImageDatabase) AddImageThumbnail(id ImageId, size int, data []byte) (bool, error) {
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

func (db SqliteImageDatabase) AddImageTags(id ImageId, tags []string) (int, error) {
	modified := 0
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
	}
	if _, err = db.AddTags(tags); err != nil {
		return 0, err
	}
	tagIds, err := db.GetTags(tags)
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO ImageTags (imageId, tagId) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	for _, tagId := range tagIds {
		res, err := stmt.Exec(id, tagId)
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

func (db SqliteImageDatabase) RemoveImage(id ImageId) (bool, error) {
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

func (db SqliteImageDatabase) RemoveImageTags(id ImageId, tags []string) (int, error) {
	if len(tags) == 0 {
		return 0, nil
	}
	var query strings.Builder
	args := []any{id}
	query.WriteString("DELETE FROM ImageTags WHERE imageId = ? AND tagId IN (")
	tagIds, err := db.GetTags(tags)
	if err != nil {
		return 0, err
	}
	for i, id := range tagIds {
		args = append(args, id)
		query.WriteRune('?')
		if i < len(tagIds)-1 {
			query.WriteString(", ")
		}
	}
	query.WriteRune(')')
	res, err := db.db.Exec(query.String(), args...)
	if err != nil {
		return 0, err
	}
	mod, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(mod), nil
}

func (db SqliteImageDatabase) RemoveImageThumbnail(id ImageId, size int) (bool, error) {
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

func (db SqliteImageDatabase) GetTagCounts(imageIds []string) (map[string]int64, error) {
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
