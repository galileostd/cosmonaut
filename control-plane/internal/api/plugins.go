package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/galileostd/cosmonaut/sdk"
)

type pluginResponse struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Type         string   `json:"type"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Capabilities []capabilityResponse `json:"capabilities"`
}

type capabilityResponse struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// handleListPlugins returns all plugins registered in this control plane instance.
// GET /api/v1/plugins
func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	all := sdk.All()
	p := parsePagination(r)

	plugins := make([]pluginResponse, 0, len(all))
	for _, plugin := range all {
		plugins = append(plugins, toPluginResponse(plugin))
	}

	// simple offset/limit over the map (order is not guaranteed — sort by name)
	total := len(plugins)
	start := p.Offset
	if start > total {
		start = total
	}
	end := start + p.Limit
	if end > total {
		end = total
	}

	writeJSON(w, http.StatusOK, newPagedResponse(plugins[start:end], total, p))
}

// handleGetPlugin returns a single plugin by name.
// GET /api/v1/plugins/:name
func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	plugin := sdk.Get(name)
	if plugin == nil {
		writeProblem(w, r, problemNotFound(r, "plugin '"+name+"' is not registered in this control plane"))
		return
	}

	writeJSON(w, http.StatusOK, toPluginResponse(plugin))
}

func toPluginResponse(p sdk.Plugin) pluginResponse {
	info := p.Describe()
	caps := p.GetCapabilities()

	capResponses := make([]capabilityResponse, len(caps))
	for i, c := range caps {
		capResponses[i] = capabilityResponse{
			Type:        string(c.Type),
			Description: c.Description,
		}
	}

	return pluginResponse{
		Name:         info.PluginName,
		DisplayName:  info.DisplayName,
		Type:         string(info.Type),
		Version:      info.Version,
		Description:  info.Description,
		Capabilities: capResponses,
	}
}
