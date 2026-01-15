package services

import (
	"freezetag/backend/pkg/repositories"
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

	jobs := make([]*repositories.FileJob, len(batch.InProgress))
	copy(jobs, batch.InProgress)
	for _, file := range jobs {
		go func(name string, data []byte) {
			result := s.imageRepository.StoreImageBytes(batch.Ctx, data, name)
			_ = s.jobRepository.CompleteFileJob(batch.UUID, name, result)
		}(file.Name, file.Bytes)
	}
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

