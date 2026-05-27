package subscription

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var supportedTransports = map[string]bool{
	"tcp":   true,
	"ws":    true,
	"grpc":  true,
	"xhttp": true,
}

func ValidateParsedNode(node ParsedNode) error {
	if node.Protocol != "vless" {
		return fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
	if _, err := uuid.Parse(node.UUID); err != nil {
		return fmt.Errorf("invalid uuid")
	}
	if strings.TrimSpace(node.Address) == "" {
		return fmt.Errorf("address is required")
	}
	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("port is out of range")
	}
	if !supportedTransports[node.Transport] {
		return fmt.Errorf("unsupported transport: %s", node.Transport)
	}
	switch node.Security {
	case "none", "tls", "reality":
		return nil
	default:
		return fmt.Errorf("unsupported security: %s", node.Security)
	}
}
