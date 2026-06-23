package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type jobResponse struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Error     string     `json:"error,omitempty"`
	Result    any        `json:"result,omitempty"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

func toJobResponse(j *Job) jobResponse {
	resp := jobResponse{
		ID:        j.ID,
		Status:    string(j.Status),
		Error:     j.Error,
		CreatedAt: j.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt: j.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if j.Result != nil {
		resp.Result = j.Result
	}
	return resp
}

// handleListJobs returns a paginated list of async jobs.
// GET /api/v1/jobs
// Query params: limit, offset, status (optional filter: pending, running, done, failed, canceled)
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	p := parsePagination(r)
	statusFilter := r.URL.Query().Get("status")

	all := s.jobs.List(statusFilter)
	total := len(all)

	start := p.Offset
	if start > total {
		start = total
	}
	end := start + p.Limit
	if end > total {
		end = total
	}

	responses := make([]jobResponse, len(all[start:end]))
	for i, j := range all[start:end] {
		responses[i] = toJobResponse(j)
	}

	writeJSON(w, http.StatusOK, newPagedResponse(responses, total, p))
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

// handleCancelJob attempts to cancel a running or pending job.
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
