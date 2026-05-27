package diagnostics

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type IPCheckRequest struct {
	Route         string `json:"route"`
	ProxyAddress  string `json:"proxyAddress"`
	ProxyProtocol string `json:"proxyProtocol"`
}

type IPCheckResult struct {
	Route     string `json:"route"`
	Status    string `json:"status"`
	IP        string `json:"ip,omitempty"`
	Country   string `json:"country,omitempty"`
	Provider  string `json:"provider"`
	LatencyMS int    `json:"latencyMs"`
	CheckedAt string `json:"checkedAt"`
	Error     string `json:"error,omitempty"`
}

type ipProvider struct {
	Name string
	URL  string
}

var ipProviders = []ipProvider{
	{Name: "ipapi.co", URL: "https://ipapi.co/json/"},
	{Name: "api.ipify.org", URL: "https://api.ipify.org?format=json"},
	{Name: "ifconfig.co", URL: "https://ifconfig.co/json"},
}

func (s *Service) IPCheck(ctx context.Context, req IPCheckRequest) IPCheckResult {
	if req.Route == "" {
		req.Route = "direct"
	}
	result := IPCheckResult{
		Route:     req.Route,
		Status:    "failed",
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}

	client, err := ipCheckHTTPClient(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	var lastErr error
	for _, provider := range ipProviders {
		start := time.Now()
		ip, country, err := fetchIP(ctx, client, provider)
		result.LatencyMS = int(time.Since(start).Milliseconds())
		result.Provider = provider.Name
		if err != nil {
			lastErr = err
			continue
		}
		result.Status = "ok"
		result.IP = ip
		result.Country = country
		result.Error = ""
		return result
	}
	if lastErr != nil {
		result.Error = lastErr.Error()
	}
	return result
}

func ipCheckHTTPClient(req IPCheckRequest) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   8 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
	}

	if req.Route == "rayflow_proxy" {
		if strings.TrimSpace(req.ProxyAddress) == "" {
			return nil, fmt.Errorf("proxy address is required for rayflow_proxy checks")
		}
		protocol := strings.ToLower(strings.TrimSpace(req.ProxyProtocol))
		if protocol == "" {
			protocol = "socks5"
		}
		switch protocol {
		case "http":
			proxyURL := &url.URL{Scheme: "http", Host: req.ProxyAddress}
			transport.Proxy = http.ProxyURL(proxyURL)
		case "socks5", "socks":
			dialer := socks5Dialer{proxyAddress: req.ProxyAddress}
			transport.Proxy = nil
			transport.DialContext = dialer.DialContext
		default:
			return nil, fmt.Errorf("unsupported proxy protocol %q", req.ProxyProtocol)
		}
	}

	return &http.Client{Transport: transport, Timeout: 12 * time.Second}, nil
}

func fetchIP(ctx context.Context, client *http.Client, provider ipProvider) (string, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.URL, nil)
	if err != nil {
		return "", "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "RayFlow/0.1")
	response, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", "", fmt.Errorf("%s returned HTTP %d", provider.Name, response.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&payload); err != nil {
		return "", "", err
	}
	ip := firstString(payload, "ip", "query", "ip_addr")
	if net.ParseIP(ip) == nil {
		return "", "", fmt.Errorf("%s returned no valid IP", provider.Name)
	}
	country := firstString(payload, "country_name", "country", "country_code")
	return ip, country, nil
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type socks5Dialer struct {
	proxyAddress string
}

func (d socks5Dialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	conn, err := (&net.Dialer{Timeout: 8 * time.Second}).DialContext(ctx, "tcp", d.proxyAddress)
	if err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}
	if err := socks5Handshake(conn, address); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func socks5Handshake(conn net.Conn, destination string) error {
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return err
	}
	greeting := make([]byte, 2)
	if _, err := io.ReadFull(conn, greeting); err != nil {
		return err
	}
	if greeting[0] != 0x05 || greeting[1] != 0x00 {
		return fmt.Errorf("SOCKS5 proxy rejected no-auth handshake")
	}

	host, portText, err := net.SplitHostPort(destination)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid destination port %q", portText)
	}

	request := []byte{0x05, 0x01, 0x00}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			request = append(request, 0x01)
			request = append(request, ip4...)
		} else {
			request = append(request, 0x04)
			request = append(request, ip.To16()...)
		}
	} else {
		if len(host) > 255 {
			return fmt.Errorf("destination host is too long")
		}
		request = append(request, 0x03, byte(len(host)))
		request = append(request, []byte(host)...)
	}
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	request = append(request, portBytes...)
	if _, err := conn.Write(request); err != nil {
		return err
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 0x05 {
		return fmt.Errorf("invalid SOCKS5 response")
	}
	if header[1] != 0x00 {
		return fmt.Errorf("SOCKS5 connect failed with code 0x%02x", header[1])
	}

	switch header[3] {
	case 0x01:
		_, err = io.CopyN(io.Discard, conn, 4)
	case 0x03:
		length := []byte{0}
		if _, err = io.ReadFull(conn, length); err == nil {
			_, err = io.CopyN(io.Discard, conn, int64(length[0]))
		}
	case 0x04:
		_, err = io.CopyN(io.Discard, conn, 16)
	default:
		err = fmt.Errorf("unknown SOCKS5 address type 0x%02x", header[3])
	}
	if err != nil {
		return err
	}
	_, err = io.CopyN(io.Discard, conn, 2)
	return err
}
