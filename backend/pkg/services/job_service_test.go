package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	mockImageRepo "freezetag/backend/mocks/ImageRepository"
	mockJobRepo "freezetag/backend/mocks/JobRepository"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRunUploadJobsAsyncExecution(t *testing.T) {

	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	jobService := InitDefaultJobService(m, i)

	batchID := repositories.NewJobBatchID()
	fileName := "test.png"
	fileData := []byte("test data")

	ctx := t.Context()

	jobBatch := &repositories.JobBatch{
		UUID: batchID,
		Ctx:  ctx,
		InProgress: []*repositories.FileJob{
			{Name: fileName, Bytes: fileData},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// expectedResult := repositories.UploadResult{
	// 	Success: &repositories.ImageUploadSuccess{Filename: fileName},
	// }

	i.EXPECT().
		StoreImageBytes(fileData, fileName).
		Return(database.ImageId(42), nil).
		Once()
	m.EXPECT().
		CompleteFileJob(batchID, fileName, database.ImageId(42)).
		Run(func(uuid uuid.UUID, name string, id database.ImageId) {
			wg.Done() // Signal that CompleteFileJob was called
		}).
		Return(nil).
		Once()

	err := jobService.RunUploadJobs(jobBatch)
	require.NoError(t, err)
	done := make(chan struct{})

	// make sure that the goroutine has time to execute and call CompleteFileJob
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out: Background goroutine never called CompleteFileJob after a 5 second wait")
	}
}
func TestRunUploadJobsAsyncExecutionStress(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	jobService := InitDefaultJobService(m, i)
	iterations := 100

	batchID := repositories.NewJobBatchID()
	fileData := []byte("test data")

	ctx := t.Context()

	jobs := make([]*repositories.FileJob, iterations)
	for i := range jobs {
		jobs[i] = &repositories.FileJob{Name: fmt.Sprint(i), Bytes: fileData}
	}

	jobBatch := &repositories.JobBatch{
		UUID:       batchID,
		Ctx:        ctx,
		InProgress: jobs,
	}

	var wg sync.WaitGroup
	wg.Add(iterations)

	i.EXPECT().
		StoreImageBytes(fileData, mock.AnythingOfType("string")).
		Return(database.ImageId(42), nil).
		Times(iterations)
	m.EXPECT().
		CompleteFileJob(batchID, mock.AnythingOfType("string"), mock.Anything).
		Run(func(BatchId uuid.UUID, name string, ImageId database.ImageId) {
			wg.Done() // Decrement counter 100 times
		}).
		Return(nil).
		Times(iterations)

	err := jobService.RunUploadJobs(jobBatch)
	require.NoError(t, err)
	done := make(chan struct{})

	// make sure that the goroutine has time to execute and call CompleteFileJob
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out: Background goroutine never called CompleteFileJob after a 5 second wait")
	}
}

func TestCreateJobBatch(t *testing.T) {
	i := mockImageRepo.NewMockImageRepository(t)
	m := mockJobRepo.NewMockJobRepository(t)
	jobService := InitDefaultJobService(m, i)

	jobs := []*repositories.FileJob{
		{Name: "file1.png", Bytes: []byte("data1")},
		{Name: "file2.png", Bytes: []byte("data2")},
	}
	id := repositories.NewJobBatchID()
	expectedBatch := &repositories.JobBatch{
		UUID:       id,
		InProgress: jobs,
	}

	m.EXPECT().
		Create(mock.AnythingOfType("*repositories.JobBatch")).
		Run(func(batch *repositories.JobBatch) {
			batch.UUID = id // Set the UUID to match expected value for assertion
		}).Return(nil).Once()

	resultBatch, err := jobService.CreateJobBatch(jobs)
	require.NoError(t, err)
	require.Equal(t, expectedBatch.UUID, resultBatch.UUID)
	require.Equal(t, expectedBatch.InProgress, resultBatch.InProgress)

}

func TestCreateJobBatch2(t *testing.T) {
	// Setup
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)

	files := []*repositories.FileJob{
		{Name: "a.png", Bytes: []byte("a")},
		{Name: "b.png", Bytes: []byte("b")},
	}

	m.EXPECT().
		Create(mock.MatchedBy(func(b *repositories.JobBatch) bool {
			return len(b.InProgress) == 2 && b.UUID != uuid.Nil
		})).
		Return(nil).
		Once()

	batch, err := service.CreateJobBatch(files)
	require.NoError(t, err)
	assert.NotNil(t, batch)
	assert.NotEqual(t, uuid.Nil, batch.UUID)
	assert.Equal(t, 2, len(batch.InProgress))
}

func TestCreateJobBatchRepoFailure(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)

	files := []*repositories.FileJob{{Name: "a.png"}}
	dbError := assert.AnError

	m.EXPECT().Create(mock.Anything).Return(dbError).Once()

	batch, err := service.CreateJobBatch(files)
	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.Nil(t, batch)
}

func TestRunUploadJobsEmptyBatch(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)

	batch := &repositories.JobBatch{
		UUID:       repositories.NewJobBatchID(),
		Ctx:        context.Background(),
		InProgress: []*repositories.FileJob{},
	}

	i.AssertNotCalled(t, "StoreImageBytes")
	m.AssertNotCalled(t, "CompleteFileJob")

	err := service.RunUploadJobs(batch)
	require.NoError(t, err)
}

func TestRunUploadJobsRespectsContextCancellation(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Error(t, ctx.Err())

	batch := &repositories.JobBatch{
		UUID: repositories.NewJobBatchID(),
		Ctx:  ctx,
		InProgress: []*repositories.FileJob{
			{Name: "test.png", Bytes: []byte("data")},
		},
	}

	i.AssertNotCalled(t, "StoreImageBytes")
	m.AssertNotCalled(t, "CompleteFileJob")

	err := service.RunUploadJobs(batch)
	require.NoError(t, err)

	batch.WG.Wait() // Wait for the background goroutine to finish
}

func TestRunUploadJobsRespectsContextCancellationStress(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)

	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, ctx.Err())

	jobs := make([]*repositories.FileJob, 100)
	for i := range jobs {
		jobs[i] = &repositories.FileJob{Name: fmt.Sprint(i), Bytes: fmt.Append(nil, "data", i)}
	}

	batchID := repositories.NewJobBatchID()
	batch := &repositories.JobBatch{
		UUID:       batchID,
		Ctx:        ctx,
		InProgress: jobs,
	}

	limit := ThreadLimit
	var wg sync.WaitGroup
	wg.Add(limit)
	done := make(chan struct{})

	i.EXPECT().
		StoreImageBytes(mock.Anything, mock.AnythingOfType("string")).
		Run(func(data []byte, name string) {
			wg.Done()
			<-done
		}).
		Return(database.ImageId(42), nil).Times(limit)
	m.EXPECT().
		CompleteFileJob(batchID, mock.AnythingOfType("string"), mock.Anything).
		Return(nil).
		Maybe()
	err := service.RunUploadJobs(batch)
	require.NoError(t, err)

	wg.Wait()
	cancel()
	close(done)
	batch.WG.Wait() // Wait for the background goroutine to finish
	assert.Error(t, ctx.Err())
}

func TestRunUploadJobsCompletesStress(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	service := InitDefaultJobService(m, i)



	jobs := make([]*repositories.FileJob, 100)
	for i := range jobs {
		jobs[i] = &repositories.FileJob{Name: fmt.Sprint(i), Bytes: fmt.Append(nil, "data", i)}
	}

	batchID := repositories.NewJobBatchID()
	batch := &repositories.JobBatch{
		UUID:       batchID,
		Ctx:        context.Background(),
		InProgress: jobs,
	}

	var wg sync.WaitGroup
	wg.Add(100)

	i.EXPECT().
		StoreImageBytes(mock.Anything, mock.AnythingOfType("string")).
		Run(func(data []byte, name string) {
			wg.Done()
		}).
		Return(database.ImageId(42), nil).Times(100)
	m.EXPECT().
		CompleteFileJob(batchID, mock.AnythingOfType("string"), mock.Anything).
		Return(nil).
		Maybe()
	err := service.RunUploadJobs(batch)
	require.NoError(t, err)

	wg.Wait()
	batch.WG.Wait() // Wait for the background goroutine to finish
}
