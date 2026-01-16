package services

import (
	"fmt"
	"log"

	"freezetag/backend/pkg/repositories"

	"golang.org/x/sync/errgroup"
)

const (
	ThreadLimit = 10
)

type JobService interface {
	RunUploadJobs(batch *repositories.JobBatch) error
	CreateJobBatch(jobs []*repositories.FileJob) (*repositories.JobBatch, error)
}

type defaultJobService struct {
	jobRepository   repositories.JobRepository
	imageRepository repositories.ImageRepository
}

func InitDefaultJobService(jobRepository repositories.JobRepository, imageRepository repositories.ImageRepository) JobService {
	return &defaultJobService{
		jobRepository:   jobRepository,
		imageRepository: imageRepository,
	}
}

func (s *defaultJobService) RunUploadJobs(batch *repositories.JobBatch) error {
	// if the batch context gets canceled, the rest of the jobs will stop processing
	g, ctx := errgroup.WithContext(batch.Ctx)
	g.SetLimit(ThreadLimit)

	jobs := make([]*repositories.FileJob, len(batch.InProgress))
	copy(jobs, batch.InProgress)

	batch.WG.Add(1)
	go func() {
		defer batch.WG.Done()
		for _, file := range jobs {
			f := file
			if ctx.Err() != nil {
				log.Printf("Job batch %v canceled, stopping remaining jobs", batch.UUID)
				break
			}
			g.Go(func() error {
				if ctx.Err() != nil {
					log.Printf("Job batch %v canceled, skipping job for file %s", batch.UUID, file.Name)
					return ctx.Err()
				}
				result := s.imageRepository.StoreImageBytes(f.Bytes, f.Name)
				if result.Err != nil {
					_ = s.jobRepository.UpdateJobStatus(batch.UUID, f.Name, "failure")
					return fmt.Errorf("%s", result.Err.Reason)
				}
				if err := s.jobRepository.CompleteFileJob(batch.UUID, f.Name, result); err != nil {
					return err
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			log.Printf("Error running upload jobs: %v", err)
		}

		// wait for the last group of jobs to be done before releasing the batch waitgroup
		_ = g.Wait()
	}()
	return nil
}

func (s *defaultJobService) CreateJobBatch(jobs []*repositories.FileJob) (*repositories.JobBatch, error) {
	UUID := repositories.NewJobBatchID()
	jobBatch := repositories.JobBatch{
		UUID:       UUID,
		InProgress: jobs,
	}
	err := s.jobRepository.Create(&jobBatch)
	if err != nil {
		return nil, err
	}
	return &jobBatch, nil
}
