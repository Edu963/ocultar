package connector

import (
	"fmt"
	"log"
	"plugin"
	"sync"

	"github.com/Edu963/ocultar/pkg/refinery"
)

// Registry maps connector types to their factory functions.
var (
	registry      = make(map[string]func() Connector)
	registryMutex sync.RWMutex
)

// Register adds a new connector factory to the global registry.
// This is typically called from a connector's init() function.
func Register(connectorType string, factory func() Connector) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	registry[connectorType] = factory
}

// New instantiates a new connector of the given type from the registry.
func New(connectorType string) (Connector, error) {
	registryMutex.RLock()
	factory, ok := registry[connectorType]
	registryMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("connector type %q not found in registry", connectorType)
	}

	return factory(), nil
}

// Manager handles the lifecycle of active connector instances.
type Manager struct {
	refinery     *refinery.Refinery
	connectors map[string]Connector
	mutex      sync.Mutex
}

// NewManager creates a new connector manager for the given refinery.
func NewManager(eng *refinery.Refinery) *Manager {
	return &Manager{
		refinery:     eng,
		connectors: make(map[string]Connector),
	}
}

// LoadAndStart instantiates and starts a connector based on the provided configuration.
func (m *Manager) LoadAndStart(id string, connectorType string, config map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.connectors[id]; exists {
		return fmt.Errorf("connector with ID %q already exists", id)
	}

	registryMutex.RLock()
	factory, ok := registry[connectorType]
	registryMutex.RUnlock()

	if !ok {
		return fmt.Errorf("connector type %q not found in registry", connectorType)
	}

	c := factory()
	if err := c.Init(config, m.refinery); err != nil {
		return fmt.Errorf("failed to initialize connector %q: %w", id, err)
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("failed to start connector %q: %w", id, err)
	}

	m.connectors[id] = c
	log.Printf("[INFO] Connector %q (%s) started matching 'Zero Egress' policy.", id, connectorType)
	return nil
}

// StopAll gracefully shuts down all active connectors.
func (m *Manager) StopAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, c := range m.connectors {
		log.Printf("[INFO] Stopping connector %q...", id)
		if err := c.Stop(); err != nil {
			log.Printf("[ERROR] Failed to stop connector %q: %v", id, err)
		}
		delete(m.connectors, id)
	}
}

// LoadPlugin loads a connector from a Go plugin (.so file).
// enterprise feature (Pro Connector bitmask check belongs in the caller).
func (m *Manager) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %q: %w", path, err)
	}

	sym, err := p.Lookup("NewConnector")
	if err != nil {
		return fmt.Errorf("plugin %q missing NewConnector symbol: %w", path, err)
	}

	factory, ok := sym.(func() Connector)
	if !ok {
		return fmt.Errorf("plugin %q: NewConnector has incorrect signature", path)
	}

	// Register it under a specific name or just store the factory?
	// For now, plugins will register themselves in their init() or we use this factory.
	// Let's assume the plugin wants to be managed.
	c := factory()
	registryMutex.Lock()
	registry[c.Type()] = factory
	registryMutex.Unlock()

	log.Printf("[INFO] Loaded connector plugin from %q (Type: %s)", path, c.Type())
	return nil
}
