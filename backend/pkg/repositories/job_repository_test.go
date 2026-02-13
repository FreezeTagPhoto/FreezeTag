package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testJobInput int

func (t testJobInput) ID() int {
	return int(t)
}

func TestJobBatchBasic(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := repo.Create(t.Context(), []JobInput{testJobInput(1)}, AtomicJob(func(i testJobInput) (int, error) {
		return int(i), nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Failed)
	require.Len(t, got.Completed, 1)
	assert.Equal(t, 1, got.Completed[0])
}

func TestJobBatchError(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := repo.Create(t.Context(), []JobInput{testJobInput(1)}, AtomicJob(func(i testJobInput) (int, error) {
		return 0, fmt.Errorf("expected error")
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Completed)
	require.Len(t, got.Failed, 1)
	assert.Equal(t, "expected error", got.Failed[0].Reason)
	assert.Equal(t, testJobInput(1), got.Failed[0].Input)
}

func TestJobBatchMixture(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := repo.Create(t.Context(), []JobInput{testJobInput(1), testJobInput(2)}, AtomicJob(func(i testJobInput) (int, error) {
		if i == testJobInput(1) {
			return 3, nil
		} else {
			return 0, fmt.Errorf("expected error")
		}
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	require.Len(t, got.Completed, 1)
	assert.Equal(t, 3, got.Completed[0])
	require.Len(t, got.Failed, 1)
	assert.Equal(t, "expected error", got.Failed[0].Reason)
	assert.Equal(t, testJobInput(2), got.Failed[0].Input)
}

func TestJobPreemptiveCancel(t *testing.T) {
	repo := NewDefaultJobRepository()
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	id := repo.Create(cancelled, []JobInput{testJobInput(1)}, AtomicJob(func(i testJobInput) (int, error) {
		return 1, nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Completed)
	require.Len(t, got.Failed, 1)
	assert.Equal(t, "job cancelled", got.Failed[0].Reason)
}

func TestJobMiddleCancelOuterContext(t *testing.T) {
	repo := NewDefaultJobRepository()
	cancellable, cancel := context.WithCancel(t.Context())
	id := repo.Create(cancellable, []JobInput{testJobInput(1)}, Job(func(i testJobInput, c context.Context) (int, error) {
		time.Sleep(time.Millisecond * 500)
		if c.Err() != nil {
			return 0, fmt.Errorf("mid-job cancel")
		}
		return 1, nil
	}))
	time.Sleep(time.Millisecond * 250)
	cancel()
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Completed)
	require.Len(t, got.Failed, 1)
	assert.Equal(t, "mid-job cancel", got.Failed[0].Reason)
}

func TestJobMiddleCancelInnerContext(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := repo.Create(t.Context(), []JobInput{testJobInput(1)}, Job(func(i testJobInput, c context.Context) (int, error) {
		time.Sleep(time.Millisecond * 500)
		if c.Err() != nil {
			return 0, fmt.Errorf("mid-job cancel")
		}
		return 1, nil
	}))
	time.Sleep(time.Millisecond * 250)
	got := repo.Get(id)
	require.NotNil(t, got)
	got.Cancel()
	<-got.WaitFinished()
	assert.Empty(t, got.Completed)
	require.Len(t, got.Failed, 1)
	assert.Equal(t, "mid-job cancel", got.Failed[0].Reason)
}

func TestManyJobs(t *testing.T) {
	repo := NewDefaultJobRepository()
	inputs := []JobInput{}
	for i := range 20 {
		inputs = append(inputs, testJobInput(i))
	}
	id := repo.Create(t.Context(), inputs, AtomicJob(func(i testJobInput) (int, error) {
		return int(i), nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Failed)
	assert.Len(t, got.Completed, 20)
	for i := range 20 {
		assert.Contains(t, got.Completed, i)
	}
}

func TestManyBigJobs(t *testing.T) {
	repo := NewDefaultJobRepository()
	inputs := []JobInput{}
	for i := range 20 {
		inputs = append(inputs, testJobInput(i))
	}
	id := repo.Create(t.Context(), inputs, AtomicJob(func(i testJobInput) (int, error) {
		// pretend to be a big job (require using the queue)
		time.Sleep(100 * time.Millisecond)
		return int(i), nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	assert.Empty(t, got.Failed)
	assert.Len(t, got.Completed, 20)
	for i := range 20 {
		assert.Contains(t, got.Completed, i)
	}
}

func TestSynchronousJobs(t *testing.T) {
	repo := NewDefaultJobRepository()
	inputs := []JobInput{}
	for i := range 50 {
		inputs = append(inputs, testJobInput(i))
	}
	index := 0
	ensureSync := func(i int) {
		t.Helper()
		assert.Equal(t, index, i)
		index++
	}
	id := repo.CreateSync(t.Context(), inputs, AtomicJob(func(i testJobInput) (int, error) {
		// waking from sleep would jumble them up if these weren't synchronous
		time.Sleep(5 * time.Millisecond)
		ensureSync(int(i))
		return int(i), nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
}

func TestDeleteJob(t *testing.T) {
	repo := NewDefaultJobRepository()
	id := repo.Create(t.Context(), []JobInput{testJobInput(1)}, AtomicJob(func(i testJobInput) (int, error) {
		return int(i), nil
	}))
	got := repo.Get(id)
	require.NotNil(t, got)
	<-got.WaitFinished()
	err := repo.Delete(id)
	assert.NoError(t, err)
	got = repo.Get(id)
	assert.Nil(t, got)
}
