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

// OpenAIAdapter wraps the OpenAI-compatible chat completions API.
type OpenAIAdapter struct {
	name     string
	endpoint string
	keyEnv   string
	client   *http.Client
}

// NewOpenAI creates an adapter for OpenAI.
func NewOpenAI(name, endpoint, keyEnv string) *OpenAIAdapter {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	return &OpenAIAdapter{
		name:     name,
		endpoint: endpoint,
		keyEnv:   keyEnv,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (o *OpenAIAdapter) Name() string { return o.name }

func (o *OpenAIAdapter) Endpoint() string { return o.endpoint }

func (o *OpenAIAdapter) Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error) {
	apiKey := os.Getenv(o.keyEnv)
	if apiKey == "" {
		return "", fmt.Errorf("openai: env var %q not set", o.keyEnv)
	}

	openAIMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		openAIMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	payload := map[string]interface{}{
		"model":    o.name,
		"messages": openAIMessages,
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

		req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("openai: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := o.client.Do(req)
		if err != nil {
			lastErr = err
			continue // network error, retry
		}
		
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, string(respBody))
			continue // rate limit or server error, retry
		}

		if resp.StatusCode >= 400 {
			return "", fmt.Errorf("openai: HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("openai: parse rx: %w", err)
		}
		if len(result.Choices) == 0 {
			return "", fmt.Errorf("openai: no choices")
		}
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("openai failed after %d retries, last error: %w", maxRetries, lastErr)
}

func (o *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
