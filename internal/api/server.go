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
	sessions *sessionStore
	events  *EventBus
	db      *gorm.DB
	uiFS    http.FileSystem
	http    *http.Server
}

// New creates a new API Server.
func New(cfg Config, k8s client.Client, plugins *plugin.Manager) *Server {
	cfg.defaults()

	s := &Server{
		cfg:      cfg,
		k8s:      k8s,
		plugins:  plugins,
		jobs:     newJobStore(),
		sessions: newSessionStore(),
		events:   newEventBus(),
		db:       cfg.DB,
		uiFS:     cfg.UIFS,
	}

	if s.uiFS == nil {
		slog.Warn("UIFS not provided, using fallback ./ui/build")
		s.uiFS = http.Dir("./ui/build")
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
	slog.Info("SetUI called", "fsys_nil", fsys == nil)
	if fsys != nil {
		s.uiFS = http.FS(fsys)
	} else {
		slog.Warn("SetUI called with nil fsys, using fallback ./ui/build")
		s.uiFS = http.Dir("./ui/build")
	}
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

	r.Get("/api/v1/system/health", s.handleHealth)
	r.Get("/api/v1/system/version", s.handleVersion)

	r.Post("/api/v1/auth/login", s.handleLogin)
	r.Post("/api/v1/auth/logout", s.handleLogout)
	r.Get("/api/v1/auth/me", s.handleMe)

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

	// ─── UI ──────────────────────────────────────────────────────
	r.Handle("/_app/*", http.FileServer(s.uiFS))

	r.Handle("/favicon.ico", http.FileServer(s.uiFS))
	r.Handle("/robots.txt", http.FileServer(s.uiFS))

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = "/"
		http.FileServer(s.uiFS).ServeHTTP(w, req)
	}))

	return r
}

// handleUI serves the static UI files.
func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	slog.Info("handleUI called", "path", r.URL.Path, "uiFS_nil", s.uiFS == nil)

	if s.uiFS == nil {
		slog.Error("uiFS is nil, trying fallback to ./ui/build")
		s.uiFS = http.Dir("./ui/build")
	}

	// strip leading slash
	cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if cleanPath == "" || cleanPath == "." {
		cleanPath = "index.html"
	}

	slog.Info("Opening file", "cleanPath", cleanPath)

	// Try to open the file
	file, err := s.uiFS.Open(cleanPath)
	if err == nil {
		defer file.Close()
		stat, err := file.Stat()
		if err == nil {
			if strings.Contains(cleanPath, ".") {
				w.Header().Set("Cache-Control", "public, max-age=86400")
			}
			http.ServeContent(w, r, path.Base(cleanPath), stat.ModTime(), file)
			return
		}
	}

	// SPA fallback: serve index.html
	slog.Info("File not found or error, serving index.html", "err", err)
	indexFile, err := s.uiFS.Open("index.html")
	if err != nil {
		slog.Error("index.html not found", "err", err)
		http.NotFound(w, r)
		return
	}
	defer indexFile.Close()
	stat, _ := indexFile.Stat()
	http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
}