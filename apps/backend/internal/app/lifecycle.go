package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/api"
	"github.com/rayflow/rayflow-client/apps/backend/internal/config"
	"github.com/rayflow/rayflow-client/apps/backend/internal/connection"
	"github.com/rayflow/rayflow-client/apps/backend/internal/core/singbox"
	"github.com/rayflow/rayflow-client/apps/backend/internal/core/xray"
	"github.com/rayflow/rayflow-client/apps/backend/internal/core/zapret"
	"github.com/rayflow/rayflow-client/apps/backend/internal/diagnostics"
	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
	"github.com/rayflow/rayflow-client/apps/backend/internal/process"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/systemproxy"
)

type App struct {
	cfg       config.AppConfig
	logger    zerolog.Logger
	db        *storage.Database
	processes *process.Manager
	server    *http.Server
}

func Stdout() io.Writer {
	return os.Stdout
}

func New(cfg config.AppConfig, logger zerolog.Logger) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	for _, dir := range []string{cfg.ConfigsDir, cfg.LogsDir, cfg.PresetsDir, filepath.Dir(cfg.DatabasePath)} {
		if dir == "." || dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := storage.Open(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	if err := storage.RunMigrations(db.SQL(), cfg.MigrationsDir); err != nil {
		_ = db.Close()
		return nil, err
	}

	repos := storage.NewRepositories(db.SQL())
	if err := repos.Settings.EnsureDefault(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := repos.Connection.EnsureDefault(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := repos.SystemProxy.EnsureDefault(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	logManager := logs.NewManager(repos.Logs, logger)
	processManager := process.NewManager(logManager, logger)
	diagnosticsService := diagnostics.NewService(repos.Checks, logManager)
	systemProxyService := systemproxy.NewService(repos.SystemProxy, logManager)
	if err := systemProxyService.RestoreOwnedIfCurrent(context.Background()); err != nil {
		logManager.Warn("system-proxy", "startup restore check failed: "+err.Error())
	}
	singBoxClient := singbox.NewClient(cfg, processManager, logManager, repos.Settings)
	xrayClient := xray.NewClient(cfg, processManager, logManager, repos.Settings)
	zapretClient := zapret.NewClient(cfg, processManager, logManager, repos.Presets, repos.Settings)
	connectionManager := connection.NewManager(repos.Nodes, repos.Settings, repos.Connection, logManager, systemProxyService, diagnosticsService, singBoxClient, xrayClient)

	deps := api.Dependencies{
		Config:        cfg,
		Logger:        logger,
		Nodes:         repos.Nodes,
		Subscriptions: repos.Subscriptions,
		Presets:       repos.Presets,
		RoutingRules:  repos.RoutingRules,
		Checks:        repos.Checks,
		Settings:      repos.Settings,
		Logs:          logManager,
		Diagnostics:   diagnosticsService,
		Processes:     processManager,
		SingBox:       singBoxClient,
		Xray:          xrayClient,
		Zapret:        zapretClient,
		Connection:    connectionManager,
		SystemProxy:   systemProxyService,
	}

	router := api.NewRouter(deps)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:       cfg,
		logger:    logger,
		db:        db,
		processes: processManager,
		server:    server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info().Str("addr", a.cfg.HTTPAddr).Msg("rayflowd listening")
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		a.logger.Info().Msg("shutting down rayflowd")
		_ = a.processes.StopAll(shutdownCtx)
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			_ = a.db.Close()
			return err
		}
		_ = a.db.Close()
		return nil
	case err := <-errCh:
		_ = a.processes.StopAll(context.Background())
		_ = a.db.Close()
		return err
	}
}
