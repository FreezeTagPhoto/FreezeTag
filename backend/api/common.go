package api

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"strconv"
)

func GetUserIDFromString(userIDString string) (database.UserID, error) {
	var id database.UserID
	if num, err := strconv.ParseInt(userIDString, 10, 64); err != nil {
		return id, fmt.Errorf("invalid user ID parameter: %s", userIDString)
	} else {
		id = database.UserID(num)
	}
	return id, nil
}