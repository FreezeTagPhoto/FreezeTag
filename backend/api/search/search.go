package search

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func parseNearParam(near string) ([3]float64, error) {
	parts := strings.Split(near, ",")
	if len(parts) != 3 {
		return [3]float64{}, fmt.Errorf("invalid near parameter")
	}
	var lat, long, dist float64
	if f, err := strconv.ParseFloat(parts[0], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid latitude in near parameter")
	} else {
		lat = f
	}
	if f, err := strconv.ParseFloat(parts[1], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid longitude in near parameter")
	} else {
		long = f
	}
	if f, err := strconv.ParseFloat(parts[2], 64); err != nil {
		return [3]float64{}, fmt.Errorf("invalid distance in near parameter")
	} else {
		dist = f
	}
	return [3]float64{lat, long, dist}, nil
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
// @success     200 {array}  database.ImageId
// @failure     400 {object} api.StatusBadRequestResponse
// @failure     500 {object} api.StatusServerErrorResponse
func (se SearchEndpoint) Search(c *gin.Context) {
	query := queries.CreateImageQuery()
	if make := c.Query("make"); make != "" {
		query.WithMake(make)
	}
	if makeLike := c.Query("makeLike"); makeLike != "" {
		query.WithMakeLike(makeLike)
	}
	if model := c.Query("model"); model != "" {
		query.WithModel(model)
	}
	if modelLike := c.Query("modelLike"); modelLike != "" {
		query.WithModelLike(modelLike)
	}
	for _, tag := range c.QueryArray("tag") {
		query.WithTag(tag)
	}
	for _, tagLike := range c.QueryArray("tagLike") {
		query.WithTagLike(tagLike)
	}
	if nearParam := c.Query("near"); nearParam != "" {
		near, err := parseNearParam(nearParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
			return
		}
		query.WithLocation(near[0], near[1], near[2])
	}
	if tbParam := c.Query("takenBefore"); tbParam != "" {
		tb, err := strconv.ParseInt(tbParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad takenBefore parameter"})
			return
		}
		query.TakenBefore(time.Unix(tb, 0))
	}
	if taParam := c.Query("takenAfter"); taParam != "" {
		ta, err := strconv.ParseInt(taParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad takenAfter parameter"})
			return
		}
		query.TakenAfter(time.Unix(ta, 0))
	}
	if ubParam := c.Query("uploadedBefore"); ubParam != "" {
		ub, err := strconv.ParseInt(ubParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad uploadedBefore parameter"})
			return
		}
		query.UploadedBefore(time.Unix(ub, 0))
	}
	if uaParam := c.Query("uploadedAfter"); uaParam != "" {
		ua, err := strconv.ParseInt(uaParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "bad uploadedAfter parameter"})
			return
		}
		query.UploadedAfter(time.Unix(ua, 0))
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
	images, err := se.imageRepository.SearchImageOrdered(query, field, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}
