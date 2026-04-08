package albums

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/AlbumDatabase"
	"freezetag/backend/pkg/database"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAlbumTestRouter(endpoint AlbumEndpoint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if userID := c.GetHeader("uid"); userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	})

	router.POST("/album", endpoint.CreateAlbum)
	router.GET("/album", endpoint.ListAlbums)
	router.GET("/album/:id", endpoint.GetAlbumInfo)
	router.PATCH("/album/:id/name", endpoint.RenameAlbum)
	router.DELETE("/album/:id", endpoint.DeleteAlbum)
	router.DELETE("/album/:id/images/:image_id", endpoint.RemoveImageFromAlbum)
	router.GET("/album/:id/images", endpoint.ListAlbumImages)
	router.POST("/album/:id/images", endpoint.AddImageToAlbum)
	router.PATCH("/album/:id/visibility", endpoint.ChangeAlbumVisibility)
	router.PUT("/album/:id/permissions", endpoint.SetUserAlbumPermission)
	router.GET("/album/:id/permissions", endpoint.GetAlbumPermissions)
	router.GET("/album/image/:id", endpoint.ListImageAlbums)

	return router
}

func newAlbumTest(t *testing.T) (*mocks.MockAlbumDatabase, *gin.Engine) {
	t.Helper()
	m := mocks.NewMockAlbumDatabase(t)
	return m, newAlbumTestRouter(InitAlbumEndpoint(m))
}

func runRequest(router *gin.Engine, method, path, uid string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if uid != "" {
		req.Header.Set("uid", uid)
	}
	router.ServeHTTP(w, req)
	return w
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func decodeResponseToJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var got T
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	return got
}

func TestCreateAlbumSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		CreateAlbum("Vacation", database.UserID(9), database.ALBUM_PUBLIC).
		Return(database.AlbumID(42), nil)

	body := mustMarshal(t, AlbumCreateRequest{Name: "Vacation", VisibilityMode: database.ALBUM_PUBLIC})
	w := runRequest(router, http.MethodPost, "/album", "9", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.AlbumCreateResponse](t, w)
	assert.Equal(t, api.AlbumCreateResponse{AlbumID: database.AlbumID(42)}, got)
}

func TestCreateAlbumBadBody(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodPost, "/album", "9", bytes.NewBufferString(`{"name":`), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestListAlbumsSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	testAlbums := []database.Album{
		{ID: 1, Name: "A", OwnerID: 11, AlbumPrivacy: database.ALBUM_PRIVATE, VisbilityLevel: database.USER_PRIVATE},
		{ID: 2, Name: "B", OwnerID: 11, AlbumPrivacy: database.ALBUM_PUBLIC, VisbilityLevel: database.USER_PUBLIC},
	}
	m.EXPECT().GetAlbums(database.UserID(11)).Return(testAlbums, nil)

	w := runRequest(router, http.MethodGet, "/album", "11", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[[]database.Album](t, w)
	assert.ElementsMatch(t, testAlbums, got)
}

func TestRenameAlbumSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().RenameAlbum(database.AlbumID(7), "Renamed", database.UserID(15)).Return(nil)

	body := mustMarshal(t, RenameAlbumRequest{NewName: "Renamed"})
	w := runRequest(router, http.MethodPatch, "/album/7/name", "15", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "Album renamed successfully"}, got)
}

func TestDeleteAlbumMissingUserID(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodDelete, "/album/7", "", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: "userID not found in context"}, got)
}

func TestListAlbumImagesBadAlbumID(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodGet, "/album/not-an-id/images", "4", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func testMissingUserID(t *testing.T, method, path string) {
	t.Helper()
	_, router := newAlbumTest(t)
	w := runRequest(router, method, path, "", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: "userID not found in context"}, got)
}

func testUnparseableUserID(t *testing.T, method, path string) {
	t.Helper()
	_, router := newAlbumTest(t)
	w := runRequest(router, method, path, "not-a-number", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: "userID in context is not of type UserID"}, got)
}

func TestCreateAlbumMissingUserID(t *testing.T) {
	testMissingUserID(t, http.MethodPost, "/album")
	testMissingUserID(t, http.MethodGet, "/album")
	testMissingUserID(t, http.MethodGet, "/album/1")
	testMissingUserID(t, http.MethodPatch, "/album/1/name")
	testMissingUserID(t, http.MethodDelete, "/album/1")
	testMissingUserID(t, http.MethodGet, "/album/1/images")
	testMissingUserID(t, http.MethodPost, "/album/1/images")
	testMissingUserID(t, http.MethodDelete, "/album/1/images/10")
	testMissingUserID(t, http.MethodPatch, "/album/1/visibility")
	testMissingUserID(t, http.MethodPut, "/album/1/permissions")
	testMissingUserID(t, http.MethodGet, "/album/1/permissions")
	testMissingUserID(t, http.MethodGet, "/album/image/1")
}

func TestCreateAlbumUnparseableUserID(t *testing.T) {
	testUnparseableUserID(t, http.MethodPost, "/album")
	testUnparseableUserID(t, http.MethodGet, "/album")
	testUnparseableUserID(t, http.MethodGet, "/album/1")
	testUnparseableUserID(t, http.MethodPatch, "/album/1/name")
	testUnparseableUserID(t, http.MethodDelete, "/album/1")
	testUnparseableUserID(t, http.MethodGet, "/album/1/images")
	testUnparseableUserID(t, http.MethodPost, "/album/1/images")
	testUnparseableUserID(t, http.MethodDelete, "/album/1/images/10")
	testUnparseableUserID(t, http.MethodPatch, "/album/1/visibility")
	testUnparseableUserID(t, http.MethodPut, "/album/1/permissions")
	testUnparseableUserID(t, http.MethodGet, "/album/1/permissions")
	testUnparseableUserID(t, http.MethodGet, "/album/image/1")
}

func TestCreateAlbumFail(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		CreateAlbum("Vacation", database.UserID(9), database.ALBUM_PUBLIC).
		Return(database.AlbumID(0), assert.AnError)

	body := mustMarshal(t, AlbumCreateRequest{Name: "Vacation", VisibilityMode: database.ALBUM_PUBLIC})
	w := runRequest(router, http.MethodPost, "/album", "9", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestAddImageToAlbumSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetImageAlbum(database.ImageID(10), database.AlbumID(5), database.UserID(3)).
		Return(nil)

	body := mustMarshal(t, AlbumImageRequest{ImageID: 10})
	w := runRequest(router, http.MethodPost, "/album/5/images", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "Image added to album successfully"}, got)
}

func TestAddImageToAlbumBadBody(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, []byte(`asdasdgfjhbiou`))
	w := runRequest(router, http.MethodPost, "/album/5/images", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestAddImageToAlbumBadJSONBody(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, struct{ InvalidField string }{
		InvalidField: "not a valid request body",
	})
	w := runRequest(router, http.MethodPost, "/album/5/images", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestAddImageToAlbumFailParse(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, AlbumImageRequest{ImageID: 10})
	w := runRequest(router, http.MethodPost, "/album/not-an-id/images", "3", bytes.NewReader(body), "application/json")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestAddImageToAlbumFailDatabase(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetImageAlbum(database.ImageID(10), database.AlbumID(5), database.UserID(3)).
		Return(assert.AnError)

	body := mustMarshal(t, AlbumImageRequest{ImageID: 10})
	w := runRequest(router, http.MethodPost, "/album/5/images", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestRemoveImageFromAlbumSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		RemoveImageFromAlbum(database.ImageID(10), database.AlbumID(5), database.UserID(3)).
		Return(nil)

	w := runRequest(router, http.MethodDelete, "/album/5/images/10", "3", nil, "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "Image removed from album successfully"}, got)
}

func TestRemoveImageFromAlbumFailParse(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodDelete, "/album/not-an-id/images/10", "3", nil, "application/json")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestRemoveImageFromAlbumFailParseTwo(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodDelete, "/album/5/images/asdfasdf", "3", nil, "application/json")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'asdfasdf' into type database.ImageID"}, got)
}

func TestRemoveImageFromAlbumFailDatabase(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		RemoveImageFromAlbum(database.ImageID(10), database.AlbumID(5), database.UserID(3)).
		Return(assert.AnError)
	body := mustMarshal(t, AlbumImageRequest{ImageID: 10})
	w := runRequest(router, http.MethodDelete, "/album/5/images/10", "3", bytes.NewReader(body), "application/json")
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestChangeAlbumVisibilitySuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetAlbumVisibility(database.AlbumID(5), database.ALBUM_PRIVATE, database.UserID(3)).
		Return(nil)

	vis := database.ALBUM_PRIVATE
	body := mustMarshal(t, AlbumModifyRequest{AlbumID: database.AlbumID(5), VisibilityMode: &vis})
	w := runRequest(router, http.MethodPatch, "/album/5/visibility", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "Album visibility updated successfully"}, got)
}

func TestChangeAlbumVisibilityBadBody(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodPatch, "/album/5/visibility", "3", bytes.NewBufferString(`{"album_id":`), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestChangeAlbumVisibilityMissingVisibilityMode(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, AlbumModifyRequest{AlbumID: database.AlbumID(5)})
	w := runRequest(router, http.MethodPatch, "/album/5/visibility", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "visibility_mode is required"}, got)
}

func TestChangeAlbumVisibilityInvalidVisibilityMode(t *testing.T) {
	_, router := newAlbumTest(t)

	vis := database.GlobalPrivacy(2)
	body := mustMarshal(t, AlbumModifyRequest{AlbumID: database.AlbumID(5), VisibilityMode: &vis})
	w := runRequest(router, http.MethodPatch, "/album/5/visibility", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestChangeAlbumVisibilityFailDatabase(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetAlbumVisibility(database.AlbumID(5), database.ALBUM_PRIVATE, database.UserID(3)).
		Return(assert.AnError)

	vis := database.ALBUM_PRIVATE
	body := mustMarshal(t, AlbumModifyRequest{AlbumID: database.AlbumID(5), VisibilityMode: &vis})
	w := runRequest(router, http.MethodPatch, "/album/5/visibility", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestSetUserAlbumPermissionSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetUserAlbumPermission(database.AlbumID(5), database.UserID(7), database.USER_PUBLIC, database.UserID(3)).
		Return(nil)

	body := mustMarshal(t, UserAlbumPermissionRequest{
		AlbumID:      database.AlbumID(5),
		TargetUserID: database.UserID(7),
		Permission:   database.USER_PUBLIC,
	})
	w := runRequest(router, http.MethodPut, "/album/5/permissions", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "User album permission updated successfully"}, got)
}

func TestSetUserAlbumPermissionBadBody(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodPut, "/album/5/permissions", "3", bytes.NewBufferString(`{"album_id":`), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestSetUserAlbumPermissionMissingTargetUserID(t *testing.T) {
	_, router := newAlbumTest(t)

	body := mustMarshal(t, UserAlbumPermissionRequest{
		AlbumID:    database.AlbumID(5),
		Permission: database.USER_PUBLIC,
	})
	w := runRequest(router, http.MethodPut, "/album/5/permissions", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestSetUserAlbumPermissionInvalidPermission(t *testing.T) {
	_, router := newAlbumTest(t)

	body := mustMarshal(t, UserAlbumPermissionRequest{
		AlbumID:      database.AlbumID(5),
		TargetUserID: database.UserID(7),
		Permission:   database.UserPrivacy(5),
	})
	w := runRequest(router, http.MethodPut, "/album/5/permissions", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestSetUserAlbumPermissionFailDatabase(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		SetUserAlbumPermission(database.AlbumID(5), database.UserID(7), database.USER_PUBLIC, database.UserID(3)).
		Return(assert.AnError)

	body := mustMarshal(t, UserAlbumPermissionRequest{
		AlbumID:      database.AlbumID(5),
		TargetUserID: database.UserID(7),
		Permission:   database.USER_PUBLIC,
	})
	w := runRequest(router, http.MethodPut, "/album/5/permissions", "3", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestGetAlbumPermissionsSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	sharedUsers := []database.AlbumSharedUser{
		{UserID: 7, Permission: database.ALBUM_PRIVATE},
		{UserID: 8, Permission: database.ALBUM_PUBLIC},
	}
	m.EXPECT().GetAlbumSharedUsers(database.AlbumID(5), database.UserID(3)).Return(sharedUsers, nil)

	w := runRequest(router, http.MethodGet, "/album/5/permissions", "3", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[[]database.AlbumSharedUser](t, w)
	assert.ElementsMatch(t, sharedUsers, got)
}

func TestGetAlbumPermissionsBadAlbumID(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodGet, "/album/not-an-id/permissions", "3", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestGetAlbumPermissionsDatabaseFailure(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().GetAlbumSharedUsers(database.AlbumID(5), database.UserID(3)).Return(nil, assert.AnError)
	w := runRequest(router, http.MethodGet, "/album/5/permissions", "3", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestListAlbumsRepositoryFailure(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().GetAlbums(database.UserID(11)).Return(nil, assert.AnError)
	w := runRequest(router, http.MethodGet, "/album", "11", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestListAlbumRepositoryImages(t *testing.T) {
	m, router := newAlbumTest(t)
	testImages := []database.ImageID{100, 101, 102}
	m.EXPECT().GetAlbumImages(database.AlbumID(5), database.UserID(11)).Return(testImages, nil)
	w := runRequest(router, http.MethodGet, "/album/5/images", "11", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[[]database.ImageID](t, w)
	assert.ElementsMatch(t, testImages, got)
}

func TestListAlbumRepositoryImagesDatabaseFailure(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().GetAlbumImages(database.AlbumID(5), database.UserID(11)).Return(nil, assert.AnError)
	w := runRequest(router, http.MethodGet, "/album/5/images", "11", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestListImageAlbumsSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	testAlbums := []database.Album{
		{ID: 1, Name: "A", OwnerID: 11, AlbumPrivacy: database.ALBUM_PRIVATE, VisbilityLevel: database.USER_PRIVATE},
		{ID: 2, Name: "B", OwnerID: 12, AlbumPrivacy: database.ALBUM_PUBLIC, VisbilityLevel: database.USER_PUBLIC},
	}
	m.EXPECT().GetAssociatedAlbums(database.ImageID(9), database.UserID(11)).Return(testAlbums, nil)
	w := runRequest(router, http.MethodGet, "/album/image/9", "11", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[[]database.Album](t, w)
	assert.ElementsMatch(t, testAlbums, got)
}

func TestListImageAlbumsBadImageID(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodGet, "/album/image/not-an-id", "11", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.ImageID"}, got)
}

func TestListImageAlbumsDatabaseFailure(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().GetAssociatedAlbums(database.ImageID(9), database.UserID(11)).Return(nil, assert.AnError)
	w := runRequest(router, http.MethodGet, "/album/image/9", "11", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestGetAlbumInfoSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	testAlbum := database.Album{ID: 5, Name: "Road Trip", OwnerID: 3, AlbumPrivacy: database.ALBUM_PRIVATE, VisbilityLevel: database.USER_ADMIN}
	m.EXPECT().GetAlbum(database.AlbumID(5), database.UserID(3)).Return(testAlbum, nil)
	w := runRequest(router, http.MethodGet, "/album/5", "3", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[database.Album](t, w)
	assert.Equal(t, testAlbum, got)
}

func TestGetAlbumInfoBadAlbumID(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodGet, "/album/not-an-id", "3", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestGetAlbumInfoRepositoryFailure(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().GetAlbum(database.AlbumID(5), database.UserID(3)).Return(database.Album{}, assert.AnError)
	w := runRequest(router, http.MethodGet, "/album/5", "3", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestDeleteAlbumIDParseError(t *testing.T) {
	_, router := newAlbumTest(t)
	w := runRequest(router, http.MethodDelete, "/album/not-an-id", "3", nil, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestDeleteAlbumSuccess(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		RemoveAlbum(database.AlbumID(5), database.UserID(3)).
		Return(nil)
	w := runRequest(router, http.MethodDelete, "/album/5", "3", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	got := decodeResponseToJSON[api.MessageResponse](t, w)
	assert.Equal(t, api.MessageResponse{Message: "Album deleted successfully"}, got)
}

func TestDeleteAlbumRepositoryError(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().
		RemoveAlbum(database.AlbumID(5), database.UserID(3)).
		Return(assert.AnError)
	w := runRequest(router, http.MethodDelete, "/album/5", "3", nil, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestRenameAlbumFailToParse(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, RenameAlbumRequest{NewName: "Renamed"})
	w := runRequest(router, http.MethodPatch, "/album/not-an-id/name", "15", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func TestRenameAlbumFailToMarshal(t *testing.T) {
	_, router := newAlbumTest(t)
	body := mustMarshal(t, struct{ InvalidField string }{
		InvalidField: "not a valid request body",
	})
	w := runRequest(router, http.MethodPatch, "/album/7/name", "15", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	got := decodeResponseToJSON[api.BadRequestResponse](t, w)
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestRenameAlbumRepositoryFails(t *testing.T) {
	m, router := newAlbumTest(t)
	m.EXPECT().RenameAlbum(database.AlbumID(7), "Renamed", database.UserID(15)).Return(assert.AnError)
	body := mustMarshal(t, RenameAlbumRequest{NewName: "Renamed"})
	w := runRequest(router, http.MethodPatch, "/album/7/name", "15", bytes.NewReader(body), "application/json")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	got := decodeResponseToJSON[api.ServerErrorResponse](t, w)
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}
