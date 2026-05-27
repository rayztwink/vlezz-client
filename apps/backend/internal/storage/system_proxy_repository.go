package storage

import (
	"context"
	"database/sql"
	"time"
)

type SystemProxyRepository struct {
	db *sql.DB
}

func NewSystemProxyRepository(db *sql.DB) *SystemProxyRepository {
	return &SystemProxyRepository{db: db}
}

func (r *SystemProxyRepository) EnsureDefault(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO system_proxy_state(id, enabled_by_rayflow, previous_proxy_enable, previous_proxy_server, previous_proxy_override, current_proxy_server, updated_at) VALUES (1, 0, 0, '', '', '', ?)`, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *SystemProxyRepository) Get(ctx context.Context) (SystemProxyState, error) {
	var state SystemProxyState
	var enabledByRayFlow int
	var previousEnabled int
	err := r.db.QueryRowContext(ctx, `SELECT id, enabled_by_rayflow, previous_proxy_enable, previous_proxy_server, previous_proxy_override, current_proxy_server, updated_at FROM system_proxy_state WHERE id = 1`).
		Scan(&state.ID, &enabledByRayFlow, &previousEnabled, &state.PreviousProxyServer, &state.PreviousProxyOverride, &state.CurrentProxyServer, &state.UpdatedAt)
	if err != nil {
		return SystemProxyState{}, err
	}
	state.EnabledByRayFlow = enabledByRayFlow == 1
	state.PreviousProxyEnable = previousEnabled == 1
	return state, nil
}

func (r *SystemProxyRepository) Save(ctx context.Context, state SystemProxyState) error {
	enabledByRayFlow := 0
	if state.EnabledByRayFlow {
		enabledByRayFlow = 1
	}
	previousEnabled := 0
	if state.PreviousProxyEnable {
		previousEnabled = 1
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO system_proxy_state(id, enabled_by_rayflow, previous_proxy_enable, previous_proxy_server, previous_proxy_override, current_proxy_server, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET enabled_by_rayflow=excluded.enabled_by_rayflow, previous_proxy_enable=excluded.previous_proxy_enable, previous_proxy_server=excluded.previous_proxy_server, previous_proxy_override=excluded.previous_proxy_override, current_proxy_server=excluded.current_proxy_server, updated_at=excluded.updated_at`,
		enabledByRayFlow, previousEnabled, state.PreviousProxyServer, state.PreviousProxyOverride, state.CurrentProxyServer, state.UpdatedAt)
	return err
}
