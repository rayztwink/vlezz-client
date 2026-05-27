package api

import "net/http"

type HealthHandler struct{}

func (HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rayflowd"})
}
