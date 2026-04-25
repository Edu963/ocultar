package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Edu963/ocultar/apps/sombra/pkg/router"
	"github.com/Edu963/ocultar/pkg/proxy"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
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

		go g.processSlackMessageAsynchronously(payload)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (g *Gateway) processSlackMessageAsynchronously(payload SlackEventPayload) {
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
		g.sendSlackMessage(slackToken, payload.Event.Channel, "⚠️ *Security Block*: Ocultar blocked this message due to a processing failure.")
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
	
	aiRespString, err := g.router.Send(context.Background(), modelName, msg, router.ModelOpts{})
	if err != nil {
		log.Printf("[ERROR] AI Routing failed: %v", err)
		g.sendSlackMessage(slackToken, payload.Event.Channel, "⚠️ *Gateway Error*: Upstream AI provider failed.")
		return
	}

	// 4. Security Re-Hydration (Inbound from LLM to Slack)
	rehydratedResponse, err := proxy.RehydrateString(g.vault, g.masterKey, aiRespString)
	if err != nil {
		log.Printf("[ERROR] Re-hydration failed: %v", err)
		g.sendSlackMessage(slackToken, payload.Event.Channel, "⚠️ *Security Block*: Re-hydration error. Data cannot be securely returned.")
		return
	}

	// 5. Send back to Slack
	g.sendSlackMessage(slackToken, payload.Event.Channel, rehydratedResponse)
	
	if g.auditor != nil {
		g.auditor.Log(actor, "SLACK_QUERY", modelName, "SUCCESS", "End-to-End Slack response delivered safely")
	}
}

func (g *Gateway) sendSlackMessage(token, channel, text string) {
	slackAPI := "https://slack.com/api/chat.postMessage"
	
	reqBody := map[string]string{
		"channel": channel,
		"text":    text,
	}
	
	bodyData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", slackAPI, bytes.NewReader(bodyData))
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
