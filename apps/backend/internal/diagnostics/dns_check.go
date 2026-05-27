package diagnostics

import (
	"context"
	"net"
	"time"
)

func DNSResolve(ctx context.Context, hostname string) (time.Duration, []string, error) {
	start := time.Now()
	addrs, err := net.DefaultResolver.LookupHost(ctx, hostname)
	return time.Since(start), addrs, err
}
