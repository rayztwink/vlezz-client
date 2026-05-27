package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/subscription"
)

type SubscriptionsHandler struct {
	deps Dependencies
}

type createSubscriptionRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	UpdateInterval int    `json:"updateInterval"`
}

func (h SubscriptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.deps.Subscriptions.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h SubscriptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.URL == "" {
		writeError(w, http.StatusBadRequest, "name and url are required")
		return
	}
	if req.UpdateInterval <= 0 {
		req.UpdateInterval = 1440
	}
	item := storage.Subscription{
		ID:             uuid.NewString(),
		Name:           req.Name,
		URL:            req.URL,
		UpdateInterval: req.UpdateInterval,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.deps.Subscriptions.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h SubscriptionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.deps.Subscriptions.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	links, err := subscription.NewUpdater().FetchLinks(r.Context(), item.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	imported := 0
	skipped := 0
	failed := 0
	for _, link := range links {
		exists, existsErr := h.deps.Nodes.ExistsRawLink(r.Context(), link)
		if existsErr != nil {
			writeError(w, http.StatusInternalServerError, existsErr.Error())
			return
		}
		if exists {
			skipped++
			continue
		}
		parsed, parseErr := subscription.ParseLink(link)
		if parseErr != nil {
			failed++
			continue
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
			failed++
			continue
		}
		imported++
	}

	if err := h.deps.Subscriptions.MarkUpdated(r.Context(), id, time.Now().UTC().Format(time.RFC3339)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.deps.Logs.Info("subscriptions", fmt.Sprintf("updated subscription %s: imported=%d skipped=%d failed=%d", item.Name, imported, skipped, failed))
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "updated",
		"id":       id,
		"total":    len(links),
		"imported": imported,
		"skipped":  skipped,
		"failed":   failed,
	})
}

func (h SubscriptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.Subscriptions.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
