//go:build ignore

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type LicensePayload struct {
	CustomerName string `json:"CustomerName"`
	Tier         string `json:"Tier"` // "community" or "enterprise"
	ExpiryDate   int64  `json:"ExpiryDate"`
}

func main() {
	customer := flag.String("customer", "Acme Corp", "Name of the customer")
	tier := flag.String("tier", "enterprise", "License tier (community or enterprise)")
	expiryYears := flag.Int("expiry", 1, "Number of years until expiry")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// 1. Generate new Ed25519 keypair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Error generating keypair: %v\n", err)
		os.Exit(1)
	}

	pubBase64 := base64.StdEncoding.EncodeToString(pubKey)
	privBase64 := base64.StdEncoding.EncodeToString(privKey)

	fmt.Println("=== KEYPAIR FOR OCULTAR (Save Securely) ===")
	fmt.Printf("Public Key (Embed in license.go): %s\n", pubBase64)
	fmt.Printf("Private Key (Strictly Confidential): %s\n\n", privBase64)

	// 2. Draft the license payload based on flags
	payload := LicensePayload{
		CustomerName: *customer,
		Tier:         *tier,
		ExpiryDate:   time.Now().AddDate(*expiryYears, 0, 0).Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		os.Exit(1)
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)

	// 3. Sign the payload bytes
	signature := ed25519.Sign(privKey, payloadBytes)
	sigB64 := base64.StdEncoding.EncodeToString(signature)

	// 4. Construct final license token
	licenseToken := fmt.Sprintf("%s.%s", sigB64, payloadB64)

	fmt.Printf("=== LICENSE FOR: %s (%s) ===\n", *customer, *tier)
	fmt.Printf("Expiry: %v\n", time.Unix(payload.ExpiryDate, 0).Format(time.RFC3339))
	fmt.Printf("OCU_LICENSE_KEY=%s\n", licenseToken)
}
