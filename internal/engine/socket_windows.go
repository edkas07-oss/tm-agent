//go:build windows

package engine

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

func createTransport(socketPath string) (*http.Transport, error) {
	if strings.HasPrefix(socketPath, "tcp://") {
		addr := strings.TrimPrefix(socketPath, "tcp://")
		return &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				var d net.Dialer
				d.Timeout = 1 * time.Second
				return d.DialContext(ctx, "tcp", addr)
			},
		}, nil
	}

	return &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			d.Timeout = 1 * time.Second
			return d.DialContext(ctx, "tcp", "127.0.0.1:2375")
		},
	}, nil
}
