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

		// Ignore ISO dates (YYYY-MM-DD) and European-format dates (DD-MM-YYYY, DD/MM/YYYY)
		if regexp.MustCompile(`^\d{4}[-/.]\d{2}[-/.]\d{2}`).MatchString(matchedStr) {
			continue
		}
		if regexp.MustCompile(`^\d{1,2}[-/.]\d{1,2}[-/.]\d{4}`).MatchString(matchedStr) {
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

// nonPhoneContextRe matches known non-phone identifier labels in the text window
// immediately preceding a Tier-B candidate. When present, the candidate is NOT
// a phone number and must not be masked as PHONE (VAT, fiber ref, subscriber ID, etc.).
// Educational/clinical scoring terms prevent test-score tables from being misclassified.
var nonPhoneContextRe = regexp.MustCompile(`(?i)\b(?:tva|vat|siret|siren|iban|fibre|fiber|r[eé]f[eé]rence|identifiant|abonn[eé]|account|compte|registr[ae]|rang|percentile|subtest|brut|standard|quotient|composite|[eé]chelle|s[ae]m\b|note\b|vitesse|m[eé]moire|comp[eé]tence|processus)`)

// ParseAndReplacePhonesRaw is the underlying implementation that propagates errors.
func ParseAndReplacePhonesRaw(input string) [][]int {
	re := regexp.MustCompile(`(?:(?:\+|00)\s*)?(\d(?:[\s\-\.()]*\d){7,16})`)
	matches := re.FindAllStringIndex(input, -1)

	// Pre-compute Ki! allowlist sentinel spans — __KII_{hex}_{n}__ — so that
	// hex digits inside sentinels are never mistaken for phone numbers.
	kiiSentinelRe := regexp.MustCompile(`__KII_[0-9a-f]+_\d+__`)
	sentinelSpans := kiiSentinelRe.FindAllStringIndex(input, -1)

	// confidenceIntervalRe matches candidates that END with a space-separated hyphenated range,
	// e.g. "93 114-127", "66 97-113" — confidence intervals, score bands, page ranges.
	// This pattern ("digits space digits-digits" at end of string) never occurs in a phone number.
	confidenceIntervalRe := regexp.MustCompile(`\d+ \d+-\d+$`)

	// structuredValueSuffixRe matches suffixes that prove the candidate is NOT a phone:
	//   ":digits"  — age/years:months (11:10), ratio (7:10), time (1:30)
	//   ",digits"  — French decimal continuation (12,12 → the match ends at "12" and
	//                `,12` follows; French uses comma as decimal separator)
	// A real phone number is never immediately followed by these patterns.
	structuredValueSuffixRe := regexp.MustCompile(`^[;:,]\d`)

	var validMatches [][]int
	for _, match := range matches {
		matchedStr := input[match[0]:match[1]]

		// Skip matches that overlap with a Ki! allowlist sentinel.
		sentinelOverlap := false
		for _, s := range sentinelSpans {
			if match[0] < s[1] && match[1] > s[0] {
				sentinelOverlap = true
				break
			}
		}
		if sentinelOverlap {
			continue
		}

		// Skip if the match is immediately followed by a structured-value suffix.
		if match[1] < len(input) && structuredValueSuffixRe.MatchString(input[match[1]:]) {
			continue
		}
		// Skip confidence intervals and score ranges (e.g. "93 114-127", "66 97-113").
		if confidenceIntervalRe.MatchString(matchedStr) {
			continue
		}

		if regexp.MustCompile(`^\d{4}[-/.]?\d{2}[-/.]\d{2}`).MatchString(matchedStr) {
			continue // skip ISO dates (YYYY-MM-DD)
		}
		if regexp.MustCompile(`^\d{1,2}[-/.]\d{1,2}[-/.]\d{4}`).MatchString(matchedStr) {
			continue // skip European-format dates (DD-MM-YYYY, DD/MM/YYYY)
		}
		// Skip invoice-line patterns: year followed by quantity/amount (e.g. "2025 1 75.00")
		if yearAmountRe.MatchString(matchedStr) {
			continue
		}

		// Shared context check: skip when surrounding text contains non-phone identifiers.
		// Applied to BOTH Tier A and Tier B — prevents score-table rows from being masked
		// even when libphonenumber happens to accept the digit sequence.
		ctxStart := match[0] - 120
		if ctxStart < 0 {
			ctxStart = 0
		}
		ctx := strings.ToLower(input[ctxStart:match[0]])
		if nonPhoneContextRe.MatchString(ctx) {
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
		// Guard: skip when the surrounding 60-char window contains non-phone identifiers
		// (VAT numbers, fiber references, IBAN, SIRET, subscriber IDs) to avoid misclassification.
		digits := regexp.MustCompile(`\D`).ReplaceAllString(matchedStr, "")
		if len(digits) >= 7 && len(digits) <= 15 && broadPhoneRegex.MatchString(strings.TrimSpace(matchedStr)) {
			ctxStart := match[0] - 120
			if ctxStart < 0 {
				ctxStart = 0
			}
			ctx := strings.ToLower(input[ctxStart:match[0]])
			if nonPhoneContextRe.MatchString(ctx) {
				continue
			}
			validMatches = append(validMatches, match)
		}
	}
	return validMatches
}
