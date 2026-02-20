PRAGMA integrity_check;
PRAGMA foreign_keys=true;

CREATE TABLE IF NOT EXISTS Users (
    id INTEGER PRIMARY KEY NOT NULL,
    username TEXT NOT NULL UNIQUE,
    passwordHash TEXT NOT NULL,
    createdAt INTEGER NOT NULL
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
-- admin users might not necessarily want to give all permissions to a token
CREATE TABLE IF NOT EXISTS Token_Permissions (
    tokenId INTEGER NOT NULL,
    permissionId INTEGER NOT NULL,
    FOREIGN KEY (tokenId) REFERENCES API_Token(id) ON DELETE CASCADE,
    FOREIGN KEY (permissionId) REFERENCES App_Permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (tokenId, permissionId)
);