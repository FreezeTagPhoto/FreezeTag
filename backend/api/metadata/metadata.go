package metadata

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MetadataEndpoint struct {
	imageRepository repositories.ImageRepository
}

func InitMetadataEndpoint(repo repositories.ImageRepository) MetadataEndpoint {
	return MetadataEndpoint{
		repo,
	}
}

// @summary     Get metadata
// @description Retrieve metadata for an image
// @produce     application/json
// @tags        metadata, images, search
// @router      /metadata/{id} [get]
// @param       id path int true "Image ID"
// @success     200 {object} api.MetadataResponse
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
func (me MetadataEndpoint) Metadata(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}

	meta, err := me.imageRepository.GetImageMetadata(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	width, height, err := me.imageRepository.GetImageResolution(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	result := api.MetadataResponse{
		Metadata: meta,
		Width:    width,
		Height:   height,
	}
	c.JSON(http.StatusOK, result)
}
