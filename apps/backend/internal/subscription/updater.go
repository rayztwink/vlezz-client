package subscription

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Updater struct {
	client *http.Client
}

func NewUpdater() *Updater {
	return &Updater{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (u *Updater) FetchLinks(ctx context.Context, subscriptionURL string) ([]string, error) {
	parsedURL, err := url.Parse(subscriptionURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("subscription URL must use http or https")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, subscriptionURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RayFlow Client/0.1")
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, err
	}
	payload := strings.TrimSpace(string(body))
	if decoded, err := decodeMaybeBase64(payload); err == nil && strings.Contains(string(decoded), "://") {
		payload = string(decoded)
	}

	var links []string
	for _, line := range strings.Fields(payload) {
		if strings.HasPrefix(line, "vless://") {
			links = append(links, line)
		}
	}
	return links, nil
}

func decodeMaybeBase64(payload string) ([]byte, error) {
	normalized := strings.NewReplacer("\r", "", "\n", "", " ", "", "\t", "").Replace(payload)
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, encoding := range encodings {
		decoded, err := encoding.DecodeString(normalized)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
