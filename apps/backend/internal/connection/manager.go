package connection

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/diagnostics"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/platform"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/systemproxy"
)

type NodeStore interface {
	Get(ctx context.Context, id string) (storage.Node, error)
}

type SettingsStore interface {
	Get(ctx context.Context) (storage.AppSettings, error)
}

type StateStore interface {
	EnsureDefault(ctx context.Context) error
	Get(ctx context.Context) (storage.ConnectionState, error)
	Save(ctx context.Context, state storage.ConnectionState) error
}

type CoreClient interface {
	ConnectWithOptions(ctx context.Context, node storage.Node, options core.ConnectOptions) error
	Disconnect(ctx context.Context) error
	Status() process.Snapshot
}

type SystemProxyController interface {
	Status(ctx context.Context) systemproxy.Status
	Enable(ctx context.Context, proxyServer string) error
	Disable(ctx context.Context) error
}

var isAdmin = platform.IsAdmin

type Manager struct {
	nodes      NodeStore
	settings   SettingsStore
	state      StateStore
	logs       *logs.Manager
	system     SystemProxyController
	diagnostic *diagnostics.Service
	singBox    CoreClient
	xray       CoreClient
}

type ConnectRequest struct {
	NodeID      string `json:"nodeId"`
	Core        string `json:"core"`
	NetworkMode string `json:"networkMode"`
}

type Status struct {
	storage.ConnectionState
	ProcessStatus process.Snapshot   `json:"processStatus"`
	SystemProxy   systemproxy.Status `json:"systemProxy"`
}

type Report struct {
	Status Status             `json:"status"`
	Logs   []storage.LogEntry `json:"logs"`
	Checks map[string]any     `json:"checks"`
}

func NewManager(nodes NodeStore, settings SettingsStore, state StateStore, logManager *logs.Manager, system SystemProxyController, diagnostic *diagnostics.Service, singBox CoreClient, xray CoreClient) *Manager {
	return &Manager{
		nodes:      nodes,
		settings:   settings,
		state:      state,
		logs:       logManager,
		system:     system,
		diagnostic: diagnostic,
		singBox:    singBox,
		xray:       xray,
	}
}

func (m *Manager) Connect(ctx context.Context, req ConnectRequest) (Status, error) {
	settings, err := m.settings.Get(ctx)
	if err != nil {
		return Status{}, err
	}
	node, err := m.nodes.Get(ctx, req.NodeID)
	if err != nil {
		return Status{}, err
	}
	selectedCore := normalizeCore(firstNonEmpty(req.Core, settings.DefaultCore, "sing-box"))
	networkMode := normalizeNetworkMode(firstNonEmpty(req.NetworkMode, settings.PreferredNetworkMode, core.NetworkModeLocalProxy))
	if networkMode == core.NetworkModeTUN && selectedCore != "sing-box" {
		return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, "TUN mode is currently supported only with sing-box")
	}
	if networkMode == core.NetworkModeTUN && !settings.TUNEnabled {
		return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, "TUN mode is disabled in settings")
	}
	if networkMode == core.NetworkModeTUN && !isAdmin() {
		return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, "admin privileges required for TUN mode")
	}

	localProxyAddress := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", settings.LocalProxyPort))
	connecting := storage.ConnectionState{
		ActiveMode:        "proxy",
		SelectedNodeID:    node.ID,
		SelectedNodeName:  node.Name,
		SelectedCore:      selectedCore,
		NetworkMode:       networkMode,
		LocalProxyAddress: localProxyAddress,
		Status:            "connecting",
	}
	_ = m.state.Save(ctx, connecting)

	if err := m.stopCores(ctx); err != nil {
		return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, err.Error())
	}
	if m.system != nil {
		_ = m.system.Disable(ctx)
	}

	options := core.ConnectOptions{
		NetworkMode:    networkMode,
		LocalProxyPort: settings.LocalProxyPort,
		TUNStack:       settings.TUNStack,
		TUNAutoRoute:   settings.TUNAutoRoute,
		TUNStrictRoute: settings.TUNStrictRoute,
	}

	client := m.client(selectedCore)
	if err := client.ConnectWithOptions(ctx, node, options); err != nil {
		return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, err.Error())
	}

	if networkMode == core.NetworkModeSystemProxy {
		if m.system == nil {
			_ = client.Disconnect(ctx)
			return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, "system proxy service is not available")
		}
		if err := m.system.Enable(ctx, systemProxyServer(selectedCore, localProxyAddress)); err != nil {
			_ = client.Disconnect(ctx)
			return m.fail(ctx, node, selectedCore, networkMode, settings.LocalProxyPort, err.Error())
		}
		networkMode = core.NetworkModeSystemProxy
	}

	connected := storage.ConnectionState{
		ActiveMode:        "proxy",
		SelectedNodeID:    node.ID,
		SelectedNodeName:  node.Name,
		SelectedCore:      selectedCore,
		NetworkMode:       networkMode,
		LocalProxyAddress: localProxyAddress,
		Status:            "connected",
	}
	if err := m.state.Save(ctx, connected); err != nil {
		return Status{}, err
	}
	m.logs.Info("connection", fmt.Sprintf("connected node %s via %s/%s at %s", node.Name, selectedCore, networkMode, localProxyAddress))
	return m.Status(ctx)
}

func (m *Manager) Disconnect(ctx context.Context) (Status, error) {
	if m.system != nil {
		if err := m.system.Disable(ctx); err != nil {
			m.logs.Warn("connection", fmt.Sprintf("failed to restore system proxy: %v", err))
		}
	}
	if err := m.stopCores(ctx); err != nil {
		return Status{}, err
	}
	state, _ := m.state.Get(ctx)
	state.ActiveMode = "direct"
	state.SelectedNodeID = ""
	state.SelectedNodeName = ""
	state.Status = "disconnected"
	state.LastError = ""
	if state.LocalProxyAddress == "" {
		state.LocalProxyAddress = "127.0.0.1:2080"
	}
	if err := m.state.Save(ctx, state); err != nil {
		return Status{}, err
	}
	m.logs.Info("connection", "disconnected")
	return m.Status(ctx)
}

func (m *Manager) Status(ctx context.Context) (Status, error) {
	if err := m.state.EnsureDefault(ctx); err != nil {
		return Status{}, err
	}
	state, err := m.state.Get(ctx)
	if err != nil {
		return Status{}, err
	}
	snapshot := m.client(normalizeCore(state.SelectedCore)).Status()
	status := Status{ConnectionState: state, ProcessStatus: snapshot}
	if m.system != nil {
		status.SystemProxy = m.system.Status(ctx)
	}
	if state.Status == "connected" {
		switch snapshot.Status {
		case process.StatusFailed:
			status.Status = "failed"
			if status.LastError == "" {
				status.LastError = snapshot.Error
			}
			m.restoreSystemProxyIfNeeded(ctx, state)
			if m.system != nil {
				status.SystemProxy = m.system.Status(ctx)
			}
			state.Status = status.Status
			state.LastError = status.LastError
			_ = m.state.Save(ctx, state)
		case process.StatusStopped:
			status.Status = "disconnected"
			status.ActiveMode = "direct"
			m.restoreSystemProxyIfNeeded(ctx, state)
			if m.system != nil {
				status.SystemProxy = m.system.Status(ctx)
			}
			state.Status = "disconnected"
			state.ActiveMode = "direct"
			_ = m.state.Save(ctx, state)
		}
	}
	return status, nil
}

func (m *Manager) Report(ctx context.Context) (Report, error) {
	status, err := m.Status(ctx)
	if err != nil {
		return Report{}, err
	}
	entries, err := m.logs.ListBySource(ctx, "", 50)
	if err != nil {
		return Report{}, err
	}
	report := Report{Status: status, Logs: entries, Checks: map[string]any{}}
	if status.SelectedNodeID != "" {
		if node, err := m.nodes.Get(ctx, status.SelectedNodeID); err == nil {
			target := net.JoinHostPort(node.Address, fmt.Sprintf("%d", node.Port))
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			check, checkErr := m.diagnostic.Check(checkCtx, diagnostics.CheckRequest{Target: target, Mode: status.NetworkMode, Type: "tcp"})
			report.Checks["tcp"] = check
			if checkErr != nil {
				report.Checks["tcpError"] = checkErr.Error()
			}
		}
	}
	return report, nil
}

func (m *Manager) fail(ctx context.Context, node storage.Node, selectedCore string, networkMode string, localProxyPort int, message string) (Status, error) {
	state := storage.ConnectionState{
		ActiveMode:        "proxy",
		SelectedNodeID:    node.ID,
		SelectedNodeName:  node.Name,
		SelectedCore:      selectedCore,
		NetworkMode:       networkMode,
		LocalProxyAddress: net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", localProxyPort)),
		Status:            "failed",
		LastError:         message,
	}
	_ = m.state.Save(ctx, state)
	m.logs.Error("connection", message)
	status, _ := m.Status(ctx)
	return status, fmt.Errorf("%s", message)
}

func (m *Manager) stopCores(ctx context.Context) error {
	if err := m.singBox.Disconnect(ctx); err != nil {
		return err
	}
	if err := m.xray.Disconnect(ctx); err != nil {
		return err
	}
	return nil
}

func (m *Manager) restoreSystemProxyIfNeeded(ctx context.Context, state storage.ConnectionState) {
	if m.system != nil && state.NetworkMode == core.NetworkModeSystemProxy {
		if err := m.system.Disable(ctx); err != nil {
			m.logs.Warn("connection", fmt.Sprintf("failed to restore system proxy: %v", err))
		}
	}
}

func (m *Manager) client(coreName string) CoreClient {
	if normalizeCore(coreName) == "xray" {
		return m.xray
	}
	return m.singBox
}

func normalizeCore(value string) string {
	switch strings.ToLower(value) {
	case "xray", "xray-core":
		return "xray"
	default:
		return "sing-box"
	}
}

func normalizeNetworkMode(value string) string {
	switch strings.ToLower(value) {
	case core.NetworkModeSystemProxy, "system":
		return core.NetworkModeSystemProxy
	case core.NetworkModeTUN:
		return core.NetworkModeTUN
	default:
		return core.NetworkModeLocalProxy
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func systemProxyServer(selectedCore string, localProxyAddress string) string {
	if normalizeCore(selectedCore) == "xray" {
		return "socks=" + localProxyAddress
	}
	return localProxyAddress
}
