package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eddywiyatno/tm-agent/internal/buildinfo"
	"github.com/eddywiyatno/tm-agent/internal/collector"
	"github.com/eddywiyatno/tm-agent/internal/config"
	"github.com/eddywiyatno/tm-agent/internal/engine"
	"github.com/eddywiyatno/tm-agent/internal/service"
	"github.com/eddywiyatno/tm-agent/internal/validator"
	"github.com/eddywiyatno/tm-agent/pkg/termutil"
)

func printHelp() {
	fmt.Printf(`tm-agent — Unified Cross-Platform Event Collector Daemon for Tomcat Monitoring

Usage:
  tm-agent [flags]

Flags:
  --config <path>       Path to configuration file (default: "CONFIG" if present)
  --run-once            Execute single snapshot and retention prune cycle, then exit
  --target <name>       Target container name (default: "tomcat-jmx-exporter")
  --target-id <id>      Canonical target identifier (default: "lab/tomcat-01/default")
  --spool-dir <path>    Evidence spool directory
  --engine <name>       Container engine override (podman / docker)
  --socket <path>       Container engine socket path or named pipe
  --validate            Execute baseline repository validation and exit
  --version, -v         Display version and build information
  --help, -h            Display this help message

Examples:
  # Start daemon listening to real-time container socket events
  tm-agent

  # Run one-shot snapshot for testing or CI verification
  tm-agent --run-once

  # Explicit container engine socket and spool directory
  tm-agent --target tomcat-jmx-exporter --engine podman --spool-dir /var/lib/monitoring/spool
`)
}

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	runOnce := flag.Bool("run-once", false, "Execute single snapshot and exit")
	targetContainer := flag.String("target", "", "Target container name")
	targetID := flag.String("target-id", "", "Target identifier")
	spoolDir := flag.String("spool-dir", "", "Spool directory path")
	engineType := flag.String("engine", "", "Container engine (podman / docker)")
	socketPath := flag.String("socket", "", "Container engine socket path")
	runValidation := flag.Bool("validate", false, "Run repository validation")
	showVersion := flag.Bool("version", false, "Display version info")
	showVersionShort := flag.Bool("v", false, "Display version info (shorthand)")
	showHelp := flag.Bool("help", false, "Display help")
	showHelpShort := flag.Bool("h", false, "Display help (shorthand)")

	flag.Usage = printHelp
	flag.Parse()

	if *showHelp || *showHelpShort {
		printHelp()
		os.Exit(0)
	}

	if *showVersion || *showVersionShort {
		fmt.Println(buildinfo.String())
		os.Exit(0)
	}

	if *runValidation {
		dir := "."
		if flag.NArg() > 0 {
			dir = flag.Arg(0)
		}
		if err := validator.ValidateRepository(dir); err != nil {
			termutil.PrintError("Validation failed: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 1. Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		termutil.PrintError("Configuration error: %v", err)
		os.Exit(1)
	}

	// Apply CLI flag overrides
	if *runOnce {
		cfg.RunOnce = true
	}
	if *targetContainer != "" {
		cfg.TargetContainer = *targetContainer
	}
	if *targetID != "" {
		cfg.TargetID = *targetID
	}
	if *spoolDir != "" {
		cfg.SpoolDir = *spoolDir
	}
	if *engineType != "" {
		cfg.ContainerEngine = *engineType
	}
	if *socketPath != "" {
		cfg.SocketPath = *socketPath
	}

	// 2. Initialize Container Engine client
	engineClient, err := engine.NewEngineClient(cfg.ContainerEngine, cfg.SocketPath)
	if err != nil {
		termutil.PrintWarning("Container engine socket initialization warning: %v", err)
	}

	// Test engine connectivity
	if engineClient != nil {
		info, pErr := engineClient.Ping(context.Background())
		if pErr != nil {
			termutil.PrintWarning("Container engine ping warning: %v (will retry during event streaming)", pErr)
		} else {
			termutil.PrintSuccess("Connected to container engine: %s (API: %s, OS: %s)", info.EngineType, info.APIVersion, info.OSType)
		}
	}

	// 3. Initialize Collector
	c := collector.NewCollector(cfg, engineClient)

	// 4. Run via Service / Daemon runner
	if err := service.RunService(c); err != nil {
		termutil.PrintError("Agent execution failed: %v", err)
		os.Exit(1)
	}
}
