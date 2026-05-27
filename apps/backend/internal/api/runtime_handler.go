package api

import (
	"net/http"
	"runtime"

	"github.com/rayflow/rayflow-client/apps/backend/internal/platform"
)

type RuntimeHandler struct {
	deps Dependencies
}

type runtimeCapabilitiesResponse struct {
	Platform             string `json:"platform"`
	IsAdmin              bool   `json:"isAdmin"`
	SystemProxySupported bool   `json:"systemProxySupported"`
}

func (h RuntimeHandler) Capabilities(w http.ResponseWriter, r *http.Request) {
	supported := false
	if h.deps.SystemProxy != nil {
		supported = h.deps.SystemProxy.Supported()
	}
	writeJSON(w, http.StatusOK, runtimeCapabilitiesResponse{
		Platform:             runtime.GOOS,
		IsAdmin:              platform.IsAdmin(),
		SystemProxySupported: supported,
	})
}
