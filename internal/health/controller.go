// Package health implements the CosmoComponent health controller.
package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
	"github.com/galileostd/cosmonaut/internal/plugin"
	"github.com/galileostd/cosmonaut/internal/registry"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
	healthCheckTimeout         = 10 * time.Second
	conditionTypeHealthy       = "Healthy"
)

// EventPublisher is the interface the health controller uses to publish events.
type EventPublisher interface {
	PublishHealthChanged(namespace, name, health, message string)
}

// Controller reconciles CosmoComponent objects.
type Controller struct {
	client.Client
	Plugins  *plugin.Manager
	EventBus EventPublisher
}

// SetupWithManager registers the controller with the controller-runtime manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registry.CosmoComponent{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		}).
		Complete(c)
}

// requeueAfter returns the requeue interval for a component.
func requeueAfter(component *registry.CosmoComponent) time.Duration {
	if component.Spec.HealthCheckIntervalSeconds > 0 {
		return time.Duration(component.Spec.HealthCheckIntervalSeconds) * time.Second
	}
	return defaultHealthCheckInterval
}

// Reconcile is called whenever a CosmoComponent spec is created or updated,
// and periodically via RequeueAfter for health checks.
func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := slog.With("component", req.NamespacedName)

	var component registry.CosmoComponent
	if err := c.Get(ctx, req.NamespacedName, &component); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("fetching CosmoComponent: %w", err)
	}

	interval := requeueAfter(&component)

	pluginClient := c.Plugins.Get(component.Spec.Plugin)
	if pluginClient == nil {
		log.Warn("no plugin registered for component", "plugin", component.Spec.Plugin)
		if err := c.setStatus(ctx, req.NamespacedName,
			pluginv1.HealthState_HEALTH_STATE_UNKNOWN,
			fmt.Sprintf("plugin %q is not registered in this control plane", component.Spec.Plugin),
			nil,
		); err != nil {
			return reconcile.Result{RequeueAfter: interval}, err
		}
		return reconcile.Result{RequeueAfter: interval}, nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	sdkComponent := &pluginv1.Component{
		Name:      component.Name,
		Namespace: component.Namespace,
		Endpoint:  component.Spec.Endpoint,
		Config:    component.Spec.Config,
	}

	resp, err := pluginClient.HealthCheck(checkCtx, sdkComponent)
	if err != nil {
		log.Error("health check error", "plugin", component.Spec.Plugin, "err", err)
		if setErr := c.setStatus(ctx, req.NamespacedName,
			pluginv1.HealthState_HEALTH_STATE_UNKNOWN,
			fmt.Sprintf("health check failed: %v", err),
			nil,
		); setErr != nil {
			return reconcile.Result{RequeueAfter: interval}, setErr
		}
		return reconcile.Result{RequeueAfter: interval}, nil
	}

	log.Info("health check complete",
		"plugin", component.Spec.Plugin,
		"state", healthStateString(resp.State),
		"message", resp.Message,
	)

	if c.EventBus != nil {
		c.EventBus.PublishHealthChanged(
			component.Namespace,
			component.Name,
			healthStateString(resp.State),
			resp.Message,
		)
	}

	var capabilities []string
	if resp.State == pluginv1.HealthState_HEALTH_STATE_HEALTHY {
		desc, descErr := pluginClient.Describe(checkCtx)
		if descErr == nil {
			for _, cap := range desc.Capabilities {
				capabilities = append(capabilities, cap.Type)
			}
		}
	}

	if err := c.setStatus(ctx, req.NamespacedName, resp.State, resp.Message, capabilities); err != nil {
		return reconcile.Result{RequeueAfter: interval}, err
	}

	return reconcile.Result{RequeueAfter: interval}, nil
}

// setStatus updates the CosmoComponent status with retry on conflict.
// Re-fetches the object before each attempt to ensure the resourceVersion is current.
func (c *Controller) setStatus(
	ctx context.Context,
	key types.NamespacedName,
	state pluginv1.HealthState,
	message string,
	capabilities []string,
) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current registry.CosmoComponent
		if err := c.Get(ctx, key, &current); err != nil {
			return fmt.Errorf("fetching CosmoComponent for status update: %w", err)
		}

		now := metav1.Now()
		current.Status.Health = healthStateString(state)
		current.Status.Message = message
		current.Status.LastChecked = &now
		current.Status.ObservedGeneration = current.Generation

		if capabilities != nil {
			current.Status.Capabilities = capabilities
		}

		healthy := state == pluginv1.HealthState_HEALTH_STATE_HEALTHY
		conditionStatus := metav1.ConditionTrue
		conditionReason := "HealthCheckPassed"
		if !healthy {
			conditionStatus = metav1.ConditionFalse
			conditionReason = "HealthCheckFailed"
		}

		setCondition(&current.Status.Conditions, metav1.Condition{
			Type:               conditionTypeHealthy,
			Status:             conditionStatus,
			Reason:             conditionReason,
			Message:            message,
			LastTransitionTime: now,
			ObservedGeneration: current.Generation,
		})

		if err := c.Status().Update(ctx, &current); err != nil {
			slog.Error("failed to update status",
				"component", key.Name,
				"health", current.Status.Health,
				"err", err,
			)
			return err
		}

		slog.Info("status updated",
			"component", key.Name,
			"health", current.Status.Health,
		)
		return nil
	})
}

func healthStateString(state pluginv1.HealthState) string {
	switch state {
	case pluginv1.HealthState_HEALTH_STATE_HEALTHY:
		return "healthy"
	case pluginv1.HealthState_HEALTH_STATE_DEGRADED:
		return "degraded"
	case pluginv1.HealthState_HEALTH_STATE_UNHEALTHY:
		return "unhealthy"
	default:
		return "unknown"
	}
}

func setCondition(conditions *[]metav1.Condition, newCondition metav1.Condition) {
	for i, existing := range *conditions {
		if existing.Type == newCondition.Type {
			if existing.Status == newCondition.Status {
				newCondition.LastTransitionTime = existing.LastTransitionTime
			}
			(*conditions)[i] = newCondition
			return
		}
	}
	*conditions = append(*conditions, newCondition)
}
