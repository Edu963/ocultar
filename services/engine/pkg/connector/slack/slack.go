package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Edu963/ocultar/pkg/connector"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/license"
)

func init() {
	connector.Register("slack", func() connector.Connector {
		return &SlackConnector{}
	})
}

// SlackConnector implements the Connector interface for Slack workspace ingestion.
type SlackConnector struct {
	id     string
	engine *engine.Engine
	config map[string]interface{}
	client *http.Client
}

func (s *SlackConnector) ID() string   { return s.id }
func (s *SlackConnector) Type() string { return "slack" }

func (s *SlackConnector) Init(config map[string]interface{}, eng *engine.Engine) error {
	s.id = fmt.Sprintf("slack-%v", config["workspace_id"])
	s.engine = eng
	s.config = config
	s.client = &http.Client{Timeout: 15 * time.Second}

	if _, ok := config["token"]; !ok {
		return fmt.Errorf("slack connector requires 'token'")
	}

	// Fail-Closed License Check: Slack requires Enterprise Tier + Bit 0 (CapProSlack)
	if !license.HasProConnector(license.CapProSlack) {
		return fmt.Errorf("slack connector: requires Enterprise License with Slack capability enabled (Fail-Closed)")
	}

	return nil
}

func (s *SlackConnector) Start() error {
	log.Printf("[INFO] Slack Connector %q initialized (on-demand pull mode).", s.id)
	return nil
}

func (s *SlackConnector) Stop() error {
	return nil
}

// Fetch implements the on-demand data pull for Slack using conversations.history.
func (s *SlackConnector) Fetch(ctx context.Context, params map[string]interface{}) ([]byte, error) {
	channelID, ok := params["channel_id"].(string)
	if !ok || channelID == "" {
		// Fallback to source_id if present (Sombra often uses source_id for the primary resource)
		if sid, ok := params["source_id"].(string); ok && sid != "" {
			channelID = sid
		}
	}

	if channelID == "" {
		return nil, fmt.Errorf("slack fetch: channel_id is required")
	}

	token, ok := s.config["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("slack fetch: token is missing from configuration")
	}

	url := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s&limit=20", channelID)

	log.Printf("[INFO] Slack Connector %q pulling history for channel %s...", s.id, channelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create slack request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("slack api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read slack response: %w", err)
	}

	// Check if Slack returned an error in the JSON (Slack often returns 200 OK with "ok": false)
	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &slackResp); err == nil {
		if !slackResp.OK {
			return nil, fmt.Errorf("slack api error: %s", slackResp.Error)
		}
	}

	// SECURITY CHECK: We return the raw JSON to Sombra.
	// We DO NOT redact here. Redaction is Sombra's responsibility.
	return body, nil
}
