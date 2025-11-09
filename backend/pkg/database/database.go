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
	GetImages(queries.DatabaseQuery) ([]ImageId, error)
	AddImage(string, imagedata.Data) (ImageId, error)
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
