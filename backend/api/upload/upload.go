package upload

import (
	"bytes"
	"freezetag/backend/api"
	"freezetag/backend/pkg/repositories"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* Types */
type UploadEndpoint struct {
	imageRepository repositories.ImageRepository
}

/* Functions */

// Creates a new UploadEndpoint with the given image repository.
func InitUploadEndpoint(repository repositories.ImageRepository) UploadEndpoint {
	return UploadEndpoint{
		imageRepository: repository,
	}
}

// Registers the upload endpoints to the given Gin engine.
func (ue UploadEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.POST("/upload", ue.HandlePost)
}

// @summary     Upload files
// @description Upload a set of image files to the server
// @produce     application/json
// @router      /upload [post]
// @param       file formData []file true "image file to upload" collectionFormat(multi)
// @success     200 {object} api.StatusOkUploadResponse
// @failure     400 {object} api.StatusBadRequestResponse
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

	results := make(chan repositories.UploadResult, len(files))
	for _, file := range files {
		bytes, err := readFileBytes(file)

		if err != nil {
			c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "error reading file bytes in file: " + file.Filename + " with error: " + err.Error()})
			return
		}
		go func(data []byte, filename string) {
			results <- ue.imageRepository.StoreImageBytes(data, filename)
		}(bytes, file.Filename)
	}
	// id := uuid.New()
	// c.JSON(http.StatusOK, api.Job{Uuid: id, Status: "processing", Name: "uploading files"})

	// uploaded := make([]repositories.ImageUploadSuccess, 0, len(files))
	// errors := make([]repositories.ImageUploadFail, 0)
	// for range files {
	// 	result := <-results
	// 	if result.Err != nil {
	// 		errors = append(errors, *result.Err)
	// 	} else {
	// 		uploaded = append(uploaded, *result.Success)
	// 	}
	// }

	// response := api.StatusOkUploadResponse{
	// 	Uploaded: uploaded,
	// 	Errors:   errors,
	// }
	// c.JSON(http.StatusOK, response)
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
