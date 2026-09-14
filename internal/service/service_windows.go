//go:build windows

package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eddywiyatno/tm-agent/internal/collector"
	"github.com/eddywiyatno/tm-agent/pkg/termutil"
	"golang.org/x/sys/windows/svc"
)

type windowsService struct {
	collector *collector.Collector
}

func (m *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.collector.Run(ctx)
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case err := <-errCh:
			if err != nil {
				termutil.PrintError("Service execution failed: %v", err)
			}
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				break loop
			default:
				termutil.PrintWarning("Unexpected service control request #%d", c.Cmd)
			}
		}
	}

	changes <- svc.Status{State: svc.Stopped}
	return
}

// RunService runs the collector daemon on Windows either as an interactive console app or as a Windows Service.
func RunService(c *collector.Collector) error {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		termutil.PrintWarning("Failed to determine interactive session: %v", err)
		isInteractive = true
	}

	if isInteractive {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		defer stop()
		return c.Run(ctx)
	}

	// Running as Windows Service SCM
	return svc.Run("TomcatMonitoringAgent", &windowsService{collector: c})
}
