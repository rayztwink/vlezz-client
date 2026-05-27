package singbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rayflow/rayflow-client/apps/backend/internal/config"
	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/platform"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	goruntime "runtime"
	"strings"
)

const processID = "sing-box"

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
		return fmt.Errorf("sing-box binary path is not configured")
	}
	if options.LocalProxyPort <= 0 {
		options.LocalProxyPort = runtime.localProxyPort
	}
	options = options.Normalized()
	data, err := GenerateConfigWithOptions(node, options)
	if err != nil {
		return err
	}
	configPath := filepath.Join(c.cfg.ConfigsDir, "sing-box-generated.json")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return err
	}
	binaryPath := runtime.binaryPath
	args := []string{"run", "-c", configPath}

	if options.NetworkMode == core.NetworkModeTUN && (goruntime.GOOS == "linux" || goruntime.GOOS == "darwin") {
		if !platform.IsAdmin() {
			if goruntime.GOOS == "linux" {
				args = append([]string{binaryPath}, args...)
				binaryPath = "/usr/bin/pkexec"
			} else if goruntime.GOOS == "darwin" {
				shellCmd := fmt.Sprintf("%q %s", binaryPath, strings.Join(args, " "))
				args = []string{"-e", fmt.Sprintf("do shell script %q with administrator privileges", shellCmd)}
				binaryPath = "/usr/bin/osascript"
			}
		}
	}

	return c.processes.Restart(ctx, process.Spec{
		ID:         processID,
		Name:       "sing-box",
		BinaryPath: binaryPath,
		Args:       args,
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
	result := runtimeSettings{binaryPath: c.cfg.SingBoxPath, localProxyPort: c.cfg.LocalProxyPort}
	if c.settings == nil {
		return result
	}
	settings, err := c.settings.Get(ctx)
	if err != nil {
		c.logs.Warn(processID, fmt.Sprintf("failed to load settings: %v", err))
		return result
	}
	if settings.SingBoxPath != "" {
		result.binaryPath = settings.SingBoxPath
	}
	if settings.LocalProxyPort > 0 {
		result.localProxyPort = settings.LocalProxyPort
	}
	return result
}
