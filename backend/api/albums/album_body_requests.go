package albums

import (
	"freezetag/backend/pkg/database"
)

type AlbumCreateRequest struct {
	Name           string                 `json:"name" binding:"required"`
	VisibilityMode database.GlobalPrivacy `json:"visibility_mode" binding:"oneof=0 1 2"`
}

type AlbumModifyRequest struct {
	AlbumId        database.AlbumID        `json:"album_id" binding:"required"`
	Name           *string                 `json:"name,omitempty"`
	VisibilityMode *database.GlobalPrivacy `json:"visibility_mode,omitempty" binding:"omitempty,oneof=0 1"`
}

type UserAlbumPermissionRequest struct {
	AlbumId      database.AlbumID       `json:"album_id" binding:"required"`
	TargetUserId database.UserID        `json:"target_user_id" binding:"required"`
	Permission   database.GlobalPrivacy `json:"permission" binding:"oneof=0 1 2"`
}

type AlbumImageRequest struct {
	ImageId database.ImageId `json:"image_id" binding:"required"`
}

type RenameAlbumRequest struct {
	NewName string `json:"new_name" binding:"required"`
}

type DeleteAlbumRequest struct {
	AlbumName string `json:"album_name" binding:"required"`
}
