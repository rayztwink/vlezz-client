package xray

import (
	"encoding/json"
	"fmt"

	"github.com/rayflow/rayflow-client/apps/backend/internal/storage"
	"github.com/rayflow/rayflow-client/apps/backend/internal/subscription"
)

func GenerateConfig(node storage.Node, localPort int) ([]byte, error) {
	profile, err := profileFromNode(node)
	if err != nil {
		return nil, err
	}

	streamSettings := map[string]any{
		"network":  profile.Transport,
		"security": profile.Security,
	}
	if profile.Security == "reality" {
		streamSettings["realitySettings"] = map[string]any{
			"serverName":  first(profile.Params["sni"], profile.Address),
			"fingerprint": profile.Params["fp"],
			"publicKey":   profile.Params["pbk"],
			"shortId":     profile.Params["sid"],
			"spiderX":     first(profile.Params["spx"], "/"),
		}
	}
	if profile.Security == "tls" {
		streamSettings["tlsSettings"] = map[string]any{
			"serverName":  first(profile.Params["sni"], profile.Address),
			"fingerprint": profile.Params["fp"],
		}
	}
	applyTransport(profile, streamSettings)

	user := map[string]any{"id": profile.UUID, "encryption": "none"}
	if flow := profile.Params["flow"]; flow != "" {
		user["flow"] = flow
	}
	cfg := map[string]any{
		"log": map[string]any{"loglevel": "warning"},
		"inbounds": []map[string]any{
			{
				"listen":   "127.0.0.1",
				"port":     localPort,
				"protocol": "socks",
				"settings": map[string]any{"udp": true, "auth": "noauth"},
				"sniffing": map[string]any{"enabled": true, "destOverride": []string{"http", "tls", "quic"}},
			},
		},
		"outbounds": []map[string]any{
			{
				"protocol": "vless",
				"tag":      "proxy",
				"settings": map[string]any{
					"vnext": []map[string]any{
						{
							"address": profile.Address,
							"port":    profile.Port,
							"users":   []map[string]any{user},
						},
					},
				},
				"streamSettings": streamSettings,
			},
			{"protocol": "freedom", "tag": "direct"},
			{"protocol": "blackhole", "tag": "block"},
		},
	}

	result, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	if !json.Valid(result) {
		return nil, fmt.Errorf("generated xray config is not valid JSON")
	}
	return result, nil
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

func applyTransport(profile subscription.ParsedNode, stream map[string]any) {
	switch profile.Transport {
	case "ws":
		stream["wsSettings"] = map[string]any{"path": first(profile.Params["path"], "/"), "headers": map[string]string{"Host": profile.Params["host"]}}
	case "grpc":
		stream["grpcSettings"] = map[string]any{"serviceName": profile.Params["serviceName"]}
	case "xhttp":
		stream["xhttpSettings"] = map[string]any{"path": first(profile.Params["path"], "/")}
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
