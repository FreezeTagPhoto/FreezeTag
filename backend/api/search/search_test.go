package search

import (
	"encoding/json"
	"fmt"
	"freezetag/backend/api"
	mocks "freezetag/backend/mocks/ImageRepository"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (se SearchEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.GET("/search", se.Search)
}

func TestSearchSuccessNoQueries(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.AnythingOfType("*queries.ImageQuery"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]database.ImageId{1, 2, 3}, nil)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := []database.ImageId{1, 2, 3}
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestSearchSuccessBasicQueries(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.AnythingOfType("*queries.ImageQuery"), queries.DateCreated, queries.Ascending, mock.Anything, mock.Anything).
		Return([]database.ImageId{1}, nil)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	// Build query parameters
	params := url.Values{}
	params.Add("make", "test")
	params.Add("makeLike", "test2")
	params.Add("model", "testModel")
	params.Add("modelLike", "testModelLike")
	params.Add("takenBefore", "0")
	params.Add("takenAfter", "0")
	params.Add("uploadedBefore", "10")
	params.Add("uploadedAfter", "10")
	params.Add("sortBy", "DateCreated")
	params.Add("sortOrder", "ASC")

	reqURL := "/search?" + params.Encode() // properly encodes & joins parameters

	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := []database.ImageId{1}
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestSearchSuccessBasicQueries2(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.AnythingOfType("*queries.ImageQuery"), queries.DateAdded, queries.Descending, mock.Anything, mock.Anything).
		Return([]database.ImageId{1}, nil)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	// Build query parameters
	params := url.Values{}
	params.Add("make", "test")
	params.Add("makeLike", "test2")
	params.Add("model", "testModel")
	params.Add("modelLike", "testModelLike")
	params.Add("takenBefore", "0")
	params.Add("takenAfter", "0")
	params.Add("uploadedBefore", "10")
	params.Add("uploadedAfter", "10")
	params.Add("sortBy", "DateAdded")
	params.Add("sortOrder", "DESC")

	reqURL := "/search?" + params.Encode() // properly encodes & joins parameters

	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := []database.ImageId{1}
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestSearchSuccessTags(t *testing.T) {
	expectedQuery := queries.CreateImageQuery()
	expectedQuery.WithTag("1")
	expectedQuery.WithTag("2")
	expectedQuery.WithTag("3")
	expectedQuery.WithTagLike("4")

	params := url.Values{}
	params.Add("tag", "1")
	params.Add("tag", "2")
	params.Add("tag", "3")
	params.Add("tagLike", "4")
	reqURL := "/search?" + params.Encode() // properly encodes & joins parameters

	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.AnythingOfType("*queries.ImageQuery"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(
			func(actualQuery queries.DatabaseQuery, _ queries.SortField, _ queries.SortOrder, _ uint, _ uint) ([]database.ImageId, error) {
				assert.Equal(t, expectedQuery, actualQuery)
				return []database.ImageId{1, 2, 3, 4}, nil
			})

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := []database.ImageId{1, 2, 3, 4}
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func runNearErrorTest(t *testing.T, near string, expectedErr string) {
	t.Helper()

	params := url.Values{}
	params.Add("near", near)
	reqURL := "/search?" + params.Encode()

	m := mocks.NewMockImageRepository(t)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := api.BadRequestResponse{Error: expectedErr}
	var got api.BadRequestResponse

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestSearchNearBadNearTooManyCoords(t *testing.T) {
	runNearErrorTest(t, "1,2,3,4", "invalid near parameter")

}

func TestSearchNearBadLatitude(t *testing.T) {
	runNearErrorTest(t, "a,2,3", "invalid latitude in near parameter")
}

func TestSearchNearBadLongitude(t *testing.T) {
	runNearErrorTest(t, "1,b,3", "invalid longitude in near parameter")
}

func TestSearchNearBadDistance(t *testing.T) {
	runNearErrorTest(t, "1,2,c", "invalid distance in near parameter")
}

func TestSearchNearSuccess(t *testing.T) {
	expectedQuery := queries.CreateImageQuery()
	expectedQuery.WithLocation(1, 2, 3)
	params := url.Values{}
	params.Add("near", "1,2,3")
	reqURL := "/search?" + params.Encode()

	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(
			func(actualQuery queries.DatabaseQuery, _ queries.SortField, _ queries.SortOrder, _ uint, _ uint) ([]database.ImageId, error) {
				assert.Equal(t, expectedQuery, actualQuery)
				return []database.ImageId{1}, nil
			})

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	expected := []database.ImageId{1}
	var got []database.ImageId
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func runBadTakenXTest(t *testing.T, query, location, expectedErr string) {
	t.Helper()

	params := url.Values{}
	params.Add(query, location)
	reqURL := "/search?" + params.Encode()

	m := mocks.NewMockImageRepository(t)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := api.BadRequestResponse{Error: expectedErr}
	var got api.BadRequestResponse

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestBadTakenBefore(t *testing.T) {
	runBadTakenXTest(t, "takenBefore", "notint", "bad takenBefore parameter")
}
func TestBadTakenAfter(t *testing.T) {
	runBadTakenXTest(t, "takenAfter", "notint", "bad takenAfter parameter")
}
func TestBadUploadedBefore(t *testing.T) {
	runBadTakenXTest(t, "uploadedBefore", "notint", "bad uploadedBefore parameter")
}

func TestBadUploadedAfter(t *testing.T) {
	runBadTakenXTest(t, "uploadedAfter", "notint", "bad uploadedAfter parameter")
}

func TestSearchImageFail(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		SearchImageOrderedPaged(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("mock error"))

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expected := api.ServerErrorResponse{Error: "mock error"}
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)

}

func runBadSortTest(t *testing.T, param string, value string, expected string) {
	t.Helper()

	params := url.Values{}
	params.Add(param, value)
	reqURL := "/search?" + params.Encode()

	m := mocks.NewMockImageRepository(t)

	router := gin.Default()
	InitSearchEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	expectedErr := api.BadRequestResponse{Error: expected}
	var got api.BadRequestResponse

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expectedErr, got)
}

func TestSearchImageBadSort(t *testing.T) {
	runBadSortTest(t, "sortBy", "foo", "bad sortBy parameter")
	runBadSortTest(t, "sortOrder", "bar", "bad sortOrder parameter")
}

func TestSearchImageBadPage(t *testing.T) {
	runBadSortTest(t, "pageSize", "foo", "bad pageSize parameter")
	runBadSortTest(t, "pageNo", "-2", "bad pageNo parameter")
}
