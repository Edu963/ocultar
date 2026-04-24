package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PrivacyFilterScanner implements Scanner by delegating to a Python service
// running openai/privacy-filter (bidirectional token classifier, Apache 2.0).
// The service must speak the same HTTP contract as this sidecar:
//
//	POST /scan  {"text": "..."}  →  {"PERSON": ["John"], "EMAIL": ["j@x.com"]}
//	GET  /health                 →  {"status": "ok"}
//
// Set PRIVACY_FILTER_URL to its base URL (default http://localhost:8086).
type PrivacyFilterScanner struct {
	endpoint string
	client   *http.Client
}

func NewPrivacyFilterScanner(endpoint string) (*PrivacyFilterScanner, error) {
	s := &PrivacyFilterScanner{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	resp, err := s.client.Get(endpoint + "/health")
	if err != nil {
		return nil, fmt.Errorf("privacy-filter service unreachable at %s: %w", endpoint, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("privacy-filter health check failed: HTTP %d", resp.StatusCode)
	}
	return s, nil
}

func (s *PrivacyFilterScanner) ScanForPII(text string) (map[string][]string, error) {
	body, _ := json.Marshal(map[string]string{"text": text})
	resp, err := s.client.Post(s.endpoint+"/scan", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("privacy-filter request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("privacy-filter returned HTTP %d", resp.StatusCode)
	}
	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("privacy-filter response parse failed: %w", err)
	}
	return result, nil
}

func (s *PrivacyFilterScanner) Close() {}
