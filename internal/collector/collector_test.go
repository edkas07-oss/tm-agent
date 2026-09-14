package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eddywiyatno/tm-agent/internal/config"
	"github.com/eddywiyatno/tm-agent/internal/engine"
	"github.com/eddywiyatno/tm-agent/internal/schema"
)

// MockEngineClient simulates container engine inspect and event streams.
type MockEngineClient struct {
	inspectResult *engine.ContainerInspect
	inspectErr    error
	events        []engine.EventMessage
	info          *engine.EngineInfo
}

func (m *MockEngineClient) Ping(ctx context.Context) (*engine.EngineInfo, error) {
	return m.info, nil
}

func (m *MockEngineClient) InspectContainer(ctx context.Context, nameOrID string) (*engine.ContainerInspect, error) {
	return m.inspectResult, m.inspectErr
}

func (m *MockEngineClient) StreamEvents(ctx context.Context, targetContainer string) (<-chan engine.EventMessage, <-chan error) {
	eventCh := make(chan engine.EventMessage, len(m.events))
	errCh := make(chan error, 1)

	for _, ev := range m.events {
		eventCh <- ev
	}
	close(eventCh)

	return eventCh, errCh
}

func (m *MockEngineClient) GetInfo() *engine.EngineInfo {
	return m.info
}

func TestCollectorRecordSnapshotRunningContainer(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	cfg := &config.Config{
		PlatformName:       "tomcat-monitoring",
		TargetContainer:    "tomcat-jmx-exporter",
		TargetID:           "lab/tomcat-01/default",
		Generation:         1,
		SchemaVersion:      1,
		MaxRecordBytes:     16384,
		SpoolDir:           spoolDir,
		MaxSpoolAgeHours:   24,
		MaxSpoolFiles:      1000,
		StaleTmpAgeMinutes: 60,
	}

	inspect := &engine.ContainerInspect{
		ID:   "container-12345",
		Name: "tomcat-jmx-exporter",
	}
	inspect.State.Running = true
	inspect.State.OOMKilled = false
	inspect.State.ExitCode = 0

	mockEngine := &MockEngineClient{
		inspectResult: inspect,
		info: &engine.EngineInfo{
			EngineType: "podman",
			SocketPath: "/tmp/mock.sock",
		},
	}

	c := NewCollector(cfg, mockEngine)
	files, err := c.RecordSnapshot(context.Background())
	if err != nil {
		t.Fatalf("RecordSnapshot failed: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 written files (container_state and runtime_oom), got %d", len(files))
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("Failed to read created spool file %s: %v", f, err)
		}

		var record schema.EventRecord
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatalf("Failed to parse record JSON: %v", err)
		}

		if err := schema.ValidateRecord(&record); err != nil {
			t.Fatalf("Record validation failed: %v", err)
		}

		if record.TargetID != "lab/tomcat-01/default" {
			t.Errorf("Unexpected target_id: %s", record.TargetID)
		}
	}
}

func TestCollectorRecordSnapshotNotFoundContainer(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	cfg := &config.Config{
		PlatformName:       "tomcat-monitoring",
		TargetContainer:    "non-existent-target",
		TargetID:           "lab/tomcat-01/default",
		Generation:         1,
		SchemaVersion:      1,
		MaxRecordBytes:     16384,
		SpoolDir:           spoolDir,
		MaxSpoolAgeHours:   24,
		MaxSpoolFiles:      1000,
		StaleTmpAgeMinutes: 60,
	}

	mockEngine := &MockEngineClient{
		inspectResult: nil, // not found
		info: &engine.EngineInfo{
			EngineType: "podman",
			SocketPath: "/tmp/mock.sock",
		},
	}

	c := NewCollector(cfg, mockEngine)
	files, err := c.RecordSnapshot(context.Background())
	if err != nil {
		t.Fatalf("RecordSnapshot failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 written file for not_found, got %d", len(files))
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var record schema.EventRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Failed to parse record: %v", err)
	}

	if record.Type != "container_state" || record.Status != schema.StatusNotFound {
		t.Errorf("Expected container_state with status not_found, got type=%s status=%s", record.Type, record.Status)
	}
}

func TestCollectorRecordSnapshotEngineError(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	cfg := &config.Config{
		PlatformName:       "tomcat-monitoring",
		TargetContainer:    "tomcat-jmx-exporter",
		TargetID:           "lab/tomcat-01/default",
		Generation:         1,
		SchemaVersion:      1,
		MaxRecordBytes:     16384,
		SpoolDir:           spoolDir,
		MaxSpoolAgeHours:   24,
		MaxSpoolFiles:      1000,
		StaleTmpAgeMinutes: 60,
	}

	mockEngine := &MockEngineClient{
		inspectErr: fmt.Errorf("connection refused"),
		info: &engine.EngineInfo{
			EngineType: "podman",
			SocketPath: "/tmp/mock.sock",
		},
	}

	c := NewCollector(cfg, mockEngine)
	files, err := c.RecordSnapshot(context.Background())
	if err != nil {
		t.Fatalf("RecordSnapshot failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 written file for collector_status, got %d", len(files))
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var record schema.EventRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Failed to parse record: %v", err)
	}

	if record.Type != "collector_status" || record.Status != schema.StatusUnavailable {
		t.Errorf("Expected collector_status with status unavailable, got type=%s status=%s", record.Type, record.Status)
	}
}

func TestCollectorRunOnce(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	cfg := &config.Config{
		PlatformName:       "tomcat-monitoring",
		TargetContainer:    "tomcat-jmx-exporter",
		TargetID:           "lab/tomcat-01/default",
		Generation:         1,
		SchemaVersion:      1,
		MaxRecordBytes:     16384,
		SpoolDir:           spoolDir,
		MaxSpoolAgeHours:   24,
		MaxSpoolFiles:      1000,
		StaleTmpAgeMinutes: 60,
		RunOnce:            true,
	}

	mockEngine := &MockEngineClient{
		inspectResult: nil,
		info: &engine.EngineInfo{
			EngineType: "podman",
			SocketPath: "/tmp/mock.sock",
		},
	}

	c := NewCollector(cfg, mockEngine)
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run once failed: %v", err)
	}

	entries, err := os.ReadDir(spoolDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("Expected spool entries written, found none")
	}
}
