package tags

import (
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


func TestGetAllTags(t *testing.T) { 
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveAllTags().Return([]string{"1", "2", "3"}, nil)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := []string{"1", "2", "3"}
	var got []string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.ElementsMatch(t, expected, got)
}

func TestGetAllTagsError(t *testing.T) { 
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveAllTags().Return([]string{}, fmt.Errorf("mock error"))
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.StatusServerErrorResponse{Error: "mock error"}
	var got api.StatusServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandleGetImageTagsSuccess(t *testing.T) { 
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		RetrieveImageTags(mock.Anything).
		Return([]string{"1", "2", "3"}, nil)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := []string{"1", "2", "3"}
	var got []string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.ElementsMatch(t, expected, got)
}

func TestHandleGetImageTagsBadId(t *testing.T) { 
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list/a", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandleGetImageTagsBadDatabaseRequest(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		RetrieveImageTags(mock.Anything).
		Return([]string{}, fmt.Errorf("mock error"))
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list/1", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.StatusServerErrorResponse{Error: "mock error"}
	var got api.StatusServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandleGetImageTagsIntOverflow(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("GET","/tag/list/9223372036854775808" , nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}




func TestRemoveImageTags(t *testing.T) { 
	// result := repositories.ImageTagResult{
	// 	Success: &repositories.ImageTagSuccess{
	// 		Id: database.ImageId(67),
	// 		Count: 3,
	// 	},
	// 	Err: nil,
	// }
}