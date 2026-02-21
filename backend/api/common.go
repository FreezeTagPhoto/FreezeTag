package api

import (
	"fmt"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/database/queries"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IDType interface {
	~uint64
}

// attempts to parse a value into an IDType
func ParseParamIntoID[T IDType](value any) (T, error) {
	strValue, ok := value.(string)
	if !ok {
		return T(0), fmt.Errorf("cannot parse %T into %T", value, T(0))
	}
	num, err := strconv.ParseUint(strValue, 10, 64)
	if err != nil {
		return T(0), fmt.Errorf("Could not parse value '%s' into type %T", strValue, T(0))
	}
	return T(num), nil
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

func QueryPermissionsFromRequest(c *gin.Context) (data.Permissions, error) {
	permissions := c.QueryArray("permission")
	if len(permissions) == 0 {
		return nil, fmt.Errorf("no permissions provided")
	}
	var perms data.Permissions
	for _, perm := range permissions {
		permission, ok := data.GetPermissionFromSlug(perm)
		if !ok {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
		perms = append(perms, permission)
	}
	return perms, nil
}
