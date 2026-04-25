package connector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// APIConnector is a generic REST API adapter.
// It makes authenticated HTTP requests to external services (banks, medical
// portals, financial APIs) and returns the raw response for PII scrubbing.
type APIConnector struct {
	name     string
	endpoint string
	authType string // "bearer", "api_key", "none"
	keyEnv   string // env var holding the API key / token
	policy   DataPolicy
	client   *http.Client
}

// APIConnectorConfig holds the YAML-driven configuration for an API connector.
type APIConnectorConfig struct {
	Name     string     `yaml:"name"`
	Endpoint string     `yaml:"endpoint"`
	AuthType string     `yaml:"auth_type"`
	KeyEnv   string     `yaml:"api_key_env"`
	Policy   DataPolicy `yaml:"policy"`
}

// NewAPIConnector creates an API connector from its configuration.
func NewAPIConnector(cfg APIConnectorConfig) *APIConnector {
	return &APIConnector{
		name:     cfg.Name,
		endpoint: cfg.Endpoint,
		authType: cfg.AuthType,
		keyEnv:   cfg.KeyEnv,
		policy:   cfg.Policy,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *APIConnector) Name() string       { return a.name }
func (a *APIConnector) Policy() DataPolicy { return a.policy }

// Fetch makes an HTTP GET to the configured endpoint, appending the SourceID
// as a path suffix. Authentication is applied based on authType.
func (a *APIConnector) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	url := a.endpoint
	if req.SourceID != "" {
		url = fmt.Sprintf("%s/%s", a.endpoint, req.SourceID)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("api connector %q: build request: %w", a.name, err)
	}

	// Apply authentication.
	if err := a.applyAuth(httpReq); err != nil {
		return nil, err
	}

	// Add any custom parameters as query params.
	if len(req.Parameters) > 0 {
		q := httpReq.URL.Query()
		for k, v := range req.Parameters {
			q.Set(k, v)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("api connector %q: request failed: %w", a.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("api connector %q: HTTP %d: %s", a.name, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api connector %q: read body: %w", a.name, err)
	}

	// Enforce size limit.
	if a.policy.MaxBodyBytes > 0 && int64(len(body)) > a.policy.MaxBodyBytes {
		return nil, fmt.Errorf("api connector %q: response size %d exceeds policy limit %d", a.name, len(body), a.policy.MaxBodyBytes)
	}

	return &FetchResponse{
		ContentType: resp.Header.Get("Content-Type"),
		Body:        body,
		Metadata: map[string]string{
			"source":      a.endpoint,
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		},
	}, nil
}

// applyAuth sets the appropriate auth headers on the request.
func (a *APIConnector) applyAuth(req *http.Request) error {
	switch a.authType {
	case "bearer":
		token := os.Getenv(a.keyEnv)
		if token == "" {
			return fmt.Errorf("api connector %q: env var %q is not set", a.name, a.keyEnv)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case "api_key":
		key := os.Getenv(a.keyEnv)
		if key == "" {
			return fmt.Errorf("api connector %q: env var %q is not set", a.name, a.keyEnv)
		}
		req.Header.Set("X-API-Key", key)
	case "none", "":
		// No auth required.
	default:
		return fmt.Errorf("api connector %q: unknown auth_type %q", a.name, a.authType)
	}
	return nil
}
