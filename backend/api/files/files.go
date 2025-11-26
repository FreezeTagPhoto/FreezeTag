package files

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileEndpoint struct {
	imageRepository repositories.ImageRepository
}

func InitFileEndpoint(repo repositories.ImageRepository) FileEndpoint {
	return FileEndpoint{
		repo,
	}
}

func (fe FileEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.GET("/file/:id", fe.HandleGet)
}

// @summary     Get file
// @description Get an image file given an ID
// @produce     application/octet-stream
// @router      /file/{id} [get]
// @param       id path int true "Image ID"
// @success     200 {file}   string "thumbnail file data"
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
func (fe FileEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}

	result, err := fe.imageRepository.GetImageFilepath(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	c.File(result)
}
