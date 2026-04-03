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

// NewImmutableLogger initializes a logger on disk. In a production
// scenario, the private key would be securely loaded from a KMS.
// Here we generate an ephemeral key for the session to prove the capability.
func NewImmutableLogger(filePath string) (*ImmutableLogger, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 keys: %v", err)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %v", err)
	}

	return &ImmutableLogger{
		privateKey: priv,
		publicKey:  pub,
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
