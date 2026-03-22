package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	seedHex := "30e053baa214d8dbcd8541ab97e3694770734be9660cc4015d348c8160ebdd57"
	expectedPubBase64 := "aeWbzWffHRzlbPJEgEIjNY0ymlZBAIkvUgrF0fYvoF8="

	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		fmt.Printf("Error decoding hex: %v\n", err)
		return
	}

	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubBase64 := base64.StdEncoding.EncodeToString(pubKey)

	fmt.Printf("Derived Public Key: %s\n", pubBase64)
	fmt.Printf("Expected Public Key: %s\n", expectedPubBase64)

	if pubBase64 == expectedPubBase64 {
		fmt.Println("MATCH!")
	} else {
		fmt.Println("NO MATCH!")
	}
}
