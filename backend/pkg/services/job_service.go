package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"freezetag/backend/pkg/plugins"
	"freezetag/backend/pkg/repositories"

	"github.com/google/uuid"
)

// only one plugin can run at a time
// (TODO: change this eventually to be plugin-specific)
var pluginLock sync.Mutex
var currentPlugin *plugins.HookedPlugin

func lockNewPlugin(p *plugins.HookedPlugin) {
	if p == currentPlugin {
		return // already locked in
	}
	currentPlugin = p
	go func() {
		<-currentPlugin.WaitFinished()
		pluginLock.Unlock()
	}()
}

type FileJob struct {
	Name  string `json:"name"`
	Bytes []byte `json:"-"`
}

type JobSummary struct {
	UUID       uuid.UUID `json:"uuid"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	InProgress int       `json:"in_progress"`
	Complete   int       `json:"complete"`
	Errors     int       `json:"errors"`
}

type innerFileJob struct {
	FileJob
	id int `json:"-"`
}

func (j innerFileJob) ID() int {
	return j.id
}

type JobService interface {
	GetBatch(uuid.UUID) *repositories.JobBatch[repositories.JobInput, any]
	GetSummary(uuid.UUID) *JobSummary
	AllJobs() []JobSummary
	RunUploadJob(files []FileJob) uuid.UUID
	SchedulePostUploads(upload uuid.UUID)
	SchedulePluginHook(plugin string, hook string, input any) uuid.UUID
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

func (s *defaultJobService) GetBatch(id uuid.UUID) *repositories.JobBatch[repositories.JobInput, any] {
	return s.jobRepository.Get(id)
}

func (s *defaultJobService) GetSummary(id uuid.UUID) *JobSummary {
	job := s.jobRepository.Get(id)
	if job == nil {
		return nil
	}
	job.Lock.Lock()
	defer job.Lock.Unlock()
	return &JobSummary{
		UUID:       job.UUID,
		Title:      job.Title,
		Status:     job.Status,
		InProgress: len(job.InProgress),
		Complete:   len(job.Completed),
		Errors:     len(job.Failed),
	}
}

func (s *defaultJobService) AllJobs() []JobSummary {
	jobs := s.jobRepository.AllJobs()
	summaries := make([]JobSummary, 0, len(jobs))
	for _, job := range jobs {
		job.Lock.Lock()
		summaries = append(summaries, JobSummary{
			UUID:       job.UUID,
			Title:      job.Title,
			Status:     job.Status,
			InProgress: len(job.InProgress),
			Complete:   len(job.Completed),
			Errors:     len(job.Failed),
		})
		job.Lock.Unlock()
	}
	return summaries
}

func (s *defaultJobService) uploadOneFile(f innerFileJob) (repositories.ImageUploadSuccess, error) {
	id, err := s.imageRepository.StoreImageBytes(f.Bytes, f.Name)
	if err != nil {
		return repositories.ImageUploadSuccess{}, err
	}
	return repositories.ImageUploadSuccess{Id: id, Filename: f.Name}, nil
}

func (s *defaultJobService) RunUploadJob(batch []FileJob) uuid.UUID {
	jobs := make([]repositories.JobInput, len(batch))
	for i, job := range batch {
		jobs[i] = innerFileJob{job, i}
	}
	id := s.jobRepository.Create(fmt.Sprintf("Uploading %d files", len(batch)), context.Background(), jobs, repositories.SimpleJob(s.uploadOneFile))
	s.SchedulePostUploads(id)
	return id
}

type pluginRun struct {
	Name  string `json:"name"`
	Hook  string `json:"hook"`
	Input any    `json:"input"`
	id    int    `json:"-"`
}

func (p pluginRun) ID() int {
	return p.id
}

type dummy struct {
	id int `json:"-"`
}

func (d dummy) ID() int {
	return d.id
}

// manual hooks run outside the "one plugin at a time" system
// since they only run one job, and they should run when the user wants
func (s *defaultJobService) SchedulePluginHook(plugin string, hook string, input any) uuid.UUID {
	return s.jobRepository.Create(
		"Manual plugin run "+plugin+":"+hook,
		context.Background(),
		[]repositories.JobInput{pluginRun{
			Name:  plugin,
			Hook:  hook,
			Input: input,
			id:    1,
		}},
		repositories.Job(func(in pluginRun, c context.Context, _ func(string)) (plugins.PluginResult, error) {
			log.Printf("[INFO] launching plugin %s", in.Name)
			p, err := s.plugins.LaunchPlugin(in.Name, c)
			if err != nil {
				log.Printf("[ERR]  failed to launch plugin %s", in.Name)
				return nil, fmt.Errorf("plugin failed to launch: %w", err)
			}
			return p.RunHook(in.Hook, in.Input, s.imageRepository)
		}),
	)
}

func (s *defaultJobService) SchedulePostUploads(upload uuid.UUID) {
	batch := s.GetBatch(upload)
	s.jobRepository.Create(
		"Schedule PostUpload plugin hooks for upload "+batch.UUID.String(),
		batch.Context,
		[]repositories.JobInput{dummy{1}},
		repositories.Job(func(_ dummy, c context.Context, status func(string)) (plugins.PluginResult, error) {
			status("Waiting")
			<-batch.WaitFinished()
			status("Running")
			if batch.Cancelled {
				return nil, fmt.Errorf("cancelled")
			}
			uploads := make([]repositories.ImageUploadSuccess, len(batch.Completed))
			for i, succ := range batch.Completed {
				uploads[i] = succ.(repositories.ImageUploadSuccess)
			}
			// create a job for every plugin that has PostUpload hooks
			for _, plugin := range s.plugins.Plugins() {
				if !plugin.Enabled {
					continue
				}
				hooks := plugin.HooksWithType(plugins.PostUpload)
				if len(hooks) == 0 {
					continue
				}
				var jobs []repositories.JobInput
				count := 0
				// schedule all the batchwise PostUpload hooks to run first
				batchHooks := plugin.FilterHooks(plugins.PostUpload, plugins.ProcessImageBatch)
				for _, hook := range batchHooks {
					jobs = append(jobs, pluginRun{Name: plugin.Name, Hook: hook.Name, Input: uploads, id: count})
					count++
				}
				// schedule all the individual PostUpload hooks to run in sequence next
				indivHooks := plugin.FilterHooks(plugins.PostUpload, plugins.ProcessOneImage)
				for _, hook := range indivHooks {
					for _, upload := range uploads {
						jobs = append(jobs, pluginRun{Name: plugin.Name, Hook: hook.Name, Input: upload, id: count})
						count++
					}
				}
				// launch the job
				// new context because being here means that the upload wasn't cancelled
				// and we don't want something like a 15 minute run to break
				ctx := context.Background()
				idx := 0
				dead := false
				s.jobRepository.CreateSync(
					fmt.Sprintf("%s PostUpload on upload %s", plugin.Name, batch.UUID.String()),
					ctx,
					jobs,
					repositories.Job(func(in pluginRun, c context.Context, status func(string)) (plugins.PluginResult, error) {
						defer func() {
							idx++
							if idx == count && !dead {
								log.Printf("[INFO] shutting down plugin %s", in.Name)
								err := currentPlugin.Shutdown()
								if err != nil {
									log.Printf("[WARN] plugin failed to shut down gracefully: %v", err)
								}
								log.Printf("[INFO] shut down plugin %s", in.Name)
							}
						}()
						if idx == 0 {
							status("Waiting")
							pluginLock.Lock()
							if c.Err() != nil {
								pluginLock.Unlock()
								return nil, fmt.Errorf("cancelled")
							}
							log.Printf("[INFO] launching plugin %s", in.Name)
							p, err := s.plugins.LaunchPlugin(in.Name, ctx)
							if err != nil {
								log.Printf("[ERR]  failed to launch plugin %s", in.Name)
								dead = true
								pluginLock.Unlock()
								return nil, fmt.Errorf("plugin failed to launch: %w", err)
							}
							lockNewPlugin(p)
						} else if c.Err() != nil || dead {
							return nil, fmt.Errorf("cancelled")
						}
						status(fmt.Sprintf("Running hook %s", in.Hook))
						return currentPlugin.RunHook(in.Hook, in.Input, s.imageRepository)
					}),
				)
			}
			return map[string]any{"success": true}, nil
		}),
	)
}
