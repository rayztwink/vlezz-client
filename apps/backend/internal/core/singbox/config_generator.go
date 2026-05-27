package singbox

import (
	"encoding/json"
	"fmt"

	"github.com/rayflow/rayflow-client/apps/backend/internal/core"
	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/subscription"
)

func GenerateConfig(node storage.Node, localPort int) ([]byte, error) {
	return GenerateConfigWithOptions(node, core.ConnectOptions{NetworkMode: core.NetworkModeLocalProxy, LocalProxyPort: localPort})
}

func GenerateConfigWithOptions(node storage.Node, options core.ConnectOptions) ([]byte, error) {
	options = options.Normalized()
	profile, err := profileFromNode(node)
	if err != nil {
		return nil, err
	}

	outbound := map[string]any{
		"type":        "vless",
		"tag":         "proxy",
		"server":      profile.Address,
		"server_port": profile.Port,
		"uuid":        profile.UUID,
	}
	if flow := profile.Params["flow"]; flow != "" {
		outbound["flow"] = flow
	}
	if tls := buildSingBoxTLS(profile); tls != nil {
		outbound["tls"] = tls
	}
	if transport := buildSingBoxTransport(profile); transport != nil {
		outbound["transport"] = transport
	}

	cfg := map[string]any{
		"log": map[string]any{
			"level":     "info",
			"timestamp": true,
		},
		"inbounds": buildSingBoxInbounds(options),
		"outbounds": []map[string]any{
			outbound,
			{"type": "direct", "tag": "direct"},
			{"type": "block", "tag": "block"},
		},
		"route": map[string]any{
			"auto_detect_interface": true,
			"final":                 "proxy",
		},
	}

	result, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	if !json.Valid(result) {
		return nil, fmt.Errorf("generated sing-box config is not valid JSON")
	}
	return result, nil
}

func buildSingBoxInbounds(options core.ConnectOptions) []map[string]any {
	if options.NetworkMode == core.NetworkModeTUN {
		return []map[string]any{
			{
				"type":           "tun",
				"tag":            "tun-in",
				"interface_name": "rayflow0",
				"address":        []string{"172.19.0.1/30"},
				"mtu":            9000,
				"stack":          options.TUNStack,
				"auto_route":     options.TUNAutoRoute,
				"strict_route":   options.TUNStrictRoute,
				"sniff":          true,
			},
		}
	}
	return []map[string]any{
		{
			"type":        "mixed",
			"tag":         "mixed-in",
			"listen":      "127.0.0.1",
			"listen_port": options.LocalProxyPort,
			"sniff":       true,
		},
	}
}

func profileFromNode(node storage.Node) (subscription.ParsedNode, error) {
	if node.RawLink != "" {
		return subscription.ParseVLESS(node.RawLink)
	}
	return subscription.ParsedNode{
		Name:      node.Name,
		Protocol:  node.Protocol,
		Address:   node.Address,
		Port:      node.Port,
		UUID:      node.UUID,
		Security:  node.Security,
		Transport: node.Transport,
		Params:    map[string]string{},
	}, nil
}

func buildSingBoxTLS(profile subscription.ParsedNode) map[string]any {
	if profile.Security != "tls" && profile.Security != "reality" {
		return nil
	}
	tls := map[string]any{
		"enabled":     true,
		"server_name": first(profile.Params["sni"], profile.Params["host"], profile.Address),
	}
	if fp := profile.Params["fp"]; fp != "" {
		tls["utls"] = map[string]any{"enabled": true, "fingerprint": fp}
	}
	if profile.Security == "reality" {
		reality := map[string]any{"enabled": true}
		if pbk := profile.Params["pbk"]; pbk != "" {
			reality["public_key"] = pbk
		}
		if sid := profile.Params["sid"]; sid != "" {
			reality["short_id"] = sid
		}
		tls["reality"] = reality
	}
	return tls
}

func buildSingBoxTransport(profile subscription.ParsedNode) map[string]any {
	switch profile.Transport {
	case "ws":
		transport := map[string]any{"type": "ws"}
		if path := profile.Params["path"]; path != "" {
			transport["path"] = path
		}
		if host := profile.Params["host"]; host != "" {
			transport["headers"] = map[string]string{"Host": host}
		}
		return transport
	case "grpc":
		transport := map[string]any{"type": "grpc"}
		if serviceName := profile.Params["serviceName"]; serviceName != "" {
			transport["service_name"] = serviceName
		}
		return transport
	case "xhttp":
		transport := map[string]any{"type": "xhttp"}
		if path := profile.Params["path"]; path != "" {
			transport["path"] = path
		}
		return transport
	default:
		return nil
	}
}

func first(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
