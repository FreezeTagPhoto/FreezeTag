package database

import (
	"database/sql"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"strings"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

type ImageId int64

type ImageDatabase interface {
	// Get a list of image IDs corresponding to the provided query
	GetImages(queries.DatabaseQuery) ([]ImageId, error)
	// Get the image filename pointed to by the image ID
	GetImageFile(ImageId) (*string, error)
	// Get the thumbnail data at the given thumbnail level for the image with the given ID
	GetImageThumbnail(ImageId, int) ([]byte, error)
	// Get the tags attached to an image
	GetImageTags(ImageId) ([]string, error)
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

func (db SqliteImageDatabase) GetImageThumbnail(id ImageId, size int) ([]byte, error) {
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
	rows, err := db.db.Query("SELECT tag FROM Tags WHERE imageId = ?", id)
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
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO Tags (imageId, tag) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() //nolint:errcheck
	for _, tag := range tags {
		res, err := stmt.Exec(id, tag)
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
	query.WriteString("DELETE FROM Tags WHERE imageId = ? AND tag IN (")
	for i, tag := range tags {
		args = append(args, tag)
		query.WriteRune('?')
		if i != len(tags)-1 {
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
