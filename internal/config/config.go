package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Config represents runtime configuration for tm-agent.
type Config struct {
	PlatformName       string `json:"platform_name"`
	TargetContainer    string `json:"target_container"`
	TargetID           string `json:"target_id"`
	Generation         int    `json:"generation"`
	SchemaVersion      int    `json:"schema_version"`
	MaxRecordBytes     int64  `json:"max_record_bytes"`
	SpoolDir           string `json:"spool_dir"`
	ContainerEngine    string `json:"container_engine"`
	SocketPath         string `json:"socket_path"`
	MaxSpoolAgeHours   int    `json:"max_spool_age_hours"`
	MaxSpoolFiles      int    `json:"max_spool_files"`
	StaleTmpAgeMinutes int    `json:"stale_tmp_age_minutes"`
	RunOnce            bool   `json:"run_once"`
}

// DefaultSpoolDir returns the standard spool directory based on the running OS.
func DefaultSpoolDir() string {
	if runtime.GOOS == "windows" {
		return `C:\monitoring\spool`
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = "/tmp"
	}
	return filepath.Join(home, ".local", "share", "tomcat-monitoring", "spool")
}

// NewDefaultConfig initializes configuration with default values.
func NewDefaultConfig() *Config {
	return &Config{
		PlatformName:       "tomcat-monitoring",
		TargetContainer:    "tomcat-jmx-exporter",
		TargetID:           "lab/tomcat-01/default",
		Generation:         1,
		SchemaVersion:      1,
		MaxRecordBytes:     16384,
		SpoolDir:           DefaultSpoolDir(),
		ContainerEngine:    "",
		SocketPath:         "",
		MaxSpoolAgeHours:   24,
		MaxSpoolFiles:      1000,
		StaleTmpAgeMinutes: 60,
		RunOnce:            false,
	}
}

// LoadConfig loads configuration with precedence: Defaults -> CONFIG file -> Environment Variables.
func LoadConfig(configPath string) (*Config, error) {
	cfg := NewDefaultConfig()

	// 1. If explicit configPath provided or default CONFIG file exists, parse it
	if configPath == "" {
		if _, err := os.Stat("CONFIG"); err == nil {
			configPath = "CONFIG"
		}
	}

	if configPath != "" {
		if err := parseConfigFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
		}
	}

	// 2. Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

func parseConfigFile(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		// Strip parameter expansion syntax like ${VAR:-default}
		if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
			inner := val[2 : len(val)-1]
			if subParts := strings.SplitN(inner, ":-", 2); len(subParts) == 2 {
				envKey := subParts[0]
				defaultVal := subParts[1]
				if envVal := os.Getenv(envKey); envVal != "" {
					val = envVal
				} else {
					val = defaultVal
				}
			}
		}

		switch key {
		case "PLATFORM_NAME":
			cfg.PlatformName = val
		case "TARGET_CONTAINER", "DEFAULT_TARGET_CONTAINER":
			cfg.TargetContainer = val
		case "TARGET_ID", "DEFAULT_TARGET_ID":
			cfg.TargetID = val
		case "GENERATION":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.Generation = n
			}
		case "SCHEMA_VERSION":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.SchemaVersion = n
			}
		case "MAX_RECORD_BYTES":
			if n, err := strconv.ParseInt(val, 10, 64); err == nil {
				cfg.MaxRecordBytes = n
			}
		case "DEFAULT_SPOOL_DIR", "SPOOL_DIR":
			if val != "" {
				cfg.SpoolDir = expandHome(val)
			}
		case "CONTAINER_ENGINE":
			cfg.ContainerEngine = val
		case "SOCKET_PATH":
			cfg.SocketPath = val
		case "DEFAULT_MAX_SPOOL_AGE_HOURS", "MAX_SPOOL_AGE_HOURS":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.MaxSpoolAgeHours = n
			}
		case "DEFAULT_MAX_SPOOL_FILES", "MAX_SPOOL_FILES":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.MaxSpoolFiles = n
			}
		case "DEFAULT_STALE_TMP_AGE_MINUTES", "STALE_TMP_AGE_MINUTES":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.StaleTmpAgeMinutes = n
			}
		case "RUN_ONCE":
			cfg.RunOnce = strings.ToLower(val) == "true" || val == "1"
		}
	}

	return scanner.Err()
}

func applyEnvOverrides(cfg *Config) {
	if val := os.Getenv("TARGET_CONTAINER"); val != "" {
		cfg.TargetContainer = val
	}
	if val := os.Getenv("TARGET_ID"); val != "" {
		cfg.TargetID = val
	}
	if val := os.Getenv("GENERATION"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Generation = n
		}
	}
	if val := os.Getenv("SPOOL_DIR"); val != "" {
		cfg.SpoolDir = expandHome(val)
	}
	if val := os.Getenv("CONTAINER_ENGINE"); val != "" {
		cfg.ContainerEngine = val
	}
	if val := os.Getenv("SOCKET_PATH"); val != "" {
		cfg.SocketPath = val
	}
	if val := os.Getenv("MAX_SPOOL_AGE_HOURS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.MaxSpoolAgeHours = n
		}
	}
	if val := os.Getenv("MAX_SPOOL_FILES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.MaxSpoolFiles = n
		}
	}
	if val := os.Getenv("STALE_TMP_AGE_MINUTES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.StaleTmpAgeMinutes = n
		}
	}
	if val := os.Getenv("RUN_ONCE"); val != "" {
		cfg.RunOnce = strings.ToLower(val) == "true" || val == "1"
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "${HOME}") {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/tmp"
		}
		return strings.Replace(path, "${HOME}", home, 1)
	}
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/tmp"
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
