package repositories

import (
	"context"
	"fmt"
	"freezetag/backend/pkg/database"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JobBatch struct {
	UUID      uuid.UUID             `json:"uuid"`
	Completed []*ImageUploadSuccess `json:"completed"`
	Failed    []*ImageUploadFailure `json:"failed"`

	InProgress []*FileJob `json:"in_progress"`

	Finished chan struct{} `json:"-"`

	mutex sync.Mutex  `json:"-"`
	timer *time.Timer `json:"-"`

	Ctx    context.Context    `json:"-"` // Given to the service
	Cancel context.CancelFunc `json:"-"` // Called by the repository when the job is killed due to idleness or when the batch is deleted
	WG     sync.WaitGroup     `json:"-"` // Used by the service to wait for a job to finish
}

func (b *JobBatch) MarkFinished() {
	b.Finished <- struct{}{} // notify finished
	close(b.Finished)
}

type FileJob struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Bytes  []byte `json:"-"`
}

type JobBatchID uuid.UUID

const (
	MaxIdleTime   = 1 * time.Hour    // if a job batch is idle for 1 hour, assume something has happened to the job and kill it
	RetentionTime = 15 * time.Minute // keep completed job batches for 15 minutes after completion
)

type JobRepository interface {
	Create(*JobBatch) error
	Get(uuid.UUID) (*JobBatch, error)
	Delete(uuid.UUID) error
	WaitFinished(uuid.UUID) <-chan []*ImageUploadSuccess

	AddInProgressFileJob(batchID uuid.UUID, file FileJob) error
	UpdateJobStatus(batchID uuid.UUID, fileName string, status string) error
	CompleteFileJob(batchID uuid.UUID, fileName string, id database.ImageId) error
	FailFileJob(batchID uuid.UUID, fileName string, reason error) error
}

type DefaultJobRepository struct {
	jobs sync.Map // map[uuid.UUID]*JobBatch
}

func NewDefaultJobRepository() JobRepository {
	return &DefaultJobRepository{}
}

func (r *DefaultJobRepository) Create(batch *JobBatch) error {
	ctx, cancel := context.WithCancel(context.Background())
	batch.Ctx = ctx
	batch.Cancel = cancel
	batch.Finished = make(chan struct{}, 1)

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
	value, err := r.Get(id)
	if err != nil {
		return err
	}

	batch := value
	batch.Cancel()
	r.jobs.Delete(id)
	return nil
}

func (r *DefaultJobRepository) WaitFinished(id uuid.UUID) <-chan []*ImageUploadSuccess {
	finished := make(chan []*ImageUploadSuccess, 1)
	batch, err := r.Get(id)
	if err != nil {
		finished <- nil
		return finished
	}
	go func() {
		<-batch.Finished
		if batch.Ctx.Err() != nil {
			finished <- nil
		} else {
			finished <- batch.Completed
		}
		close(finished)
	}()
	return finished
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

func (r *DefaultJobRepository) CompleteFileJob(batchID uuid.UUID, fileName string, id database.ImageId) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	batch.updateIdleStatus(MaxIdleTime)

	for i, job := range batch.InProgress {
		if job.Name == fileName {
			batch.Completed = append(batch.Completed, &ImageUploadSuccess{
				Filename: fileName,
				Id:       id,
			})
			batch.InProgress = append(batch.InProgress[:i], batch.InProgress[i+1:]...)

			if len(batch.InProgress) == 0 {
				if batch.timer != nil {
					batch.timer.Stop()
					batch.timer.Reset(RetentionTime)
				}
				batch.MarkFinished()
			}
			return nil
		}
	}
	return fmt.Errorf("file name not found")
}

func (r *DefaultJobRepository) FailFileJob(batchID uuid.UUID, fileName string, reason error) error {
	batch, err := r.Get(batchID)
	if err != nil {
		return err
	}
	batch.mutex.Lock()
	defer batch.mutex.Unlock()
	batch.updateIdleStatus(MaxIdleTime)

	for i, job := range batch.InProgress {
		if job.Name == fileName {
			batch.Failed = append(batch.Failed, &ImageUploadFailure{
				Filename: job.Name,
				Reason:   reason.Error(),
			})
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
