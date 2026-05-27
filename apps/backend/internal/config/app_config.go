package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AppConfig struct {
	HTTPAddr       string
	DatabasePath   string
	MigrationsDir  string
	ConfigsDir     string
	LogsDir        string
	PresetsDir     string
	AuthToken      string
	DefaultCore    string
	SingBoxPath    string
	XrayPath       string
	ZapretPath     string
	LocalProxyPort int
}

func (c AppConfig) Validate() error {
	if len(c.AuthToken) < 16 {
		return fmt.Errorf("security validation failed: RAYFLOW_AUTH_TOKEN must be at least 16 characters long (got %d)", len(c.AuthToken))
	}
	return nil
}

func LoadAppConfig() AppConfig {
	dbPath := env("RAYFLOW_DB_PATH", "data/rayflow.db")
	authToken := os.Getenv("RAYFLOW_AUTH_TOKEN")
	if authToken == "" {
		// Fallback: Try to read from local file data/.auth_token in development/production
		tokenFile := filepath.Join(filepath.Dir(dbPath), ".auth_token")
		if data, err := os.ReadFile(tokenFile); err == nil {
			authToken = strings.TrimSpace(string(data))
		}
	}

	return AppConfig{
		HTTPAddr:       env("RAYFLOW_HTTP_ADDR", "127.0.0.1:8787"),
		DatabasePath:   dbPath,
		MigrationsDir:  env("RAYFLOW_MIGRATIONS_DIR", "migrations"),
		ConfigsDir:     env("RAYFLOW_CONFIGS_DIR", "configs"),
		LogsDir:        env("RAYFLOW_LOGS_DIR", "logs"),
		PresetsDir:     env("RAYFLOW_PRESETS_DIR", "presets"),
		AuthToken:      authToken,
		DefaultCore:    env("RAYFLOW_DEFAULT_CORE", "sing-box"),
		SingBoxPath:    os.Getenv("RAYFLOW_SING_BOX_PATH"),
		XrayPath:       os.Getenv("RAYFLOW_XRAY_PATH"),
		ZapretPath:     os.Getenv("RAYFLOW_ZAPRET_PATH"),
		LocalProxyPort: envInt("RAYFLOW_LOCAL_PROXY_PORT", 2080),
	}
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
