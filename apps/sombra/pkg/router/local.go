package router

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// SendStream implements Streamer. Local servers (Ollama, llama.cpp, LM Studio)
// speak the OpenAI SSE protocol natively — same parsing as OpenAIAdapter.
func (l *LocalAdapter) SendStream(ctx context.Context, messages []Message, opts ModelOpts, onDelta func(string) error) error {
	localMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		localMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"messages": localMessages,
		"stream":   true,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", l.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("local stream: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("local stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("local stream: HTTP %d: %s", resp.StatusCode, string(errBody))
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
			return fmt.Errorf("local stream read: %w", err)
		}
	}
	return nil
}
