package pii

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)


// DetectionResult defines the structured output of a PII hit per Phase 4 requirements
type DetectionResult struct {
	Entity     string   `json:"entity"`
	Value      string   `json:"-"` // Not serialized by default
	ValueHash  string   `json:"value_hash"`
	Confidence float64  `json:"confidence"`
	Method     []string `json:"method"`
	Location   struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"location"`
}

type Engine struct {
	registry []EntityDef
}

func NewEngine() *Engine {
	return &Engine{registry: Registry}
}

// Scan performs the exhaustive deterministic sweep with validation
func (e *Engine) Scan(input string) []DetectionResult {
	var results []DetectionResult

	for _, entity := range e.registry {
		// Use FindAllStringSubmatchIndex to get submatch boundaries
		matches := entity.Pattern.FindAllStringSubmatchIndex(input, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			
			// Default to full match
			start, end := match[0], match[1]
			
			// If CaptureGroup is specified, use its boundaries
			if entity.CaptureGroup > 0 && len(match) >= (entity.CaptureGroup*2)+2 {
				cgStart := match[entity.CaptureGroup*2]
				cgEnd := match[(entity.CaptureGroup*2)+1]
				if cgStart != -1 && cgEnd != -1 {
					start, end = cgStart, cgEnd
				}
			}

			if start == -1 || end == -1 || start >= end {
				continue
			}

			matchedStr := input[start:end]

			if len(matchedStr) < entity.MinLength {
				continue
			}

			valueForValidation := matchedStr
			if entity.Normalization {
				valueForValidation = Normalize(matchedStr)
			}

			method := []string{"regex"}
			valid := true

			if entity.Validator != ValNone {
				valFunc := GetValidator(entity.Validator)
				if valFunc != nil {
					valid = valFunc(valueForValidation)
					if valid {
						method = append(method, "checksum")
					}
				}
			}

			if valid {
				h := sha256.Sum256([]byte(matchedStr))
				res := DetectionResult{
					Entity:     entity.Type,
					Value:      matchedStr,
					ValueHash:  fmt.Sprintf("%x", h),
					Confidence: 1.0,
					Method:     method,
				}
				res.Location.Start = start
				res.Location.End = end
				results = append(results, res)
			}
		}
	}
	return results
}

// Redact uses Scan to find PII and calls the tokenFunc to store/get a token,
// then replaces the PII with the token in the returned string.
func (e *Engine) Redact(input string, tokenFunc func(string, string, int, int) (string, error)) (string, error) {
	detections := e.Scan(input)
	if len(detections) == 0 {
		return input, nil
	}

	// Sort detections by start index, then longest first. If ties, prioritize contextual (ACCOUNT_NUMBER/PERSON)
	sort.Slice(detections, func(i, j int) bool {
		if detections[i].Location.Start != detections[j].Location.Start {
			return detections[i].Location.Start < detections[j].Location.Start
		}
		lenI := detections[i].Location.End - detections[i].Location.Start
		lenJ := detections[j].Location.End - detections[j].Location.Start
		if lenI != lenJ {
			return lenI > lenJ
		}
		// Tie-breaker: contextual extractors take precedence over pure regex IDs
		score := func(t string) int {
			if t == "ACCOUNT_NUMBER" || t == "PERSON" || t == "MEMO_TEXT" || t == "IBAN" {
				return 2
			}
			return 1
		}
		return score(detections[i].Entity) > score(detections[j].Entity)
	})

	var out strings.Builder
	lastPos := 0
	
	for _, d := range detections {
		if d.Location.Start < lastPos { // Skip overlaps
			continue
		}
		
		token, err := tokenFunc(d.Value, d.Entity, d.Location.Start, d.Location.End)
		if err != nil {
			return "", err
		}
		
		// If tokenFunc returns an empty string or the original value, it means SKIP replacement
		if token == "" || token == d.Value {
			continue
		}

		out.WriteString(input[lastPos:d.Location.Start])
		out.WriteString(token)
		lastPos = d.Location.End
	}
	out.WriteString(input[lastPos:])
	return out.String(), nil
}
