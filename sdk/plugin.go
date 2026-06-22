// Package sdk provides the plugin interface for Cosmonaut.
// Any tool can be integrated into the control plane by implementing the Plugin interface.
package sdk

import "context"

// PluginType categorizes what kind of component a plugin represents.
type PluginType string

const (
	PluginTypeQueryEngine    PluginType = "query-engine"
	PluginTypeProcessing     PluginType = "processing"
	PluginTypeOrchestration  PluginType = "orchestration"
	PluginTypeStorage        PluginType = "storage"
	PluginTypeCatalog        PluginType = "catalog"
	PluginTypeStreaming       PluginType = "streaming"
	PluginTypeML             PluginType = "ml"
	PluginTypeObservability  PluginType = "observability"
	PluginTypeNotebook       PluginType = "notebook"
	PluginTypeBI             PluginType = "bi"
	PluginTypeGeneric        PluginType = "generic"
)

// HealthState represents the current health of a component.
type HealthState string

const (
	HealthStateHealthy   HealthState = "healthy"
	HealthStateDegraded  HealthState = "degraded"
	HealthStateUnhealthy HealthState = "unhealthy"
	HealthStateUnknown   HealthState = "unknown"
)

// Component holds the configuration of a registered component.
// This maps to the CosmoComponent CRD in Kubernetes.
type Component struct {
	// Name is the unique identifier of this component instance.
	Name string

	// Namespace is the Kubernetes namespace where this component lives.
	Namespace string

	// Endpoint is the base URL to reach this component.
	Endpoint string

	// Config holds arbitrary plugin-specific configuration.
	Config map[string]string
}

// HealthStatus is the result of a health check.
type HealthStatus struct {
	State   HealthState
	Message string
}

// CapabilityType describes what a component can do.
type CapabilityType string

const (
	CapabilityQuery       CapabilityType = "query"
	CapabilitySubmitJob   CapabilityType = "submit-job"
	CapabilityListJobs    CapabilityType = "list-jobs"
	CapabilityKillJob     CapabilityType = "kill-job"
	CapabilityListTables  CapabilityType = "list-tables"
	CapabilityListFiles   CapabilityType = "list-files"
	CapabilityTriggerDAG  CapabilityType = "trigger-dag"
	CapabilityListDAGs    CapabilityType = "list-dags"
)

// Capability describes a single action a component can perform.
type Capability struct {
	Type        CapabilityType
	Description string
}

// Action represents a request to execute something on a component.
type Action struct {
	Type    CapabilityType
	Payload map[string]any
}

// Result is the response from executing an action.
type Result struct {
	Success bool
	Data    map[string]any
	Error   string
}

// ComponentInfo provides static metadata about a plugin.
type ComponentInfo struct {
	// PluginName is the unique identifier of the plugin (e.g. "trino", "spark").
	PluginName string

	// DisplayName is the human-readable name shown in the UI.
	DisplayName string

	// Type categorizes the plugin.
	Type PluginType

	// Version is the plugin version.
	Version string

	// Description is a short description of what this plugin integrates.
	Description string
}

// Plugin is the interface every Cosmonaut plugin must implement.
//
// To write a plugin:
//  1. Import this package: github.com/galileostd/cosmonaut/sdk
//  2. Implement all methods below
//  3. Register your plugin in your plugin's init() function
//
// See plugins/trino/ for a reference implementation.
type Plugin interface {
	// HealthCheck verifies the component is reachable and operational.
	// Called periodically by the health controller.
	HealthCheck(ctx context.Context, component Component) (HealthStatus, error)

	// GetCapabilities returns the list of actions this plugin supports.
	// Used by the API and UI to render available actions.
	GetCapabilities() []Capability

	// Execute performs an action on the component.
	// The action type must be one returned by GetCapabilities.
	Execute(ctx context.Context, component Component, action Action) (Result, error)

	// Describe returns static metadata about this plugin.
	Describe() ComponentInfo
}

// Registry is the global plugin registry.
// Plugins register themselves via Register() in their init() function.
var registry = map[string]Plugin{}

// Register adds a plugin to the global registry.
// Call this from your plugin's init() function.
func Register(p Plugin) {
	info := p.Describe()
	registry[info.PluginName] = p
}

// Get returns a registered plugin by name.
// Returns nil if the plugin is not registered.
func Get(name string) Plugin {
	return registry[name]
}

// All returns all registered plugins.
func All() map[string]Plugin {
	result := make(map[string]Plugin, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}
