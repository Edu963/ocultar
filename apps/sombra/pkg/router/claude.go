package router

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

// SendStream implements Streamer using Anthropic's streaming Messages API.
// Only content_block_delta events with type=text_delta are forwarded.
func (c *ClaudeAdapter) SendStream(ctx context.Context, messages []Message, opts ModelOpts, onDelta func(string) error) error {
	apiKey := os.Getenv(c.keyEnv)
	if apiKey == "" {
		return fmt.Errorf("claude stream: env var %q not set", c.keyEnv)
	}

	claudeMessages := make([]map[string]string, 0, len(messages))
	var systemPrompt string
	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
		} else {
			claudeMessages = append(claudeMessages, map[string]string{"role": m.Role, "content": m.Content})
		}
	}

	maxTokens := 4096
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	payload := map[string]interface{}{
		"model":      c.name,
		"max_tokens": maxTokens,
		"messages":   claudeMessages,
		"stream":     true,
	}
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("claude stream: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return fmt.Errorf("claude stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("claude stream: HTTP %d: %s", resp.StatusCode, string(errBody))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		// Anthropic SSE: "data: " lines carry the payload; "event: " lines carry the type.
		// We only need content_block_delta events — identified by the data payload's "type" field.
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if jsonErr := json.Unmarshal([]byte(data), &event); jsonErr == nil {
				if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					if cbErr := onDelta(event.Delta.Text); cbErr != nil {
						return cbErr
					}
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("claude stream read: %w", err)
		}
	}
	return nil
}
