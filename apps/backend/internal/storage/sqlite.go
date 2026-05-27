package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func Open(path string) (*Database, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", abs)
	if err != nil {
		return nil, err
	}
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return &Database{db: db}, nil
}

func (d *Database) SQL() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}

type Repositories struct {
	Nodes         *NodeRepository
	Subscriptions *SubscriptionRepository
	Presets       *PresetRepository
	RoutingRules  *RoutingRuleRepository
	Checks        *CheckRepository
	Settings      *SettingsRepository
	Logs          *LogRepository
	Connection    *ConnectionRepository
	SystemProxy   *SystemProxyRepository
}

func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		Nodes:         NewNodeRepository(db),
		Subscriptions: NewSubscriptionRepository(db),
		Presets:       NewPresetRepository(db),
		RoutingRules:  NewRoutingRuleRepository(db),
		Checks:        NewCheckRepository(db),
		Settings:      NewSettingsRepository(db),
		Logs:          NewLogRepository(db),
		Connection:    NewConnectionRepository(db),
		SystemProxy:   NewSystemProxyRepository(db),
	}
}
