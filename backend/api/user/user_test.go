package user

import (
	"encoding/json"
	"errors"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/UserRepository"
	"freezetag/backend/pkg/database"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserOK(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	testUser := &database.PublicUser{
		ID:        1,
		Username:  "testuser",
		CreatedAt: 0,
	}
	mockRepo.EXPECT().GetUserByID(database.UserID(1)).Return(testUser, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got database.PublicUser
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, *testUser, got)
}

func TestGetUserUserIDServerError(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockRepo.EXPECT().GetUserByID(database.UserID(1)).Return(&database.PublicUser{}, errors.New("not found"))

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.StatusBadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.StatusBadRequestResponse{Error: "User not found"}, got)
}

func TestGetUserUserIDBadIDParse(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/one", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.StatusBadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.StatusBadRequestResponse{Error: "Invalid user ID parameter: one"}, got)
}

func TestListAllUsers(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{
		{ID: 1, Username: "testuser1", CreatedAt: 0},
		{ID: 2, Username: "testuser2", CreatedAt: 0},
	}
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []*database.PublicUser
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, testUsers, got)
}

func TestListAllUsersNoUsers(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{} // Empty list of users
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAllUsersInternalError(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockRepo.EXPECT().ListAllUsers().Return(nil, errors.New("database error"))

	router := gin.Default()
	InitUserEndpoint(mockRepo).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{} // Empty list of users
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.StatusBadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.StatusBadRequestResponse{Error: "Failed to list users"}, got)
}
