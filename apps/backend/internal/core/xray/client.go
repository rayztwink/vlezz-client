package xray

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rayflow/rayflow-client/apps/backend/internal/config"
	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

const processID = "xray-core"

type Client struct {
	cfg       config.AppConfig
	processes *process.Manager
	logs      *logs.Manager
	settings  settingsStore
}

type settingsStore interface {
	Get(ctx context.Context) (storage.AppSettings, error)
}

func NewClient(cfg config.AppConfig, manager *process.Manager, logManager *logs.Manager, settings settingsStore) *Client {
	return &Client{cfg: cfg, processes: manager, logs: logManager, settings: settings}
}

func (c *Client) Connect(ctx context.Context, node storage.Node) error {
	return c.ConnectWithOptions(ctx, node, core.ConnectOptions{NetworkMode: core.NetworkModeLocalProxy})
}

func (c *Client) ConnectWithOptions(ctx context.Context, node storage.Node, options core.ConnectOptions) error {
	runtime := c.runtimeSettings(ctx)
	if runtime.binaryPath == "" {
		return fmt.Errorf("xray-core binary path is not configured")
	}
	if options.LocalProxyPort <= 0 {
		options.LocalProxyPort = runtime.localProxyPort
	}
	options = options.Normalized()
	if options.NetworkMode == core.NetworkModeTUN {
		return fmt.Errorf("TUN mode is currently supported only with sing-box")
	}
	data, err := GenerateConfig(node, options.LocalProxyPort)
	if err != nil {
		return err
	}
	configPath := filepath.Join(c.cfg.ConfigsDir, "xray-generated.json")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return err
	}
	return c.processes.Restart(ctx, process.Spec{
		ID:         processID,
		Name:       "xray-core",
		BinaryPath: runtime.binaryPath,
		Args:       []string{"run", "-config", configPath},
	})
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.processes.Stop(ctx, processID)
}

func (c *Client) Status() process.Snapshot {
	return c.processes.Status(processID)
}

type runtimeSettings struct {
	binaryPath     string
	localProxyPort int
}

func (c *Client) runtimeSettings(ctx context.Context) runtimeSettings {
	result := runtimeSettings{binaryPath: c.cfg.XrayPath, localProxyPort: c.cfg.LocalProxyPort}
	if c.settings == nil {
		return result
	}
	settings, err := c.settings.Get(ctx)
	if err != nil {
		c.logs.Warn(processID, fmt.Sprintf("failed to load settings: %v", err))
		return result
	}
	if settings.XrayPath != "" {
		result.binaryPath = settings.XrayPath
	}
	if settings.LocalProxyPort > 0 {
		result.localProxyPort = settings.LocalProxyPort
	}
	return result
}
