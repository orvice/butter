package daemon

import (
	"sync"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Registry tracks connected daemon instances.
type Registry struct {
	mu      sync.RWMutex
	conns   map[string]map[string]*Connection // workspace_id → daemon_id → connection
	metrics *Metrics
}

// NewRegistry creates an empty daemon registry with a fresh bridge metrics
// collector.
func NewRegistry() *Registry {
	return &Registry{
		conns:   make(map[string]map[string]*Connection),
		metrics: NewMetrics(60),
	}
}

// Metrics returns the bridge metrics collector backing this registry. Bridges
// read this to record per-invocation latency; the dashboard reads it for the
// diagnostics view.
func (r *Registry) Metrics() *Metrics {
	return r.metrics
}

// Register adds a connected daemon to the registry.
func (r *Registry) Register(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	workspaceID := conn.WorkspaceID
	if workspaceID == "" {
		workspaceID = conn.Info.GetWorkspaceId()
	}
	if workspaceID == "" {
		return
	}
	bucket := r.conns[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*Connection)
		r.conns[workspaceID] = bucket
	}
	bucket[conn.Info.DaemonId] = conn
}

// Unregister removes a daemon from the registry by its ID.
func (r *Registry) Unregister(workspaceID, daemonID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bucket := r.conns[workspaceID]; bucket != nil {
		delete(bucket, daemonID)
		if len(bucket) == 0 {
			delete(r.conns, workspaceID)
		}
	}
}

// FindByCapability returns the first connected daemon that declares the given
// capability, or nil if none is available.
func (r *Registry) FindByCapability(workspaceID, capability string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.conns[workspaceID] {
		for _, cap := range conn.Info.Capabilities {
			if cap == capability {
				return conn
			}
		}
	}
	return nil
}

// ListConnected returns info for all currently connected daemons.
func (r *Registry) ListConnected(workspaceID string) []*agentsv1.DaemonInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*agentsv1.DaemonInfo
	if workspaceID != "" {
		for _, conn := range r.conns[workspaceID] {
			result = append(result, conn.Info)
		}
		return result
	}
	for _, bucket := range r.conns {
		for _, conn := range bucket {
			result = append(result, conn.Info)
		}
	}
	return result
}

// ListConnections returns all currently connected daemon connections. The
// returned slice is a snapshot; the underlying connections are still managed
// by the registry.
func (r *Registry) ListConnections(workspaceID string) []*Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*Connection
	if workspaceID != "" {
		for _, conn := range r.conns[workspaceID] {
			result = append(result, conn)
		}
		return result
	}
	for _, bucket := range r.conns {
		for _, conn := range bucket {
			result = append(result, conn)
		}
	}
	return result
}

// Get returns the connection for the given daemon id, or nil if not found.
func (r *Registry) Get(workspaceID, daemonID string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if bucket := r.conns[workspaceID]; bucket != nil {
		return bucket[daemonID]
	}
	return nil
}
