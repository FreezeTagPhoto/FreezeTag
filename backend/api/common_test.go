package api

import (
	"bytes"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParamStringSuccess(t *testing.T) {
	id, err := ParseParamIntoID[database.UserID]("123")
	assert.NoError(t, err)
	assert.Equal(t, database.UserID(123), id)
}

func TestParseParamStringNonNumeric(t *testing.T) {
	id, err := ParseParamIntoID[database.UserID]("abc")
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestParseParamEmpty(t *testing.T) {
	id, err := ParseParamIntoID[database.TokenID]("")
	assert.Error(t, err)
	assert.Equal(t, database.TokenID(0), id)
}

func TestParseParamNonString(t *testing.T) {
	id, err := ParseParamIntoID[database.UserID](123)
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestParseParamStringInvalid(t *testing.T) {
	id, err := ParseParamIntoID[database.UserID]("one")
	assert.Error(t, err)
	assert.Equal(t, database.UserID(0), id)
}

func TestGetUserIDFromStringNegative(t *testing.T) {
	id, err := ParseParamIntoID[database.UserID]("-5")
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

func TestQueryPermissionsFromRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/permissions", nil)
		q := req.URL.Query()
		q.Add("permission", "read:user")
		q.Add("permission", "write:user")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		perms, err := QueryPermissionsFromRequest(ctx)
		require.NoError(t, err)
		assert.Len(t, perms, 2)
	})

	t.Run("failNoPermissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/permissions", nil)
		ctx.Request = req
		perms, err := QueryPermissionsFromRequest(ctx)
		assert.Error(t, err)
		assert.Nil(t, perms)
	})

	t.Run("failInvalidPermission", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/permissions", nil)
		q := req.URL.Query()
		q.Add("permission", "invalid")
		req.URL.RawQuery = q.Encode()
		ctx.Request = req
		perms, err := QueryPermissionsFromRequest(ctx)
		assert.Error(t, err)
		assert.Nil(t, perms)
	})
}

func writeTestFile(writer *multipart.Writer, fieldname string, filename string, content []byte) error {
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}

func TestReadFileBytes(t *testing.T) {
	router := gin.Default()
	router.POST("/testEndpoint", func(ctx *gin.Context) {
		form, err := ctx.FormFile("file")
		require.NoError(t, err)
		fileBytes, err := ReadFileBytes(form)
		require.NoError(t, err)
		assert.Equal(t, []byte("test content"), fileBytes)
	})

	w := httptest.NewRecorder()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writeTestFile(writer, "file", "testfile.txt", []byte("test content")))

	require.NoError(t, writer.Close())

	req, err := http.NewRequest("POST", "/testEndpoint", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	err = writer.Close()
	require.NoError(t, err)

}
