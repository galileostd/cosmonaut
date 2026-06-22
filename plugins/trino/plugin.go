// Package trino implements the Cosmonaut plugin for Apache Trino.
package trino

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/galileostd/cosmonaut/sdk"
)

func init() {
	sdk.Register(&Plugin{})
}

// Plugin implements sdk.Plugin for Apache Trino.
type Plugin struct{}

func (p *Plugin) Describe() sdk.ComponentInfo {
	return sdk.ComponentInfo{
		PluginName:  "trino",
		DisplayName: "Apache Trino",
		Type:        sdk.PluginTypeQueryEngine,
		Version:     "v0.1.0",
		Description: "Distributed SQL query engine for the lakehouse.",
	}
}

func (p *Plugin) GetCapabilities() []sdk.Capability {
	return []sdk.Capability{
		{
			Type:        sdk.CapabilityQuery,
			Description: "Execute SQL queries against Iceberg tables via Trino.",
		},
		{
			Type:        sdk.CapabilityListTables,
			Description: "List tables available in the catalog.",
		},
	}
}

func (p *Plugin) HealthCheck(ctx context.Context, component sdk.Component) (sdk.HealthStatus, error) {
	url := fmt.Sprintf("%s/v1/info", component.Endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return sdk.HealthStatus{State: sdk.HealthStateUnknown}, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return sdk.HealthStatus{
			State:   sdk.HealthStateUnhealthy,
			Message: fmt.Sprintf("failed to reach Trino at %s: %v", component.Endpoint, err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sdk.HealthStatus{
			State:   sdk.HealthStateDegraded,
			Message: fmt.Sprintf("Trino returned HTTP %d", resp.StatusCode),
		}, nil
	}

	return sdk.HealthStatus{
		State:   sdk.HealthStateHealthy,
		Message: "Trino is reachable and responding",
	}, nil
}

func (p *Plugin) Execute(ctx context.Context, action sdk.Action) (sdk.Result, error) {
	switch action.Type {
	case sdk.CapabilityQuery:
		return p.executeQuery(ctx, action)
	case sdk.CapabilityListTables:
		return p.listTables(ctx, action)
	default:
		return sdk.Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", action.Type),
		}, nil
	}
}

func (p *Plugin) executeQuery(_ context.Context, action sdk.Action) (sdk.Result, error) {
	sql, ok := action.Payload["sql"].(string)
	if !ok || sql == "" {
		return sdk.Result{Success: false, Error: "missing required payload field: sql"}, nil
	}
	// TODO: implement Trino HTTP API client
	return sdk.Result{
		Success: true,
		Data:    map[string]any{"sql": sql, "status": "submitted"},
	}, nil
}

func (p *Plugin) listTables(_ context.Context, _ sdk.Action) (sdk.Result, error) {
	// TODO: implement via Trino information_schema query
	return sdk.Result{
		Success: true,
		Data:    map[string]any{"tables": []string{}},
	}, nil
}
