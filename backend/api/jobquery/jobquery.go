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
