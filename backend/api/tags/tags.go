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

// @summary     Delete tags
// @description Delete tags from the database
// @tags        tags
// @produce     application/json
// @router      /tag/delete [delete]
// @param       tag query []string true "tags to delete" collectionFormat(multi)
// @success     200 {object} api.TagDeleteResponse
// @failure     400 {object} api.BadRequestResponse
func (te TagEndpoint) HandleDeleteFull(c *gin.Context) {
	tags := c.QueryArray("tag")
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no tags to delete"})
		return
	}
	count, err := te.imageRepository.DeleteTags(tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.TagDeleteResponse{Deleted: count})
}

// @summary     Delete image tags
// @description Delete tags from images
// @tags        tags
// @produce     application/json
// @router      /tag/remove [delete]
// @param       tag query []string true "tags to remove"           collectionFormat(multi)
// @param       id  query []int    true "image IDs to remove from" collectionFormat(multi)
// @success     200 {object} api.TagRemoveResponse
// @failure     400 {object} api.BadRequestResponse
func (te TagEndpoint) HandleDelete(c *gin.Context) {
	tags := c.QueryArray("tag")
	ids := c.QueryArray("id")
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no tags to remove"})
		return
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no ids to remove tags from"})
		return
	}

	deleted := make([]repositories.ImageTagSuccess, 0)
	failed := make([]repositories.ImageTagFail, 0)

	for _, idStr := range c.QueryArray("id") {
		id, err := api.ParseParamIntoID[uint64](idStr)
		if err != nil {
			failed = append(failed, repositories.ImageTagFail{Reason: fmt.Sprintf("unknown id %s", idStr)})
			continue
		}
		res := te.imageRepository.RemoveImageTags(database.ImageId(id), tags)
		if res.Success != nil {
			deleted = append(deleted, repositories.ImageTagSuccess{
				Id:    res.Success.Id,
				Count: res.Success.Count,
			})
		}
		if res.Err != nil {
			failed = append(failed, repositories.ImageTagFail{
				Id:     res.Err.Id,
				Reason: res.Err.Reason,
			})
		}
	}

	response := api.TagRemoveResponse{
		Deleted: deleted,
		Errors:  failed,
	}
	c.JSON(http.StatusOK, response)
}

// @summary     Add tags
// @description Add tags to images (or to the database if no images are specified)
// @tags        tags, upload
// @produce     application/json
// @router      /tag/add [post]
// @param       tag query []string true "tags to add"         collectionFormat(multi)
// @param       id  query []int    false "image IDs to add to (optional)" collectionFormat(multi)
// @success     200 {object} api.TagAddResponse
// @failure     400 {object} api.BadRequestResponse
func (te TagEndpoint) HandlePost(c *gin.Context) {
	tags := c.QueryArray("tag")
	ids := c.QueryArray("id")
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no tags to add"})
		return
	}
	if len(ids) == 0 {
		te.addTagsOnly(tags, c)
		return
	}

	added := make([]repositories.ImageTagSuccess, 0)
	failed := make([]repositories.ImageTagFail, 0)
	for _, idStr := range ids {
		id, err := api.ParseParamIntoID[uint64](idStr)
		if err != nil {
			failed = append(failed, repositories.ImageTagFail{
				Reason: fmt.Sprintf("unknown id %s", idStr),
			})
			continue
		}
		// TODO: I think to truly fix the bug, this needs to check for image id validation in general.
		// TODO: Or we need to fix the foreign key locking everything
		if id == 0 {
			failed = append(failed, repositories.ImageTagFail{
				Reason: fmt.Sprintf("bad image id: %d", id),
			})
			continue
		}
		res := te.imageRepository.AddImageTags(database.ImageId(id), tags)
		if res.Success != nil {
			added = append(added, repositories.ImageTagSuccess{
				Id:    res.Success.Id,
				Count: res.Success.Count,
			})
		}
		if res.Err != nil {
			failed = append(failed, repositories.ImageTagFail{
				Id:     res.Err.Id,
				Reason: res.Err.Reason,
			})
		}
	}

	response := api.TagAddResponse{
		Added:  added,
		Errors: failed,
	}
	c.JSON(http.StatusOK, response)
}

func (te TagEndpoint) addTagsOnly(tags []string, c *gin.Context) {
	res := te.imageRepository.AddTags(tags)
	if res.Success {
		c.JSON(http.StatusOK, api.TagAddResponse{
			Added:  []repositories.ImageTagSuccess{{Count: len(tags)}},
			Errors: []repositories.ImageTagFail{},
		})
	} else {
		c.JSON(http.StatusOK, api.TagAddResponse{
			Added:  []repositories.ImageTagSuccess{},
			Errors: []repositories.ImageTagFail{{Reason: res.Err}},
		})
	}
}

// @summary     List all tags
// @description Get all the tags in the database
// @produce     application/json
// @tags        tags
// @router      /tag/list [get]
// @success     200 {object} api.TagCounts
// @failure     500 {object} api.ServerErrorResponse
func (te TagEndpoint) ListTags(c *gin.Context) {
	result, err := te.imageRepository.RetrieveAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.TagCounts(result))
}

// @summary     List image tags
// @description Get all the tags associated with an image
// @produce     application/json
// @tags        tags
// @router      /tag/list/{id} [get]
// @param       id path int true "image ID to get the tags of"
// @success     200 {array} string
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (te TagEndpoint) ImageTags(c *gin.Context) {
	idParam := c.Param("id")
	var id database.ImageId
	if num, err := strconv.ParseInt(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid image ID parameter"})
		return
	} else {
		id = database.ImageId(num)
	}
	result, err := te.imageRepository.RetrieveImageTags(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @summary     Get tag counts
// @description Get all tags and the total count of the overlap for each tag associated with the provided image IDs
// @tags        tags
// @router      /tag/counts [get]
// @param       id query []string true "image IDs to get tag counts for" collectionFormat(multi)
// @success     200 {object} api.TagCounts
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (te TagEndpoint) ListCounts(c *gin.Context) {
	idParam := c.QueryArray("id")
	if len(idParam) == 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "no ids specified"})
		return
	}
	ids := make([]database.ImageId, len(idParam))
	for i, idStr := range idParam {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad id parameter"})
			return
		}
		ids[i] = database.ImageId(id)
	}

	result, err := te.imageRepository.GetTagCounts(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.TagCounts(result))
}

// @summary     Get tag counts from search
// @description Get all tags and the total overlap count for a search query
// @tags        search, tags
// @router      /tag/search [get]
// @param       make           query string   false "camera make"
// @param       makeLike       query string   false "camera make fuzzy"
// @param       model          query string   false "camera model"
// @param       modelLike      query string   false "camera model fuzzy"
// @param       takenBefore    query string   false "picture taken before (unix epoch)"
// @param       takenAfter     query string   false "picture taken after (unix epoch)"
// @param       uploadedBefore query string   false "picture uploaded before (unix epoch)"
// @param       uploadedAfter  query string   false "picture uploaded after (unix epoch)"
// @param       near           query string   false "latitude/longitude/distance (degrees)" example(100.0,12.0,1.0)
// @param       tag            query []string false "picture tag"                           collectionFormat(multi)
// @param       tagLike        query []string false "picture tag fuzzy"                     collectionFormat(multi)
// @param       sortBy         query string   false "sort by"                               Enums(DateAdded,DateCreated) default(DateAdded)
// @param       sortOrder      query string   false "sort order"                            Enums(ASC,DESC) default(DESC)
// @param       pageSize       query uint     false "page size"
// @param       pageNo         query uint     false "page number (zero indexed)"
// @success     200 {object} api.TagCounts
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (te TagEndpoint) ListCountsQuery(c *gin.Context) {
	userId, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userId not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userId in context is not of type UserID"})
		return
	}
	query := api.GetRequestQuery(c)
	if query == nil {
		return
	}
	tc, err := te.imageRepository.GetQueryTagCounts(query, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.TagCounts(tc))
}
