package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RemoteScanner implements the AIScanner interface by sending HTTP requests
// to the SLM Sidecar microservice. This decouples CGO dependencies from the 
// high-throughput Gateway paths to prevent production panics/OOMs.
type RemoteScanner struct {
	client     *http.Client
	sidecarURL string
	available  bool
	domain     string
}

// NewRemoteScanner creates a scanner that hits the out-of-process SLM engine.
func NewRemoteScanner(sidecarURL string) *RemoteScanner {
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8085"
	}
	return &RemoteScanner{
		client:     &http.Client{Timeout: 30 * time.Second},
		sidecarURL: sidecarURL,
		available:  true,
	}
}

// ScanForPII forwards the text payload to the SLM sidecar.
func (s *RemoteScanner) ScanForPII(text string) (map[string][]string, error) {
	if !s.available {
		return nil, fmt.Errorf("remote scanner not available (fail-closed)")
	}

	payload := map[string]string{"text": text}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", s.sidecarURL+"/scan", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SLM sidecar unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("SLM sidecar err: HTTP %d", resp.StatusCode)
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse SLM sidecar response: %w", err)
	}

	return result, nil
}

// CheckHealth verifies the sidecar is responsive for Tier 2 scanning.
func (s *RemoteScanner) CheckHealth(host string) {
	resp, err := s.client.Get(s.sidecarURL + "/health")
	if err == nil && resp.StatusCode == 200 {
		s.available = true
	} else {
		s.available = false
	}
}

func (s *RemoteScanner) IsAvailable() bool { return s.available }

func (s *RemoteScanner) SetDomain(domain string) { s.domain = domain }
