package database

import _ "embed"

import "database/sql"

type Manager struct {
	db      *sql.DB
	ImageDB ImageDatabase
	UserDB  UserDatabase
	AlbumDB AlbumDatabase
}

//go:embed schema.sql
var schema string

func NewDefaultManager(dbPath string) (*Manager, error) {
	registerExtendedSqlite("sqlite3_extrafunc")
	db, err := sql.Open("sqlite3_extrafunc", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	ImageDb := SqliteImageDatabase{db}
	UserDb := SqliteUserDatabase{db}
	AlbumDb := SqliteAlbumDatabase{db}
	if err := UserDb.seedPermissions(); err != nil {
		return nil, err
	}
	return &Manager{db: db, ImageDB: ImageDb, UserDB: UserDb, AlbumDB: AlbumDb}, nil
}
