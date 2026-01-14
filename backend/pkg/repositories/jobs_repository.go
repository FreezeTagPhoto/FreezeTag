package repositories

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JobBatch struct {
	UUID       uuid.UUID       `json:"uuid"`
	Results    []*UploadResult `json:"results"`
	InProgress []*FileJob      `json:"in_progress"`

	mutex sync.Mutex  `json:"-"`
	timer *time.Timer `json:"-"`
}

type FileJob struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Bytes  []byte `json:"-"`
}

const (
	MaxIdleTime   = 1 * time.Hour    // if a job batch is idle for 1 hour, assume something has happened to the job and kill it
	RetentionTime = 15 * time.Minute // keep completed job batches for 15 minutes after completion
)

type JobRepository interface {
	Create(*JobBatch) error
	Get(uuid.UUID) (*JobBatch, error)
	Delete(uuid.UUID) error

	AddInProgressFileJob(batchID uuid.UUID, file FileJob) error
	CompleteFileJob(batchID uuid.UUID, fileName string, result UploadResult) error
	UpdateJobStatus(batchID uuid.UUID, fileName string, status string) error
}

type DefaultJobRepository struct {
	jobs sync.Map // map[uuid.UUID]*JobBatch
}

func NewDefaultJobRepository() *DefaultJobRepository {
	return &DefaultJobRepository{}
}

func (r *DefaultJobRepository) Create(batch *JobBatch) error {
	r.jobs.Store(batch.UUID, batch)
	batch.timer = time.AfterFunc(MaxIdleTime, func() {
		_ = r.Delete(batch.UUID)
	})
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

func (r *DefaultJobRepository) AddInProgressFileJob(batchID uuid.UUID, fileJob FileJob) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	batch.InProgress = append(batch.InProgress, &fileJob)
	return nil
}

func (r *DefaultJobRepository) CompleteFileJob(batchID uuid.UUID, fileName string, result UploadResult) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	batch.updateIdleStatus(MaxIdleTime)

	for i, job := range batch.InProgress {
		if job.Name == fileName {
			batch.Results = append(batch.Results, &result)
			batch.InProgress = append(batch.InProgress[:i], batch.InProgress[i+1:]...)

			if len(batch.InProgress) == 0 {
				if batch.timer != nil {
					batch.timer.Stop()
					batch.timer.Reset(RetentionTime)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("file name not found")
}

func (r *DefaultJobRepository) UpdateJobStatus(batchID uuid.UUID, fileName string, status string) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	batch.updateIdleStatus(MaxIdleTime)

	for _, job := range batch.InProgress {
		if job.Name == fileName {
			job.Status = status
			return nil
		}
	}
	return fmt.Errorf("file name not found")
}

func NewJobBatchID() uuid.UUID {
	return uuid.New()
}

func (batch *JobBatch) updateIdleStatus(reset time.Duration) {
	if batch.timer != nil {
		batch.timer.Stop()
		batch.timer.Reset(reset)
	}
}
