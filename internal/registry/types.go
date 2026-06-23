// Package registry defines the CosmoComponent CRD types and their lifecycle.
package registry

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is the API group and version for Cosmonaut CRDs.
var GroupVersion = schema.GroupVersion{
	Group:   "cosmonaut.galileostd.io",
	Version: "v1",
}

// SchemeBuilder registers the CRD types into a scheme.
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme adds all CRD types to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&CosmoComponent{},
		&CosmoComponentList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Plugin",type="string",JSONPath=".spec.plugin"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Health",type="string",JSONPath=".status.health"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CosmoComponent represents a data platform component registered with Cosmonaut.
// Any tool — Trino, Spark, Flink, Airflow, Dremio, OpenSearch, or anything else —
// becomes a first-class citizen in the control plane by creating a CosmoComponent.
type CosmoComponent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CosmoComponentSpec   `json:"spec,omitempty"`
	Status CosmoComponentStatus `json:"status,omitempty"`
}

// CosmoComponentSpec defines the desired state of a CosmoComponent.
type CosmoComponentSpec struct {
	// Plugin is the name of the plugin that handles this component.
	// Must match a registered plugin name (e.g. "trino", "spark", "airflow").
	// +kubebuilder:validation:Required
	Plugin string `json:"plugin"`

	// Type categorizes the component (e.g. "query-engine", "processing", "orchestration").
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Endpoint is the base URL to reach this component.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://`
	Endpoint string `json:"endpoint"`

	// Auth defines how the control plane authenticates with this component.
	// +optional
	Auth *AuthSpec `json:"auth,omitempty"`

	// Config holds arbitrary plugin-specific configuration key-value pairs.
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// HealthCheckIntervalSeconds defines how often the health controller
	// checks this component. Defaults to 30 seconds.
	// +optional
	// +kubebuilder:default=30
	HealthCheckIntervalSeconds int32 `json:"healthCheckIntervalSeconds,omitempty"`
}

// AuthSpec defines authentication configuration for a component.
type AuthSpec struct {
	// VaultPath is the Vault secret path containing credentials.
	// Example: "secret/data/trino/credentials"
	// +optional
	VaultPath string `json:"vaultPath,omitempty"`

	// SecretRef references a Kubernetes Secret in the same namespace.
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`
}

// SecretReference points to a Kubernetes Secret.
type SecretReference struct {
	// Name is the name of the Secret.
	Name string `json:"name"`

	// Key is the key within the Secret to use.
	Key string `json:"key"`
}

// CosmoComponentStatus defines the observed state of a CosmoComponent.
// Updated by the health controller — do not edit manually.
type CosmoComponentStatus struct {
	// Health is the current health state of the component.
	// One of: healthy, degraded, unhealthy, unknown.
	// +kubebuilder:validation:Enum=healthy;degraded;unhealthy;unknown
	Health string `json:"health,omitempty"`

	// Message provides a human-readable explanation of the current health state.
	// +optional
	Message string `json:"message,omitempty"`

	// LastChecked is the timestamp of the last health check.
	// +optional
	LastChecked *metav1.Time `json:"lastChecked,omitempty"`

	// Capabilities lists the actions this component supports,
	// as reported by its plugin. Populated on first successful health check.
	// +optional
	Capabilities []string `json:"capabilities,omitempty"`

	// ObservedGeneration is the generation of the spec that was last reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions provides detailed status conditions following the K8s convention.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// CosmoComponentList contains a list of CosmoComponent.
type CosmoComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CosmoComponent `json:"items"`
}
