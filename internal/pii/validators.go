package pii

import "strconv"

// Validator is a function that returns true if the stripped value is mathematically valid
type Validator func(string) bool

// IsLuhnValid implements the Luhn algorithm (Mod-10) for credit cards.
func IsLuhnValid(s string) bool {
	digits := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	sum := 0
	shouldDouble := false
	for i := len(digits) - 1; i >= 0; i-- {
		n := int(digits[i] - '0')
		if shouldDouble {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		shouldDouble = !shouldDouble
	}
	return (sum % 10) == 0
}

// IsIBANValid implements the Mod-97 checksum (ISO 7064) for IBANs.
func IsIBANValid(s string) bool {
	norm := Normalize(s)
	if len(norm) < 15 || len(norm) > 34 {
		return false
	}
	rearranged := norm[4:] + norm[:4]
	var numeric string
	for _, ch := range rearranged {
		if ch >= 'A' && ch <= 'Z' {
			val := int(ch - 'A' + 10)
			numeric += strconv.Itoa(val)
		} else if ch >= '0' && ch <= '9' {
			numeric += string(ch)
		} else {
			return false
		}
	}
	mod := 0
	for _, ch := range numeric {
		mod = (mod*10 + int(ch-'0')) % 97
	}
	return mod == 1
}

// GetValidator returns the appropriate function for a given validation method
func GetValidator(name ValidationMethod) Validator {
	switch name {
	case ValLuhn:
		return IsLuhnValid
	case ValMod97:
		return IsIBANValid
	default:
		return nil
	}
}
