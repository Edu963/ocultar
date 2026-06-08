package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// The public key used by the refinery to verify licenses.
var licensePubKeyBase64 = "Hs7yFmulrC4S7kFbRT3MlKc5/U2GfTQ8p4XDJIl2gNQ="

type Payload struct {
	CustomerName string `json:"CustomerName"`
	Tier         string `json:"Tier"` // "community" or "enterprise"
	ExpiryDate   int64  `json:"ExpiryDate"`
	Capabilities uint64 `json:"Capabilities"` // Bitmask for Pro features
}

const (
	CapProSlack      uint64 = 1 << 0
	CapProSharePoint uint64 = 1 << 1
)

// Active holds the currently loaded license state.
var Active Payload

// IsEnterprise returns true if the active license tier explicitly grants enterprise capabilities.
func IsEnterprise() bool {
	return Active.Tier == "enterprise"
}

// HasProConnector returns true if the current license is enterprise AND has the specific bit set.
// If the license is enterprise but Capabilities is 0, all connectors are allowed (backward compat).
func HasProConnector(cap uint64) bool {
	if Active.Tier != "enterprise" {
		return false
	}
	if Active.Capabilities == 0 {
		return true
	}
	return (Active.Capabilities & cap) != 0
}

// verifyLicense verifies a signed license string.
// Format: base64(signature) + "." + base64(json_payload)
func verifyLicense(signedLicense string) (Payload, bool) {
	parts := strings.Split(signedLicense, ".")
	if len(parts) != 2 {
		return Payload{}, false
	}

	sigBytes, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return Payload{}, false
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return Payload{}, false
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(licensePubKeyBase64)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return Payload{}, false
	}

	if !ed25519.Verify(ed25519.PublicKey(pubKeyBytes), payloadBytes, sigBytes) {
		return Payload{}, false
	}

	var p Payload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return Payload{}, false
	}

	// ExpiryDate == 0 is treated as expired (not "no expiry") — every valid license
	// must carry a future timestamp. This prevents zero-expiry payloads from
	// becoming eternal enterprise licenses if the signing key is ever compromised.
	if p.ExpiryDate == 0 || time.Now().Unix() > p.ExpiryDate {
		return Payload{}, false
	}

	return p, true
}

// Load reads OCU_LICENSE_KEY from the environment and populates Active.
// If the key is absent or invalid, Active defaults to community tier (fail-safe).
func Load() {
	key := os.Getenv("OCU_LICENSE_KEY")
	if key == "" {
		log.Println("[license] OCU_LICENSE_KEY not set — running community tier")
		Active = Payload{Tier: "community"}
		return
	}

	payload, ok := verifyLicense(key)
	if !ok {
		log.Println("[license] OCU_LICENSE_KEY is set but invalid — defaulting to community tier")
		Active = Payload{Tier: "community"}
		return
	}

	Active = payload
	log.Printf("[license] Loaded license for %q — tier: %s", payload.CustomerName, payload.Tier)
}
