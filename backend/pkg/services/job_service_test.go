package services

import (
	"testing"

	mockImageRepo "freezetag/backend/mocks/ImageRepository"
	mockJobRepo "freezetag/backend/mocks/JobRepository"
	mockPluginService "freezetag/backend/mocks/PluginService"
	"freezetag/backend/pkg/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	p.EXPECT().
		AllPlugins().
		Return([]string{})

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
