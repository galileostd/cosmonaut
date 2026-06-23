// Package api implements the Cosmonaut REST API.
// This file contains the in-memory async job store.
package api

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
)

// JobStatus represents the lifecycle state of an async job.
type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusDone     JobStatus = "done"
	JobStatusFailed   JobStatus = "failed"
	JobStatusCanceled JobStatus = "canceled"
)

// Job represents an async execution submitted via /exec.
type Job struct {
	ID          string
	Status      JobStatus
	PluginJobID string // job ID returned by the plugin
	Result      map[string]string
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	cancel      context.CancelFunc
}

// JobFunc is the function executed by a job.
type JobFunc func(ctx context.Context) (*pluginv1.ExecuteResponse, error)

// JobStore manages async jobs in memory.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func newJobStore() *JobStore {
	js := &JobStore{jobs: make(map[string]*Job)}
	go js.pruneLoop()
	return js
}

func (js *JobStore) Submit(fn JobFunc) *Job {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)

	job := &Job{
		ID:        uuid.New().String(),
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		cancel:    cancel,
	}

	js.mu.Lock()
	js.jobs[job.ID] = job
	js.mu.Unlock()

	go js.run(ctx, job, fn)
	return job
}

func (js *JobStore) Get(id string) (*Job, bool) {
	js.mu.RLock()
	defer js.mu.RUnlock()
	j, ok := js.jobs[id]
	return j, ok
}

func (js *JobStore) List(statusFilter string) []*Job {
	js.mu.RLock()
	defer js.mu.RUnlock()

	jobs := make([]*Job, 0, len(js.jobs))
	for _, j := range js.jobs {
		if statusFilter == "" || string(j.Status) == statusFilter {
			jobs = append(jobs, j)
		}
	}
	return jobs
}

func (js *JobStore) Cancel(id string) bool {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return false
	}
	if j.Status == JobStatusRunning || j.Status == JobStatusPending {
		j.cancel()
		j.Status = JobStatusCanceled
		j.UpdatedAt = time.Now()
		return true
	}
	return false
}

func (js *JobStore) run(ctx context.Context, job *Job, fn JobFunc) {
	js.updateStatus(job.ID, JobStatusRunning, "", nil, "")

	resp, err := fn(ctx)
	if err != nil {
		slog.Error("job failed", "job_id", job.ID, "err", err)
		js.updateStatus(job.ID, JobStatusFailed, "", nil, err.Error())
		return
	}

	switch resp.State {
	case pluginv1.JobState_JOB_STATE_FAILED:
		js.updateStatus(job.ID, JobStatusFailed, resp.JobId, resp.Result, resp.Message)
	case pluginv1.JobState_JOB_STATE_CANCELED:
		js.updateStatus(job.ID, JobStatusCanceled, resp.JobId, resp.Result, resp.Message)
	default:
		js.updateStatus(job.ID, JobStatusDone, resp.JobId, resp.Result, "")
	}
}

func (js *JobStore) updateStatus(id string, status JobStatus, pluginJobID string, result map[string]string, errMsg string) {
	js.mu.Lock()
	defer js.mu.Unlock()

	j, ok := js.jobs[id]
	if !ok {
		return
	}
	j.Status = status
	if pluginJobID != "" {
		j.PluginJobID = pluginJobID
	}
	j.Result = result
	j.Error = errMsg
	j.UpdatedAt = time.Now()
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
	defer js.mu.Unlock()

	for id, j := range js.jobs {
		terminal := j.Status == JobStatusDone ||
			j.Status == JobStatusFailed ||
			j.Status == JobStatusCanceled
		if terminal && j.UpdatedAt.Before(cutoff) {
			delete(js.jobs, id)
		}
	}
}
