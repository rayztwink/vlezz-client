package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/rayflow/rayflow-client/apps/backend/internal/connection"
	"github.com/rayflow/rayflow-client/apps/backend/internal/diagnostics"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/subscription"
)

type NodesHandler struct {
	deps Dependencies
}

type importNodeRequest struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

type connectNodeRequest struct {
	Core        string `json:"core"`
	NetworkMode string `json:"networkMode"`
}

func (h NodesHandler) List(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.deps.Nodes.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (h NodesHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req importNodeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	parsed, err := subscription.ParseLink(req.Link)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Name != "" {
		parsed.Name = req.Name
	}
	node := storage.Node{
		ID:        uuid.NewString(),
		Name:      parsed.Name,
		Protocol:  parsed.Protocol,
		Address:   parsed.Address,
		Port:      parsed.Port,
		UUID:      parsed.UUID,
		Security:  parsed.Security,
		Transport: parsed.Transport,
		RawLink:   parsed.RawLink,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.deps.Nodes.Create(r.Context(), node); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deps.Logs.Info("nodes", fmt.Sprintf("imported node %s", node.Name))
	writeJSON(w, http.StatusCreated, node)
}

func (h NodesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.Nodes.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h NodesHandler) Check(w http.ResponseWriter, r *http.Request) {
	node, err := h.deps.Nodes.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	target := net.JoinHostPort(node.Address, fmt.Sprintf("%d", node.Port))
	check, err := h.deps.Diagnostics.Check(r.Context(), diagnostics.CheckRequest{Target: target, Mode: "direct", Type: "tcp"})
	if check.LatencyMS != nil {
		_ = h.deps.Nodes.UpdateLatency(r.Context(), node.ID, *check.LatencyMS)
	}
	if err != nil {
		writeJSON(w, http.StatusOK, check)
		return
	}
	writeJSON(w, http.StatusOK, check)
}

func (h NodesHandler) Connect(w http.ResponseWriter, r *http.Request) {
	var req connectNodeRequest
	_ = decodeJSON(r, &req)
	status, err := h.deps.Connection.Connect(r.Context(), connection.ConnectRequest{
		NodeID:      chi.URLParam(r, "id"),
		Core:        req.Core,
		NetworkMode: req.NetworkMode,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h NodesHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	status, err := h.deps.Connection.Disconnect(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}
