package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"freezetag/backend/pkg/repositories"

	"github.com/google/uuid"
)

// only one plugin can run at a time
// (TODO: change this eventually to be plugin-specific)
var pluginLock sync.Mutex

type FileJob struct {
	Name  string `json:"name"`
	Bytes []byte `json:"-"`
}

type JobSummary struct {
	UUID       uuid.UUID `json:"uuid"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	InProgress int       `json:"in_progress"`
	Complete   int       `json:"complete"`
	Errors     int       `json:"errors"`
}

// required to make FileJobs satisfy the JobInput interface
// without requiring someone else to assign them IDs
type innerFileJob struct {
	FileJob
	id int `json:"-"`
}

func (j innerFileJob) ID() int {
	return j.id
}

type JobService interface {
	GetBatch(uuid.UUID) *repositories.JobBatch[repositories.JobInput, any]
	GetSummary(uuid.UUID) *JobSummary
	AllJobs() []JobSummary
	RunUploadJob(files []FileJob) uuid.UUID
	SchedulePostUploads(upload uuid.UUID)
}

type defaultJobService struct {
	jobRepository   repositories.JobRepository
	imageRepository repositories.ImageRepository
	plugins         PluginService
}

func InitDefaultJobService(jobRepository repositories.JobRepository, imageRepository repositories.ImageRepository, plugins PluginService) JobService {
	return &defaultJobService{
		jobRepository:   jobRepository,
		imageRepository: imageRepository,
		plugins:         plugins,
	}
}

func (s *defaultJobService) GetBatch(id uuid.UUID) *repositories.JobBatch[repositories.JobInput, any] {
	return s.jobRepository.Get(id)
}

func (s *defaultJobService) GetSummary(id uuid.UUID) *JobSummary {
	job := s.jobRepository.Get(id)
	if job == nil {
		return nil
	}
	job.Lock.Lock()
	defer job.Lock.Unlock()
	return &JobSummary{
		UUID:       job.UUID,
		Title:      job.Title,
		Status:     job.Status,
		InProgress: len(job.InProgress),
		Complete:   len(job.Completed),
		Errors:     len(job.Failed),
	}
}

func (s *defaultJobService) AllJobs() []JobSummary {
	jobs := s.jobRepository.AllJobs()
	summaries := make([]JobSummary, 0, len(jobs))
	for _, job := range jobs {
		job.Lock.Lock()
		summaries = append(summaries, JobSummary{
			UUID:       job.UUID,
			Title:      job.Title,
			Status:     job.Status,
			InProgress: len(job.InProgress),
			Complete:   len(job.Completed),
			Errors:     len(job.Failed),
		})
		job.Lock.Unlock()
	}
	return summaries
}

func (s *defaultJobService) uploadOneFile(f innerFileJob) (repositories.ImageUploadSuccess, error) {
	id, err := s.imageRepository.StoreImageBytes(f.Bytes, f.Name)
	if err != nil {
		return repositories.ImageUploadSuccess{}, err
	}
	return repositories.ImageUploadSuccess{Id: id, Filename: f.Name}, nil
}

func (s *defaultJobService) RunUploadJob(batch []FileJob) uuid.UUID {
	jobs := make([]repositories.JobInput, len(batch))
	for i, job := range batch {
		jobs[i] = innerFileJob{job, i}
	}
	id := s.jobRepository.Create(fmt.Sprintf("Uploading %d files", len(batch)), context.Background(), jobs, repositories.SimpleJob(s.uploadOneFile))
	s.SchedulePostUploads(id)
	return id
}

type pluginRun struct {
	Name string `json:"name"`
	Hook string `json:"hook"`
	id   int    `json:"-"`
}
type pluginResult map[string]any

func (p pluginRun) ID() int {
	return p.id
}

func (s *defaultJobService) SchedulePostUploads(upload uuid.UUID) {
	batch := s.GetBatch(upload)
	var jobs []repositories.JobInput
	// create a giant list of plugin hooks to run synchronously in order
	for i, plugin := range s.plugins.AllPlugins() {
		jobs = append(jobs, pluginRun{plugin, "PostUpload", i})
	}
	s.jobRepository.Create("PostUpload plugin hooks for upload "+batch.UUID.String(), batch.Context, jobs, repositories.Job(func(p pluginRun, c context.Context, status func(string)) (pluginResult, error) {
		<-batch.WaitFinished()
		result := make(map[string]any)
		if batch.Cancelled {
			return result, fmt.Errorf("cancelled")
		}
		uploads := make([]repositories.ImageUploadSuccess, len(batch.Completed))
		for i, succ := range batch.Completed {
			uploads[i] = succ.(repositories.ImageUploadSuccess)
		}
		// only one plugin can run at once
		pluginLock.Lock()
		defer pluginLock.Unlock()
		log.Printf("[INFO] running plugin '%s' on upload batch %v", p.Name, upload)
		status(fmt.Sprintf("Running plugin %s", p.Name))
		results, err := s.plugins.RunPostUpload(p.Name, c, uploads)
		if err != nil {
			return result, err
		}
		for {
			err, ok := <-results
			if !ok {
				break
			}
			if err != nil {
				log.Printf("[ERR]  %s: %s", p.Name, err.Error())
			}
		}
		result["name"] = p.Name
		result["status"] = "success"
		log.Printf("[INFO] finished plugin '%s' on upload batch %v", p.Name, upload)
		return result, nil
	}))
}
