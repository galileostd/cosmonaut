package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/galileostd/cosmonaut/sdk"
	"github.com/galileostd/cosmonaut/control-plane/internal/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ── response types ─────────────────────────────────────────────────────────────

type componentResponse struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Plugin      string            `json:"plugin"`
	Type        string            `json:"type"`
	Endpoint    string            `json:"endpoint"`
	Config      map[string]string `json:"config,omitempty"`
	Health      string            `json:"health"`
	Message     string            `json:"message,omitempty"`
	LastChecked *time.Time        `json:"last_checked,omitempty"`
	Capabilities []string         `json:"capabilities,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type createComponentRequest struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Plugin                     string            `json:"plugin"`
	Type                       string            `json:"type"`
	Endpoint                   string            `json:"endpoint"`
	Config                     map[string]string `json:"config,omitempty"`
	HealthCheckIntervalSeconds int32             `json:"health_check_interval_seconds,omitempty"`
}

type execRequest struct {
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload,omitempty"`
}

type execResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ── handlers ───────────────────────────────────────────────────────────────────

// handleListComponents returns a paginated list of all CosmoComponents.
// GET /api/v1/components
// Query params: limit, offset, namespace (optional filter), plugin (optional filter)
func (s *Server) handleListComponents(w http.ResponseWriter, r *http.Request) {
	p := parsePagination(r)
	nsFilter := r.URL.Query().Get("namespace")
	pluginFilter := r.URL.Query().Get("plugin")

	var list registry.CosmoComponentList
	opts := []client.ListOption{}
	if nsFilter != "" {
		opts = append(opts, client.InNamespace(nsFilter))
	}

	if err := s.k8s.List(r.Context(), &list, opts...); err != nil {
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to list components: %v", err)))
		return
	}

	items := list.Items

	// filter by plugin if requested
	if pluginFilter != "" {
		filtered := items[:0]
		for _, item := range items {
			if item.Spec.Plugin == pluginFilter {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	total := len(items)

	// apply pagination
	start := p.Offset
	if start > total {
		start = total
	}
	end := start + p.Limit
	if end > total {
		end = total
	}

	responses := make([]componentResponse, len(items[start:end]))
	for i, item := range items[start:end] {
		responses[i] = toComponentResponse(item)
	}

	writeJSON(w, http.StatusOK, newPagedResponse(responses, total, p))
}

// handleGetComponent returns a single CosmoComponent by namespace and name.
// GET /api/v1/components/:namespace/:name
func (s *Server) handleGetComponent(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	var component registry.CosmoComponent
	if err := s.k8s.Get(r.Context(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &component); err != nil {
		if isNotFound(err) {
			writeProblem(w, r, problemNotFound(r, fmt.Sprintf(
				"component '%s' not found in namespace '%s'", name, namespace,
			)))
			return
		}
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to get component: %v", err)))
		return
	}

	writeJSON(w, http.StatusOK, toComponentResponse(component))
}

// handleCreateComponent creates a new CosmoComponent in the cluster.
// POST /api/v1/components
func (s *Server) handleCreateComponent(w http.ResponseWriter, r *http.Request) {
	var req createComponentRequest
	if prob := decodeJSON(r, &req); prob != nil {
		writeProblem(w, r, prob)
		return
	}

	// validate required fields
	if req.Name == "" {
		writeProblem(w, r, problemUnprocessable(r, "name is required"))
		return
	}
	if req.Namespace == "" {
		req.Namespace = "cosmonaut"
	}
	if req.Plugin == "" {
		writeProblem(w, r, problemUnprocessable(r, "plugin is required"))
		return
	}
	if req.Type == "" {
		writeProblem(w, r, problemUnprocessable(r, "type is required"))
		return
	}
	if req.Endpoint == "" {
		writeProblem(w, r, problemUnprocessable(r, "endpoint is required"))
		return
	}

	// verify the plugin is registered
	if sdk.Get(req.Plugin) == nil {
		writeProblem(w, r, problemUnprocessable(r, fmt.Sprintf(
			"plugin '%s' is not registered in this control plane", req.Plugin,
		)))
		return
	}

	interval := req.HealthCheckIntervalSeconds
	if interval == 0 {
		interval = 30
	}

	component := &registry.CosmoComponent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: registry.CosmoComponentSpec{
			Plugin:                     req.Plugin,
			Type:                       req.Type,
			Endpoint:                   req.Endpoint,
			Config:                     req.Config,
			HealthCheckIntervalSeconds: interval,
		},
	}

	if err := s.k8s.Create(r.Context(), component); err != nil {
		if isAlreadyExists(err) {
			writeProblem(w, r, problemConflict(r, fmt.Sprintf(
				"component '%s' already exists in namespace '%s'", req.Name, req.Namespace,
			)))
			return
		}
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to create component: %v", err)))
		return
	}

	writeJSON(w, http.StatusCreated, toComponentResponse(*component))
}

// handleDeleteComponent removes a CosmoComponent from the cluster.
// DELETE /api/v1/components/:namespace/:name
func (s *Server) handleDeleteComponent(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	var component registry.CosmoComponent
	if err := s.k8s.Get(r.Context(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &component); err != nil {
		if isNotFound(err) {
			writeProblem(w, r, problemNotFound(r, fmt.Sprintf(
				"component '%s' not found in namespace '%s'", name, namespace,
			)))
			return
		}
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to get component: %v", err)))
		return
	}

	if err := s.k8s.Delete(r.Context(), &component); err != nil {
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to delete component: %v", err)))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleExecComponent submits an async action to a component via its plugin.
// POST /api/v1/components/:namespace/:name/exec
//
// Returns a job ID immediately. The caller tracks progress via:
//   GET  /api/v1/jobs/:id
//   WS   /api/v1/events (job.* events)
func (s *Server) handleExecComponent(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")

	var req execRequest
	if prob := decodeJSON(r, &req); prob != nil {
		writeProblem(w, r, prob)
		return
	}

	if req.Action == "" {
		writeProblem(w, r, problemUnprocessable(r, "action is required"))
		return
	}

	// fetch the component
	var component registry.CosmoComponent
	if err := s.k8s.Get(r.Context(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &component); err != nil {
		if isNotFound(err) {
			writeProblem(w, r, problemNotFound(r, fmt.Sprintf(
				"component '%s' not found in namespace '%s'", name, namespace,
			)))
			return
		}
		writeProblem(w, r, problemInternal(r, fmt.Sprintf("failed to get component: %v", err)))
		return
	}

	// verify the plugin is available
	plugin := sdk.Get(component.Spec.Plugin)
	if plugin == nil {
		writeProblem(w, r, problemServiceUnavailable(r, fmt.Sprintf(
			"plugin '%s' is not registered in this control plane", component.Spec.Plugin,
		)))
		return
	}

	// verify the action is supported
	action := sdk.Action{
		Type:    sdk.CapabilityType(req.Action),
		Payload: req.Payload,
	}

	supported := false
	for _, cap := range plugin.GetCapabilities() {
		if cap.Type == action.Type {
			supported = true
			break
		}
	}
	if !supported {
		writeProblem(w, r, problemUnprocessable(r, fmt.Sprintf(
			"action '%s' is not supported by plugin '%s'", req.Action, component.Spec.Plugin,
		)))
		return
	}

	sdkComponent := sdk.Component{
		Name:      component.Name,
		Namespace: component.Namespace,
		Endpoint:  component.Spec.Endpoint,
		Config:    component.Spec.Config,
	}

	job := s.jobs.Submit(func(ctx context.Context) (sdk.Result, error) {
		return plugin.Execute(ctx, sdkComponent, action)
	})

	writeJSON(w, http.StatusAccepted, execResponse{
		JobID:     job.ID,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt,
	})
}

// ── helpers ────────────────────────────────────────────────────────────────────

func toComponentResponse(c registry.CosmoComponent) componentResponse {
	resp := componentResponse{
		Name:         c.Name,
		Namespace:    c.Namespace,
		Plugin:       c.Spec.Plugin,
		Type:         c.Spec.Type,
		Endpoint:     c.Spec.Endpoint,
		Config:       c.Spec.Config,
		Health:       c.Status.Health,
		Message:      c.Status.Message,
		Capabilities: c.Status.Capabilities,
		CreatedAt:    c.CreationTimestamp.Time,
		UpdatedAt:    c.CreationTimestamp.Time,
	}

	if c.Status.LastChecked != nil {
		t := c.Status.LastChecked.Time
		resp.LastChecked = &t
	}

	return resp
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return client.IgnoreNotFound(err) == nil
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	// k8s returns 409 for already exists
	return fmt.Sprintf("%T", err) == "*errors.StatusError" &&
		fmt.Sprintf("%v", err) != fmt.Sprintf("%v", err)
}
