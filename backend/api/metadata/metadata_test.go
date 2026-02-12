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
	m.EXPECT().
		GetImageResolution(mock.AnythingOfType("database.ImageId")).
		Return(0, 0, nil)

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got api.MetadataResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, imagedata.Metadata{}, got.Metadata)
}

func TestGetMetadataSuccessOneValue(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{
			CameraMake: ptrString("Canon"),
		}, nil)
	m.EXPECT().
		GetImageResolution(mock.AnythingOfType("database.ImageId")).
		Return(69, 420, nil)

	router := gin.Default()
	InitMetadataEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metadata/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var expected = api.MetadataResponse{
		Metadata: imagedata.Metadata{CameraMake: ptrString("Canon")},
		Width:    69,
		Height:   420,
	}
	var got api.MetadataResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	t.Logf("Got metadata: %s", w.Body.String())
	assert.Equal(t, expected, got)
}

func TestGetMetadataError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{}, fmt.Errorf("mock error")).Maybe()
	m.EXPECT().
		GetImageResolution(mock.AnythingOfType("database.ImageId")).
		Return(10, 12, nil).Maybe()

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

func TestGetMetadataError2(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetImageMetadata(mock.AnythingOfType("database.ImageId")).
		Return(imagedata.Metadata{}, nil).Maybe()
	m.EXPECT().
		GetImageResolution(mock.AnythingOfType("database.ImageId")).
		Return(0, 0, fmt.Errorf("mock error")).Maybe()

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
