CREATE TABLE IF NOT EXISTS app_settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  theme TEXT NOT NULL DEFAULT 'system',
  language TEXT NOT NULL DEFAULT 'system',
  autostart INTEGER NOT NULL DEFAULT 0,
  active_mode TEXT NOT NULL DEFAULT 'direct',
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS nodes (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  protocol TEXT NOT NULL,
  address TEXT NOT NULL,
  port INTEGER NOT NULL,
  uuid TEXT NOT NULL,
  security TEXT NOT NULL,
  transport TEXT NOT NULL,
  raw_link TEXT,
  latency_ms INTEGER,
  country TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  update_interval INTEGER NOT NULL,
  last_update_at TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS zapret_presets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source TEXT NOT NULL,
  command TEXT NOT NULL,
  description TEXT,
  is_active INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS routing_rules (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL,
  mode TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS checks (
  id TEXT PRIMARY KEY,
  target TEXT NOT NULL,
  mode TEXT NOT NULL,
  status TEXT NOT NULL,
  latency_ms INTEGER,
  error TEXT,
  checked_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS logs (
  id TEXT PRIMARY KEY,
  source TEXT NOT NULL,
  level TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_protocol ON nodes(protocol);
CREATE INDEX IF NOT EXISTS idx_checks_checked_at ON checks(checked_at);
CREATE INDEX IF NOT EXISTS idx_logs_source_created_at ON logs(source, created_at);
CREATE INDEX IF NOT EXISTS idx_routing_rules_domain ON routing_rules(domain);

INSERT OR IGNORE INTO app_settings(id, theme, language, autostart, active_mode, updated_at)
VALUES (1, 'system', 'system', 0, 'direct', CURRENT_TIMESTAMP);

