package files

import (
	"encoding/json"
	"fmt"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/ImageRepository"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (fe FileEndpoint) RegisterEndpoints(e gin.IRouter) {
	e.GET("/file/:id", fe.HandleGet)
}

func TestServeFileSuccess(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageFilepath(mock.Anything).
		Return("./test_resources/gopher1.png", nil)

	router := gin.Default()
	InitFileEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected, err := os.ReadFile("./test_resources/gopher1.png")
	require.NoError(t, err)
	assert.Equal(t, expected, w.Body.Bytes())
}

func TestServeFileFail(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageFilepath(mock.Anything).
		Return("", fmt.Errorf("mock error"))

	router := gin.Default()
	InitFileEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expected := api.StatusServerErrorResponse{Error: "mock error"}
	var got api.StatusServerErrorResponse
	t.Log(w.Body.String())
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestServeFileBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitFileEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file/a", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	t.Log(w.Body.String())
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestServeFileNotFound(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageFilepath(mock.Anything).
		Return("./test_resources/notreal.png", nil)

	router := gin.Default()
	InitFileEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/file/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
