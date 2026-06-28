package api

import (
	"net/http"
	"sort"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
	"github.com/go-chi/chi/v5"
)

type pluginResponse struct {
	Name          string               `json:"name"`
	DisplayName   string               `json:"display_name"`
	PluginType    string               `json:"plugin_type"`
	ExecutionType string               `json:"execution_type"`
	WorkloadType  string               `json:"workload_type,omitempty"`
	Version       string               `json:"version"`
	Description   string               `json:"description"`
	Capabilities  []capabilityResponse `json:"capabilities"`
}

type capabilityResponse struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// handleListPlugins returns all plugins discovered in the cluster.
// GET /api/v1/plugins
func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	p := parsePagination(r)

	details := s.plugins.DescribeAll(r.Context())

	// sort by name for deterministic output
	sort.Slice(details, func(i, j int) bool {
		return details[i].Info.Name < details[j].Info.Name
	})

	total := len(details)
	start := p.Offset
	if start > total {
		start = total
	}
	end := start + p.Limit
	if end > total {
		end = total
	}

	responses := make([]pluginResponse, len(details[start:end]))
	for i, d := range details[start:end] {
		responses[i] = toPluginResponse(d.Describe)
	}

	writeJSON(w, http.StatusOK, newPagedResponse(responses, total, p))
}

// handleGetPlugin returns a single plugin by name.
// GET /api/v1/plugins/:name
func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	desc, err := s.plugins.Describe(r.Context(), name)
	if err != nil {
		writeProblem(w, r, problemNotFound(r, "plugin '"+name+"' is not registered in this control plane"))
		return
	}

	writeJSON(w, http.StatusOK, toPluginResponse(desc))
}

func toPluginResponse(desc *pluginv1.DescribeResponse) pluginResponse {
	caps := make([]capabilityResponse, len(desc.Capabilities))
	for i, c := range desc.Capabilities {
		caps[i] = capabilityResponse{
			Type:        c.Type,
			Description: c.Description,
		}
	}

	resp := pluginResponse{
		Name:          desc.PluginName,
		DisplayName:   desc.DisplayName,
		PluginType:    desc.PluginType.String(),
		ExecutionType: desc.ExecutionType.String(),
		Version:       desc.Version,
		Description:   desc.Description,
		Capabilities:  caps,
	}

	if desc.WorkloadType != pluginv1.WorkloadType_WORKLOAD_TYPE_UNSPECIFIED {
		resp.WorkloadType = desc.WorkloadType.String()
	}

	return resp
}
