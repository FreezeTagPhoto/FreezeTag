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

func (te ThumbnailEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.GET("/thumbnails/:id", te.HandleGet)
}

func (te ThumbnailEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
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
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	if data == nil {
		c.JSON(http.StatusNotFound, api.StatusNotFoundResponse{Error: "thumbnail not found"})
		return
	}
	c.Data(http.StatusOK, "image/webp", data)
}
