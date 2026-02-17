package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"freezetag/backend/api"
	mockService "freezetag/backend/mocks/AuthService"
	mocks "freezetag/backend/mocks/UserRepository"

	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (ue UserEndpoint) RegisterEndpoints(router gin.IRoutes) {
	router.GET("/users/:id", ue.GetUser)
	router.GET("/users/all", ue.ListUsers)

	router.POST("/createuser", ue.CreateUser)
	router.POST("/users/permissions/:id", ue.AddPermissions)

	router.DELETE("/users/:id", ue.DeleteUser)
	router.DELETE("/users/permissions/:id", ue.RevokePermissions)
}

func TestGetUserOK(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	testUser := &database.PublicUser{
		ID:        1,
		Username:  "testuser",
		CreatedAt: 0,
	}
	mockRepo.EXPECT().GetUserByID(database.UserID(1)).Return(testUser, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got database.PublicUser
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, *testUser, got)
}

func TestGetUserUserIDServerError(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	mockRepo.EXPECT().GetUserByID(database.UserID(1)).Return(&database.PublicUser{}, errors.New("not found"))

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.BadRequestResponse{Error: "User not found"}, got)
}

func TestGetUserUserIDBadIDParse(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/one", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.BadRequestResponse{Error: "Invalid user ID parameter: one"}, got)
}

func TestListAllUsers(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{
		{ID: 1, Username: "testuser1", CreatedAt: 0},
		{ID: 2, Username: "testuser2", CreatedAt: 0},
	}
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/all", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []*database.PublicUser
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, testUsers, got)
}

func TestListAllUsersNoUsers(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{} // Empty list of users
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/all", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAllUsersInternalError(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	mockRepo.EXPECT().ListAllUsers().Return(nil, errors.New("database error"))

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	testUsers := []*database.PublicUser{} // Empty list of users
	mockRepo.EXPECT().ListAllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/all", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, api.BadRequestResponse{Error: "Failed to list users"}, got)
}

func TestDeleteUserSuccess(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	mockRepo.EXPECT().DeleteUser(database.UserID(1)).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "user 1 deleted"
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestDeleteUserBadIDParse(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/one", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "invalid user ID parameter: one"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestDeleteUserInternalError(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)
	mockRepo.EXPECT().DeleteUser(database.UserID(1)).Return(errors.New("database error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "Failed to delete user"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestCreateUserSuccess(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	loginCredentials := api.LoginCredentials{
		Username: "testuser",
		Password: "password",
	}
	testUser := &database.PublicUser{
		ID:        1,
		Username:  "testuser",
		CreatedAt: 0,
	}
	mockService.EXPECT().AddUser("testuser", "password").Return(testUser, nil)
	jsonBytes, err := json.Marshal(loginCredentials)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/createuser", bytes.NewReader(jsonBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got database.PublicUser
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, *testUser, got)
}

func TestCreateUserInvalidCredentialBinds(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	invalidJSON := []byte(`garbage data`) // password should be a string

	req, err := http.NewRequest("POST", "/createuser", bytes.NewReader(invalidJSON))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Contains(t, got.Error, "invalid request")
}

func TestCreateUserAddUserFails(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	loginCredentials := api.LoginCredentials{
		Username: "testuser",
		Password: "password",
	}
	jsonBytes, err := json.Marshal(loginCredentials)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/createuser", bytes.NewReader(jsonBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	mockService.EXPECT().AddUser("testuser", "password").Return(nil, errors.New("database error"))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Contains(t, got.Error, "failed to create user")
}

// Adding Permissions
func TestAddPermissionsSuccess(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", data.CreateUser.Slug)
	params.Add("permission", data.ReadFiles.Slug)
	params.Add("permission", data.WriteFiles.Slug)
	params.Add("permission", data.DeleteUser.Slug)
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	mockRepo.EXPECT().GrantPermissions(database.UserID(1), mock.Anything).Return(nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "permissions granted"
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestAddPermissionsFailId(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/one" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "invalid user ID parameter: one"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestAddPermissionsFailGrant(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "create:user")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	mockRepo.EXPECT().GrantPermissions(database.UserID(1), data.Permissions{data.CreateUser}).Return(errors.New("database error"))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "failed to grant permissions: database error"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestAddPermissionsNoPermissions(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "no permissions provided"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

// Deleting Permissions

func TestRevokePermissionsSuccess(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "create:user")
	params.Add("permission", "read:files")
	params.Add("permission", "write:files")
	params.Add("permission", "delete:user")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	mockRepo.EXPECT().RevokePermissions(database.UserID(1), data.Permissions{data.CreateUser, data.ReadFiles, data.WriteFiles, data.DeleteUser}).Return(nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "permissions revoked"
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestRevokePermissionsFailId(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/one" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "invalid user ID parameter: one"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestRevokePermissionsFailGrant(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "create:user")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	mockRepo.EXPECT().RevokePermissions(database.UserID(1), data.Permissions{data.CreateUser}).Return(errors.New("database error"))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "failed to revoke permissions: database error"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestRevokePermissionsNoPermissions(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockRepo, mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "no permissions provided"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}
