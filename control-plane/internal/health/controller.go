// Package health implements the CosmoComponent health controller.
// It watches all CosmoComponent objects in the cluster and periodically
// calls the corresponding plugin's HealthCheck, updating the component status.
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

	"github.com/galileostd/cosmonaut/sdk"
	"github.com/galileostd/cosmonaut/control-plane/internal/api"
	"github.com/galileostd/cosmonaut/control-plane/internal/registry"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
	healthCheckTimeout         = 10 * time.Second

	conditionTypeHealthy = "Healthy"
)

// Controller reconciles CosmoComponent objects.
type Controller struct {
	client.Client
	EventBus *api.EventBus  // <-- ADICIONE ESTE CAMPO
}

// SetupWithManager registers the controller with the controller-runtime manager.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registry.CosmoComponent{}).
		Complete(c)
}

// Reconcile is called by controller-runtime whenever a CosmoComponent is
// created, updated, or deleted — and also on a periodic requeue.
func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := slog.With("component", req.NamespacedName)

	// fetch the component
	var component registry.CosmoComponent
	if err := c.Get(ctx, req.NamespacedName, &component); err != nil {
		if errors.IsNotFound(err) {
			// component was deleted — nothing to do
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("fetching CosmoComponent: %w", err)
	}

	// find the plugin
	plugin := sdk.Get(component.Spec.Plugin)
	if plugin == nil {
		log.Warn("no plugin registered for component", "plugin", component.Spec.Plugin)
		return c.setStatus(ctx, &component, sdk.HealthStatus{
			State:   sdk.HealthStateUnknown,
			Message: fmt.Sprintf("plugin %q is not registered in this control plane", component.Spec.Plugin),
		}, nil)
	}

	// run the health check with a timeout
	checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	sdkComponent := sdk.Component{
		Name:      component.Name,
		Namespace: component.Namespace,
		Endpoint:  component.Spec.Endpoint,
		Config:    component.Spec.Config,
	}

	status, err := plugin.HealthCheck(checkCtx, sdkComponent)
	if err != nil {
		log.Error("health check error", "plugin", component.Spec.Plugin, "err", err)
		status = sdk.HealthStatus{
			State:   sdk.HealthStateUnknown,
			Message: fmt.Sprintf("health check failed: %v", err),
		}
	}

	log.Info("health check complete",
		"plugin", component.Spec.Plugin,
		"state", status.State,
		"message", status.Message,
	)

	// collect capabilities on healthy state
	var capabilities []string
	if status.State == sdk.HealthStateHealthy {
		for _, cap := range plugin.GetCapabilities() {
			capabilities = append(capabilities, string(cap.Type))
		}
	}

	result, err := c.setStatus(ctx, &component, status, capabilities)

	// Publish event via EventBus if available
	if c.EventBus != nil {
		c.EventBus.PublishHealthChanged(
			component.Namespace,
			component.Name,
			string(status.State),
			status.Message,
		)
	}

	// requeue after the configured interval
	interval := defaultHealthCheckInterval
	if component.Spec.HealthCheckIntervalSeconds > 0 {
		interval = time.Duration(component.Spec.HealthCheckIntervalSeconds) * time.Second
	}
	result.RequeueAfter = interval

	return result, err
}

// setStatus patches the CosmoComponent status with the result of a health check.
func (c *Controller) setStatus(
	ctx context.Context,
	component *registry.CosmoComponent,
	status sdk.HealthStatus,
	capabilities []string,
) (reconcile.Result, error) {
	now := metav1.Now()

	patch := client.MergeFrom(component.DeepCopy())

	component.Status.Health = string(status.State)
	component.Status.Message = status.Message
	component.Status.LastChecked = &now
	component.Status.ObservedGeneration = component.Generation

	if capabilities != nil {
		component.Status.Capabilities = capabilities
	}

	// set the Healthy condition following the K8s convention
	healthy := status.State == sdk.HealthStateHealthy
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
		Message:            status.Message,
		LastTransitionTime: now,
		ObservedGeneration: component.Generation,
	})

	if err := c.Status().Patch(ctx, component, patch); err != nil {
		return reconcile.Result{}, fmt.Errorf("patching CosmoComponent status: %w", err)
	}

	return reconcile.Result{}, nil
}

// setCondition upserts a condition in the conditions slice.
func setCondition(conditions *[]metav1.Condition, newCondition metav1.Condition) {
	for i, existing := range *conditions {
		if existing.Type == newCondition.Type {
			// preserve LastTransitionTime if status did not change
			if existing.Status == newCondition.Status {
				newCondition.LastTransitionTime = existing.LastTransitionTime
			}
			(*conditions)[i] = newCondition
			return
		}
	}
	*conditions = append(*conditions, newCondition)
}