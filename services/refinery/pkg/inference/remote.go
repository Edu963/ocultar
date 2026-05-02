package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// circuitState represents the three states of the circuit breaker.
type circuitState int

const (
	stateClosed   circuitState = iota // healthy — requests flow through
	stateOpen                          // tripped — requests fail-fast to Tier 1
	stateHalfOpen                      // probing — one test request allowed
)

const (
	failureThreshold = 3                // consecutive failures before opening
	successThreshold = 2                // consecutive successes in HalfOpen before closing
	halfOpenDelay    = 30 * time.Second // wait before probing after Open
	healthInterval   = 10 * time.Second // background health probe cadence
)

// RemoteScanner implements the AIScanner interface by sending HTTP requests
// to the SLM Sidecar microservice. It wraps the outbound calls in a
// three-state circuit breaker so a dead sidecar degrades gracefully to
// Tier 1-only mode instead of blocking every request.
type RemoteScanner struct {
	client     *http.Client
	sidecarURL string
	domain     string

	mu                  sync.Mutex
	state               circuitState
	consecutiveFailures int
	consecutiveSuccesses int
	lastStateChange     time.Time

	stopHealth chan struct{}
}

// NewRemoteScanner creates a scanner that hits the out-of-process SLM engine.
// A background goroutine probes /health every 10 s and moves the circuit
// from Open → HalfOpen → Closed as the sidecar recovers.
func NewRemoteScanner(sidecarURL string) *RemoteScanner {
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8085"
	}
	s := &RemoteScanner{
		client:          &http.Client{Timeout: 30 * time.Second},
		sidecarURL:      sidecarURL,
		state:           stateClosed,
		lastStateChange: time.Now(),
		stopHealth:      make(chan struct{}),
	}
	go s.runHealthLoop()
	return s
}

// ScanForPII forwards the text payload to the SLM sidecar.
// Returns an error (and an empty map) if the circuit is Open so the
// caller can fall through to Tier 1 detection without panicking.
func (s *RemoteScanner) ScanForPII(text string) (map[string][]string, error) {
	if err := s.allow(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	payload := map[string]string{"text": text}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", s.sidecarURL+"/scan", bytes.NewReader(body))
	if err != nil {
		s.recordFailure()
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.recordFailure()
		return nil, fmt.Errorf("SLM sidecar unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.recordFailure()
		return nil, fmt.Errorf("SLM sidecar err: HTTP %d", resp.StatusCode)
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.recordFailure()
		return nil, fmt.Errorf("failed to parse SLM sidecar response: %w", err)
	}

	s.recordSuccess()
	return result, nil
}

// CheckHealth probes /health and updates the circuit state directly.
// Called by the background health loop and exposed for external use.
func (s *RemoteScanner) CheckHealth(host string) {
	resp, err := s.client.Get(s.sidecarURL + "/health")
	healthy := err == nil && resp != nil && resp.StatusCode == 200
	if resp != nil {
		resp.Body.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if healthy && s.state == stateOpen {
		s.transitionTo(stateHalfOpen)
	}
}

func (s *RemoteScanner) IsAvailable() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state != stateOpen
}

func (s *RemoteScanner) SetDomain(domain string) { s.domain = domain }

// Stop shuts down the background health goroutine.
func (s *RemoteScanner) Stop() {
	close(s.stopHealth)
}

// --- Circuit breaker internals ---

// allow checks whether the current circuit state permits a request.
func (s *RemoteScanner) allow() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.state {
	case stateClosed:
		return nil
	case stateOpen:
		// Transition to HalfOpen if enough time has passed.
		if time.Since(s.lastStateChange) >= halfOpenDelay {
			s.transitionTo(stateHalfOpen)
			return nil // allow the probe request
		}
		return fmt.Errorf("SLM circuit open — Tier 2 bypassed, Tier 1 active")
	case stateHalfOpen:
		return nil // allow one probe
	}
	return nil
}

func (s *RemoteScanner) recordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consecutiveSuccesses = 0
	s.consecutiveFailures++

	switch s.state {
	case stateClosed:
		if s.consecutiveFailures >= failureThreshold {
			s.transitionTo(stateOpen)
		}
	case stateHalfOpen:
		s.transitionTo(stateOpen) // single failure in probe → re-open immediately
	}
}

func (s *RemoteScanner) recordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consecutiveFailures = 0
	s.consecutiveSuccesses++

	if s.state == stateHalfOpen && s.consecutiveSuccesses >= successThreshold {
		s.transitionTo(stateClosed)
	}
}

// transitionTo changes state and logs the transition. Caller must hold mu.
func (s *RemoteScanner) transitionTo(next circuitState) {
	names := map[circuitState]string{
		stateClosed:   "CLOSED",
		stateOpen:     "OPEN",
		stateHalfOpen: "HALF-OPEN",
	}
	log.Printf("[CIRCUIT-BREAKER] Tier 2 SLM: %s → %s (failures=%d, successes=%d)",
		names[s.state], names[next], s.consecutiveFailures, s.consecutiveSuccesses)
	s.state = next
	s.lastStateChange = time.Now()
	s.consecutiveFailures = 0
	s.consecutiveSuccesses = 0
}

func (s *RemoteScanner) runHealthLoop() {
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
