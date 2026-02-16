package jobs

import (
	"encoding/json"
	"freezetag/backend/api"
	mockJobService "freezetag/backend/mocks/JobService"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (je JobsEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.GET("/jobs/details/:id", je.Details)
	e.GET("/jobs/summary/:id", je.Summary)
	e.GET("/jobs/list", je.List)
	e.POST("/jobs/cancel/:id", je.Cancel)
}

func TestJobDetailsEndpointBadUUID(t *testing.T) {
	j := mockJobService.NewMockJobService(t)

	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/details/badid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobSummaryEndpointBadUUID(t *testing.T) {
	j := mockJobService.NewMockJobService(t)

	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/summary/badid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobDetailsEndpointNotFound(t *testing.T) {
	uuid := uuid.New()
	j := mockJobService.NewMockJobService(t)
	j.EXPECT().GetBatch(uuid).Return(nil)

	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/details/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobSummaryEndpointNotFound(t *testing.T) {
	uuid := uuid.New()
	j := mockJobService.NewMockJobService(t)
	j.EXPECT().GetSummary(uuid).Return(nil)

	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/summary/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobDetailsEndpointFound(t *testing.T) {
	uuid := uuid.New()
	job := repositories.JobBatch[repositories.JobInput, any]{
		UUID:       uuid,
		InProgress: nil,
		Completed:  nil,
		Failed:     nil,
	}

	j := mockJobService.NewMockJobService(t)
	j.EXPECT().GetBatch(uuid).Return(&job)

	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/details/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.FileJobBatch
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, got.UUID, uuid)
}

func TestJobSummaryEndpointFound(t *testing.T) {
	uuid := uuid.New()
	summary := services.JobSummary{
		UUID: uuid,
	}
	j := mockJobService.NewMockJobService(t)
	j.EXPECT().GetSummary(uuid).Return(&summary)
	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/summary/"+uuid.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got services.JobSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, got.UUID, uuid)
}

func TestJobListEndpoint(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	list := []services.JobSummary{
		{UUID: id1},
		{UUID: id2},
	}
	j := mockJobService.NewMockJobService(t)
	j.EXPECT().AllJobs().Return(list)
	router := gin.Default()
	InitJobsEndpoint(j).RegisterEndpoints(router)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/list", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []services.JobSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.ElementsMatch(t, list, got)
}
