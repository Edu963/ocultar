package refinery

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Edu963/ocultar/internal/pii"
	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/license"
	"github.com/Edu963/ocultar/vault"
)

// AuditLogger defines the interface for the Enterprise SIEM logger
type AuditLogger interface {
	Init(filePath string) error
	Log(user, action, result, mapping string)
	Close()
}

// AIScanner defines the interface for the Enterprise Deep Scan NER
type AIScanner interface {
	ScanForPII(text string) (map[string][]string, error)
	CheckHealth(host string)
	IsAvailable() bool
	SetDomain(domain string)
}

// NoopAuditLogger is the default nil-logger for the Community tier
type NoopAuditLogger struct{}

func (n NoopAuditLogger) Init(filePath string) error               { return nil }
func (n NoopAuditLogger) Log(user, action, result, mapping string) {}
func (n NoopAuditLogger) Close()                                   {}

// NoopAIScanner is the default nil-scanner for the Community tier
type NoopAIScanner struct{}

func (n NoopAIScanner) ScanForPII(text string) (map[string][]string, error) { return nil, nil }
func (n NoopAIScanner) CheckHealth(host string)                             {}
func (n NoopAIScanner) IsAvailable() bool                                   { return false }
func (n NoopAIScanner) SetDomain(domain string)                             {}

var tokenPattern = regexp.MustCompile(`\[[A-Z_]+_[0-9a-f]+\]`)
var greetingRegex = regexp.MustCompile(`(?m)(?i)(?:Regards|Best|Cheers|Bonjour|Hello|Hi|Dear|Sincerely|Cordialement)[,.-]*\s+([A-ZÀ-Ÿ][a-zà-ÿ]+(?:\s+[A-ZÀ-Ÿ][a-zà-ÿ]+){0,2})\b`)
var base64Regex = regexp.MustCompile(`([A-Za-z0-9+/=]{20,})`) // Lowered from 40: catches short PII (emails, names, phones) encoded in Base64

// Generalized Multilingual Heuristics (Phase 1)
var conjunctionRegex = regexp.MustCompile(`(?i)\b(ET|AND|Y|UND|CON|WITH|&)\b`)
var profTitleRegex = regexp.MustCompile(`(?i)\b(DR|DOCTEUR|PROF|MME|MLLE|SR|SRA|HR|FR|MAÎTRE|AVOCAT)\b`)
var capitalizedWordRegex = regexp.MustCompile(`\b[A-ZÀ-Ÿ][A-ZÀ-Ÿa-zà-ÿ\-]{1,20}\b`)
var possessiveRegex = regexp.MustCompile(`(?i)\b[A-ZÀ-Ÿ][a-zà-ÿ\-]{1,20}['’]s\b`)
var semanticTriggerRegex = regexp.MustCompile(`(?i)\b(DIVORCE|MARIAGE|WEDDING|AVOCAT|LAWYER|HOSPITAL|CLINIQUE|TREATMENT|TRAITEMENT|CAMPAIGN|POLITICAL|CAMPAGNE|PEA)\b`)

// DryRunReport collects PII hit metadata when running in --dry-run or --report mode.
type DryRunReport struct {
	Mode       string                `json:"mode"`
	FilesIn    int                   `json:"files_scanned"`
	Hits       []pii.DetectionResult `json:"pii_hits"`
	TotalCount int                   `json:"total_pii_count"`
	Blocking   bool                  `json:"blocking"`
}

// Refinery is the OCULTAR core redaction refinery.
// The storage backend is fully abstracted behind vault.Provider; the refinery
// has no knowledge of DuckDB, PostgreSQL, or any other concrete implementation.
type Refinery struct {
	Vault        vault.Provider
	MasterKey    []byte
	DryRun       bool
	Report       bool
	Serve        string
	PilotMode    bool
	SkipDeepScan bool
	VaultCount   *atomic.Int64
	AuditLogger  AuditLogger
	AIScanner    AIScanner

	Hits      []pii.DetectionResult
	hitsMutex sync.Mutex

	// SessionCache provides a fast-path for identical strings during a single batch/recursion run.
	SessionCache sync.Map
}

// NewRefinery constructs an Refinery using a vault.Provider as its storage backend.
func NewRefinery(v vault.Provider, key []byte) *Refinery {
	count := int64(0)
	if v != nil {
		count = v.CountAll()
	}
	e := &Refinery{
		Vault:       v,
		MasterKey:   key,
		VaultCount:  &atomic.Int64{},
		Hits:        []pii.DetectionResult{},
		AuditLogger: NoopAuditLogger{},
		AIScanner:   NoopAIScanner{},
	}
	e.VaultCount.Store(count)
	return e
}

// GenerateReport aggregates the current session's PII hits into a DryRunReport.
func (e *Refinery) GenerateReport(filesScanned int) DryRunReport {
	e.hitsMutex.Lock()
	defer e.hitsMutex.Unlock()

	blocking := len(e.Hits) > 0
	mode := "report"
	if e.DryRun {
		mode = "dry-run"
	}
	if e.Serve != "" {
		mode = "serve"
	}
	total := len(e.Hits)

	// Copy hits to avoid race conditions with JSON marshaling
	hitsCopy := append([]pii.DetectionResult{}, e.Hits...)

	return DryRunReport{
		Mode:       mode,
		FilesIn:    filesScanned,
		Hits:       hitsCopy,
		TotalCount: total,
		Blocking:   blocking,
	}
}

// ResetHits clears the in-memory record of detected PII and the session cache.
func (e *Refinery) ResetHits() {
	e.hitsMutex.Lock()
	defer e.hitsMutex.Unlock()
	e.Hits = []pii.DetectionResult{}
	e.SessionCache = sync.Map{}
}

// RefineBatch processes a slice of generic objects in parallel using a bounded worker pool.
// This enables High-Density Batch Tokenization for gigabyte-scale data ingestion.
func (e *Refinery) RefineBatch(items []interface{}, actor string) ([]interface{}, error) {
	if len(items) == 0 {
		return items, nil
	}

	results := make([]interface{}, len(items))
	errs := make([]error, len(items))
	var wg sync.WaitGroup

	// Bounded worker pool to prevent memory/goroutine exhaustion
	concurrency := 100
	sem := make(chan struct{}, concurrency)

	for i, item := range items {
		wg.Add(1)
		go func(idx int, val interface{}) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			res, err := e.ProcessInterface(val, actor)
			results[idx] = res
			errs[idx] = err
		}(i, item)
	}

	wg.Wait()

	// Fail-Closed: If any item fails in batch, the entire batch fails securely.
	for _, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("RefineBatch failed securely during processing: %w", err)
		}
	}

	return results, nil
}

// ProcessInterface recursively processes dynamic JSON data to identify and redact PII.
func (e *Refinery) ProcessInterface(data interface{}, actor string) (interface{}, error) {
	// 1. If it's a large complex object, extract all text and run SLM ONCE per record
	var preScanMap map[string][]string
	if e.AIScanner.IsAvailable() && !e.SkipDeepScan {
		// Marshal the record to a flat string to scan it contextually in one go
		textBytes, err := json.Marshal(data)
		if err == nil {
			textStr := string(textBytes)
			if len(textStr) > 50 && !e.isFullyTokenised(textStr) {
				var slmErr error
				preScanMap, slmErr = e.AIScanner.ScanForPII(textStr)
				if slmErr != nil {
					return nil, fmt.Errorf("SLM inference failed: %w", slmErr)
				}
				if len(preScanMap) > 0 {
					log.Printf("[INFO] SLM Batch Scan: %d entity type(s) detected in record", len(preScanMap))
				}
			}
		}
	}

	return e.processInterfaceRecursive(data, actor, preScanMap)
}

// Refine is a convenience method that delegates to ProcessInterface.
func (e *Refinery) Refine(data interface{}) (interface{}, error) {
	return e.ProcessInterface(data, "system-refinery")
}

// processInterfaceRecursive is the internal recursive helper for traversing JSON structures.
func (e *Refinery) processInterfaceRecursive(data interface{}, actor string, preScanMap map[string][]string) (interface{}, error) {
	switch val := data.(type) {
	case string:
		// Attempt Base64 decoding
		if decodedBytes, err := decodeBase64(val); err == nil && len(decodedBytes) > 0 {
			// Try to treat decoded Base64 as JSON or string
			var unmarshaled interface{}
			if err := json.Unmarshal(decodedBytes, &unmarshaled); err == nil {
				mod, err := e.processInterfaceRecursive(unmarshaled, actor, preScanMap)
				if err != nil {
					return nil, err
				}
				if remarshed, err := json.Marshal(mod); err == nil {
					return base64.StdEncoding.EncodeToString(remarshed), nil
				}
			}
			// Fallback: treat decoded Base64 as pure string
			refinedStr, err := e.RefineString(string(decodedBytes), actor, preScanMap)
			if err != nil {
				return nil, err
			}
			return base64.StdEncoding.EncodeToString([]byte(refinedStr)), nil
		}

		// Attempt URL decoding
		if strings.Contains(val, "%") {
			if unescaped, err := url.QueryUnescape(val); err == nil && unescaped != val {
				mod, err := e.processInterfaceRecursive(unescaped, actor, preScanMap)
				if err != nil {
					return nil, err
				}
				if modStr, ok := mod.(string); ok {
					return url.QueryEscape(modStr), nil
				} else if remarshed, err := json.Marshal(mod); err == nil {
					return url.QueryEscape(string(remarshed)), nil
				}
			}
		}

		// Attempt nested JSON decoding
		var unmarshaled interface{}
		if err := json.Unmarshal([]byte(val), &unmarshaled); err == nil {
			switch unmarshaled.(type) {
			case map[string]interface{}, []interface{}:
				mod, err := e.processInterfaceRecursive(unmarshaled, actor, preScanMap)
				if err != nil {
					return nil, err
				}
				if remarshed, err := json.Marshal(mod); err == nil {
					return string(remarshed), nil
				}
			}
		}

		return e.RefineString(val, actor, preScanMap)
	case map[string]interface{}:
		for k, v := range val {
			mod, err := e.processInterfaceRecursive(v, actor, preScanMap)
			if err != nil {
				return nil, err
			}
			val[k] = mod
		}
		return val, nil
	case []interface{}:
		if len(val) < 5 {
			// Sequential for small arrays to avoid goroutine overhead
			for i, v := range val {
				mod, err := e.processInterfaceRecursive(v, actor, preScanMap)
				if err != nil {
					return nil, err
				}
				val[i] = mod
			}
			return val, nil
		}

		// Parallel for larger arrays
		results := make([]interface{}, len(val))
		errs := make([]error, len(val))
		var wg sync.WaitGroup

		// Use a bounded worker pool (shared with RefineBatch logic)
		concurrency := 50 // Conservative default for recursion
		sem := make(chan struct{}, concurrency)

		for i, v := range val {
			wg.Add(1)
			go func(idx int, item interface{}) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				mod, err := e.processInterfaceRecursive(item, actor, preScanMap)
				results[idx] = mod
				errs[idx] = err
			}(i, v)
		}
		wg.Wait()

		for _, err := range errs {
			if err != nil {
				return nil, err
			}
		}
		return results, nil
	default:
		return val, nil
	}
}

// RefineString is the core logic that orchestrates PII detection tiers (Regex, Dictionaries, SLM) on a single string.
func (e *Refinery) RefineString(input string, actor string, preScanMap map[string][]string) (string, error) {
	if len(input) < 3 {
		return input, nil
	}

	// Session Cache Lookup: if we've already refined this exact string in this batch, return it.
	if cached, ok := e.SessionCache.Load(input); ok {
		return cached.(string), nil
	}

	trimmed := strings.TrimSpace(input)
	if len(trimmed) < 3 {
		return input, nil
	}

	refined := input
	var err error

	// TIER 0.1: Embedded Base64 Evasion Shield
	base64Matches := base64Regex.FindAllStringIndex(refined, -1)
	if len(base64Matches) > 0 {
		var out strings.Builder
		lastPos := 0
		for _, match := range base64Matches {
			start, end := match[0], match[1]
			out.WriteString(refined[lastPos:start])

			b64Str := refined[start:end]
			if decodedBytes, err := decodeBase64(b64Str); err == nil && len(decodedBytes) > 0 {
				mod, procErr := e.processInterfaceRecursive(string(decodedBytes), actor, preScanMap)
				if procErr == nil {
					if modStr, ok := mod.(string); ok {
						out.WriteString(base64.StdEncoding.EncodeToString([]byte(modStr)))
					} else if modBytes, err := json.Marshal(mod); err == nil {
						if len(modBytes) >= 2 && modBytes[0] == '"' && modBytes[len(modBytes)-1] == '"' {
							out.WriteString(base64.StdEncoding.EncodeToString(modBytes[1 : len(modBytes)-1]))
						} else {
							out.WriteString(base64.StdEncoding.EncodeToString(modBytes))
						}
					} else {
						out.WriteString(b64Str)
					}
				} else {
					out.WriteString(b64Str)
				}
			} else {
				out.WriteString(b64Str)
			}
			lastPos = end
		}
		out.WriteString(refined[lastPos:])
		refined = out.String()
	}

	// TIER 0.5: Dynamic Exclusion Dictionaries (VIPs, etc)
	for _, dictRule := range config.Global.Dictionaries {
		for _, term := range dictRule.Terms {
			refined, err = e.applyReplacement(refined, term, dictRule.Type, actor)
			if err != nil {
				return "", err
			}
		}
	}

	// TIER 1: Centralized Deterministic Pipeline (pii.Engine)
	eng := pii.NewEngine()
	tokens := tokenPattern.FindAllStringIndex(refined, -1)

	refined, err = eng.Redact(refined, func(d pii.DetectionResult) (string, error) {
		overlap := false
		for _, t := range tokens {
			if d.Range.Start < t[1] && d.Range.End > t[0] {
				overlap = true
				break
			}
		}
		if overlap {
			return d.Value, nil // Skip replacement by returning original value
		}
		return e.getOrSetSecureResult(d, actor)
	})
	if err != nil {
		return "", err
	}

	// TIER 1.1: Phone Shield (libphonenumber) - Fast-path: Only run if digits are present
	if strings.ContainsAny(refined, "0123456789") {
		refined, err = parseAndReplaceWithErr(refined, ParseAndReplacePhonesRaw, func(match string) (string, error) {
			return e.getOrSetSecureToken(match, "PHONE", actor)
		})
		if err != nil {
			return "", err
		}
	}

	// TIER 1.2: Modular Address Shield - Fast-path: Only run if digits are present or keywords like 'Rue/St/Ave'
	if len(refined) > 10 && (strings.ContainsAny(refined, "0123456789") || containsAnyLower(refined, "rue", "calle", "street", "ave", "road", "str.")) {
		refined, err = parseAndReplaceWithErr(refined, ParseAndReplaceAddressesRaw, func(match string) (string, error) {
			return e.getOrSetSecureToken(match, "ADDRESS", actor)
		})
		if err != nil {
			return "", err
		}
	}


	// TIER 1.5: Greetings & Signatures Shield
	greetingMatches := greetingRegex.FindAllStringSubmatchIndex(refined, -1)
	for _, match := range greetingMatches {
		if len(match) > 2 {
			start, end := match[2], match[3]
			nameStr := refined[start:end]
			if !strings.HasPrefix(nameStr, "[") {
				refined, err = e.applyReplacement(refined, nameStr, "PERSON", actor)
				if err != nil {
					return "", err
				}
			}
		}
	}

	// TIER 1.6: [VULN-001] Noise-Stripping Sliding Window Evasion Shield (Safety Net)
	// This pass strips non-alphanumeric noise AND tokens to scan for segmented patterns
	// that escaped the specific matches above.
	if strings.ContainsAny(refined, "0123456789") && !e.isFullyTokenised(refined) {
		// 1. Strip tokens first so they don't interfere with digit counts
		stripped := tokenPattern.ReplaceAllString(refined, " ")
		noiseWords := []string{"then", "is", "at", "and", "under", "with", "plus"}
		normalized := strings.ToLower(stripped)
		for _, w := range noiseWords {
			normalized = strings.ReplaceAll(normalized, " "+w+" ", " ")
		}
		regAlnum := regexp.MustCompile(`[^0-9]`) // Only interested in raw digits now
		tightDigits := regAlnum.ReplaceAllString(normalized, "")

		// High-fidelity Segmented SSN: exactly 9 digits.
		// If we find exactly 9 digits hidden in the remaining text, we fail-closed.
		if len(tightDigits) == 9 {
			log.Printf("[VULN-001] Segmented SSN (9 digits) detected in noise-stripped buffer.")
			return e.getOrSetSecureToken(refined, "SSN", actor)
		}

		// CC: 13-16 digits.
		if len(tightDigits) >= 13 && len(tightDigits) <= 16 {
			if isLuhnValid(tightDigits) {
				log.Printf("[VULN-001] Segmented Credit Card detected in noise-stripped buffer.")
				return e.getOrSetSecureToken(refined, "CREDIT_CARD", actor)
			}
		}
	}

	// TIER 2: SLM NER Scan
	if preScanMap != nil {
		for piiType, items := range preScanMap {
			for _, item := range items {
				if len(strings.TrimSpace(item)) > 2 && strings.Contains(refined, item) {
					refined, err = e.applyReplacement(refined, item, piiType, actor)
					if err != nil {
						return "", err
					}
				}
			}
		}
	} else if e.AIScanner.IsAvailable() && !e.SkipDeepScan && len(refined) > 15 && !e.isFullyTokenised(refined) {
		piiMap, slmErr := e.AIScanner.ScanForPII(refined)
		if slmErr != nil {
			return "", fmt.Errorf("SLM inference failed: %w", slmErr)
		}
		for piiType, items := range piiMap {
			for _, item := range items {
				if len(strings.TrimSpace(item)) > 2 {
					refined, err = e.applyReplacement(refined, item, piiType, actor)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}

	// TIER 3: Phase 1 Structural Heuristics (Zero-Egress Hardening)
	refined, err = e.applyStructuralHeuristics(refined, actor)
	if err != nil {
		return "", err
	}

	// Update session cache before returning
	e.SessionCache.Store(input, refined)

	return refined, nil
}

func containsAnyLower(s string, keywords ...string) bool {
	lower := strings.ToLower(s)
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// applyStructuralHeuristics executes generalized rules for entity expansion and linkages.
func (e *Refinery) applyStructuralHeuristics(input string, actor string) (string, error) {
	refined := input

	// 1. Semantic Scrubbing: [TRIGGER] [SUBJECT]
	// Done first to ensure it runs even if no tokens are present.
	refined, _ = replaceAllStringFuncErr(semanticTriggerRegex, refined, func(match string) (string, error) {
		// Redact the trigger itself to hide the sensitive category
		return e.getOrSetSecureToken(match, "SENSITIVE_EVENT", actor)
	})

	// 2. Professional Shield: [TITLE] [CAPITALIZED_NAME]
	refined, _ = replaceAllStringFuncErr(profTitleRegex, refined, func(match string) (string, error) {
		// Lookahead for capitalized words
		remaining := refined[strings.Index(refined, match)+len(match):]
		words := strings.Fields(remaining)
		if len(words) > 0 && capitalizedWordRegex.MatchString(words[0]) {
			// Redact the title and the following word(s)
			expanded := match + " " + words[0]
			// Greedy expansion for multi-part names after title
			for j := 1; j < len(words); j++ {
				if capitalizedWordRegex.MatchString(words[j]) {
					expanded += " " + words[j]
				} else {
					break
				}
			}
			return e.getOrSetSecureToken(expanded, "HEALTH_ENTITY", actor)
		}
		return match, nil // No expansion
	})

	// 3. Possessive Catch: [CAPITALIZED_WORD]'s
	refined, _ = replaceAllStringFuncErr(possessiveRegex, refined, func(match string) (string, error) {
		return e.getOrSetSecureToken(match, "PERSON", actor)
	})

	// 4. Greedy Neighborhood & Conjunctions: [TOKEN] [CONJUNCTION] [CAPITALIZED_NAME]
	tokens := tokenPattern.FindAllStringIndex(refined, -1)
	if len(tokens) == 0 {
		return refined, nil
	}

	var out strings.Builder
	lastPos := 0
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		start, end := t[0], t[1]
		if start < lastPos {
			continue // Already processed in an expanded token
		}

		out.WriteString(refined[lastPos:start])

		currentToken := refined[start:end]
		lookaheadEnd := end

		// Iterative Greedy Expansion
		for {
			remaining := refined[lookaheadEnd:]
			words := strings.Fields(remaining)
			if len(words) == 0 {
				break
			}

			firstWord := words[0]
			expandedThisTurn := false

			// Case A: Conjunction linkage (e.g. [TOKEN] ET MULLER)
			if conjunctionRegex.MatchString(firstWord) && len(words) > 1 && capitalizedWordRegex.MatchString(words[1]) {
				lookaheadEnd += strings.Index(remaining, words[1]) + len(words[1])
				expandedThisTurn = true
			} else if capitalizedWordRegex.MatchString(firstWord) || possessiveRegex.MatchString(firstWord) {
				// Case B: Direct surname proximity or possessive
				lookaheadEnd += strings.Index(remaining, firstWord) + len(firstWord)
				expandedThisTurn = true
			}

			if !expandedThisTurn {
				break
			}
		}

		if lookaheadEnd > end {
			// Expansion occurred
			expandedPII := refined[start:lookaheadEnd]
			piiType := strings.Split(strings.Trim(currentToken, "[]"), "_")[0]
			newToken, err := e.getOrSetSecureToken(expandedPII, piiType, actor)
			if err != nil {
				return "", err
			}
			out.WriteString(newToken)
			lastPos = lookaheadEnd
		} else {
			out.WriteString(currentToken)
			lastPos = end
		}
	}
	out.WriteString(refined[lastPos:])
	return out.String(), nil
}

// isFullyTokenised checks if a string consists entirely of redacted tokens and formatting characters.
func (e *Refinery) isFullyTokenised(s string) bool {
	stripped := tokenPattern.ReplaceAllString(s, "")
	return regexp.MustCompile(`^[\s\p{P}\p{Z}>*_|\-=+#@~]+$`).MatchString(stripped)
}

// applyRegexShield applies a specific regex pattern to discover and redact PII, avoiding overlapping matches.
func (e *Refinery) applyRegexShield(input string, re *regexp.Regexp, piiType string, actor string) (string, error) {
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

		out.WriteString(input[lastPos:start])

		tok, err := e.getOrSetSecureToken(matchedStr, piiType, actor)
		if err != nil {
			return "", err
		}

		out.WriteString(tok)
		lastPos = end
	}
	out.WriteString(input[lastPos:])
	return out.String(), nil
}

// applyReplacement replaces exact target strings with vaulted tokens.
func (e *Refinery) applyReplacement(line, target, piiType string, actor string) (string, error) {
	target = strings.TrimSpace(target)
	if len(target) < 3 {
		return line, nil
	}

	target = strings.ToValidUTF8(target, "")
	if len(target) < 3 {
		return line, nil
	}

	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(target))
	if err != nil {
		log.Printf("[WARN] applyReplacement: skipping invalid pattern for %q: %v", target, err)
		return line, nil
	}

	return replaceAllStringFuncErr(re, line, func(match string) (string, error) {
		return e.getOrSetSecureToken(match, piiType, actor)
	})
}

func (e *Refinery) getOrSetSecureToken(val, piiType string, actor string) (string, error) {
	res := pii.DetectionResult{
		Entity:     piiType,
		Value:      val,
		Confidence: 1.0,
		Method:     []string{"heuristic"},
	}
	// Note: We don't have start/end offsets here for simple string replacements
	return e.getOrSetSecureResult(res, actor)
}

// getOrSetSecureResult retrieves an existing token from the vault or generates, encrypts, and stores a new one.
func (e *Refinery) getOrSetSecureResult(res pii.DetectionResult, actor string) (string, error) {
	// [VULN-003] Enforce checksum validation for high-fidelity types
	if res.Entity == "CREDIT_CARD" && !isLuhnValid(res.Value) {
		// False positive avoidance: if it's not Luhn-valid, it's not a PII credit card
		return res.Value, nil
	}

	hash := sha256Hash(res.Value)
	token := fmt.Sprintf("[%s_%s]", res.Entity, hash[:8])

	if e.DryRun || e.Report || e.Serve != "" {
		e.hitsMutex.Lock()
		res.ValueHash = hash
		if res.Location == "" && res.Range.End > 0 {
			res.Location = fmt.Sprintf("%d-%d", res.Range.Start, res.Range.End)
		}
		e.Hits = append(e.Hits, res)
		e.hitsMutex.Unlock()
	}

	// Check vault for an existing token
	if existing, found := e.Vault.GetToken(hash); found {
		if license.Active.Tier == "enterprise" && !e.DryRun {
			e.AuditLogger.Log(actor, "matched", existing, getComplianceMapping(res.Entity))
		}
		return existing, nil
	}

	if license.Active.Tier == "enterprise" && !e.DryRun {
		e.AuditLogger.Log(actor, "vaulted", token, getComplianceMapping(res.Entity))
	}

	encrypted, encErr := encrypt([]byte(res.Value), e.MasterKey)
	if encErr != nil {
		return "", fmt.Errorf("encryption failed: %w", encErr)
	}
	inserted, err := e.Vault.StoreToken(hash, token, encrypted)
	if err != nil {
		return "", fmt.Errorf("vault storage failed: %w", err)
	}
	if inserted {
		e.VaultCount.Add(1)
	}
	return token, nil
}

func sha256Hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// isLuhnValid implements the Luhn algorithm (mod 10) for credit card checksum validation.
func isLuhnValid(s string) bool {
	// Strip non-digits
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

// Encrypt encrypts plaintext with AES-256-GCM using the provided key.
// The result is a hex-encoded string prefixed with the nonce.
func Encrypt(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Keep the unexported alias so internal call-sites are unaffected.
func encrypt(plaintext, key []byte) (string, error) { return Encrypt(plaintext, key) }

// Decrypt decrypts a hex-encoded AES-256-GCM ciphertext produced by Encrypt.
func Decrypt(hexCiphertext string, key []byte) ([]byte, error) {
	data, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// DecryptToken looks up a OCULTAR vault token (e.g. "[PERSON_abc12345]") using the
// original token unchanged if it is not found in the vault (safe fallback).
func DecryptToken(v vault.Provider, masterKey []byte, token string) (string, error) {
	// DecryptToken is called by the proxy layer; it needs to look up the encrypted
	// value by token, not by hash. We iterate — this is an infrequent operation.
	// We expose a raw SQL lookup only on the duckdbProvider via a type assertion.
	type tokenLookup interface {
		GetEncryptedByToken(token string) (string, bool)
	}
	if tl, ok := v.(tokenLookup); ok {
		encryptedPII, found := tl.GetEncryptedByToken(token)
		if !found {
			return token, nil
		}
		plaintext, err := Decrypt(encryptedPII, masterKey)
		if err != nil {
			log.Printf("[ERROR] decrypt error for token %s (key rotation?). Fail-safe: returning unhydrated token: %v", token, err)
			return token, nil
		}
		return string(plaintext), nil
	}
	// Fall back: no reverse-lookup capability — return token as-is (safe)
	return token, nil
}

// replaceAllStringFuncErr applies a replacement function that can return an error
func replaceAllStringFuncErr(re *regexp.Regexp, input string, repl func(string) (string, error)) (string, error) {
	matches := re.FindAllStringIndex(input, -1)
	if len(matches) == 0 {
		return input, nil
	}

	var out strings.Builder
	lastPos := 0
	for _, match := range matches {
		start, end := match[0], match[1]
		out.WriteString(input[lastPos:start])

		r, err := repl(input[start:end])
		if err != nil {
			return "", err
		}
		out.WriteString(r)
		lastPos = end
	}
	out.WriteString(input[lastPos:])
	return out.String(), nil
}

// Helper types for migrating address/phone parsers to support errors
func parseAndReplaceWithErr(input string, extractor func(string) [][]int, repl func(string) (string, error)) (string, error) {
	matches := extractor(input)
	if len(matches) == 0 {
		return input, nil
	}

	tokens := tokenPattern.FindAllStringIndex(input, -1)

	var out strings.Builder
	lastPos := 0
	for _, match := range matches {
		start, end := match[0], match[1]

		// Ensure we aren't carving into already tokenized variables
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

		// If matches overlap due to nested tokens or bad indices, skip
		if start < lastPos {
			continue
		}

		out.WriteString(input[lastPos:start])

		r, err := repl(input[start:end])
		if err != nil {
			return "", err
		}
		out.WriteString(r)
		lastPos = end
	}
	out.WriteString(input[lastPos:])
	return out.String(), nil
}

// decodeBase64 attempts to decode standard base64 strings, and falls back to raw
// unpadded decoding to catch obfuscated PII.
func decodeBase64(s string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}

func getComplianceMapping(piiType string) string {
	if config.Global.RegulatoryPolicy == nil {
		return "GENERAL_PII"
	}

	mappings, ok := config.Global.RegulatoryPolicy["mappings"].(map[string]interface{})
	if !ok {
		return "GENERAL_PII"
	}

	if m, ok := mappings[piiType].(map[string]interface{}); ok {
		if reg, ok := m["regulation"].(string); ok {
			return reg
		}
	}

	return "GENERAL_PII"
}
