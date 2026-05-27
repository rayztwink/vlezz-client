package api

import (
	"net/http"
)

type ConnectionHandler struct {
	deps Dependencies
}

func (h ConnectionHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.deps.Connection.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h ConnectionHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	status, err := h.deps.Connection.Disconnect(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h ConnectionHandler) Report(w http.ResponseWriter, r *http.Request) {
	report, err := h.deps.Connection.Report(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}
