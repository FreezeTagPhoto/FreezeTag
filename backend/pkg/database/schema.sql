PRAGMA integrity_check;
PRAGMA foreign_keys=true;

CREATE TABLE IF NOT EXISTS Images (
    id INTEGER PRIMARY KEY NOT NULL,
    imageFile TEXT NOT NULL,
    dateTaken TEXT,
    dateUploaded TEXT,
    cameraMake TEXT,
    cameraModel TEXT,
    latitude REAL,
    longitude REAL
);

CREATE TABLE IF NOT EXISTS Tags (
    imageId INTEGER,
    tag TEXT,
    FOREIGN KEY(imageId) REFERENCES Images(imageId) ON DELETE CASCADE,
    UNIQUE(imageId, tag)
);

CREATE TABLE IF NOT EXISTS Thumbnails (
    imageId INTEGER,
    thumbnailSize INTEGER,
    thumbnailData BLOB,
    FOREIGN KEY(imageId) REFERENCES Images(imageId) ON DELETE CASCADE,
    CHECK(thumbnailSize > 0),
    UNIQUE(imageId, thumbnailSize)
);