package diagnostics

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

type CheckSink interface {
	Create(ctx context.Context, c storage.DiagnosticCheck) error
	History(ctx context.Context, limit int) ([]storage.DiagnosticCheck, error)
}

type Service struct {
	checks CheckSink
	logs   *logs.Manager
}

func NewService(checks CheckSink, logManager *logs.Manager) *Service {
	return &Service{checks: checks, logs: logManager}
}

type CheckRequest struct {
	Target string `json:"target"`
	Mode   string `json:"mode"`
	Type   string `json:"type"`
}

func (s *Service) Check(ctx context.Context, req CheckRequest) (storage.DiagnosticCheck, error) {
	if req.Mode == "" {
		req.Mode = "direct"
	}
	if req.Type == "" {
		req.Type = "tcp"
	}
	var latency time.Duration
	var err error

	switch req.Type {
	case "dns":
		latency, _, err = DNSResolve(ctx, req.Target)
	case "http":
		latency, _, err = HTTPAvailability(ctx, req.Target)
	default:
		target := req.Target
		if !strings.Contains(target, ":") {
			target = net.JoinHostPort(target, "443")
		}
		latency, err = TCPConnect(ctx, target, 8*time.Second)
	}

	latencyMs := int(latency.Milliseconds())
	status := "ok"
	errorMessage := ""
	if err != nil {
		status = "failed"
		errorMessage = err.Error()
	}

	check := storage.DiagnosticCheck{
		ID:        uuid.NewString(),
		Target:    req.Target,
		Mode:      req.Mode,
		Status:    status,
		LatencyMS: &latencyMs,
		Error:     errorMessage,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s.checks != nil {
		if createErr := s.checks.Create(ctx, check); createErr != nil {
			s.logs.Warn("diagnostics", fmt.Sprintf("failed to persist check: %v", createErr))
		}
	}
	return check, err
}

func (s *Service) History(ctx context.Context, limit int) ([]storage.DiagnosticCheck, error) {
	if s.checks == nil {
		return []storage.DiagnosticCheck{}, nil
	}
	return s.checks.History(ctx, limit)
}
