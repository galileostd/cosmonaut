// Package health implements the CosmoComponent health controller.
package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
	"github.com/galileostd/cosmonaut/control-plane/internal/plugin"
	"github.com/galileostd/cosmonaut/control-plane/internal/registry"
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
		Complete(c)
}

// Reconcile is called whenever a CosmoComponent is created, updated, or deleted.
func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := slog.With("component", req.NamespacedName)

	var component registry.CosmoComponent
	if err := c.Get(ctx, req.NamespacedName, &component); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("fetching CosmoComponent: %w", err)
	}

	pluginClient := c.Plugins.Get(component.Spec.Plugin)
	if pluginClient == nil {
		log.Warn("no plugin registered for component", "plugin", component.Spec.Plugin)
		return c.setStatus(ctx, &component, pluginv1.HealthState_HEALTH_STATE_UNKNOWN,
			fmt.Sprintf("plugin %q is not registered in this control plane", component.Spec.Plugin),
			nil,
		)
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
		return c.setStatus(ctx, &component, pluginv1.HealthState_HEALTH_STATE_UNKNOWN,
			fmt.Sprintf("health check failed: %v", err),
			nil,
		)
	}

	log.Info("health check complete",
		"plugin", component.Spec.Plugin,
		"state", resp.State,
		"message", resp.Message,
	)

	if c.EventBus != nil {
		c.EventBus.PublishHealthChanged(
			component.Namespace,
			component.Name,
			resp.State.String(),
			resp.Message,
		)
	}

	// collect capabilities on healthy state
	var capabilities []string
	if resp.State == pluginv1.HealthState_HEALTH_STATE_HEALTHY {
		desc, err := pluginClient.Describe(checkCtx)
		if err == nil {
			for _, cap := range desc.Capabilities {
				capabilities = append(capabilities, cap.Type)
			}
		}
	}

	result, err := c.setStatus(ctx, &component, resp.State, resp.Message, capabilities)

	interval := defaultHealthCheckInterval
	if component.Spec.HealthCheckIntervalSeconds > 0 {
		interval = time.Duration(component.Spec.HealthCheckIntervalSeconds) * time.Second
	}
	result.RequeueAfter = interval

	return result, err
}

func (c *Controller) setStatus(
	ctx context.Context,
	component *registry.CosmoComponent,
	state pluginv1.HealthState,
	message string,
	capabilities []string,
) (reconcile.Result, error) {
	now := metav1.Now()
	patch := client.MergeFrom(component.DeepCopy())

	component.Status.Health = healthStateString(state)
	component.Status.Message = message
	component.Status.LastChecked = &now
	component.Status.ObservedGeneration = component.Generation

	if capabilities != nil {
		component.Status.Capabilities = capabilities
	}

	healthy := state == pluginv1.HealthState_HEALTH_STATE_HEALTHY
	conditionStatus := metav1.ConditionTrue
	conditionReason := "HealthCheckPassed"
	if !healthy {
		conditionStatus = metav1.ConditionFalse
		conditionReason = "HealthCheckFailed"
	}

	setCondition(&component.Status.Conditions, metav1.Condition{
		Type:               conditionTypeHealthy,
		Status:             conditionStatus,
		Reason:             conditionReason,
		Message:            message,
		LastTransitionTime: now,
		ObservedGeneration: component.Generation,
	})

	if err := c.Status().Patch(ctx, component, patch); err != nil {
		return reconcile.Result{}, fmt.Errorf("patching CosmoComponent status: %w", err)
	}

	return reconcile.Result{}, nil
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
