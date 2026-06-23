package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const errorBaseURL = "https://cosmonaut.galileostd.io/errors"

// Problem represents an RFC 7807 Problem Details object.
type Problem struct {
	// Type is a URI reference identifying the problem type.
	Type string `json:"type"`

	// Title is a short, human-readable summary of the problem.
	Title string `json:"title"`

	// Status is the HTTP status code.
	Status int `json:"status"`

	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`

	// Instance is a URI reference identifying the specific occurrence.
	Instance string `json:"instance,omitempty"`

	// Extensions holds additional problem-specific fields.
	Extensions map[string]any `json:"-"`
}

func (p *Problem) Error() string {
	return fmt.Sprintf("%s: %s", p.Title, p.Detail)
}

// MarshalJSON merges extensions into the top-level object.
func (p *Problem) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":   p.Type,
		"title":  p.Title,
		"status": p.Status,
	}
	if p.Detail != "" {
		m["detail"] = p.Detail
	}
	if p.Instance != "" {
		m["instance"] = p.Instance
	}
	for k, v := range p.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

// writeProblem writes a Problem as an RFC 7807 JSON response.
func writeProblem(w http.ResponseWriter, r *http.Request, p *Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}

// ── pre-built problem constructors ───────────────────────────────────────────

func problemNotFound(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/not-found",
		Title:    "Not Found",
		Status:   http.StatusNotFound,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemBadRequest(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/bad-request",
		Title:    "Bad Request",
		Status:   http.StatusBadRequest,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemUnauthorized(r *http.Request) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/unauthorized",
		Title:    "Unauthorized",
		Status:   http.StatusUnauthorized,
		Detail:   "valid authentication credentials are required",
		Instance: r.URL.Path,
	}
}

func problemForbidden(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/forbidden",
		Title:    "Forbidden",
		Status:   http.StatusForbidden,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemConflict(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/conflict",
		Title:    "Conflict",
		Status:   http.StatusConflict,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemUnprocessable(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/unprocessable",
		Title:    "Unprocessable Entity",
		Status:   http.StatusUnprocessableEntity,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemInternal(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/internal",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}

func problemServiceUnavailable(r *http.Request, detail string) *Problem {
	return &Problem{
		Type:     errorBaseURL + "/service-unavailable",
		Title:    "Service Unavailable",
		Status:   http.StatusServiceUnavailable,
		Detail:   detail,
		Instance: r.URL.Path,
	}
}
