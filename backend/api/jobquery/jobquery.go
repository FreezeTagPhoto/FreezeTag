package jobquery

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobQueryEndpoint struct {
	jobRepository repositories.JobRepository
}

// Creates a new JobQueryEndpoint with the given image repository.
func InitJobQueryEndpoint(jobRepository repositories.JobRepository) JobQueryEndpoint {
	return JobQueryEndpoint{
		jobRepository: jobRepository,
	}
}

// Registers the job query endpoints to the given Gin engine.
func (je JobQueryEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.GET("/jobquery/:id", je.HandleGet)
}

type exampleResponse struct { //nolint:unused
	UUID      uuid.UUID                         `json:"uuid"`
	Completed []repositories.ImageUploadSuccess `json:"completed"`
	Failed    []struct {
		Input  services.FileJob `json:"input"`
		Reason string           `json:"reason"`
	} `json:"failed"`
	InProgress []services.FileJob `json:"in_progress"`
	Cancelled  bool               `json:"cancelled"`
}

// @summary     Query job batch status
// @description Retrieves the current status of a job batch, including pending files and completed results.
// @produce     application/json
// @router      /jobquery/{id} [get]
// @tags        jobs
// @param       id   path      string  true  "Job Batch UUID" format(uuid)
// @success     200  {object}  exampleResponse
// @failure     400  {object}  api.StatusBadRequestResponse
// @failure     404  {object}  api.StatusNotFoundResponse
func (je JobQueryEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	var id uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to parse job batch UUID: " + err.Error()})
		return
	}

	jobBatch := je.jobRepository.Get(id)
	if jobBatch == nil {
		c.JSON(http.StatusNotFound, api.StatusNotFoundResponse{Error: "job batch not found"})
		return
	}
	c.JSON(http.StatusOK, jobBatch)
}
