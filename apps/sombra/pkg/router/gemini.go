package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GeminiAdapter wraps Google Gemini API.
type GeminiAdapter struct {
	name     string
	endpoint string
	keyEnv   string
	client   *http.Client
}

func NewGemini(name, endpoint, keyEnv string) *GeminiAdapter {
	if endpoint == "" {
		endpoint = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &GeminiAdapter{
		name:     name,
		endpoint: endpoint,
		keyEnv:   keyEnv,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (g *GeminiAdapter) Name() string { return g.name }

func (g *GeminiAdapter) Endpoint() string { return g.endpoint }

func (g *GeminiAdapter) Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error) {
	apiKey := os.Getenv(g.keyEnv)
	if apiKey == "" {
		return "", fmt.Errorf("gemini: env var %q not set", g.keyEnv)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.endpoint, g.name, apiKey)
	
	contents := make([]map[string]interface{}, len(messages))
	for i, m := range messages {
		role := m.Role
		if role == "assistant" || role == "system" {
			role = "model"
		}
		contents[i] = map[string]interface{}{
			"role": role,
			"parts": []map[string]string{{"text": m.Content}},
		}
	}

	payload := map[string]interface{}{
		"contents": contents,
	}

	body, _ := json.Marshal(payload)

	var lastErr error
	backoff := 1 * time.Second
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // exponential backoff
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("gemini: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := g.client.Do(req)
		if err != nil {
			lastErr = err
			continue // network error, retry
		}
		
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(respBody))
			continue // rate limit or server error, retry
		}

		if resp.StatusCode >= 400 {
			return "", fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		var result struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("gemini: parse rx: %w", err)
		}
		if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("gemini: no content")
		}
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("gemini failed after %d retries, last error: %w", maxRetries, lastErr)
}

func (g *GeminiAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
