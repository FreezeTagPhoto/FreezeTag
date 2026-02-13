package jobquery

import (
	"encoding/json"
	"freezetag/backend/api"
	mockJobRepo "freezetag/backend/mocks/JobRepository"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobQueryEndpointBadUUID(t *testing.T) {
	j := mockJobRepo.NewMockJobRepository(t)

	router := gin.Default()
	InitJobQueryEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobquery/badid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobQueryEndpointExpiredUUID(t *testing.T) {
	uuid := uuid.New()
	j := mockJobRepo.NewMockJobRepository(t)
	j.EXPECT().Get(uuid).Return(nil)

	router := gin.Default()
	InitJobQueryEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobquery/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobQueryEndpointCorrectUUID(t *testing.T) {

	uuid := uuid.New()
	job := repositories.JobBatch[repositories.JobInput, any]{
		UUID:       uuid,
		InProgress: nil,
		Completed:  nil,
		Failed:     nil,
	}

	j := mockJobRepo.NewMockJobRepository(t)
	j.EXPECT().Get(uuid).Return(&job)

	router := gin.Default()
	InitJobQueryEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobquery/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.FileJobBatch
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, got.UUID, uuid)
}
