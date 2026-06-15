package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/router"
	"github.com/ocultar-dev/ocultar/pkg/proxy"
)

type SlackEventPayload struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge,omitempty"` // For URL verification
	Event     struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text"`
		BotID   string `json:"bot_id,omitempty"`
	} `json:"event"`
}

// HandleSlackEvent handles incoming webhook requests from the Slack Events API.
// It acts as the "One Killer Connector", turning Sombra into a safe AI chatbot for Teams.
func (g *Gateway) HandleSlackEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body first — needed for HMAC verification before any parsing.
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB cap
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Verify Slack request signature before processing any payload.
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if signingSecret == "" {
		log.Printf("[ERROR] SLACK_SIGNING_SECRET not configured — rejecting Slack event")
		http.Error(w, "slack signing secret not configured", http.StatusInternalServerError)
		return
	}
	if !verifySlackSignature(signingSecret, r.Header.Get("X-Slack-Request-Timestamp"), r.Header.Get("X-Slack-Signature"), body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload SlackEventPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// 1. URL Verification Challenge
	if payload.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(payload.Challenge))
		return
	}

	// Only process valid messages from users (ignore bot loops)
	if payload.Type == "event_callback" && payload.Event.Type == "message" && payload.Event.BotID == "" {
		// Acknowledge Slack event immediately (Slack requires 200 OK within 3s)
		w.WriteHeader(http.StatusOK)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			g.processSlackMessageAsynchronously(ctx, payload)
		}()
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// verifySlackSignature validates the X-Slack-Signature header per the Slack signing
// secret protocol: HMAC-SHA256("v0:{timestamp}:{body}", signingSecret).
// Returns false if the timestamp differs from now by more than 5 minutes in
// either direction (replay protection + future-timestamp rejection).
func verifySlackSignature(signingSecret, timestamp, sigHeader string, body []byte) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	diff := time.Now().Unix() - ts
	if diff < 0 {
		diff = -diff
	}
	if diff > 300 {
		return false
	}
	sigBase := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(sigBase))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sigHeader), []byte(expected))
}

func (g *Gateway) processSlackMessageAsynchronously(ctx context.Context, payload SlackEventPayload) {
	actor := fmt.Sprintf("slack-user-%s", payload.Event.User)
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackToken == "" {
		log.Printf("[ERROR] SLACK_BOT_TOKEN not configured.")
		return
	}

	// 2. Fail-Closed Redaction (Outbound from Slack to LLM)
	refinedPrompt, err := g.eng.RefineString(payload.Event.Text, actor, nil)
	if err != nil {
		log.Printf("[ERROR] Refinery failed to process Slack message: %v", err)
		g.sendSlackMessage(ctx, slackToken, payload.Event.Channel, "⚠️ *Security Block*: Ocultar blocked this message due to a processing failure.")
		return
	}

	// 3. Route to LLM
	modelName := os.Getenv("SLACK_LLM_MODEL")
	if modelName == "" {
		modelName = "gpt-4o"
	}

	msg := []router.Message{
		{Role: "system", Content: "You are a helpful AI assistant connected via Ocultar Gateway. When answering questions, just answer naturally."},
		{Role: "user", Content: refinedPrompt},
	}

	aiRespString, err := g.router.Send(ctx, modelName, msg, router.ModelOpts{})
	if err != nil {
		log.Printf("[ERROR] AI Routing failed: %v", err)
		g.sendSlackMessage(ctx, slackToken, payload.Event.Channel, "⚠️ *Gateway Error*: Upstream AI provider failed.")
		return
	}

	// 4. Security Re-Hydration (Inbound from LLM to Slack)
	rehydratedResponse, err := proxy.RehydrateString(g.vault, g.masterKey, aiRespString)
	if err != nil {
		log.Printf("[ERROR] Re-hydration failed: %v", err)
		g.sendSlackMessage(ctx, slackToken, payload.Event.Channel, "⚠️ *Security Block*: Re-hydration error. Data cannot be securely returned.")
		return
	}

	// 5. Send back to Slack
	g.sendSlackMessage(ctx, slackToken, payload.Event.Channel, rehydratedResponse)

	if g.auditor != nil {
		g.auditor.Log(actor, "SLACK_QUERY", modelName, "SUCCESS", "End-to-End Slack response delivered safely")
	}
}

func (g *Gateway) sendSlackMessage(ctx context.Context, token, channel, text string) {
	slackAPI := "https://slack.com/api/chat.postMessage"

	reqBody := map[string]string{
		"channel": channel,
		"text":    text,
	}

	bodyData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal Slack message body: %v", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackAPI, bytes.NewReader(bodyData))
	if err != nil {
		log.Printf("[ERROR] Failed to build Slack API request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to send slack message: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Slack API returned %d", resp.StatusCode)
	}
}
