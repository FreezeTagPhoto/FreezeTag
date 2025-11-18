package api

import (
	"freezetag/backend/pkg/repositories"

	"github.com/gin-gonic/gin"
)


type StatusOkResponse struct {
	Uploaded []repositories.ImageUploadSuccess `json:"uploaded"`
	Errors   []repositories.ImageUploadFail    `json:"errors"`
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
