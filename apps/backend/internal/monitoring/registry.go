package monitoring

import "sync"

type ConnectorRegistry struct {
	mu         sync.RWMutex
	connectors map[string]Connector
}

func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[string]Connector),
	}
}

func (r *ConnectorRegistry) Register(connector Connector) {
	if r == nil || connector == nil {
		return
	}

	r.mu.Lock()
	r.connectors[connector.ID()] = connector
	r.mu.Unlock()
}

func (r *ConnectorRegistry) Unregister(id string) {
	if r == nil {
		return
	}

	r.mu.Lock()
	delete(r.connectors, id)
	r.mu.Unlock()
}

func (r *ConnectorRegistry) Get(id string) (Connector, bool) {
	if r == nil {
		return nil, false
	}

	r.mu.RLock()
	connector, ok := r.connectors[id]
	r.mu.RUnlock()

	return connector, ok
}

func (r *ConnectorRegistry) GetAll() []Connector {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	connectors := make([]Connector, 0, len(r.connectors))
	for _, connector := range r.connectors {
		connectors = append(connectors, connector)
	}
	r.mu.RUnlock()

	return connectors
}

func (r *ConnectorRegistry) GetByProvider(provider ProviderType) []Connector {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	connectors := make([]Connector, 0)
	for _, connector := range r.connectors {
		if connector.Provider() == provider {
			connectors = append(connectors, connector)
		}
	}
	r.mu.RUnlock()

	return connectors
}

func (r *ConnectorRegistry) Exists(id string) bool {
	if r == nil {
		return false
	}

	r.mu.RLock()
	_, ok := r.connectors[id]
	r.mu.RUnlock()

	return ok
}
