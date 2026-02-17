package api

import (
	"freezetag/backend/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserIdFromStringSuccess(t *testing.T) {
	id, err := GetUserIDFromString("123")
	assert.NoError(t, err)
	assert.Equal(t, database.UserID(123), id)
}

func TestGetUserIdFromStringInvalid(t *testing.T) {
	id, err := GetUserIDFromString("one")
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestGetUserIdFromStringNegative(t *testing.T) {
	id, err := GetUserIDFromString("-5")
	assert.NoError(t, err)
	assert.Equal(t, database.UserID(-5), id)
}
