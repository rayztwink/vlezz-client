package singbox

import (
	"encoding/json"
	"testing"

	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
)

func TestGenerateConfigProducesJSON(t *testing.T) {
	cfg, err := GenerateConfig(storage.Node{
		Name:      "example",
		Protocol:  "vless",
		Address:   "example.com",
		Port:      443,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Security:  "reality",
		Transport: "tcp",
		RawLink:   "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=tcp&security=reality&sni=example.com&fp=chrome&pbk=public-key&sid=abcd#example",
	}, 2080)
	if err != nil {
		t.Fatalf("GenerateConfig returned error: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(cfg, &payload); err != nil {
		t.Fatalf("generated config is not JSON: %v", err)
	}
	if payload["inbounds"] == nil || payload["outbounds"] == nil {
		t.Fatalf("generated config missing core sections")
	}
}

func TestGenerateConfigWithTUNInbound(t *testing.T) {
	cfg, err := GenerateConfigWithOptions(storage.Node{
		Name:      "example",
		Protocol:  "vless",
		Address:   "example.com",
		Port:      443,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Security:  "reality",
		Transport: "tcp",
		RawLink:   "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=tcp&security=reality&sni=example.com&fp=chrome&pbk=public-key&sid=abcd#example",
	}, core.ConnectOptions{
		NetworkMode:    core.NetworkModeTUN,
		TUNStack:       "system",
		TUNAutoRoute:   true,
		TUNStrictRoute: true,
	})
	if err != nil {
		t.Fatalf("GenerateConfigWithOptions returned error: %v", err)
	}
	var payload struct {
		Inbounds []map[string]any `json:"inbounds"`
	}
	if err := json.Unmarshal(cfg, &payload); err != nil {
		t.Fatalf("generated config is not JSON: %v", err)
	}
	if len(payload.Inbounds) != 1 || payload.Inbounds[0]["type"] != "tun" {
		t.Fatalf("expected tun inbound, got %#v", payload.Inbounds)
	}
}

func TestGenerateConfigWithXHTTP(t *testing.T) {
	cfg, err := GenerateConfig(storage.Node{
		Name:      "example-xhttp",
		Protocol:  "vless",
		Address:   "example.com",
		Port:      443,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Security:  "reality",
		Transport: "xhttp",
		RawLink:   "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?type=xhttp&security=reality&sni=example.com&fp=chrome&pbk=public-key&sid=abcd&path=%2Fcustompath#example-xhttp",
	}, 2080)
	if err != nil {
		t.Fatalf("GenerateConfig returned error: %v", err)
	}

	var payload struct {
		Outbounds []struct {
			Type      string `json:"type"`
			Transport struct {
				Type string `json:"type"`
				Path string `json:"path"`
			} `json:"transport"`
		} `json:"outbounds"`
	}

	if err := json.Unmarshal(cfg, &payload); err != nil {
		t.Fatalf("generated config is not JSON: %v", err)
	}

	foundProxy := false
	for _, outbound := range payload.Outbounds {
		if outbound.Type == "vless" {
			foundProxy = true
			if outbound.Transport.Type != "xhttp" {
				t.Errorf("expected transport type xhttp, got %q", outbound.Transport.Type)
			}
			if outbound.Transport.Path != "/custompath" {
				t.Errorf("expected transport path /custompath, got %q", outbound.Transport.Path)
			}
		}
	}

	if !foundProxy {
		t.Fatalf("expected to find vless outbound in generated config")
	}
}
