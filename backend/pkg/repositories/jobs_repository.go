package repositories

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	MaxIdleTime   = 1 * time.Hour    // if a job batch is idle for 1 hour, assume something has happened to the job and kill it
	RetentionTime = 15 * time.Minute // keep completed job batches for 15 minutes after completion
	MaxJobThreads = 10
)

// global job counter for keeping thread counts to a reasonable level
var jobCounter = struct {
	lock       sync.Mutex
	activeJobs uint
	waiting    []chan<- func()
}{}

func decreaseJobCounter() {
	jobCounter.lock.Lock()
	defer jobCounter.lock.Unlock()
	if jobCounter.activeJobs < MaxJobThreads && len(jobCounter.waiting) > 0 {
		sub := jobCounter.waiting[0]
		sub <- decreaseJobCounter
		jobCounter.waiting = jobCounter.waiting[1:]
	} else {
		jobCounter.activeJobs--
	}
}

func waitForJob() <-chan func() {
	jobCounter.lock.Lock()
	defer jobCounter.lock.Unlock()
	sub := make(chan func(), 1)
	if jobCounter.activeJobs < MaxJobThreads {
		sub <- decreaseJobCounter
		jobCounter.activeJobs++
	} else {
		jobCounter.waiting = append(jobCounter.waiting, sub)
	}
	return sub
}

type JobInput interface {
	ID() int // ID should be unique between JobInputs in a batch
}

type JobFunction[I JobInput, O any] func(I, context.Context, func(string)) (O, error)

func Job[I JobInput, O any](f func(I, context.Context, func(string)) (O, error)) func(JobInput, context.Context, func(string)) (any, error) {
	return func(i JobInput, c context.Context, s func(string)) (any, error) {
		return f(i.(I), c, s)
	}
}

func AtomicJob[I JobInput, O any](f func(I, func(string)) (O, error)) func(JobInput, context.Context, func(string)) (any, error) {
	return func(i JobInput, c context.Context, s func(string)) (any, error) {
		if c.Err() != nil {
			return *new(O), fmt.Errorf("job cancelled")
		}
		return f(i.(I), s)
	}
}

func SimpleJob[I JobInput, O any](f func(I) (O, error)) func(JobInput, context.Context, func(string)) (any, error) {
	return func(i JobInput, c context.Context, s func(string)) (any, error) {
		if c.Err() != nil {
			return *new(O), fmt.Errorf("job cancelled")
		}
		return f(i.(I))
	}
}

type JobBatch[I JobInput, C any] struct {
	UUID        uuid.UUID          `json:"uuid"`
	Title       string             `json:"title"`
	Status      string             `json:"status"`
	Completed   []C                `json:"completed"`
	Failed      []jobFailure[I]    `json:"failed"`
	InProgress  []I                `json:"in_progress"`
	Cancelled   bool               `json:"cancelled"`
	Context     context.Context    `json:"-"`
	Lock        sync.Mutex         `json:"-"`
	finished    bool               `json:"-"`
	operation   JobFunction[I, C]  `json:"-"`
	cancel      context.CancelFunc `json:"-"`
	timer       *time.Timer        `json:"-"`
	subscribers []chan<- struct{}  `json:"-"`
	workers     sync.WaitGroup     `json:"-"`
	synchronous bool               `json:"-"` // whether or not the job should be run in order
}

func (jb *JobBatch[I, C]) removeByID(id int) {
	jb.Lock.Lock()
	defer jb.Lock.Unlock()
	jb.InProgress = slices.DeleteFunc(jb.InProgress, func(i I) bool {
		return i.ID() == id
	})
}

type jobFailure[I any] struct {
	Input  I      `json:"input"`
	Reason string `json:"reason"`
}

func (jb *JobBatch[I, C]) addResults(i I, c C, f error) {
	jb.Lock.Lock()
	defer jb.Lock.Unlock()
	if f != nil {
		jb.Failed = append(jb.Failed, jobFailure[I]{Input: i, Reason: f.Error()})
	} else {
		jb.Completed = append(jb.Completed, c)
	}
}

// run a job batch.
// this function does not block. use WaitFinished after run if you want to block.
func (jb *JobBatch[I, C]) run() {
	if len(jb.InProgress) == 0 {
		// empty job special case
		jb.finish()
		return
	}
	var count atomic.Int64
	count.Store(int64(len(jb.InProgress)))
	waitPrev := make(chan struct{}, 1)
	waitPrev <- struct{}{}
	close(waitPrev)
	jb.Lock.Lock()
	jb.Status = "Running"
	changeStatus := func(s string) {
		jb.Lock.Lock()
		defer jb.Lock.Unlock()
		jb.Status = s
	}
	for _, input := range jb.InProgress {
		wait := make(chan struct{}, 1)
		workerWait := waitPrev
		jb.workers.Go(func() {
			// wait for previous
			<-workerWait
			// if async is allowed, notify now
			if !jb.synchronous {
				wait <- struct{}{}
				close(wait)
			}
			finished := <-waitForJob()
			thing, err := jb.operation(input, jb.Context, changeStatus)
			jb.removeByID(input.ID())
			jb.addResults(input, thing, err)
			finished()
			// decrement the count and maybe notify complete job
			if count.Add(-1) == 0 {
				jb.finish()
			}
			// if async is not allowed, notify now
			if jb.synchronous {
				wait <- struct{}{}
				close(wait)
			}
		})
		waitPrev = wait
	}
	jb.Lock.Unlock()
}

func (jb *JobBatch[I, C]) notifySubscribers() {
	for _, sub := range jb.subscribers {
		sub <- struct{}{}
		close(sub)
	}
}

func (jb *JobBatch[I, C]) adjustKeepTime(d time.Duration) {
	if !jb.timer.Reset(d) {
		// stop the timer if the delete job already ran (how did we get here)
		jb.timer.Stop()
	}
}

// cancel a job batch. This function is non-blocking
func (jb *JobBatch[I, C]) Cancel() {
	jb.Cancelled = true
	jb.cancel()
	// finish will get called eventually and notify subscribers
}

// this should only be called once, by the job that finishes the batch
func (jb *JobBatch[I, C]) finish() {
	jb.finished = true
	if jb.Cancelled {
		jb.Status = "Cancelled"
	} else {
		jb.Status = "Finished"
	}
	jb.notifySubscribers()
	jb.adjustKeepTime(RetentionTime)
}

// return a channel that completes when the job batch finishes
func (jb *JobBatch[I, C]) WaitFinished() <-chan struct{} {
	sub := make(chan struct{}, 1)
	if jb.finished {
		// if it's already finished, it's obviously finished
		sub <- struct{}{}
		close(sub)
	} else {
		jb.subscribers = append(jb.subscribers, sub)
	}
	return sub
}

type UploadJob struct {
	Name  string `json:"name"`
	Bytes []byte `json:"-"`
}

type JobRepository interface {
	// create and start a job batch that uses the operator on the passed data
	Create(string, context.Context, []JobInput, func(JobInput, context.Context, func(string)) (any, error)) uuid.UUID
	// create and start a synchronous job batch
	CreateSync(string, context.Context, []JobInput, func(JobInput, context.Context, func(string)) (any, error)) uuid.UUID
	// list all jobs in the repository
	AllJobs() []*JobBatch[JobInput, any]
	// get a job by UUID
	Get(uuid.UUID) *JobBatch[JobInput, any]
	// cancel and delete a job by UUID
	Delete(uuid.UUID) error
}

type DefaultJobRepository struct {
	jobs map[uuid.UUID]*JobBatch[JobInput, any]
	lock sync.RWMutex
}

func NewDefaultJobRepository() JobRepository {
	return &DefaultJobRepository{jobs: make(map[uuid.UUID]*JobBatch[JobInput, any])}
}

func (r *DefaultJobRepository) AllJobs() []*JobBatch[JobInput, any] {
	r.lock.RLock()
	defer r.lock.RUnlock()
	jobs := make([]*JobBatch[JobInput, any], 0, len(r.jobs))
	for _, value := range r.jobs {
		jobs = append(jobs, value)
	}
	return jobs
}

func (r *DefaultJobRepository) Create(title string, ctx context.Context, data []JobInput, operator func(JobInput, context.Context, func(string)) (any, error)) uuid.UUID {
	r.lock.Lock()
	defer r.lock.Unlock()
	id := uuid.New()
	ctx, cancel := context.WithCancel(ctx)
	r.jobs[id] = &JobBatch[JobInput, any]{
		UUID:        id,
		Title:       title,
		Status:      "Created",
		Completed:   []any{},
		Failed:      []jobFailure[JobInput]{},
		InProgress:  data,
		Context:     ctx,
		operation:   operator,
		cancel:      cancel,
		synchronous: false,
		timer: time.AfterFunc(MaxIdleTime, func() {
			r.Delete(id) //nolint:errcheck
		}),
	}
	r.jobs[id].run()
	return id
}

func (r *DefaultJobRepository) CreateSync(title string, ctx context.Context, data []JobInput, operator func(JobInput, context.Context, func(string)) (any, error)) uuid.UUID {
	r.lock.Lock()
	defer r.lock.Unlock()
	id := uuid.New()
	ctx, cancel := context.WithCancel(ctx)
	r.jobs[id] = &JobBatch[JobInput, any]{
		UUID:        id,
		Title:       title,
		Status:      "Created",
		Completed:   []any{},
		Failed:      []jobFailure[JobInput]{},
		InProgress:  data,
		Context:     ctx,
		operation:   operator,
		cancel:      cancel,
		synchronous: true,
		timer: time.AfterFunc(MaxIdleTime, func() {
			r.Delete(id) //nolint:errcheck
		}),
	}
	r.jobs[id].run()
	return id
}

func (r *DefaultJobRepository) Get(id uuid.UUID) *JobBatch[JobInput, any] {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.jobs[id]
}

func (r *DefaultJobRepository) Delete(id uuid.UUID) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil
	}
	job.Cancel()
	delete(r.jobs, id)
	return nil
}
