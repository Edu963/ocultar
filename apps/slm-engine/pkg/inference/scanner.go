package inference

// Tier2Engine is the engine-agnostic contract for all in-process PII backends.
// Implementations: PrivacyFilterEngine (HTTP→Python).
// TODO: implement LlamaCppEngine for air-gapped deployments
// requiring local .gguf model execution. See Tier2Engine interface.
type Tier2Engine interface {
	// TODO: migrate to []Entity with offsets and confidence
	// scores once Privacy Filter is validated in production.
	ScanForPII(text string) (map[string][]string, error)
	Name() string
	Close()
}
