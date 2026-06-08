package policy

import (
	"strings"

	"github.com/Edu963/ocultar/internal/pii"
	"github.com/Edu963/ocultar/pkg/config"
)

// Decision is the result of evaluating all policies against a set of detected entities.
type Decision struct {
	Blocked       bool
	PolicyName    string
	BlockedEntity string // entity TYPE only — never the raw PII value
}

// Evaluate checks each detected PII hit against the active policies in order.
// The first matching "block" policy short-circuits and returns immediately.
// If no policy matches, the zero Decision (Blocked=false) is returned,
// meaning standard redaction proceeds unchanged.
func Evaluate(policies []config.Policy, hits []pii.DetectionResult) Decision {
	for _, hit := range hits {
		for _, p := range policies {
			if p.Action != "block" {
				continue
			}
			if !matchesEntity(p.When.Entity, hit.Entity) {
				continue
			}
			if p.When.MinConfidence > 0 && hit.Confidence < p.When.MinConfidence {
				continue
			}
			return Decision{
				Blocked:       true,
				PolicyName:    p.Name,
				BlockedEntity: hit.Entity,
			}
		}
	}
	return Decision{}
}

func matchesEntity(patterns []string, entity string) bool {
	for _, p := range patterns {
		if p == "*" || strings.EqualFold(p, entity) {
			return true
		}
	}
	return false
}
