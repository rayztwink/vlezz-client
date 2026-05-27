package systemproxy

import (
	"context"
	"fmt"
	"time"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

type StateStore interface {
	Get(ctx context.Context) (storage.SystemProxyState, error)
	Save(ctx context.Context, state storage.SystemProxyState) error
}

type Service struct {
	store StateStore
	logs  *logs.Manager
}

var (
	readProxy              = readPlatformProxy
	writeProxy             = writePlatformProxy
	systemProxySupportedFn = platformSupported
)

type PlatformState struct {
	ProxyEnable   bool   `json:"proxyEnable"`
	ProxyServer   string `json:"proxyServer"`
	ProxyOverride string `json:"proxyOverride"`
}

type Status struct {
	Supported          bool   `json:"supported"`
	ProxyEnable        bool   `json:"proxyEnable"`
	ProxyServer        string `json:"proxyServer"`
	ProxyOverride      string `json:"proxyOverride"`
	EnabledByRayFlow   bool   `json:"enabledByRayflow"`
	CurrentProxyServer string `json:"currentProxyServer"`
	Error              string `json:"error,omitempty"`
}

func NewService(store StateStore, logManager *logs.Manager) *Service {
	return &Service{store: store, logs: logManager}
}

func (s *Service) Supported() bool {
	return systemProxySupportedFn()
}

func (s *Service) Status(ctx context.Context) Status {
	state, stateErr := s.store.Get(ctx)
	platform, platformErr := readProxy()
	status := Status{Supported: s.Supported()}
	if stateErr == nil {
		status.CurrentProxyServer = state.CurrentProxyServer
	}
	if platformErr != nil {
		status.Error = platformErr.Error()
		return status
	}
	status.ProxyEnable = platform.ProxyEnable
	status.ProxyServer = platform.ProxyServer
	status.ProxyOverride = platform.ProxyOverride
	if stateErr == nil && state.EnabledByRayFlow && state.CurrentProxyServer != "" && platform.ProxyEnable && platform.ProxyServer == state.CurrentProxyServer {
		status.EnabledByRayFlow = true
	}
	return status
}

func (s *Service) Enable(ctx context.Context, proxyServer string) error {
	if proxyServer == "" {
		return fmt.Errorf("proxy server is required")
	}
	current, err := readProxy()
	if err != nil {
		return err
	}
	if err := writeProxy(PlatformState{ProxyEnable: true, ProxyServer: proxyServer, ProxyOverride: "<local>"}); err != nil {
		return err
	}
	if err := s.store.Save(ctx, storage.SystemProxyState{
		EnabledByRayFlow:      true,
		PreviousProxyEnable:   current.ProxyEnable,
		PreviousProxyServer:   current.ProxyServer,
		PreviousProxyOverride: current.ProxyOverride,
		CurrentProxyServer:    proxyServer,
		UpdatedAt:             time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return err
	}
	s.logs.Info("system-proxy", fmt.Sprintf("enabled system proxy at %s", proxyServer))
	return nil
}

func (s *Service) Disable(ctx context.Context) error {
	return s.restoreOwnedProxy(ctx, "restored previous system proxy settings")
}

func (s *Service) RestoreOwnedIfCurrent(ctx context.Context) error {
	return s.restoreOwnedProxy(ctx, "restored RayFlow-owned system proxy settings")
}

func (s *Service) restoreOwnedProxy(ctx context.Context, logMessage string) error {
	state, err := s.store.Get(ctx)
	if err != nil {
		return err
	}
	if !state.EnabledByRayFlow {
		return nil
	}
	current, err := readProxy()
	if err != nil {
		return err
	}
	if state.CurrentProxyServer == "" || !current.ProxyEnable || current.ProxyServer != state.CurrentProxyServer {
		state.EnabledByRayFlow = false
		state.CurrentProxyServer = ""
		if saveErr := s.store.Save(ctx, state); saveErr != nil {
			return saveErr
		}
		s.logs.Warn("system-proxy", "skipped restore because Windows proxy is no longer owned by RayFlow")
		return nil
	}
	if err := writeProxy(PlatformState{
		ProxyEnable:   state.PreviousProxyEnable,
		ProxyServer:   state.PreviousProxyServer,
		ProxyOverride: state.PreviousProxyOverride,
	}); err != nil {
		return err
	}
	state.EnabledByRayFlow = false
	state.CurrentProxyServer = ""
	if err := s.store.Save(ctx, state); err != nil {
		return err
	}
	s.logs.Info("system-proxy", logMessage)
	return nil
}
