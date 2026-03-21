package proxy

import (
	"regexp"
	"strings"

	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/vault"
)

// tokenRe matches OCULTAR vault tokens of the form [TYPE_hexhash8].
// Examples: [PERSON_ab3c12ef], [EMAIL_00fa9b12], [PHONE_cc847211]
var tokenRe = regexp.MustCompile(`\[[A-Z_]+_[0-9a-f]{8}\]`)

// RehydrateString scans s for vault tokens and replaces each with the
// original PII recovered from the vault. Tokens not found in the vault
// are left unchanged (safe fallback — they simply stay as redacted stubs).
func RehydrateString(v vault.Provider, masterKey []byte, s string) (string, error) {
	var firstErr error
	out := tokenRe.ReplaceAllStringFunc(s, func(tok string) string {
		plain, err := engine.DecryptToken(v, masterKey, tok)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return tok
		}
		return plain
	})
	return out, firstErr
}

// RehydrateInterface recursively walks a decoded JSON structure (as returned
// by json.Unmarshal into an interface{}) and re-hydrates every string value.
func RehydrateInterface(v vault.Provider, masterKey []byte, val interface{}) (interface{}, error) {
	switch typed := val.(type) {
	case string:
		return RehydrateString(v, masterKey, typed)
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for k, item := range typed {
			rehydrated, err := RehydrateInterface(v, masterKey, item)
			if err != nil {
				return nil, err
			}
			out[k] = rehydrated
		}
		return out, nil
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, item := range typed {
			rehydrated, err := RehydrateInterface(v, masterKey, item)
			if err != nil {
				return nil, err
			}
			out[i] = rehydrated
		}
		return out, nil
	default:
		return val, nil
	}
}

// ContainsTokens reports whether s contains at least one OCULTAR token.
func ContainsTokens(s string) bool {
	return tokenRe.MatchString(s)
}

// ContainsTokensInBody reports whether a raw byte slice contains any tokens,
// using a fast string search before invoking the full regex.
func ContainsTokensInBody(body []byte) bool {
	return strings.Contains(string(body), "[") && tokenRe.Match(body)
}
