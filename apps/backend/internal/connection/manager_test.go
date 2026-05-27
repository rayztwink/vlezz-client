package connection

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/systemproxy"
)

func TestConnectAndDisconnectUpdateState(t *testing.T) {
	ctx := context.Background()
	state := &fakeStateStore{}
	singBox := &fakeCoreClient{}
	manager := NewManager(
		fakeNodeStore{},
		fakeSettingsStore{},
		state,
		logs.NewManager(nil, zerolog.Nop()),
		nil,
		nil,
		singBox,
		&fakeCoreClient{},
	)

	status, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeLocalProxy})
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if status.Status != "connected" {
		t.Fatalf("expected connected status, got %s", status.Status)
	}
	if !singBox.connected {
		t.Fatal("expected fake core to be connected")
	}

	status, err = manager.Disconnect(ctx)
	if err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if status.Status != "disconnected" {
		t.Fatalf("expected disconnected status, got %s", status.Status)
	}
}

func TestLocalProxyDoesNotEnableSystemProxy(t *testing.T) {
	ctx := context.Background()
	system := &fakeSystemProxyController{}
	manager := NewManager(
		fakeNodeStore{},
		fakeSettingsStore{},
		&fakeStateStore{},
		logs.NewManager(nil, zerolog.Nop()),
		system,
		nil,
		&fakeCoreClient{},
		&fakeCoreClient{},
	)

	if _, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeLocalProxy}); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if system.enableCalls != 0 {
		t.Fatalf("expected local_proxy not to enable system proxy, got %d calls", system.enableCalls)
	}
}

func TestSystemProxyEnablesAndRestoresSystemProxy(t *testing.T) {
	ctx := context.Background()
	system := &fakeSystemProxyController{}
	manager := NewManager(
		fakeNodeStore{},
		fakeSettingsStore{},
		&fakeStateStore{},
		logs.NewManager(nil, zerolog.Nop()),
		system,
		nil,
		&fakeCoreClient{},
		&fakeCoreClient{},
	)

	if _, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeSystemProxy}); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if system.enableCalls != 1 {
		t.Fatalf("expected system_proxy to enable system proxy once, got %d calls", system.enableCalls)
	}
	if _, err := manager.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect returned error: %v", err)
	}
	if system.disableCalls != 2 {
		t.Fatalf("expected pre-connect cleanup and disconnect restore, got %d calls", system.disableCalls)
	}
}

func TestTUNValidation(t *testing.T) {
	ctx := context.Background()
	t.Run("rejects xray", func(t *testing.T) {
		manager := NewManager(fakeNodeStore{}, fakeSettingsStore{}, &fakeStateStore{}, logs.NewManager(nil, zerolog.Nop()), nil, nil, &fakeCoreClient{}, &fakeCoreClient{})
		_, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "xray", NetworkMode: core.NetworkModeTUN})
		if err == nil || err.Error() != "TUN mode is currently supported only with sing-box" {
			t.Fatalf("expected xray TUN error, got %v", err)
		}
	})

	t.Run("rejects disabled setting", func(t *testing.T) {
		manager := NewManager(fakeNodeStore{}, fakeSettingsStore{}, &fakeStateStore{}, logs.NewManager(nil, zerolog.Nop()), nil, nil, &fakeCoreClient{}, &fakeCoreClient{})
		_, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeTUN})
		if err == nil || err.Error() != "TUN mode is disabled in settings" {
			t.Fatalf("expected disabled TUN error, got %v", err)
		}
	})

	t.Run("rejects non admin", func(t *testing.T) {
		restoreAdmin := setAdminForTest(t, false)
		defer restoreAdmin()
		manager := NewManager(
			fakeNodeStore{},
			fakeSettingsStore{settings: storage.AppSettings{DefaultCore: "sing-box", LocalProxyPort: 2080, PreferredNetworkMode: core.NetworkModeLocalProxy, TUNEnabled: true, TUNStack: "system", TUNAutoRoute: true, TUNStrictRoute: true}},
			&fakeStateStore{},
			logs.NewManager(nil, zerolog.Nop()),
			nil,
			nil,
			&fakeCoreClient{},
			&fakeCoreClient{},
		)
		_, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeTUN})
		if err == nil || err.Error() != "admin privileges required for TUN mode" {
			t.Fatalf("expected admin TUN error, got %v", err)
		}
	})

	t.Run("accepts sing-box admin enabled", func(t *testing.T) {
		restoreAdmin := setAdminForTest(t, true)
		defer restoreAdmin()
		singBox := &fakeCoreClient{}
		manager := NewManager(
			fakeNodeStore{},
			fakeSettingsStore{settings: storage.AppSettings{DefaultCore: "sing-box", LocalProxyPort: 2080, PreferredNetworkMode: core.NetworkModeLocalProxy, TUNEnabled: true, TUNStack: "system", TUNAutoRoute: true, TUNStrictRoute: true}},
			&fakeStateStore{},
			logs.NewManager(nil, zerolog.Nop()),
			nil,
			nil,
			singBox,
			&fakeCoreClient{},
		)
		if _, err := manager.Connect(ctx, ConnectRequest{NodeID: "node-1", Core: "sing-box", NetworkMode: core.NetworkModeTUN}); err != nil {
			t.Fatalf("Connect returned error: %v", err)
		}
		if singBox.options.NetworkMode != core.NetworkModeTUN {
			t.Fatalf("expected TUN options, got %#v", singBox.options)
		}
	})
}

func TestSystemProxyServerUsesSocksForXray(t *testing.T) {
	if got := systemProxyServer("xray", "127.0.0.1:2080"); got != "socks=127.0.0.1:2080" {
		t.Fatalf("expected xray system proxy to use SOCKS scheme, got %q", got)
	}
	if got := systemProxyServer("sing-box", "127.0.0.1:2080"); got != "127.0.0.1:2080" {
		t.Fatalf("expected sing-box system proxy to keep mixed proxy address, got %q", got)
	}
}

type fakeNodeStore struct{}

func (fakeNodeStore) Get(ctx context.Context, id string) (storage.Node, error) {
	return storage.Node{
		ID:        id,
		Name:      "Test Node",
		Protocol:  "vless",
		Address:   "example.com",
		Port:      443,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Security:  "reality",
		Transport: "tcp",
	}, nil
}

type fakeSettingsStore struct {
	settings storage.AppSettings
}

func (s fakeSettingsStore) Get(ctx context.Context) (storage.AppSettings, error) {
	if s.settings.DefaultCore != "" {
		return s.settings, nil
	}
	return storage.AppSettings{
		ActiveMode:           "direct",
		DefaultCore:          "sing-box",
		LocalProxyPort:       2080,
		PreferredNetworkMode: core.NetworkModeLocalProxy,
		TUNStack:             "system",
		TUNAutoRoute:         true,
		TUNStrictRoute:       true,
	}, nil
}

type fakeStateStore struct {
	state storage.ConnectionState
}

func (s *fakeStateStore) EnsureDefault(ctx context.Context) error {
	if s.state.ID == 0 {
		s.state = storage.ConnectionState{ID: 1, Status: "disconnected", SelectedCore: "sing-box", NetworkMode: core.NetworkModeLocalProxy, LocalProxyAddress: "127.0.0.1:2080"}
	}
	return nil
}

func (s *fakeStateStore) Get(ctx context.Context) (storage.ConnectionState, error) {
	_ = s.EnsureDefault(ctx)
	return s.state, nil
}

func (s *fakeStateStore) Save(ctx context.Context, state storage.ConnectionState) error {
	state.ID = 1
	s.state = state
	return nil
}

type fakeCoreClient struct {
	connected bool
	options   core.ConnectOptions
}

func (c *fakeCoreClient) ConnectWithOptions(ctx context.Context, node storage.Node, options core.ConnectOptions) error {
	c.connected = true
	c.options = options
	return nil
}

func (c *fakeCoreClient) Disconnect(ctx context.Context) error {
	c.connected = false
	return nil
}

func (c *fakeCoreClient) Status() process.Snapshot {
	if c.connected {
		return process.Snapshot{ID: "sing-box", Name: "sing-box", Status: process.StatusRunning, PID: 1234}
	}
	return process.Snapshot{ID: "sing-box", Name: "sing-box", Status: process.StatusStopped}
}

type fakeSystemProxyController struct {
	enableCalls  int
	disableCalls int
}

func (s *fakeSystemProxyController) Status(ctx context.Context) systemproxy.Status {
	return systemproxy.Status{Supported: true}
}

func (s *fakeSystemProxyController) Enable(ctx context.Context, proxyServer string) error {
	s.enableCalls++
	return nil
}

func (s *fakeSystemProxyController) Disable(ctx context.Context) error {
	s.disableCalls++
	return nil
}

func setAdminForTest(t *testing.T, value bool) func() {
	t.Helper()
	original := isAdmin
	isAdmin = func() bool {
		return value
	}
	return func() {
		isAdmin = original
	}
}
