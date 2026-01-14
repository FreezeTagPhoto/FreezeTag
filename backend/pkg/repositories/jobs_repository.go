package repositories

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type JobBatch struct {
	Uuid       uuid.UUID `json:"uuid"`
	Completed  []*FileJob `json:"completed"`
	InProgress []*FileJob `json:"in_progress"`
}

type FileJob struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	bytes  []byte `json:"-"`
}

type JobRepository interface {
    Create(JobBatch) error
    Get(uuid.UUID) (*JobBatch, error)
    Delete(uuid.UUID) error

    AddInProgress(batchID uuid.UUID, file *FileJob) error
    SetFileStatus(batchID uuid.UUID, fileName string, status string) error
}

type DefaultJobRepository struct {
	jobs sync.Map // map[uuid.UUID]*JobBatch
}

func (r *DefaultJobRepository) Create(batch JobBatch) error {
	r.jobs.Store(batch.Uuid, &batch)
	return nil
}

func (r *DefaultJobRepository) Get(id uuid.UUID) (*JobBatch, error) {
	value, ok := r.jobs.Load(id)
	if !ok {
		return nil, fmt.Errorf("job batch not found")
	}
	return value.(*JobBatch), nil
}

func (r *DefaultJobRepository) Delete(id uuid.UUID) error {
	r.jobs.Delete(id)
	return nil
}

func (r *DefaultJobRepository) AddInProgress(batchID uuid.UUID, fileJob *FileJob) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.InProgress = append(batch.InProgress, fileJob)
	return nil
}

func (r *DefaultJobRepository) SetFileJobStatus(batchID uuid.UUID, fileJob *FileJob, status string) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	fileJob.Status = status
	if status == "completed" {
		for i, job := range batch.InProgress {
			if job.Name == fileJob.Name {
				batch.Completed = append(batch.Completed, job)
				batch.InProgress = append(batch.InProgress[:i], batch.InProgress[i+1:]...)
				break
			}
		}
	}
	return nil
}

