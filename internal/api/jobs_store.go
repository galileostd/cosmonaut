package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
	dbmodel "github.com/galileostd/cosmonaut/internal/db"
)

// JobStatus represents the lifecycle state of an async job.
type JobStatus string

const (
	JobStatusQueued   JobStatus = "queued"
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusDone     JobStatus = "done"
	JobStatusFailed   JobStatus = "failed"
	JobStatusCanceled JobStatus = "canceled"
)

// Job is the runtime representation. Mirrors db.Job with an added cancel func.
type Job struct {
	ID          string
	ComponentID string
	Name        string
	Plugin      string
	PluginJobID string
	Action      string
	Status      JobStatus
	SubmitTime  time.Time
	StartTime   *time.Time
	EndTime     *time.Time
	Duration    string
	ErrorMsg    string
	Result      map[string]string
	Logs        string
	CreatedAt   time.Time

	cancel context.CancelFunc // runtime only — never persisted
}

// JobFunc is the function executed by a job.
type JobFunc func(ctx context.Context) (*pluginv1.ExecuteResponse, error)

// JobStore manages async jobs with in-memory execution and DB persistence.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	db   *gorm.DB
}

func newJobStore(db *gorm.DB) *JobStore {
	js := &JobStore{
		jobs: make(map[string]*Job),
		db:   db,
	}
	go js.pruneLoop()
	return js
}

// Submit creates a job, persists it, and starts async execution.
func (js *JobStore) Submit(componentID, name, plugin, action string, fn JobFunc) *Job {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	now := time.Now()

	job := &Job{
		ID:          uuid.New().String(),
		ComponentID: componentID,
		Name:        name,
		Plugin:      plugin,
		Action:      action,
		Status:      JobStatusQueued,
		SubmitTime:  now,
		CreatedAt:   now,
		cancel:      cancel,
	}

	js.mu.Lock()
	js.jobs[job.ID] = job
	js.mu.Unlock()

	js.persist(job)
	go js.run(ctx, job, fn)
	return job
}

// Get returns a job by ID. Falls back to DB if not in memory.
func (js *JobStore) Get(id string) (*Job, bool) {
	js.mu.RLock()
	j, ok := js.jobs[id]
	js.mu.RUnlock()

	if !ok && js.db != nil {
		var dbJob dbmodel.Job
		if err := js.db.Where("id = ?", id).First(&dbJob).Error; err == nil {
			j = dbToMem(&dbJob)
			ok = true
		}
	}
	return j, ok
}

// List returns jobs matching filters.
func (js *JobStore) List(componentFilter, statusFilter string, limit, offset int) ([]*Job, int) {
	if js.db != nil {
		js.mu.RLock()
		memCount := len(js.jobs)
		js.mu.RUnlock()

		if memCount == 0 {
			return js.listFromDB(componentFilter, statusFilter, limit, offset)
		}
	}

	js.mu.RLock()
	defer js.mu.RUnlock()

	jobs := make([]*Job, 0, len(js.jobs))
	for _, j := range js.jobs {
		if componentFilter != "" && j.Plugin != componentFilter && j.Name != componentFilter {
			continue
		}
		if statusFilter == "" || string(j.Status) == statusFilter {
			jobs = append(jobs, j)
		}
	}

	total := len(jobs)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return jobs[start:end], total
}

// Cancel cancels a running or pending job.
func (js *JobStore) Cancel(id string) bool {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return false
	}
	if j.Status == JobStatusRunning || j.Status == JobStatusPending || j.Status == JobStatusQueued {
		j.cancel()
		j.Status = JobStatusCanceled
		now := time.Now()
		j.EndTime = &now
		js.persist(j)
		return true
	}
	return false
}

func (js *JobStore) run(ctx context.Context, job *Job, fn JobFunc) {
	now := time.Now()
	job.StartTime = &now
	js.updateStatus(job, JobStatusRunning, "")

	resp, err := fn(ctx)
	endTime := time.Now()
	job.EndTime = &endTime

	if job.StartTime != nil {
		job.Duration = endTime.Sub(*job.StartTime).Round(time.Millisecond).String()
	}

	if err != nil {
		slog.Error("job failed", "job_id", job.ID, "name", job.Name, "err", err)
		js.updateStatus(job, JobStatusFailed, err.Error())
		return
	}

	job.PluginJobID = resp.JobId
	job.Result = resp.Result

	switch resp.State {
	case pluginv1.JobState_JOB_STATE_FAILED:
		js.updateStatus(job, JobStatusFailed, resp.Message)
	case pluginv1.JobState_JOB_STATE_CANCELED:
		js.updateStatus(job, JobStatusCanceled, resp.Message)
	default:
		js.updateStatus(job, JobStatusDone, "")
	}
}

func (js *JobStore) updateStatus(job *Job, status JobStatus, errMsg string) {
	js.mu.Lock()
	job.Status = status
	if errMsg != "" {
		job.ErrorMsg = errMsg
	}
	js.mu.Unlock()

	js.persist(job)
}

func (js *JobStore) persist(j *Job) {
	if js.db == nil {
		return
	}

	resultJSON := ""
	if j.Result != nil {
		b, _ := json.Marshal(j.Result)
		resultJSON = string(b)
	}

	dbJob := dbmodel.Job{
		ID:          j.ID,
		ComponentID: j.ComponentID,
		Name:        j.Name,
		Plugin:      j.Plugin,
		PluginJobID: j.PluginJobID,
		Action:      j.Action,
		Status:      string(j.Status),
		SubmitTime:  j.SubmitTime,
		StartTime:   j.StartTime,
		EndTime:     j.EndTime,
		Duration:    j.Duration,
		ErrorMsg:    j.ErrorMsg,
		Result:      resultJSON,
		Logs:        j.Logs,
		CreatedAt:   j.CreatedAt,
	}

	if err := js.db.Save(&dbJob).Error; err != nil {
		slog.Error("failed to persist job", "job_id", j.ID, "err", err)
	}
}

func (js *JobStore) listFromDB(componentFilter, statusFilter string, limit, offset int) ([]*Job, int) {
	var total int64
	query := js.db.Model(&dbmodel.Job{})

	if componentFilter != "" {
		query = query.Where("plugin = ? OR name = ?", componentFilter, componentFilter)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)

	var dbJobs []dbmodel.Job
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&dbJobs)

	jobs := make([]*Job, len(dbJobs))
	for i, dbJob := range dbJobs {
		jobs[i] = dbToMem(&dbJob)
	}

	return jobs, int(total)
}

func (js *JobStore) pruneLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		js.prune()
	}
}

func (js *JobStore) prune() {
	cutoff := time.Now().Add(-24 * time.Hour)
	js.mu.Lock()
	for id, j := range js.jobs {
		terminal := j.Status == JobStatusDone ||
			j.Status == JobStatusFailed ||
			j.Status == JobStatusCanceled
		if terminal && j.CreatedAt.Before(cutoff) {
			delete(js.jobs, id)
		}
	}
	js.mu.Unlock()
}

func dbToMem(dbJob *dbmodel.Job) *Job {
	var result map[string]string
	if dbJob.Result != "" {
		json.Unmarshal([]byte(dbJob.Result), &result)
	}
	return &Job{
		ID:          dbJob.ID,
		ComponentID: dbJob.ComponentID,
		Name:        dbJob.Name,
		Plugin:      dbJob.Plugin,
		PluginJobID: dbJob.PluginJobID,
		Action:      dbJob.Action,
		Status:      JobStatus(dbJob.Status),
		SubmitTime:  dbJob.SubmitTime,
		StartTime:   dbJob.StartTime,
		EndTime:     dbJob.EndTime,
		Duration:    dbJob.Duration,
		ErrorMsg:    dbJob.ErrorMsg,
		Result:      result,
		Logs:        dbJob.Logs,
		CreatedAt:   dbJob.CreatedAt,
	}
}
