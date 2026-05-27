package api

import (
	"net/http"
	"strconv"
)

type SystemProxyHandler struct {
	deps Dependencies
}

type enableSystemProxyRequest struct {
	ProxyServer string `json:"proxyServer"`
}

func (h SystemProxyHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.deps.SystemProxy.Status(r.Context()))
}

func (h SystemProxyHandler) Enable(w http.ResponseWriter, r *http.Request) {
	var req enableSystemProxyRequest
	_ = decodeJSON(r, &req)
	if req.ProxyServer == "" {
		settings, err := h.deps.Settings.Get(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		req.ProxyServer = "127.0.0.1:" + strconv.Itoa(settings.LocalProxyPort)
	}
	if err := h.deps.SystemProxy.Enable(r.Context(), req.ProxyServer); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.deps.SystemProxy.Status(r.Context()))
}

func (h SystemProxyHandler) Disable(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.SystemProxy.Disable(r.Context()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.deps.SystemProxy.Status(r.Context()))
}
