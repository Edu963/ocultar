package inference

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRemoteScanner_ScanForPII(t *testing.T) {
	tests := []struct {
		name       string
		serverResp interface{}
		statusCode int
		wantKeys   []string
		wantErr    bool
	}{
		{
			name:       "returns detected entities",
			serverResp: map[string][]string{"PERSON": {"Alice Johnson"}, "EMAIL": {"alice@example.com"}},
			statusCode: 200,
			wantKeys:   []string{"PERSON", "EMAIL"},
		},
		{
			name:       "returns empty map when no PII",
			serverResp: map[string][]string{},
			statusCode: 200,
			wantKeys:   []string{},
		},
		{
			name:       "server error triggers circuit failure",
			serverResp: "internal error",
			statusCode: 500,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/health" {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(tc.statusCode)
				json.NewEncoder(w).Encode(tc.serverResp)
			}))
			defer srv.Close()

			scanner := NewRemoteScanner(srv.URL)
			defer scanner.Stop()

			result, err := scanner.ScanForPII("Alice Johnson email alice@example.com")
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, key := range tc.wantKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("missing key %q in result: %v", key, result)
				}
			}
		})
	}
}

func TestRemoteScanner_IsAvailable_InitiallyTrue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	if !scanner.IsAvailable() {
		t.Error("new scanner should be available (circuit closed)")
	}
	if scanner.CircuitStateName() != "closed" {
		t.Errorf("expected closed, got %s", scanner.CircuitStateName())
	}
}

func TestRemoteScanner_CircuitOpensAfterConsecutiveFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	for i := 0; i < failureThreshold; i++ {
		scanner.ScanForPII("test")
	}

	if scanner.IsAvailable() {
		t.Error("circuit should be open after consecutive failures")
	}
	if scanner.CircuitStateName() != "open" {
		t.Errorf("expected open, got %s", scanner.CircuitStateName())
	}
}

func TestRemoteScanner_CircuitRejectsWhenOpen(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	// Trip the circuit
	for i := 0; i < failureThreshold; i++ {
		scanner.ScanForPII("test")
	}

	// Next call should fail fast without hitting the server
	_, err := scanner.ScanForPII("should fail fast")
	if err == nil {
		t.Error("expected error when circuit is open, got nil")
	}
}

func TestRemoteScanner_HalfOpen_OnlyOneProbeAtATime(t *testing.T) {
	var inFlight atomic.Int32
	var maxConcurrent atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Track concurrent requests reaching the server
		cur := inFlight.Add(1)
		for {
			if old := maxConcurrent.Load(); cur > old {
				if maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			} else {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		inFlight.Add(-1)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string][]string{})
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	// Force circuit to half-open by manipulating internal state
	scanner.mu.Lock()
	scanner.state = stateHalfOpen
	scanner.probeInFlight = false
	scanner.mu.Unlock()

	// Launch multiple goroutines simultaneously in HalfOpen
	const goroutines = 10
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner.ScanForPII("concurrent probe test")
		}()
	}
	wg.Wait()

	// At most 1 request should have reached the server concurrently
	if max := maxConcurrent.Load(); max > 1 {
		t.Errorf("expected at most 1 concurrent probe in half-open, got %d", max)
	}
}

func TestRemoteScanner_HalfOpen_TransitionsToClosedAfterSuccesses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string][]string{})
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	// Force to HalfOpen
	scanner.mu.Lock()
	scanner.state = stateHalfOpen
	scanner.consecutiveSuccesses = 0
	scanner.probeInFlight = false
	scanner.mu.Unlock()

	// Two successful probes should close the circuit
	for i := 0; i < successThreshold; i++ {
		_, err := scanner.ScanForPII("probe")
		if err != nil {
			t.Fatalf("probe %d failed: %v", i+1, err)
		}
	}

	if !scanner.IsAvailable() {
		t.Error("circuit should be closed after successful probes")
	}
	if scanner.CircuitStateName() != "closed" {
		t.Errorf("expected closed, got %s", scanner.CircuitStateName())
	}
}

func TestRemoteScanner_CircuitStateName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	scanner := NewRemoteScanner(srv.URL)
	defer scanner.Stop()

	states := []struct {
		set  circuitState
		want string
	}{
		{stateClosed, "closed"},
		{stateOpen, "open"},
		{stateHalfOpen, "half-open"},
	}
	for _, tc := range states {
		scanner.mu.Lock()
		scanner.state = tc.set
		scanner.mu.Unlock()
		if got := scanner.CircuitStateName(); got != tc.want {
			t.Errorf("state %d: want %q, got %q", tc.set, tc.want, got)
		}
	}
}
