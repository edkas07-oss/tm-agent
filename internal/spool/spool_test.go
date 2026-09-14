package spool

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/eddywiyatno/tm-agent/internal/schema"
)

func TestWriteRecordAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	now := time.Now().UTC()
	record := schema.NewEventRecord(
		"container_state",
		"lab/tomcat-01/default",
		1,
		schema.StatusCollected,
		schema.StrengthDirect,
		schema.ContainerStateValue{State: "running"},
		now,
	)

	filePath, err := WriteRecord(spoolDir, record, 16384)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Expected final file %s to exist", filePath)
	}

	// Verify no .tmp files remain
	tmpFiles, err := filepath.Glob(filepath.Join(spoolDir, "*.tmp"))
	if err != nil || len(tmpFiles) > 0 {
		t.Errorf("Found left-over .tmp files: %v", tmpFiles)
	}

	// Permissions check on Unix
	if runtime.GOOS != "windows" {
		dirInfo, err := os.Stat(spoolDir)
		if err != nil {
			t.Fatalf("Failed to stat spool dir: %v", err)
		}
		if perm := dirInfo.Mode().Perm(); perm != 0700 {
			t.Errorf("Expected spool dir permission 0700, got %o", perm)
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat final file: %v", err)
		}
		if perm := fileInfo.Mode().Perm(); perm != 0600 {
			t.Errorf("Expected file permission 0600, got %o", perm)
		}
	}
}

func TestWriteRecordSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")

	// Payload that exceeds 100 bytes with limit 100
	hugeMap := make(map[string]string)
	for i := 0; i < 50; i++ {
		hugeMap[fmt.Sprintf("key_%d", i)] = "a large string value to exceed size"
	}

	record := schema.NewEventRecord(
		"large_event",
		"lab/tomcat-01/default",
		1,
		schema.StatusCollected,
		schema.StrengthDirect,
		hugeMap,
		time.Now().UTC(),
	)

	_, err := WriteRecord(spoolDir, record, 100)
	if err == nil {
		t.Errorf("Expected error for oversized record, got nil")
	}

	// Verify no file remained
	files, _ := os.ReadDir(spoolDir)
	if len(files) > 0 {
		t.Errorf("Expected no files created for rejected record, found %d", len(files))
	}
}

func TestPruneStaleTmpAndJson(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")
	if err := EnsureSpoolDir(spoolDir); err != nil {
		t.Fatalf("EnsureSpoolDir failed: %v", err)
	}

	// 1. Create a stale .tmp file (2 hours old)
	staleTmp := filepath.Join(spoolDir, "1000000000_stale.tmp")
	_ = os.WriteFile(staleTmp, []byte("stale tmp"), 0600)
	oldTime := time.Now().Add(-2 * time.Hour)
	_ = os.Chtimes(staleTmp, oldTime, oldTime)

	// 2. Create a fresh .tmp file (5 mins old)
	freshTmp := filepath.Join(spoolDir, "9999999999_fresh.tmp")
	_ = os.WriteFile(freshTmp, []byte("fresh tmp"), 0600)

	// 3. Create a stale .json file (48 hours old)
	staleJson := filepath.Join(spoolDir, "1000000000_stale.json")
	_ = os.WriteFile(staleJson, []byte("{}"), 0600)
	veryOldTime := time.Now().Add(-48 * time.Hour)
	_ = os.Chtimes(staleJson, veryOldTime, veryOldTime)

	// 4. Create a fresh .json file (10 mins old)
	freshJson := filepath.Join(spoolDir, "9999999999_fresh.json")
	_ = os.WriteFile(freshJson, []byte("{}"), 0600)

	report, err := PruneSpool(spoolDir, 24, 1000, 60)
	if err != nil {
		t.Fatalf("PruneSpool failed: %v", err)
	}

	if report.StaleTmpPruned != 1 {
		t.Errorf("Expected 1 stale tmp pruned, got %d", report.StaleTmpPruned)
	}
	if report.StaleJsonPruned != 1 {
		t.Errorf("Expected 1 stale json pruned, got %d", report.StaleJsonPruned)
	}

	if _, err := os.Stat(staleTmp); !os.IsNotExist(err) {
		t.Errorf("Expected stale .tmp to be deleted")
	}
	if _, err := os.Stat(freshTmp); os.IsNotExist(err) {
		t.Errorf("Expected fresh .tmp to be preserved")
	}
	if _, err := os.Stat(staleJson); !os.IsNotExist(err) {
		t.Errorf("Expected stale .json to be deleted")
	}
	if _, err := os.Stat(freshJson); os.IsNotExist(err) {
		t.Errorf("Expected fresh .json to be preserved")
	}
}

func TestPruneFIFOQuota(t *testing.T) {
	tmpDir := t.TempDir()
	spoolDir := filepath.Join(tmpDir, "spool")
	if err := EnsureSpoolDir(spoolDir); err != nil {
		t.Fatalf("EnsureSpoolDir failed: %v", err)
	}

	// Create 8 files
	for i := 1; i <= 8; i++ {
		fileName := filepath.Join(spoolDir, fmt.Sprintf("200000000%d_event.json", i))
		_ = os.WriteFile(fileName, []byte("{}"), 0600)
	}

	// Enforce quota of 5 files
	report, err := PruneSpool(spoolDir, 24, 5, 60)
	if err != nil {
		t.Fatalf("PruneSpool failed: %v", err)
	}

	if report.QuotaPruned != 3 {
		t.Errorf("Expected 3 files quota-pruned, got %d", report.QuotaPruned)
	}
	if report.TotalRemaining != 5 {
		t.Errorf("Expected 5 remaining files, got %d", report.TotalRemaining)
	}

	// Oldest files (1, 2, 3) must be deleted
	for i := 1; i <= 3; i++ {
		oldFile := filepath.Join(spoolDir, fmt.Sprintf("200000000%d_event.json", i))
		if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
			t.Errorf("Expected oldest file %s to be deleted by FIFO pruning", oldFile)
		}
	}

	// Newer files (4, 5, 6, 7, 8) must exist
	for i := 4; i <= 8; i++ {
		newFile := filepath.Join(spoolDir, fmt.Sprintf("200000000%d_event.json", i))
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			t.Errorf("Expected newer file %s to be preserved", newFile)
		}
	}
}
