package api

import (
	"net/http"
	"strconv"
)

type LogsHandler struct {
	deps Dependencies
}

func (h LogsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	source := r.URL.Query().Get("source")
	entries, err := h.deps.Logs.ListBySource(r.Context(), source, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}
