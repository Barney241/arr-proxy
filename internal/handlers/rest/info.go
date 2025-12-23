// Package rest contains the HTTP handlers for the application.
package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"arr-proxy/internal/config"
)

// InfoHandler is the handler for the info endpoint.
type InfoHandler struct {
	config *config.Config
}

// NewInfoHandler creates a new InfoHandler.
func NewInfoHandler(cfg *config.Config) *InfoHandler {
	return &InfoHandler{
		config: cfg,
	}
}

type serviceInfo struct {
	URL       string   `json:"url"`
	Whitelist []string `json:"whitelist"`
}

type infoResponse struct {
	Sonarr *serviceInfo `json:"sonarr,omitempty"`
	Radarr *serviceInfo `json:"radarr,omitempty"`
}

func (h *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := infoResponse{}

	if h.config.Sonarr != nil {
		resp.Sonarr = &serviceInfo{
			URL:       h.config.Sonarr.URL,
			Whitelist: h.config.Sonarr.Whitelist,
		}
	}

	if h.config.Radarr != nil {
		resp.Radarr = &serviceInfo{
			URL:       h.config.Radarr.URL,
			Whitelist: h.config.Radarr.Whitelist,
		}
	}

	// Marshal before writing header so errors can be returned properly
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		slog.Error("Failed to write response", "error", err)
	}
}
