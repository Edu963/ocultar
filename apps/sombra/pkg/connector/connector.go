// Package connector defines the data source abstraction for Sombra.
// Every data source (file uploads, REST APIs, FHIR, Plaid, etc.) implements
// the Connector interface and declares a DataPolicy governing what PII
// categories are stripped, redacted, or passed through.
package connector

import "context"

// Connector is the contract every data source adapter must satisfy.
type Connector interface {
	// Name returns a human-readable identifier (e.g. "file", "plaid", "fhir").
	Name() string

	// Fetch retrieves data from the external source.
	// The returned FetchResponse contains raw content that will be run through
	// the OCULTAR refinery before reaching any AI model.
	Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error)

	// Policy returns the data governance rules for this connector.
	Policy() DataPolicy
}

// FetchRequest describes what data to pull from the source.
type FetchRequest struct {
	// SourceID identifies the resource (file path, account ID, patient ID, etc.)
	SourceID string
	// RawBody is the raw request body (used by the file connector for uploads).
	RawBody []byte
	// ContentType is the MIME type of RawBody, if provided.
	ContentType string
	// Parameters holds connector-specific key/value options.
	Parameters map[string]string
	// Actor is the identity of the requester (for audit logging).
	Actor string
}

// FetchResponse is the raw data retrieved from a source.
type FetchResponse struct {
	// ContentType is the MIME type of the body (e.g. "application/json", "text/csv").
	ContentType string
	// Body is the raw content bytes.
	Body []byte
	// Metadata holds any source-specific metadata (transaction count, date range, etc.)
	Metadata map[string]string
}

// DataPolicy defines the PII governance rules for a given connector.
type DataPolicy struct {
	// StripCategories lists PII types to remove entirely (no vault, no token).
	// Example: ["SSN", "ACCOUNT_NUMBER"]
	StripCategories []string `yaml:"strip_categories"`

	// RedactCategories lists PII types to vault-tokenise.
	// If empty, all detected PII is redacted (default behaviour).
	RedactCategories []string `yaml:"redact_categories"`

	// AllowedModels restricts which AI backends may receive data from this connector.
	// An empty list means all configured models are allowed.
	AllowedModels []string `yaml:"allowed_models"`

	// MaxBodyBytes limits the size of data pulled per request (0 = unlimited).
	MaxBodyBytes int64 `yaml:"max_body_bytes"`
}

// IsModelAllowed checks whether a given model name is permitted by this policy.
// An empty AllowedModels list means all models are allowed.
func (p DataPolicy) IsModelAllowed(model string) bool {
	if len(p.AllowedModels) == 0 {
		return true
	}
	for _, m := range p.AllowedModels {
		if m == model {
			return true
		}
	}
	return false
}

// ShouldStrip checks whether a PII category should be stripped (removed entirely).
func (p DataPolicy) ShouldStrip(category string) bool {
	for _, c := range p.StripCategories {
		if c == category {
			return true
		}
	}
	return false
}
