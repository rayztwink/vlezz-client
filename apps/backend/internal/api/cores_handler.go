package api

import (
	"bytes"
	"context"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
)

type CoresHandler struct {
	deps Dependencies
}

type validateCoreRequest struct {
	Core string `json:"core"`
	Path string `json:"path"`
}

type validateCoreResponse struct {
	OK      bool   `json:"ok"`
	Core    string `json:"core"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h CoresHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]process.Snapshot{
		"singBox": h.deps.SingBox.Status(),
		"xray":    h.deps.Xray.Status(),
		"zapret":  h.deps.Zapret.Status(),
	})
}

func (h CoresHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req validateCoreRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp := validateCoreResponse{Core: req.Core, Path: req.Path}
	if err := process.ValidateBinaryPath(req.Path); err != nil {
		resp.Error = err.Error()
		writeJSON(w, http.StatusOK, resp)
		return
	}

	version, err := probeCoreVersion(r.Context(), req.Core, req.Path)
	if err != nil {
		resp.Error = err.Error()
		writeJSON(w, http.StatusOK, resp)
		return
	}
	resp.OK = true
	resp.Version = version
	writeJSON(w, http.StatusOK, resp)
}

func probeCoreVersion(parent context.Context, core string, path string) (string, error) {
	args := versionArgs(core)
	if args == nil {
		return "file validation passed; version probe skipped", nil
	}
	ctx, cancel := context.WithTimeout(parent, 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return "", err
	}
	version := strings.TrimSpace(output.String())
	if version == "" {
		return "version command returned no output", nil
	}
	if len(version) > 600 {
		version = version[:600]
	}
	return version, nil
}

func versionArgs(core string) []string {
	switch strings.ToLower(core) {
	case "sing-box", "singbox":
		return []string{"version"}
	case "xray", "xray-core":
		return []string{"version"}
	case "zapret":
		return nil
	default:
		return []string{"version"}
	}
}
