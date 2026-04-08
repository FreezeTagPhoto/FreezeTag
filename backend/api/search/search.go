package search

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchEndpoint struct {
	imageRepository repositories.ImageRepository
}

func InitSearchEndpoint(repository repositories.ImageRepository) SearchEndpoint {
	return SearchEndpoint{
		repository,
	}
}

// @summary     Search images
// @description Search for an image given information
// @produce     application/json
// @router      /search [get]
// @tags        search, images
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
// @success     200 {array}  database.ImageID
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
func (se SearchEndpoint) Search(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	query := api.GetRequestQuery(c)
	if query == nil {
		return
	}
	var pageSize uint
	if psParam := c.Query("pageSize"); psParam != "" {
		ps, err := strconv.ParseUint(psParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad pageSize parameter"})
			return
		}
		pageSize = uint(ps)
	}
	var pageNo uint
	if pcParam := c.Query("pageNo"); pcParam != "" {
		pn, err := strconv.ParseUint(pcParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad pageNo parameter"})
			return
		}
		pageNo = uint(pn)
	}
	var field queries.SortField
	var order queries.SortOrder
	if sfParam := c.Query("sortBy"); sfParam != "" {
		switch sfParam {
		case "DateAdded":
			field = queries.DateAdded
		case "DateCreated":
			field = queries.DateCreated
		default:
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad sortBy parameter"})
			return
		}
	}
	if soParam := c.Query("sortOrder"); soParam != "" {
		switch soParam {
		case "ASC":
			order = queries.Ascending
		case "DESC":
			order = queries.Descending
		default:
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad sortOrder parameter"})
			return
		}
	}
	images, err := se.imageRepository.SearchImageOrderedPaged(query, field, order, pageSize, pageNo, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}
