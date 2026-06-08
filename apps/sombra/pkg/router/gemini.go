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

// GeminiAdapter wraps Google Gemini API.
type GeminiAdapter struct {
	name       string // router key sent by clients in the "model" field
	apiModelID string // actual Google Gemini API model identifier used in the URL
	endpoint   string
	keyEnv     string
	client     *http.Client
}

// NewGemini creates a Gemini adapter. apiModelID is the Google-side model name
// (e.g. "gemini-2.0-flash-latest"); if empty it falls back to name.
func NewGemini(name, endpoint, keyEnv, apiModelID string) *GeminiAdapter {
	if endpoint == "" {
		endpoint = "https://generativelanguage.googleapis.com/v1beta"
	}
	if apiModelID == "" {
		apiModelID = name
	}
	return &GeminiAdapter{
		name:       name,
		apiModelID: apiModelID,
		endpoint:   endpoint,
		keyEnv:     keyEnv,
		client:     &http.Client{Timeout: 120 * time.Second},
	}
}

func (g *GeminiAdapter) Name() string { return g.name }

func (g *GeminiAdapter) Endpoint() string { return g.endpoint }

func (g *GeminiAdapter) Send(ctx context.Context, messages []Message, opts ModelOpts) (string, error) {
	apiKey := os.Getenv(g.keyEnv)
	if apiKey == "" {
		return "", fmt.Errorf("gemini: env var %q not set", g.keyEnv)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.endpoint, g.apiModelID, apiKey)
	
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

		if resp.StatusCode == 429 {
			// Quota-exceeded is permanent — bail immediately instead of burning retries
			if bytes.Contains(respBody, []byte("quota")) {
				return "", fmt.Errorf("gemini: HTTP 429: %s", string(respBody))
			}
			lastErr = fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(respBody))
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("gemini: HTTP %d: %s", resp.StatusCode, string(respBody))
			continue
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

// SendStream implements Streamer using Gemini's streamGenerateContent endpoint.
// Adding ?alt=sse requests SSE-format output — identical wire format to OpenAI/Claude,
// so parsing is consistent across all three adapters.
func (g *GeminiAdapter) SendStream(ctx context.Context, messages []Message, opts ModelOpts, onDelta func(string) error) error {
	apiKey := os.Getenv(g.keyEnv)
	if apiKey == "" {
		return fmt.Errorf("gemini stream: env var %q not set", g.keyEnv)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s",
		g.endpoint, g.apiModelID, apiKey)

	contents := make([]map[string]interface{}, len(messages))
	for i, m := range messages {
		role := m.Role
		if role == "assistant" || role == "system" {
			role = "model"
		}
		contents[i] = map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		}
	}

	body, _ := json.Marshal(map[string]interface{}{"contents": contents})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gemini stream: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("gemini stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("gemini stream: HTTP %d: %s", resp.StatusCode, string(errBody))
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
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}
			if jsonErr := json.Unmarshal([]byte(data), &chunk); jsonErr == nil {
				if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
					if text := chunk.Candidates[0].Content.Parts[0].Text; text != "" {
						if cbErr := onDelta(text); cbErr != nil {
							return cbErr
						}
					}
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("gemini stream read: %w", err)
		}
	}
	return nil
}
