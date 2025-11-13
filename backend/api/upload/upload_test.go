package upload

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/mockery"
	"freezetag/backend/pkg/repositories"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T) *gin.Engine {
	m := mocks.NewMockImageRepository(t)
	m.EXPECT().
    StoreImageBytes(mock.Anything, mock.AnythingOfType("string")).
    RunAndReturn(func(_ []byte, filename string) repositories.Result {
        return repositories.Result{
            Success: &repositories.ImageHandleSuccess{
                Id:       67,
                Filename: filename,
            },
        }
    }).Maybe()

	router := gin.Default()
	InitUploadEndpoint(m).RegisterEndpoints(router)
	return router
}

func writeTestFile(writer *multipart.Writer, fieldname string, filename string, content []byte) error {
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}

func TestPostFileSuccess(t *testing.T) {
	router := initTest(t)

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	require.NoError(t, writeTestFile(writer, "file", "testfile.png", []byte("hello world image")))
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	expected := StatusOkResponse{
		Uploaded: []repositories.ImageHandleSuccess{{Id: 67, Filename: "testfile.png"}},
		Errors:   []repositories.ImageHandleFail{},
	}
	var got StatusOkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestPostWithNoFiles(t *testing.T) {
	router := initTest(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := StatusBadRequestResponse{Error: "multipart form has no file field or no files were uploaded"}
	var got StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestPostWithMalformedMultipartForm(t *testing.T) {
	router := initTest(t)

	body := bytes.NewBufferString("this is not a multipart payload")
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data") // missing boundary

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := StatusBadRequestResponse{Error: "failed to parse multipart form: no multipart boundary param in Content-Type"}
	var got StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestPostTextNoMultipartForm(t *testing.T) {
	router := initTest(t)

	body := bytes.NewBufferString("this is not a multipart payload")
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", "text/html") // no media type

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := StatusBadRequestResponse{Error: "failed to parse multipart form: request Content-Type isn't multipart/form-data"}
	var got StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}

func TestPostWithMultipleFilesSuccess(t *testing.T) {
	router := initTest(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	require.NoError(t, writeTestFile(writer, "file", "testfile1.png", []byte("hello world image 1")))
	require.NoError(t, writeTestFile(writer, "file", "testfile2.jpg", []byte("hello world image 2")))
	require.NoError(t, writeTestFile(writer, "file", "testfile3.txt", []byte("hello world text")))

	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	expected := StatusOkResponse{
		Uploaded: []repositories.ImageHandleSuccess{
			{Id: 67, Filename: "testfile1.png"},
			{Id: 67, Filename: "testfile2.jpg"},
			{Id: 67, Filename: "testfile3.txt"},
		},
		Errors: []repositories.ImageHandleFail{},
	}
	var got StatusOkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	//because the order of uploaded files is not guaranteed, sort the same way before comparing.
	sort.Slice(expected.Uploaded, func(i, j int) bool {
		return expected.Uploaded[i].Filename < expected.Uploaded[j].Filename
	})
	sort.Slice(got.Uploaded, func(i, j int) bool {
		return got.Uploaded[i].Filename < got.Uploaded[j].Filename
	})
	assert.Equal(t, expected, got)
}

func TestPostWithNoFileField(t *testing.T) {
	router := initTest(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	require.NoError(t, writeTestFile(writer, "not_file", "testfile1.png", []byte("hello world image 1")))
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	expected := StatusBadRequestResponse{Error: "multipart form has no file field or no files were uploaded"}
	var got StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}
