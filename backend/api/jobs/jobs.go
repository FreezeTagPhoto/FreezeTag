package jobs

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobsEndpoint struct {
	jobService services.JobService
}

// Creates a new JobQueryEndpoint with the given image repository.
func InitJobsEndpoint(jobService services.JobService) JobsEndpoint {
	return JobsEndpoint{
		jobService: jobService,
	}
}

// Registers the job query endpoints to the given Gin engine.

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

// @summary     Query job batch details
// @description Retrieves the current details of a job batch, including specifics of individual in-progress/completed/failed jobs
// @produce     application/json
// @router      /jobs/details/{id} [get]
// @tags        jobs
// @param       id   path      string  true  "Job Batch UUID" format(uuid)
// @success     200  {object}  exampleResponse
// @failure     400  {object}  api.BadRequestResponse
// @failure     404  {object}  api.NotFoundResponse
func (je JobsEndpoint) Details(c *gin.Context) {
	idParam := c.Param("id")
	var id uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse job batch UUID: " + err.Error()})
		return
	}

	jobBatch := je.jobService.GetBatch(id)
	if jobBatch == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: "job batch not found"})
		return
	}
	c.JSON(http.StatusOK, jobBatch)
}

// @summary     Query job batch summary
// @description Retrieves a summary of a job batch, with a high-level overview of progress
// @produce     application/json
// @router      /jobs/summary/{id} [get]
// @tags        jobs
// @param       id   path      string  true  "Job Batch UUID" format(uuid)
// @success     200  {object}  services.JobSummary
// @failure     400  {object}  api.BadRequestResponse
// @failure     404  {object}  api.NotFoundResponse
func (je JobsEndpoint) Summary(c *gin.Context) {
	idParam := c.Param("id")
	var id uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse job batch UUID: " + err.Error()})
		return
	}

	summary := je.jobService.GetSummary(id)
	if summary == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: "job batch not found"})
		return
	}
	c.JSON(http.StatusOK, *summary)
}

// @summary     List all job batches
// @description Lists all job batches, including currently running, scheduled, and cancelled/finished (before deletion)
// @produce     application/json
// @router      /jobs/list [get]
// @tags        jobs
// @success     200 {array} services.JobSummary
func (je JobsEndpoint) List(c *gin.Context) {
	jobs := je.jobService.AllJobs()
	c.JSON(http.StatusOK, jobs)
}

// @summary     Cancel a job
// @description Cancel a job given a job ID
// @produce     application/json
// @router      /jobs/cancel/{id} [post]
// @tags        jobs
// @param       id   path      string  true  "Job Batch UUID" format(uuid)
// @success     200 {object} api.CancelledJobResponse
// @failure     404 {object} api.NotFoundResponse
func (je JobsEndpoint) Cancel(c *gin.Context) {
	idParam := c.Param("id")
	var id uuid.UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to parse job batch UUID: " + err.Error()})
		return
	}

	batch := je.jobService.GetBatch(id)
	if batch == nil {
		c.JSON(http.StatusNotFound, api.NotFoundResponse{Error: "job batch not found"})
		return
	}
	batch.Cancel()
	c.JSON(http.StatusOK, api.CancelledJobResponse{UUID: batch.UUID})
}
