package audit

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Event represents a single entry in the immutable audit log.
type Event struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Status    string `json:"status"`
	Details   string `json:"details,omitempty"`
	PrevHash  string `json:"prev_hash"`
	Signature string `json:"signature"`
}

// ImmutableLogger handles cryptographically signed append-only logging
// to satisfy NIS2 and GDPR Article 30 requirements.
type ImmutableLogger struct {
	mu         sync.Mutex
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	logFile    *os.File
	lastHash   string
}

// NewImmutableLogger initializes a logger on disk with an ephemeral Ed25519
// key pair. Signatures are verifiable within the current process session only.
// For cross-session verifiability use NewImmutableLoggerWithKey.
func NewImmutableLogger(filePath string) (*ImmutableLogger, error) {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 keys: %v", err)
	}
	return newLogger(filePath, priv)
}

// NewImmutableLoggerWithKey initializes a logger with a caller-supplied
// private key. The same key must be retained to verify signatures after
// process restarts. privateKey must be a valid ed25519.PrivateKey (64 bytes).
func NewImmutableLoggerWithKey(filePath string, privateKey ed25519.PrivateKey) (*ImmutableLogger, error) {
	return newLogger(filePath, privateKey)
}

// LoadPrivateKeyFromHex decodes a hex-encoded 32-byte Ed25519 seed (as stored
// in OCU_AUDIT_PRIVATE_KEY) into a PrivateKey suitable for NewImmutableLoggerWithKey.
// Generate with: openssl rand -hex 32
func LoadPrivateKeyFromHex(seedHex string) (ed25519.PrivateKey, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("OCU_AUDIT_PRIVATE_KEY is not valid hex: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("OCU_AUDIT_PRIVATE_KEY: want %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func newLogger(filePath string, priv ed25519.PrivateKey) (*ImmutableLogger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %v", err)
	}
	return &ImmutableLogger{
		privateKey: priv,
		publicKey:  priv.Public().(ed25519.PublicKey),
		logFile:    f,
		lastHash:   "0000000000000000000000000000000000000000000000000000000000000000",
	}, nil
}

// Log records an event, chaining it to the previous hash and signing it.
func (l *ImmutableLogger) Log(actor, action, resource, status, details string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	e := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Status:    status,
		Details:   details,
		PrevHash:  l.lastHash,
	}

	// Payload to sign
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		e.Timestamp, e.Actor, e.Action, e.Resource, e.Status, e.Details, e.PrevHash)

	sig := ed25519.Sign(l.privateKey, []byte(payload))
	e.Signature = hex.EncodeToString(sig)

	// Hash the payload for the next event in the chain
	h := sha256.Sum256([]byte(payload))
	l.lastHash = hex.EncodeToString(h[:])

	logData, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if _, err := l.logFile.Write(append(logData, '\n')); err != nil {
		return fmt.Errorf("audit write failed: %v", err)
	}
	
	// Ensure it hits disk immediately (sync/fsync)
	return l.logFile.Sync()
}

// Close gracefully closes the log file.
func (l *ImmutableLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logFile.Close()
}

// PublicKey returns the hex string of the public key for verification.
func (l *ImmutableLogger) PublicKeyHex() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return hex.EncodeToString(l.publicKey)
}
