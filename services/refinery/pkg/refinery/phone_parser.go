package refinery

import (
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ParseAndReplacePhones extracts and normalizes international and local phone numbers,
// safely skipping overlapping OCULTAR tokens and ISO dates, and invokes replaceFn on valid matches.
func ParseAndReplacePhones(input string, replaceFn func(match string) string) string {
	// A broad regex for international and local phone sequences.
	// Captures optional leading +, spaces, or parentheses, and ensures it ends on a digit.
	re := regexp.MustCompile(`(?:(?:\+|00)\s*)?\(?\d(?:[\s\-\.()]*\d){7,16}`)
	// tokenPattern is defined in refinery.go
	tokens := tokenPattern.FindAllStringIndex(input, -1)
	matches := re.FindAllStringIndex(input, -1)

	var out strings.Builder
	lastPos := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		overlap := false
		for _, t := range tokens {
			if start < t[1] && end > t[0] {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}

		matchedStr := input[start:end]

		// Ignore ISO dates that look like phone numbers
		if regexp.MustCompile(`^\d{4}[-/.]\d{2}[-/.]\d{2}`).MatchString(matchedStr) {
			continue
		}
		// Skip invoice-line patterns: year followed by quantity/amount (e.g. "2025 1 75.00")
		if yearAmountRe.MatchString(matchedStr) {
			continue
		}

		// Normalize European trunk-prefix notation: "+33 (0)4 ..." → "+33 4 ..."
		normalized := trunkPrefixRe.ReplaceAllString(matchedStr, "")

		valid := false
		// Try default regions to catch local formats if no + is provided.
		// Expanded to cover NA, EU, and LATAM comprehensively.
		for _, region := range []string{"FR", "US", "CO", "DE", "GB", "ES", "AR", "MX", "BR", "CA", "IT", "CH"} {
			num, err := phonenumbers.Parse(normalized, region)
			if err == nil && phonenumbers.IsValidNumber(num) {
				valid = true
				break
			}
		}

		if !valid {
			continue
		}

		out.WriteString(input[lastPos:start])
		out.WriteString(replaceFn(matchedStr))
		lastPos = end
	}
	out.WriteString(input[lastPos:])
	return out.String()
}

// broadPhoneRegex matches structurally valid phone sequences (7–15 digits, optional
// leading + or 00, separated by spaces/dashes/dots) that libphonenumber may reject
// as "unallocated" but which a human would clearly read as a phone number.
var broadPhoneRegex = regexp.MustCompile(`^(?:\+|00)?[\d\s\-\.\(\)]{7,20}$`)

// trunkPrefixRe strips the European `(0)` trunk-prefix notation so libphonenumber
// can validate "+33 (0)4 76 98 93 85" as a standard E.164 number.
var trunkPrefixRe = regexp.MustCompile(`\(\s*0\s*\)\s*`)

// yearAmountRe guards against invoice lines where "YYYY quantity amount" (e.g. "2025 1 75.00")
// is mistakenly read as a phone number by the broad Tier-B fallback.
var yearAmountRe = regexp.MustCompile(`^(?:19|20)\d{2}[\s/]`)

// ParseAndReplacePhonesRaw is the underlying implementation that propagates errors.
func ParseAndReplacePhonesRaw(input string) [][]int {
	re := regexp.MustCompile(`(?:(?:\+|00)\s*)?(\d(?:[\s\-\.()]*\d){7,16})`)
	matches := re.FindAllStringIndex(input, -1)

	var validMatches [][]int
	for _, match := range matches {
		matchedStr := input[match[0]:match[1]]
		if regexp.MustCompile(`^\d{4}[-/.]?\d{2}[-/.]\d{2}`).MatchString(matchedStr) {
			continue // skip ISO dates
		}
		// Skip invoice-line patterns: year followed by quantity/amount (e.g. "2025 1 75.00")
		if yearAmountRe.MatchString(matchedStr) {
			continue
		}

		// Normalize European trunk-prefix notation before libphonenumber validation.
		// "+33 (0)4 76 98 93 85" → "+33 4 76 98 93 85"
		normalized := trunkPrefixRe.ReplaceAllString(matchedStr, "")

		// Tier A: libphonenumber strict validation (most precise)
		libValid := false
		for _, region := range []string{"FR", "US", "CO", "DE", "GB", "ES", "AR", "MX", "BR", "CA", "IT", "CH"} {
			num, err := phonenumbers.Parse(normalized, region)
			if err == nil && phonenumbers.IsValidNumber(num) {
				libValid = true
				break
			}
		}
		if libValid {
			validMatches = append(validMatches, match)
			continue
		}

		// Tier B: Broad "looks like a phone" fallback (Fail-Closed for placeholders)
		// Count just the digits — any sequence of 7–15 digits with phone separators is suspicious.
		digits := regexp.MustCompile(`\D`).ReplaceAllString(matchedStr, "")
		if len(digits) >= 7 && len(digits) <= 15 && broadPhoneRegex.MatchString(strings.TrimSpace(matchedStr)) {
			validMatches = append(validMatches, match)
		}
	}
	return validMatches
}
