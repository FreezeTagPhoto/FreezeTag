package api

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUserIDFromString(userIDString string) (database.UserID, error) {
	var id database.UserID
	if num, err := strconv.ParseUint(userIDString, 10, 64); err != nil {
		return id, fmt.Errorf("invalid user ID parameter: %s", userIDString)
	} else {
		id = database.UserID(num)
	}
	return id, nil
}

func GetRequestQuery(c *gin.Context) *queries.ImageQuery {
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
			c.JSON(http.StatusBadRequest, BadRequestResponse{Error: err.Error()})
			return nil
		}
		query.WithLocation(near[0], near[1], near[2])
	}
	if tbParam := c.Query("takenBefore"); tbParam != "" {
		tb, err := strconv.ParseInt(tbParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, BadRequestResponse{Error: "bad takenBefore parameter"})
			return nil
		}
		query.TakenBefore(time.Unix(tb, 0))
	}
	if taParam := c.Query("takenAfter"); taParam != "" {
		ta, err := strconv.ParseInt(taParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, BadRequestResponse{Error: "bad takenAfter parameter"})
			return nil
		}
		query.TakenAfter(time.Unix(ta, 0))
	}
	if ubParam := c.Query("uploadedBefore"); ubParam != "" {
		ub, err := strconv.ParseInt(ubParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, BadRequestResponse{Error: "bad uploadedBefore parameter"})
			return nil
		}
		query.UploadedBefore(time.Unix(ub, 0))
	}
	if uaParam := c.Query("uploadedAfter"); uaParam != "" {
		ua, err := strconv.ParseInt(uaParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, BadRequestResponse{Error: "bad uploadedAfter parameter"})
			return nil
		}
		query.UploadedAfter(time.Unix(ua, 0))
	}
	return query
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
