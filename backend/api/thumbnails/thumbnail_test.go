package thumbnails

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/ImageRepository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:embed test_resources/gopher.webp
var gopherBytes []byte //embed thes


func (te ThumbnailEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.GET("/thumbnails/:id", te.HandleGet)
}


func TestGetThumbnailSuccess(t *testing.T) {

	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		RetrieveThumbnail(mock.Anything, mock.Anything).
		Return(gopherBytes, nil)
	router := gin.Default()
	InitThumbnailEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/thumbnails/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, gopherBytes, w.Body.Bytes())
}

func TestGetThumbnailBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitThumbnailEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/thumbnails/bad123data", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestGetThumbnailSpecifySize(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveThumbnail(mock.Anything, mock.Anything).Return(gopherBytes, nil)
	router := gin.Default()
	InitThumbnailEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/thumbnails/1?size=67", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, gopherBytes, w.Body.Bytes())
}

func TestGetThumbnailDatabaseFail(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveThumbnail(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("mock error"))
	router := gin.Default()
	InitThumbnailEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/thumbnails/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expected := api.StatusServerErrorResponse{Error: "mock error"}
	var got api.StatusServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestGetThumbnailDatabaseNilReturn(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveThumbnail(mock.Anything, mock.Anything).Return(nil, nil)
	router := gin.Default()
	InitThumbnailEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/thumbnails/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	expected := api.StatusNotFoundResponse{Error: "thumbnail not found"}
	var got api.StatusNotFoundResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}
