package connector_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/router"
)

// mockSSEServer returns an httptest.Server that emits the given text chunks
// as OpenAI-compatible SSE deltas, then sends [DONE].
func mockSSEServer(chunks []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, chunk := range chunks {
			data := fmt.Sprintf(`{"choices":[{"delta":{"content":%q},"finish_reason":null}]}`, chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
}

// mockGeminiSSEServer emits Gemini-format SSE chunks.
func mockGeminiSSEServer(chunks []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, chunk := range chunks {
			data := fmt.Sprintf(`{"candidates":[{"content":{"parts":[{"text":%q}],"role":"model"}}]}`, chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		// Gemini ends without [DONE] — just closes the connection.
	}))
}

// collectDeltas drives SendStream and collects all emitted deltas.
func collectDeltas(t *testing.T, streamer router.Streamer, model string, srv *httptest.Server) []string {
	t.Helper()
	var got []string
	err := streamer.SendStream(
		context.Background(),
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(delta string) error {
			got = append(got, delta)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("SendStream error: %v", err)
	}
	return got
}

// --- OpenAI streaming ---

func TestOpenAIAdapter_SendStream_BasicChunks(t *testing.T) {
	srv := mockSSEServer([]string{"Hello", " world", "!"})
	defer srv.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")
	adapter := router.NewOpenAI("gpt-4o", srv.URL, "OPENAI_API_KEY")

	got := collectDeltas(t, adapter, "gpt-4o", srv)

	if strings.Join(got, "") != "Hello world!" {
		t.Errorf("reassembled: %q", strings.Join(got, ""))
	}
}

func TestOpenAIAdapter_SendStream_EmptyDeltas_Skipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mix of empty-content chunks (role-only) and real content.
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\"},\"finish_reason\":null}]}\n\n")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"},\"finish_reason\":null}]}\n\n")
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")
	adapter := router.NewOpenAI("gpt-4o", srv.URL, "OPENAI_API_KEY")
	got := collectDeltas(t, adapter, "gpt-4o", srv)

	if len(got) != 1 || got[0] != "Hi" {
		t.Errorf("expected [Hi], got %v", got)
	}
}

func TestOpenAIAdapter_SendStream_HTTP400_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	t.Setenv("OPENAI_API_KEY", "bad-key")
	adapter := router.NewOpenAI("gpt-4o", srv.URL, "OPENAI_API_KEY")

	err := adapter.SendStream(context.Background(),
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(string) error { return nil },
	)
	if err == nil {
		t.Fatal("expected error for HTTP 401")
	}
}

// --- Gemini streaming ---

func TestGeminiAdapter_SendStream_BasicChunks(t *testing.T) {
	srv := mockGeminiSSEServer([]string{"Bonjour", " le monde"})
	defer srv.Close()

	t.Setenv("GEMINI_API_KEY", "test-key")
	// Point endpoint at mock server; apiModelID doesn't matter for mock.
	adapter := router.NewGemini("gemini-flash-latest", srv.URL, "GEMINI_API_KEY", "gemini-mock")
	got := collectDeltas(t, adapter, "gemini-flash-latest", srv)

	if strings.Join(got, "") != "Bonjour le monde" {
		t.Errorf("reassembled: %q", strings.Join(got, ""))
	}
}

func TestGeminiAdapter_SendStream_NoMissingAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	adapter := router.NewGemini("gemini-flash-latest", "http://localhost", "GEMINI_API_KEY", "gemini-mock")

	err := adapter.SendStream(context.Background(),
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(string) error { return nil },
	)
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Fatalf("expected missing-key error, got: %v", err)
	}
}

// --- Local streaming ---

func TestLocalAdapter_SendStream_BasicChunks(t *testing.T) {
	srv := mockSSEServer([]string{"ciao", " mondo"})
	defer srv.Close()

	adapter := router.NewLocal("local-slm", srv.URL)
	got := collectDeltas(t, adapter, "local-slm", srv)

	if strings.Join(got, "") != "ciao mondo" {
		t.Errorf("reassembled: %q", strings.Join(got, ""))
	}
}

// --- Router.SendStream fallback ---

// nonStreamingAdapter is a ModelAdapter that does NOT implement Streamer.
// Router.SendStream should fall back to Send() and emit one delta.
type nonStreamingAdapter struct{ response string }

func (a *nonStreamingAdapter) Name() string     { return "non-streamer" }
func (a *nonStreamingAdapter) Endpoint() string { return "http://127.0.0.1" }
func (a *nonStreamingAdapter) HealthCheck(_ context.Context) error { return nil }
func (a *nonStreamingAdapter) Send(_ context.Context, _ []router.Message, _ router.ModelOpts) (string, error) {
	return a.response, nil
}

func TestRouter_SendStream_FallsBackToSend_WhenStreamerNotImplemented(t *testing.T) {
	r := router.New("non-streamer", []string{"127.0.0.1"})
	r.Register(&nonStreamingAdapter{response: "buffered response"})

	var got []string
	err := r.SendStream(context.Background(), "non-streamer",
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(delta string) error {
			got = append(got, delta)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "buffered response" {
		t.Errorf("expected single buffered delta, got %v", got)
	}
}

func TestRouter_SendStream_UnknownModel_ReturnsError(t *testing.T) {
	r := router.New("default", nil)
	err := r.SendStream(context.Background(), "no-such-model",
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(string) error { return nil },
	)
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
}

func TestRouter_SendStream_BlocksUnapprovedDomain(t *testing.T) {
	srv := mockSSEServer([]string{"hello"})
	defer srv.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")
	r := router.New("gpt-4o", []string{"approved.example.com"})
	r.Register(router.NewOpenAI("gpt-4o", srv.URL, "OPENAI_API_KEY"))

	err := r.SendStream(context.Background(), "gpt-4o",
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(string) error { return nil },
	)
	if err == nil || !strings.Contains(err.Error(), "Zero-Egress Block") {
		t.Fatalf("expected zero-egress block error, got: %v", err)
	}
}

func TestRouter_SendStream_CallbackError_AbortsStream(t *testing.T) {
	srv := mockSSEServer([]string{"a", "b", "c"})
	defer srv.Close()

	t.Setenv("OPENAI_API_KEY", "test-key")
	adapter := router.NewOpenAI("gpt-4o", srv.URL, "OPENAI_API_KEY")

	count := 0
	err := adapter.SendStream(context.Background(),
		[]router.Message{{Role: "user", Content: "hi"}},
		router.ModelOpts{},
		func(delta string) error {
			count++
			if count == 2 {
				return fmt.Errorf("client disconnected")
			}
			return nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "client disconnected") {
		t.Fatalf("expected abort error, got: %v", err)
	}
	if count != 2 {
		t.Errorf("expected stream aborted after 2 deltas, got %d", count)
	}
}
