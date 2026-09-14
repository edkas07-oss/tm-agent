package engine

import "time"

// ContainerInspect represents container state returned by Engine Inspect API.
type ContainerInspect struct {
	ID      string `json:"Id"`
	Name    string `json:"Name"`
	Created string `json:"Created"`
	State   struct {
		Status     string    `json:"Status"`
		Running    bool      `json:"Running"`
		Paused     bool      `json:"Paused"`
		Restarting bool      `json:"Restarting"`
		OOMKilled  bool      `json:"OOMKilled"`
		Dead       bool      `json:"Dead"`
		Pid        int       `json:"Pid"`
		ExitCode   int       `json:"ExitCode"`
		Error      string    `json:"Error"`
		StartedAt  time.Time `json:"StartedAt"`
		FinishedAt time.Time `json:"FinishedAt"`
	} `json:"State"`
	Config struct {
		Image  string            `json:"Image"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
}

// EventActor represents event actor details in Docker/Podman events.
type EventActor struct {
	ID         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}

// EventMessage represents a raw event payload streamed from the container engine.
type EventMessage struct {
	Type     string     `json:"Type"`
	Action   string     `json:"Action"`
	Status   string     `json:"status,omitempty"`
	Actor    EventActor `json:"Actor"`
	ID       string     `json:"id,omitempty"`
	From     string     `json:"from,omitempty"`
	Time     int64      `json:"time"`
	TimeNano int64      `json:"timeNano"`
	// Libpod specific compatibility fields
	Name string `json:"Name,omitempty"`
}

// GetAction returns normalized action string (e.g. "die", "stop", "oom", "start").
func (e *EventMessage) GetAction() string {
	if e.Action != "" {
		return e.Action
	}
	if e.Status != "" {
		return e.Status
	}
	return ""
}

// GetContainerName returns container name from Attributes or Name field.
func (e *EventMessage) GetContainerName() string {
	if e.Actor.Attributes != nil {
		if name, ok := e.Actor.Attributes["name"]; ok && name != "" {
			return name
		}
	}
	if e.Name != "" {
		return e.Name
	}
	return ""
}

// EngineInfo holds engine metadata.
type EngineInfo struct {
	EngineType string // "podman" or "docker"
	APIVersion string
	OSType     string
	SocketPath string
}
