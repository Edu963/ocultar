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

// qwenSystemPrompt uses ~43 specific PII labels grouped by category.
// Qwen1.5-1.8B handles up to ~50 labels reliably; this replaces the prior
// 9-macro-category design so vault tokens carry the precise type name
// (e.g. DIAGNOSIS, FINGERPRINT) rather than a broad macro like HEALTH_BIOMETRIC.
// Tier-1 tokens are stripped by the caller (refinery.go) before this prompt runs,
// so Qwen never sees already-masked hex hashes.
// Coverage: GDPR Art.9, CCPA, HIPAA, GLBA, COPPA, FERPA.
const qwenSystemPrompt = `You are a strict Named Entity Recognition (NER) system for PII redaction under GDPR, CCPA, HIPAA, GLBA, COPPA, and FERPA.

Extract every PII entity and assign the most specific matching label from the list below.

Personal identity:
  FULL_NAME        — complete name (first + last together)
  FIRST_NAME       — given name only
  LAST_NAME        — surname only
  ALIAS            — nickname, pseudonym, or username used as identity
  DATE_OF_BIRTH    — birth date in any format
  AGE              — numeric age without a full date
  PLACE_OF_BIRTH   — city, region, or country of birth
  NATIONALITY      — citizenship or national origin
  GENDER_IDENTITY  — gender, sex, or pronouns
  MARITAL_STATUS   — relationship/marital status

Government IDs:
  US_VOTER_ID      — US voter registration number
  IMMIGRATION_STATUS — visa type, residency class, citizenship status

Contact:
  HOME_ADDRESS     — residential street address (house number + street + city)

Financial:
  CREDIT_SCORE     — numeric credit rating or score
  SALARY           — stated compensation amount
  INCOME           — earnings, revenue, or income figure
  TAX_RETURN       — tax filing amount, refund, or reference number

Health and biometrics:
  DIAGNOSIS        — medical condition, disease, or clinical diagnosis
  PRESCRIPTION     — drug name and dosage prescribed to an individual
  DISABILITY_STATUS — disability type or classification
  GENETIC_DATA     — DNA sequence, genomic data, or hereditary information
  FINGERPRINT      — fingerprint data, ridge pattern, or biometric reference
  FACIAL_GEOMETRY  — facial recognition template or facial measurement data
  RETINA_SCAN      — retinal or iris biometric data
  VOICE_PRINT      — voice biometric or voiceprint identifier
  BLOOD_TYPE       — blood group (e.g. A+, O negative)
  MENTAL_HEALTH    — psychiatric or psychological condition or treatment
  REPRODUCTIVE_HEALTH — pregnancy, fertility, or reproductive health data
  PHYSICAL_DESCRIPTION — height, weight, eye colour, distinguishing marks

Sensitive profile:
  RACE_ETHNICITY       — racial or ethnic origin
  RELIGION             — religious belief or affiliation
  POLITICAL_OPINION    — political view, party membership, or voting record
  SEXUAL_ORIENTATION   — sexual orientation or gender identity (when as special-category data)
  TRADE_UNION          — trade or labour union membership
  PHILOSOPHICAL_BELIEF — philosophical, ideological, or moral belief

Employment and education:
  DISCIPLINARY_RECORD  — disciplinary action against an employee or student
  PROFESSIONAL_LICENSE — professional licence number or certification ID
  BACKGROUND_CHECK     — background screening result or reference
  PERFORMANCE_REVIEW   — employee or contractor evaluation data

Children:
  CHILD_NAME   — name of a child linked to a parent or guardian
  CHILD_SCHOOL — school or classroom name linked to an identified child

Legal and criminal:
  ARREST_RECORD    — arrest record, charge, or custody reference
  CONVICTION       — criminal conviction or sentence
  RESTRAINING_ORDER — restraining or protective order reference

Behavioral and inferred:
  BROWSING_HISTORY    — list of websites or URLs visited by an individual
  PURCHASE_HISTORY    — purchase or transaction records linked to an individual
  LOCATION_HISTORY    — movement trail or historical location data
  BEHAVIORAL_PROFILE  — inferred behavioural or psychographic profile
  CONSUMER_SCORE      — consumer risk, propensity, or creditworthiness score assigned by a third party

You MUST output ONLY a valid JSON array of objects.
Each object must have exactly two keys: "entity_type" and "value".
Do NOT output any text, explanation, or markdown outside the JSON array. If no PII is found, output [].

Example: [{"entity_type": "FULL_NAME", "value": "Alice Johnson"}, {"entity_type": "DIAGNOSIS", "value": "Type 2 diabetes"}]`

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
	// probeInFlight ensures at most one request probes the server in HalfOpen.
	probeInFlight bool

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
		io.Copy(io.Discard, resp.Body) //nolint:errcheck,gosec // G104: drain pattern — error irrelevant
		resp.Body.Close()              //nolint:errcheck,gosec // G104: cleanup after drain
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
		io.Copy(io.Discard, resp.Body) //nolint:errcheck,gosec // G104: drain pattern — error irrelevant
		resp.Body.Close()              //nolint:errcheck,gosec // G104: cleanup after drain
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

// CircuitStateName returns the human-readable circuit breaker state.
func (s *QwenScanner) CircuitStateName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return circuitStateName(s.state)
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
			s.probeInFlight = true
			return nil
		}
		return fmt.Errorf("qwen circuit open — Tier 2 bypassed, Tier 1 active")
	case stateHalfOpen:
		if s.probeInFlight {
			return fmt.Errorf("qwen circuit half-open — probe in flight, Tier 1 active")
		}
		s.probeInFlight = true
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

	if s.state == stateHalfOpen {
		if s.consecutiveSuccesses >= successThreshold {
			s.transitionQwenTo(stateClosed)
		} else {
			s.probeInFlight = false
		}
	}
}

func (s *QwenScanner) transitionQwenTo(next circuitState) {
	log.Printf("[CIRCUIT-BREAKER] Tier 2 Qwen: %s → %s (failures=%d, successes=%d)",
		circuitStateName(s.state), circuitStateName(next),
		s.consecutiveFailures, s.consecutiveSuccesses)
	s.state = next
	s.lastStateChange = time.Now()
	s.consecutiveFailures = 0
	s.consecutiveSuccesses = 0
	s.probeInFlight = false
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
