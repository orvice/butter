package application

import (
	"context"
	"time"

	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DaemonServiceServer exposes read-only views over the daemon registry.
type DaemonServiceServer struct {
	registry *daemon.Registry
}

func NewDaemonServiceServer(registry *daemon.Registry) *DaemonServiceServer {
	return &DaemonServiceServer{registry: registry}
}

func (s *DaemonServiceServer) ListDaemons(ctx context.Context, _ *agentsv1.ListDaemonsRequest) (*agentsv1.ListDaemonsResponse, error) {
	if s.registry == nil {
		return &agentsv1.ListDaemonsResponse{}, nil
	}
	conns := s.registry.ListConnections()
	now := time.Now()
	out := make([]*agentsv1.DaemonStatus, 0, len(conns))
	for _, c := range conns {
		out = append(out, connectionToStatus(c, now))
	}
	return &agentsv1.ListDaemonsResponse{Daemons: out}, nil
}

func (s *DaemonServiceServer) GetDaemon(ctx context.Context, req *agentsv1.GetDaemonRequest) (*agentsv1.GetDaemonResponse, error) {
	if s.registry == nil {
		return nil, twirp.NotFoundError("daemon not found")
	}
	conn := s.registry.Get(req.GetDaemonId())
	if conn == nil {
		return nil, twirp.NotFoundError("daemon not found")
	}
	return &agentsv1.GetDaemonResponse{Daemon: connectionToStatus(conn, time.Now())}, nil
}

func (s *DaemonServiceServer) CancelDaemonTask(ctx context.Context, req *agentsv1.CancelDaemonTaskRequest) (*agentsv1.CancelDaemonTaskResponse, error) {
	if s.registry == nil {
		return nil, twirp.NotFoundError("daemon registry not available")
	}
	if req.GetTaskId() == "" {
		return nil, twirp.RequiredArgumentError("task_id")
	}

	var target *daemon.Connection
	if hint := req.GetDaemonId(); hint != "" {
		conn := s.registry.Get(hint)
		if conn == nil || !conn.HasTask(req.GetTaskId()) {
			return nil, twirp.NotFoundError("task not found on daemon")
		}
		target = conn
	} else {
		for _, conn := range s.registry.ListConnections() {
			if conn.HasTask(req.GetTaskId()) {
				target = conn
				break
			}
		}
		if target == nil {
			return nil, twirp.NotFoundError("task not found on any connected daemon")
		}
	}

	if err := target.CancelTask(req.GetTaskId()); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.CancelDaemonTaskResponse{DaemonId: target.Info.GetDaemonId()}, nil
}

func connectionToStatus(c *daemon.Connection, now time.Time) *agentsv1.DaemonStatus {
	active := c.ActiveTaskCount()
	state := agentsv1.DaemonStatus_STATE_ONLINE
	if active == 0 {
		state = agentsv1.DaemonStatus_STATE_IDLE
	}
	return &agentsv1.DaemonStatus{
		DaemonId:     c.Info.GetDaemonId(),
		Name:         c.Info.GetName(),
		Capabilities: c.Info.GetCapabilities(),
		Labels:       c.Info.GetLabels(),
		State:        state,
		ConnectedAt:  timestamppb.New(c.ConnectedAt),
		Uptime:       durationpb.New(now.Sub(c.ConnectedAt)),
		ActiveTasks:  int32(active),
	}
}
