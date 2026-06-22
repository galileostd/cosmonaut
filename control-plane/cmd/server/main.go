// Cosmonaut Control Plane
// Open-source control plane for on-premises data platforms.
package main

import (
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/galileostd/cosmonaut/control-plane/internal/health"
	"github.com/galileostd/cosmonaut/control-plane/internal/registry"

	// register native plugins
	_ "github.com/galileostd/cosmonaut/plugins/trino"
)

var scheme = runtime.NewScheme()

func init() {
	// register standard K8s types
	_ = clientgoscheme.AddToScheme(scheme)
	// register Cosmonaut CRD types
	_ = registry.AddToScheme(scheme)
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(envOr("COSMONAUT_DEV", "") == "true")))

	slog.Info("starting cosmonaut control plane", "version", "v0.1.0")

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

	// register the health controller
	if err := (&health.Controller{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		slog.Error("unable to set up health controller", "err", err)
		os.Exit(1)
	}

	// liveness and readiness probes
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
