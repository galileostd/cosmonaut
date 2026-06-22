package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// decodeJSON decodes the request body into v.
// Returns a Problem if decoding fails.
func decodeJSON(r *http.Request, v any) *Problem {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return problemBadRequest(r, "invalid JSON body: "+err.Error())
	}
	return nil
}

// Pagination holds pagination parameters parsed from query string.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// parsePagination parses limit and offset from query string.
// Defaults: limit=50, offset=0. Max limit: 500.
func parsePagination(r *http.Request) Pagination {
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 500 {
		limit = 500
	}
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	return Pagination{Limit: limit, Offset: offset}
}

// PagedResponse wraps a list response with pagination metadata.
type PagedResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func newPagedResponse[T any](items []T, total int, p Pagination) PagedResponse[T] {
	if items == nil {
		items = []T{}
	}
	return PagedResponse[T]{
		Items:  items,
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
	}
}

// queryInt reads an integer query parameter, returning the default if missing or invalid.
func queryInt(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	var v int
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
		return defaultVal
	}
	return v
}
