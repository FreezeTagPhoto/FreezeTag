package albums

// I hate getting stuff out of query params,
// so I am trying out a new approach where all the album related endpoints take a json body with the relevant info.
// This is the file that defines the structs for those json bodies.
// pretty easy to change back to the way we do it now if this doesn't work out the way I want it to

import (
	"freezetag/backend/pkg/database"
)

type AlbumCreateRequest struct {
	Name           string                `json:"name" binding:"required"`
	VisibilityMode database.PrivacyLevel `json:"visibility_mode" binding:"oneof=0 1 2"`
}

type AlbumModifyRequest struct {
	AlbumId        database.AlbumID       `json:"album_id" binding:"required"`
	Name           *string                `json:"name,omitempty"`
	VisibilityMode *database.PrivacyLevel `json:"visibility_mode,omitempty" binding:"omitempty,oneof=0 1"`
}

type UserAlbumPermissionRequest struct {
	AlbumId      database.AlbumID      `json:"album_id" binding:"required"`
	TargetUserId database.UserID       `json:"target_user_id" binding:"required"`
	Permission   database.PrivacyLevel `json:"permission" binding:"oneof=0 1 2"`
}

type AlbumImageRequest struct {
	ImageId database.ImageId `json:"image_id"`
	AlbumId database.AlbumID `json:"album_id"`
}

type RenameAlbumRequest struct {
	AlbumID database.AlbumID `json:"album_id" binding:"required"`
	NewName string           `json:"new_name" binding:"required"`
}

type DeleteAlbumRequest struct {
	AlbumName string `json:"album_name" binding:"required"`
}
