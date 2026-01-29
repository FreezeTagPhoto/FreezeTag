package services

import (
	"context"
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
	plugins         PluginService
}

func InitDefaultJobService(jobRepository repositories.JobRepository, imageRepository repositories.ImageRepository, plugins PluginService) JobService {
	return &defaultJobService{
		jobRepository:   jobRepository,
		imageRepository: imageRepository,
		plugins:         plugins,
	}
}

func (s *defaultJobService) RunUploadJobs(batch *repositories.JobBatch) error {
	// if the batch context gets canceled, the rest of the jobs will stop processing
	g, ctx := errgroup.WithContext(batch.Ctx)
	g.SetLimit(ThreadLimit)
	if batch.Finished == nil {
		batch.Finished = make(chan struct{}, 1)
	}

	jobs := make([]*repositories.FileJob, len(batch.InProgress))
	copy(jobs, batch.InProgress)
	if len(jobs) == 0 {
		batch.MarkFinished()
	}

	batch.WG.Add(1)
	go func() {
		defer batch.WG.Done()
		for _, file := range jobs {
			f := file
			if ctx.Err() != nil {
				log.Printf("[INFO] Job batch %v canceled, stopping remaining jobs", batch.UUID)
				break
			}
			g.Go(func() error {
				if ctx.Err() != nil {
					log.Printf("[INFO] Job batch %v canceled, skipping job for file %s", batch.UUID, file.Name)
					return ctx.Err()
				}

				// if a single job fails, it doesnt necessarily mean
				// all the other jobs should be canceled, so we log the error but return nil to let other jobs keep running
				id, err := s.imageRepository.StoreImageBytes(f.Bytes, f.Name)
				if err != nil {
					log.Printf("[ERR] Failed to store image bytes for file %s in batch %v: %v", f.Name, batch.UUID, err)
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
			log.Printf("[ERR]  Error running upload jobs: %v", err)
		}

		// wait for the last group of jobs to be done before releasing the batch waitgroup
		_ = g.Wait()
	}()
	// set up PostUpload hooks to run
	go func() {
		uploads := <-s.jobRepository.WaitFinished(batch.UUID)
		if uploads == nil {
			return // cancelled job
		}
		// TODO: change this to something reportable job-style so that plugin jobs can be cancelled
		ctx := context.Background()
		log.Printf("[INFO] running plugins on files from batch %v", batch.UUID)
		for _, plugin := range s.plugins.AllPlugins() {
			results, err := s.plugins.RunPostUpload(plugin, ctx, uploads)
			if err != nil {
				log.Printf("[ERR]  %s: %s", plugin, err.Error())
				continue
			}
			for {
				err, ok := <-results
				if !ok {
					break // finished this plugin, move to the next
				}
				if err != nil {
					log.Printf("[ERR]  %s: %s", plugin, err.Error())
				}
			}
		}
		log.Printf("[INFO] finished running plugins on files from batch %v", batch.UUID)
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
