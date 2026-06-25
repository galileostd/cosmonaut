package api

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/galileostd/cosmonaut/internal/plugin"
)

// Config holds the API server configuration.
type Config struct {
	Addr               string
	OIDC               OIDCConfig
	CORSAllowedOrigins string
	RequestTimeout     time.Duration
	DB                 *gorm.DB
	UIFS               http.FileSystem
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
	cfg     Config
	k8s     client.Client
	plugins *plugin.Manager
	jobs    *JobStore
	events  *EventBus
	db      *gorm.DB
	uiFS    http.FileSystem
	http    *http.Server
}

// New creates a new API Server.
func New(cfg Config, k8s client.Client, plugins *plugin.Manager) *Server {
	cfg.defaults()

	s := &Server{
		cfg:     cfg,
		k8s:     k8s,
		plugins: plugins,
		jobs:    newJobStore(),
		events:  newEventBus(),
		db:      cfg.DB,
		uiFS:    cfg.UIFS,
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

// NeedLeaderElection implements the LeaderElectionRunnable interface.
func (s *Server) NeedLeaderElection() bool {
	return false
}

// SetUI defines the filesystem to serve static UI files.
func (s *Server) SetUI(fsys fs.FS) {
	s.uiFS = http.FS(fsys)
}

// routes builds and returns the chi router.
func (s *Server) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewareRequestID)
	r.Use(middlewareRecover)
	r.Use(middlewareLogger)
	r.Use(middlewareCORS(s.cfg.CORSAllowedOrigins))
	r.Use(middlewareTimeout(s.cfg.RequestTimeout))
	r.Use(chimiddleware.StripSlashes)

	// unauthenticated
	r.Get("/api/v1/system/health", s.handleHealth)
	r.Get("/api/v1/system/version", s.handleVersion)

	// authenticated
	r.Group(func(r chi.Router) {
		r.Use(middlewareOIDC(s.cfg.OIDC))

		r.Get("/api/v1/components", s.handleListComponents)
		r.Post("/api/v1/components", s.handleCreateComponent)
		r.Get("/api/v1/components/{namespace}/{name}", s.handleGetComponent)
		r.Delete("/api/v1/components/{namespace}/{name}", s.handleDeleteComponent)
		r.Post("/api/v1/components/{namespace}/{name}/exec", s.handleExecComponent)

		r.Get("/api/v1/jobs", s.handleListJobs)
		r.Get("/api/v1/jobs/{id}", s.handleGetJob)
		r.Delete("/api/v1/jobs/{id}", s.handleCancelJob)

		r.Get("/api/v1/plugins", s.handleListPlugins)
		r.Get("/api/v1/plugins/{name}", s.handleGetPlugin)

		r.Get("/api/v1/events", s.handleEvents)
	})

	// UI - serve static files
	if s.uiFS != nil {
		r.Get("/*", s.handleUI)
	}

	return r
}

// handleUI serves the static UI files.
func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	if requestPath == "" || requestPath == "/" {
		requestPath = "/index.html"
	}

	file, err := s.uiFS.Open(path.Clean(requestPath))
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "not found") {
			indexFile, err := s.uiFS.Open("/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer indexFile.Close()

			stat, _ := indexFile.Stat()
			http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
			return
		}
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if strings.Contains(requestPath, ".") {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}

	http.ServeContent(w, r, path.Base(requestPath), stat.ModTime(), file)
}