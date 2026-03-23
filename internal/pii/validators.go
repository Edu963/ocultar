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
	if len(digits) < 9 || len(digits) > 19 {
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

// IsESDNIValid validates Spanish DNI/NIE using Modulo 23.
func IsESDNIValid(s string) bool {
	s = Normalize(s)
	if len(s) != 9 {
		return false
	}
	table := "TRWAGMYFPDXBNJZSQVHLCKE"
	prefix := s[0]
	var numericPart string
	if prefix >= 'X' && prefix <= 'Z' {
		val := map[byte]string{'X': "0", 'Y': "1", 'Z': "2"}[prefix]
		numericPart = val + s[1:8]
	} else if prefix >= '0' && prefix <= '9' {
		numericPart = s[:8]
	} else {
		return false
	}
	num, err := strconv.Atoi(numericPart)
	if err != nil {
		return false
	}
	return table[num%23] == s[8]
}

// IsITCFValid validates Italian Codice Fiscale check character.
func IsITCFValid(s string) bool {
	s = Normalize(s)
	if len(s) != 16 {
		return false
	}
	oddValues := map[byte]int{
		'0': 1, '1': 0, '2': 5, '3': 7, '4': 9, '5': 13, '6': 15, '7': 17, '8': 19, '9': 21,
		'A': 1, 'B': 0, 'C': 5, 'D': 7, 'E': 9, 'F': 13, 'G': 15, 'H': 17, 'I': 19, 'J': 21,
		'K': 2, 'L': 4, 'M': 18, 'N': 20, 'O': 11, 'P': 3, 'Q': 6, 'R': 8, 'S': 12, 'T': 14,
		'U': 16, 'V': 10, 'W': 22, 'X': 25, 'Y': 24, 'Z': 23,
	}
	evenValues := map[byte]int{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'A': 0, 'B': 1, 'C': 2, 'D': 3, 'E': 4, 'F': 5, 'G': 6, 'H': 7, 'I': 8, 'J': 9,
		'K': 10, 'L': 11, 'M': 12, 'N': 13, 'O': 14, 'P': 15, 'Q': 16, 'R': 17, 'S': 18, 'T': 19,
		'U': 20, 'V': 21, 'W': 22, 'X': 23, 'Y': 24, 'Z': 25,
	}
	sum := 0
	for i := 0; i < 15; i++ {
		if (i+1)%2 != 0 {
			sum += oddValues[s[i]]
		} else {
			sum += evenValues[s[i]]
		}
	}
	return byte('A'+(sum%26)) == s[15]
}

// IsNLBSNValid validates Dutch BSN using the 11-test.
func IsNLBSNValid(s string) bool {
	s = Normalize(s)
	if len(s) != 8 && len(s) != 9 {
		return false
	}
	if len(s) == 8 {
		s = "0" + s
	}
	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(s[i]-'0') * (9 - i)
	}
	sum -= int(s[8]-'0') * 1
	return sum != 0 && sum%11 == 0
}

// IsPLPESELValid validates Polish PESEL using weights.
func IsPLPESELValid(s string) bool {
	s = Normalize(s)
	if len(s) != 11 {
		return false
	}
	weights := []int{1, 3, 7, 9, 1, 3, 7, 9, 1, 3}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(s[i]-'0') * weights[i]
	}
	checksum := (10 - (sum % 10)) % 10
	return checksum == int(s[10]-'0')
}

// IsDESTIDValid validates German Steuer-ID using Modulo 11.
func IsDESTIDValid(s string) bool {
	s = Normalize(s)
	if len(s) != 11 {
		return false
	}
	// Simplified DE Steuer-ID check (Mod 11, RFC-style)
	remainder := 10
	for i := 0; i < 10; i++ {
		digit := int(s[i] - '0')
		sum := (digit + remainder) % 10
		if sum == 0 {
			sum = 10
		}
		remainder = (sum * 2) % 11
	}
	check := 11 - remainder
	if check == 10 {
		check = 0
	}
	return check == int(s[10]-'0')
}

// IsDKCPRValid validates Danish CPR using Modulo 11 with weights.
func IsDKCPRValid(s string) bool {
	s = Normalize(s)
	if len(s) != 10 {
		return false
	}
	weights := []int{4, 3, 2, 7, 6, 5, 4, 3, 2, 1}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(s[i]-'0') * weights[i]
	}
	return sum%11 == 0
}

// IsFIHETUValid validates Finnish HETU using Modulo 31.
func IsFIHETUValid(s string) bool {
	// Format: DDMMYY[+-A]ZZZX (X is checksum)
	// We only validate the checksum character here.
	if len(s) != 11 {
		return false
	}
	table := "0123456789ABCDEFHJKLMNPRSTUVWXY"
	// DDMMYY (6) + ZZZ (3) = 9 digits
	numStr := s[0:6] + s[7:10]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return false
	}
	return table[num%31] == s[10]
}

// IsSEPINValid validates Swedish Personal Identity Number using Luhn on the last 10 digits.
func IsSEPINValid(s string) bool {
	s = Normalize(s)
	if len(s) == 12 {
		s = s[2:] // Strip century
	}
	if len(s) != 10 {
		return false
	}
	return IsLuhnValid(s)
}

// IsBRCPFValid validates Brazil CPF using Dual Modulo 11.
func IsBRCPFValid(s string) bool {
	s = Normalize(s)
	if len(s) != 11 {
		return false
	}
	// All same digits are invalid
	allSame := true
	for i := 1; i < 11; i++ {
		if s[i] != s[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// First digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(s[i]-'0') * (10 - i)
	}
	remainder := (sum * 10) % 11
	if remainder == 10 {
		remainder = 0
	}
	if remainder != int(s[9]-'0') {
		return false
	}

	// Second digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(s[i]-'0') * (11 - i)
	}
	remainder = (sum * 10) % 11
	if remainder == 10 {
		remainder = 0
	}
	return remainder == int(s[10]-'0')
}

// IsCLRUTValid validates Chile RUT using Modulo 11.
func IsCLRUTValid(s string) bool {
	s = Normalize(s)
	if len(s) < 8 || len(s) > 9 {
		return false
	}
	body := s[:len(s)-1]
	checkDigit := s[len(s)-1]
	if checkDigit >= 'a' && checkDigit <= 'z' {
		checkDigit -= 32 // To upper
	}

	sum := 0
	multiplier := 2
	for i := len(body) - 1; i >= 0; i-- {
		sum += int(body[i]-'0') * multiplier
		multiplier++
		if multiplier > 7 {
			multiplier = 2
		}
	}

	expectedMod := 11 - (sum % 11)
	var expectedDigit byte
	if expectedMod == 11 {
		expectedDigit = '0'
	} else if expectedMod == 10 {
		expectedDigit = 'K'
	} else {
		expectedDigit = byte('0' + expectedMod)
	}

	return expectedDigit == checkDigit
}

// IsESCIFValid validates Spanish CIF using documented check digit logic.
func IsESCIFValid(s string) bool {
	s = Normalize(s)
	if len(s) != 9 {
		return false
	}
	prefix := s[0]
	if prefix >= 'a' && prefix <= 'z' {
		prefix -= 32
	}
	
	sum := 0
	for i := 1; i < 8; i++ {
		digit := int(s[i] - '0')
		if i%2 != 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	
	control := (10 - (sum % 10)) % 10
	suffix := s[8]
	if suffix >= 'a' && suffix <= 'j' {
		suffix -= 32
	}

	// Suffix types:
	// - Numeric only: P, Q, R, S, W (non-standard but common)
	// - Letter only: A, B, C, D, E, F, G, H
	// - Both: others
	
	isLetterOnly := prefix == 'P' || prefix == 'Q' || prefix == 'R' || prefix == 'S' || prefix == 'W'
	isNumericOnly := prefix == 'A' || prefix == 'B' || prefix == 'C' || prefix == 'D' || prefix == 'E' || prefix == 'F' || prefix == 'G' || prefix == 'H'
	
	if isLetterOnly {
		return suffix == "ABCDEFGHIJ"[control]
	}
	if isNumericOnly {
		return suffix == byte('0'+control)
	}
	
	// Both allowed
	return suffix == byte('0'+control) || suffix == "ABCDEFGHIJ"[control]
}

// IsAadhaarValid validates India's 12-digit Aadhaar number using Verhoeff algorithm.
func IsAadhaarValid(s string) bool {
	s = Normalize(s)
	if len(s) != 12 {
		return false
	}

	d := [][]int{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{1, 2, 3, 4, 0, 6, 7, 8, 9, 5},
		{2, 3, 4, 0, 1, 7, 8, 9, 5, 6},
		{3, 4, 0, 1, 2, 8, 9, 5, 6, 7},
		{4, 0, 1, 2, 3, 9, 5, 6, 7, 8},
		{5, 9, 8, 7, 6, 0, 4, 3, 2, 1},
		{6, 5, 9, 8, 7, 1, 0, 4, 3, 2},
		{7, 6, 5, 9, 8, 2, 1, 0, 4, 3},
		{8, 7, 6, 5, 9, 3, 2, 1, 0, 4},
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	}
	p := [][]int{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{1, 5, 7, 6, 2, 8, 3, 0, 9, 4},
		{5, 8, 0, 3, 7, 9, 6, 1, 4, 2},
		{8, 9, 1, 6, 0, 4, 3, 5, 2, 7},
		{9, 4, 5, 3, 1, 2, 6, 8, 7, 0},
		{4, 2, 8, 6, 5, 7, 3, 9, 0, 1},
		{2, 7, 9, 3, 8, 0, 6, 4, 1, 5},
		{7, 0, 4, 6, 9, 1, 3, 2, 5, 8},
	}

	c := 0
	for i := 0; i < len(s); i++ {
		digit := int(s[len(s)-1-i] - '0')
		c = d[c][p[i%8][digit]]
	}
	return c == 0
}

// IsSGIDValid validates Singapore NRIC/FIN using weighted checksum.
func IsSGIDValid(s string) bool {
	if len(s) != 9 {
		return false
	}
	prefix := s[0]
	if prefix >= 'a' && prefix <= 'z' {
		prefix -= 32
	}
	suffix := s[8]
	if suffix >= 'a' && suffix <= 'z' {
		suffix -= 32
	}

	digits := s[1:8]
	weights := []int{2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 7; i++ {
		sum += int(digits[i]-'0') * weights[i]
	}

	offset := 0
	switch prefix {
	case 'S':
		offset = 0
	case 'T':
		offset = 4
	case 'F':
		offset = 1
	case 'G':
		offset = 5
	case 'M':
		offset = 3 // M series uses a different logic for mapping, but offset 3 is standard for sum
	default:
		return false
	}

	sum += offset
	remainder := sum % 11
	
	var table string
	switch prefix {
	case 'S':
		table = "JZIHGFEDCBA"
	case 'T':
		table = "GFEDCBAJZIH"
	case 'F':
		table = "XWUTRQPNMLK"
	case 'G':
		table = "TRQPNMLKXWU"
	case 'M':
		table = "KLJN PQR TUWX"
		// M series: 0=K, 1=L, 2=J, 3=N, 4=P, 5=Q, 6=R, 7=T, 8=U, 9=W, 10=X
		table = "KLJN PQR TUWX" // Note: space at index 4 is just for clarity or if it's 1-indexed? No.
		table = "KLJN PQRTUWX" // 11 characters
		table = "KLJNPQRTUWX"
	}

	return table[remainder] == suffix
}

// GetValidator returns the appropriate function for a given validation method
func GetValidator(name ValidationMethod) Validator {
	switch name {
	case ValLuhn:
		return IsLuhnValid
	case ValMod97:
		return IsIBANValid
	case ValESDNI:
		return IsESDNIValid
	case ValITCF:
		return IsITCFValid
	case ValNLBSN:
		return IsNLBSNValid
	case ValPLPSL:
		return IsPLPESELValid
	case ValDESTID:
		return IsDESTIDValid
	case ValDKCPR:
		return IsDKCPRValid
	case ValFIHETU:
		return IsFIHETUValid
	case ValSEPIN:
		return IsSEPINValid
	case ValBRCPF:
		return IsBRCPFValid
	case ValCLRUT:
		return IsCLRUTValid
	case ValIndiaAadhaar:
		return IsAadhaarValid
	case ValSingaporeID:
		return IsSGIDValid
	case ValESCIF:
		return IsESCIFValid
	default:
		return nil
	}
}
