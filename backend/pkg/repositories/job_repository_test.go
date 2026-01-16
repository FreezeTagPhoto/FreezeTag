package repositories

import (
	"freezetag/backend/pkg/database"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobBatchBasic(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	batch := &JobBatch{
		UUID: id,
		InProgress: nil, 
		Completed: nil,
		Failed: nil,
	}

	err := repo.Create(batch)
	require.NoError(t, err)
	assert.NotNil(t, batch.timer, "Timer should be initialized on Create")

	got, err := repo.Get(id)
	require.NoError(t, err)
	assert.Equal(t, id, got.UUID)

	_, err = repo.Get(uuid.New())
	assert.Error(t, err)
	assert.Equal(t, "job batch not found", err.Error())

	err = repo.Delete(id)
	require.NoError(t, err)

	_, err = repo.Get(id)
	assert.Error(t, err, "Should not find batch after deletion")
}

func TestFileJobLifecycle(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	batch := &JobBatch{UUID: id}
	require.NoError(t, repo.Create(batch))


	filename := "test.png"
	file := FileJob{Name: filename, Status: "pending"}

	err := repo.AddInProgressFileJob(id, file)
	require.NoError(t, err)

	batch, _ = repo.Get(id)
	require.Len(t, batch.InProgress, 1)
	assert.Equal(t, filename, batch.InProgress[0].Name)

	err = repo.UpdateJobStatus(id, filename, "uploading")
	require.NoError(t, err)
	assert.Equal(t, "uploading", batch.InProgress[0].Status)

	mockResult := database.ImageId(42)
	err = repo.CompleteFileJob(id, filename, mockResult)
	require.NoError(t, err)

	require.Len(t, batch.InProgress, 0, "InProgress should be empty")
	require.Len(t, batch.Completed, 1, "Results should have 1 item")
}

func TestCompleteFileJobFileNotFound(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	require.NoError(t, repo.Create(&JobBatch{UUID: id}))

	err := repo.CompleteFileJob(id, "ghost_file.png", database.ImageId(0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file name not found")
}

func TestUpdateJobStatusFileNotFound(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	require.NoError(t, repo.Create(&JobBatch{UUID: id}))

	err := repo.UpdateJobStatus(id, "test.png", "done")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file name not found")
}

func TestWithConcurrencyBasic(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	require.NoError(t, repo.Create(&JobBatch{UUID: id}))

	require.NoError(t, repo.AddInProgressFileJob(id, FileJob{Name: "race.png"}))
	done := make(chan bool)

	go func() {
		_ = repo.UpdateJobStatus(id, "race.png", "A")
		done <- true
	}()

	go func() {
		_ = repo.UpdateJobStatus(id, "race.png", "B")
		done <- true
	}()

	<-done
	<-done
}

func TestWithConcurrencyStress(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := NewJobBatchID()
	require.NoError(t, repo.Create(&JobBatch{UUID: id}))
	iterations := 100

	require.NoError(t, repo.AddInProgressFileJob(id, FileJob{Name: "race.png"}))
	done := make(chan error, iterations * 2)

	for i := range iterations {
		go func(i int) {
			done <- repo.UpdateJobStatus(id, "race.png", strconv.Itoa(i))
		}(i)			
	}

	jobs := make(chan FileJob, iterations)
	for i := range iterations { 
		go func(i int) { 
			FileJob := FileJob{Name: "file" + strconv.Itoa(i)}
			_ = repo.AddInProgressFileJob(id, FileJob)
			jobs <- FileJob
		}(i)
	}

	for range iterations {
		job := <- jobs
		go func(job FileJob) {
			done <- repo.CompleteFileJob(id, job.Name, database.ImageId(0))
		}(job)
	}

	for range iterations * 2 {	
		assert.NoError(t, <-done)
	}
}