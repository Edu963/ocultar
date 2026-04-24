package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PrivacyFilterEngine implements Tier2Engine by delegating to a Python service
// running openai/privacy-filter (bidirectional token classifier, Apache 2.0).
// The service must speak the same HTTP contract as this sidecar:
//
//	POST /scan  {"text": "..."}  →  {"PERSON": ["John"], "EMAIL": ["j@x.com"]}
//	GET  /health                 →  {"status": "ok"}
//
// Set PYTHON_SIDECAR_URL to its base URL (default http://localhost:8086).
type PrivacyFilterEngine struct {
	endpoint  string
	modelPath string
	client    *http.Client
}

func NewPrivacyFilterEngine(endpoint string, modelPath string) (*PrivacyFilterEngine, error) {
	s := &PrivacyFilterEngine{
		endpoint:  endpoint,
		modelPath: modelPath,
		client:    &http.Client{Timeout: 10 * time.Second},
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

func (s *PrivacyFilterEngine) ScanForPII(text string) (map[string][]string, error) {
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

func (s *PrivacyFilterEngine) Name() string {
	return fmt.Sprintf("privacy-filter (%s)", s.modelPath)
}

func (s *PrivacyFilterEngine) Close() {}
