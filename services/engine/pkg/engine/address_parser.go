package engine

import (
	"regexp"
	"sort"
	"strings"
)

// ParsedAddress holds the extracted components of an address
type ParsedAddress struct {
	Original   string
	Number     string
	Street     string
	PostalCode string
	City       string
	Country    string
}

// localizedAddressParsers defines bounded heuristics for different global regions.
var localizedAddressParsers = []*regexp.Regexp{
	// LATAM: "Calle 123 # 45-67", "Cra. 15 #10-20"
	regexp.MustCompile(`(?i)(?:^|\W)((?:calle|carrera|cra\.?|avenida|av\.?|diagonal|transversal)\s+(?:[a-záéíóúñ0-9\-]+\s+){0,2}[a-záéíóúñ0-9\-]+\s*#\s*\d+\s*-\s*\d+)(?:$|\W)`),

	// EU FR: "12 rue de la paix 75001"
	regexp.MustCompile(`(?i)(?:^|\W)(\d{1,4}[a-z]?\s+(?:rue|avenue|boulevard|blvd|chemin|all[eé]e|place)\s+(?:[a-záéíóúñß\-]+\s+){0,3}[a-záéíóúñß\-]+(?:[\s,]+\d{4,5})?)(?:$|\W)`),

	// EU DE: "Hauptstraße 15, 10115"
	regexp.MustCompile(`(?i)(?:^|\W)([a-záéíóúñß\-]{2,30}\s?(?:strasse|straße|str\.?)\s+\d{1,4}[a-z]?(?:[\s,]+\d{4,5})?)(?:$|\W)`),

	// EU ES: "Calle Atocha 123"
	regexp.MustCompile(`(?i)(?:^|\W)((?:calle|paseo|plaza)\s+(?:[a-záéíóúñß\-]+\s+){0,2}[a-záéíóúñß\-]+\s+\d{1,4}(?:[\s,]+\d{4,5})?)(?:$|\W)`),

	// North America: "1600 Pennsylvania Avenue, Washington, DC 20500"
	regexp.MustCompile(`(?i)(?:^|\W)(\d{1,5}\s+(?:[a-z\-]+\s+){1,3}(?:street|st\.?|avenue|ave\.?|road|rd\.?|boulevard|blvd\.?|lane|ln\.?|drive|dr\.?)(?:[\s,]+[a-z\-]+){0,2}(?:[\s,]+[a-z]{2})?(?:[\s,]+\d{5}(?:-\d{4})?)?)(?:$|\W)`),

	// Isolated High-Risk Cities/Locales (fallback for loose mentions)
	regexp.MustCompile(`(?i)(?:^|\W)(grenoble|bogot[áa]|buenos aires|mallifaud|paris|madrid|berlin|london|new york|los angeles)(?:$|\W)`),
}

func ParseAndReplaceAddresses(input string, replaceFn func(match string) string) string {
	matches := ParseAndReplaceAddressesRaw(input)
	if len(matches) == 0 {
		return input
	}
	var out strings.Builder
	lastPos := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		if start < lastPos {
			continue // skip overlaps
		}
		out.WriteString(input[lastPos:start])
		out.WriteString(replaceFn(input[start:end]))
		lastPos = end
	}
	out.WriteString(input[lastPos:])
	return out.String()
}

// ParseAndReplaceAddressesRaw locates local address formats and returns their byte indices.
func ParseAndReplaceAddressesRaw(input string) [][]int {
	refined := input
	var allMatches [][]int

	for _, parser := range localizedAddressParsers {
		matches := parser.FindAllStringSubmatchIndex(refined, -1)
		if len(matches) == 0 {
			continue
		}

		for _, match := range matches {
			if len(match) < 4 {
				continue
			}
			// match[2] and match[3] bound the first capturing group (the address itself)
			start, end := match[2], match[3]
			if start == -1 || end == -1 {
				continue
			}

			// Clean any trailing punctuation safely captured
			matchedStr := refined[start:end]
			trimCount := 0
			for i := len(matchedStr) - 1; i >= 0; i-- {
				if matchedStr[i] == '.' || matchedStr[i] == ',' || matchedStr[i] == ' ' {
					trimCount++
				} else {
					break
				}
			}
			if trimCount > 0 {
				end -= trimCount
			}

			allMatches = append(allMatches, []int{start, end})
		}
	}

	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i][0] < allMatches[j][0]
	})

	return allMatches
}
