package daemon

import (
	"sync"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Registry tracks connected daemon instances.
type Registry struct {
	mu    sync.RWMutex
	conns map[string]*Connection // daemon_id → connection
}

// NewRegistry creates an empty daemon registry.
func NewRegistry() *Registry {
	return &Registry{
		conns: make(map[string]*Connection),
	}
}

// Register adds a connected daemon to the registry.
func (r *Registry) Register(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[conn.Info.DaemonId] = conn
}

// Unregister removes a daemon from the registry by its ID.
func (r *Registry) Unregister(daemonID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, daemonID)
}

// FindByCapability returns the first connected daemon that declares the given
// capability, or nil if none is available.
func (r *Registry) FindByCapability(capability string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.conns {
		for _, cap := range conn.Info.Capabilities {
			if cap == capability {
				return conn
			}
		}
	}
	return nil
}

// ListConnected returns info for all currently connected daemons.
func (r *Registry) ListConnected() []*agentsv1.DaemonInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*agentsv1.DaemonInfo, 0, len(r.conns))
	for _, conn := range r.conns {
		result = append(result, conn.Info)
	}
	return result
}

// ListConnections returns all currently connected daemon connections. The
// returned slice is a snapshot; the underlying connections are still managed
// by the registry.
func (r *Registry) ListConnections() []*Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Connection, 0, len(r.conns))
	for _, conn := range r.conns {
		result = append(result, conn)
	}
	return result
}

// Get returns the connection for the given daemon id, or nil if not found.
func (r *Registry) Get(daemonID string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conns[daemonID]
}
