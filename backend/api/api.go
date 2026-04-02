package api

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"

	"github.com/google/uuid"
)

type UploadResponse struct {
	Uploaded []repositories.ImageUploadSuccess `json:"uploaded"`
	Errors   []repositories.ImageUploadFailure `json:"errors"`
}

type TagRemoveResponse struct {
	Deleted []repositories.ImageTagSuccess `json:"deleted"`
	Errors  []repositories.ImageTagFail    `json:"errors"`
}

type TagDeleteResponse struct {
	Deleted int `json:"deleted"`
}

type ImageDeleteResponse struct {
	Id   database.ImageId `json:"id"`
	File string           `json:"file"`
}

type TagAddResponse struct {
	Added  []repositories.ImageTagSuccess `json:"added"`
	Errors []repositories.ImageTagFail    `json:"errors"`
}

type LoginUserResponse struct {
	UserID      database.UserID  `json:"user_id"`
	Permissions data.Permissions `json:"permissions"`
}

type BadRequestResponse struct {
	Error string `json:"error"`
}

type ServerErrorResponse struct {
	Error string `json:"error"`
}

type NotFoundResponse struct {
	Error string `json:"error"`
}

type LoginFailResponse struct {
	Error string `json:"error"`
}

type LoginSuccessResponse struct {
	Token string `json:"token"`
}

type LogoutSuccessResponse struct {
	Status string `json:"status"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type MetadataResponse struct {
	imagedata.Metadata
	Width  int `json:"width"`
	Height int `json:"height"`
}

type CancelledJobResponse struct {
	UUID uuid.UUID `json:"uuid"`
}

type PluginDisabledResponse struct {
	Disabled bool `json:"disabled"`
}

type LoginCredentials struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type AlbumCreateResponse struct {
	AlbumID database.AlbumID `json:"album_id"`
}

type AlbumNameResponse struct {
	Name string `json:"name"`
}

type AlbumOwnerResponse struct {
	OwnerID database.UserID `json:"owner_id"`
}

type AlbumSharedUsersResponse struct {
	SharedUserIDs []database.UserID `json:"shared_user_ids"`
}

type TagCounts map[string]int64

type PasswordChangeRequest struct {
	CurrentPassword string `form:"current_password" json:"current_password" binding:"required"`
	NewPassword     string `form:"new_password" json:"new_password" binding:"required"`
}

type innerFileJob struct {
	services.FileJob
	id int `json:"-"`
}

func (j innerFileJob) ID() int {
	return j.id
}

type FileJobBatch repositories.JobBatch[innerFileJob, repositories.ImageUploadSuccess]
