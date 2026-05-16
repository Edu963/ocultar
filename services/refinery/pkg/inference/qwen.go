package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// qwenSystemPrompt uses 9 macro-categories instead of granular labels.
// Qwen1.5-1.8B collapses under high-cardinality classification (100+ labels);
// macro-categories keep accuracy high while covering GDPR, CCPA, HIPAA, GLBA,
// COPPA, and FERPA. The refinery maps category names to vault token labels via
// strings.ToUpper, so "HEALTH_BIOMETRIC" becomes [HEALTH_BIOMETRIC_xxxx].
// Tier-1 tokens are stripped by the caller (refinery.go) before this prompt runs,
// so Qwen never sees already-masked text.
const qwenSystemPrompt = `You are a strict Named Entity Recognition (NER) system tailored for comprehensive PII redaction across EU and US regulations (GDPR, CCPA, HIPAA, GLBA, COPPA, FERPA).
Extract all PII entities and classify them into the following macro-categories:

- IDENTITY: Full names, aliases, date/place of birth, age, nationality, marital status, signatures, gender identity, photographs.
- GOV_ID: Social Security Numbers (SSN), passports, driver's licenses, national/state IDs (e.g., PESEL, NIR, NIF, Codice Fiscale, BSN), tax IDs, immigration/visa status, voter/military IDs, criminal records.
- CONTACT: Home/mailing addresses, email addresses, phone/fax numbers.
- DIGITAL_NETWORK: IP addresses, precise geolocation/GPS, MAC addresses, device identifiers (IMEI/UDID), usernames, passwords, crypto wallet addresses, browser/cookie identifiers.
- FINANCIAL: Bank account numbers (IBAN/routing), credit/debit cards, CVV, insurance policy numbers, salary/income data, loan/mortgage details, credit scores, transaction history.
- HEALTH_BIOMETRIC: Medical record numbers (MRN), diagnoses, prescriptions, disability status, genetic data, fingerprints, facial geometry/retina scans, voice prints, physical descriptions, health insurance IDs.
- SENSITIVE_PROFILE: Race/ethnicity, religious/philosophical beliefs, political opinions, sexual orientation, trade union membership, criminal records, drug/substance use records.
- EDU_EMP: Student IDs, grades/transcripts, disciplinary records, employee badges, background checks, performance reviews, professional licenses, salary compensation.
- CHILDREN_DATA: Data of individuals under 16 (EU/GDPR Art.8) or under 13 (US/COPPA): child's name linked to parent, child's school/class, child's geolocation, persistent child device IDs (IDFA/GAID), child photos/audio/video.

You MUST output ONLY a valid JSON array of objects.
Each object must have exactly two keys: "entity_type" and "value".
Do NOT output any conversational text, explanations, or markdown formatting blocks. If no entities are found, output an empty array [].

Example output: [{"entity_type": "IDENTITY", "value": "John Doe"}, {"entity_type": "DIGITAL_NETWORK", "value": "192.168.1.1"}]`

type qwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type qwenChatRequest struct {
	Messages    []qwenMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type qwenChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type qwenEntity struct {
	EntityType string `json:"entity_type"`
	Value      string `json:"value"`
}

// QwenScanner implements AIScanner using an OpenAI-compatible llama.cpp server
// (e.g. ghcr.io/ggml-org/llama.cpp:server running Qwen1.5-1.8B-Chat-Q4_K_M).
// It wraps calls in the same three-state circuit breaker as RemoteScanner so a
// dead container degrades gracefully to Tier 1-only mode.
type QwenScanner struct {
	client     *http.Client
	sidecarURL string
	domain     string

	mu                   sync.Mutex
	state                circuitState
	consecutiveFailures  int
	consecutiveSuccesses int
	lastStateChange      time.Time

	stopHealth chan struct{}
}

// NewQwenScanner creates a scanner that calls the llama.cpp /v1/chat/completions
// endpoint. A background goroutine probes /health every 10 s to recover from
// transient failures.
func NewQwenScanner(sidecarURL string) *QwenScanner {
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8080"
	}
	transport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	s := &QwenScanner{
		// 5-second hard timeout: llama.cpp must respond within this window or
		// the circuit breaker records a failure and Tier 1 takes over.
		client:          &http.Client{Timeout: 5 * time.Second, Transport: transport},
		sidecarURL:      sidecarURL,
		state:           stateClosed,
		lastStateChange: time.Now(),
		stopHealth:      make(chan struct{}),
	}
	go s.runHealthLoop()
	return s
}

// ScanForPII sends text to the llama.cpp chat completions endpoint and parses
// the returned JSON entity array into a map keyed by entity type.
func (s *QwenScanner) ScanForPII(text string) (map[string][]string, error) {
	if err := s.allowQwen(); err != nil {
		return nil, err
	}

	result := map[string][]string{}

	reqBody := qwenChatRequest{
		Messages: []qwenMessage{
			{Role: "system", Content: qwenSystemPrompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.1,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.sidecarURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: llama.cpp unreachable: %w", err)
	}
	defer func() {
		// Drain body to allow TCP connection reuse.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: HTTP %d — %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var chatResp qwenChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		s.recordQwenSuccess()
		return result, nil
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Extract JSON array even if the model wrapped it in markdown fences.
	if start := strings.Index(content, "["); start != -1 {
		if end := strings.LastIndex(content, "]"); end > start {
			content = content[start : end+1]
		}
	}

	var entities []qwenEntity
	if err := json.Unmarshal([]byte(content), &entities); err != nil {
		log.Printf("[WARN] qwen: JSON parse failed: %v. Raw: %s", err, content)
		s.recordQwenFailure()
		return result, fmt.Errorf("qwen: parse entity array: %w", err)
	}

	for _, e := range entities {
		val := strings.TrimSpace(e.Value)
		if len(val) > 2 {
			result[e.EntityType] = append(result[e.EntityType], val)
		}
	}

	s.recordQwenSuccess()
	return result, nil
}

// CheckHealth probes /health and moves the circuit from Open → HalfOpen if
// the server is reachable. The host parameter is ignored; sidecarURL is used.
func (s *QwenScanner) CheckHealth(_ string) {
	resp, err := s.client.Get(s.sidecarURL + "/health")
	healthy := err == nil && resp != nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if healthy && s.state == stateOpen {
		s.transitionQwenTo(stateHalfOpen)
	}
}

func (s *QwenScanner) IsAvailable() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state != stateOpen
}

func (s *QwenScanner) SetDomain(domain string) { s.domain = domain }

// Stop shuts down the background health goroutine.
func (s *QwenScanner) Stop() { close(s.stopHealth) }

// --- Circuit breaker internals (mirrors RemoteScanner) ---

func (s *QwenScanner) allowQwen() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.state {
	case stateClosed:
		return nil
	case stateOpen:
		if time.Since(s.lastStateChange) >= halfOpenDelay {
			s.transitionQwenTo(stateHalfOpen)
			return nil
		}
		return fmt.Errorf("qwen circuit open — Tier 2 bypassed, Tier 1 active")
	case stateHalfOpen:
		return nil
	}
	return nil
}

func (s *QwenScanner) recordQwenFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consecutiveSuccesses = 0
	s.consecutiveFailures++

	switch s.state {
	case stateClosed:
		if s.consecutiveFailures >= failureThreshold {
			s.transitionQwenTo(stateOpen)
		}
	case stateHalfOpen:
		s.transitionQwenTo(stateOpen)
	}
}

func (s *QwenScanner) recordQwenSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consecutiveFailures = 0
	s.consecutiveSuccesses++

	if s.state == stateHalfOpen && s.consecutiveSuccesses >= successThreshold {
		s.transitionQwenTo(stateClosed)
	}
}

func (s *QwenScanner) transitionQwenTo(next circuitState) {
	names := map[circuitState]string{
		stateClosed:   "CLOSED",
		stateOpen:     "OPEN",
		stateHalfOpen: "HALF-OPEN",
	}
	log.Printf("[CIRCUIT-BREAKER] Tier 2 Qwen: %s → %s (failures=%d, successes=%d)",
		names[s.state], names[next], s.consecutiveFailures, s.consecutiveSuccesses)
	s.state = next
	s.lastStateChange = time.Now()
	s.consecutiveFailures = 0
	s.consecutiveSuccesses = 0
}

func (s *QwenScanner) runHealthLoop() {
	ticker := time.NewTicker(healthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			st := s.state
			s.mu.Unlock()
			if st != stateClosed {
				s.CheckHealth("")
			}
		case <-s.stopHealth:
			return
		}
	}
}
