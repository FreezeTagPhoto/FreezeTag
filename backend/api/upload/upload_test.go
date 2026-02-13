package upload

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/api"
	mockJobService "freezetag/backend/mocks/JobService"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T) *gin.Engine {
	t.Helper()
	j := mockJobService.NewMockJobService(t)
	j.EXPECT().RunUploadJob(mock.Anything).Return(uuid.New()).Maybe()

	router := gin.Default()
	InitUploadEndpoint(j).RegisterEndpoints(router)
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
	assert.Equal(t, http.StatusAccepted, w.Code)
	var got uuid.UUID
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEqual(t, uuid.Nil, got)
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

	expected := api.StatusBadRequestResponse{Error: "multipart form has no file field or no files were uploaded"}
	var got api.StatusBadRequestResponse
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

	expected := api.StatusBadRequestResponse{Error: "failed to parse multipart form: no multipart boundary param in Content-Type"}
	var got api.StatusBadRequestResponse
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

	expected := api.StatusBadRequestResponse{Error: "failed to parse multipart form: request Content-Type isn't multipart/form-data"}
	var got api.StatusBadRequestResponse
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
	assert.Equal(t, http.StatusAccepted, w.Code)
	var got uuid.UUID
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.NotEqual(t, uuid.Nil, got)
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

	expected := api.StatusBadRequestResponse{Error: "multipart form has no file field or no files were uploaded"}
	var got api.StatusBadRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

	assert.Equal(t, expected, got)
}
