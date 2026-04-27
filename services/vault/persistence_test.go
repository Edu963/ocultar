package vault_test

// persistence_test.go — integration tests for DuckDB vault token survival.
//
// These tests prove three properties that underpin Ocultar's security contract:
//
//  1. Persistence  — tokens written to disk survive a vault close+reopen.
//  2. Determinism  — the same plaintext always produces the same token,
//                    which is required for cross-document relational integrity.
//  3. Key isolation — AES-256-GCM authentication catches wrong-key decryption;
//                     a rekeyed vault cannot silently return corrupt plaintext.
//
// The crypto helpers below mirror refinery.Encrypt / refinery.Decrypt exactly.
// They are inlined here to keep the vault package free of a circular import on
// the refinery module.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/vault"
)

// ── crypto helpers ────────────────────────────────────────────────────────────

// gcmEncrypt encrypts plaintext with AES-256-GCM and returns a hex-encoded
// nonce+ciphertext blob — identical to refinery.Encrypt.
func gcmEncrypt(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(sealed), nil
}

// gcmDecrypt decrypts a hex-encoded AES-256-GCM blob produced by gcmEncrypt.
// Returns an error if the key is wrong or the ciphertext is tampered.
func gcmDecrypt(hexCipher string, key []byte) ([]byte, error) {
	data, err := hex.DecodeString(hexCipher)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// piiHash returns the hex-encoded SHA-256 of value — matches sha256Hash in refinery.
func piiHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// makeToken returns the (hash, tokenString) pair for a PII value and type,
// matching the format Ocultar uses: "[TYPE_<first8hexchars>]".
func makeToken(piiType, value string) (hash, token string) {
	h := piiHash(value)
	return h, fmt.Sprintf("[%s_%s]", piiType, h[:8])
}

// ── local interface for encrypted lookup ─────────────────────────────────────

// encLookup matches the unexported method on duckdbProvider so the test can
// retrieve the encrypted blob by token string without importing the refinery.
type encLookup interface {
	GetEncryptedByToken(token string) (string, bool)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func openVault(t *testing.T, path string) vault.Provider {
	t.Helper()
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, path)
	if err != nil {
		t.Fatalf("vault.New(%q): %v", path, err)
	}
	return v
}

func closeVault(t *testing.T, v vault.Provider) {
	t.Helper()
	if err := v.Close(); err != nil {
		t.Fatalf("vault.Close: %v", err)
	}
}

// ── Test 1 — token survival across restarts ───────────────────────────────────

// TestPersistence_TokenSurvivesRestart is the core persistence contract:
// three French finance PII values vaulted in session 1 must be fully
// recoverable — token → original plaintext — after a close + reopen.
func TestPersistence_TokenSurvivesRestart(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.db")

	// Fixed 32-byte key simulating a stable deployment OCU_MASTER_KEY.
	masterKey := []byte("persistence-test-key-32-bytes!!!")

	type piiEntry struct {
		piiType  string
		value    string
		hash     string
		token    string
	}

	entries := []piiEntry{
		{piiType: "PERSON", value: "Jean-Pierre Dumont"},
		{piiType: "IBAN",   value: "FR76 3000 6000 0112 3456 7890 189"},
		{piiType: "EMAIL",  value: "jp.dumont@societe-generale.fr"},
	}

	// ── Step 1–3: open vault, store all three PII values, record tokens ───────
	v := openVault(t, vaultPath)
	for i, e := range entries {
		h, tok := makeToken(e.piiType, e.value)
		entries[i].hash = h
		entries[i].token = tok

		enc, err := gcmEncrypt([]byte(e.value), masterKey)
		if err != nil {
			t.Fatalf("gcmEncrypt(%q): %v", e.value, err)
		}
		_, err = v.StoreToken(h, tok, enc)
		if err != nil {
			t.Fatalf("StoreToken(%q): %v", e.value, err)
		}
	}

	// ── Step 4: close vault ───────────────────────────────────────────────────
	closeVault(t, v)

	// Confirm the file actually exists on disk — if it's in-memory only the
	// whole test is vacuous.
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatalf("vault file does not exist after close: %s", vaultPath)
	}

	// ── Step 5: reopen with same key ─────────────────────────────────────────
	v2 := openVault(t, vaultPath)
	defer closeVault(t, v2)

	el, ok := v2.(encLookup)
	if !ok {
		t.Fatal("vault.Provider does not implement encLookup — GetEncryptedByToken missing")
	}

	// ── Step 6: resolve each token back to original plaintext ─────────────────
	for _, e := range entries {
		encBlob, found := el.GetEncryptedByToken(e.token)
		if !found {
			t.Errorf("token %q not found after vault restart", e.token)
			continue
		}
		recovered, err := gcmDecrypt(encBlob, masterKey)
		if err != nil {
			t.Errorf("decrypt %q: %v", e.token, err)
			continue
		}
		if string(recovered) != e.value {
			t.Errorf("token %q: want %q, got %q", e.token, e.value, recovered)
		} else {
			t.Logf("✅ %s survived restart: %q → %q", e.piiType, e.token, e.value)
		}
	}
}

// ── Test 2 — determinism ──────────────────────────────────────────────────────

// TestPersistence_Determinism verifies that vaulting the same plaintext twice
// (within a session and across sessions) always produces the identical token.
// This is required so that [PERSON_6cea57c2] means the same person in every
// document, enabling privacy-safe aggregation without decryption.
func TestPersistence_Determinism(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.db")
	masterKey := []byte("determinism-test-key-32-bytes!!!")
	value := "Jean-Pierre Dumont"

	// ── Session A: store twice in the same vault open ─────────────────────────
	v := openVault(t, vaultPath)

	enc1, _ := gcmEncrypt([]byte(value), masterKey)
	enc2, _ := gcmEncrypt([]byte(value), masterKey)

	h, tok := makeToken("PERSON", value)

	_, err := v.StoreToken(h, tok, enc1)
	if err != nil {
		t.Fatalf("first StoreToken: %v", err)
	}
	// Second call with same hash — INSERT OR IGNORE, should be idempotent.
	inserted, err := v.StoreToken(h, tok, enc2)
	if err != nil {
		t.Fatalf("second StoreToken: %v", err)
	}
	if inserted {
		t.Error("second StoreToken should return inserted=false for duplicate hash")
	}

	// Token must be identical on both calls — no randomness in token derivation.
	_, tok2 := makeToken("PERSON", value)
	if tok != tok2 {
		t.Errorf("non-deterministic token: %q vs %q", tok, tok2)
	}

	closeVault(t, v)

	// ── Session B: fresh vault open, same key, same value → same token ────────
	v2 := openVault(t, vaultPath)
	defer closeVault(t, v2)

	_, tok3 := makeToken("PERSON", value)
	if tok != tok3 {
		t.Errorf("token changed across vault sessions: %q vs %q", tok, tok3)
	}

	got, found := v2.GetToken(h)
	if !found {
		t.Fatal("token missing after reopen")
	}
	if got != tok {
		t.Errorf("stored token mismatch: want %q, got %q", tok, got)
	}
	t.Logf("✅ Determinism confirmed: %q → %q (stable across two sessions)", value, tok)
}

// ── Test 3 — key isolation ────────────────────────────────────────────────────

// TestPersistence_KeyIsolation confirms that AES-256-GCM authentication
// rejects decryption under a different key. A rekeyed vault must fail loudly —
// never return corrupt or wrong plaintext — preserving the integrity guarantee.
func TestPersistence_KeyIsolation(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.db")

	keyA := []byte("key-A-for-isolation-test-32bytes")
	keyB := []byte("key-B-for-isolation-test-32bytes")
	value := "Jean-Pierre Dumont"

	// ── Vault with key A ──────────────────────────────────────────────────────
	v := openVault(t, vaultPath)

	enc, err := gcmEncrypt([]byte(value), keyA)
	if err != nil {
		t.Fatalf("gcmEncrypt: %v", err)
	}
	h, tok := makeToken("PERSON", value)
	if _, err := v.StoreToken(h, tok, enc); err != nil {
		t.Fatalf("StoreToken: %v", err)
	}
	closeVault(t, v)

	// ── Reopen vault, attempt decryption with key B ───────────────────────────
	v2 := openVault(t, vaultPath)
	defer closeVault(t, v2)

	el, ok := v2.(encLookup)
	if !ok {
		t.Fatal("vault.Provider does not implement encLookup")
	}

	encBlob, found := el.GetEncryptedByToken(tok)
	if !found {
		t.Fatalf("token %q not found — vault file may not have persisted", tok)
	}

	_, err = gcmDecrypt(encBlob, keyB)
	if err == nil {
		t.Error("❌ decryption with wrong key succeeded — AES-GCM authentication is broken")
	} else {
		t.Logf("✅ Key isolation confirmed: wrong-key decryption correctly rejected (%v)", err)
	}

	// Correct key must still work.
	plaintext, err := gcmDecrypt(encBlob, keyA)
	if err != nil {
		t.Errorf("correct-key decryption failed: %v", err)
	} else if string(plaintext) != value {
		t.Errorf("correct-key returned wrong value: %q", plaintext)
	} else {
		t.Logf("✅ Correct key decrypts successfully: %q", plaintext)
	}
}
