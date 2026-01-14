package jobquery

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobQueryEndpoint struct {
	jobRepository repositories.JobRepository
}

/* Functions */

// Creates a new JobQueryEndpoint with the given image repository.
func InitJobQueryEndpoint(jobRepository repositories.JobRepository) JobQueryEndpoint {
	return JobQueryEndpoint{
		jobRepository: jobRepository,
	}
}

// Registers the job query endpoints to the given Gin engine.
func (je JobQueryEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.GET("/jobquery/:id", je.HandleGet)
}

// @summary     Query job batch status
// @description Query the status of a job batch using its UUID
// @produce     application/json
// @router      /jobquery/{id} [get]
// @param       id path string true "the UUID of the job batch to query"
// @success     200 {object} repositories.JobBatch "the job batch with the given UUID"
// @failure     400 {object} api.StatusBadRequestResponse "if the provided UUID is not valid"
// @failure     404 {object} api.StatusNotFoundResponse "if no job batch with the given UUID is found"
func (je JobQueryEndpoint) HandleGet(c *gin.Context) {
	idParam := c.Param("id")
	var id uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to parse job batch UUID: " + err.Error()})
		return
	}

	jobBatch, err := je.jobRepository.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, api.StatusNotFoundResponse{Error: "job batch not found"})
		return
	}
	c.JSON(http.StatusOK, jobBatch)
}
