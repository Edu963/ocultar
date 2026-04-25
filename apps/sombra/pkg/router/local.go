package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LocalAdapter wraps a local llama.cpp / SLM server.
type LocalAdapter struct {
	name     string
	endpoint string
	client   *http.Client
}

func NewLocal(name, endpoint string) *LocalAdapter {
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}
	return &LocalAdapter{
		name:     name,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 300 * time.Second},
	}
}

func (l *LocalAdapter) Name() string { return l.name }

func (l *LocalAdapter) Endpoint() string { return l.endpoint }

func (l *LocalAdapter) Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error) {
	localMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		localMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	payload := map[string]interface{}{
		"messages": localMessages,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", l.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("local: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("local: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("local: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("local: parse rx: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("local: no choices")
	}
	return result.Choices[0].Message.Content, nil
}

func (l *LocalAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
