package application

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/cron"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/twitchtv/twirp"
)

// DashboardServiceServer aggregates read-only metrics for the dashboard
// Overview screen.
type DashboardServiceServer struct {
	agentRepo    configrepo.AgentRepository
	mcpRepo      configrepo.MCPServerRepository
	remoteRepo   configrepo.RemoteAgentRepository
	channelRepo  configrepo.ChannelRepository
	cronJobRepo  cron.JobRepo
	daemonReg    *daemon.Registry
	mongoDB      *mongo.Database
	rdb          *redis.Client
	runnerReady  func() bool
	invRepo      invocation.Repository
	cronExecRepo cron.ExecutionRepo
	sessionCount func(ctx context.Context) (int64, error)
}

// DashboardConfigStore is the union of config repository interfaces required
// for dashboard counts.
type DashboardConfigStore interface {
	configrepo.AgentRepository
	configrepo.MCPServerRepository
	configrepo.RemoteAgentRepository
	configrepo.ChannelRepository
}

func NewDashboardServiceServer(store DashboardConfigStore, daemonReg *daemon.Registry) *DashboardServiceServer {
	return &DashboardServiceServer{
		agentRepo:   store,
		mcpRepo:     store,
		remoteRepo:  store,
		channelRepo: store,
		daemonReg:   daemonReg,
	}
}

// SetMongo wires the MongoDB database used for liveness probes.
func (s *DashboardServiceServer) SetMongo(db *mongo.Database) { s.mongoDB = db }

// SetRedis wires the Redis client used for liveness probes.
func (s *DashboardServiceServer) SetRedis(rdb *redis.Client) { s.rdb = rdb }

// SetRunnerReady wires a callback returning whether the runner service has
// finished bootstrap. The callback may be nil before wiring.
func (s *DashboardServiceServer) SetRunnerReady(ready func() bool) { s.runnerReady = ready }

// SetCronJobRepo wires the cron job repository for cron_jobs counts.
func (s *DashboardServiceServer) SetCronJobRepo(repo cron.JobRepo) { s.cronJobRepo = repo }

// SetSessionCounter wires a callable that returns the total number of ADK
// sessions for the overview counts.
func (s *DashboardServiceServer) SetSessionCounter(fn func(ctx context.Context) (int64, error)) {
	s.sessionCount = fn
}

// SetInvocationRepo wires the invocation repository used by GetActivityFeed.
func (s *DashboardServiceServer) SetInvocationRepo(repo invocation.Repository) { s.invRepo = repo }

// SetCronExecutionRepo wires the cron execution repository used for
// GetCronExecutionTimeseries.
func (s *DashboardServiceServer) SetCronExecutionRepo(repo cron.ExecutionRepo) {
	s.cronExecRepo = repo
}

func (s *DashboardServiceServer) GetCronExecutionTimeseries(ctx context.Context, req *agentsv1.GetCronExecutionTimeseriesRequest) (*agentsv1.GetCronExecutionTimeseriesResponse, error) {
	if s.cronExecRepo == nil {
		return &agentsv1.GetCronExecutionTimeseriesResponse{}, nil
	}

	end := time.Now().UTC()
	var bucketSize time.Duration
	var span time.Duration
	switch req.GetRange() {
	case agentsv1.GetCronExecutionTimeseriesRequest_RANGE_7D:
		span = 7 * 24 * time.Hour
		bucketSize = 24 * time.Hour
	case agentsv1.GetCronExecutionTimeseriesRequest_RANGE_30D:
		span = 30 * 24 * time.Hour
		bucketSize = 24 * time.Hour
	default: // RANGE_1D and unspecified
		span = 24 * time.Hour
		bucketSize = time.Hour
	}
	start := end.Add(-span).Truncate(bucketSize)
	end = end.Truncate(bucketSize).Add(bucketSize)

	execs, err := s.cronExecRepo.ListByTimeRange(ctx, req.GetJobName(), start, end)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	bucketCount := int(end.Sub(start) / bucketSize)
	buckets := make([]*agentsv1.CronExecutionBucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		buckets[i] = &agentsv1.CronExecutionBucket{
			Start: timestamppb.New(start.Add(time.Duration(i) * bucketSize)),
		}
	}

	for _, exec := range execs {
		t := exec.GetStartedAt().AsTime()
		idx := int(t.Sub(start) / bucketSize)
		if idx < 0 || idx >= bucketCount {
			continue
		}
		if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR {
			buckets[idx].Error++
		} else {
			buckets[idx].Success++
		}
	}

	return &agentsv1.GetCronExecutionTimeseriesResponse{Buckets: buckets}, nil
}

func (s *DashboardServiceServer) GetActivityFeed(ctx context.Context, req *agentsv1.GetActivityFeedRequest) (*agentsv1.GetActivityFeedResponse, error) {
	if s.invRepo == nil {
		return &agentsv1.GetActivityFeedResponse{}, nil
	}
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	invs, next, err := s.invRepo.ListRecent(ctx, limit, req.GetPageToken())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	events := make([]*agentsv1.ActivityEvent, 0, len(invs))
	for _, inv := range invs {
		events = append(events, invocationToActivity(inv))
	}
	return &agentsv1.GetActivityFeedResponse{Events: events, NextPageToken: next}, nil
}

func invocationToActivity(inv *agentsv1.Invocation) *agentsv1.ActivityEvent {
	ts := inv.GetFinishedAt()
	if ts == nil {
		ts = inv.GetStartedAt()
	}
	kind := "invocation"
	msg := inv.GetInput()
	switch inv.GetStatus() {
	case agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED:
		kind = "execution_completed"
	case agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED:
		kind = "error"
		if inv.GetError() != "" {
			msg = inv.GetError()
		}
	}
	return &agentsv1.ActivityEvent{
		Id:        inv.GetId(),
		Kind:      kind,
		Actor:     inv.GetAgentName(),
		Message:   msg,
		Timestamp: ts,
	}
}

func (s *DashboardServiceServer) GetOverview(ctx context.Context, _ *agentsv1.GetOverviewRequest) (*agentsv1.GetOverviewResponse, error) {
	counts, err := s.counts(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	health := s.health(ctx)
	latest := s.latestHandshake()
	return &agentsv1.GetOverviewResponse{
		Counts:                counts,
		Health:                health,
		LatestDaemonHandshake: latest,
	}, nil
}

func (s *DashboardServiceServer) counts(ctx context.Context) (*agentsv1.OverviewCounts, error) {
	agents, err := s.agentRepo.ListAgents(ctx)
	if err != nil {
		return nil, err
	}
	mcp, err := s.mcpRepo.ListMCPServers(ctx)
	if err != nil {
		return nil, err
	}
	remote, err := s.remoteRepo.ListRemoteAgents(ctx)
	if err != nil {
		return nil, err
	}
	channels, err := s.channelRepo.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	out := &agentsv1.OverviewCounts{
		ActiveAgents: int32(len(agents)),
		McpServers:   int32(len(mcp)),
		RemoteAgents: int32(len(remote)),
		Channels:     int32(len(channels)),
	}
	if s.daemonReg != nil {
		out.ConnectedDaemons = int32(len(s.daemonReg.ListConnected()))
	}
	if s.cronJobRepo != nil {
		jobs, err := s.cronJobRepo.List(ctx)
		if err == nil {
			out.CronJobs = int32(len(jobs))
		}
	}
	if s.sessionCount != nil {
		if n, err := s.sessionCount(ctx); err == nil {
			out.ActiveSessions = int32(n)
		}
	}
	return out, nil
}

func (s *DashboardServiceServer) health(ctx context.Context) *agentsv1.HealthSummary {
	return &agentsv1.HealthSummary{
		Mongodb: s.checkMongo(ctx),
		Redis:   s.checkRedis(ctx),
		Runner:  s.checkRunner(),
	}
}

func (s *DashboardServiceServer) checkMongo(ctx context.Context) *agentsv1.ComponentHealth {
	if s.mongoDB == nil {
		return &agentsv1.ComponentHealth{
			Status: agentsv1.ComponentHealth_STATUS_UNSPECIFIED,
			Detail: "mongodb client not wired",
		}
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	start := time.Now()
	err := s.mongoDB.Client().Ping(pingCtx, nil)
	latency := time.Since(start).Milliseconds()
	now := timestamppb.New(time.Now().UTC())
	if err != nil {
		return &agentsv1.ComponentHealth{
			Status:    agentsv1.ComponentHealth_STATUS_DOWN,
			Detail:    err.Error(),
			CheckedAt: now,
			LatencyMs: latency,
		}
	}
	return &agentsv1.ComponentHealth{
		Status:    agentsv1.ComponentHealth_STATUS_HEALTHY,
		Detail:    s.mongoDB.Name(),
		CheckedAt: now,
		LatencyMs: latency,
	}
}

func (s *DashboardServiceServer) checkRedis(ctx context.Context) *agentsv1.ComponentHealth {
	if s.rdb == nil {
		return &agentsv1.ComponentHealth{
			Status: agentsv1.ComponentHealth_STATUS_UNSPECIFIED,
			Detail: "redis client not wired",
		}
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	start := time.Now()
	_, err := s.rdb.Ping(pingCtx).Result()
	latency := time.Since(start).Milliseconds()
	now := timestamppb.New(time.Now().UTC())
	if err != nil {
		return &agentsv1.ComponentHealth{
			Status:    agentsv1.ComponentHealth_STATUS_DEGRADED,
			Detail:    err.Error(),
			CheckedAt: now,
			LatencyMs: latency,
		}
	}
	return &agentsv1.ComponentHealth{
		Status:    agentsv1.ComponentHealth_STATUS_HEALTHY,
		CheckedAt: now,
		LatencyMs: latency,
	}
}

func (s *DashboardServiceServer) checkRunner() *agentsv1.ComponentHealth {
	now := timestamppb.New(time.Now().UTC())
	if s.runnerReady == nil {
		return &agentsv1.ComponentHealth{
			Status:    agentsv1.ComponentHealth_STATUS_UNSPECIFIED,
			Detail:    "runner readiness probe not wired",
			CheckedAt: now,
		}
	}
	if !s.runnerReady() {
		return &agentsv1.ComponentHealth{
			Status:    agentsv1.ComponentHealth_STATUS_DOWN,
			Detail:    "runner service not initialized",
			CheckedAt: now,
		}
	}
	return &agentsv1.ComponentHealth{Status: agentsv1.ComponentHealth_STATUS_HEALTHY, CheckedAt: now}
}

func (s *DashboardServiceServer) latestHandshake() *agentsv1.DaemonHandshake {
	if s.daemonReg == nil {
		return nil
	}
	conns := s.daemonReg.ListConnections()
	if len(conns) == 0 {
		return nil
	}
	latest := conns[0]
	for _, c := range conns[1:] {
		if c.ConnectedAt.After(latest.ConnectedAt) {
			latest = c
		}
	}
	return &agentsv1.DaemonHandshake{
		DaemonId:     latest.Info.GetDaemonId(),
		Name:         latest.Info.GetName(),
		Capabilities: latest.Info.GetCapabilities(),
		ConnectedAt:  timestamppb.New(latest.ConnectedAt),
		Os:           latest.Info.GetOs(),
	}
}
