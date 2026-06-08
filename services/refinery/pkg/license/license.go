// Package license stubs out license enforcement for the open-source build.
// OCULTAR is fully open source — all features are available without a license key.
package license

// Payload is retained so that code referencing license.Active compiles unchanged.
type Payload struct {
	CustomerName string `json:"CustomerName"`
	Tier         string `json:"Tier"`
	ExpiryDate   int64  `json:"ExpiryDate"`
	Capabilities uint64 `json:"Capabilities"`
}

const (
	CapProSlack      uint64 = 1 << 0
	CapProSharePoint uint64 = 1 << 1
)

// Active always reflects open-source mode (all capabilities enabled).
var Active = Payload{Tier: "enterprise", Capabilities: ^uint64(0)}

// IsEnterprise always returns true — all features are available.
func IsEnterprise() bool { return true }

// HasProConnector always returns true — all connectors are available.
func HasProConnector(_ uint64) bool { return true }

// Load is a no-op; retained for call-site compatibility.
func Load() {}
