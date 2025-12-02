package metadata

import (
	"encoding/json"
	"fmt"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/images/imagedata"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string {
	return &s
}

func TestGetMetadataSuccessNils(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{}, nil)

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got imagedata.Metadata
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, imagedata.Metadata{}, got)
}

func TestGetMetadataSuccessOneValue(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{
			CameraMake:       ptrString("Canon"),
		}, nil)

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var expected = imagedata.Metadata{
		CameraMake: ptrString("Canon"),
	}
	var got imagedata.Metadata
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	t.Logf("Got metadata: %s", w.Body.String())
	assert.Equal(t, expected, got)
}

func TestGetMetadataError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{}, fmt.Errorf("mock error"))

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.StatusServerErrorResponse{Error: "mock error"}
	var got api.StatusServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetMetadataBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/abc", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}