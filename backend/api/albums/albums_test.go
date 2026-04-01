package albums

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"freezetag/backend/api"
	albumMocks "freezetag/backend/mocks/AlbumDatabase"
	"freezetag/backend/pkg/database"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAlbumRouter(t *testing.T, repo database.AlbumDatabase, withUser bool, userID any) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if withUser {
		router.Use(func(c *gin.Context) {
			c.Set("userID", fmt.Sprintf("%v", userID))
			c.Next()
		})
	}

	ae := InitAlbumEndpoint(repo)
	albumGroup := router.Group("/album")
	albumGroup.POST("/create", ae.CreateAlbum)
	albumGroup.POST("/add_image", ae.AddImageToAlbum)
	albumGroup.POST("/add_image_by_name", ae.AddImageToAlbumByName)
	albumGroup.GET("/names", ae.ListVisibleAlbums)
	albumGroup.GET("/album_names/:id", ae.ListImageAlbums)
	albumGroup.GET("/images/:name", ae.GetAlbumImagesByName)
	albumGroup.POST("/rename", ae.RenameAlbum)
	albumGroup.POST("/delete", ae.DeleteAlbum)
	albumGroup.POST("/set_visibility", ae.SetAlbumVisibility)
	albumGroup.POST("/set_permission", ae.SetUserAlbumPermission)
	albumGroup.GET("/list", ae.ListAlbums)

	return router
}

func doJSONRequest(t *testing.T, router *gin.Engine, method string, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		jsonData, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCreateAlbumSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().CreateAlbum("trip", database.UserID(7), database.PrivacyLevel(1)).Return(database.AlbumId(42), nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(7))
	w := doJSONRequest(t, router, http.MethodPost, "/album/create", AlbumCreateRequest{Name: "trip", VisibilityMode: 1})

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.AlbumCreateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, database.AlbumId(42), got.AlbumID)
}

func TestCreateAlbumMissingUserID(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	router := setupAlbumRouter(t, repo, false, nil)
	w := doJSONRequest(t, router, http.MethodPost, "/album/create", AlbumCreateRequest{Name: "trip", VisibilityMode: 1})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "userID not found in context", got.Error)
}

func TestAddImageToAlbumByNameNotFound(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumIdByName("missing", database.UserID(9)).Return(database.AlbumId(0), nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(9))
	w := doJSONRequest(t, router, http.MethodPost, "/album/add_image_by_name", AddImageByNameRequest{ImageId: 11, AlbumName: "missing"})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "album not found", got.Error)
}

func TestAddImageToAlbumByNameSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumIdByName("shared", database.UserID(9)).Return(database.AlbumId(88), nil).Once()
	repo.EXPECT().SetImageAlbum(database.ImageId(13), database.AlbumId(88), database.UserID(9)).Return(nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(9))
	w := doJSONRequest(t, router, http.MethodPost, "/album/add_image_by_name", AddImageByNameRequest{ImageId: 13, AlbumName: "shared"})

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "Image added to album successfully", got.Message)
}

func TestSetAlbumVisibilityMissingValue(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	router := setupAlbumRouter(t, repo, true, database.UserID(2))

	w := doJSONRequest(t, router, http.MethodPost, "/album/set_visibility", map[string]any{"album_id": 5})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "visibility_mode is required", got.Error)
}

func TestSetUserAlbumPermissionSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().SetUserAlbumPermission(database.AlbumId(3), database.UserID(44), database.PrivacyLevel(1), database.UserID(7)).Return(nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(7))
	w := doJSONRequest(t, router, http.MethodPost, "/album/set_permission", UserAlbumPermissionRequest{AlbumId: 3, TargetUserId: 44, Permission: 1})

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "User album permission updated successfully", got.Message)
}

func TestListAlbumsSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumIds(database.UserID(7)).Return([]database.AlbumId{3}, nil).Once()
	repo.EXPECT().GetAlbumNameById(database.AlbumId(3), database.UserID(7)).Return("shared-album", nil).Once()
	repo.EXPECT().GetAlbumOwner(database.AlbumId(3)).Return(database.UserID(7), nil).Once()
	repo.EXPECT().CanManageAlbum(database.AlbumId(3), database.UserID(7)).Return(true, nil).Once()
	repo.EXPECT().GetAlbumSharedUsers(database.AlbumId(3)).Return([]database.AlbumSharedUser{{UserID: 44, Permission: 2}}, nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(7))
	w := doJSONRequest(t, router, http.MethodGet, "/album/list", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []AlbumListItemResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, database.AlbumId(3), got[0].ID)
	assert.Equal(t, "shared-album", got[0].Name)
	assert.Equal(t, database.UserID(7), got[0].OwnerID)
	assert.True(t, got[0].CanManageSharing)
	assert.Equal(t, []database.UserID{44}, got[0].SharedUserIDs)
	require.Len(t, got[0].SharedUsers, 1)
	assert.Equal(t, database.UserID(44), got[0].SharedUsers[0].UserID)
	assert.Equal(t, database.PrivacyLevel(2), got[0].SharedUsers[0].Permission)
}

func TestListVisibleAlbumsSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumNames(database.UserID(2)).Return([]string{"a", "b"}, nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(2))
	w := doJSONRequest(t, router, http.MethodGet, "/album/names", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestListImageAlbumsBadImageID(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	router := setupAlbumRouter(t, repo, true, database.UserID(5))
	w := doJSONRequest(t, router, http.MethodGet, "/album/album_names/not-an-id", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEmpty(t, got.Error)
}

func TestGetAlbumImagesByNameNotFound(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumIdByName("missing", database.UserID(1)).Return(database.AlbumId(0), nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(1))
	w := doJSONRequest(t, router, http.MethodGet, "/album/images/missing", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "album not found", got.Error)
}

func TestGetAlbumImagesByNameSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().GetAlbumIdByName("vacation", database.UserID(1)).Return(database.AlbumId(12), nil).Once()
	repo.EXPECT().GetAlbumImages(database.AlbumId(12), database.UserID(1)).Return([]database.ImageId{3, 4}, nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(1))
	w := doJSONRequest(t, router, http.MethodGet, "/album/images/vacation", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, []database.ImageId{3, 4}, got)
}

func TestRenameAlbumRepositoryError(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().RenameAlbum("a", "b", database.UserID(2)).Return(errors.New("boom")).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(2))
	w := doJSONRequest(t, router, http.MethodPost, "/album/rename", RenameAlbumRequest{OldName: "a", NewName: "b"})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "boom", got.Error)
}

func TestDeleteAlbumSuccess(t *testing.T) {
	repo := albumMocks.NewMockAlbumDatabase(t)
	repo.EXPECT().RemoveAlbum("old", database.UserID(3)).Return(nil).Once()

	router := setupAlbumRouter(t, repo, true, database.UserID(3))
	w := doJSONRequest(t, router, http.MethodPost, "/album/delete", DeleteAlbumRequest{AlbumName: "old"})

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "Album deleted successfully", got.Message)
}
