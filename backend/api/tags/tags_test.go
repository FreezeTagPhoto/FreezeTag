package tags

import (
	"encoding/json"
	"errors"
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

func (te TagEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.DELETE("/tag/remove", te.HandleDelete)
	e.DELETE("/tag/delete", te.HandleDeleteFull)
	e.POST("/tag/add", te.HandlePost)
	e.GET("/tag/list", te.ListTags)
	e.GET("/tag/list/:id", te.ImageTags)
	e.GET("/tag/counts", te.ListCounts)
	e.GET("/tag/search", te.ListCountsQuery)
}

func TestGetAllTags(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveAllTags().Return(map[string]int64{"1": 1, "2": 1, "3": 1}, nil)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := map[string]int64{"1": 1, "2": 1, "3": 1}
	var got map[string]int64
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetAllTagsError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RetrieveAllTags().Return(map[string]int64{}, fmt.Errorf("mock error"))
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/list", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.ServerErrorResponse{Error: "mock error"}
	var got api.ServerErrorResponse
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
	expected := api.BadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.BadRequestResponse
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
	expected := api.ServerErrorResponse{Error: "mock error"}
	var got api.ServerErrorResponse
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
	expected := api.BadRequestResponse{Error: "Invalid image ID parameter"}
	var got api.BadRequestResponse
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
	expected := api.TagAddResponse{
		Added:  []repositories.ImageTagSuccess{*result.Success},
		Errors: []repositories.ImageTagFail{},
	}
	var got api.TagAddResponse
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
						Id:    id,
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

	expected := api.TagAddResponse{
		Added:  []repositories.ImageTagSuccess{{Id: 1, Count: 3}, {Id: 2, Count: 3}, {Id: 3, Count: 3}},
		Errors: []repositories.ImageTagFail{{Reason: "unknown id c"}},
	}
	var got api.TagAddResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	t.Log(w.Body.String())
	assert.ElementsMatch(t, expected.Added, got.Added)
	assert.ElementsMatch(t, expected.Errors, got.Errors)
}

func TestHandlePostNoIds(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().AddTags([]string{"tagtest"}).Return(repositories.TagResult{Success: true})
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "tagtest")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.TagAddResponse{Added: []repositories.ImageTagSuccess{{Count: 1}}, Errors: []repositories.ImageTagFail{}}
	var got api.TagAddResponse
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
	expected := api.BadRequestResponse{Error: "no tags to add"}
	var got api.BadRequestResponse
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
	params.Add("id", "124912481509128491718251935710248712971029285912729012754")
	reqURL := "/tag/add?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("POST", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.TagAddResponse{
		Added:  []repositories.ImageTagSuccess{},
		Errors: []repositories.ImageTagFail{{Reason: "unknown id a"}, {Reason: "unknown id 124912481509128491718251935710248712971029285912729012754"}},
	}
	var got api.TagAddResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

// tests below are the same as for post, as they have similar behaviors

func TestHandleDeleteSimple(t *testing.T) {
	result := repositories.ImageTagResult{
		Success: &repositories.ImageTagSuccess{
			Id:    database.ImageId(1),
			Count: 3,
		},
		Err: nil,
	}
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().RemoveImageTags(mock.Anything, mock.Anything).Return(result)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "1")
	params.Add("tag", "a")
	params.Add("tag", "3")
	params.Add("id", "1")
	reqURL := "/tag/remove?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.TagRemoveResponse{
		Deleted: []repositories.ImageTagSuccess{*result.Success},
		Errors:  []repositories.ImageTagFail{},
	}
	var got api.TagRemoveResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandleDeleteComplex(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		RemoveImageTags(mock.AnythingOfType("database.ImageId"), mock.Anything).
		RunAndReturn(
			func(id database.ImageId, _ []string) repositories.ImageTagResult {
				return repositories.ImageTagResult{
					Success: &repositories.ImageTagSuccess{
						Id:    id,
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
	reqURL := "/tag/remove?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	expected := api.TagRemoveResponse{
		Deleted: []repositories.ImageTagSuccess{{Id: 1, Count: 3}, {Id: 2, Count: 3}, {Id: 3, Count: 3}},
		Errors:  []repositories.ImageTagFail{{Reason: "unknown id c"}},
	}
	var got api.TagRemoveResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	t.Log(w.Body.String())
	assert.ElementsMatch(t, expected.Deleted, got.Deleted)
	assert.ElementsMatch(t, expected.Errors, got.Errors)
}

func TestHandleDeleteNoIds(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "tagtest")
	reqURL := "/tag/remove?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.BadRequestResponse{Error: "no ids to remove tags from"}
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestHandleDeleteNoTags(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("id", "1")
	reqURL := "/tag/remove?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.BadRequestResponse{Error: "no tags to remove"}
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)

}

func TestHandleDeleteBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)
	params := url.Values{}
	params.Add("tag", "1")
	params.Add("id", "a")
	params.Add("id", "124912481509128491718251935710248712971029285912729012754")
	reqURL := "/tag/remove?" + params.Encode() // properly encodes & joins parameters

	w := httptest.NewRecorder()
	//max signed int64 type
	req, _ := http.NewRequest("DELETE", reqURL, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := api.TagRemoveResponse{
		Deleted: []repositories.ImageTagSuccess{},
		Errors:  []repositories.ImageTagFail{{Reason: "unknown id a"}, {Reason: "unknown id 124912481509128491718251935710248712971029285912729012754"}},
	}
	var got api.TagRemoveResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCounts(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetTagCounts(mock.Anything).
		Return(map[string]int64{"1": 2, "2": 3}, nil)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/counts", nil)
	q := req.URL.Query()
	q.Add("id", "1")
	q.Add("id", "2")
	req.URL.RawQuery = q.Encode()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := map[string]int64{"1": 2, "2": 3}
	var got map[string]int64
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsBadId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetTagCounts(mock.Anything).
		Return(map[string]int64{"1": 2, "2": 3}, nil)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/counts", nil)
	q := req.URL.Query()
	q.Add("id", "1")
	q.Add("id", "3")
	req.URL.RawQuery = q.Encode()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	expected := map[string]int64{"1": 2, "2": 3}
	var got map[string]int64
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsNoIds(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/counts", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.BadRequestResponse{Error: "no ids specified"}
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsDatabaseError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetTagCounts(mock.Anything).
		Return(nil, errors.New("database error"))
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/counts", nil)
	q := req.URL.Query()
	q.Add("id", "1")
	q.Add("id", "3")
	req.URL.RawQuery = q.Encode()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.ServerErrorResponse{Error: "database error"}
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsInvalidId(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	router := gin.Default()
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/counts", nil)
	q := req.URL.Query()
	q.Add("id", "1")
	q.Add("id", "foo")
	req.URL.RawQuery = q.Encode()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.BadRequestResponse{Error: "bad id parameter"}
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsQuerySuccess(t *testing.T) {
    m := mocks.NewMockImageRepository(t)
    m.EXPECT().
        GetQueryTagCounts(mock.Anything, mock.Anything).
        Return(map[string]int64{"foo": 2}, nil)
        
    router := gin.Default()
    router.Use(func(c *gin.Context) {
        c.Set("userID", "1") 
        c.Next()
    })
    
    InitTagEndpoint(m).RegisterEndpoints(router)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/tag/search", nil)
    q := req.URL.Query()
    q.Add("make", "Apple")
    req.URL.RawQuery = q.Encode()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    expected := api.TagCounts(map[string]int64{"foo": 2})
    var got api.TagCounts
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
    assert.Equal(t, expected, got)
}

func TestGetTagCountsQueryError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		GetQueryTagCounts(mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("foo"))
	
	router := gin.Default()
	router.Use(func(c *gin.Context) { c.Set("userID", "1"); c.Next() })
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/search?make=Apple", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expected := api.ServerErrorResponse{Error: "foo"}
	var got api.ServerErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestGetTagCountsQueryInvalidField(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	
	router := gin.Default()
	router.Use(func(c *gin.Context) { c.Set("userID", "1"); c.Next() })
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tag/search?near=foo", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	expected := api.BadRequestResponse{Error: "invalid near parameter"}
	var got api.BadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expected, got)
}

func TestFullDeleteSuccess(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		DeleteTags([]string{"foo", "bar"}).
		Return(2, nil)
	
	router := gin.Default()
	router.Use(func(c *gin.Context) { c.Set("userID", "1"); c.Next() })
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tag/delete?tag=foo&tag=bar", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFullDeleteNoTags(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	
	router := gin.Default()
	router.Use(func(c *gin.Context) { c.Set("userID", "1"); c.Next() })
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tag/delete", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFullDeleteRepoError(t *testing.T) {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
		DeleteTags(mock.Anything).
		Return(0, fmt.Errorf("test error"))
	
	router := gin.Default()
	router.Use(func(c *gin.Context) { c.Set("userID", "1"); c.Next() })
	InitTagEndpoint(m).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tag/delete?tag=foo", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}