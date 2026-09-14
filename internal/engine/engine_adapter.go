package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RESTEngineAdapter implements EngineClient over Container Engine REST API.
type RESTEngineAdapter struct {
	client  *http.Client
	info    *EngineInfo
	baseURL string
}

func (a *RESTEngineAdapter) GetInfo() *EngineInfo {
	return a.info
}

func (a *RESTEngineAdapter) Ping(ctx context.Context) (*EngineInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/_ping", nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("engine ping failed at socket %s: %w", a.info.SocketPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("engine ping returned non-200 status: %d", resp.StatusCode)
	}

	// Fetch version info
	vReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/version", nil)
	if vResp, vErr := a.client.Do(vReq); vErr == nil {
		defer vResp.Body.Close()
		var vData struct {
			Version    string `json:"Version"`
			APIVersion string `json:"ApiVersion"`
			OSType     string `json:"Os"`
		}
		if json.NewDecoder(vResp.Body).Decode(&vData) == nil {
			a.info.APIVersion = vData.APIVersion
			if vData.OSType != "" {
				a.info.OSType = vData.OSType
			}
		}
	}

	return a.info, nil
}

func (a *RESTEngineAdapter) InspectContainer(ctx context.Context, nameOrID string) (*ContainerInspect, error) {
	endpoint := fmt.Sprintf("%s/containers/%s/json", a.baseURL, url.PathEscape(nameOrID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Container not found
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("inspect container %s failed (%d): %s", nameOrID, resp.StatusCode, string(body))
	}

	var inspect ContainerInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		return nil, fmt.Errorf("failed to decode container inspect response: %w", err)
	}
	return &inspect, nil
}

func (a *RESTEngineAdapter) StreamEvents(ctx context.Context, targetContainer string) (<-chan EventMessage, <-chan error) {
	eventCh := make(chan EventMessage, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		// Build filter parameters
		filterMap := make(map[string][]string)
		if targetContainer != "" {
			filterMap["container"] = []string{targetContainer}
		}
		filterMap["type"] = []string{"container"}

		filterBytes, _ := json.Marshal(filterMap)
		endpoint := fmt.Sprintf("%s/events?filters=%s", a.baseURL, url.QueryEscape(string(filterBytes)))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			errCh <- fmt.Errorf("failed to create events request: %w", err)
			return
		}

		resp, err := a.client.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("failed to connect to event stream: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("event stream returned status %d: %s", resp.StatusCode, string(body))
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF || ctx.Err() != nil {
						return
					}
					errCh <- fmt.Errorf("error reading event stream: %w", err)
					return
				}

				lineStr := strings.TrimSpace(string(line))
				if lineStr == "" {
					continue
				}

				var event EventMessage
				if err := json.Unmarshal([]byte(lineStr), &event); err != nil {
					// In some Podman versions, events are JSON objects without trailing newlines or formatted differently
					dec := json.NewDecoder(strings.NewReader(lineStr))
					if decErr := dec.Decode(&event); decErr != nil {
						continue
					}
				}

				eventCh <- event
			}
		}
	}()

	return eventCh, errCh
}
