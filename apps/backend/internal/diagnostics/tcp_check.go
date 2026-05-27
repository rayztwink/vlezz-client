package diagnostics

import (
	"context"
	"net"
	"time"
)

func TCPConnect(ctx context.Context, address string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(start), nil
}
