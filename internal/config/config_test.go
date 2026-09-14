package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.TargetContainer != "tomcat-jmx-exporter" {
		t.Errorf("expected TargetContainer 'tomcat-jmx-exporter', got '%s'", cfg.TargetContainer)
	}
	if cfg.SchemaVersion != 1 {
		t.Errorf("expected SchemaVersion 1, got %d", cfg.SchemaVersion)
	}
	if cfg.MaxSpoolFiles != 1000 {
		t.Errorf("expected MaxSpoolFiles 1000, got %d", cfg.MaxSpoolFiles)
	}
	if cfg.MaxSpoolAgeHours != 24 {
		t.Errorf("expected MaxSpoolAgeHours 24, got %d", cfg.MaxSpoolAgeHours)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "CONFIG")

	content := `
PLATFORM_NAME="tomcat-monitoring-test"
TARGET_CONTAINER="custom-tomcat"
TARGET_ID="lab/custom/test"
GENERATION=2
SCHEMA_VERSION=1
MAX_RECORD_BYTES=8192
DEFAULT_SPOOL_DIR="/tmp/test-spool"
MAX_SPOOL_AGE_HOURS=12
MAX_SPOOL_FILES=500
STALE_TMP_AGE_MINUTES=30
RUN_ONCE="true"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.TargetContainer != "custom-tomcat" {
		t.Errorf("expected TargetContainer 'custom-tomcat', got '%s'", cfg.TargetContainer)
	}
	if cfg.TargetID != "lab/custom/test" {
		t.Errorf("expected TargetID 'lab/custom/test', got '%s'", cfg.TargetID)
	}
	if cfg.Generation != 2 {
		t.Errorf("expected Generation 2, got %d", cfg.Generation)
	}
	if cfg.MaxRecordBytes != 8192 {
		t.Errorf("expected MaxRecordBytes 8192, got %d", cfg.MaxRecordBytes)
	}
	if cfg.SpoolDir != "/tmp/test-spool" {
		t.Errorf("expected SpoolDir '/tmp/test-spool', got '%s'", cfg.SpoolDir)
	}
	if cfg.MaxSpoolAgeHours != 12 {
		t.Errorf("expected MaxSpoolAgeHours 12, got %d", cfg.MaxSpoolAgeHours)
	}
	if cfg.MaxSpoolFiles != 500 {
		t.Errorf("expected MaxSpoolFiles 500, got %d", cfg.MaxSpoolFiles)
	}
	if cfg.StaleTmpAgeMinutes != 30 {
		t.Errorf("expected StaleTmpAgeMinutes 30, got %d", cfg.StaleTmpAgeMinutes)
	}
	if !cfg.RunOnce {
		t.Errorf("expected RunOnce true, got false")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	os.Setenv("TARGET_CONTAINER", "env-tomcat")
	os.Setenv("TARGET_ID", "env/test/01")
	os.Setenv("RUN_ONCE", "true")
	defer func() {
		os.Unsetenv("TARGET_CONTAINER")
		os.Unsetenv("TARGET_ID")
		os.Unsetenv("RUN_ONCE")
	}()

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.TargetContainer != "env-tomcat" {
		t.Errorf("expected TargetContainer 'env-tomcat', got '%s'", cfg.TargetContainer)
	}
	if cfg.TargetID != "env/test/01" {
		t.Errorf("expected TargetID 'env/test/01', got '%s'", cfg.TargetID)
	}
	if !cfg.RunOnce {
		t.Errorf("expected RunOnce true, got false")
	}
}
