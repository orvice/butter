package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"

	"go.orx.me/apps/butter/internal/repo/apitoken"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DaemonServiceServer exposes read-only views over the daemon registry.
type DaemonServiceServer struct {
	repo      configrepo.DaemonRuntimeRepository
	registry  *daemon.Registry
	tokenRepo apitoken.Repository
}

func NewDaemonServiceServer(repo configrepo.DaemonRuntimeRepository, registry *daemon.Registry) *DaemonServiceServer {
	return &DaemonServiceServer{repo: repo, registry: registry}
}

func (s *DaemonServiceServer) SetAPITokenRepo(repo apitoken.Repository) {
	s.tokenRepo = repo
}

func (s *DaemonServiceServer) ListDaemonRuntimes(ctx context.Context, _ *connect.Request[agentsv1.ListDaemonRuntimesRequest]) (*connect.Response[agentsv1.ListDaemonRuntimesResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	runtimes, err := s.repo.ListDaemonRuntimes(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.ListDaemonRuntimesResponse{Runtimes: runtimes}), nil
}

func (s *DaemonServiceServer) GetDaemonRuntime(ctx context.Context, req *connect.Request[agentsv1.GetDaemonRuntimeRequest]) (*connect.Response[agentsv1.GetDaemonRuntimeResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	d, err := s.repo.GetDaemonRuntime(ctx, wsID, req.Msg.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetDaemonRuntimeResponse{Runtime: d}), nil
}

func (s *DaemonServiceServer) CreateDaemonRuntime(ctx context.Context, req *connect.Request[agentsv1.CreateDaemonRuntimeRequest]) (*connect.Response[agentsv1.CreateDaemonRuntimeResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetRuntime() == nil {
		return nil, connectx.RequiredArgument("runtime")
	}
	d := proto.Clone(req.Msg.GetRuntime()).(*agentsv1.DaemonRuntime)
	if d.GetId() == "" {
		d.Id = uuid.NewString()
	}
	d.Id = strings.TrimSpace(d.GetId())
	d.Name = strings.TrimSpace(d.GetName())
	d.Description = strings.TrimSpace(d.GetDescription())
	if d.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	d.WorkspaceId = wsID
	d.CreatedAt = timestamppb.New(time.Now().UTC())
	if user, ok := auth.UserFromContext(ctx); ok {
		d.CreatedBy = user.GetId()
	}
	out, err := s.repo.CreateDaemonRuntime(ctx, wsID, d)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.CreateDaemonRuntimeResponse{Runtime: out}), nil
}

func (s *DaemonServiceServer) UpdateDaemonRuntime(ctx context.Context, req *connect.Request[agentsv1.UpdateDaemonRuntimeRequest]) (*connect.Response[agentsv1.UpdateDaemonRuntimeResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetRuntime() == nil {
		return nil, connectx.RequiredArgument("runtime")
	}
	d := proto.Clone(req.Msg.GetRuntime()).(*agentsv1.DaemonRuntime)
	d.Id = strings.TrimSpace(d.GetId())
	d.Name = strings.TrimSpace(d.GetName())
	d.Description = strings.TrimSpace(d.GetDescription())
	if d.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if d.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	prev, err := s.repo.GetDaemonRuntime(ctx, wsID, d.GetId())
	if err != nil {
		return nil, toConnectError(err)
	}
	d.WorkspaceId = wsID
	d.CreatedAt = prev.GetCreatedAt()
	d.CreatedBy = prev.GetCreatedBy()
	out, err := s.repo.UpdateDaemonRuntime(ctx, wsID, d)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.UpdateDaemonRuntimeResponse{Runtime: out}), nil
}

func (s *DaemonServiceServer) DeleteDaemonRuntime(ctx context.Context, req *connect.Request[agentsv1.DeleteDaemonRuntimeRequest]) (*connect.Response[agentsv1.DeleteDaemonRuntimeResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetId() == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.repo.DeleteDaemonRuntime(ctx, wsID, req.Msg.GetId()); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.DeleteDaemonRuntimeResponse{}), nil
}

func (s *DaemonServiceServer) CreateDaemonRuntimeToken(ctx context.Context, req *connect.Request[agentsv1.CreateDaemonRuntimeTokenRequest]) (*connect.Response[agentsv1.CreateDaemonRuntimeTokenResponse], error) {
	if s.tokenRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("api token store not available"))
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	runtimeID := strings.TrimSpace(req.Msg.GetDaemonRuntimeId())
	if runtimeID == "" {
		return nil, connectx.RequiredArgument("daemon_runtime_id")
	}
	if _, err := s.repo.GetDaemonRuntime(ctx, wsID, runtimeID); err != nil {
		return nil, toConnectError(err)
	}
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		name = "daemon runtime " + runtimeID
	}
	secret, err := generateSecret()
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	token := &agentsv1.APIToken{
		Id:              uuid.NewString(),
		Name:            name,
		Prefix:          tokenPrefix(secret),
		CreatedAt:       timestamppb.New(time.Now().UTC()),
		WorkspaceId:     wsID,
		Kind:            agentsv1.APITokenKind_API_TOKEN_KIND_DAEMON,
		Scopes:          []string{"daemon:connect"},
		DaemonRuntimeId: runtimeID,
	}
	if ttl := req.Msg.GetTtl(); ttl != nil && ttl.AsDuration() > 0 {
		token.ExpiresAt = timestamppb.New(time.Now().UTC().Add(ttl.AsDuration()))
	}
	if err := s.tokenRepo.Create(ctx, token, HashAPITokenSecret(secret)); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.CreateDaemonRuntimeTokenResponse{Token: token, Secret: secret}), nil
}

func (s *DaemonServiceServer) ListDaemons(ctx context.Context, _ *connect.Request[agentsv1.ListDaemonsRequest]) (*connect.Response[agentsv1.ListDaemonsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.registry == nil {
		return connect.NewResponse(&agentsv1.ListDaemonsResponse{}), nil
	}
	conns := s.registry.ListConnections(wsID)
	now := time.Now()
	out := make([]*agentsv1.DaemonStatus, 0, len(conns))
	for _, c := range conns {
		out = append(out, connectionToStatus(c, now))
	}
	return connect.NewResponse(&agentsv1.ListDaemonsResponse{Daemons: out}), nil
}

func (s *DaemonServiceServer) GetDaemon(ctx context.Context, req *connect.Request[agentsv1.GetDaemonRequest]) (*connect.Response[agentsv1.GetDaemonResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.registry == nil {
		return nil, connectx.NotFound("daemon not found")
	}
	conn := s.registry.Get(wsID, req.Msg.GetDaemonRuntimeId())
	if conn == nil {
		return nil, connectx.NotFound("daemon not found")
	}
	return connect.NewResponse(&agentsv1.GetDaemonResponse{Daemon: connectionToStatus(conn, time.Now())}), nil
}

func (s *DaemonServiceServer) CancelDaemonTask(ctx context.Context, req *connect.Request[agentsv1.CancelDaemonTaskRequest]) (*connect.Response[agentsv1.CancelDaemonTaskResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.registry == nil {
		return nil, connectx.NotFound("daemon registry not available")
	}
	if req.Msg.GetTaskId() == "" {
		return nil, connectx.RequiredArgument("task_id")
	}

	var target *daemon.Connection
	if hint := req.Msg.GetDaemonRuntimeId(); hint != "" {
		conn := s.registry.Get(wsID, hint)
		if conn == nil || !conn.HasTask(req.Msg.GetTaskId()) {
			return nil, connectx.NotFound("task not found on daemon")
		}
		target = conn
	} else {
		for _, conn := range s.registry.ListConnections(wsID) {
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
		"daemon_runtime_id", target.Info.GetDaemonRuntimeId(),
		"daemon_name", target.Info.GetName(),
	)
	if err := target.CancelTask(req.Msg.GetTaskId()); err != nil {
		logger.Error("cancel daemon task failed",
			"task_id", req.Msg.GetTaskId(),
			"daemon_runtime_id", target.Info.GetDaemonRuntimeId(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("daemon task cancelled", "task_id", req.Msg.GetTaskId(), "daemon_runtime_id", target.Info.GetDaemonRuntimeId())
	return connect.NewResponse(&agentsv1.CancelDaemonTaskResponse{DaemonRuntimeId: target.Info.GetDaemonRuntimeId()}), nil
}

func (s *DaemonServiceServer) ListDaemonTasks(ctx context.Context, req *connect.Request[agentsv1.ListDaemonTasksRequest]) (*connect.Response[agentsv1.ListDaemonTasksResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.registry == nil {
		return connect.NewResponse(&agentsv1.ListDaemonTasksResponse{}), nil
	}
	var conns []*daemon.Connection
	if hint := req.Msg.GetDaemonRuntimeId(); hint != "" {
		if c := s.registry.Get(wsID, hint); c != nil {
			conns = []*daemon.Connection{c}
		}
	} else {
		conns = s.registry.ListConnections(wsID)
	}
	now := time.Now()
	var out []*agentsv1.DaemonTaskInFlight
	for _, c := range conns {
		for _, snap := range c.ActiveTaskSnapshots() {
			acpRuntime := snap.AcpRuntime
			if acpRuntime == "" {
				if runtimes := c.Info.GetAcpRuntimes(); len(runtimes) > 0 {
					acpRuntime = runtimes[0]
				}
			}
			task := &agentsv1.DaemonTaskInFlight{
				TaskId:          snap.TaskID,
				DaemonRuntimeId: c.Info.GetDaemonRuntimeId(),
				DaemonName:      c.Info.GetName(),
				AcpRuntime:      acpRuntime,
				AgentName:       snap.AgentName,
				CurrentStep:     snap.CurrentStep,
				Progress:        snap.Progress,
				WorkspaceId:     snap.WorkspaceID,
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
		DaemonRuntimeId: c.Info.GetDaemonRuntimeId(),
		Name:            c.Info.GetName(),
		AcpRuntimes:     c.Info.GetAcpRuntimes(),
		Labels:          c.Info.GetLabels(),
		State:           state,
		ConnectedAt:     timestamppb.New(c.ConnectedAt),
		Uptime:          durationpb.New(now.Sub(c.ConnectedAt)),
		ActiveTasks:     int32(active),
		Version:         c.Info.GetVersion(),
		Os:              c.Info.GetOs(),
		Executors:       c.Info.GetExecutors(),
		RemoteAddr:      c.RemoteAddr,
		WorkspaceId:     c.WorkspaceID,
	}
}
