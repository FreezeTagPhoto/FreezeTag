package user

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"freezetag/backend/api"
	mockService "freezetag/backend/mocks/AuthService"
	"mime/multipart"

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
	router.GET("/users/permissions/:id", ue.GetPermissions)

	router.POST("/createuser", ue.CreateUser)
	router.POST("/users/permissions/:id", ue.AddPermissions)

	router.DELETE("/users/:id", ue.DeleteUser)
	router.DELETE("/users/permissions/:id", ue.RevokePermissions)

	router.GET("/users/profile-picture/:id", ue.GetProfilePicture)
	router.POST("/users/profile-picture/:id", ue.SetProfilePicture)

}

func TestGetUserOK(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	testUser := &database.PublicUser{
		ID:        1,
		Username:  "testuser",
		CreatedAt: 0,
	}
	mockService.EXPECT().GetUserById(database.UserID(1)).Return(testUser, nil)

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
	mockService := mockService.NewMockAuthService(t)

	mockService.EXPECT().GetUserById(database.UserID(1)).Return(&database.PublicUser{}, errors.New("not found"))

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	testUsers := []database.PublicUser{
		{ID: 1, Username: "testuser1", CreatedAt: 0},
		{ID: 2, Username: "testuser2", CreatedAt: 0},
	}
	mockService.EXPECT().AllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/all", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []database.PublicUser
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, testUsers, got)
}

func TestListAllUsersNoUsers(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	testUsers := []database.PublicUser{} // Empty list of users
	mockService.EXPECT().AllUsers().Return(testUsers, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/all", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAllUsersInternalError(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	mockService.EXPECT().AllUsers().Return(nil, errors.New("database error"))

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	testUsers := []database.PublicUser{} // Empty list of users
	mockService.EXPECT().AllUsers().Return(testUsers, nil)

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	mockService.EXPECT().DeleteUser(database.UserID(1)).Return(nil)

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/one", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestDeleteUserInternalError(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)
	mockService.EXPECT().DeleteUser(database.UserID(1)).Return(errors.New("database error"))

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

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
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Contains(t, got.Error, "failed to create user")
}

// Adding Permissions
func TestAddPermissionsSuccess(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", data.WriteUser.Slug)
	params.Add("permission", data.ReadFiles.Slug)
	params.Add("permission", data.WriteFiles.Slug)
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	mockService.EXPECT().GrantPermissions(database.UserID(1), mock.Anything).Return(nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "permissions granted"
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestAddPermissionsFailId(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/one" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestAddPermissionsFailGrant(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "write:user")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	mockService.EXPECT().GrantPermissions(database.UserID(1), data.Permissions{data.WriteUser}).Return(errors.New("database error"))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "failed to grant permissions: database error"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestAddPermissionsNoPermissions(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

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
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "write:user")
	params.Add("permission", "read:files")
	params.Add("permission", "write:files")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	mockService.EXPECT().RevokePermissions(database.UserID(1), data.Permissions{data.WriteUser, data.ReadFiles, data.WriteFiles}).Return(nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "permissions revoked"
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestRevokePermissionsFailId(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/one" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestRevokePermissionsFailGrant(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	params := url.Values{}
	params.Add("permission", "write:user")
	reqURL := "/users/permissions/1?" + params.Encode() // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	mockService.EXPECT().RevokePermissions(database.UserID(1), data.Permissions{data.WriteUser}).Return(errors.New("database error"))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedError := "failed to revoke permissions: database error"
	assert.Equal(t, api.BadRequestResponse{Error: expectedError}, got)
}

func TestRevokePermissionsNoPermissions(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

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

func TestGetPermissions(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	expected := data.Permissions{{Slug: "read:user", Name: "Read Users", Description: "foo"}}
	mockService.EXPECT().
		GetUserPermissions(mock.Anything).
		Return(expected, nil)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got data.Permissions
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetPermissionsFail(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)
	mockService.EXPECT().
		GetUserPermissions(mock.Anything).
		Return(nil, fmt.Errorf("big deal"))

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.ServerErrorResponse{Error: "big deal"}, got)
}

func TestGetPermissionsBadId(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/permissions/foo" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}

func TestGetProfilePicture(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)
	mockService.EXPECT().
		GetUserProfilePicture(database.UserID(1)).
		Return([]byte("fake image bytes"), nil)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)

	expected := []byte("fake image bytes")
	assert.Equal(t, expected, w.Body.Bytes())
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetProfilePictureBadId(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/foo" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}

func TestGetProfilePictureInternalError(t *testing.T) {
	mockService := mockService.NewMockAuthService(t)
	mockService.EXPECT().
		GetUserProfilePicture(database.UserID(1)).
		Return(nil, errors.New("database error"))

	router := gin.Default()
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}

func writeTestFile(writer *multipart.Writer, fieldname string, filename string, content []byte) error {
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}

func TestSetUserProfilePictureSuccess(t *testing.T) {
	router := gin.Default()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writeTestFile(writer, "picture", "testfile.png", []byte("new fake image bytes")))
	require.NoError(t, writer.Close())

	mockService := mockService.NewMockAuthService(t)
	mockService.EXPECT().
		SetUserProfilePicture(database.UserID(1), []byte("new fake image bytes"), mock.Anything).
		Return(nil)
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expectedMessage := "profile picture updated"
	t.Logf("got: %+v", got)
	assert.Equal(t, api.MessageResponse{Message: expectedMessage}, got)
}

func TestSetUserProfilePictureBadId(t *testing.T) {
	router := gin.Default()
	InitUserEndpoint(mockService.NewMockAuthService(t)).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/foo" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}

func TestSetUserProfilePictureInternalError(t *testing.T) {
	router := gin.Default()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writeTestFile(writer, "picture", "testfile.png", []byte("new fake image bytes")))
	require.NoError(t, writer.Close())

	mockService := mockService.NewMockAuthService(t)
	mockService.EXPECT().
		SetUserProfilePicture(database.UserID(1), []byte("new fake image bytes"), mock.Anything).
		Return(errors.New("database error"))
	InitUserEndpoint(mockService).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}

func TestSetUserProfilePictureNoPictureField(t *testing.T) {
	router := gin.Default()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())

	InitUserEndpoint(mockService.NewMockAuthService(t)).RegisterEndpoints(router)

	reqURL := "/users/profile-picture/1" // properly encodes & joins parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got)
}