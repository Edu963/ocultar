package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Edu963/ocultar/pkg/refinery"
)

// EventWebhook is an http.Handler for the Slack Events API (push mode).
// It verifies every request with HMAC-SHA256 and passes message content
// through the Refinery before any downstream use.
//
// Mount it at the path you configure in your Slack App's Event Subscriptions:
//
//	http.Handle("/slack/events", connector.NewEventWebhook(signingSecret, eng))
type EventWebhook struct {
	signingSecret string
	refinery      *refinery.Refinery
}

// NewEventWebhook creates an EventWebhook.
// signingSecret is the "Signing Secret" from your Slack App's Basic Information page.
func NewEventWebhook(signingSecret string, eng *refinery.Refinery) *EventWebhook {
	return &EventWebhook{signingSecret: signingSecret, refinery: eng}
}

func (h *EventWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	if err := h.verifySignature(r, body); err != nil {
		log.Printf("[SLACK-WEBHOOK] Signature verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var envelope struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"` // url_verification
		Event     *struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			User    string `json:"user"`
			Channel string `json:"channel"`
			Ts      string `json:"ts"`
			BotID   string `json:"bot_id"` // non-empty for bot messages — skip
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	switch envelope.Type {
	case "url_verification":
		// Slack sends this once when you save the Events subscription URL.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": envelope.Challenge})
		return

	case "event_callback":
		if envelope.Event == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		ev := envelope.Event
		if ev.Type != "message" || ev.BotID != "" || ev.Text == "" {
			w.WriteHeader(http.StatusOK)
			return
		}
		go h.processMessage(ev.Channel, ev.User, ev.Ts, ev.Text)

	default:
		log.Printf("[SLACK-WEBHOOK] Unhandled event type %q", envelope.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EventWebhook) processMessage(channel, user, ts, text string) {
	refined, err := h.refinery.RefineString(text, "slack-events/"+channel, nil)
	if err != nil {
		log.Printf("[SLACK-WEBHOOK] Refinery error for channel %s ts %s: %v", channel, ts, err)
		return
	}
	log.Printf("[SLACK-WEBHOOK] channel=%s user=%s ts=%s original_len=%d refined_len=%d",
		channel, user, ts, len(text), len(refined))
	// refined is the sanitized text — forward to downstream here.
	_ = refined
}

// verifySignature checks the Slack HMAC-SHA256 request signature.
// Spec: https://api.slack.com/authentication/verifying-requests-from-slack
func (h *EventWebhook) verifySignature(r *http.Request, body []byte) error {
	tsHeader := r.Header.Get("X-Slack-Request-Timestamp")
	sigHeader := r.Header.Get("X-Slack-Signature")

	if tsHeader == "" || sigHeader == "" {
		return fmt.Errorf("missing Slack signature headers")
	}

	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	// Reject requests older than 5 minutes to prevent replay attacks.
	if abs(time.Now().Unix()-ts) > 300 {
		return fmt.Errorf("timestamp too old: %d", ts)
	}

	baseString := fmt.Sprintf("v0:%s:%s", tsHeader, string(body))
	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write([]byte(baseString))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sigHeader)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

// WebhookHandler returns the EventWebhook handler registered on the SlackConnector.
// The connector must have been Init'd with a "signing_secret" config key.
func (s *SlackConnector) WebhookHandler() (http.Handler, error) {
	secret, ok := s.config["signing_secret"].(string)
	if !ok || secret == "" {
		return nil, fmt.Errorf("slack connector: signing_secret required for Events API webhook")
	}
	return NewEventWebhook(secret, s.refinery), nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func init() {
	// Ensure the slack package imports strings (used in webhook routing checks).
	_ = strings.Contains
}
