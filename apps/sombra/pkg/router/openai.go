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

// SendStream implements Streamer. It requests server-sent events from the
// OpenAI-compatible endpoint and calls onDelta for each non-empty text chunk.
func (o *OpenAIAdapter) SendStream(ctx context.Context, messages []Message, opts ModelOpts, onDelta func(string) error) error {
	apiKey := os.Getenv(o.keyEnv)
	if apiKey == "" {
		return fmt.Errorf("openai: env var %q not set", o.keyEnv)
	}

	openAIMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		openAIMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	payload := map[string]interface{}{
		"model":    o.name,
		"messages": openAIMessages,
		"stream":   true,
	}
	if opts.Temperature > 0 {
		payload["temperature"] = opts.Temperature
	}
	if opts.MaxTokens > 0 {
		payload["max_tokens"] = opts.MaxTokens
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("openai stream: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// No client-level timeout for streaming; context handles cancellation.
	streamClient := &http.Client{}
	resp, err := streamClient.Do(req)
	if err != nil {
		return fmt.Errorf("openai stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("openai stream: HTTP %d: %s", resp.StatusCode, string(errBody))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if jsonErr := json.Unmarshal([]byte(data), &chunk); jsonErr == nil {
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					if cbErr := onDelta(chunk.Choices[0].Delta.Content); cbErr != nil {
						return cbErr
					}
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("openai stream read: %w", err)
		}
	}
	return nil
}
