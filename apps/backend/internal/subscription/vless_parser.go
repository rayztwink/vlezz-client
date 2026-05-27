package subscription

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func ParseVLESS(raw string) (ParsedNode, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ParsedNode{}, err
	}
	if parsed.Scheme != "vless" {
		return ParsedNode{}, fmt.Errorf("expected vless link")
	}

	uuidValue := parsed.User.Username()
	host := parsed.Hostname()
	portValue := parsed.Port()
	if portValue == "" {
		return ParsedNode{}, fmt.Errorf("port is required")
	}
	port, err := strconv.Atoi(portValue)
	if err != nil {
		return ParsedNode{}, fmt.Errorf("invalid port")
	}

	query := parsed.Query()
	params := make(map[string]string, len(query))
	for key, values := range query {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	transport := firstNonEmpty(params["type"], params["network"], "tcp")
	security := firstNonEmpty(params["security"], "none")
	name := parsed.Fragment
	if name == "" {
		name = net.JoinHostPort(host, strconv.Itoa(port))
	}

	node := ParsedNode{
		Name:      name,
		Protocol:  "vless",
		Address:   host,
		Port:      port,
		UUID:      uuidValue,
		Security:  strings.ToLower(security),
		Transport: strings.ToLower(transport),
		RawLink:   raw,
		Params:    params,
	}

	if err := ValidateParsedNode(node); err != nil {
		return ParsedNode{}, err
	}
	return node, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
