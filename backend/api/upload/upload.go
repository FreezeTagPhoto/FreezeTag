package upload

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadEndpoint struct {
	jobService services.JobService
}

// Creates a new UploadEndpoint with the given image repository.
func InitUploadEndpoint(jobService services.JobService) UploadEndpoint {
	return UploadEndpoint{
		jobService: jobService,
	}
}

// Registers the upload endpoints to the given Gin engine.

// @summary     Upload files
// @description Upload a set of image files to the server
// @produce     application/json
// @router      /upload [post]
// @tags        upload
// @param       file formData []file true "image file to upload" collectionFormat(multi)
// @success     202 {object} string "the UUID of the created job batch for the upload"
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (ue UploadEndpoint) Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse multipart form: " + err.Error()})
		return
	}

	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "multipart form has no file field or no files were uploaded"})
		return
	}

	jobs := []services.FileJob{}
	for _, file := range files {
		bytes, err := api.ReadFileBytes(file)

		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "error reading file bytes in file: " + file.Filename + " with error: " + err.Error()})
			return
		}
		jobs = append(jobs, services.FileJob{Name: file.Filename, Bytes: bytes})
	}
	id := ue.jobService.RunUploadJob(jobs)
	c.JSON(http.StatusAccepted, id)
}
