package application

import (
	"context"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/transport/connectx"
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

func (s *DaemonServiceServer) ListDaemons(ctx context.Context, _ *connect.Request[agentsv1.ListDaemonsRequest]) (*connect.Response[agentsv1.ListDaemonsResponse], error) {
	if s.registry == nil {
		return connect.NewResponse(&agentsv1.ListDaemonsResponse{}), nil
	}
	conns := s.registry.ListConnections()
	now := time.Now()
	out := make([]*agentsv1.DaemonStatus, 0, len(conns))
	for _, c := range conns {
		out = append(out, connectionToStatus(c, now))
	}
	return connect.NewResponse(&agentsv1.ListDaemonsResponse{Daemons: out}), nil
}

func (s *DaemonServiceServer) GetDaemon(ctx context.Context, req *connect.Request[agentsv1.GetDaemonRequest]) (*connect.Response[agentsv1.GetDaemonResponse], error) {
	if s.registry == nil {
		return nil, connectx.NotFound("daemon not found")
	}
	conn := s.registry.Get(req.Msg.GetDaemonId())
	if conn == nil {
		return nil, connectx.NotFound("daemon not found")
	}
	return connect.NewResponse(&agentsv1.GetDaemonResponse{Daemon: connectionToStatus(conn, time.Now())}), nil
}

func (s *DaemonServiceServer) CancelDaemonTask(ctx context.Context, req *connect.Request[agentsv1.CancelDaemonTaskRequest]) (*connect.Response[agentsv1.CancelDaemonTaskResponse], error) {
	if s.registry == nil {
		return nil, connectx.NotFound("daemon registry not available")
	}
	if req.Msg.GetTaskId() == "" {
		return nil, connectx.RequiredArgument("task_id")
	}

	var target *daemon.Connection
	if hint := req.Msg.GetDaemonId(); hint != "" {
		conn := s.registry.Get(hint)
		if conn == nil || !conn.HasTask(req.Msg.GetTaskId()) {
			return nil, connectx.NotFound("task not found on daemon")
		}
		target = conn
	} else {
		for _, conn := range s.registry.ListConnections() {
			if conn.HasTask(req.Msg.GetTaskId()) {
				target = conn
				break
			}
		}
		if target == nil {
			return nil, connectx.NotFound("task not found on any connected daemon")
		}
	}

	logger := log.FromContext(ctx)
	logger.Info("cancelling daemon task",
		"task_id", req.Msg.GetTaskId(),
		"daemon_id", target.Info.GetDaemonId(),
		"daemon_name", target.Info.GetName(),
	)
	if err := target.CancelTask(req.Msg.GetTaskId()); err != nil {
		logger.Error("cancel daemon task failed",
			"task_id", req.Msg.GetTaskId(),
			"daemon_id", target.Info.GetDaemonId(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("daemon task cancelled", "task_id", req.Msg.GetTaskId(), "daemon_id", target.Info.GetDaemonId())
	return connect.NewResponse(&agentsv1.CancelDaemonTaskResponse{DaemonId: target.Info.GetDaemonId()}), nil
}

func (s *DaemonServiceServer) ListDaemonTasks(ctx context.Context, req *connect.Request[agentsv1.ListDaemonTasksRequest]) (*connect.Response[agentsv1.ListDaemonTasksResponse], error) {
	if s.registry == nil {
		return connect.NewResponse(&agentsv1.ListDaemonTasksResponse{}), nil
	}
	var conns []*daemon.Connection
	if hint := req.Msg.GetDaemonId(); hint != "" {
		if c := s.registry.Get(hint); c != nil {
			conns = []*daemon.Connection{c}
		}
	} else {
		conns = s.registry.ListConnections()
	}
	now := time.Now()
	var out []*agentsv1.DaemonTaskInFlight
	for _, c := range conns {
		for _, snap := range c.ActiveTaskSnapshots() {
			cap := snap.Capability
			if cap == "" {
				if caps := c.Info.GetCapabilities(); len(caps) > 0 {
					cap = caps[0]
				}
			}
			task := &agentsv1.DaemonTaskInFlight{
				TaskId:      snap.TaskID,
				DaemonId:    c.Info.GetDaemonId(),
				DaemonName:  c.Info.GetName(),
				Capability:  cap,
				AgentName:   snap.AgentName,
				CurrentStep: snap.CurrentStep,
				Progress:    snap.Progress,
			}
			if !snap.StartedAt.IsZero() {
				task.StartedAt = timestamppb.New(snap.StartedAt)
				task.Elapsed = durationpb.New(now.Sub(snap.StartedAt))
			}
			out = append(out, task)
		}
	}
	return connect.NewResponse(&agentsv1.ListDaemonTasksResponse{Tasks: out}), nil
}

func (s *DaemonServiceServer) GetBridgeDiagnostics(ctx context.Context, _ *connect.Request[agentsv1.GetBridgeDiagnosticsRequest]) (*connect.Response[agentsv1.GetBridgeDiagnosticsResponse], error) {
	if s.registry == nil || s.registry.Metrics() == nil {
		return connect.NewResponse(&agentsv1.GetBridgeDiagnosticsResponse{Diagnostics: &agentsv1.BridgeDiagnostics{
			CheckedAt: timestamppb.New(time.Now().UTC()),
		}}), nil
	}
	snap := s.registry.Metrics().Snapshot()
	points := make([]*agentsv1.LatencyPoint, 0, len(snap.Latency))
	for _, sample := range snap.Latency {
		points = append(points, &agentsv1.LatencyPoint{
			Timestamp: timestamppb.New(sample.Timestamp),
			LatencyMs: sample.Latency.Milliseconds(),
		})
	}
	diag := &agentsv1.BridgeDiagnostics{
		CpuPercent:      snap.CPUPercent,
		MemoryUsedBytes: int64(snap.MemoryUsedBytes),
		Goroutines:      int32(snap.Goroutines),
		CheckedAt:       timestamppb.New(time.Now().UTC()),
		Latency:         points,
	}
	return connect.NewResponse(&agentsv1.GetBridgeDiagnosticsResponse{Diagnostics: diag}), nil
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
		Version:      c.Info.GetVersion(),
		Os:           c.Info.GetOs(),
		Executors:    c.Info.GetExecutors(),
		RemoteAddr:   c.RemoteAddr,
	}
}
