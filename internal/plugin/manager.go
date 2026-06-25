// Package plugin manages the lifecycle of remote plugin connections.
// Plugins are discovered automatically via Kubernetes Service labels
// and communicated with via gRPC using the cosmonaut-sdk contract.
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	sdkclient "github.com/galileostd/cosmonaut-sdk/go/client"
	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
)

// PluginInfo holds metadata about a discovered plugin.
type PluginInfo struct {
	// Name is the unique plugin identifier (e.g. "trino", "spark").
	Name string

	// Endpoint is the gRPC address (e.g. "cosmonaut-plugin-trino.cosmonaut:50051").
	Endpoint string

	// Namespace is the K8s namespace where the plugin Service lives.
	Namespace string

	// ServiceName is the K8s Service name.
	ServiceName string

	// Describe cache — populated on first successful call.
	describe *pluginv1.DescribeResponse
}

// Manager maintains gRPC connections to all discovered plugins.
// It is the single point of contact between the control-plane and plugin instances.
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*entry // keyed by plugin name
}

type entry struct {
	info   PluginInfo
	client *sdkclient.Client
}

// NewManager creates a new plugin Manager.
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]*entry),
	}
}

// Register adds or updates a plugin in the manager.
// If a plugin with the same name already exists, its connection is replaced.
func (m *Manager) Register(info PluginInfo) error {
	client, err := sdkclient.New(info.Endpoint)
	if err != nil {
		return fmt.Errorf("connecting to plugin %q at %s: %w", info.Name, info.Endpoint, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// close existing connection if replacing
	if existing, ok := m.plugins[info.Name]; ok {
		_ = existing.client.Close()
	}

	m.plugins[info.Name] = &entry{info: info, client: client}
	slog.Info("plugin registered", "name", info.Name, "endpoint", info.Endpoint)
	return nil
}

// Unregister removes a plugin and closes its connection.
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.plugins[name]; ok {
		_ = e.client.Close()
		delete(m.plugins, name)
		slog.Info("plugin unregistered", "name", name)
	}
}


// UnregisterByService removes a plugin by looking up the Service name.
// This is necessary because when the Service is deleted, we no longer have access to its labels.
func (m *Manager) UnregisterByService(serviceName, namespace string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for name, e := range m.plugins {
        if e.info.ServiceName == serviceName && e.info.Namespace == namespace {
            _ = e.client.Close()
            delete(m.plugins, name)
            slog.Info("plugin unregistered by service", "name", name, "service", serviceName)
            return
        }
    }
}

// Get returns the gRPC client for a plugin by name.
// Returns nil if the plugin is not registered.
func (m *Manager) Get(name string) *sdkclient.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if e, ok := m.plugins[name]; ok {
		return e.client
	}
	return nil
}

// GetInfo returns metadata about a registered plugin.
func (m *Manager) GetInfo(name string) (PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if e, ok := m.plugins[name]; ok {
		return e.info, true
	}
	return PluginInfo{}, false
}

// All returns all registered plugins with their describe responses.
// Plugins that fail to describe are skipped with a warning.
func (m *Manager) All(ctx context.Context) []PluginDetail {
	m.mu.RLock()
	entries := make([]*entry, 0, len(m.plugins))
	for _, e := range m.plugins {
		entries = append(entries, e)
	}
	m.mu.RUnlock()

	details := make([]PluginDetail, 0, len(entries))
	for _, e := range entries {
		desc, err := e.client.Describe(ctx)
		if err != nil {
			slog.Warn("failed to describe plugin", "name", e.info.Name, "err", err)
			continue
		}
		details = append(details, PluginDetail{
			Info:     e.info,
			Describe: desc,
		})
	}
	return details
}

// Describe returns the describe response for a single plugin.
func (m *Manager) Describe(ctx context.Context, name string) (*pluginv1.DescribeResponse, error) {
	client := m.Get(name)
	if client == nil {
		return nil, fmt.Errorf("plugin %q is not registered", name)
	}
	return client.Describe(ctx)
}

// IsAlive checks if a plugin gRPC server is reachable.
func (m *Manager) IsAlive(ctx context.Context, name string) bool {
	client := m.Get(name)
	if client == nil {
		return false
	}
	return client.IsAlive(ctx)
}

// Names returns the names of all registered plugins.
func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}



// PluginDetail combines plugin info with its describe response.
type PluginDetail struct {
	Info     PluginInfo
	Describe *pluginv1.DescribeResponse
}
