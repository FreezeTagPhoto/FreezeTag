package api

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/repositories"

	"github.com/gin-gonic/gin"
)

type StatusOkUploadResponse struct {
	Uploaded []repositories.ImageUploadSuccess `json:"uploaded"`
	Errors   []repositories.ImageUploadFailure `json:"errors"`
}

type StatusOkTagDeleteResponse struct {
	Deleted []repositories.ImageTagSuccess `json:"deleted"`
	Errors  []repositories.ImageTagFail    `json:"errors"`
}

type StatusOkTagAddResponse struct {
	Added  []repositories.ImageTagSuccess `json:"added"`
	Errors []repositories.ImageTagFail    `json:"errors"`
}

type StatusBadRequestResponse struct {
	Error string `json:"error"`
}

type StatusServerErrorResponse struct {
	Error string `json:"error"`
}

type StatusNotFoundResponse struct {
	Error string `json:"error"`
}

type ApiEndpoint interface {
	RegisterEndpoints(e *gin.Engine)
}

type StatusLoginFail struct {
	Error string `json:"error"`
}

type StatusLoginSuccess struct {
	Token string `json:"token"`
}

type StatusLogoutSuccess struct {
	Status string `json:"status"`
}

type LoginCredentials struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type StatusLoginUser struct {
	UserID database.UserID `json:"user_id"`
}

type TagCounts map[string]int64

type MetadataResponse struct {
	imagedata.Metadata
	Width  int `json:"width"`
	Height int `json:"height"`
}

type JobBatch repositories.JobBatch
