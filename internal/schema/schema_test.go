package schema

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidEventRecords(t *testing.T) {
	now := time.Now().UTC()

	// 1. Container state record
	r1 := NewEventRecord(
		"container_state",
		"lab/tomcat-01/default",
		1,
		StatusCollected,
		StrengthDirect,
		ContainerStateValue{State: "running"},
		now,
	)
	if err := ValidateRecord(r1); err != nil {
		t.Fatalf("Validation failed for valid container_state record: %v", err)
	}

	data, err := r1.MarshalIndent()
	if err != nil {
		t.Fatalf("Failed to marshal record: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if parsed["schema_version"] != float64(1) {
		t.Errorf("expected schema_version 1, got %v", parsed["schema_version"])
	}
	if parsed["type"] != "container_state" {
		t.Errorf("expected type container_state, got %v", parsed["type"])
	}
	if parsed["target_id"] != "lab/tomcat-01/default" {
		t.Errorf("expected target_id lab/tomcat-01/default, got %v", parsed["target_id"])
	}

	// 2. Runtime OOM record
	r2 := NewEventRecord(
		"runtime_oom",
		"lab/tomcat-01/default",
		1,
		StatusCollected,
		StrengthDirect,
		RuntimeOOMValue{OOMKilled: false, ExitCode: 0},
		now,
	)
	if err := ValidateRecord(r2); err != nil {
		t.Fatalf("Validation failed for valid runtime_oom record: %v", err)
	}

	// 3. Collector status record
	r3 := NewEventRecord(
		"collector_status",
		"lab/tomcat-01/default",
		1,
		StatusUnavailable,
		StrengthContextual,
		CollectorStatusValue{Error: "container_engine_unreachable"},
		now,
	)
	if err := ValidateRecord(r3); err != nil {
		t.Fatalf("Validation failed for valid collector_status record: %v", err)
	}
}

func TestInvalidEventRecords(t *testing.T) {
	now := time.Now().UTC()

	// Missing type
	r1 := &EventRecord{
		SchemaVersion: 1,
		TargetID:      "lab/test",
		ObservedAt:    now.Format(time.RFC3339),
		Status:        StatusCollected,
		Strength:      StrengthDirect,
		Value:         map[string]string{"state": "running"},
	}
	if err := ValidateRecord(r1); err == nil {
		t.Errorf("expected error for missing type, got nil")
	}

	// Invalid status
	r2 := &EventRecord{
		SchemaVersion: 1,
		Type:          "container_state",
		TargetID:      "lab/test",
		ObservedAt:    now.Format(time.RFC3339),
		Status:        "unknown_status",
		Strength:      StrengthDirect,
		Value:         map[string]string{"state": "running"},
	}
	if err := ValidateRecord(r2); err == nil {
		t.Errorf("expected error for invalid status, got nil")
	}

	// Invalid date-time
	r3 := &EventRecord{
		SchemaVersion: 1,
		Type:          "container_state",
		TargetID:      "lab/test",
		ObservedAt:    "invalid-date",
		Status:        StatusCollected,
		Strength:      StrengthDirect,
		Value:         map[string]string{"state": "running"},
	}
	if err := ValidateRecord(r3); err == nil {
		t.Errorf("expected error for invalid observed_at, got nil")
	}
}
