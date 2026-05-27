package logs

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

type Sink interface {
	Create(ctx context.Context, entry storage.LogEntry) error
	ListBySource(ctx context.Context, source string, limit int) ([]storage.LogEntry, error)
}

type Manager struct {
	sink   Sink
	logger zerolog.Logger
}

var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

func NewManager(sink Sink, logger zerolog.Logger) *Manager {
	return &Manager{sink: sink, logger: logger}
}

func (m *Manager) Info(source string, message string) {
	m.Record(context.Background(), source, "info", message)
}

func (m *Manager) Warn(source string, message string) {
	m.Record(context.Background(), source, "warn", message)
}

func (m *Manager) Error(source string, message string) {
	m.Record(context.Background(), source, "error", message)
}

func (m *Manager) Record(ctx context.Context, source string, level string, message string) {
	safeMessage := MaskSecrets(message)
	entry := storage.LogEntry{
		ID:        uuid.NewString(),
		Source:    source,
		Level:     level,
		Message:   safeMessage,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if m.sink != nil {
		if err := m.sink.Create(ctx, entry); err != nil {
			m.logger.Warn().Err(err).Msg("failed to persist log entry")
		}
	}

	switch strings.ToLower(level) {
	case "error":
		m.logger.Error().Str("source", source).Msg(safeMessage)
	case "warn", "warning":
		m.logger.Warn().Str("source", source).Msg(safeMessage)
	default:
		m.logger.Info().Str("source", source).Msg(safeMessage)
	}
}

func (m *Manager) ListBySource(ctx context.Context, source string, limit int) ([]storage.LogEntry, error) {
	if m.sink == nil {
		return []storage.LogEntry{}, nil
	}
	return m.sink.ListBySource(ctx, source, limit)
}

func MaskSecrets(message string) string {
	masked := uuidPattern.ReplaceAllStringFunc(message, func(match string) string {
		if len(match) <= 8 {
			return "***"
		}
		return match[:8] + "-****-****-****-************"
	})

	for _, scheme := range []string{"vless://", "vmess://", "trojan://", "ss://"} {
		if strings.Contains(masked, scheme) {
			masked = maskURLScheme(masked, scheme)
		}
	}
	return masked
}

func maskURLScheme(input string, scheme string) string {
	idx := strings.Index(input, scheme)
	if idx < 0 {
		return input
	}
	end := len(input)
	for i := idx; i < len(input); i++ {
		if input[i] == ' ' || input[i] == '\n' || input[i] == '\t' {
			end = i
			break
		}
	}
	return input[:idx] + scheme + "***" + input[end:]
}
