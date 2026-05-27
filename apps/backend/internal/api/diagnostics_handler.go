package api

import (
	"net/http"
	"strconv"

	"github.com/rayflow/rayflow-client/apps/backend/internal/diagnostics"
)

type DiagnosticsHandler struct {
	deps Dependencies
}

func (h DiagnosticsHandler) Check(w http.ResponseWriter, r *http.Request) {
	var req diagnostics.CheckRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	check, err := h.deps.Diagnostics.Check(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusOK, check)
		return
	}
	writeJSON(w, http.StatusOK, check)
}

func (h DiagnosticsHandler) IPCheck(w http.ResponseWriter, r *http.Request) {
	var req diagnostics.IPCheckRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	writeJSON(w, http.StatusOK, h.deps.Diagnostics.IPCheck(r.Context(), req))
}

func (h DiagnosticsHandler) History(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.deps.Diagnostics.History(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}
