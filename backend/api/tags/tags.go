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

func InitTagEndpoint(repo repositories.ImageRepository) TagEndpoint {
	return TagEndpoint{
		repo,
	}
}

func (te TagEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.DELETE("/tag/remove", te.HandleDelete)
	e.POST("/tag/add", te.HandlePost)
	e.GET("/tag/list", te.HandleGetAllTags)
	e.GET("/tag/list/:id", te.HandleGetImageTags)
	e.GET("/tag/counts", te.HandleGetTagCounts)
}

// @summary     Delete tags
// @description Delete tags from images
// @tags        tags
// @produce     application/json
// @router      /tag/remove [delete]
// @param       tag query []string true "tags to remove"           collectionFormat(multi)
// @param       id  query []int    true "image IDs to remove from" collectionFormat(multi)
// @success     200 {object} api.StatusOkTagDeleteResponse
// @failure     400 {object} api.StatusBadRequestResponse
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
		} else {
			go func(id database.ImageId, tags []string) {
				results <- te.imageRepository.RemoveImageTags(id, tags)
			}(database.ImageId(idstr), tags)
		}
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

// @summary     Add tags
// @description Add tags to images
// @tags        tags, upload
// @produce     application/json
// @router      /tag/add [post]
// @param       tag query []string true "tags to add"         collectionFormat(multi)
// @param       id  query []int    true "image IDs to add to" collectionFormat(multi)
// @success     200 {object} api.StatusOkTagAddResponse
// @failure     400 {object} api.StatusBadRequestResponse
func (te TagEndpoint) HandlePost(c *gin.Context) {
	if len(c.QueryArray("tag")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no tags to add"})
		return
	}
	if len(c.QueryArray("id")) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no ids to add tags to"})
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
		} else {
			go func(id database.ImageId, tags []string) {
				results <- te.imageRepository.AddImageTags(id, tags)
			}(database.ImageId(idstr), tags)
		}

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

	response := api.StatusOkTagAddResponse{
		Added:  deleted,
		Errors: errors,
	}
	c.JSON(http.StatusOK, response)
}

// @summary     List all tags
// @description Get all the tags in the database
// @produce     application/json
// @tags        tags, search
// @router      /tag/list [get]
// @success     200 {array}  string
// @failure     500 {object} api.StatusServerErrorResponse
func (te TagEndpoint) HandleGetAllTags(c *gin.Context) {
	result, err := te.imageRepository.RetrieveAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @summary     List image tags
// @description Get all the tags associated with an image
// @produce     application/json
// @tags        tags
// @router      /tag/list/{id} [get]
// @param       id path int true "image ID to get the tags of"
// @success     200 {array}  string
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
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

// @summary     Get tag counts
// @description Get all tags and the total count of the overlap for each tag associated with the provided image IDs
// @tags        tags, search
// @router      /tag/counts [get]
// @param       ids query []string true "image IDs to get tag counts for" collectionFormat(multi)
// @success     200 {object} api.TagCounts
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
// @produce     application/json
func (te TagEndpoint) HandleGetTagCounts(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "no ids specified"})
		return
	}

	result, err := te.imageRepository.GetTagCounts(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.TagCounts(result))
}
