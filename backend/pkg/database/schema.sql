PRAGMA integrity_check;
PRAGMA foreign_keys=true;

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

CREATE TABLE IF NOT EXISTS AlbumImages (
    albumId INTEGER,
    imageId INTEGER,
    FOREIGN KEY(albumId) REFERENCES Albums(id) ON DELETE CASCADE,
    FOREIGN KEY(imageId) REFERENCES Images(id) ON DELETE CASCADE,
    UNIQUE(albumId, imageId)
);


CREATE TABLE IF NOT EXISTS Albums (
    id INTEGER PRIMARY KEY NOT NULL,
    userId INTEGER NOT NULL,
    -- 0: private album, visibility to no users by default, but can give access to specific users
    -- 1: public album,  visible to all users by default, but can restrict access to specific users
    visibility_mode INTEGER NOT NULL DEFAULT 0,
    -- 0: users who can see this album can only view images in it
    -- 1: users who can see this album can view and modify images in it
    edit_mode INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(userId) REFERENCES Users(id) ON DELETE CASCADE,
    CHECK(visibility_mode IN (0, 1)),
    CHECK(edit_mode IN (0, 1)),
    CHECK((visibility_mode = 0 AND edit_mode = 0) OR (visibility_mode = 1)),
    UNIQUE(userId, name),
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS User_Per_Album_Settings (
    albumId INTEGER NOT NULL,
    userId INTEGER NOT NULL,
    -- 0: Block user from seeing any images associated with this album, 
    -- 1: Allow user to see images in this album 
    visibility_mode INTEGER NOT NULL DEFAULT 0,
    -- 0: User can only view images in this album,
    -- 1: User can view and modify images in this album
    edit_mode INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(albumId) REFERENCES Albums(id) ON DELETE CASCADE,
    FOREIGN KEY(userId) REFERENCES Users(id) ON DELETE CASCADE,
    CHECK(visibility_mode IN (0, 1)),
    CHECK(edit_mode IN (0, 1)),
    CHECK((visibility_mode = 0 AND edit_mode = 0) OR (visibility_mode = 1)),
    PRIMARY KEY (albumId, userId)
);

CREATE TABLE IF NOT EXISTS Users (
    id INTEGER PRIMARY KEY NOT NULL,
    username TEXT NOT NULL UNIQUE,
    passwordHash TEXT NOT NULL,
    createdAt INTEGER NOT NULL,
    -- 0: Only see images in albums they have access to
    -- 1: See most images, but not images in whitelisted albums they dont have access to
    visibility_mode INTEGER NOT NULL DEFAULT 1,
    profilePicture BLOB
    CHECK(visibility_mode IN (0, 1))
);

CREATE TABLE IF NOT EXISTS API_Token (
    id INTEGER PRIMARY KEY NOT NULL,
    userId INTEGER NOT NULL,
    tokenHash TEXT NOT NULL UNIQUE,
    createdAt INTEGER NOT NULL,
    expiresAt INTEGER,
    label TEXT NOT NULL,
    revoked INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (userId) REFERENCES Users(id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS App_Permissions (
    id INTEGER PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL UNIQUE,          
    name TEXT NOT NULL UNIQUE,                
    description TEXT                  
);

CREATE TABLE IF NOT EXISTS User_Permissions (
    userId INTEGER NOT NULL,
    permissionId INTEGER NOT NULL,
    FOREIGN KEY (userId) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (permissionId) REFERENCES App_Permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (userId, permissionId)
);

CREATE TABLE IF NOT EXISTS Token_Permissions (
    tokenId INTEGER NOT NULL,
    permissionId INTEGER NOT NULL,
    FOREIGN KEY (tokenId) REFERENCES API_Token(id) ON DELETE CASCADE,
    FOREIGN KEY (permissionId) REFERENCES App_Permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (tokenId, permissionId)
);

