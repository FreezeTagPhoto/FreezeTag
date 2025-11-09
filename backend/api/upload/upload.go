package upload

import (
	"bytes"
	"freezetag/backend/pkg/repositories"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* Types */
type StatusOkResponse struct {
	Uploaded []repositories.ImageHandleSuccess `json:"uploaded"`
	Errors   []repositories.ImageHandleFail    `json:"errors"`
}

type StatusBadRequestResponse struct {
	Error string `json:"error"`
}

type UploadEndpoint struct {
	imageRepository repositories.ImageRepository
}

/* Functions */

// Creates a new UploadEndpoint with the given image repository.
func InitUploadEndpoint(repository repositories.ImageRepository) UploadEndpoint {
	return UploadEndpoint{
		repository,
	}
}

// Registers the upload endpoints to the given Gin engine.
func (ue UploadEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.POST("/upload", ue.handlePost)
}

// Handles the POST /upload endpoint, sending uploaded files to the image repository that was
// initialized in the InitUploadEndpoint function.
func (ue UploadEndpoint) handlePost(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, StatusBadRequestResponse{Error: "failed to parse multipart form: " + err.Error()})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, StatusBadRequestResponse{Error: "No files uploaded"})
		return
	}

	results := make(chan repositories.Result, len(files))
	for _, file := range files {
		bytes, err := readFileBytes(file)

		if err != nil {
			c.JSON(http.StatusBadRequest, StatusBadRequestResponse{Error: "error reading file bytes in file: " + file.Filename + " with error: " + err.Error()})
			return
		}
		go func(data []byte, filename string) {
			results <- ue.imageRepository.StoreImageBytes(data, filename)
		}(bytes, file.Filename)
	}

	uploaded := make([]repositories.ImageHandleSuccess, 0, len(files))
	errors := make([]repositories.ImageHandleFail, 0)
	for range files {
		result := <-results
		if result.Err != nil {
			errors = append(errors, *result.Err)
		}
		uploaded = append(uploaded, *result.Success)
	}

	response := StatusOkResponse{
		Uploaded: uploaded,
		Errors:   errors,
	}
	c.JSON(http.StatusOK, response)
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


