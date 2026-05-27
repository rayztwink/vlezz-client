package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/rayflow/rayflow-client/apps/backend/internal/routing"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

type RoutingHandler struct {
	deps Dependencies
}

type createRoutingRuleRequest struct {
	Domain  string `json:"domain"`
	Mode    string `json:"mode"`
	Enabled bool   `json:"enabled"`
}

func (h RoutingHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.deps.RoutingRules.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h RoutingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoutingRuleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "domain is required")
		return
	}
	rule := storage.RoutingRule{
		ID:      uuid.NewString(),
		Domain:  req.Domain,
		Mode:    routing.NormalizeMode(req.Mode),
		Enabled: req.Enabled,
	}
	if err := h.deps.RoutingRules.Create(r.Context(), rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (h RoutingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.RoutingRules.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
