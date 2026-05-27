package api

import (
	"net/http"

	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
)

type SettingsHandler struct {
	deps Dependencies
}

type patchSettingsRequest struct {
	Theme                      *string `json:"theme"`
	Language                   *string `json:"language"`
	Autostart                  *bool   `json:"autostart"`
	ActiveMode                 *string `json:"activeMode"`
	DefaultCore                *string `json:"defaultCore"`
	LocalProxyPort             *int    `json:"localProxyPort"`
	SingBoxPath                *string `json:"singBoxPath"`
	XrayPath                   *string `json:"xrayPath"`
	ZapretPath                 *string `json:"zapretPath"`
	EnableSystemProxyOnConnect *bool   `json:"enableSystemProxyOnConnect"`
	PreferredNetworkMode       *string `json:"preferredNetworkMode"`
	TUNEnabled                 *bool   `json:"tunEnabled"`
	TUNStack                   *string `json:"tunStack"`
	TUNAutoRoute               *bool   `json:"tunAutoRoute"`
	TUNStrictRoute             *bool   `json:"tunStrictRoute"`
}

func (h SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.deps.Settings.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h SettingsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	var req patchSettingsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	patch, err := h.deps.Settings.Get(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if req.Theme != nil {
		patch.Theme = *req.Theme
	}
	if req.Language != nil {
		patch.Language = *req.Language
	}
	if req.Autostart != nil {
		patch.Autostart = *req.Autostart
	}
	if req.ActiveMode != nil {
		patch.ActiveMode = *req.ActiveMode
	}
	if req.DefaultCore != nil {
		patch.DefaultCore = *req.DefaultCore
	}
	if req.LocalProxyPort != nil {
		patch.LocalProxyPort = *req.LocalProxyPort
	}
	if req.SingBoxPath != nil {
		if *req.SingBoxPath != "" {
			if err := process.ValidateBinaryPath(*req.SingBoxPath); err != nil {
				writeError(w, http.StatusBadRequest, "invalid singBoxPath: "+err.Error())
				return
			}
		}
		patch.SingBoxPath = *req.SingBoxPath
	}
	if req.XrayPath != nil {
		if *req.XrayPath != "" {
			if err := process.ValidateBinaryPath(*req.XrayPath); err != nil {
				writeError(w, http.StatusBadRequest, "invalid xrayPath: "+err.Error())
				return
			}
		}
		patch.XrayPath = *req.XrayPath
	}
	if req.ZapretPath != nil {
		if *req.ZapretPath != "" {
			if err := process.ValidateBinaryPath(*req.ZapretPath); err != nil {
				writeError(w, http.StatusBadRequest, "invalid zapretPath: "+err.Error())
				return
			}
		}
		patch.ZapretPath = *req.ZapretPath
	}
	if req.EnableSystemProxyOnConnect != nil {
		patch.EnableSystemProxyOnConnect = *req.EnableSystemProxyOnConnect
	}
	if req.PreferredNetworkMode != nil {
		patch.PreferredNetworkMode = *req.PreferredNetworkMode
	}
	if req.TUNEnabled != nil {
		patch.TUNEnabled = *req.TUNEnabled
	}
	if req.TUNStack != nil {
		patch.TUNStack = *req.TUNStack
	}
	if req.TUNAutoRoute != nil {
		patch.TUNAutoRoute = *req.TUNAutoRoute
	}
	if req.TUNStrictRoute != nil {
		patch.TUNStrictRoute = *req.TUNStrictRoute
	}
	settings, err := h.deps.Settings.Patch(r.Context(), patch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}
