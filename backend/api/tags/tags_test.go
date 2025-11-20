package tags

import (
	"encoding/json"
	"fmt"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	req, _ := http.NewRequest("GET", "/tag/list/9223372036854775808", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandlePostSimple(t *testing.T) {
	result := repositories.ImageTagResult{
		Success: &repositories.ImageTagSuccess{
			Id:    database.ImageId(1),
			Count: 3,
		},
		Err: nil,
	}
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().AddImageTags(mock.Anything, mock.Anything).Return(result)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "1")
	params.Add("tag", "a")
	params.Add("tag", "3")
	params.Add("id", "1")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.StatusOkTagAddResponse{
		Added:  []repositories.ImageTagSuccess{*result.Success},
		Errors: []repositories.ImageTagFail{},
	}
	var got api.StatusOkTagAddResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}


func TestHandlePostComplex(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		AddImageTags(mock.AnythingOfType("database.ImageId"), mock.Anything).
		RunAndReturn(
			func(id database.ImageId, _ []string) repositories.ImageTagResult { 
				return repositories.ImageTagResult{
					Success: &repositories.ImageTagSuccess{
						Id:    id ,
						Count: 3,
					},
					Err: nil,
				}
			})
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "1")
	params.Add("tag", "2")
	params.Add("tag", "3")
	params.Add("id", "1")
	params.Add("id", "2")
	params.Add("id", "3")
	params.Add("id", "c")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := api.StatusOkTagAddResponse{
		Added:  []repositories.ImageTagSuccess{{Id: 1, Count: 3}, {Id: 2, Count: 3}, {Id: 3, Count: 3}},
		Errors: []repositories.ImageTagFail{{Reason: "unknown id c", Id: -1}},
	}
	var got api.StatusOkTagAddResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	t.Log(w.Body.String())
	assert.ElementsMatch(t, expected.Added, got.Added)
	assert.ElementsMatch(t, expected.Errors, got.Errors)
}

func TestHandlePostNoIds(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "tagtest")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "no ids to remove tags from"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandlePostNoTags(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("id", "1")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.StatusBadRequestResponse{Error: "no tags to remove"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)

}

func TestHandlePostBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "1")
	params.Add("id", "a")
	reqURL := "/tag/add?" + params.Encode() // pro,perly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.StatusOkTagAddResponse{
		Added:  []repositories.ImageTagSuccess{},
		Errors: []repositories.ImageTagFail{{Reason: "unknown id a", Id: -1}},
	}
	var got api.StatusOkTagAddResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}
