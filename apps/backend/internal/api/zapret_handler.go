package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

type ZapretHandler struct {
	deps Dependencies
}

func (h ZapretHandler) ListPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := h.deps.Presets.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, presets)
}

func (h ZapretHandler) UpdatePresets(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC().Format(time.RFC3339)
	defaults := []storage.ZapretPreset{
		{ID: "flowseal-youtube", Name: "Flowseal YouTube", Source: "built-in", Command: h.deps.Config.ZapretPath + " --preset youtube", Description: "Starter placeholder preset for YouTube diagnostics", UpdatedAt: now},
		{ID: "flowseal-discord", Name: "Flowseal Discord", Source: "built-in", Command: h.deps.Config.ZapretPath + " --preset discord", Description: "Starter placeholder preset for Discord diagnostics", UpdatedAt: now},
	}
	for _, preset := range defaults {
		if preset.Command == " --preset youtube" || preset.Command == " --preset discord" {
			preset.Command = "C:\\Path\\To\\zapret.exe --preset " + preset.ID
		}
		if err := h.deps.Presets.Upsert(r.Context(), preset); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h ZapretHandler) StartPreset(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.Zapret.StartPreset(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "starting"})
}

func (h ZapretHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.deps.Zapret.Stop(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h ZapretHandler) Logs(w http.ResponseWriter, r *http.Request) {
	entries, err := h.deps.Logs.ListBySource(r.Context(), "zapret", 300)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
