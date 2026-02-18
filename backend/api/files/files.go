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

// @summary     Download file
// @description Get an image file given an ID
// @produce     application/octet-stream
// @tags        files, images
// @router      /file/download/{id} [get]
// @param       id path int true "Image ID"
// @success     200 {file}   string "thumbnail file data"
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (fe FileEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}

	result, err := fe.imageRepository.GetImageFilepath(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.File(result)
}

// @summary     Delete file
// @description Delete an image file given an ID
// @produce     application/json
// @tags        files, images
// @router      /file/delete/:id [delete]
// @param       id path int true "Image ID"
// @success     200 {object} api.ImageDeleteResponse
// @failure     400 {object} api.BadRequestResponse
func (fe FileEndpoint) HandleDelete(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}

	file, err := fe.imageRepository.DeleteImage(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.ImageDeleteResponse{Id: id, File: file})
}
