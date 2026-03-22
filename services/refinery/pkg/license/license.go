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

// Global active license state
var Active Payload

// IsEnterprise returns true if the active license tier explicitly grants enterprise capabilities.
func IsEnterprise() bool {
	return Active.Tier == "enterprise"
}

// HasProConnector returns true if the current license is enterprise AND has the specific bit set.
// If the license is enterprise but Capabilities is 0, we allow all for backward compatibility.
func HasProConnector(cap uint64) bool {
	if Active.Tier != "enterprise" {
		return false
	}
	if Active.Capabilities == 0 {
		return true // Default: enterprise allows all if bitmask is empty (backward compatibility)
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
		log.Printf("[ERROR] Invalid embedded public key configuration.")
		return Payload{}, false
	}

	if !ed25519.Verify(pubKeyBytes, payloadBytes, sigBytes) {
		return Payload{}, false // Invalid signature
	}

	var payload Payload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Payload{}, false
	}

	if payload.ExpiryDate < time.Now().Unix() {
		log.Printf("[WARN] License for %s expired.", payload.CustomerName)
		return Payload{}, false
	}

	return payload, true
}

// Load reads the license from the environment variable OCU_LICENSE_KEY
// or from a license.key file. If valid, sets Active and returns true.
// If invalid or missing, defaults to Community tier ("community").
func Load() {
	licenseData := os.Getenv("OCU_LICENSE_KEY")
	if licenseData == "" {
		if content, err := os.ReadFile("license.key"); err == nil {
			licenseData = strings.TrimSpace(string(content))
		}
	}

	if licenseData == "" {
		log.Printf("[INFO] No license key found. Reverting to Community Mode.")
		Active = Payload{Tier: "community"}
		return
	}

	payload, ok := verifyLicense(licenseData)
	if !ok {
		log.Printf("[WARN] Invalid or expired license key. Reverting to Community Mode.")
		Active = Payload{Tier: "community"}
		return
	}

	Active = payload
	log.Printf("[INFO] License verified. Tier: %s, Customer: %s", strings.ToUpper(Active.Tier), Active.CustomerName)
}
