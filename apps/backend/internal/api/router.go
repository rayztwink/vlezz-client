package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

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

type Dependencies struct {
	Config        config.AppConfig
	Logger        zerolog.Logger
	Nodes         *storage.NodeRepository
	Subscriptions *storage.SubscriptionRepository
	Presets       *storage.PresetRepository
	RoutingRules  *storage.RoutingRuleRepository
	Checks        *storage.CheckRepository
	Settings      *storage.SettingsRepository
	Logs          *logs.Manager
	Diagnostics   *diagnostics.Service
	Processes     *process.Manager
	SingBox       *singbox.Client
	Xray          *xray.Client
	Zapret        *zapret.Client
	Connection    *connection.Manager
	SystemProxy   *systemproxy.Service
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(CORS)
	r.Use(ValidateHost([]string{"localhost", "127.0.0.1", "[::1]"}))
	r.Use(LocalOnly)
	r.Use(AuthToken(deps.Config.AuthToken))

	health := HealthHandler{}
	nodes := NodesHandler{deps: deps}
	subscriptions := SubscriptionsHandler{deps: deps}
	zapretHandler := ZapretHandler{deps: deps}
	routing := RoutingHandler{deps: deps}
	diagnosticsHandler := DiagnosticsHandler{deps: deps}
	settings := SettingsHandler{deps: deps}
	cores := CoresHandler{deps: deps}
	logsHandler := LogsHandler{deps: deps}
	connectionHandler := ConnectionHandler{deps: deps}
	systemProxyHandler := SystemProxyHandler{deps: deps}
	runtimeHandler := RuntimeHandler{deps: deps}

	r.Get("/", health.Get)
	r.Get("/health", health.Get)

	r.Route("/nodes", func(r chi.Router) {
		r.Get("/", nodes.List)
		r.Post("/import", nodes.Import)
		r.Delete("/{id}", nodes.Delete)
		r.Post("/{id}/check", nodes.Check)
		r.Post("/{id}/connect", nodes.Connect)
		r.Post("/disconnect", nodes.Disconnect)
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", subscriptions.List)
		r.Post("/", subscriptions.Create)
		r.Post("/{id}/update", subscriptions.Update)
		r.Delete("/{id}", subscriptions.Delete)
	})

	r.Route("/zapret", func(r chi.Router) {
		r.Get("/presets", zapretHandler.ListPresets)
		r.Post("/presets/update", zapretHandler.UpdatePresets)
		r.Post("/presets/{id}/start", zapretHandler.StartPreset)
		r.Post("/stop", zapretHandler.Stop)
		r.Get("/logs", zapretHandler.Logs)
	})

	r.Route("/routing", func(r chi.Router) {
		r.Get("/rules", routing.List)
		r.Post("/rules", routing.Create)
		r.Delete("/rules/{id}", routing.Delete)
	})

	r.Route("/diagnostics", func(r chi.Router) {
		r.Post("/check", diagnosticsHandler.Check)
		r.Post("/ip-check", diagnosticsHandler.IPCheck)
		r.Get("/history", diagnosticsHandler.History)
	})

	r.Get("/settings", settings.Get)
	r.Patch("/settings", settings.Patch)
	r.Get("/cores/status", cores.Status)
	r.Post("/cores/validate", cores.Validate)
	r.Get("/logs", logsHandler.List)
	r.Get("/connection/status", connectionHandler.Status)
	r.Post("/connection/disconnect", connectionHandler.Disconnect)
	r.Get("/connection/report", connectionHandler.Report)
	r.Get("/runtime/capabilities", runtimeHandler.Capabilities)
	r.Get("/system-proxy/status", systemProxyHandler.Status)
	r.Post("/system-proxy/enable", systemProxyHandler.Enable)
	r.Post("/system-proxy/disable", systemProxyHandler.Disable)

	return r
}
