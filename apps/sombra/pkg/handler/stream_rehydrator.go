package handler

import (
	"regexp"
	"strings"

	"github.com/ocultar-dev/ocultar/pkg/proxy"
	"github.com/ocultar-dev/ocultar/vault"
)

// Vault tokens have the form [TYPE_8hexchars], e.g. [PERSON_ab3c12ef].
// Both regexes below are anchored to the start of the string for efficient use
// against the tail of the accumulated buffer.
var (
	// completeToken matches a fully-formed vault token at the start of a string.
	completeToken = regexp.MustCompile(`^\[[A-Z_]+_[0-9a-f]{8}\]`)

	// incompleteToken matches text that looks like the opening of a vault token
	// but is not yet complete. Used to detect where to hold the buffer.
	// Matches: "[", "[PERSON", "[PERSON_", "[PERSON_ab12"
	// Does NOT match: "[markdown text]", "[1234]" (lowercase / digit after bracket)
	incompleteToken = regexp.MustCompile(`^\[[A-Z_]*(_[0-9a-f]{0,8})?$`)
)

// SplitAtTokenBoundary splits s into:
//   - safe: text that contains no incomplete vault token — safe to rehydrate and emit
//   - hold: text at the end that might be the opening of a vault token spanning the
//     next chunk — must be held until more data arrives
//
// Exported for testing.
func SplitAtTokenBoundary(s string) (safe, hold string) {
	last := strings.LastIndex(s, "[")
	if last == -1 {
		return s, ""
	}
	tail := s[last:]

	// If the last '[' starts a COMPLETE token, everything is safe.
	if completeToken.MatchString(tail) {
		return s, ""
	}

	// If the tail looks like the opening of a vault token, hold it.
	if incompleteToken.MatchString(tail) {
		return s[:last], tail
	}

	// '[' is present but doesn't match vault token syntax (e.g. markdown).
	return s, ""
}

// StreamRehydrator accumulates upstream SSE text deltas and emits rehydrated
// text that is safe to forward to the client without exposing vault token syntax.
//
// Problem: a vault token like [PERSON_ab3c12ef] may arrive split across multiple
// upstream chunks ("[PERSON_" in one chunk, "ab3c12ef]" in the next). Emitting
// either half would expose raw vault syntax to the client.
//
// Solution: after each Push(), the largest safe prefix — everything up to the
// last '[' that could be the start of an incomplete token — is rehydrated and
// returned. The remainder is buffered until the next Push() or Flush().
type StreamRehydrator struct {
	vault     vault.Provider
	masterKey []byte
	buf       strings.Builder
}

// NewStreamRehydrator creates a StreamRehydrator. Exported for testing.
func NewStreamRehydrator(v vault.Provider, masterKey []byte) *StreamRehydrator {
	return &StreamRehydrator{vault: v, masterKey: masterKey}
}

// newStreamRehydrator is the unexported alias used within the handler package.
func newStreamRehydrator(v vault.Provider, masterKey []byte) *StreamRehydrator {
	return NewStreamRehydrator(v, masterKey)
}

// Push accepts the next upstream delta. It returns the rehydrated text that is
// safe to emit now. Any incomplete vault token tail is held internally.
func (r *StreamRehydrator) Push(delta string) (string, error) {
	r.buf.WriteString(delta)
	safe, hold := SplitAtTokenBoundary(r.buf.String())
	r.buf.Reset()
	r.buf.WriteString(hold)
	if safe == "" {
		return "", nil
	}
	if r.vault == nil {
		return safe, nil // test/noop mode
	}
	return proxy.RehydrateString(r.vault, r.masterKey, safe)
}

// Flush drains any remaining buffer at end-of-stream.
func (r *StreamRehydrator) Flush() (string, error) {
	remaining := r.buf.String()
	r.buf.Reset()
	if remaining == "" {
		return "", nil
	}
	if r.vault == nil {
		return remaining, nil // test/noop mode
	}
	return proxy.RehydrateString(r.vault, r.masterKey, remaining)
}
