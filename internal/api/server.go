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
	"k8s.io/client-go/rest"
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
	cfg        Config
	k8s        client.Client
	restConfig *rest.Config
	plugins    *plugin.Manager
	jobs       *JobStore
	sessions   *sessionStore
	events     *EventBus
	db         *gorm.DB
	uiFS       http.FileSystem
	http       *http.Server
}

// New creates a new API Server.
func New(cfg Config, k8s client.Client, restConfig *rest.Config, plugins *plugin.Manager) *Server {
	cfg.defaults()

	s := &Server{
		cfg:        cfg,
		k8s:        k8s,
		restConfig: restConfig,
		plugins:    plugins,
		jobs:       newJobStore(cfg.DB),
		sessions:   newSessionStore(),
		events:     newEventBus(),
		db:         cfg.DB,
		uiFS:       cfg.UIFS,
	}

	if s.uiFS == nil {
		slog.Warn("UIFS not provided, using fallback ./ui/build")
		s.uiFS = http.Dir("./ui/build")
	}

	s.http = &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// Start begins listening for requests. Blocks until the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	// Build routes here, after SetUI() has been called by main.go
	s.http.Handler = s.routes()

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

	// ─── ROTAS DA API ──────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// Rotas públicas
		r.Get("/system/health", s.handleHealth)
		r.Get("/system/version", s.handleVersion)
		r.Get("/cluster", s.handleCluster)

		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/logout", s.handleLogout)
		r.Get("/auth/me", s.handleMe)

		// Rotas com autenticação
		r.Group(func(r chi.Router) {
			r.Use(middlewareOIDC(s.cfg.OIDC))

			r.Get("/components", s.handleListComponents)
			r.Post("/components", s.handleCreateComponent)
			r.Get("/components/{namespace}/{name}", s.handleGetComponent)
			r.Delete("/components/{namespace}/{name}", s.handleDeleteComponent)
			r.Post("/components/{namespace}/{name}/exec", s.handleExecComponent)
			r.Get("/components/{namespace}/{name}/logs", s.handleComponentLogs)

			r.Get("/jobs", s.handleListJobs)
			r.Get("/jobs/{id}", s.handleGetJob)
			r.Delete("/jobs/{id}", s.handleCancelJob)

			r.Get("/plugins", s.handleListPlugins)
			r.Get("/plugins/{name}", s.handleGetPlugin)

			r.Get("/events", s.handleEvents)
		})
	})

	// ─── UI ──────────────────────────────────────────────────────
	r.Handle("/_app/*", http.FileServer(s.uiFS))
	r.Handle("/favicon.ico", http.FileServer(s.uiFS))
	r.Handle("/robots.txt", http.FileServer(s.uiFS))

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api/") {
			http.NotFound(w, req)
			return
		}
		// SPA: paths without a file extension are client-side routes — serve index.html
		if !strings.Contains(path.Base(req.URL.Path), ".") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			req.URL.Path = "/"
		}
		http.FileServer(s.uiFS).ServeHTTP(w, req)
	}))

	return r
}
