package database

import (
	"database/sql"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
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
	// Add an image file and its metadata to the database
	//
	// returns: the database ID of the image
	AddImage(string, imagedata.Data) (ImageId, error)
	// Add a thumbnail of the given size with the given data to the database
	//
	// returns: whether the thumbnail was added successfully
	AddImageThumbnail(ImageId, int, []byte) (bool, error)
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

func (db SqliteImageDatabase) AddImage(file string, data imagedata.Data) (ImageId, error) {
	var datetaken *string
	dateuploaded := time.Now().Format("2006-01-02 15:04:05")
	var make *string
	var model *string
	var latitude *float64
	var longitude *float64
	if data.DateCreated != nil {
		formatted := data.DateCreated.Format("2006-01-02 15:04:05")
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
	if rows == 0 {
		return false, nil
	}
	return true, nil
}
