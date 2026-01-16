package api

import (
	"freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/repositories"

	"github.com/gin-gonic/gin"
)

type StatusOkUploadResponse struct {
	Uploaded []repositories.ImageUploadSuccess `json:"uploaded"`
	Errors   []repositories.ImageUploadFail    `json:"errors"`
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

type MetadataResponse imagedata.Metadata

type JobBatch repositories.JobBatch
