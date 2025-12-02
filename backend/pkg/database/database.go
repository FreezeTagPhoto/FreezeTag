package database

import (
	"database/sql"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"math"
	"strings"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

type ImageId int64
type TagId int64

type ImageDatabase interface {
	// Get a list of image IDs corresponding to the provided query
	GetImages(queries.DatabaseQuery) ([]ImageId, error)
	// Get the image filename pointed to by the image ID
	GetImageFile(ImageId) (*string, error)
	// Get the thumbnail data at the given thumbnail level for the image with the given ID
	GetImageThumbnail(ImageId, uint) ([]byte, error)
	// Get the tags attached to an image
	GetImageTags(ImageId) ([]string, error)
	// Get the metadata for an image
	GetImageMetadata(ImageId) (imagedata.Metadata, error)
	// Get all tags matching the given names in the database
	//
	// Note: This function WILL fail silently if you ask for tags that haven't been added. Only use this if you don't care about the tags absolutely existing.
	GetTags([]string) ([]TagId, error)
	// Get all tags present in the database
	GetAllTags() ([]string, error)
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
}

type SqliteImageDatabase struct {
	db *sql.DB
}

//go:embed schema.sql
var schema string

func InitSQLiteImageDatabase(datasource string) (SqliteImageDatabase, error) {
	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		return SqliteImageDatabase{}, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return SqliteImageDatabase{}, err
	}
	return SqliteImageDatabase{db}, nil
}

func cosineDistanceDegrees(latA float64, latB float64, longA float64, longB float64) float64 {
	var phi1, phi2, lambda1, lambda2 = latA * math.Pi / 180., latB * math.Pi / 180., longA * math.Pi / 180., longB * math.Pi / 180.
	return (180. / math.Pi) * math.Acos(math.Sin(phi1)*math.Sin(phi2)+math.Cos(phi1)*math.Cos(phi2)*math.Cos(math.Abs(lambda1-lambda2)))
}

func (db SqliteImageDatabase) GetImages(q queries.DatabaseQuery) ([]ImageId, error) {
	s, as := queries.ImageIdPreparable(q)
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
		var lat, long sql.NullFloat64
		if err := rows.Scan(&id, &lat, &long); err != nil {
			return []ImageId{}, err
		}
		if dq, ok := q.(*queries.ImageQuery); ok && dq.NearLocation != nil {
			// skip adding if there is no location
			if !lat.Valid || !long.Valid {
				continue
			}
			// skip adding if distance is too large
			if cosineDistanceDegrees(lat.Float64, dq.NearLocation[0], long.Float64, dq.NearLocation[1]) > dq.NearLocation[2] {
				continue
			}
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

func (db SqliteImageDatabase) GetAllTags() ([]string, error) {
	rows, err := db.db.Query("SELECT tag FROM Tags")
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
			FileName:    nullStringPtr(fileName),
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

	stmt, err := db.db.Prepare(`INSERT INTO Images (imageFile, dateTaken, dateUploaded, cameraMake, cameraModel, latitude, longitude) values (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	res, err := stmt.Exec(file, datetaken, &dateuploaded, make, model, latitude, longitude)
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
