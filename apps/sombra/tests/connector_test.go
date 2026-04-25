package connector_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/apps/sombra/pkg/connector"
)

func TestFileConnector_Fetch(t *testing.T) {
	policy := connector.DataPolicy{
		MaxBodyBytes: 1024,
	}
	fc := connector.NewFileConnector("file", policy)

	t.Run("In-memory raw body CSV", func(t *testing.T) {
		req := connector.FetchRequest{
			RawBody:     []byte("name,email\nJohn Doe,john@test.com"),
			ContentType: "text/csv",
		}
		resp, err := fc.Fetch(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ContentType != "text/plain" {
			t.Errorf("expected plain text, got %q", resp.ContentType)
		}

		bodyStr := string(resp.Body)
		if !strings.Contains(bodyStr, "Record 1") || !strings.Contains(bodyStr, "john@test.com") {
			t.Errorf("failed to normalise CSV: %s", bodyStr)
		}
	})

	t.Run("Size limit exceeded", func(t *testing.T) {
		req := connector.FetchRequest{
			RawBody: make([]byte, 2048),
		}
		_, err := fc.Fetch(context.Background(), req)
		if err == nil || !strings.Contains(err.Error(), "exceeds policy limit") {
			t.Errorf("expected size limit error, got %v", err)
		}
	})
}

func TestAPIConnector_Fetch(t *testing.T) {
	// Create a mock API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-secret-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/v1/accounts/123" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"account_id": "123", "balance": 1000}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	cfg := connector.APIConnectorConfig{
		Name:     "mock-bank",
		Endpoint: ts.URL + "/v1/accounts",
		AuthType: "bearer",
		KeyEnv:   "MOCK_BANK_KEY",
		Policy:   connector.DataPolicy{},
	}
	ac := connector.NewAPIConnector(cfg)

	t.Run("Missing auth token", func(t *testing.T) {
		os.Unsetenv("MOCK_BANK_KEY")
		_, err := ac.Fetch(context.Background(), connector.FetchRequest{SourceID: "123"})
		if err == nil {
			t.Error("expected auth error")
		}
	})

	t.Run("Successful authenticated fetch", func(t *testing.T) {
		os.Setenv("MOCK_BANK_KEY", "test-secret-token")
		defer os.Unsetenv("MOCK_BANK_KEY")

		resp, err := ac.Fetch(context.Background(), connector.FetchRequest{SourceID: "123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		bodyStr := string(resp.Body)
		if !strings.Contains(bodyStr, "1000") {
			t.Errorf("expected balance 1000 in response, got %s", bodyStr)
		}
	})
}
