package services

import (
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

				// if a single job fails, it doesnt necessarily mean
				// all the other jobs should be canceled, so we log the error but return nil to let other jobs keep running
				id, err := s.imageRepository.StoreImageBytes(f.Bytes, f.Name)
				if err != nil {
					log.Printf("Failed to store image bytes for file %s in batch %v: %v", f.Name, batch.UUID, err)
					if err := s.jobRepository.FailFileJob(batch.UUID, f.Name, err); err != nil {
						return err
					}
				} else {
					if err := s.jobRepository.CompleteFileJob(batch.UUID, f.Name, id); err != nil {
						return err
					}
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
