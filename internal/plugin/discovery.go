package plugin

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// LabelPlugin is the label that marks a K8s Service as a Cosmonaut plugin.
	// Any Service with this label set to "true" will be discovered automatically.
	LabelPlugin = "cosmonaut.galileostd.io/plugin"

	// LabelPluginName optionally overrides the plugin name.
	// If not set, the Service name is used.
	LabelPluginName = "cosmonaut.galileostd.io/plugin-name"

	// AnnotationPluginPort optionally overrides the gRPC port.
	// Default: 50051
	AnnotationPluginPort = "cosmonaut.galileostd.io/plugin-port"

	defaultPluginPort = "50051"
)

// DiscoveryController watches Kubernetes Services labeled as Cosmonaut plugins
// and registers/unregisters them in the plugin Manager automatically.
//
// Any Service with label cosmonaut.galileostd.io/plugin=true is treated as a plugin.
// The gRPC endpoint is derived from:
//
//	{service-name}.{namespace}.svc.cluster.local:{port}
type DiscoveryController struct {
	client.Client
	Manager *Manager
}

// SetupWithManager registers the discovery controller with controller-runtime.
func (d *DiscoveryController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return isPluginService(e.Object.GetLabels())
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isPluginService(e.ObjectNew.GetLabels()) ||
					isPluginService(e.ObjectOld.GetLabels())
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return isPluginService(e.Object.GetLabels())
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return isPluginService(e.Object.GetLabels())
			},
		}).
		Complete(d)
}

// Reconcile handles Service create/update/delete events.
func (d *DiscoveryController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := slog.With("service", req.NamespacedName)

	var svc corev1.Service
	if err := d.Get(ctx, req.NamespacedName, &svc); err != nil {
		if errors.IsNotFound(err) {
			// service was deleted — unregister by service name
			// we use the request name as a best-effort plugin name
			d.Manager.Unregister(req.Name)
			d.Manager.UnregisterByService(req.Name, req.Namespace)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("fetching Service: %w", err)
	}

	// if the label was removed, unregister
	if !isPluginService(svc.Labels) {
		d.Manager.Unregister(pluginName(svc))
		return reconcile.Result{}, nil
	}

	// build the gRPC endpoint
	port := svc.Annotations[AnnotationPluginPort]
	if port == "" {
		port = defaultPluginPort
	}

	endpoint := fmt.Sprintf("%s.%s.svc.cluster.local:%s",
		svc.Name, svc.Namespace, port,
	)

	info := PluginInfo{
		Name:        pluginName(svc),
		Endpoint:    endpoint,
		Namespace:   svc.Namespace,
		ServiceName: svc.Name,
	}

	if err := d.Manager.Register(info); err != nil {
		log.Error("failed to register plugin", "name", info.Name, "endpoint", endpoint, "err", err)
		// don't requeue — registration will be retried on next Service update
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

// isPluginService returns true if the labels contain the plugin marker.
func isPluginService(labels map[string]string) bool {
	return labels[LabelPlugin] == "true"
}

// pluginName derives the plugin name from a Service.
// Uses the cosmonaut.galileostd.io/plugin-name label if set, otherwise the Service name.
func pluginName(svc corev1.Service) string {
	if name, ok := svc.Labels[LabelPluginName]; ok && name != "" {
		return name
	}
	return svc.Name
}

// ListPluginServices returns all Services in the cluster labeled as plugins.
// Useful for initial bootstrap and debugging.
func ListPluginServices(ctx context.Context, c client.Client) ([]corev1.Service, error) {
	var list corev1.ServiceList
	if err := c.List(ctx, &list, client.MatchingLabels{
		LabelPlugin: "true",
	}); err != nil {
		return nil, fmt.Errorf("listing plugin services: %w", err)
	}
	return list.Items, nil
}
