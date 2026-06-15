package connector_test

import (
	"testing"

	"github.com/ocultar-dev/ocultar/apps/sombra/pkg/handler"
)

// splitAtTokenBoundary is tested via the exported behaviour of streamRehydrator.
// We drive it through push()/flush() to confirm the boundary logic is correct.

func TestSplitAtTokenBoundary_NoTokens(t *testing.T) {
	safe, hold := handler.SplitAtTokenBoundary("Hello world")
	if safe != "Hello world" || hold != "" {
		t.Errorf("got safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_CompleteToken(t *testing.T) {
	safe, hold := handler.SplitAtTokenBoundary("Hello [PERSON_ab3c12ef] world")
	if safe != "Hello [PERSON_ab3c12ef] world" || hold != "" {
		t.Errorf("complete token should be fully safe: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_IncompleteToken(t *testing.T) {
	safe, hold := handler.SplitAtTokenBoundary("Hello [PERSON_ab3c")
	if safe != "Hello " || hold != "[PERSON_ab3c" {
		t.Errorf("incomplete token should be held: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_OnlyOpenBracket(t *testing.T) {
	safe, hold := handler.SplitAtTokenBoundary("Hello [")
	if safe != "Hello " || hold != "[" {
		t.Errorf("bare '[' should be held: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_MarkdownBracket(t *testing.T) {
	// Lowercase content — not a vault token start.
	safe, hold := handler.SplitAtTokenBoundary("See [link text]")
	if safe != "See [link text]" || hold != "" {
		t.Errorf("markdown bracket should be safe: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_TypeOnly(t *testing.T) {
	// "[PERSON" — type present but no underscore/hash yet.
	safe, hold := handler.SplitAtTokenBoundary("name: [PERSON")
	if safe != "name: " || hold != "[PERSON" {
		t.Errorf("partial type should be held: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_TypePlusSeparator(t *testing.T) {
	safe, hold := handler.SplitAtTokenBoundary("name: [PERSON_")
	if safe != "name: " || hold != "[PERSON_" {
		t.Errorf("type+separator should be held: safe=%q hold=%q", safe, hold)
	}
}

func TestSplitAtTokenBoundary_MultipleTokensLastIncomplete(t *testing.T) {
	// First token is complete, second is in progress.
	safe, hold := handler.SplitAtTokenBoundary("[EMAIL_00fa9b12] and [PHONE_cc84")
	if safe != "[EMAIL_00fa9b12] and " || hold != "[PHONE_cc84" {
		t.Errorf("got safe=%q hold=%q", safe, hold)
	}
}

// --- streamRehydrator integration (vault-free, no actual token lookup needed) ---

// newNoopRehydrator builds a rehydrator backed by nil vault/key.
// Since no actual vault tokens appear in the test strings, RehydrateString
// will pass them through unchanged — confirming the boundary logic alone.
func newNoopRehydrator() *handler.StreamRehydrator {
	return handler.NewStreamRehydrator(nil, nil)
}

func TestStreamRehydrator_SimpleText(t *testing.T) {
	r := newNoopRehydrator()
	out, err := r.Push("Hello world")
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hello world" {
		t.Errorf("got %q", out)
	}
	tail, _ := r.Flush()
	if tail != "" {
		t.Errorf("expected empty flush, got %q", tail)
	}
}

func TestStreamRehydrator_TokenSpanningChunks(t *testing.T) {
	r := newNoopRehydrator()

	// Chunk 1: text + start of a token
	out1, err := r.Push("The person is [PERSON_")
	if err != nil {
		t.Fatal(err)
	}
	if out1 != "The person is " {
		t.Errorf("chunk 1: expected prefix only, got %q", out1)
	}

	// Chunk 2: rest of token + more text
	out2, err := r.Push("ab3c12ef] today")
	if err != nil {
		t.Fatal(err)
	}
	// "[PERSON_ab3c12ef] today" is now in buffer — token is complete so all safe.
	if out2 != "[PERSON_ab3c12ef] today" {
		t.Errorf("chunk 2: expected full token + tail, got %q", out2)
	}

	tail, _ := r.Flush()
	if tail != "" {
		t.Errorf("expected empty flush, got %q", tail)
	}
}

func TestStreamRehydrator_TokenAtEndHeldUntilFlush(t *testing.T) {
	r := newNoopRehydrator()

	out, err := r.Push("prefix [PERSON_ab3c")
	if err != nil {
		t.Fatal(err)
	}
	if out != "prefix " {
		t.Errorf("incomplete token should be held: got %q", out)
	}

	tail, _ := r.Flush()
	if tail != "[PERSON_ab3c" {
		t.Errorf("flush should drain incomplete token: got %q", tail)
	}
}

func TestStreamRehydrator_MarkdownNotHeld(t *testing.T) {
	r := newNoopRehydrator()
	out, _ := r.Push("Click [here] to continue")
	if out != "Click [here] to continue" {
		t.Errorf("markdown brackets should pass through: got %q", out)
	}
}
