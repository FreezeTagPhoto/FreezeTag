package services

import (
	"os"
	"strings"
	"testing"
	"time"

	mockImageRepo "freezetag/backend/mocks/ImageRepository"
	mockJobRepo "freezetag/backend/mocks/JobRepository"
	mockPluginService "freezetag/backend/mocks/PluginService"
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
	p := mockPluginService.NewMockPluginService(t)
	jobService := InitDefaultJobService(m, i, p)

	id := uuid.New()

	fileName := "test.png"
	fileData := []byte("test data")
	m.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(id).
		Times(2)
	m.EXPECT().
		Get(id).
		Return(&repositories.JobBatch[repositories.JobInput, any]{}).
		Once()

	testid := jobService.RunUploadJob([]FileJob{
		{Name: fileName, Bytes: fileData},
	})
	assert.Equal(t, id, testid)
}

func TestGetJobSummary(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	p := mockPluginService.NewMockPluginService(t)
	jobService := InitDefaultJobService(m, i, p)

	id := uuid.New()
	m.EXPECT().
		Get(id).
		Return(&repositories.JobBatch[repositories.JobInput, any]{
			UUID:   id,
			Title:  "foo",
			Status: "bar",
		})
	summary := jobService.GetSummary(id)
	assert.NotNil(t, summary)
	assert.Equal(t, id, summary.UUID)
	assert.Equal(t, "foo", summary.Title)
	assert.Equal(t, "bar", summary.Status)
}

func TestGetJobList(t *testing.T) {
	m := mockJobRepo.NewMockJobRepository(t)
	i := mockImageRepo.NewMockImageRepository(t)
	p := mockPluginService.NewMockPluginService(t)
	jobService := InitDefaultJobService(m, i, p)

	id1 := uuid.New()
	id2 := uuid.New()
	m.EXPECT().
		AllJobs().
		Return([]*repositories.JobBatch[repositories.JobInput, any]{
			{UUID: id1, Title: "foo", Status: "bar"},
			{UUID: id2, Title: "abc", Status: "def"},
		})
	list := jobService.AllJobs()
	assert.ElementsMatch(t, []JobSummary{
		{UUID: id1, Title: "foo", Status: "bar"},
		{UUID: id2, Title: "abc", Status: "def"},
	}, list)
}

func TestPostUploadPlugin(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("integration/foo/.venv") //nolint:errcheck
	})
	i := mockImageRepo.NewMockImageRepository(t)
	p, err := InitDefaultPluginService("integration", i)
	require.NoError(t, err)
	j := repositories.NewDefaultJobRepository()
	jobService := InitDefaultJobService(j, i, p)
	batch := []FileJob{{Name: "foo.png", Bytes: []byte{}}}
	i.EXPECT().StoreImageBytes(mock.Anything, mock.Anything).Return(database.ImageId(1), nil)
	i.EXPECT().AddImageTags(mock.Anything, mock.Anything).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 1, Count: 1}}).Times(2)
	data, err := os.ReadFile("test_resources/gopher.webp")
	require.NoError(t, err)
	i.EXPECT().RetrieveThumbnail(mock.Anything, mock.Anything).Return(data, nil)
	j1 := jobService.RunUploadJob(batch)
	job1 := jobService.GetBatch(j1)
	require.NotNil(t, job1)
	<-job1.WaitFinished()
	time.Sleep(1 * time.Second) // plenty of time for plugin jobs to kick off
	jobs := jobService.AllJobs()
	job1 = nil
	for _, job := range jobs {
		if strings.Contains(job.Title, "foo") {
			job1 = jobService.GetBatch(job.UUID)
			break
		}
	}
	require.NotNil(t, job1)
	<-job1.WaitFinished()
	assert.Empty(t, job1.Failed)
}

func TestManualPlugin(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("integration/foo/.venv") //nolint:errcheck
	})
	i := mockImageRepo.NewMockImageRepository(t)
	data, err := os.ReadFile("test_resources/gopher.webp")
	require.NoError(t, err)
	i.EXPECT().RetrieveThumbnail(mock.Anything, mock.Anything).Return(data, nil)
	i.EXPECT().AddImageTags(database.ImageId(42), []string{"foo"}).Return(repositories.ImageTagResult{Success: &repositories.ImageTagSuccess{Id: 42, Count: 1}})
	p, err := InitDefaultPluginService("integration", i)
	require.NoError(t, err)
	j := repositories.NewDefaultJobRepository()
	jobService := InitDefaultJobService(j, i, p)
	jobId := jobService.SchedulePluginHook("foo", "foo_image", database.ImageId(42))
	job := jobService.GetBatch(jobId)
	require.NotNil(t, job)
	<-job.WaitFinished()
}
