package albums

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/AlbumDatabase"
	"freezetag/backend/pkg/database"
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
	router.PATCH("/album/:id/name", endpoint.RenameAlbum)
	router.DELETE("/album/:id", endpoint.DeleteAlbum)
	router.GET("/album/:id/images", endpoint.ListAlbumImages)
	router.POST("/album/:id/images", endpoint.AddImageToAlbum)

	return router
}

func TestCreateAlbumSuccess(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	m.EXPECT().
		CreateAlbum("Vacation", database.UserID(9), database.ALBUM_PUBLIC).
		Return(database.AlbumID(42), nil)

	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal(AlbumCreateRequest{Name: "Vacation", VisibilityMode: database.ALBUM_PUBLIC})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "9")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got api.AlbumCreateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.AlbumCreateResponse{AlbumID: database.AlbumID(42)}, got)
}

func TestCreateAlbumBadBody(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "9")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestListAlbumsSuccess(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	testAlbums := []database.Album{
		{ID: 1, Name: "A", OwnerID: 11, AlbumPrivacy: database.ALBUM_PRIVATE, VisbilityLevel: database.VIS_PRIVATE},
		{ID: 2, Name: "B", OwnerID: 11, AlbumPrivacy: database.ALBUM_PUBLIC, VisbilityLevel: database.VIS_PUBLIC},
	}
	m.EXPECT().GetAlbums(database.UserID(11)).Return(testAlbums, nil)

	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/album", nil)
	req.Header.Set("uid", "11")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got []database.Album
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.ElementsMatch(t, testAlbums, got)
}

func TestRenameAlbumSuccess(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	m.EXPECT().RenameAlbum(database.AlbumID(7), "Renamed", database.UserID(15)).Return(nil)

	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal(RenameAlbumRequest{NewName: "Renamed"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/album/7/name", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "15")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got api.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.MessageResponse{Message: "Album renamed successfully"}, got)
}

func TestDeleteAlbumMissingUserID(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/album/7", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.ServerErrorResponse{Error: "userID not found in context"}, got)
}

func TestListAlbumImagesBadAlbumID(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/album/not-an-id/images", nil)
	req.Header.Set("uid", "4")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.BadRequestResponse{Error: "could not parse value 'not-an-id' into type database.AlbumID"}, got)
}

func testMissingUserID(t *testing.T, method, path string) {
	t.Helper()
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.ServerErrorResponse{Error: "userID not found in context"}, got)
}

func testUnparseableUserID(t *testing.T, method, path string) {
	t.Helper()
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("uid", "not-a-number")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.ServerErrorResponse{Error: "userID in context is not of type UserID"}, got)
}

func TestCreateAlbumMissingUserID(t *testing.T) {
	testMissingUserID(t, http.MethodPost, "/album")
	testMissingUserID(t, http.MethodGet, "/album")
	testMissingUserID(t, http.MethodPatch, "/album/1/name")
	testMissingUserID(t, http.MethodDelete, "/album/1")
	testMissingUserID(t, http.MethodGet, "/album/1/images")
	testMissingUserID(t, http.MethodPost, "/album/1/images")
}

func TestCreateAlbumUnparseableUserID(t *testing.T) {
	testUnparseableUserID(t, http.MethodPost, "/album")
	testUnparseableUserID(t, http.MethodGet, "/album")
	testUnparseableUserID(t, http.MethodPatch, "/album/1/name")
	testUnparseableUserID(t, http.MethodDelete, "/album/1")
	testUnparseableUserID(t, http.MethodGet, "/album/1/images")
	testUnparseableUserID(t, http.MethodPost, "/album/1/images")
}

func TestCreateAlbumFail(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	m.EXPECT().
		CreateAlbum("Vacation", database.UserID(9), database.ALBUM_PUBLIC).
		Return(database.AlbumID(0), assert.AnError)

	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal(AlbumCreateRequest{Name: "Vacation", VisibilityMode: database.ALBUM_PUBLIC})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "9")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.ServerErrorResponse{Error: assert.AnError.Error()}, got)
}

func TestAddImageToAlbumSuccess(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	m.EXPECT().
		SetImageAlbum(database.ImageID(10), database.AlbumID(5), database.UserID(3)).
		Return(nil)

	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal(AlbumImageRequest{ImageID: 10})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album/5/images", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "3")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got api.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.MessageResponse{Message: "Image added to album successfully"}, got)
}

func TestAddImageToAlbumBadBody(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal([]byte(`asdasdgfjhbiou`))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album/5/images", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "3")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}

func TestAddImageToAlbumBadJSONBody(t *testing.T) {
	m := mocks.NewMockAlbumDatabase(t)
	router := newAlbumTestRouter(InitAlbumEndpoint(m))
	body, err := json.Marshal(struct {InvalidField string}{
		InvalidField: "not a valid request body",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/album/5/images", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("uid", "3")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, api.BadRequestResponse{Error: "invalid request body"}, got)
}
