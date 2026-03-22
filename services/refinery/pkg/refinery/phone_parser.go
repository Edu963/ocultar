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

		valid := false
		// Try default regions to catch local formats if no + is provided.
		// Expanded to cover NA, EU, and LATAM comprehensively.
		for _, region := range []string{"FR", "US", "CO", "DE", "GB", "ES", "AR", "MX", "BR", "CA", "IT", "CH"} {
			num, err := phonenumbers.Parse(matchedStr, region)
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

// ParseAndReplacePhonesRaw is the underlying implementation that propagates errors.
func ParseAndReplacePhonesRaw(input string) [][]int {
	re := regexp.MustCompile(`(?:(?:\+|00)\s*)?\(?\d(?:[\s\-\.()]*\d){7,16}`)
	matches := re.FindAllStringIndex(input, -1)

	var validMatches [][]int
	for _, match := range matches {
		matchedStr := input[match[0]:match[1]]
		if regexp.MustCompile(`^\d{4}[-/.]\d{2}[-/.]\d{2}`).MatchString(matchedStr) {
			continue
		}
		valid := false
		for _, region := range []string{"FR", "US", "CO", "DE", "GB", "ES", "AR", "MX", "BR", "CA", "IT", "CH"} {
			num, err := phonenumbers.Parse(matchedStr, region)
			if err == nil && phonenumbers.IsValidNumber(num) {
				valid = true
				break
			}
		}
		if valid {
			validMatches = append(validMatches, match)
		}
	}
	return validMatches
}
