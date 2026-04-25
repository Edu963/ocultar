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

// ClaudeAdapter wraps Anthropic Messages API.
type ClaudeAdapter struct {
	name     string
	endpoint string
	keyEnv   string
	client   *http.Client
}

func NewClaude(name, endpoint, keyEnv string) *ClaudeAdapter {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1"
	}
	return &ClaudeAdapter{
		name:     name,
		endpoint: endpoint,
		keyEnv:   keyEnv,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *ClaudeAdapter) Name() string { return c.name }

func (c *ClaudeAdapter) Endpoint() string { return c.endpoint }

func (c *ClaudeAdapter) Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error) {
	apiKey := os.Getenv(c.keyEnv)
	if apiKey == "" {
		return "", fmt.Errorf("claude: env var %q not set", c.keyEnv)
	}

	claudeMessages := make([]map[string]string, 0)
	var systemPrompt string
	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
		} else {
			claudeMessages = append(claudeMessages, map[string]string{"role": m.Role, "content": m.Content})
		}
	}

	payload := map[string]interface{}{
		"model":      c.name,
		"max_tokens": 4096,
		"messages":   claudeMessages,
	}
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("claude: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("claude: parse rx: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("claude: no content")
	}

	var text string
	for _, block := range result.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return text, nil
}

func (c *ClaudeAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
