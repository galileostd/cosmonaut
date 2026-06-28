package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
	"github.com/galileostd/cosmonaut/internal/registry"
	"github.com/go-chi/chi/v5"
)

// ── Response types ───────────────────────────────────────────────────────────

type jobResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Plugin      string            `json:"plugin"`
	PluginJobID string            `json:"plugin_job_id,omitempty"`
	Action      string            `json:"action"`
	Status      string            `json:"status"`
	ErrorMsg    string            `json:"error_msg,omitempty"`
	Result      map[string]string `json:"result,omitempty"`
	SubmitTime  string            `json:"submit_time"`
	StartTime   string            `json:"start_time,omitempty"`
	EndTime     string            `json:"end_time,omitempty"`
	Duration    string            `json:"duration,omitempty"`
	CreatedAt   string            `json:"created_at"`
}

type jobsResponse struct {
	Jobs  []*pluginv1.GetJobResponse `json:"jobs"`
	Total int32                      `json:"total"`
}

func toJobResponse(j *Job) jobResponse {
	r := jobResponse{
		ID:          j.ID,
		Name:        j.Name,
		Plugin:      j.Plugin,
		PluginJobID: j.PluginJobID,
		Action:      j.Action,
		Status:      string(j.Status),
		ErrorMsg:    j.ErrorMsg,
		Result:      j.Result,
		SubmitTime:  j.SubmitTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:    j.Duration,
		CreatedAt:   j.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if j.StartTime != nil {
		r.StartTime = j.StartTime.UTC().Format("2006-01-02T15:04:05Z")
	}
	if j.EndTime != nil {
		r.EndTime = j.EndTime.UTC().Format("2006-01-02T15:04:05Z")
	}
	return r
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// handleListJobs returns all jobs across all workload plugins.
// Merges jobs submitted via Cosmonaut with jobs discovered directly
// from plugins (Airflow, kubectl, scripts, etc). Deduplicates by job_id.
//
// GET /api/v1/jobs
//
//	?component=spark
//	?state=running|succeeded|failed|pending
//	?job_name=my-pipeline
//	?job_group=etl-diaria
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	stateFilter := parseJobState(q.Get("state"))
	jobName := q.Get("job_name")
	jobGroup := q.Get("job_group")
	componentFilter := q.Get("component")

	var componentList registry.CosmoComponentList
	if err := s.k8s.List(ctx, &componentList); err != nil {
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("listing components: %v", err)))
		return
	}

	seen := make(map[string]bool)
	var allJobs []*pluginv1.GetJobResponse

	for _, component := range componentList.Items {
		if componentFilter != "" && component.Name != componentFilter {
			continue
		}
		if component.Spec.Type != "processing" && component.Spec.Type != "orchestration" {
			continue
		}
		if !hasCapability(component.Status.Capabilities, "list-jobs") {
			continue
		}

		pluginClient := s.plugins.Get(component.Spec.Plugin)
		if pluginClient == nil {
			continue
		}

		sdkComponent := &pluginv1.Component{
			Name:      component.Name,
			Namespace: component.Namespace,
			Endpoint:  component.Spec.Endpoint,
			Config:    component.Spec.Config,
		}

		resp, err := pluginClient.ListJobs(ctx, sdkComponent, stateFilter, jobName, jobGroup, 0, 0)
		if err != nil {
			slog.Warn("listJobs: plugin error",
				"plugin", component.Spec.Plugin,
				"component", component.Name,
				"err", err,
			)
			continue
		}

		for _, job := range resp.Jobs {
			if seen[job.JobId] {
				continue
			}
			seen[job.JobId] = true

			if job.Details == nil {
				job.Details = make(map[string]string)
			}
			job.Details["_component"] = component.Name
			job.Details["_plugin"] = component.Spec.Plugin

			allJobs = append(allJobs, job)
		}
	}

	if allJobs == nil {
		allJobs = []*pluginv1.GetJobResponse{}
	}

	writeJSON(w, http.StatusOK, jobsResponse{
		Jobs:  allJobs,
		Total: int32(len(allJobs)),
	})
}

// handleGetJob returns a single job by ID.
// GET /api/v1/jobs/:id
func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	job, ok := s.jobs.Get(id)
	if !ok {
		writeProblem(w, r, problemNotFound(r, fmt.Sprintf("job '%s' not found", id)))
		return
	}

	writeJSON(w, http.StatusOK, toJobResponse(job))
}

// handleCancelJob cancels a running or pending job.
// DELETE /api/v1/jobs/:id
func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, ok := s.jobs.Get(id)
	if !ok {
		writeProblem(w, r, problemNotFound(r, fmt.Sprintf("job '%s' not found", id)))
		return
	}

	if canceled := s.jobs.Cancel(id); !canceled {
		writeProblem(w, r, problemConflict(r, fmt.Sprintf(
			"job '%s' cannot be canceled — it is already in a terminal state", id,
		)))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func parseJobState(s string) pluginv1.JobState {
	switch strings.ToLower(s) {
	case "running":
		return pluginv1.JobState_JOB_STATE_RUNNING
	case "succeeded", "completed":
		return pluginv1.JobState_JOB_STATE_SUCCEEDED
	case "failed":
		return pluginv1.JobState_JOB_STATE_FAILED
	case "pending":
		return pluginv1.JobState_JOB_STATE_PENDING
	case "canceled":
		return pluginv1.JobState_JOB_STATE_CANCELED
	default:
		return pluginv1.JobState_JOB_STATE_UNSPECIFIED
	}
}

func hasCapability(caps []string, target string) bool {
	for _, c := range caps {
		if c == target {
			return true
		}
	}
	return false
}
