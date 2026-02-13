package upload

import (
	"bytes"
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* Types */
type UploadEndpoint struct {
	jobService services.JobService
}

/* Functions */

// Creates a new UploadEndpoint with the given image repository.
func InitUploadEndpoint(jobService services.JobService) UploadEndpoint {
	return UploadEndpoint{
		jobService: jobService,
	}
}

// Registers the upload endpoints to the given Gin engine.
func (ue UploadEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/upload", ue.HandlePost)
}

// @summary     Upload files
// @description Upload a set of image files to the server
// @produce     application/json
// @router      /upload [post]
// @tags        upload
// @param       file formData []file true "image file to upload" collectionFormat(multi)
// @success     202 {object} string "the UUID of the created job batch for the upload"
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
func (ue UploadEndpoint) HandlePost(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to parse multipart form: " + err.Error()})
		return
	}

	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "multipart form has no file field or no files were uploaded"})
		return
	}

	jobs := []services.FileJob{}
	for _, file := range files {
		bytes, err := readFileBytes(file)

		if err != nil {
			c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "error reading file bytes in file: " + file.Filename + " with error: " + err.Error()})
			return
		}
		jobs = append(jobs, services.FileJob{Name: file.Filename, Bytes: bytes})
	}
	id := ue.jobService.RunUploadJob(jobs)
	c.JSON(http.StatusAccepted, id)
}

// Reads the bytes from a multipart.FileHeader
func readFileBytes(fh *multipart.FileHeader) ([]byte, error) {
	var buf bytes.Buffer

	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	if _, err := io.Copy(&buf, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
