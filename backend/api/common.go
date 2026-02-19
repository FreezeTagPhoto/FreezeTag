package api

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"net/http"
	"strconv"

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
	query := queries.ImageQueryParams{
		Make:           c.Query("make"),
		Model:          c.Query("model"),
		MakeLike:       c.Query("makeLike"),
		ModelLike:      c.Query("modelLike"),
		TakenBefore:    c.Query("takenBefore"),
		TakenAfter:     c.Query("takenAfter"),
		UploadedBefore: c.Query("uploadedBefore"),
		UploadedAfter:  c.Query("uploadedAfter"),
		Near:           c.Query("near"),
		Tags:           c.QueryArray("tag"),
		TagsLike:       c.QueryArray("tagLike"),
	}
	q, err := queries.QueryFromStruct(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, BadRequestResponse{Error: err.Error()})
		return nil
	}
	return q
}
