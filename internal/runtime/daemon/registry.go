package daemon

import (
	"sync"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Registry tracks connected daemon instances.
type Registry struct {
	mu      sync.RWMutex
	conns   map[string]map[string]*Connection // workspace_id → daemon_runtime_id → connection
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
func (r *Registry) Register(conn *Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	workspaceID := conn.WorkspaceID
	if workspaceID == "" {
		workspaceID = conn.Info.GetWorkspaceId()
	}
	if workspaceID == "" {
		return nil
	}
	bucket := r.conns[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*Connection)
		r.conns[workspaceID] = bucket
	}
	runtimeID := conn.Info.GetDaemonRuntimeId()
	if _, exists := bucket[runtimeID]; exists {
		return ErrRuntimeAlreadyConnected
	}
	bucket[runtimeID] = conn
	return nil
}

// RegisterOrReplace adds a daemon connection, replacing any existing
// connection for the same workspace/runtime. This is used by poll transport,
// where a daemon restart may reconnect before the stale connection expires.
func (r *Registry) RegisterOrReplace(conn *Connection) bool {
	var replaced *Connection

	r.mu.Lock()
	workspaceID := conn.WorkspaceID
	if workspaceID == "" {
		workspaceID = conn.Info.GetWorkspaceId()
	}
	if workspaceID == "" {
		r.mu.Unlock()
		return false
	}
	bucket := r.conns[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*Connection)
		r.conns[workspaceID] = bucket
	}
	runtimeID := conn.Info.GetDaemonRuntimeId()
	if existing := bucket[runtimeID]; existing != nil {
		replaced = existing
	}
	bucket[runtimeID] = conn
	r.mu.Unlock()

	if replaced != nil {
		replaced.Close()
		return true
	}
	return false
}

// Unregister removes a daemon from the registry by its ID.
func (r *Registry) Unregister(workspaceID, runtimeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bucket := r.conns[workspaceID]; bucket != nil {
		delete(bucket, runtimeID)
		if len(bucket) == 0 {
			delete(r.conns, workspaceID)
		}
	}
}

// PruneStalePollConnections removes poll-mode connections that have not
// checked in within maxIdle.
func (r *Registry) PruneStalePollConnections(maxIdle time.Duration) {
	if maxIdle <= 0 {
		return
	}

	now := time.Now()
	var stale []*Connection

	r.mu.Lock()
	for workspaceID, bucket := range r.conns {
		for runtimeID, conn := range bucket {
			if conn.stalePollConnection(now, maxIdle) {
				delete(bucket, runtimeID)
				stale = append(stale, conn)
			}
		}
		if len(bucket) == 0 {
			delete(r.conns, workspaceID)
		}
	}
	r.mu.Unlock()

	for _, conn := range stale {
		conn.Close()
	}
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

// Get returns the connection for the given runtime id, or nil if not found.
func (r *Registry) Get(workspaceID, runtimeID string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if bucket := r.conns[workspaceID]; bucket != nil {
		return bucket[runtimeID]
	}
	return nil
}
