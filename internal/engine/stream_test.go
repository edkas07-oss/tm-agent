package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEventMessageNormalization(t *testing.T) {
	// Docker format event
	dockerJSON := `{
		"status": "die",
		"id": "abc123456789",
		"from": "localhost/tomcat-jmx-exporter:1.0.0",
		"Type": "container",
		"Action": "die",
		"Actor": {
			"ID": "abc123456789",
			"Attributes": {
				"name": "tomcat-jmx-exporter",
				"image": "localhost/tomcat-jmx-exporter:1.0.0",
				"exitCode": "143"
			}
		},
		"scope": "local",
		"time": 1726272000,
		"timeNano": 1726272000000000000
	}`

	var msg1 EventMessage
	if err := json.Unmarshal([]byte(dockerJSON), &msg1); err != nil {
		t.Fatalf("Failed to unmarshal docker event JSON: %v", err)
	}

	if msg1.GetAction() != "die" {
		t.Errorf("expected action 'die', got '%s'", msg1.GetAction())
	}
	if msg1.GetContainerName() != "tomcat-jmx-exporter" {
		t.Errorf("expected container name 'tomcat-jmx-exporter', got '%s'", msg1.GetContainerName())
	}

	// Podman compat event
	podmanJSON := `{
		"Action": "died",
		"Name": "tomcat-jmx-exporter",
		"Type": "container",
		"time": 1726272000
	}`

	var msg2 EventMessage
	if err := json.Unmarshal([]byte(podmanJSON), &msg2); err != nil {
		t.Fatalf("Failed to unmarshal podman event JSON: %v", err)
	}

	if msg2.GetAction() != "died" {
		t.Errorf("expected action 'died', got '%s'", msg2.GetAction())
	}
	if msg2.GetContainerName() != "tomcat-jmx-exporter" {
		t.Errorf("expected container name 'tomcat-jmx-exporter', got '%s'", msg2.GetContainerName())
	}
}

func TestStreamingEventParsing(t *testing.T) {
	streamData := strings.TrimSpace(`
{"Action":"stop","Actor":{"ID":"123","Attributes":{"name":"tomcat-jmx-exporter"}},"Type":"container","time":1000}
{"Action":"die","Actor":{"ID":"123","Attributes":{"name":"tomcat-jmx-exporter"}},"Type":"container","time":1001}
{"Action":"start","Actor":{"ID":"123","Attributes":{"name":"tomcat-jmx-exporter"}},"Type":"container","time":1002}
`)

	dec := json.NewDecoder(strings.NewReader(streamData))
	actions := []string{}
	for dec.More() {
		var ev EventMessage
		if err := dec.Decode(&ev); err == nil {
			actions = append(actions, ev.GetAction())
		}
	}

	if len(actions) != 3 {
		t.Fatalf("Expected 3 parsed events, got %d", len(actions))
	}
	if actions[0] != "stop" || actions[1] != "die" || actions[2] != "start" {
		t.Errorf("Unexpected action sequence: %v", actions)
	}
}
