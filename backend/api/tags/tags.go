package tags

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TagEndpoint struct {
	imageRepository repositories.ImageRepository
}

func InitTagEndpoint(repository repositories.ImageRepository) TagEndpoint {
	return TagEndpoint{
		repository,
	}
}

func (te TagEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.DELETE("/tag/remove", te.HandleDelete)
	e.POST("/tag/add", te.HandlePost)
	e.GET("/tag/list", te.HandleGetAllTags)
	e.GET("/tag/list/:id", te.HandleGetImageTags)
}

func (te TagEndpoint) HandleDelete(c *gin.Context) {
	if len(c.QueryArray("tag")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no tags to remove"})
		return
	}
	if len(c.QueryArray("id")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no ids to remove tags from"})
		return
	}

	tags := c.QueryArray("tag")
	idlen := len(c.QueryArray("id"))
	results := make(chan repositories.ImageTagResult, idlen)
	for _, id := range c.QueryArray("id") {
		idstr, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			results <- repositories.ImageTagResult{
				Success: nil,
				Err: &repositories.ImageTagFail{
					Id:     -1,
					Reason: fmt.Sprintf("unknown id %s", id),
				},
			}
		}
		go func(id database.ImageId, tags []string) {
			results <- te.imageRepository.RemoveImageTags(id, tags)
		}(database.ImageId(idstr), tags)
	}

	deleted := make([]repositories.ImageTagSuccess, 0)
	errors := make([]repositories.ImageTagFail, 0)
	for range idlen {
		result := <-results
		if result.Err != nil {
			errors = append(errors, *result.Err)
		} else {
			deleted = append(deleted, *result.Success)
		}
	}

	response := api.StatusOkTagDeleteResponse{
		Deleted: deleted,
		Errors:  errors,
	}
	c.JSON(http.StatusOK, response)
}

func (te TagEndpoint) HandlePost(c *gin.Context) {
	if len(c.QueryArray("tag")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no tags to remove"})
		return
	}
	if len(c.QueryArray("id")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no ids to remove tags from"})
		return
	}

	tags := c.QueryArray("tag")
	idlen := len(c.QueryArray("id"))
	results := make(chan repositories.ImageTagResult, idlen)
	for _, id := range c.QueryArray("id") {
		idstr, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			results <- repositories.ImageTagResult{
				Success: nil,
				Err: &repositories.ImageTagFail{
					Id:     -1,
					Reason: fmt.Sprintf("unknown id %s", id),
				},
			}
		}
		go func(id database.ImageId, tags []string) {
			results <- te.imageRepository.AddImageTags(id, tags)
		}(database.ImageId(idstr), tags)
	}

	deleted := make([]repositories.ImageTagSuccess, 0)
	errors := make([]repositories.ImageTagFail, 0)
	for range idlen {
		result := <-results
		if result.Err != nil {
			errors = append(errors, *result.Err)
		} else {
			deleted = append(deleted, *result.Success)
		}
	}

	response := api.StatusOkTagDeleteResponse{
		Deleted: deleted,
		Errors:  errors,
	}
	c.JSON(http.StatusOK, response)
}

func (te TagEndpoint) HandleGetAllTags(c *gin.Context) {
	result, err := te.imageRepository.RetrieveAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (te TagEndpoint) HandleGetImageTags(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}
	result, err := te.imageRepository.RetrieveImageTags(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
