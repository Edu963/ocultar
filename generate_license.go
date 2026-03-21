package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Payload struct {
	CustomerName string `json:"CustomerName"`
	Tier         string `json:"Tier"` // "community" or "enterprise"
	ExpiryDate   int64  `json:"ExpiryDate"`
	Capabilities uint64 `json:"Capabilities"` // Bitmask for Pro features
}

func main() {
	// Seed from OCU_MASTER_KEY in .env
	seedHex := "30e053baa214d8dbcd8541ab97e3694770734be9660cc4015d348c8160ebdd57"
	
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		fmt.Printf("Error decoding hex: %v\n", err)
		return
	}

	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubBase64 := base64.StdEncoding.EncodeToString(pubKey)

	payload := Payload{
		CustomerName: "Development",
		Tier:         "enterprise",
		ExpiryDate:   time.Now().AddDate(1, 0, 0).Unix(),
		Capabilities: 3, // Slack (1) + SharePoint (2)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)

	signature := ed25519.Sign(privKey, payloadBytes)
	sigB64 := base64.StdEncoding.EncodeToString(signature)

	licenseToken := fmt.Sprintf("%s.%s", sigB64, payloadB64)

	fmt.Printf("New Public Key (Put in license.go): %s\n", pubBase64)
	fmt.Printf("OCU_LICENSE_KEY=%s\n", licenseToken)
}
