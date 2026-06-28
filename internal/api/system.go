package api

import (
	"net/http"
	"runtime"
	"time"
)

var startTime = time.Now()

// build info — set via ldflags at build time:
// go build -ldflags "-X api.BuildVersion=v0.1.0 -X api.BuildCommit=abc123 -X api.BuildDate=2026-01-01"
var (
	BuildVersion = "v0.1.0"
	BuildCommit  = "dev"
	BuildDate    = "unknown"
)

type healthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

type versionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

// handleHealth returns the health status of the control plane.
// GET /api/v1/system/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Uptime: time.Since(startTime).Round(time.Second).String(),
	})
}

// handleVersion returns build information about the control plane.
// GET /api/v1/system/version
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{
		Version:   BuildVersion,
		Commit:    BuildCommit,
		Date:      BuildDate,
		GoVersion: runtime.Version(),
	})
}
