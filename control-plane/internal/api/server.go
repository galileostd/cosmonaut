package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config holds the API server configuration.
type Config struct {
	// Addr is the address to listen on. Default: ":8080"
	Addr string

	// OIDC holds the authentication configuration.
	OIDC OIDCConfig

	// CORSAllowedOrigins is a comma-separated list of allowed origins.
	// Use "*" for development only.
	CORSAllowedOrigins string

	// RequestTimeout is the maximum duration for a request. Default: 60s.
	RequestTimeout time.Duration
}

func (c *Config) defaults() {
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 60 * time.Second
	}
	if c.CORSAllowedOrigins == "" {
		c.CORSAllowedOrigins = "*"
	}
}

// Server is the Cosmonaut API server.
type Server struct {
	cfg    Config
	k8s    client.Client
	jobs   *JobStore
	events *EventBus
	http   *http.Server
}

// New creates a new API Server.
func New(cfg Config, k8s client.Client) *Server {
	cfg.defaults()

	s := &Server{
		cfg:    cfg,
		k8s:    k8s,
		jobs:   newJobStore(),
		events: newEventBus(),
	}

	s.http = &http.Server{
		Addr:         cfg.Addr,
		Handler:      s.routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// Start begins listening for requests. Blocks until the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("API server listening", "addr", s.cfg.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("API server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		slog.Info("API server shutting down")
		return s.http.Shutdown(shutdownCtx)
	}
}

// EventBus returns the event bus so the health controller can publish events.
func (s *Server) EventBus() *EventBus {
	return s.events
}

// routes builds and returns the chi router with all middleware and handlers.
func (s *Server) routes() http.Handler {
	r := chi.NewRouter()

	// ── global middleware ──────────────────────────────────────────────────────
	r.Use(middlewareRequestID)
	r.Use(middlewareRecover)
	r.Use(middlewareLogger)
	r.Use(middlewareCORS(s.cfg.CORSAllowedOrigins))
	r.Use(middlewareTimeout(s.cfg.RequestTimeout))
	r.Use(chimiddleware.StripSlashes)

	// ── unauthenticated routes ─────────────────────────────────────────────────
	r.Get("/api/v1/system/health", s.handleHealth)
	r.Get("/api/v1/system/version", s.handleVersion)

	// ── authenticated routes ───────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middlewareOIDC(s.cfg.OIDC))

		// components
		r.Get("/api/v1/components", s.handleListComponents)
		r.Post("/api/v1/components", s.handleCreateComponent)
		r.Get("/api/v1/components/{namespace}/{name}", s.handleGetComponent)
		r.Delete("/api/v1/components/{namespace}/{name}", s.handleDeleteComponent)
		r.Post("/api/v1/components/{namespace}/{name}/exec", s.handleExecComponent)

		// jobs
		r.Get("/api/v1/jobs", s.handleListJobs)
		r.Get("/api/v1/jobs/{id}", s.handleGetJob)
		r.Delete("/api/v1/jobs/{id}", s.handleCancelJob)

		// plugins
		r.Get("/api/v1/plugins", s.handleListPlugins)
		r.Get("/api/v1/plugins/{name}", s.handleGetPlugin)

		// events (WebSocket)
		r.Get("/api/v1/events", s.handleEvents)
	})

	return r
}

// NeedLeaderElection implements the LeaderElectionRunnable interface.
// The API server runs on all replicas, not just the leader.
func (s *Server) NeedLeaderElection() bool {
	return false
}
