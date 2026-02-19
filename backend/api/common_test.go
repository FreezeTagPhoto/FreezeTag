package api

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserIdFromStringSuccess(t *testing.T) {
	id, err := GetUserIDFromString("123")
	assert.NoError(t, err)
	assert.Equal(t, database.UserID(123), id)
}

func TestGetUserIdFromStringInvalid(t *testing.T) {
	id, err := GetUserIDFromString("one")
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestGetUserIdFromStringNegative(t *testing.T) {
	id, err := GetUserIDFromString("-5")
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestExtractDatabaseQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		q := req.URL.Query()
		q.Add("make", "Apple")
		q.Add("model", "iPhone 16 Plus")
		q.Add("makeLike", "app")
		q.Add("modelLike", "iphone")
		q.Add("tag", "foo")
		q.Add("tagLike", "bar")
		q.Add("near", "100,100,5")
		q.Add("takenBefore", "50")
		q.Add("takenAfter", "50")
		q.Add("uploadedBefore", "50")
		q.Add("uploadedAfter", "50")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		require.NotNil(t, dbq)
		expected := queries.CreateImageQuery().
			WithLocation(100, 100, 5).
			WithMake("Apple").
			WithModel("iPhone 16 Plus").
			WithMakeLike("app").
			WithModelLike("iphone").
			WithTag("foo").
			WithTagLike("bar").
			TakenBefore(time.Unix(50, 0)).
			TakenAfter(time.Unix(50, 0)).
			UploadedBefore(time.Unix(50, 0)).
			UploadedAfter(time.Unix(50, 0))
		assert.Equal(t, *expected, *dbq)
	})

	t.Run("successNothing", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		require.NotNil(t, dbq)
		assert.Equal(t, *queries.CreateImageQuery(), *dbq)
	})

	for _, param := range []string{"takenBefore", "takenAfter", "uploadedBefore", "uploadedAfter"} {
		t.Run("failParse"+param, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/search", nil)
			q := req.URL.Query()
			q.Add(param, "foo")
			req.URL.RawQuery = q.Encode()
			ctx.Request = req
			dbq := GetRequestQuery(ctx)
			assert.Nil(t, dbq)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}

	t.Run("failParseNear1", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		q := req.URL.Query()
		q.Add("near", "123")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		assert.Nil(t, dbq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("failParseNear2", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		q := req.URL.Query()
		q.Add("near", "foo,2,3")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		assert.Nil(t, dbq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("failParseNear3", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		q := req.URL.Query()
		q.Add("near", "1,foo,3")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		assert.Nil(t, dbq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("failParseNear4", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/search", nil)
		q := req.URL.Query()
		q.Add("near", "1,2,foo")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		dbq := GetRequestQuery(ctx)
		assert.Nil(t, dbq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
