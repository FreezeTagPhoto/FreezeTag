PRAGMA integrity_check;
PRAGMA foreign_keys=true;

CREATE TABLE IF NOT EXISTS Images (
    id INTEGER PRIMARY KEY NOT NULL,
    imageFile TEXT NOT NULL,
    dateTaken INTEGER, -- unix epoch
    dateUploaded INTEGER, -- unix epoch
    cameraMake TEXT,
    cameraModel TEXT,
    latitude REAL,
    longitude REAL,
    width INTEGER,
    height INTEGER
);

CREATE TABLE IF NOT EXISTS Tags (
    id INTEGER PRIMARY KEY NOT NULL,
    tag TEXT,
    UNIQUE(tag)
);

CREATE TABLE IF NOT EXISTS ImageTags (
    imageId INTEGER,
    tagId INTEGER,
    FOREIGN KEY(imageId) REFERENCES Images(id) ON DELETE CASCADE,
    FOREIGN KEY(tagId) REFERENCES Tags(id) ON DELETE CASCADE,
    UNIQUE(imageId, tagId)
);

CREATE TABLE IF NOT EXISTS Thumbnails (
    imageId INTEGER,
    thumbnailSize INTEGER,
    thumbnailData BLOB,
    FOREIGN KEY(imageId) REFERENCES Images(id) ON DELETE CASCADE,
    CHECK(thumbnailSize > 0),
    UNIQUE(imageId, thumbnailSize)
);

CREATE TABLE IF NOT EXISTS Albums (
    id INTEGER PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS AlbumImages (
    albumId INTEGER,
    imageId INTEGER,
    FOREIGN KEY(albumId) REFERENCES Albums(id) ON DELETE CASCADE,
    FOREIGN KEY(imageId) REFERENCES Images(id) ON DELETE CASCADE,
    UNIQUE(albumId, imageId)
);