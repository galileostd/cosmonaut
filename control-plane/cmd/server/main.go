// Cosmonaut Control Plane
// Open-source control plane for on-premises data platforms.
package main

import (
	"log/slog"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/galileostd/cosmonaut/control-plane/internal/api"
	"github.com/galileostd/cosmonaut/control-plane/internal/health"
	"github.com/galileostd/cosmonaut/control-plane/internal/plugin"
	"github.com/galileostd/cosmonaut/control-plane/internal/registry"
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

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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

	// plugin manager — maintains gRPC connections to discovered plugins
	pluginManager := plugin.NewManager()

	// API server
	apiServer := api.New(api.Config{
		Addr: envOr("COSMONAUT_API_ADDR", ":8080"),
		OIDC: api.OIDCConfig{
			IssuerURL: envOr("COSMONAUT_OIDC_ISSUER", ""),
			ClientID:  envOr("COSMONAUT_OIDC_CLIENT_ID", "cosmonaut"),
			Enabled:   envOr("COSMONAUT_AUTH_ENABLED", "false") == "true",
		},
		CORSAllowedOrigins: envOr("COSMONAUT_CORS_ORIGINS", "*"),
		RequestTimeout:     60 * time.Second,
	}, mgr.GetClient(), pluginManager)

	// health controller
	if err := (&health.Controller{
		Client:   mgr.GetClient(),
		Plugins:  pluginManager,
		EventBus: apiServer.EventBus(),
	}).SetupWithManager(mgr); err != nil {
		slog.Error("unable to set up health controller", "err", err)
		os.Exit(1)
	}

	// plugin discovery controller
	if err := (&plugin.DiscoveryController{
		Client:  mgr.GetClient(),
		Manager: pluginManager,
	}).SetupWithManager(mgr); err != nil {
		slog.Error("unable to set up plugin discovery controller", "err", err)
		os.Exit(1)
	}

	// API server as a runnable
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
