package inference

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQwenScanner_ScanForPII(t *testing.T) {
	tests := []struct {
		name       string
		serverResp string
		statusCode int
		wantKeys   []string
		wantErr    bool
	}{
		{
			name: "extracts full name and diagnosis",
			serverResp: `{"choices":[{"message":{"content":"[{\"entity_type\":\"FULL_NAME\",\"value\":\"Alice Johnson\"},{\"entity_type\":\"DIAGNOSIS\",\"value\":\"Type 2 diabetes\"}]"}}]}`,
			statusCode: 200,
			wantKeys:   []string{"FULL_NAME", "DIAGNOSIS"},
		},
		{
			name:       "empty array when no PII",
			serverResp: `{"choices":[{"message":{"content":"[]"}}]}`,
			statusCode: 200,
			wantKeys:   []string{},
		},
		{
			name:       "tolerates markdown fences around JSON",
			serverResp: `{"choices":[{"message":{"content":"` + "```" + `json\n[{\"entity_type\":\"FINGERPRINT\",\"value\":\"right index\"}]\n` + "```" + `"}}]}`,
			statusCode: 200,
			wantKeys:   []string{"FINGERPRINT"},
		},
		{
			name:       "server error opens circuit",
			serverResp: `internal error`,
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
				if r.URL.Path != "/v1/chat/completions" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var req qwenChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if len(req.Messages) != 2 || req.Messages[0].Role != "system" {
					t.Errorf("expected system+user messages, got %d messages", len(req.Messages))
				}
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.serverResp))
			}))
			defer srv.Close()

			scanner := NewQwenScanner(srv.URL)
			defer scanner.Stop()

			result, err := scanner.ScanForPII("Please process: Alice Johnson has Type 2 diabetes.")

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

func TestQwenScanner_IsAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	scanner := NewQwenScanner(srv.URL)
	defer scanner.Stop()

	if !scanner.IsAvailable() {
		t.Error("new scanner should be available (circuit closed)")
	}
}

func TestQwenScanner_CircuitOpensAfterFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`error`))
	}))
	defer srv.Close()

	scanner := NewQwenScanner(srv.URL)
	defer scanner.Stop()

	for i := 0; i < failureThreshold; i++ {
		scanner.ScanForPII("test")
	}

	if scanner.IsAvailable() {
		t.Error("circuit should be open after consecutive failures")
	}
}
