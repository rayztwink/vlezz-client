package zapret

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rayflow/rayflow-client/apps/backend/internal/config"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

const processID = "zapret"

type PresetStore interface {
	Get(ctx context.Context, id string) (storage.ZapretPreset, error)
	MarkActive(ctx context.Context, id string) error
}

type Client struct {
	cfg       config.AppConfig
	processes *process.Manager
	logs      *logs.Manager
	presets   PresetStore
	settings  settingsStore
}

type settingsStore interface {
	Get(ctx context.Context) (storage.AppSettings, error)
}

func NewClient(cfg config.AppConfig, manager *process.Manager, logManager *logs.Manager, presets PresetStore, settings settingsStore) *Client {
	return &Client{cfg: cfg, processes: manager, logs: logManager, presets: presets, settings: settings}
}

func (c *Client) StartPreset(ctx context.Context, id string) error {
	preset, err := c.presets.Get(ctx, id)
	if err != nil {
		return err
	}
	command, err := ParsePresetCommand(preset.Command)
	if err != nil {
		return err
	}
	if path := c.zapretPath(ctx); path != "" {
		command.Executable = path
	}
	if !filepath.IsAbs(command.Executable) {
		return fmt.Errorf("zapret executable path must be absolute")
	}
	if err := c.processes.Restart(ctx, process.Spec{
		ID:         processID,
		Name:       "zapret",
		BinaryPath: command.Executable,
		Args:       command.Args,
	}); err != nil {
		return err
	}
	return c.presets.MarkActive(ctx, id)
}

func (c *Client) Stop(ctx context.Context) error {
	if err := c.processes.Stop(ctx, processID); err != nil {
		return err
	}
	return c.presets.MarkActive(ctx, "")
}

func (c *Client) Status() process.Snapshot {
	return c.processes.Status(processID)
}

func (c *Client) zapretPath(ctx context.Context) string {
	path := c.cfg.ZapretPath
	if c.settings == nil {
		return path
	}
	settings, err := c.settings.Get(ctx)
	if err != nil {
		c.logs.Warn(processID, fmt.Sprintf("failed to load settings: %v", err))
		return path
	}
	if settings.ZapretPath != "" {
		return settings.ZapretPath
	}
	return path
}
