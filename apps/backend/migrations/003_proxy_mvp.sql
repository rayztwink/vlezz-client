ALTER TABLE app_settings ADD COLUMN enable_system_proxy_on_connect INTEGER NOT NULL DEFAULT 0;
ALTER TABLE app_settings ADD COLUMN preferred_network_mode TEXT NOT NULL DEFAULT 'local_proxy';
ALTER TABLE app_settings ADD COLUMN tun_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE app_settings ADD COLUMN tun_stack TEXT NOT NULL DEFAULT 'system';
ALTER TABLE app_settings ADD COLUMN tun_auto_route INTEGER NOT NULL DEFAULT 1;
ALTER TABLE app_settings ADD COLUMN tun_strict_route INTEGER NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS connection_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  active_mode TEXT NOT NULL DEFAULT 'direct',
  selected_node_id TEXT,
  selected_node_name TEXT,
  selected_core TEXT NOT NULL DEFAULT 'sing-box',
  network_mode TEXT NOT NULL DEFAULT 'local_proxy',
  local_proxy_address TEXT NOT NULL DEFAULT '127.0.0.1:2080',
  status TEXT NOT NULL DEFAULT 'disconnected',
  last_error TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS system_proxy_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  enabled_by_rayflow INTEGER NOT NULL DEFAULT 0,
  previous_proxy_enable INTEGER NOT NULL DEFAULT 0,
  previous_proxy_server TEXT NOT NULL DEFAULT '',
  previous_proxy_override TEXT NOT NULL DEFAULT '',
  current_proxy_server TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);

INSERT OR IGNORE INTO connection_state(id, active_mode, selected_core, network_mode, local_proxy_address, status, updated_at)
VALUES (1, 'direct', 'sing-box', 'local_proxy', '127.0.0.1:2080', 'disconnected', CURRENT_TIMESTAMP);

INSERT OR IGNORE INTO system_proxy_state(id, enabled_by_rayflow, previous_proxy_enable, previous_proxy_server, previous_proxy_override, current_proxy_server, updated_at)
VALUES (1, 0, 0, '', '', '', CURRENT_TIMESTAMP);

