package diagnostics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
)

func TestIPCheckDirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"203.0.113.10","country_name":"Testland"}`))
	}))
	defer server.Close()
	restoreProviders := setIPProvidersForTest(t, []ipProvider{{Name: "test", URL: server.URL}})
	defer restoreProviders()

	service := NewService(nil, logs.NewManager(nil, zerolog.Nop()))
	result := service.IPCheck(context.Background(), IPCheckRequest{Route: "direct"})
	if result.Status != "ok" || result.IP != "203.0.113.10" || result.Country != "Testland" {
		t.Fatalf("unexpected IP check result: %#v", result)
	}
}

func TestIPCheckHTTPProxyRoute(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "http://provider.test/ip" {
			t.Fatalf("expected proxied absolute URL, got %q", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"198.51.100.20"}`))
	}))
	defer proxy.Close()
	restoreProviders := setIPProvidersForTest(t, []ipProvider{{Name: "proxy-test", URL: "http://provider.test/ip"}})
	defer restoreProviders()

	service := NewService(nil, logs.NewManager(nil, zerolog.Nop()))
	result := service.IPCheck(context.Background(), IPCheckRequest{
		Route:         "rayflow_proxy",
		ProxyAddress:  proxy.Listener.Addr().String(),
		ProxyProtocol: "http",
	})
	if result.Status != "ok" || result.IP != "198.51.100.20" {
		t.Fatalf("unexpected proxy IP check result: %#v", result)
	}
}

func TestIPCheckProviderErrorReturnsFailedResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()
	restoreProviders := setIPProvidersForTest(t, []ipProvider{{Name: "bad", URL: server.URL}})
	defer restoreProviders()

	service := NewService(nil, logs.NewManager(nil, zerolog.Nop()))
	result := service.IPCheck(context.Background(), IPCheckRequest{Route: "direct"})
	if result.Status != "failed" || result.Error == "" {
		t.Fatalf("expected failed result with error, got %#v", result)
	}
}

func setIPProvidersForTest(t *testing.T, providers []ipProvider) func() {
	t.Helper()
	original := ipProviders
	ipProviders = providers
	return func() {
		ipProviders = original
	}
}
