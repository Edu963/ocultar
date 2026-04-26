package connector_test

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Edu963/ocultar/pkg/audit"
)

// ── NewImmutableLogger ────────────────────────────────────────────────────────

func TestNewImmutableLogger_CreatesFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("expected log file to be created on disk")
	}
}

func TestNewImmutableLogger_PublicKeyHex_NonEmpty(t *testing.T) {
	logger, err := audit.NewImmutableLogger(filepath.Join(t.TempDir(), "audit.log"))
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	pk := logger.PublicKeyHex()
	if pk == "" {
		t.Error("expected non-empty public key hex")
	}
	// Ed25519 public keys are 32 bytes → 64 hex chars.
	if len(pk) != 64 {
		t.Errorf("expected 64-char hex public key, got %d chars: %s", len(pk), pk)
	}
	if _, err := hex.DecodeString(pk); err != nil {
		t.Errorf("PublicKeyHex is not valid hex: %v", err)
	}
}

func TestNewImmutableLogger_BadPath_ReturnsError(t *testing.T) {
	_, err := audit.NewImmutableLogger("/nonexistent/path/audit.log")
	if err == nil {
		t.Error("expected error for unwritable path, got nil")
	}
}

// ── Log ───────────────────────────────────────────────────────────────────────

func TestLog_WritesEventToFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	if err := logger.Log("test-actor", "REDACT", "/api/v1/data", "SUCCESS", "email redacted"); err != nil {
		t.Fatalf("Log: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if len(data) == 0 {
		t.Error("log file is empty after Log() call")
	}
}

func TestLog_EventIsValidJSON(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	_ = logger.Log("actor", "ACTION", "resource", "SUCCESS", "details")

	f, _ := os.Open(logPath)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected at least one line in log")
	}
	var event audit.Event
	if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
		t.Fatalf("log line is not valid JSON: %v\nline: %s", err, scanner.Text())
	}
}

func TestLog_EventContainsExpectedFields(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	_ = logger.Log("sombra-gateway", "UPLOAD", "/connector/file", "SUCCESS", "1 entity redacted")

	f, _ := os.Open(logPath)
	defer f.Close()

	var event audit.Event
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	json.Unmarshal(scanner.Bytes(), &event)

	if event.Actor != "sombra-gateway" {
		t.Errorf("expected actor=sombra-gateway, got %q", event.Actor)
	}
	if event.Action != "UPLOAD" {
		t.Errorf("expected action=UPLOAD, got %q", event.Action)
	}
	if event.Status != "SUCCESS" {
		t.Errorf("expected status=SUCCESS, got %q", event.Status)
	}
	if event.Signature == "" {
		t.Error("expected non-empty signature on event")
	}
	if event.PrevHash == "" {
		t.Error("expected non-empty PrevHash on event")
	}
	if event.Timestamp == "" {
		t.Error("expected non-empty Timestamp on event")
	}
}

// ── Ed25519 signature verification ───────────────────────────────────────────

func TestLog_SignatureIsVerifiable(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	_ = logger.Log("actor", "DELETE", "vault/token/abc", "SUCCESS", "")

	pubKeyHex := logger.PublicKeyHex()
	pubKeyBytes, _ := hex.DecodeString(pubKeyHex)
	pubKey := ed25519.PublicKey(pubKeyBytes)

	f, _ := os.Open(logPath)
	defer f.Close()

	var event audit.Event
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	json.Unmarshal(scanner.Bytes(), &event)

	// Reconstruct the signed payload in the same format as ImmutableLogger.Log.
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		event.Timestamp, event.Actor, event.Action, event.Resource,
		event.Status, event.Details, event.PrevHash)

	sigBytes, err := hex.DecodeString(event.Signature)
	if err != nil {
		t.Fatalf("decode signature hex: %v", err)
	}
	if !ed25519.Verify(pubKey, []byte(payload), sigBytes) {
		t.Error("Ed25519 signature verification failed — log entry may have been tampered with")
	}
}

// ── Hash chaining ─────────────────────────────────────────────────────────────

func TestLog_HashChain_SecondEventReferencesFirstHash(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	_ = logger.Log("actor", "FIRST", "res", "SUCCESS", "")
	_ = logger.Log("actor", "SECOND", "res", "SUCCESS", "")

	f, _ := os.Open(logPath)
	defer f.Close()

	var events []audit.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e audit.Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			events = append(events, e)
		}
	}

	if len(events) < 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	first, second := events[0], events[1]

	// The first event's PrevHash must be the genesis hash (all zeros).
	if first.PrevHash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("first event should have genesis PrevHash, got: %s", first.PrevHash)
	}

	// The second event's PrevHash must not be the genesis hash.
	if second.PrevHash == first.PrevHash {
		t.Error("second event PrevHash should chain from first event, not reuse genesis hash")
	}
	if second.PrevHash == "" {
		t.Error("second event PrevHash is empty — chain is broken")
	}
}

// ── Concurrent Log calls ──────────────────────────────────────────────────────

func TestLog_ConcurrentCalls_NoDataRace(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	logger, err := audit.NewImmutableLogger(logPath)
	if err != nil {
		t.Fatalf("NewImmutableLogger: %v", err)
	}
	defer logger.Close()

	done := make(chan struct{})
	for i := range 10 {
		i := i
		go func() {
			_ = logger.Log("concurrent-actor", fmt.Sprintf("EVENT_%d", i), "resource", "SUCCESS", "")
			done <- struct{}{}
		}()
	}
	for range 10 {
		<-done
	}
}
