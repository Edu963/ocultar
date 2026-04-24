package inference

// Scanner is the engine-agnostic contract for all in-process PII backends.
// Implementations: LlamaScanner (llama.cpp CGO), PrivacyFilterScanner (HTTP→Python).
type Scanner interface {
	ScanForPII(text string) (map[string][]string, error)
	Close()
}
