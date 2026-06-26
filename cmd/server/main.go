package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/galileostd/cosmonaut/internal/api"
	"github.com/galileostd/cosmonaut/internal/db"
	"github.com/galileostd/cosmonaut/internal/health"
	"github.com/galileostd/cosmonaut/internal/plugin"
	"github.com/galileostd/cosmonaut/internal/registry"
	"github.com/galileostd/cosmonaut/internal/ui"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = registry.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(envOr("COSMONAUT_DEV", "") == "true")))

	slog.Info("starting cosmonaut control plane", "version", api.BuildVersion)

	dbCfg := &db.Config{
		Driver:   envOr("COSMONAUT_DB_DRIVER", "cockroachdb"),
		Host:     envOr("COSMONAUT_DB_HOST", "localhost"),
		Port:     envOr("COSMONAUT_DB_PORT", "26257"),
		User:     envOr("COSMONAUT_DB_USER", "root"),
		Password: envOr("COSMONAUT_DB_PASSWORD", ""),
		Database: envOr("COSMONAUT_DB_NAME", "cosmonaut"),
		SSLMode:  envOr("COSMONAUT_DB_SSL_MODE", "disable"),
		Debug:    envOr("COSMONAUT_DB_DEBUG", "false") == "true",
	}

	dbConn, err := db.Connect(dbCfg)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(dbConn); err != nil {
		slog.Error("failed to run database migrations", "err", err)
		os.Exit(1)
	}

	slog.Info("database connected", "driver", dbCfg.Driver, "host", dbCfg.Host)

	// -----------------------------------------------------------------
	// UI
	// -----------------------------------------------------------------

	uiFS, err := ui.GetFS()
	if err != nil {
		slog.Error("failed to load embedded UI", "err", err)
		os.Exit(1)
	}

	slog.Info("embedded UI loaded")

	// -----------------------------------------------------------------
	// Kubernetes Config
	// -----------------------------------------------------------------

	k8sConfig := ctrl.GetConfigOrDie()

	// -----------------------------------------------------------------
	// Manager
	// -----------------------------------------------------------------

	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: envOr("COSMONAUT_METRICS_ADDR", ":9090"),
		},
		HealthProbeBindAddress: envOr("COSMONAUT_HEALTH_ADDR", ":8081"),
		LeaderElection:         envOr("COSMONAUT_LEADER_ELECT", "false") == "true",
		LeaderElectionID:       "cosmonaut.galileostd.io",
	})
	if err != nil {
		slog.Error("unable to create manager", "err", err)
		os.Exit(1)
	}

	pluginManager := plugin.NewManager()

	apiServer := api.New(api.Config{
		Addr: envOr("COSMONAUT_API_ADDR", ":8080"),
		OIDC: api.OIDCConfig{
			IssuerURL: envOr("COSMONAUT_OIDC_ISSUER", ""),
			ClientID:  envOr("COSMONAUT_OIDC_CLIENT_ID", "cosmonaut"),
			Enabled:   envOr("COSMONAUT_AUTH_ENABLED", "false") == "true",
		},
		CORSAllowedOrigins: envOr("COSMONAUT_CORS_ORIGINS", "*"),
		RequestTimeout:     60 * time.Second,
		DB:                 dbConn,
		UIFS:               http.FS(uiFS),
	}, mgr.GetClient(), k8sConfig, pluginManager)

	if err := (&health.Controller{
		Client:   mgr.GetClient(),
		Plugins:  pluginManager,
		EventBus: apiServer.EventBus(),
	}).SetupWithManager(mgr); err != nil {
		slog.Error("unable to set up health controller", "err", err)
		os.Exit(1)
	}

	if err := (&plugin.DiscoveryController{
		Client:  mgr.GetClient(),
		Manager: pluginManager,
	}).SetupWithManager(mgr); err != nil {
		slog.Error("unable to set up plugin discovery controller", "err", err)
		os.Exit(1)
	}

	if err := mgr.Add(apiServer); err != nil {
		slog.Error("unable to add API server to manager", "err", err)
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		slog.Error("unable to set up healthz", "err", err)
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		slog.Error("unable to set up readyz", "err", err)
		os.Exit(1)
	}

	slog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		slog.Error("manager exited with error", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}