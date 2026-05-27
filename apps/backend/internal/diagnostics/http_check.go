package diagnostics

import (
	"context"
	"net/http"
	"time"
)

func HTTPAvailability(ctx context.Context, target string) (time.Duration, int, error) {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, 0, err
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	return time.Since(start), resp.StatusCode, nil
}
