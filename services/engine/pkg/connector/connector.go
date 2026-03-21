package connector

import (
	"context"
	"github.com/Edu963/ocultar/pkg/engine"
)

// Connector is the interface that all OCULTAR connectors must implement.
// Connectors are responsible for ingesting data from external sources
// (Slack, SharePoint, etc.) and feeding it into the refinery engine.
type Connector interface {
	// ID returns the unique identifier for this connector instance.
	ID() string

	// Type returns the type of the connector (e.g., "slack", "sharepoint").
	Type() string

	// Init initializes the connector with its configuration and a reference to the engine.
	Init(config map[string]interface{}, eng *engine.Engine) error

	// Start begins the data ingestion process. This is typically a non-blocking call
	// that starts background goroutines (e.g., listening for webhooks or polling).
	Start() error

	// Stop gracefully shuts down the connector and stops all background work.
	Stop() error

	// Fetch retrieves data from the source on-demand (pull-based).
	// Params are connector-specific (e.g., file path, message ID).
	Fetch(ctx context.Context, params map[string]interface{}) ([]byte, error)
}
