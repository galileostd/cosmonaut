// Package trino implements the Cosmonaut plugin for Apache Trino.
package trino

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/galileostd/cosmonaut/sdk"
)

func init() {
	sdk.Register(&Plugin{})
}

// Plugin implements sdk.Plugin for Apache Trino.
type Plugin struct {
	httpClient *http.Client
}

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
	url := strings.TrimRight(component.Endpoint, "/") + "/v1/info"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return sdk.HealthStatus{State: sdk.HealthStateUnknown}, err
	}

	client := p.getClient()
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

func (p *Plugin) Execute(ctx context.Context, component sdk.Component, action sdk.Action) (sdk.Result, error) {
	switch action.Type {
	case sdk.CapabilityQuery:
		return p.executeQuery(ctx, component, action)
	case sdk.CapabilityListTables:
		return p.listTables(ctx, component, action)
	default:
		return sdk.Result{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", action.Type),
		}, nil
	}
}

func (p *Plugin) getClient() *http.Client {
	if p.httpClient == nil {
		p.httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return p.httpClient
}

// ─── Query Execution ──────────────────────────────────────────────────────────

// TrinoQueryResponse represents the response from Trino's /v1/statement endpoint.
type TrinoQueryResponse struct {
	ID          string          `json:"id"`
	InfoURI     string          `json:"infoUri"`
	Columns     []TrinoColumn   `json:"columns"`
	Data        [][]interface{} `json:"data"`
	Stats       TrinoStats      `json:"stats"`
	Error       *TrinoError     `json:"error,omitempty"`
	NextURI     string          `json:"nextUri,omitempty"`
	PartialData bool            `json:"partialData,omitempty"`
}

type TrinoColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TrinoStats struct {
	State          string  `json:"state"`
	Queued         bool    `json:"queued"`
	Scheduled      bool    `json:"scheduled"`
	Nodes          int     `json:"nodes"`
	TotalSplits    int     `json:"totalSplits"`
	QueuedSplits   int     `json:"queuedSplits"`
	RunningSplits  int     `json:"runningSplits"`
	CompletedSplits int    `json:"completedSplits"`
	CPUTimeMillis  int64   `json:"cpuTimeMillis"`
	WallTimeMillis int64   `json:"wallTimeMillis"`
	QueuedTimeMillis int64 `json:"queuedTimeMillis"`
	ElapsedTimeMillis int64 `json:"elapsedTimeMillis"`
	ProcessedRows  int64   `json:"processedRows"`
	ProcessedBytes int64   `json:"processedBytes"`
	PeakMemoryBytes int64  `json:"peakMemoryBytes"`
	SpilledBytes   int64   `json:"spilledBytes"`
}

type TrinoError struct {
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
	ErrorName string `json:"errorName"`
	ErrorType string `json:"errorType"`
}

func (p *Plugin) executeQuery(ctx context.Context, component sdk.Component, action sdk.Action) (sdk.Result, error) {
	sql, ok := action.Payload["sql"].(string)
	if !ok || sql == "" {
		return sdk.Result{Success: false, Error: "missing required payload field: sql"}, nil
	}

	catalog := "iceberg"
	if v, ok := action.Payload["catalog"].(string); ok && v != "" {
		catalog = v
	}
	schema := "default"
	if v, ok := action.Payload["schema"].(string); ok && v != "" {
		schema = v
	}

	endpoint := strings.TrimRight(component.Endpoint, "/")

	// Step 1: POST the query to /v1/statement
	queryURL := fmt.Sprintf("%s/v1/statement", endpoint)
	reqBody := bytes.NewBufferString(sql)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queryURL, reqBody)
	if err != nil {
		return sdk.Result{Success: false, Error: err.Error()}, nil
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Trino-Catalog", catalog)
	req.Header.Set("X-Trino-Schema", schema)
	req.Header.Set("X-Trino-User", "cosmonaut")
	req.Header.Set("X-Trino-Source", "cosmonaut")
	req.Header.Set("X-Trino-Client-Info", fmt.Sprintf("cosmonaut/%s", component.Name))

	client := p.getClient()
	resp, err := client.Do(req)
	if err != nil {
		return sdk.Result{Success: false, Error: fmt.Sprintf("Trino request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return sdk.Result{
			Success: false,
			Error:   fmt.Sprintf("Trino returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var queryResp TrinoQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return sdk.Result{Success: false, Error: fmt.Sprintf("failed to decode Trino response: %v", err)}, nil
	}

	// Check for immediate error
	if queryResp.Error != nil {
		return sdk.Result{
			Success: false,
			Error:   fmt.Sprintf("Trino error: %s (code: %d)", queryResp.Error.Message, queryResp.Error.ErrorCode),
		}, nil
	}

	// Step 2: Fetch all pages if there's a nextUri
	var allData [][]interface{}
	allData = append(allData, queryResp.Data...)

	nextURI := queryResp.NextURI
	for nextURI != "" {
		select {
		case <-ctx.Done():
			return sdk.Result{Success: false, Error: "query cancelled"}, nil
		default:
		}

		resp, err := client.Get(nextURI)
		if err != nil {
			return sdk.Result{Success: false, Error: fmt.Sprintf("failed to fetch next page: %v", err)}, nil
		}

		var pageResp TrinoQueryResponse
		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			resp.Body.Close()
			return sdk.Result{Success: false, Error: fmt.Sprintf("failed to decode page: %v", err)}, nil
		}
		resp.Body.Close()

		if pageResp.Error != nil {
			return sdk.Result{
				Success: false,
				Error:   fmt.Sprintf("Trino error on page: %s", pageResp.Error.Message),
			}, nil
		}

		allData = append(allData, pageResp.Data...)
		nextURI = pageResp.NextURI
	}

	// Step 3: Build result
	columns := make([]string, len(queryResp.Columns))
	for i, col := range queryResp.Columns {
		columns[i] = col.Name
	}

	return sdk.Result{
		Success: true,
		Data: map[string]any{
			"columns":      columns,
			"rows":         allData,
			"row_count":    len(allData),
			"stats": map[string]any{
				"processed_rows":   queryResp.Stats.ProcessedRows,
				"processed_bytes":  queryResp.Stats.ProcessedBytes,
				"elapsed_ms":       queryResp.Stats.ElapsedTimeMillis,
				"cpu_ms":          queryResp.Stats.CPUTimeMillis,
				"state":           queryResp.Stats.State,
			},
		},
	}, nil
}

// ─── List Tables ──────────────────────────────────────────────────────────────

func (p *Plugin) listTables(ctx context.Context, component sdk.Component, action sdk.Action) (sdk.Result, error) {
	catalog := "iceberg"
	if v, ok := action.Payload["catalog"].(string); ok && v != "" {
		catalog = v
	}
	schema := "default"
	if v, ok := action.Payload["schema"].(string); ok && v != "" {
		schema = v
	}

	sql := fmt.Sprintf("SELECT table_name, table_type, comment FROM %s.information_schema.tables WHERE table_schema = '%s'", catalog, schema)

	// Reuse executeQuery with a modified payload
	queryAction := sdk.Action{
		Type: sdk.CapabilityQuery,
		Payload: map[string]any{
			"sql":     sql,
			"catalog": catalog,
			"schema":  schema,
		},
	}

	result, err := p.executeQuery(ctx, component, queryAction)
	if err != nil {
		return result, err
	}

	if !result.Success {
		return result, nil
	}

	// Extract tables from the query result
	rows, ok := result.Data["rows"].([][]interface{})
	if !ok {
		return sdk.Result{Success: true, Data: map[string]any{"tables": []map[string]string{}}}, nil
	}

	tables := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		if len(row) >= 3 {
			tableName, _ := row[0].(string)
			tableType, _ := row[1].(string)
			comment, _ := row[2].(string)
			tables = append(tables, map[string]string{
				"name":    tableName,
				"type":    tableType,
				"comment": comment,
			})
		}
	}

	return sdk.Result{
		Success: true,
		Data: map[string]any{
			"tables":  tables,
			"catalog": catalog,
			"schema":  schema,
			"count":   len(tables),
		},
	}, nil
}