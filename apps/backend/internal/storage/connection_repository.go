package storage

import (
	"context"
	"database/sql"
	"time"
)

type ConnectionRepository struct {
	db *sql.DB
}

func NewConnectionRepository(db *sql.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) EnsureDefault(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO connection_state(id, active_mode, selected_core, network_mode, local_proxy_address, status, updated_at) VALUES (1, 'direct', 'sing-box', 'local_proxy', '127.0.0.1:2080', 'disconnected', ?)`, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *ConnectionRepository) Get(ctx context.Context) (ConnectionState, error) {
	var state ConnectionState
	var nodeID sql.NullString
	var nodeName sql.NullString
	var lastError sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, active_mode, selected_node_id, selected_node_name, selected_core, network_mode, local_proxy_address, status, last_error, updated_at FROM connection_state WHERE id = 1`).
		Scan(&state.ID, &state.ActiveMode, &nodeID, &nodeName, &state.SelectedCore, &state.NetworkMode, &state.LocalProxyAddress, &state.Status, &lastError, &state.UpdatedAt)
	if err != nil {
		return ConnectionState{}, err
	}
	if nodeID.Valid {
		state.SelectedNodeID = nodeID.String
	}
	if nodeName.Valid {
		state.SelectedNodeName = nodeName.String
	}
	if lastError.Valid {
		state.LastError = lastError.String
	}
	return state, nil
}

func (r *ConnectionRepository) Save(ctx context.Context, state ConnectionState) error {
	if state.ActiveMode == "" {
		state.ActiveMode = "direct"
	}
	if state.SelectedCore == "" {
		state.SelectedCore = "sing-box"
	}
	if state.NetworkMode == "" {
		state.NetworkMode = "local_proxy"
	}
	if state.LocalProxyAddress == "" {
		state.LocalProxyAddress = "127.0.0.1:2080"
	}
	if state.Status == "" {
		state.Status = "disconnected"
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `INSERT INTO connection_state(id, active_mode, selected_node_id, selected_node_name, selected_core, network_mode, local_proxy_address, status, last_error, updated_at)
		VALUES (1, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, NULLIF(?, ''), ?)
		ON CONFLICT(id) DO UPDATE SET active_mode=excluded.active_mode, selected_node_id=excluded.selected_node_id, selected_node_name=excluded.selected_node_name, selected_core=excluded.selected_core, network_mode=excluded.network_mode, local_proxy_address=excluded.local_proxy_address, status=excluded.status, last_error=excluded.last_error, updated_at=excluded.updated_at`,
		state.ActiveMode, state.SelectedNodeID, state.SelectedNodeName, state.SelectedCore, state.NetworkMode, state.LocalProxyAddress, state.Status, state.LastError, state.UpdatedAt)
	return err
}
