package storage

import (
	"context"
	"database/sql"
	"time"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) EnsureDefault(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO app_settings(id, theme, language, autostart, active_mode, default_core, local_proxy_port, updated_at) VALUES (1, 'system', 'system', 0, 'direct', 'sing-box', 2080, ?)`, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *SettingsRepository) Get(ctx context.Context) (AppSettings, error) {
	var s AppSettings
	var autostart int
	var enableSystemProxy int
	var tunEnabled int
	var tunAutoRoute int
	var tunStrictRoute int
	err := r.db.QueryRowContext(ctx, `SELECT id, theme, language, autostart, active_mode, default_core, local_proxy_port, sing_box_path, xray_path, zapret_path, enable_system_proxy_on_connect, preferred_network_mode, tun_enabled, tun_stack, tun_auto_route, tun_strict_route, updated_at FROM app_settings WHERE id = 1`).
		Scan(&s.ID, &s.Theme, &s.Language, &autostart, &s.ActiveMode, &s.DefaultCore, &s.LocalProxyPort, &s.SingBoxPath, &s.XrayPath, &s.ZapretPath, &enableSystemProxy, &s.PreferredNetworkMode, &tunEnabled, &s.TUNStack, &tunAutoRoute, &tunStrictRoute, &s.UpdatedAt)
	if err != nil {
		return AppSettings{}, err
	}
	s.Autostart = autostart == 1
	s.EnableSystemProxyOnConnect = enableSystemProxy == 1
	s.TUNEnabled = tunEnabled == 1
	s.TUNAutoRoute = tunAutoRoute == 1
	s.TUNStrictRoute = tunStrictRoute == 1
	return s, nil
}

func (r *SettingsRepository) Patch(ctx context.Context, s AppSettings) (AppSettings, error) {
	current, err := r.Get(ctx)
	if err != nil {
		return AppSettings{}, err
	}
	if s.Theme != "" {
		current.Theme = s.Theme
	}
	if s.Language != "" {
		current.Language = s.Language
	}
	if s.ActiveMode != "" {
		current.ActiveMode = s.ActiveMode
	}
	if s.DefaultCore != "" {
		current.DefaultCore = s.DefaultCore
	}
	if s.LocalProxyPort > 0 {
		current.LocalProxyPort = s.LocalProxyPort
	}
	if s.PreferredNetworkMode != "" {
		current.PreferredNetworkMode = s.PreferredNetworkMode
	}
	if s.TUNStack != "" {
		current.TUNStack = s.TUNStack
	}
	current.SingBoxPath = s.SingBoxPath
	current.XrayPath = s.XrayPath
	current.ZapretPath = s.ZapretPath
	current.Autostart = s.Autostart
	current.EnableSystemProxyOnConnect = s.EnableSystemProxyOnConnect
	current.TUNEnabled = s.TUNEnabled
	current.TUNAutoRoute = s.TUNAutoRoute
	current.TUNStrictRoute = s.TUNStrictRoute
	current.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	auto := 0
	if current.Autostart {
		auto = 1
	}
	enableSystemProxy := 0
	if current.EnableSystemProxyOnConnect {
		enableSystemProxy = 1
	}
	tunEnabled := 0
	if current.TUNEnabled {
		tunEnabled = 1
	}
	tunAutoRoute := 0
	if current.TUNAutoRoute {
		tunAutoRoute = 1
	}
	tunStrictRoute := 0
	if current.TUNStrictRoute {
		tunStrictRoute = 1
	}
	_, err = r.db.ExecContext(ctx, `UPDATE app_settings SET theme = ?, language = ?, autostart = ?, active_mode = ?, default_core = ?, local_proxy_port = ?, sing_box_path = ?, xray_path = ?, zapret_path = ?, enable_system_proxy_on_connect = ?, preferred_network_mode = ?, tun_enabled = ?, tun_stack = ?, tun_auto_route = ?, tun_strict_route = ?, updated_at = ? WHERE id = 1`,
		current.Theme, current.Language, auto, current.ActiveMode, current.DefaultCore, current.LocalProxyPort, current.SingBoxPath, current.XrayPath, current.ZapretPath, enableSystemProxy, current.PreferredNetworkMode, tunEnabled, current.TUNStack, tunAutoRoute, tunStrictRoute, current.UpdatedAt)
	return current, err
}
