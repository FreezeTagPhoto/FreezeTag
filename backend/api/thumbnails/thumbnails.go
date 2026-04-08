package thumbnails

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ThumbnailEndpoint struct {
	imageRepository repositories.ImageRepository
}

func InitThumbnailEndpoint(repository repositories.ImageRepository) ThumbnailEndpoint {
	return ThumbnailEndpoint{
		repository,
	}
}

// @summary     Get thumbnail
// @description Get a WEBP format image thumbnail for an image
// @produce     image/webp
// @tags        thumbnails, images
// @router      /thumbnails/{id} [get]
// @param       id path int true "Image ID"
// @success     200 {file}   string "thumbnail file data"
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (te ThumbnailEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	id, err := api.ParseParamIntoID[database.ImageID](idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid image ID parameter"})
		return
	}
	sizeParam := c.Query("size")
	var size uint
	if num, err := strconv.ParseUint(sizeParam, 10, 32); err != nil {
		size = 1
	} else {
		size = uint(num)
	}
	data, err := te.imageRepository.RetrieveThumbnail(id, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	if data == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: "thumbnail not found"})
		return
	}
	c.Data(http.StatusOK, "image/webp", data)
}
