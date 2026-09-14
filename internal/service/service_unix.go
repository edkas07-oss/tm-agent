//go:build !windows

package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eddywiyatno/tm-agent/internal/collector"
)

// RunService runs the collector daemon on Unix systems with signal handling.
func RunService(c *collector.Collector) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	return c.Run(ctx)
}
