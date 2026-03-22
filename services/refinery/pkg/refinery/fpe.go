package refinery

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"strconv"
)

// ApplyBucketing generalises a given numeric string into a statistical range (e.g., age 34 -> "30-40")
// This preserves statistical utility for analytics/data-science teams without exposing exact PII.
func (e *Refinery) ApplyBucketing(numStr string, bucketSize int) (string, error) {
	val, err := strconv.Atoi(numStr)
	if err != nil {
		return numStr, nil // Retain original if not numeric
	}

	if bucketSize <= 0 {
		return numStr, fmt.Errorf("bucket size must be greater than zero")
	}

	lower := (val / bucketSize) * bucketSize
	upper := lower + bucketSize

	if e.DryRun || e.Report {
		e.hitsMutex.Lock()
		e.Hits["BUCKET"] = append(e.Hits["BUCKET"], fmt.Sprintf("%d-%d", lower, upper))
		e.hitsMutex.Unlock()
	}

	return fmt.Sprintf("%d-%d", lower, upper), nil
}

// ApplyFPE performs Format-Preserving Encryption on a numeric string (e.g., Credit Card numbers).
// It retains the length and digit formatting using deterministic encryption, permitting secure analytics.
func (e *Refinery) ApplyFPE(numericStr string) (string, error) {
	if len(numericStr) == 0 {
		return numericStr, nil
	}

	// Deterministic mapping via AES CTR module logic
	block, err := aes.NewCipher(e.MasterKey[:32])
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, iv)

	pt := []byte(numericStr)
	ct := make([]byte, len(pt))
	stream.XORKeyStream(ct, pt)

	// Map back to ASCII subset to preserve numeric format
	result := make([]byte, len(numericStr))
	for i := 0; i < len(numericStr); i++ {
		if numericStr[i] >= '0' && numericStr[i] <= '9' {
			offset := int(ct[i]) % 10
			digit := int(numericStr[i] - '0')
			result[i] = '0' + byte((digit+offset)%10)
		} else {
			// Keep non-numeric formatting characters untouched (e.g. '-' or ' ')
			result[i] = numericStr[i]
		}
	}

	if e.DryRun || e.Report {
		e.hitsMutex.Lock()
		e.Hits["FPE"] = append(e.Hits["FPE"], "[FPE_TOKEN]")
		e.hitsMutex.Unlock()
	}

	return string(result), nil
}
