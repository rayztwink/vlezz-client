package systemproxy

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

func TestDisableRestoresOnlyWhenCurrentProxyIsRayFlowOwned(t *testing.T) {
	ctx := context.Background()
	store := &fakeSystemProxyStore{}
	service := NewService(store, logs.NewManager(nil, zerolog.Nop()))
	var platformState PlatformState
	writes := 0
	restorePlatformForTest(t, &platformState, &writes)

	platformState = PlatformState{ProxyEnable: true, ProxyServer: "socks=127.0.0.1:10808", ProxyOverride: "<local>"}
	if err := service.Enable(ctx, "127.0.0.1:2080"); err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	platformState = PlatformState{ProxyEnable: true, ProxyServer: "socks=127.0.0.1:10808", ProxyOverride: "<local>"}
	if err := service.Disable(ctx); err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}

	if platformState.ProxyServer != "socks=127.0.0.1:10808" {
		t.Fatalf("expected external proxy to stay untouched, got %q", platformState.ProxyServer)
	}
	if store.state.EnabledByRayFlow {
		t.Fatal("expected RayFlow ownership to be cleared")
	}
}

func TestDisableRestoresPreviousProxyWhenStillOwned(t *testing.T) {
	ctx := context.Background()
	store := &fakeSystemProxyStore{}
	service := NewService(store, logs.NewManager(nil, zerolog.Nop()))
	platformState := PlatformState{ProxyEnable: true, ProxyServer: "socks=127.0.0.1:10808", ProxyOverride: "<local>"}
	writes := 0
	restorePlatformForTest(t, &platformState, &writes)

	if err := service.Enable(ctx, "127.0.0.1:2080"); err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	if err := service.Disable(ctx); err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}

	if !platformState.ProxyEnable || platformState.ProxyServer != "socks=127.0.0.1:10808" {
		t.Fatalf("expected previous proxy restored, got %#v", platformState)
	}
	if store.state.EnabledByRayFlow {
		t.Fatal("expected RayFlow ownership to be cleared")
	}
	if writes != 2 {
		t.Fatalf("expected enable and disable writes, got %d", writes)
	}
}

func TestStatusReportsExternalProxyWhenCurrentProxyChanged(t *testing.T) {
	ctx := context.Background()
	store := &fakeSystemProxyStore{}
	service := NewService(store, logs.NewManager(nil, zerolog.Nop()))
	platformState := PlatformState{ProxyEnable: true, ProxyServer: "socks=127.0.0.1:10808", ProxyOverride: "<local>"}
	writes := 0
	restorePlatformForTest(t, &platformState, &writes)

	if err := service.Enable(ctx, "127.0.0.1:2080"); err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	platformState = PlatformState{ProxyEnable: true, ProxyServer: "socks=127.0.0.1:10808", ProxyOverride: "<local>"}

	status := service.Status(ctx)
	if status.EnabledByRayFlow {
		t.Fatal("expected changed proxy to be reported as external")
	}
	if status.ProxyServer != "socks=127.0.0.1:10808" {
		t.Fatalf("expected current external proxy, got %q", status.ProxyServer)
	}
}

func restorePlatformForTest(t *testing.T, platformState *PlatformState, writes *int) {
	t.Helper()
	originalRead := readProxy
	originalWrite := writeProxy
	originalSupported := systemProxySupportedFn
	readProxy = func() (PlatformState, error) {
		return *platformState, nil
	}
	writeProxy = func(state PlatformState) error {
		*writes = *writes + 1
		*platformState = state
		return nil
	}
	systemProxySupportedFn = func() bool {
		return true
	}
	t.Cleanup(func() {
		readProxy = originalRead
		writeProxy = originalWrite
		systemProxySupportedFn = originalSupported
	})
}

type fakeSystemProxyStore struct {
	state storage.SystemProxyState
}

func (s *fakeSystemProxyStore) Get(ctx context.Context) (storage.SystemProxyState, error) {
	return s.state, nil
}

func (s *fakeSystemProxyStore) Save(ctx context.Context, state storage.SystemProxyState) error {
	state.ID = 1
	s.state = state
	return nil
}
